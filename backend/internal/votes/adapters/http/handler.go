package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	committeedomain "riksdagskollen/internal/committees/domain"
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
	r.Get("/votes/recent", h.recent)
	r.Get("/votes/{beteckning}/{punkt}", h.getDetail)
	r.Get("/documents/{dokId}/full", h.getDocumentFull)
	r.Get("/documents/{dokId}", h.getDocument)
	r.Get("/committees/{code}/votes", h.listByCommittee)
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

// listByCommittee serves the RÖSTAT row on the committee page: the
// committee's voteringar with each party's dominant position, alphabetical
// by party within a votering. States positions as fact only — no share, no
// rank, no comparison to the LOVAT goals shown alongside it; the reader
// connects the two.
func (h *Handler) listByCommittee(w http.ResponseWriter, r *http.Request) {
	// Required, no default — mirrors GET /committees/{code} exactly. An
	// unscoped query would blend mandate periods together the moment the
	// record holds more than one.
	period := r.URL.Query().Get("period")
	if period == "" {
		jsonError(w, "period is required", http.StatusBadRequest)
		return
	}

	// Canonicalise before querying, same as /committees/{code}/goals: a raw
	// path segment like "sou" must behave like "SoU".
	code := committeedomain.Canonical(chi.URLParam(r, "code"))

	// A code Canonical does not recognise as a committee (no "U" suffix, e.g.
	// "%" or "_") comes back "". The repository's prefix match is
	// starts_with(beteckning, code), and starts_with(x, "") is true for every
	// x — so an empty code would smuggle in exactly the wildcard-match-everything
	// the starts_with predicate was chosen over LIKE to prevent. Refuse it here,
	// before it reaches the query, rather than let the predicate answer it.
	if code == "" {
		jsonOK(w, map[string]any{"items": []ports.CommitteeVotering{}, "total": 0})
		return
	}

	limit := queryInt(r, "limit", 50)
	if limit > 200 {
		limit = 200
	}
	offset := queryInt(r, "offset", 0)

	items, total, err := h.svc.ListByCommitteeWithPositions(r.Context(), period, code, limit, offset)
	if err != nil {
		if errors.Is(err, votes.ErrPeriodNotFound) {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// The repository never returns nil — a zero-row result comes back as an
	// explicit []ports.CommitteeVotering{}, so no nil-guard is needed here.
	jsonOK(w, map[string]any{
		"items": items,
		"total": total,
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
			resp["debattDate"] = ds.DebattDate
			resp["beslutDate"] = ds.BeslutDate
			resp["statusText"] = ds.StatusText
			resp["notis"] = ds.Notis
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
		resp["debattDate"] = ds.DebattDate
		resp["beslutDate"] = ds.BeslutDate
		resp["statusText"] = ds.StatusText
		resp["notis"] = ds.Notis
	}
	jsonOK(w, resp)
}

// recent serves the complete, dated betänkande feed the front page's timeline
// is built on: every committee, no selection. Unlike riksdagFeed, it applies
// no committee list — see the NOTE on FetchDocuments for why one would be
// inert on the upstream call anyway.
func (h *Handler) recent(w http.ResponseWriter, r *http.Request) {
	count := queryInt(r, "count", 40)
	if count > 200 {
		count = 200
	}

	docs, err := h.svc.GetRecentBetankanden(r.Context(), count)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// decided and decisionDate travel with every item because the list mixes
	// decided betänkanden with ones that are only planned, and `date` cannot
	// tell them apart — it is a publication date on both.
	type recentItem struct {
		Title        string `json:"title"`
		Organ        string `json:"organ"`
		Date         string `json:"date"`
		Beteckning   string `json:"beteckning"`
		Decided      bool   `json:"decided"`
		DecisionDate string `json:"decisionDate"`
		Status       string `json:"status"`
	}
	items := make([]recentItem, 0, len(docs))
	for _, d := range docs {
		items = append(items, recentItem{
			Title:        d.Title,
			Organ:        d.Organ,
			Date:         d.Date,
			Beteckning:   d.Beteckning,
			Decided:      d.Decided(),
			DecisionDate: d.DecisionDate,
			Status:       d.Status,
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
