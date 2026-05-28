package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"riksdagskollen/internal/ministers"
)

var activeDeptCodes = []string{"SB", "Fi", "Ju", "UD", "Fö", "S", "A", "U", "Ku", "N", "LI"}

type ProposalsWorker struct {
	svc *ministers.Service
}

func NewProposalsWorker(svc *ministers.Service) ProposalsWorker {
	return ProposalsWorker{svc: svc}
}

func (w *ProposalsWorker) Name() string { return "sync-proposals" }

func (w *ProposalsWorker) Run(ctx context.Context) error {
	currentYear := time.Now().Year()
	years := []string{
		fmt.Sprintf("%d/%02d", currentYear-1, currentYear%100),
		fmt.Sprintf("%d/%02d", currentYear-2, (currentYear-1)%100),
	}

	for i, dept := range activeDeptCodes {
		if err := w.svc.SyncProposals(ctx, dept, years); err != nil {
			slog.Warn("proposals sync failed for dept", "dept", dept, "error", err)
		}
		if i < len(activeDeptCodes)-1 {
			time.Sleep(300 * time.Millisecond)
		}
	}
	return nil
}
