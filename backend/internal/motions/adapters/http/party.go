package http

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/motions/domain"
)

const (
	partyDecisionLimit = 10
	partyMotionLimit   = 50
)

type partyDecisionResponse struct {
	Riksmote    string  `json:"riksmote"`
	Beteckning  string  `json:"beteckning"`
	Title       string  `json:"title"`
	Utskott     string  `json:"utskott"`
	UtskottName string  `json:"utskottName"`
	Date        *string `json:"date"`
	Decided     bool    `json:"decided"`
	Outcome     string  `json:"outcome"`
	// Partilinje is the party's own line in the votering(ar), empty when the
	// förslagspunkt was decided without one (acklamation) or has none stored.
	Partilinje string `json:"partilinje"`
	Voteringar int    `json:"voteringar"`
	SourceURL  string `json:"sourceUrl"`
}

type partyRecordResponse struct {
	Party     string                  `json:"party"`
	Decisions []partyDecisionResponse `json:"decisions"`
	Motions   []listRowResponse       `json:"motions"`
}

// partyRecord serves a party's recent decisions (with its Partilinje) and its
// Kommittémotioner (with outcomes). Both are plain reads of the lists that
// /api/beslut and /api/motions serve, filtered to the party.
func (h *Handler) partyRecord(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(chi.URLParam(r, "party"))
	if !beteckningPattern.MatchString(code) {
		jsonError(w, "invalid party", http.StatusBadRequest)
		return
	}
	bres, err := h.svc.ListBeslut(r.Context(), domain.BeslutFilter{State: domain.BeslutDecided, Involved: []string{code}}, nil, partyDecisionLimit)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	mres, err := h.svc.List(r.Context(), domain.ListFilter{Parties: []string{code}, Kinds: []string{"Kommittémotion"}}, nil, partyMotionLimit)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := partyRecordResponse{Party: code, Decisions: []partyDecisionResponse{}}
	for _, row := range toBeslutResponse(bres).Data {
		d := partyDecisionResponse{
			Riksmote: row.Riksmote, Beteckning: row.Beteckning, Title: row.Title, Utskott: row.Utskott,
			UtskottName: row.UtskottName, Date: row.Date, Decided: row.Decided, Outcome: row.Outcome, SourceURL: row.SourceURL,
		}
		for _, p := range row.Parties {
			if strings.EqualFold(p.Party, code) {
				d.Partilinje, d.Voteringar = p.Partilinje, p.Voteringar
			}
		}
		out.Decisions = append(out.Decisions, d)
	}
	out.Motions = toListResponse(mres).Data
	jsonOK(w, out)
}
