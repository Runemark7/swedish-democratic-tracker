package workers

import (
	"context"

	"riksdagskollen/internal/politicians"
)

// ActiveParties is the named-party list used by the speeches and votes workers
// for per-party aggregation. The politicians worker no longer iterates it (see
// Run) because it excludes party-less members.
var ActiveParties = []string{"S", "M", "SD", "C", "V", "KD", "L", "MP"}

type PoliticiansWorker struct {
	svc *politicians.Service
}

func NewPoliticiansWorker(svc *politicians.Service) PoliticiansWorker {
	return PoliticiansWorker{svc: svc}
}

func (w *PoliticiansWorker) Name() string { return "fetch-politicians" }

func (w *PoliticiansWorker) Run(ctx context.Context) error {
	// Sync the full roster in one call (empty party filter) instead of looping
	// the named parties — that loop missed party-less members ("-", politiska
	// vildar), leaving the active roster 9 short of the 349 seats.
	return w.svc.SyncAll(ctx)
}
