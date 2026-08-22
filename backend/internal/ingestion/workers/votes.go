package workers

import (
	"context"
	"fmt"
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

// MandateRiksmoten is the 2022-2026 mandate period: four riksmöten, 2 562
// voteringar in total.
//
// This is the daily speeches worker's scope, not the site's. The record holds
// six mandate periods (mandate_periods), and cmd/backfill reads that table
// rather than this list. Widening this to all six would put 24 years of
// speeches behind a cron job that runs every day, so it stays at the current
// period until a historical speech backfill is deliberately built.
var MandateRiksmoten = []string{"2022/23", "2023/24", "2024/25", "2025/26"}

const voteringPageSize = 200

// FetchDelay throttles requests to the Riksdagen API. Exported so tests can set
// it to zero: against fakes the sleep only slows the suite down.
var FetchDelay = 200 * time.Millisecond

// CurrentRiksmote returns the riksmöte label for a date. A riksmöte runs
// September to September, so a date before September belongs to the one that
// opened the previous year.
func CurrentRiksmote(t time.Time) string {
	y := t.Year()
	if t.Month() < time.September {
		y--
	}
	return fmt.Sprintf("%d/%02d", y, (y+1)%100)
}

func (w *VotesWorker) Run(ctx context.Context) error {
	var since time.Time
	if cur, err := w.cursors.Get(ctx, "votes"); err == nil && cur != nil && cur.LastDate != nil {
		since = *cur.LastDate
	}
	rm := CurrentRiksmote(time.Now().UTC())
	slog.Info("votes: incremental sync", "riksmote", rm, "since", since.Format("2006-01-02"))

	// Enumerate newest-first and stop at the cursor. /dokumentlista paginates
	// correctly; /voteringlista does not, which is why the work is partitioned
	// by beteckning rather than paged.
	seen := map[string]bool{}
	var order []string
	stop := false
	for page := 1; !stop; page++ {
		refs, _, err := w.svc.ListVoteringar(ctx, rm, page, voteringPageSize)
		if err != nil {
			// Leave the cursor untouched: advancing past a window we failed to
			// read is what silently lost months of votes.
			slog.Warn("votes: enumeration failed", "rm", rm, "page", page, "error", err)
			return nil
		}
		if len(refs) == 0 {
			break
		}
		for _, r := range refs {
			if !since.IsZero() && !r.SystemDatum.After(since) {
				stop = true
				break
			}
			if r.Beteckning != "" && !seen[r.Beteckning] {
				seen[r.Beteckning] = true
				order = append(order, r.Beteckning)
			}
		}
		if len(refs) < voteringPageSize {
			break
		}
	}

	// One request per betänkande returns every party's ballots for it,
	// including party-less members that a per-party loop would miss.
	var newest time.Time
	for _, bet := range order {
		vv, err := w.svc.FetchAndStore(ctx, ports.FetchVotesFilter{
			Session: rm, Beteckning: bet, Since: since,
		})
		if err != nil {
			slog.Warn("votes: fetch failed", "rm", rm, "bet", bet, "error", err)
			continue
		}
		for _, v := range vv {
			if v.SystemDatum.After(newest) {
				newest = v.SystemDatum
			}
		}
		time.Sleep(FetchDelay)
	}

	if newest.IsZero() {
		slog.Info("votes: nothing fetched, cursor unchanged", "betankanden", len(order))
		return nil
	}
	return w.cursors.Upsert(ctx, ingPorts.Cursor{DataType: "votes", LastDate: &newest})
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
