package workers

import (
	"context"
	"log/slog"
	"time"

	ingPorts "riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/motions"
)

// BeslutBackfillPerRun caps how many riksmöten never listed before are read
// in one run. Each costs a few requests; older history fills over a few days.
const BeslutBackfillPerRun = 3

// BeslutListWorker stores Riksdagen's betänkande listing (status, beslutsdag,
// beslutad) for the Beslut list. The current and the previous riksmöte are
// re-read on every run, because planned betänkanden change; older riksmöten
// are read once, when nothing is stored for them. Writes are upserts, so a run
// that stops halfway resumes on the next one and rerunning changes nothing.
type BeslutListWorker struct {
	svc     *motions.Service
	periods ingPorts.MandatePeriodReader
	now     func() time.Time
}

func NewBeslutListWorker(svc *motions.Service, periods ingPorts.MandatePeriodReader) BeslutListWorker {
	return BeslutListWorker{svc: svc, periods: periods}
}

// SetNow pins the clock; set it before registration.
func (w *BeslutListWorker) SetNow(now func() time.Time) { w.now = now }

func (w *BeslutListWorker) Name() string { return "sync-beslut-list" }

func (w *BeslutListWorker) Run(ctx context.Context) error {
	now := time.Now()
	if w.now != nil {
		now = w.now()
	}
	periods, err := w.periods.ListMandatePeriods(ctx)
	if err != nil {
		return err
	}
	now = now.UTC()
	current := CurrentRiksmote(now)
	rms := motionRiksmoten(periods, now) // newest first
	backfill := 0
	for i, rm := range rms {
		if ctx.Err() != nil {
			break
		}
		if rm != current && i > 1 {
			n, err := w.svc.BeslutListed(ctx, rm)
			if err != nil {
				return err
			}
			if n > 0 {
				continue
			}
			if backfill >= BeslutBackfillPerRun {
				continue
			}
			backfill++
		}
		n, err := w.svc.SyncBeslutList(ctx, rm, FetchDelay)
		if err != nil {
			slog.Warn("beslut list: riksmöte sync failed", "rm", rm, "error", err)
			continue
		}
		slog.Info("beslut list: riksmöte synced", "rm", rm, "rows", n)
		time.Sleep(FetchDelay)
	}
	return nil
}
