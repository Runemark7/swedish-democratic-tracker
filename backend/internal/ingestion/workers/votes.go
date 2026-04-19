package workers

import (
	"context"
	"log/slog"
	"time"

	ingPorts "riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/votes"
	"riksdagskollen/internal/votes/ports"
)

type VotesWorker struct {
	svc     *votes.Service
	cursors ingPorts.CursorRepository
}

func NewVotesWorker(svc *votes.Service, cursors ingPorts.CursorRepository) VotesWorker {
	return VotesWorker{svc: svc, cursors: cursors}
}

func (w *VotesWorker) Name() string { return "fetch-votes" }

func (w *VotesWorker) Run(ctx context.Context) error {
	// Read cursor for incremental sync
	var since time.Time
	if cur, err := w.cursors.Get(ctx, "votes"); err == nil && cur != nil && cur.LastDate != nil {
		since = *cur.LastDate
		slog.Info("votes: incremental sync", "since", since.Format("2006-01-02"))
	}

	for i, party := range ActiveParties {
		f := ports.FetchVotesFilter{
			Session: currentSession,
			Party:   party,
			Size:    500,
			Since:   since,
		}
		if err := w.svc.SyncVotes(ctx, f); err != nil {
			slog.Warn("votes sync failed for party, continuing", "party", party, "error", err)
		}
		if i < len(ActiveParties)-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Update cursor to today
	now := time.Now().UTC().Truncate(24 * time.Hour)
	if err := w.cursors.Upsert(ctx, ingPorts.Cursor{
		DataType: "votes",
		LastDate: &now,
	}); err != nil {
		slog.Warn("votes: failed to update cursor", "error", err)
	}
	return nil
}

type EnrichOriginsWorker struct {
	svc *votes.Service
}

func NewEnrichOriginsWorker(svc *votes.Service) EnrichOriginsWorker {
	return EnrichOriginsWorker{svc: svc}
}

func (w *EnrichOriginsWorker) Name() string { return "enrich-vote-origins" }

func (w *EnrichOriginsWorker) Run(ctx context.Context) error {
	_, err := w.svc.EnrichOrigins(ctx, 100)
	return err
}

// RunBatch returns the number of votes enriched so the loop can stop early.
func (w *EnrichOriginsWorker) RunBatch(ctx context.Context) (int, error) {
	return w.svc.EnrichOrigins(ctx, 100)
}
