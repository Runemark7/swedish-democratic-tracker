package ingestion

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"

	"riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/ingestion/workers"
)

// JobTimeout is the limit for one worker. The chain runs workers one after
// another, so a hung fetch cannot hold the jobs that depend on it, and it
// cannot sit on the scheduler until the next midnight.
const JobTimeout = 20 * time.Minute

type Scheduler struct {
	cron       *cron.Cron
	entries    []scheduled
	running    atomic.Bool
	startOnce  sync.Once
	jobTimeout time.Duration
	now        func() time.Time
	runs       ports.IngestionRunRepository
	freshness  FreshnessSource
}

type scheduled struct {
	spec string
	w    Worker
}

type Worker interface {
	Name() string
	Run(ctx context.Context) error
}

func NewScheduler() *Scheduler {
	s := &Scheduler{
		jobTimeout: JobTimeout,
		now:        func() time.Time { return time.Now().UTC() },
	}
	s.cron = cron.New(
		cron.WithLocation(time.UTC),
		cron.WithChain(
			cron.Recover(cronLogger{}),
			cron.SkipIfStillRunning(cronLogger{}),
		),
	)
	return s
}

// SetRunRecorder stores a row for every worker the chain runs. Nil skips the
// log; a failure to write it does not skip the worker.
func (s *Scheduler) SetRunRecorder(r ports.IngestionRunRepository) { s.runs = r }

// SetFreshness checks speeches, votes and job success after each chain.
// Stale data is logged at error. Nil skips the check.
func (s *Scheduler) SetFreshness(f FreshnessSource) { s.freshness = f }

func (s *Scheduler) Register(spec string, w Worker) error {
	switch spec {
	case "@daily", "@weekly":
	default:
		return fmt.Errorf("unsupported schedule %q", spec)
	}
	s.entries = append(s.entries, scheduled{spec: spec, w: w})
	return nil
}

// RegisterSync adds a worker to the ordered chain. Startup and the daily
// cron walk that chain; there is no second list to drift out of date.
func (s *Scheduler) RegisterSync(spec string, w Worker) error {
	return s.Register(spec, w)
}

// Catalog is the registered chain in dependency order.
func (s *Scheduler) Catalog() []ScheduledJob {
	out := make([]ScheduledJob, len(s.entries))
	for i, e := range s.entries {
		out[i] = ScheduledJob{Name: e.w.Name(), Spec: e.spec}
	}
	return out
}

// RegisterDefaults wires the vote pipeline in dependency order.
//
// One daily cron runs the chain. Weekly workers run on Sunday, in their
// place in that order, so refresh-scorecards follows keyword-matcher and
// refresh-coverage follows the vote fetch. Independent @daily registrations
// all fire at 00:00 and finish in parallel, which is how coverage and
// scorecards kept publishing yesterday's inputs.
func (s *Scheduler) RegisterDefaults(
	pw workers.PoliticiansWorker,
	sw workers.SpeechesWorker,
	vw workers.VotesWorker,
	ew workers.EnrichOriginsWorker,
	kw workers.KeywordMatcherWorker,
	rw workers.RefreshScorecardsWorker,
	cw workers.CoverageWorker,
) error {
	ordered := []struct {
		spec string
		w    Worker
	}{
		{"@daily", &pw},
		{"@daily", &sw},
		{"@daily", &vw},
		{"@daily", &ew},
		{"@daily", &kw},
		{"@weekly", &rw},
		{"@daily", &cw},
	}
	for _, e := range ordered {
		if err := s.Register(e.spec, e.w); err != nil {
			return fmt.Errorf("register %s: %w", e.w.Name(), err)
		}
	}
	return nil
}

// Start arms the single daily chain. Workers registered later are still
// picked up: the cron callback reads the chain when it fires.
func (s *Scheduler) Start() {
	s.startOnce.Do(func() {
		if _, err := s.cron.AddFunc("@daily", func() {
			s.RunDue(context.Background())
		}); err != nil {
			slog.Error("failed to schedule ingestion chain", "error", err)
			return
		}
		s.cron.Start()
	})
}

func (s *Scheduler) Stop() { s.cron.Stop() }

// RunDue runs every worker whose schedule is due, in registration order.
// A chain that is already running is skipped. One worker's panic or error
// does not stop the rest: a later job is often what makes the failure visible.
func (s *Scheduler) RunDue(ctx context.Context) {
	if !s.running.CompareAndSwap(false, true) {
		slog.Warn("ingestion chain still running, skipping")
		return
	}
	defer s.running.Store(false)

	now := s.now()
	slog.Info("ingestion chain starting", "at", now.Format(time.RFC3339))
	for _, e := range s.entries {
		if ctx.Err() != nil {
			slog.Error("ingestion chain stopped", "error", ctx.Err())
			return
		}
		if !due(e.spec, now) {
			continue
		}
		s.runWorker(ctx, e.w, false)
	}
	s.reportFreshness(ctx)
}

