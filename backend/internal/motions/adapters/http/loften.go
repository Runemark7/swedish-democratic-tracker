package http

import (
	"context"
	"net/http"

	loften "riksdagskollen/internal/loften/domain"
)

// LoefteSource tells which Löften a decision touches and summarises Löftesstatus
// per party. It is read-only and counts Bekräftade kopplingar only. Wired in
// cmd/api; when nil the pages carry no Löfte data.
type LoefteSource interface {
	ForBetankanden(ctx context.Context, keys []string) (map[string][]loften.LoefteRef, error)
	PartyTiles(ctx context.Context) ([]loften.PartySummary, error)
}

// WithLoften attaches the Löfte source.
func (h *Handler) WithLoften(src LoefteSource) *Handler {
	h.loften = src
	return h
}

// loefteByKey reads the Löften of the given "riksmöte:beteckning" keys. A
// failure is reported to the caller, never swallowed into a count of zero.
func (h *Handler) loefteByKey(ctx context.Context, keys []string) (map[string][]loften.LoefteRef, error) {
	if h.loften == nil || len(keys) == 0 {
		return map[string][]loften.LoefteRef{}, nil
	}
	return h.loften.ForBetankanden(ctx, keys)
}

func orEmpty(l []loften.LoefteRef) []loften.LoefteRef {
	if l == nil {
		return []loften.LoefteRef{}
	}
	return l
}

// addLoefteCounts fills LoefteCount on beslut rows.
func (h *Handler) addLoefteCounts(ctx context.Context, w http.ResponseWriter, rows ...[]beslutRowResponse) bool {
	var keys []string
	for _, rs := range rows {
		for _, r := range rs {
			keys = append(keys, r.Riksmote+":"+r.Beteckning)
		}
	}
	got, err := h.loefteByKey(ctx, keys)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	for _, rs := range rows {
		for i := range rs {
			rs[i].LoefteCount = len(got[rs[i].Riksmote+":"+rs[i].Beteckning])
		}
	}
	return true
}

// addMotionLoefte fills the per-decision counts and the motion's own list, the
// union of the Löften of every betänkande that decided one of its yrkanden.
func (h *Handler) addMotionLoefte(ctx context.Context, m *motionResponse) error {
	var keys []string
	for _, y := range m.Yrkanden {
		for _, d := range y.Decisions {
			keys = append(keys, d.BetankandeRiksmote+":"+d.BetankandeBeteckning)
		}
	}
	got, err := h.loefteByKey(ctx, keys)
	if err != nil {
		return err
	}
	m.Loften = []loften.LoefteRef{}
	seen := map[int]bool{}
	for i := range m.Yrkanden {
		for j := range m.Yrkanden[i].Decisions {
			d := &m.Yrkanden[i].Decisions[j]
			d.Loften = orEmpty(got[d.BetankandeRiksmote+":"+d.BetankandeBeteckning])
			d.LoefteCount = len(d.Loften)
			for _, l := range d.Loften {
				if !seen[l.ID] {
					seen[l.ID] = true
					m.Loften = append(m.Loften, l)
				}
			}
		}
	}
	return nil
}
