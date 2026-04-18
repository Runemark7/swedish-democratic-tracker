package workers

import (
	"context"
	"log/slog"
	"time"

	"riksdagskollen/internal/politicians"
)

var ActiveParties = []string{"S", "M", "SD", "C", "V", "KD", "L", "MP"}

type PoliticiansWorker struct {
	svc *politicians.Service
}

func NewPoliticiansWorker(svc *politicians.Service) PoliticiansWorker {
	return PoliticiansWorker{svc: svc}
}

func (w *PoliticiansWorker) Name() string { return "fetch-politicians" }

func (w *PoliticiansWorker) Run(ctx context.Context) error {
	for i, party := range ActiveParties {
		if err := w.svc.SyncParty(ctx, party); err != nil {
			slog.Warn("failed to sync party politicians, continuing", "party", party, "error", err)
			continue
		}
		if i < len(ActiveParties)-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}
	return nil
}
