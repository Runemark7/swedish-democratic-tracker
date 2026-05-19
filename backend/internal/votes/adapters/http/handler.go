package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/votes"
	"riksdagskollen/internal/votes/domain"
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
	r.Get("/votes/riksdag-feed", h.riksdagFeed)
	r.Get("/votes/{beteckning}/{punkt}", h.getDetail)
	r.Get("/documents/{dokId}/full", h.getDocumentFull)
	r.Get("/documents/{dokId}", h.getDocument)
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

	// Build party breakdown helper
	type partyPos struct {
		Party        string `json:"party"`
		Ja           int    `json:"jaCount"`
		Nej          int    `json:"nejCount"`
		Avstar       int    `json:"avstarCount"`
		Franvarande  int    `json:"franvarandeCount"`
		DominantVote string `json:"dominantVote,omitempty"`
	}

	buildBreakdown := func(votes []*domain.Vote) []*partyPos {
		tally := map[string]*partyPos{}
		for _, v := range votes {
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
		return breakdown
	}

	computeStatus := func(breakdown []*partyPos) string {
		var totalJa, totalNej int
		for _, p := range breakdown {
			totalJa += p.Ja
			totalNej += p.Nej
		}
		if totalJa >= totalNej {
			return "Bifall"
		}
		return "Avslag"
	}

	if len(vv) == 0 {
		// No votes in DB — fetch metadata directly from Riksdagen
		info, err := h.svc.GetBetankandeInfo(r.Context(), bet)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if info == nil {
			jsonError(w, "vote not found", http.StatusNotFound)
			return
		}
		resp := map[string]any{
			"beteckning":     bet,
			"forslagspunkt":  punkt,
			"dokId":          info.DokID,
			"documentTitle":  info.Title,
			"session":        info.Session,
			"date":           info.Date,
			"status":         info.Status,
			"partyBreakdown": []any{},
		}
		if ds, err := h.svc.GetDocumentStatus(r.Context(), info.DokID); err == nil && ds != nil {
			resp["subtitle"] = ds.Subtitle
			resp["summary"] = ds.Summary
			resp["bodyHtml"] = ds.BodyHTML
		}
		jsonOK(w, resp)
		return
	}

	// Has votes in DB
	breakdown := buildBreakdown(vv)
	meta := vv[0]
	for _, v := range vv {
		if v.ProposalOrigin.DocumentTitle != "" {
			meta = v
			break
		}
	}

	resp := map[string]any{
		"beteckning":      bet,
		"forslagspunkt":   punkt,
		"dokId":           meta.DokID,
		"documentTitle":   meta.ProposalOrigin.DocumentTitle,
		"proposedByParty": meta.ProposalOrigin.ProposedByParty,
		"proposalType":    string(meta.ProposalOrigin.ProposalType),
		"session":         meta.Session,
		"partyBreakdown":  breakdown,
		"status":          computeStatus(breakdown),
	}
	if ds, err := h.svc.GetDocumentStatus(r.Context(), meta.DokID); err == nil && ds != nil {
		resp["date"] = ds.Date
		resp["subtitle"] = ds.Subtitle
		resp["summary"] = ds.Summary
		resp["bodyHtml"] = ds.BodyHTML
	}
	jsonOK(w, resp)
}

// organTag maps a Riksdag committee abbreviation to a readable Swedish topic tag.
var organTag = map[string]string{
	"SoU": "Vård",
	"TU":  "Trafik",
	"UbU": "Skola",
	"CU":  "Plan",
}

func (h *Handler) riksdagFeed(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	if level != "region" && level != "kommun" {
		jsonError(w, `level must be "region" or "kommun"`, http.StatusBadRequest)
		return
	}

	docs, err := h.svc.GetRiksdagFeed(r.Context(), level)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type feedItem struct {
		Time       string `json:"time"`
		Title      string `json:"title"`
		Status     string `json:"status"`
		Tag        string `json:"tag,omitempty"`
		Beteckning string `json:"beteckning,omitempty"`
	}
	items := make([]feedItem, 0, len(docs))
	for _, d := range docs {
		tag := organTag[d.Organ]
		items = append(items, feedItem{
			Time:       d.Date,
			Title:      d.Title,
			Status:     "Bifall", // >90% of betänkanden pass; real votering lookup is future work
			Tag:        tag,
			Beteckning: d.Beteckning,
		})
	}
	jsonOK(w, items)
}

func (h *Handler) getDocument(w http.ResponseWriter, r *http.Request) {
	dokID := chi.URLParam(r, "dokId")
	if dokID == "" {
		jsonError(w, "dokId required", http.StatusBadRequest)
		return
	}
	ds, err := h.svc.GetDocumentStatus(r.Context(), dokID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ds == nil {
		jsonError(w, "document not found", http.StatusNotFound)
		return
	}
	// Strip the heavy bodyHtml — only the beslut detail page needs it.
	resp := map[string]any{
		"dokId":      ds.DokID,
		"type":       ds.Type,
		"title":      ds.Title,
		"subtitle":   ds.Subtitle,
		"summary":    ds.Summary,
		"date":       ds.Date,
		"beteckning": ds.Beteckning,
	}
	jsonOK(w, resp)
}

func (h *Handler) getDocumentFull(w http.ResponseWriter, r *http.Request) {
	dokID := chi.URLParam(r, "dokId")
	if dokID == "" {
		jsonError(w, "dokId required", http.StatusBadRequest)
		return
	}
	ds, err := h.svc.GetDocumentStatus(r.Context(), dokID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ds == nil {
		jsonError(w, "document not found", http.StatusNotFound)
		return
	}
	intressenter := ds.Intressenter
	if intressenter == nil {
		intressenter = []domain.Intressent{}
	}
	resp := map[string]any{
		"dokId":        ds.DokID,
		"type":         ds.Type,
		"title":        ds.Title,
		"subtitle":     ds.Subtitle,
		"summary":      ds.Summary,
		"date":         ds.Date,
		"beteckning":   ds.Beteckning,
		"bodyHtml":     ds.BodyHTML,
		"intressenter": intressenter,
	}
	jsonOK(w, resp)
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
