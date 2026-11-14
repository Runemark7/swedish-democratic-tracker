package workers

import (
	"context"
	"log/slog"
	"time"

	"riksdagskollen/internal/motions"
)

// BetankandenBudget caps the betänkanden read per run. Each costs two
// requests; the first runs after deploy fill the record over days. The next
// run asks for what is still missing, so the worker resumes where it stopped.
const BetankandenBudget = 25

// BetankandenWorker reads the betänkanden that stored motions are handled in,
// with their förslagspunkter, and links each yrkande to the punkt(er) that
// decided it. It also reads the decided betänkanden of the stored listing that
// no motion points at (a vilande grundlagsförslag, say). Run daily it reads
// only those the current riksmöte's motions point at and that riksmöte's
// listed ones; older riksmöten are the operator's backfill
// (cmd/backfill-motions). It is idempotent: a betänkande is read again only
// when a motion pointing at it was read after it or it was read before its
// beslutsdag was over, and links are replaced, not appended.
type BetankandenWorker struct {
	svc    *motions.Service
	budget int
	now    func() time.Time
}

func NewBetankandenWorker(svc *motions.Service) BetankandenWorker {
	return BetankandenWorker{svc: svc, budget: BetankandenBudget}
}

// SetNow pins the clock that decides the current riksmöte. Set it before registration.
func (w *BetankandenWorker) SetNow(now func() time.Time) { w.now = now }

func (w *BetankandenWorker) Name() string { return "fetch-betankanden" }

func (w *BetankandenWorker) Run(ctx context.Context) error {
	budget := w.budget
	if budget <= 0 {
		budget = BetankandenBudget
	}
	now := time.Now
	if w.now != nil {
		now = w.now
	}
	refs, err := w.svc.Pending(ctx, CurrentRiksmote(now().UTC()), budget)
	if err != nil {
		return err
	}
	read := 0
	for _, ref := range refs {
		if ctx.Err() != nil {
			break
		}
		found, err := w.svc.SyncBetankande(ctx, ref)
		if err != nil {
			slog.Warn("betankanden: sync failed", "rm", ref.Riksmote, "bet", ref.Beteckning, "error", err)
			continue
		}
		if found {
			read++
		}
		time.Sleep(FetchDelay)
	}
	slog.Info("betankanden: run done", "pending", len(refs), "read", read)
	return nil
}
