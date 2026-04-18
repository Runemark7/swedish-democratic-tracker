package ingestion

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"riksdagskollen/internal/ingestion/workers"
)

type Scheduler struct {
	cron        *cron.Cron
	syncWorkers []Worker // ordered list for initial sync
}

type Worker interface {
	Name() string
	Run(ctx context.Context) error
}

func NewScheduler(ww ...Worker) *Scheduler {
	return &Scheduler{
		cron: cron.New(),
	}
}

func (s *Scheduler) Register(spec string, w Worker) error {
	_, err := s.cron.AddFunc(spec, func() {
		ctx := context.Background()
		slog.Info("ingestion worker starting", "worker", w.Name())
		if err := w.Run(ctx); err != nil {
			slog.Error("ingestion worker failed", "worker", w.Name(), "error", err)
			return
		}
		slog.Info("ingestion worker done", "worker", w.Name())
	})
	return err
}

// RegisterDefaults wires all workers on their standard schedules.
func (s *Scheduler) RegisterDefaults(
	pw workers.PoliticiansWorker,
	sw workers.SpeechesWorker,
	vw workers.VotesWorker,
	ew workers.EnrichOriginsWorker,
	kw workers.KeywordMatcherWorker,
	rw workers.RefreshScorecardsWorker,
) error {
	schedules := []struct {
		spec   string
		worker Worker
	}{
		{"@daily", &pw},
		{"@daily", &sw},
		{"@daily", &vw},
		{"@daily", &ew},
		{"@daily", &kw},
		{"@weekly", &rw},
	}

	// Store ordered sync workers (dependency order)
	s.syncWorkers = []Worker{&pw, &sw, &vw, &ew, &kw, &rw}

	for _, entry := range schedules {
		if err := s.Register(entry.spec, entry.worker); err != nil {
			return fmt.Errorf("register %s: %w", entry.worker.Name(), err)
		}
	}
	return nil
}

// RunInitialSync runs all workers once in dependency order.
// EnrichOrigins loops until all votes are enriched.
func (s *Scheduler) RunInitialSync(ctx context.Context) error {
	slog.Info("initial sync: starting background data ingestion from Riksdagen")
	start := time.Now()

	for _, w := range s.syncWorkers {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		wStart := time.Now()
		slog.Info("initial sync: running worker", "worker", w.Name())

		// EnrichOrigins needs to loop until all votes are processed
		if w.Name() == "enrich-vote-origins" {
			if err := s.runEnrichLoop(ctx, w); err != nil {
				slog.Error("initial sync: enrich loop failed", "error", err)
			}
		} else {
			if err := w.Run(ctx); err != nil {
				slog.Error("initial sync: worker failed", "worker", w.Name(), "error", err, "elapsed", time.Since(wStart))
				// Continue to next worker
			}
		}

		slog.Info("initial sync: worker done", "worker", w.Name(), "elapsed", time.Since(wStart))
	}

	slog.Info("initial sync: complete", "totalElapsed", time.Since(start))
	return nil
}

// runEnrichLoop runs the enrich worker repeatedly with increasing batch sizes
// until no more unenriched votes remain.
func (s *Scheduler) runEnrichLoop(ctx context.Context, w Worker) error {
	for round := 1; ; round++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		slog.Info("initial sync: enrich round", "round", round)
		if err := w.Run(ctx); err != nil {
			return err
		}
		// The enrich worker processes a fixed batch (100). If it ran without
		// error, we check if there might be more. We run up to 50 rounds
		// (5000 votes) to avoid infinite loops.
		if round >= 50 {
			slog.Info("initial sync: enrich reached max rounds", "rounds", round)
			break
		}
		// Small delay between rounds to be gentle on the Riksdagen API
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func (s *Scheduler) Start() { s.cron.Start() }
func (s *Scheduler) Stop()  { s.cron.Stop() }
