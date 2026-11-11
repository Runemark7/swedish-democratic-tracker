package motions

import (
	"context"
	"log/slog"
	"time"

	"riksdagskollen/internal/motions/domain"
	"riksdagskollen/internal/motions/ports"
	votesports "riksdagskollen/internal/votes/ports"
)

// BeslutPageSize is how many listing rows one request to Riksdagen asks for.
const BeslutPageSize = 200

// BeslutItem is one row of the Beslut list.
type BeslutItem struct {
	domain.BeslutRow
	Decided bool
	Outcome string
	Parties []domain.PartyLineSummary
}

// BeslutResult is a page of the Beslut list.
type BeslutResult struct {
	Items []BeslutItem
	Next  *domain.BeslutCursor
}

// SyncBeslutList reads Riksdagen's betänkande listing for one riksmöte and
// stores it. Every page is read before anything is written, so a failing page
// leaves the stored listing as it was. Running it twice leaves the same rows.
func (s *Service) SyncBeslutList(ctx context.Context, rm string, pause time.Duration) (int, error) {
	var all []ports.BetankandeRef
	for page := 1; ; page++ {
		refs, _, err := s.rd.ListBetankanden(ctx, rm, page, BeslutPageSize)
		if err != nil {
			return 0, err
		}
		all = append(all, refs...)
		if len(refs) < BeslutPageSize {
			break
		}
		time.Sleep(pause)
	}
	if err := s.repo.UpsertBeslutListing(ctx, all); err != nil {
		return 0, err
	}
	return len(all), nil
}

// BeslutListed reports how many listing rows a riksmöte holds.
func (s *Service) BeslutListed(ctx context.Context, rm string) (int, error) {
	return s.repo.BeslutListed(ctx, rm)
}

// ListBeslut reads one page. Whether a row is decided is
// RiksdagDocument.Decided over Riksdagen's stored status, beslutsdag and
// beslutad.
func (s *Service) ListBeslut(ctx context.Context, f domain.BeslutFilter, after *domain.BeslutCursor, limit int) (*BeslutResult, error) {
	now := s.clock()
	rows, err := s.repo.ListBeslut(ctx, f, now, after, limit+1)
	if err != nil {
		return nil, err
	}
	res := &BeslutResult{Items: []BeslutItem{}}
	if len(rows) > limit {
		rows = rows[:limit]
		last := rows[limit-1]
		res.Next = &domain.BeslutCursor{Decided: last.SQLDecided, Date: last.SortDate, Riksmote: last.Riksmote, Beteckning: last.Beteckning}
	}
	keys := make([]string, 0, len(rows))
	for _, r := range rows {
		keys = append(keys, r.Key())
	}
	details, err := s.repo.BeslutDetails(ctx, keys)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		doc := votesports.RiksdagDocument{DecisionDate: r.DecisionDate, Status: r.Status, ChamberDecided: r.ChamberDecided}
		decided := doc.Decided(now)
		if decided != r.SQLDecided {
			slog.Error("beslut list: SQL and Decided() disagree", "key", r.Key(), "sql", r.SQLDecided, "go", decided)
		}
		d := details[r.Key()]
		it := BeslutItem{BeslutRow: r, Decided: decided}
		it.Outcome = domain.BeslutOutcome(decided, d.Beslutstyper, len(d.Ballots) > 0)
		if decided {
			it.Parties = domain.SummarisePartilinje(d.Ballots, domain.RiksdagParties)
		} else {
			it.Parties = []domain.PartyLineSummary{}
		}
		res.Items = append(res.Items, it)
	}
	return res, nil
}
