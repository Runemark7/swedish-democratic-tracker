package workers

import (
	"context"
	"log/slog"

	ingPorts "riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/regions"
)

// MunicipalityBudgetWorker fetches Kolada spending KPIs for all tracked municipalities
// across the last 4 completed years, converts kr/inv × population to mnkr snapshots,
// and upserts them into municipality_budget_snapshots.
type MunicipalityBudgetWorker struct {
	regionSvc *regions.Service
	runRepo   ingPorts.IngestionRunRepository
}

func NewMunicipalityBudgetWorker(regionSvc *regions.Service, runRepo ingPorts.IngestionRunRepository) MunicipalityBudgetWorker {
	return MunicipalityBudgetWorker{regionSvc: regionSvc, runRepo: runRepo}
}

func (w *MunicipalityBudgetWorker) Name() string { return "municipality-budget" }

func (w *MunicipalityBudgetWorker) Run(ctx context.Context) error {
	runID, err := w.runRepo.StartRun(ctx, w.Name())
	if err != nil {
		slog.Error("municipality-budget: failed to start run record", "error", err)
	}

	totalRows, runErr := w.doRun(ctx)

	if runID != 0 {
		if ferr := w.runRepo.FinishRun(ctx, runID, totalRows, runErr); ferr != nil {
			slog.Warn("municipality-budget: failed to finish run record", "error", ferr)
		}
	}
	return runErr
}

func (w *MunicipalityBudgetWorker) doRun(ctx context.Context) (int, error) {
	// List all municipalities (empty regionCode = all).
	muns, err := w.regionSvc.ListMunicipalities(ctx, "")
	if err != nil {
		return 0, err
	}

	years := rollingYearsWorker(4)
	var totalRows int

	for _, mun := range muns {
		if ctx.Err() != nil {
			return totalRows, ctx.Err()
		}

		snapshots, err := w.regionSvc.GetMunicipalityBudgetMultiYear(ctx, mun.Code, mun.Population, years)
		if err != nil {
			slog.Warn("municipality-budget: fetch failed", "mun", mun.Code, "error", err)
			continue
		}
		if len(snapshots) == 0 {
			continue
		}

		n, err := w.regionSvc.UpsertMunicipalityBudgetSnapshots(ctx, snapshots)
		if err != nil {
			return totalRows, err
		}
		totalRows += n
		slog.Info("municipality-budget: upserted snapshots", "mun", mun.Code, "rows", n)
	}

	return totalRows, nil
}

var _ interface {
	Name() string
	Run(ctx context.Context) error
} = (*MunicipalityBudgetWorker)(nil)
