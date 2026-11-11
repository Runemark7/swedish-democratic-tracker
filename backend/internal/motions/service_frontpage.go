package motions

import (
	"context"
	"time"

	"riksdagskollen/internal/motions/domain"
)

// Front-page sizes.
const (
	FrontPageDecided  = 5
	FrontPageMotions  = 5
	FrontPageUpcoming = 3
)

// FrontPage is what the front page shows: the newest decided betänkanden, the
// recently decided motions and the next betänkanden on the plan. Decided and
// Upcoming come from two separate queries, so a planned betänkande cannot
// appear among the decided ones.
type FrontPage struct {
	Decided  []BeslutItem
	Motions  []ListItem
	Upcoming []BeslutItem
	// Oldest, Newest are the decision dates the Decided rows cover ("" when
	// none). Stale is true when Newest is before the start (Monday) of the
	// current week, or when nothing is decided at all.
	Oldest, Newest string
	Stale          bool
}

// ItemDate is the decision date of a decided row, the planned beslutsdag of an
// upcoming one; "" when Riksdagen gives none.
func (b BeslutItem) ItemDate() string {
	if b.DecisionDate != "" {
		return b.DecisionDate
	}
	if b.Decided {
		return b.SortDate.Format("2006-01-02")
	}
	return ""
}

// FrontPage assembles the feed.
func (s *Service) FrontPage(ctx context.Context) (*FrontPage, error) {
	dec, err := s.ListBeslut(ctx, domain.BeslutFilter{State: domain.BeslutDecided}, nil, FrontPageDecided)
	if err != nil {
		return nil, err
	}
	up, err := s.ListBeslut(ctx, domain.BeslutFilter{State: domain.BeslutUpcoming}, nil, FrontPageUpcoming)
	if err != nil {
		return nil, err
	}
	// The motion list puts decided motions first; the ones after them are
	// undecided and not part of this block.
	ml, err := s.List(ctx, domain.ListFilter{}, nil, FrontPageMotions)
	if err != nil {
		return nil, err
	}
	fp := &FrontPage{Decided: dec.Items, Upcoming: up.Items, Motions: []ListItem{}}
	for _, m := range ml.Items {
		if m.Cursor.Decided {
			fp.Motions = append(fp.Motions, m)
		}
	}
	for _, it := range fp.Decided {
		d := it.ItemDate()
		if d == "" {
			continue
		}
		if fp.Oldest == "" || d < fp.Oldest {
			fp.Oldest = d
		}
		if d > fp.Newest {
			fp.Newest = d
		}
	}
	fp.Stale = fp.Newest == "" || fp.Newest < weekStart(s.clock())
	return fp, nil
}

// weekStart is the Monday of t's week, as a date.
func weekStart(t time.Time) string {
	back := (int(t.Weekday()) + 6) % 7
	return t.AddDate(0, 0, -back).Format("2006-01-02")
}
