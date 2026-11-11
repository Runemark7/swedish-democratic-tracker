package workers

import (
	"context"
	"log/slog"
	"time"

	ingPorts "riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/questions"
)

// QuestionsDetailBudget caps the documents read in full per run and kind. The
// next run lists again and skips what is stored, so the worker resumes.
const QuestionsDetailBudget = 500

const questionListPageSize = 200

// QuestionsWorker ingests skriftliga frågor (with their svar) and
// interpellationer for the current riksmöte, then recomputes their motion
// links. A document is keyed on dok_id and read again only when Riksdagen's
// systemdatum differs from the one stored, so it is idempotent. Per riksmöte
// and kind it records how many documents Riksdagen's listing reports.
type QuestionsWorker struct {
	svc     *questions.Service
	periods ingPorts.MandatePeriodReader
	now     func() time.Time
	budget  int
}

func NewQuestionsWorker(svc *questions.Service, periods ingPorts.MandatePeriodReader) QuestionsWorker {
	return QuestionsWorker{svc: svc, periods: periods, budget: QuestionsDetailBudget}
}

func (w *QuestionsWorker) SetNow(now func() time.Time) { w.now = now }

func (w *QuestionsWorker) Name() string { return "fetch-questions" }

func (w *QuestionsWorker) at() time.Time {
	if w.now != nil {
		return w.now()
	}
	return time.Now()
}

// Run takes the current riksmöte only. Older ones are the operator's
// cmd/backfill-questions.
func (w *QuestionsWorker) Run(ctx context.Context) error {
	periods, err := w.periods.ListMandatePeriods(ctx)
	if err != nil {
		return err
	}
	now := w.at().UTC()
	current := CurrentRiksmote(now)
	for _, rm := range motionRiksmoten(periods, now) {
		if rm == current {
			if err := w.SyncRiksmote(ctx, rm, w.budget); err != nil {
				slog.Warn("questions: riksmöte sync failed", "rm", rm, "error", err)
			}
			return nil
		}
	}
	slog.Info("questions: current riksmöte is in no mandate period, nothing to do", "rm", current)
	return nil
}

// SyncRiksmote lists each kind for rm, records the listing's total, reads
// what is new or changed (up to budget per kind) and relinks the riksmöte.
func (w *QuestionsWorker) SyncRiksmote(ctx context.Context, rm string, budget int) error {
	if budget <= 0 {
		budget = QuestionsDetailBudget
	}
	for _, kind := range questions.Kinds {
		if err := w.syncKind(ctx, rm, kind, budget); err != nil {
			return err
		}
	}
	n, err := w.svc.Relink(ctx, rm)
	slog.Info("questions: relinked", "rm", rm, "links", n)
	return err
}

func (w *QuestionsWorker) syncKind(ctx context.Context, rm, kind string, budget int) error {
	stored, err := w.svc.StoredVersions(ctx, kind, rm)
	if err != nil {
		return err
	}
	var todo []string
	total := 0
	for page := 1; ; page++ {
		refs, t, err := w.svc.List(ctx, kind, rm, page, questionListPageSize)
		if err != nil {
			// A half-read listing would understate what exists.
			return err
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
		if len(refs) < questionListPageSize {
			break
		}
		time.Sleep(FetchDelay)
	}
	if total > 0 {
		if err := w.svc.RecordExpected(ctx, rm, kind, total); err != nil {
			return err
		}
	}
	read := 0
	for _, id := range todo {
		if read >= budget {
			break
		}
		read++
		if err := w.svc.FetchAndStore(ctx, kind, id); err != nil {
			slog.Warn("questions: fetch failed", "rm", rm, "kind", kind, "dok_id", id, "error", err)
			continue
		}
		time.Sleep(FetchDelay)
	}
	slog.Info("questions: synced", "rm", rm, "kind", kind, "expected", total, "changed", len(todo), "read", read)
	return nil
}
