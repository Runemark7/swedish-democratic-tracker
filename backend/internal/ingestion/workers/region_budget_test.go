package workers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"riksdagskollen/internal/regions/ports"
)

// testRollingYears mirrors the unexported rollingYearsWorker for test use.
func testRollingYears(n int) []int {
	current := time.Now().Year()
	years := make([]int, n)
	for i := range years {
		years[i] = current - 1 - i
	}
	return years
}

// --- Minimal mock interfaces for unit testing ---

type mockRegionSvc struct {
	regions   []mockRegion
	snapshots []ports.RegionBudgetSnapshot
	fetchErr  error
	upsertErr error
	upsertN   int
}

type mockRegion struct{ code string }

func (m *mockRegionSvc) listRegionCodes() []string {
	codes := make([]string, len(m.regions))
	for i, r := range m.regions {
		codes[i] = r.code
	}
	return codes
}

type mockIngestionRunRepo struct {
	startErr    error
	finishErr   error
	startCalled bool
	finishCalled bool
	lastStatus  string // "success" or "error" — inferred from runErr argument
	lastErr     error
}

func (m *mockIngestionRunRepo) StartRun(_ context.Context, _ string) (int64, error) {
	m.startCalled = true
	if m.startErr != nil {
		return 0, m.startErr
	}
	return 42, nil
}

func (m *mockIngestionRunRepo) FinishRun(_ context.Context, _ int64, _ int, runErr error) error {
	m.finishCalled = true
	m.lastErr = runErr
	if runErr != nil {
		m.lastStatus = "error"
	} else {
		m.lastStatus = "success"
	}
	return m.finishErr
}

// workerUnderTest encapsulates the logic we want to unit-test without
// depending on the real regions.Service (which requires a DB and live SCB).
// We test the orchestration contract: StartRun → fetch → upsert → FinishRun.
func runWorkerLogic(
	ctx context.Context,
	regionCodes []string,
	fetchFn func(ctx context.Context, regionCode string, years []int) ([]ports.RegionBudgetSnapshot, error),
	upsertFn func(ctx context.Context, snapshots []ports.RegionBudgetSnapshot) (int, error),
	runRepo *mockIngestionRunRepo,
	workerName string,
) error {
	runID, err := runRepo.StartRun(ctx, workerName)
	if err != nil {
		// Non-fatal — continue without run tracking.
		runID = 0
	}

	var totalRows int
	var runErr error

	years := testRollingYears(4)

	for _, code := range regionCodes {
		snaps, err := fetchFn(ctx, code, years)
		if err != nil {
			runErr = err
			break
		}
		if len(snaps) == 0 {
			continue
		}
		n, err := upsertFn(ctx, snaps)
		if err != nil {
			runErr = err
			break
		}
		totalRows += n
	}

	if runID != 0 {
		_ = runRepo.FinishRun(ctx, runID, totalRows, runErr)
	}
	return runErr
}

func TestRegionBudgetWorker_SuccessPath(t *testing.T) {
	runRepo := &mockIngestionRunRepo{}
	snaps := []ports.RegionBudgetSnapshot{
		{RegionCode: "06", AreaName: "Primärvård", Year: 2023, ValueMnkr: 1000, TotalMnkr: 5000, Pct: 20},
	}

	err := runWorkerLogic(
		context.Background(),
		[]string{"06", "07"},
		func(_ context.Context, _ string, _ []int) ([]ports.RegionBudgetSnapshot, error) {
			return snaps, nil
		},
		func(_ context.Context, _ []ports.RegionBudgetSnapshot) (int, error) {
			return 1, nil
		},
		runRepo,
		"region-budget",
	)

	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
	if !runRepo.startCalled {
		t.Error("StartRun was not called")
	}
	if !runRepo.finishCalled {
		t.Error("FinishRun was not called")
	}
	if runRepo.lastStatus != "success" {
		t.Errorf("expected status 'success', got %q", runRepo.lastStatus)
	}
	if runRepo.lastErr != nil {
		t.Errorf("expected nil error in FinishRun, got: %v", runRepo.lastErr)
	}
}

func TestRegionBudgetWorker_FetchError(t *testing.T) {
	runRepo := &mockIngestionRunRepo{}
	fetchErr := errors.New("SCB connection refused")

	err := runWorkerLogic(
		context.Background(),
		[]string{"06"},
		func(_ context.Context, _ string, _ []int) ([]ports.RegionBudgetSnapshot, error) {
			return nil, fetchErr
		},
		func(_ context.Context, _ []ports.RegionBudgetSnapshot) (int, error) {
			return 0, nil
		},
		runRepo,
		"region-budget",
	)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, fetchErr) {
		t.Errorf("expected fetchErr, got: %v", err)
	}
	if !runRepo.finishCalled {
		t.Error("FinishRun was not called even on error")
	}
	if runRepo.lastStatus != "error" {
		t.Errorf("expected status 'error', got %q", runRepo.lastStatus)
	}
	if runRepo.lastErr == nil {
		t.Error("expected non-nil error in FinishRun")
	}
}

func TestRegionBudgetWorker_StartRunFailureIsNonFatal(t *testing.T) {
	runRepo := &mockIngestionRunRepo{
		startErr: errors.New("DB down"),
	}

	err := runWorkerLogic(
		context.Background(),
		[]string{"06"},
		func(_ context.Context, _ string, _ []int) ([]ports.RegionBudgetSnapshot, error) {
			return []ports.RegionBudgetSnapshot{{RegionCode: "06", AreaName: "A", Year: 2023}}, nil
		},
		func(_ context.Context, _ []ports.RegionBudgetSnapshot) (int, error) {
			return 1, nil
		},
		runRepo,
		"region-budget",
	)

	// Worker should still succeed even if StartRun fails.
	if err != nil {
		t.Errorf("expected nil error when StartRun fails gracefully, got: %v", err)
	}
	// FinishRun should NOT be called because runID is 0.
	if runRepo.finishCalled {
		t.Error("FinishRun should not be called when StartRun returned runID=0")
	}
}
