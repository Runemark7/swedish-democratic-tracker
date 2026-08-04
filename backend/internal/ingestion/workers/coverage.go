package workers

import (
	"context"
	"log/slog"
	"time"

	ingPorts "riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/votes"
)

// CoverageWorker re-reads the denominator of the site's coverage claim from
// Riksdagen's own listing.
//
// Before this existed, expected_voteringar was written only by the hand-run
// cmd/backfill. The numerator moved daily while the denominator sat still, so
// the published figure drifted toward looking more complete than it was, and
// nothing failed while it did. A coverage claim is the site's central honesty
// claim; it cannot depend on someone remembering to run a command.
type CoverageWorker struct {
	svc      *votes.Service
	coverage ingPorts.CoverageRepository
}

func NewCoverageWorker(svc *votes.Service, coverage ingPorts.CoverageRepository) CoverageWorker {
	return CoverageWorker{svc: svc, coverage: coverage}
}

func (w *CoverageWorker) Name() string { return "refresh-coverage" }

func (w *CoverageWorker) Run(ctx context.Context) error {
	riksmoten, err := w.coverage.RiksmotenNeedingDenominator(ctx)
	if err != nil {
		return err
	}
	if len(riksmoten) == 0 {
		slog.Info("coverage: no riksmöte needs a denominator")
		return nil
	}

	for _, rm := range riksmoten {
		d, err := w.enumerate(ctx, rm)
		if err != nil {
			// Keep the denominator we already have. A half-finished enumeration
			// would understate what exists upstream, which reads as the record
			// being more complete than it is — the exact failure this worker
			// exists to prevent. One riksmöte failing must not stop the rest.
			slog.Warn("coverage: enumeration failed, keeping previous denominator",
				"rm", rm, "error", err)
			continue
		}
		if d.Expected == 0 {
			// Riksdagen reporting zero for a riksmöte we believe has votes is a
			// bad response, not news. Writing it would zero a good denominator.
			slog.Warn("coverage: upstream reported zero voteringar, not recording", "rm", rm)
			continue
		}
		if err := w.coverage.UpsertDenominator(ctx, d); err != nil {
			slog.Warn("coverage: upsert failed", "rm", rm, "error", err)
			continue
		}
		slog.Info("coverage: denominator refreshed", "rm", rm,
			"expected", d.Expected, "unreachable", d.Unreachable)
	}
	return nil
}

// enumerate walks Riksdagen's listing for one riksmöte.
//
// The count comes from @traffar, which the listing returns alongside the first
// page, but the pages are still walked: unreachable voteringar and the newest
// decision date can only be seen by looking at the records themselves.
func (w *CoverageWorker) enumerate(ctx context.Context, rm string) (ingPorts.Denominator, error) {
	d := ingPorts.Denominator{Riksmote: rm}
	var lastDecision time.Time

	for page := 1; ; page++ {
		refs, total, err := w.svc.ListVoteringar(ctx, rm, page, voteringPageSize)
		if err != nil {
			return ingPorts.Denominator{}, err
		}
		if total > 0 {
			d.Expected = total
		}
		for _, r := range refs {
			if r.Beteckning == "" {
				// A votering on a motion rather than a betänkande: there is no
				// beteckning to partition by, so this fetch strategy cannot
				// reach it. Counted so a shortfall is explained.
				d.Unreachable++
				continue
			}
			if t, err := time.Parse("2006-01-02", r.Date); err == nil && t.After(lastDecision) {
				lastDecision = t
			}
		}
		if len(refs) < voteringPageSize {
			break
		}
		time.Sleep(FetchDelay)
	}

	if !lastDecision.IsZero() {
		d.LastVoteDate = &lastDecision
	}
	return d, nil
}
