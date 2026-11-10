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
// decided it. It is idempotent: a betänkande is read again only when a motion
// pointing at it was read after it, and links are replaced, not appended.
type BetankandenWorker struct {
	svc    *motions.Service
	budget int
}

func NewBetankandenWorker(svc *motions.Service) BetankandenWorker {
	return BetankandenWorker{svc: svc, budget: BetankandenBudget}
}

func (w *BetankandenWorker) Name() string { return "fetch-betankanden" }

func (w *BetankandenWorker) Run(ctx context.Context) error {
	budget := w.budget
	if budget <= 0 {
		budget = BetankandenBudget
	}
	refs, err := w.svc.Pending(ctx, budget)
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
