package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/speeches"
	speechPorts "riksdagskollen/internal/speeches/ports"
)

type SpeechesWorker struct {
	svc     *speeches.Service
	cursors ports.CursorRepository
}

func NewSpeechesWorker(svc *speeches.Service, cursors ports.CursorRepository) SpeechesWorker {
	return SpeechesWorker{svc: svc, cursors: cursors}
}

func (w *SpeechesWorker) Name() string { return "fetch-speeches" }

// SpeechFetchLimit is the largest page /anforandelista will return. The
// endpoint ignores `p`, so a response of exactly this size means the
// party-riksmöte was truncated with no way to fetch the remainder.
const SpeechFetchLimit = 10000

func (w *SpeechesWorker) Run(ctx context.Context) error {
	var since time.Time
	if cur, err := w.cursors.Get(ctx, "speeches"); err == nil && cur != nil && cur.LastDate != nil {
		since = *cur.LastDate
		slog.Info("speeches: incremental sync", "since", since.Format("2006-01-02"))
	}

	// Newest record actually retrieved. The cursor only ever advances to this,
	// never to now(): advancing past a window we failed to read is what left
	// speeches eleven months stale while the worker reported success.
	var newest time.Time

	for _, rm := range MandateRiksmoten {
		for _, party := range ActiveParties {
			ss, err := w.svc.FetchAndStore(ctx, speechPorts.FetchSpeechesFilter{
				Session: rm,
				Party:   party,
				Size:    SpeechFetchLimit,
				Since:   since,
			})
			if err != nil {
				slog.Warn("speeches sync failed, continuing", "rm", rm, "party", party, "error", err)
				continue
			}
			if len(ss) >= SpeechFetchLimit {
				// Storing a silently truncated riksmöte would understate the
				// record with no way to detect the gap downstream.
				return fmt.Errorf(
					"speeches: %s/%s returned the fetch limit (%d); the riksmöte is truncated and cannot be paged",
					rm, party, SpeechFetchLimit)
			}
			for _, sp := range ss {
				if sp.Date.After(newest) {
					newest = sp.Date
				}
			}
			time.Sleep(FetchDelay)
		}
	}

	if newest.IsZero() {
		slog.Info("speeches: nothing fetched, cursor unchanged")
		return nil
	}
	return w.cursors.Upsert(ctx, ports.Cursor{DataType: "speeches", LastDate: &newest})
}