// RunInitialSync runs the whole chain once, including weekly workers, in the
// same order as the cron. Enrich-vote-origins loops until it drains, which
// the daily cron does not: the cron enriches one batch.
func (s *Scheduler) RunInitialSync(ctx context.Context) error {
	if !s.running.CompareAndSwap(false, true) {
		slog.Warn("ingestion chain still running, skipping initial sync")
		return nil
	}
	defer s.running.Store(false)

	slog.Info("initial sync: starting background data ingestion from Riksdagen")
	start := time.Now()
	for _, e := range s.entries {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		s.runWorker(ctx, e.w, true)
	}
	s.reportFreshness(ctx)
	slog.Info("initial sync: complete", "elapsed", time.Since(start))
	return nil
}

func (s *Scheduler) runWorker(ctx context.Context, w Worker, loopEnrich bool) {
	if ctx.Err() != nil {
		slog.Error("ingestion worker skipped", "worker", w.Name(), "error", ctx.Err())
		return
	}
	jobCtx, cancel := context.WithTimeout(ctx, s.jobTimeout)
	defer cancel()

	var runID int64
	if s.runs != nil {
		id, err := s.runs.StartRun(jobCtx, w.Name())
		if err != nil {
			slog.Warn("ingestion run log failed", "worker", w.Name(), "error", err)
		} else {
			runID = id
		}
	}

	slog.Info("ingestion worker starting", "worker", w.Name())
	started := time.Now()
	err := s.invoke(jobCtx, w, loopEnrich)
	elapsed := time.Since(started)

	if runID != 0 && s.runs != nil {
		// The job context may already be expired. Recording the outcome has
		// to outlive it, or a timeout erases the evidence of the timeout.
		finCtx, finCancel := context.WithTimeout(context.Background(), 10*time.Second)
		if ferr := s.runs.FinishRun(finCtx, runID, 0, err); ferr != nil {
			slog.Warn("ingestion run finish failed", "worker", w.Name(), "error", ferr)
		}
		finCancel()
	}

	if err != nil {
		slog.Error("ingestion worker failed", "worker", w.Name(), "error", err, "elapsed", elapsed)
		return
	}
	slog.Info("ingestion worker done", "worker", w.Name(), "elapsed", elapsed)
}

func (s *Scheduler) invoke(ctx context.Context, w Worker, loopEnrich bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("ingestion worker panicked",
				"worker", w.Name(), "panic", r, "stack", string(debug.Stack()))
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	if loopEnrich {
		if ew, ok := w.(*workers.EnrichOriginsWorker); ok {
			return s.runEnrichLoop(ctx, ew)
		}
	}
	return w.Run(ctx)
}

func (s *Scheduler) runEnrichLoop(ctx context.Context, w *workers.EnrichOriginsWorker) error {
	for round := 1; round <= 50; round++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		n, err := w.RunBatch(ctx)
		if err != nil {
			return err
		}
		slog.Info("initial sync: enrich round", "round", round, "enriched", n)
		if n == 0 {
			slog.Info("initial sync: enrich complete, no more unenriched votes", "rounds", round)
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return nil
}

func (s *Scheduler) reportFreshness(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("ingestion freshness check panicked", "panic", r, "stack", string(debug.Stack()))
		}
	}()
	if s.freshness == nil {
		return
	}
	checkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	snap, err := s.freshness.Snapshot(checkCtx)
	if err != nil {
		slog.Error("ingestion freshness check failed", "error", err)
		return
	}
	snap.Jobs = MergeCatalog(s.Catalog(), snap.Jobs)
	for _, alert := range Alerts(snap, s.now()) {
		slog.Error("ingestion data stale", "alert", alert)
	}
}

func due(spec string, now time.Time) bool {
	switch spec {
	case "@daily":
		return true
	case "@weekly":
		return now.Weekday() == time.Sunday
	default:
		return false
	}
}

// cronLogger adapts slog to robfig/cron. Skip and recover must go through it
// or a skipped overlap and a recovered panic are invisible next to the
// worker logs.
type cronLogger struct{}

func (cronLogger) Info(msg string, keysAndValues ...interface{}) {
	slog.Info(msg, keysAndValues...)
}

func (cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	args := make([]interface{}, 0, len(keysAndValues)+2)
	args = append(args, "error", err)
	args = append(args, keysAndValues...)
	slog.Error(msg, args...)
}
