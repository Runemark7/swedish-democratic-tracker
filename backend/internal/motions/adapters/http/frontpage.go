package http

import (
	"net/http"

	loften "riksdagskollen/internal/loften/domain"
	"riksdagskollen/internal/motions"
)

type frontPageMotion struct {
	listRowResponse
	// DecidedAt is the date of the newest betänkande that decided one of the
	// motion's yrkanden.
	DecidedAt string `json:"decidedAt"`
}

type frontPageCoverage struct {
	Oldest *string `json:"oldest"`
	Newest *string `json:"newest"`
	// Stale is true when the newest decision is older than the current week.
	Stale bool `json:"stale"`
}

type frontPageResponse struct {
	Decided  []beslutRowResponse `json:"decided"`
	Motions  []frontPageMotion   `json:"motions"`
	Upcoming []beslutRowResponse `json:"upcoming"`
	Coverage frontPageCoverage   `json:"coverage"`
	// PartyTiles are the eight parties in alphabetical order of party code,
	// whatever the counts, with Löftesstatus counts from Bekräftade kopplingar.
	PartyTiles []loften.PartySummary `json:"partyTiles"`
}

// frontPage serves GET /frontpage.
func (h *Handler) frontPage(w http.ResponseWriter, r *http.Request) {
	fp, err := h.svc.FrontPage(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := frontPageResponse{
		Decided:  toBeslutResponse(&motions.BeslutResult{Items: fp.Decided}).Data,
		Upcoming: toBeslutResponse(&motions.BeslutResult{Items: fp.Upcoming}).Data,
		Motions:  []frontPageMotion{},
		Coverage: frontPageCoverage{Stale: fp.Stale},
	}
	if fp.Oldest != "" {
		o, n := fp.Oldest, fp.Newest
		out.Coverage.Oldest, out.Coverage.Newest = &o, &n
	}
	rows := toListResponse(&motions.ListResult{Items: fp.Motions}).Data
	for i, row := range rows {
		out.Motions = append(out.Motions, frontPageMotion{listRowResponse: row, DecidedAt: fp.Motions[i].Cursor.Key.Format("2006-01-02")})
	}
	if !h.addLoefteCounts(r.Context(), w, out.Decided, out.Upcoming) {
		return
	}
	out.PartyTiles = []loften.PartySummary{}
	if h.loften != nil {
		tiles, err := h.loften.PartyTiles(r.Context())
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out.PartyTiles = tiles
	}
	jsonOK(w, out)
}
