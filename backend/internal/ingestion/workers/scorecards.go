package workers

import (
	"context"

	"riksdagskollen/internal/matching"
)

type RefreshScorecardsWorker struct {
	svc *matching.Service
}

func NewRefreshScorecardsWorker(svc *matching.Service) RefreshScorecardsWorker {
	return RefreshScorecardsWorker{svc: svc}
}

func (w *RefreshScorecardsWorker) Name() string { return "refresh-scorecards" }

func (w *RefreshScorecardsWorker) Run(ctx context.Context) error {
	return w.svc.RefreshScorecards(ctx)
}
