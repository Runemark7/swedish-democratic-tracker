package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/votes"
	"riksdagskollen/internal/votes/ports"
)

type Handler struct {
	svc *votes.Service
}

func NewHandler(svc *votes.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/politicians/{id}/votes", h.listByPolitician)
	r.Get("/votes", h.listAll)
	r.Get("/votes/{beteckning}/{punkt}", h.getDetail)
}

func (h *Handler) listByPolitician(w http.ResponseWriter, r *http.Request) {
	f := ports.ListVotesFilter{
		PoliticianID: chi.URLParam(r, "id"),
		Session:      r.URL.Query().Get("session"),
		Page:         queryInt(r, "page", 1),
		PageSize:     queryInt(r, "pageSize", 50),
	}

	result, err := h.svc.ListByPolitician(r.Context(), f)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]any{
		"data":     result.Votes,
		"total":    result.Total,
		"page":     f.Page,
		"pageSize": f.PageSize,
	})
}

func (h *Handler) listAll(w http.ResponseWriter, r *http.Request) {
	f := ports.ListDistinctVotesFilter{
		Page:     queryInt(r, "page", 1),
		PageSize: queryInt(r, "pageSize", 50),
	}

	result, err := h.svc.ListDistinctVotes(r.Context(), f)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]any{
		"data":     result.Votes,
		"total":    result.Total,
		"page":     f.Page,
		"pageSize": f.PageSize,
	})
}

func (h *Handler) getDetail(w http.ResponseWriter, r *http.Request) {
	bet := chi.URLParam(r, "beteckning")
	punkt := chi.URLParam(r, "punkt")

	vv, err := h.svc.ListByBeteckning(r.Context(), bet, punkt)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(vv) == 0 {
		jsonError(w, "vote not found", http.StatusNotFound)
		return
	}

	// Build party breakdown
	type partyPos struct {
		Party        string `json:"party"`
		Ja           int    `json:"jaCount"`
		Nej          int    `json:"nejCount"`
		Avstar       int    `json:"avstarCount"`
		Franvarande  int    `json:"franvarandeCount"`
		DominantVote string `json:"dominantVote,omitempty"`
	}
	tally := map[string]*partyPos{}
	for _, v := range vv {
		pos, ok := tally[v.Party]
		if !ok {
			pos = &partyPos{Party: v.Party}
			tally[v.Party] = pos
		}
		switch v.VoteResult {
		case "Ja":
			pos.Ja++
		case "Nej":
			pos.Nej++
		case "Avstår":
			pos.Avstar++
		case "Frånvarande":
			pos.Franvarande++
		}
	}
	breakdown := make([]*partyPos, 0, len(tally))
	for _, p := range tally {
		// Compute dominant vote (exclude Frånvarande)
		max := p.Ja
		if p.Nej > max {
			max = p.Nej
		}
		if p.Avstar > max {
			max = p.Avstar
		}
		ties := 0
		if p.Ja == max {
			ties++
		}
		if p.Nej == max {
			ties++
		}
		if p.Avstar == max {
			ties++
		}
		if ties == 1 && max > 0 {
			switch {
			case p.Ja == max:
				p.DominantVote = "Ja"
			case p.Nej == max:
				p.DominantVote = "Nej"
			case p.Avstar == max:
				p.DominantVote = "Avstår"
			}
		}
		breakdown = append(breakdown, p)
	}

	// Pick the first enriched vote for metadata (title, origin, etc.)
	meta := vv[0]
	for _, v := range vv {
		if v.ProposalOrigin.DocumentTitle != "" {
			meta = v
			break
		}
	}
	jsonOK(w, map[string]any{
		"beteckning":      bet,
		"forslagspunkt":   punkt,
		"dokId":           meta.DokID,
		"documentTitle":   meta.ProposalOrigin.DocumentTitle,
		"proposedByParty": meta.ProposalOrigin.ProposedByParty,
		"proposalType":    string(meta.ProposalOrigin.ProposalType),
		"session":         meta.Session,
		"partyBreakdown":  breakdown,
	})
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return def
	}
	return n
}
