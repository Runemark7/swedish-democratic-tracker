package workers

import (
	"context"
	"log/slog"
	"time"

	"riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/speeches"
	speechPorts "riksdagskollen/internal/speeches/ports"
)

const currentSession = "2024/25"

type SpeechesWorker struct {
	svc     *speeches.Service
	cursors ports.CursorRepository
}

func NewSpeechesWorker(svc *speeches.Service, cursors ports.CursorRepository) SpeechesWorker {
	return SpeechesWorker{svc: svc, cursors: cursors}
}

func (w *SpeechesWorker) Name() string { return "fetch-speeches" }

func (w *SpeechesWorker) Run(ctx context.Context) error {
	// Read cursor for incremental sync
	var since time.Time
	if cur, err := w.cursors.Get(ctx, "speeches"); err == nil && cur != nil && cur.LastDate != nil {
		since = *cur.LastDate
		slog.Info("speeches: incremental sync", "since", since.Format("2006-01-02"))
	}

	var latestDate time.Time
	for i, party := range ActiveParties {
		f := speechPorts.FetchSpeechesFilter{
			Session: currentSession,
			Party:   party,
			Size:    200,
			Since:   since,
		}
		if err := w.svc.SyncSpeeches(ctx, f); err != nil {
			slog.Warn("speeches sync failed for party, continuing", "party", party, "error", err)
		}
		if i < len(ActiveParties)-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Update cursor to today (we fetched everything up to now)
	if latestDate.IsZero() {
		now := time.Now().UTC().Truncate(24 * time.Hour)
		latestDate = now
	}
	if err := w.cursors.Upsert(ctx, ports.Cursor{
		DataType: "speeches",
		LastDate: &latestDate,
	}); err != nil {
		slog.Warn("speeches: failed to update cursor", "error", err)
	}
	return nil
}
