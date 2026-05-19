package workers

import (
	"context"
	"log/slog"
	"time"

	ingPorts "riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/regions"
)

// rollingYearsWorker returns the last n completed calendar years (excluding the current year).
// Copied from regions.rollingYears (unexported) for use in the worker package.
func rollingYearsWorker(n int) []int {
	current := time.Now().Year()
	years := make([]int, n)
	for i := range years {
		years[i] = current - 1 - i
	}
	return years
}

// RegionBudgetWorker fetches budget snapshots from SCB for all regions across
// the last 4 completed years and upserts them into region_budget_snapshots.
type RegionBudgetWorker struct {
	regionSvc *regions.Service
	runRepo   ingPorts.IngestionRunRepository
}

func NewRegionBudgetWorker(regionSvc *regions.Service, runRepo ingPorts.IngestionRunRepository) RegionBudgetWorker {
	return RegionBudgetWorker{regionSvc: regionSvc, runRepo: runRepo}
}

func (w *RegionBudgetWorker) Name() string { return "region-budget" }

func (w *RegionBudgetWorker) Run(ctx context.Context) error {
	runID, err := w.runRepo.StartRun(ctx, w.Name())
	if err != nil {
		slog.Error("region-budget: failed to start run record", "error", err)
		// Non-fatal — continue without run tracking rather than aborting.
	}

	totalRows, runErr := w.doRun(ctx)

	if runID != 0 {
		if ferr := w.runRepo.FinishRun(ctx, runID, totalRows, runErr); ferr != nil {
			slog.Warn("region-budget: failed to finish run record", "error", ferr)
		}
	}
	return runErr
}

func (w *RegionBudgetWorker) doRun(ctx context.Context) (int, error) {
	regions, err := w.regionSvc.ListRegions(ctx)
	if err != nil {
		return 0, err
	}

	years := rollingYearsWorker(4)
	var totalRows int

	for _, reg := range regions {
		if ctx.Err() != nil {
			return totalRows, ctx.Err()
		}

		snapshots, err := w.regionSvc.GetRegionBudgetMultiYear(ctx, reg.Code, years)
		if err != nil {
			slog.Warn("region-budget: fetch failed for region", "region", reg.Code, "error", err)
			continue
		}

		if len(snapshots) == 0 {
			continue
		}

		n, err := w.regionSvc.UpsertRegionBudgetSnapshots(ctx, snapshots)
		if err != nil {
			return totalRows, err
		}
		totalRows += n

		slog.Info("region-budget: upserted snapshots", "region", reg.Code, "rows", n)
	}

	return totalRows, nil
}

// compile-time interface check
var _ interface {
	Name() string
	Run(ctx context.Context) error
} = (*RegionBudgetWorker)(nil)
