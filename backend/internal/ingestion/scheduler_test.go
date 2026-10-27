package ingestion

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type seqWorker struct {
	name string
	fn   func(ctx context.Context) error
}

func (w seqWorker) Name() string { return w.name }
func (w seqWorker) Run(ctx context.Context) error {
	if w.fn == nil {
		return nil
	}
	return w.fn(ctx)
}

type fakeRuns struct {
	mu       sync.Mutex
	started  []string
	finished []error
}

func (f *fakeRuns) StartRun(_ context.Context, name string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.started = append(f.started, name)
	return int64(len(f.started)), nil
}

func (f *fakeRuns) FinishRun(_ context.Context, _ int64, _ int, runErr error) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.finished = append(f.finished, runErr)
	return nil
}

func TestRunDue_OrderAndSundayOnlyWeeklies(t *testing.T) {
	monday := time.Date(2026, 10, 26, 1, 0, 0, 0, time.UTC)
	if monday.Weekday() != time.Monday {
		t.Fatalf("fixture weekday = %s, want Monday", monday.Weekday())
	}
	var order []string
	s := NewScheduler()
	s.now = func() time.Time { return monday }
	record := func(name string) seqWorker {
		return seqWorker{name: name, fn: func(context.Context) error {
			order = append(order, name)
			return nil
		}}
	}
	for _, spec := range []struct{ spec, name string }{
		{"@daily", "fetch-votes"},
		{"@daily", "keyword-matcher"},
		{"@weekly", "refresh-scorecards"},
		{"@daily", "refresh-coverage"},
	} {
		if err := s.Register(spec.spec, record(spec.name)); err != nil {
			t.Fatal(err)
		}
	}

	s.RunDue(context.Background())
	want := []string{"fetch-votes", "keyword-matcher", "refresh-coverage"}
	if !sameStrings(order, want) {
		t.Fatalf("Monday order = %v, want %v", order, want)
	}

	order = nil
	sunday := time.Date(2026, 10, 25, 1, 0, 0, 0, time.UTC)
	if sunday.Weekday() != time.Sunday {
		t.Fatalf("fixture weekday = %s, want Sunday", sunday.Weekday())
	}
	s.now = func() time.Time { return sunday }
	s.RunDue(context.Background())
	want = []string{"fetch-votes", "keyword-matcher", "refresh-scorecards", "refresh-coverage"}
	if !sameStrings(order, want) {
		t.Fatalf("Sunday order = %v, want %v", order, want)
	}
}

func TestRunDue_PanicIsContainedAndRecorded(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic escaped the scheduler: %v", r)
		}
	}()
	var order []string
	runs := &fakeRuns{}
	s := NewScheduler()
	s.runs = runs
	s.now = func() time.Time { return time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC) }
	_ = s.Register("@daily", seqWorker{name: "boom", fn: func(context.Context) error {
		panic("boom")
	}})
	_ = s.Register("@daily", seqWorker{name: "after", fn: func(context.Context) error {
		order = append(order, "after")
		return nil
	}})

	s.RunDue(context.Background())
	if !sameStrings(order, []string{"after"}) {
		t.Fatalf("order = %v, want the worker after the panic to run", order)
	}
	runs.mu.Lock()
	defer runs.mu.Unlock()
	if len(runs.finished) != 2 {
		t.Fatalf("finished %d runs, want 2", len(runs.finished))
	}
	if runs.finished[0] == nil {
		t.Error("panic was recorded as a successful run")
	}
	if runs.finished[1] != nil {
		t.Errorf("worker after the panic finished with %v", runs.finished[1])
	}
}

func TestRunDue_TimeoutDoesNotBlockTheChain(t *testing.T) {
	var order []string
	s := NewScheduler()
	s.jobTimeout = 20 * time.Millisecond
	s.now = func() time.Time { return time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC) }
	_ = s.Register("@daily", seqWorker{name: "slow", fn: func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}})
	_ = s.Register("@daily", seqWorker{name: "next", fn: func(context.Context) error {
		order = append(order, "next")
		return nil
	}})

	done := make(chan struct{})
	go func() {
		s.RunDue(context.Background())
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("chain did not return after the worker timeout")
	}
	if !sameStrings(order, []string{"next"}) {
		t.Fatalf("order = %v, want the next worker to run", order)
	}
}

func TestRunDue_SkipsWhenAlreadyRunning(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var runs int
	var mu sync.Mutex
	s := NewScheduler()
	s.now = func() time.Time { return time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC) }
	_ = s.Register("@daily", seqWorker{name: "block", fn: func(context.Context) error {
		mu.Lock()
		runs++
		mu.Unlock()
		close(started)
		<-release
		return nil
	}})

	go s.RunDue(context.Background())
	<-started

	skipped := make(chan struct{})
	go func() {
		s.RunDue(context.Background())
		close(skipped)
	}()
	select {
	case <-skipped:
	case <-time.After(time.Second):
		t.Fatal("overlapping run was not skipped")
	}
	close(release)

	deadline := time.After(time.Second)
	for {
		mu.Lock()
		n := runs
		mu.Unlock()
		if n == 1 {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("worker ran %d times, want 1", n)
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestRunInitialSync_RunsWeeklyWorkersOnAWeekday(t *testing.T) {
	var order []string
	s := NewScheduler()
	s.now = func() time.Time { return time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC) } // Monday
	_ = s.Register("@weekly", seqWorker{name: "weekly", fn: func(context.Context) error {
		order = append(order, "weekly")
		return nil
	}})
	if err := s.RunInitialSync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !sameStrings(order, []string{"weekly"}) {
		t.Fatalf("initial sync order = %v, want the weekly worker", order)
	}
}

func TestRegister_RejectsUnknownSchedule(t *testing.T) {
	s := NewScheduler()
	err := s.Register("0 0 * * *", seqWorker{name: "x"})
	if err == nil {
		t.Fatal("expected an error for a raw cron spec")
	}
}

func TestRunDue_WorkerErrorDoesNotStopTheChain(t *testing.T) {
	var order []string
	s := NewScheduler()
	s.now = func() time.Time { return time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC) }
	_ = s.Register("@daily", seqWorker{name: "bad", fn: func(context.Context) error {
		return errors.New("upstream")
	}})
	_ = s.Register("@daily", seqWorker{name: "next", fn: func(context.Context) error {
		order = append(order, "next")
		return nil
	}})
	s.RunDue(context.Background())
	if !sameStrings(order, []string{"next"}) {
		t.Fatalf("order = %v", order)
	}
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
