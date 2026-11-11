package workers

import (
	"context"
	"log/slog"
	"sort"
	"time"

	ingPorts "riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/motions"
)

// MotionsDetailBudget caps the motions read in full per run. A riksmöte holds
// thousands, so the first runs after deploy fill the record over days instead
// of in one burst. The next run re-lists and skips what is stored, so the
// worker resumes where it stopped.
const MotionsDetailBudget = 500

const motionListPageSize = 200

// MotionsWorker ingests motions per riksmöte. The riksmöten come from the
// mandate_periods table, newest first, so a new parliament is a row, not a
// code change.
//
// It is idempotent: a motion is keyed on dok_id and is read again only when
// Riksdagen's systemdatum differs from the one stored. It records, per
// riksmöte and source, how many motions Riksdagen's listing reports.
type MotionsWorker struct {
	svc     *motions.Service
	periods ingPorts.MandatePeriodReader
	now     func() time.Time
	budget  int
}

func NewMotionsWorker(svc *motions.Service, periods ingPorts.MandatePeriodReader) MotionsWorker {
	return MotionsWorker{svc: svc, periods: periods, budget: MotionsDetailBudget}
}

// SetNow pins the clock used to decide which riksmöten have opened. Set it
// before registration; the scheduler keeps a copy of the worker.
func (w *MotionsWorker) SetNow(now func() time.Time) { w.now = now }

func (w *MotionsWorker) Name() string { return "fetch-motions" }

func (w *MotionsWorker) at() time.Time {
	if w.now != nil {
		return w.now()
	}
	return time.Now()
}

// motionRiksmoten lists the riksmöten of every declared mandate period that
// have opened, newest first.
func motionRiksmoten(periods []ingPorts.MandatePeriod, now time.Time) []string {
	seen := map[string]bool{}
	var all []string
	for _, p := range periods {
		for _, rm := range p.Riksmoten {
			if !seen[rm] {
				seen[rm] = true
				all = append(all, rm)
			}
		}
	}
	out := OpenedRiksmoten(all, now)
	sort.Slice(out, func(i, j int) bool {
		yi, oki := riksmoteStartYear(out[i])
		yj, okj := riksmoteStartYear(out[j])
		if oki && okj && yi != yj {
			return yi > yj
		}
		return out[i] > out[j]
	})
	return out
}

// Run ingests the CURRENT riksmöte only, incrementally and within the detail
// budget. Older riksmöten are the operator's backfill (cmd/backfill-motions):
// nothing in the daily chain walks the history.
func (w *MotionsWorker) Run(ctx context.Context) error {
	periods, err := w.periods.ListMandatePeriods(ctx)
	if err != nil {
		return err
	}
	now := w.at().UTC()
	current := CurrentRiksmote(now)
	found := false
	for _, rm := range motionRiksmoten(periods, now) {
		if rm == current {
			found = true
		}
	}
	if !found {
		slog.Info("motions: current riksmöte is in no mandate period, nothing to do", "rm", current)
		return nil
	}
	budget := w.budget
	if budget <= 0 {
		budget = MotionsDetailBudget
	}
	if _, err := w.syncRiksmote(ctx, current, budget); err != nil {
		slog.Warn("motions: riksmöte sync failed", "rm", current, "error", err)
	}
	return nil
}

// SyncRiksmote lists one riksmöte, records the listing's total and reads in
// full each motion that is new or changed, up to budget. It returns the number
// read. The backfill uses it with a budget large enough for the whole riksmöte.
func (w *MotionsWorker) SyncRiksmote(ctx context.Context, rm string, budget int) (int, error) {
	return w.syncRiksmote(ctx, rm, budget)
}

// syncRiksmote lists the riksmöte, records the listing's total, and reads in
// full each motion that is new or changed, up to budget. It returns the
// number read.
func (w *MotionsWorker) syncRiksmote(ctx context.Context, rm string, budget int) (int, error) {
	stored, err := w.svc.StoredVersions(ctx, rm)
	if err != nil {
		return 0, err
	}
	var todo []string
	total := 0
	for page := 1; ; page++ {
		refs, t, err := w.svc.ListMotions(ctx, rm, page, motionListPageSize)
		if err != nil {
			// A half-read listing would understate what exists.
			return 0, err
		}
		if t > 0 {
			total = t
		}
		for _, r := range refs {
			if have, ok := stored[r.DokID]; ok && !r.SystemDatum.IsZero() && have.Equal(r.SystemDatum) {
				continue
			}
			todo = append(todo, r.DokID)
		}
		if len(refs) < motionListPageSize {
			break
		}
		time.Sleep(FetchDelay)
	}
	if total > 0 {
		if err := w.svc.RecordExpected(ctx, rm, total); err != nil {
			return 0, err
		}
	}

	read := 0
	for _, id := range todo {
		if read >= budget {
			break
		}
		if err := w.svc.FetchAndStore(ctx, id); err != nil {
			slog.Warn("motions: fetch failed", "rm", rm, "dok_id", id, "error", err)
			read++
			continue
		}
		read++
		time.Sleep(FetchDelay)
	}
	slog.Info("motions: riksmöte synced", "rm", rm, "expected", total, "changed", len(todo), "read", read)
	return read, nil
}
