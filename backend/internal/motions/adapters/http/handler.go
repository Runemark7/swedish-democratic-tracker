// Package http serves motions over the public API.
package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/motions"
	"riksdagskollen/internal/motions/domain"
	"riksdagskollen/internal/motions/ports"
	votes "riksdagskollen/internal/votes/domain"
)

type Handler struct{ svc *motions.Service }

func NewHandler(svc *motions.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes(r chi.Router) {
	r.Get("/motions/coverage", h.coverage)
	// The riksmöte sits in one path segment, so its slash is encoded: 2017%2F18.
	r.Get("/motions/{riksmote}/{beteckning}", h.get)
}

var (
	riksmotePattern   = regexp.MustCompile(`^\d{4}/\d{2}$`)
	beteckningPattern = regexp.MustCompile(`^[0-9A-Za-z]{1,12}$`)
)

const sourceURLBase = "https://data.riksdagen.se/dokument/"

type motionResponse struct {
	DokID           string            `json:"dokId"`
	Riksmote        string            `json:"riksmote"`
	Beteckning      string            `json:"beteckning"`
	Title           string            `json:"title"`
	Subtyp          string            `json:"subtyp"`
	Kind            string            `json:"kind"`
	KindExplanation string            `json:"kindExplanation"`
	Date            *string           `json:"date"`
	Utskott         string            `json:"utskott"`
	UtskottName     string            `json:"utskottName"`
	Status          string            `json:"status"`
	BehandlasI      string            `json:"behandlasI"`
	SourceURL       string            `json:"sourceUrl"`
	Authors         []authorResponse  `json:"authors"`
	Yrkanden        []yrkandeResponse `json:"yrkanden"`
	Outcome         outcomeResponse   `json:"outcome"`
}

// outcomeResponse is a motion's outcome with the counts behind it.
type outcomeResponse struct {
	Word   string         `json:"word"`
	Counts map[string]int `json:"counts"`
	Total  int            `json:"total"`
}

type decisionResponse struct {
	BetankandeRiksmote   string `json:"betankandeRiksmote"`
	BetankandeBeteckning string `json:"betankandeBeteckning"`
	BetankandeTitle      string `json:"betankandeTitle"`
	BetankandeSourceURL  string `json:"betankandeSourceUrl"`
	Punkt                string `json:"punkt"`
	Rubrik               string `json:"rubrik"`
	LinkedBy             string `json:"linkedBy"`
	Beslutstyp           string `json:"beslutstyp"`
	// Mode is acklamation, votering, ej_beslutad or annat; ModeLabel is the
	// plain Swedish line ("Avgjord utan omröstning").
	Mode      string `json:"mode"`
	ModeLabel string `json:"modeLabel"`
	// Alternativ is the alternative put to the vote, as Riksdagen words it.
	Alternativ string `json:"alternativ"`
	Outcome    string `json:"outcome"`
	// VoteringStored is true when we hold the ballots of the votering.
	VoteringStored bool            `json:"voteringStored"`
	Parties        []partyResponse `json:"parties"`
}

type partyResponse struct {
	Party       string `json:"party"`
	Partilinje  string `json:"partilinje"`
	Ja          int    `json:"ja"`
	Nej         int    `json:"nej"`
	Avstar      int    `json:"avstar"`
	Franvarande int    `json:"franvarande"`
}

type authorResponse struct {
	PoliticianID string `json:"politicianId"`
	Name         string `json:"name"`
	Party        string `json:"party"`
	Role         string `json:"role"`
}

type yrkandeResponse struct {
	Number          int    `json:"number"`
	Wording         string `json:"wording"`
	UtskottProposal string `json:"utskottProposal"`
	ChamberDecision string `json:"chamberDecision"`
	BehandlasI      string `json:"behandlasI"`
	BehandlasIPunkt string `json:"behandlasIPunkt"`
	// Outcome is the plain-language outcome: Avslagen, Bifallen, Delvis
	// bifallen, Besvarad, Ej avgjord än, or Riksdagen's own word.
	Outcome   string             `json:"outcome"`
	Decisions []decisionResponse `json:"decisions"`
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	rm, err := url.PathUnescape(chi.URLParam(r, "riksmote"))
	bet := chi.URLParam(r, "beteckning")
	if err != nil || !riksmotePattern.MatchString(rm) || !beteckningPattern.MatchString(bet) {
		jsonError(w, "invalid riksmöte or beteckning; the riksmöte is written 2017/18 and sent as 2017%2F18", http.StatusBadRequest)
		return
	}
	m, err := h.svc.Get(r.Context(), rm, bet)
	if errors.Is(err, ports.ErrNotFound) {
		jsonError(w, "motion not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	decisions, err := h.svc.Decisions(r.Context(), m.DokID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, toResponse(m, decisions))
}

func toResponse(m *domain.Motion, decisions []domain.YrkandeDecision) motionResponse {
	kind, expl := m.Kind()
	out := motionResponse{
		DokID: m.DokID, Riksmote: m.Riksmote, Beteckning: m.Beteckning, Title: m.Title,
		Subtyp: m.Subtyp, Kind: kind, KindExplanation: expl,
		Utskott: m.Organ, UtskottName: m.Tilldelat, Status: m.Status, BehandlasI: m.BehandlasI,
		SourceURL: sourceURLBase + m.DokID + ".html",
		Authors:   make([]authorResponse, 0, len(m.Authors)),
		Yrkanden:  make([]yrkandeResponse, 0, len(m.Yrkanden)),
	}
	if m.Date != nil {
		d := m.Date.Format("2006-01-02")
		out.Date = &d
	}
	for _, a := range m.Authors {
		out.Authors = append(out.Authors, authorResponse{PoliticianID: a.PoliticianID, Name: a.Name, Party: a.Party, Role: a.Role})
	}
	byYrkande := map[int][]domain.YrkandeDecision{}
	for _, d := range decisions {
		byYrkande[d.YrkandeNumber] = append(byYrkande[d.YrkandeNumber], d)
	}
	var words []string
	for _, y := range m.Yrkanden {
		yr := yrkandeResponse{
			Number: y.Number, Wording: y.Wording, UtskottProposal: y.UtskottProposal,
			ChamberDecision: y.ChamberDecision, BehandlasI: y.BehandlasI, BehandlasIPunkt: y.BehandlasIPunkt,
			Decisions: []decisionResponse{},
		}
		var dws []string
		for _, d := range byYrkande[y.Number] {
			word := domain.YrkandeOutcome(y.ChamberDecision, d.Decision)
			dws = append(dws, word)
			yr.Decisions = append(yr.Decisions, toDecision(d, word))
		}
		// Without a linked förslagspunkt there is no decision to show: the
		// yrkande is Ej avgjord än, whatever the listing's own columns say.
		switch len(dws) {
		case 0:
			yr.Outcome = domain.OutcomeEjAvgjord
		case 1:
			yr.Outcome = dws[0]
		default:
			yr.Outcome = domain.MotionOutcome(dws).Word
		}
		words = append(words, yr.Outcome)
		out.Yrkanden = append(out.Yrkanden, yr)
	}
	agg := domain.MotionOutcome(words)
	out.Outcome = outcomeResponse{Word: agg.Word, Counts: agg.Counts, Total: agg.Total}
	return out
}

type coverageResponse struct {
	Riksmote         string `json:"riksmote"`
	Source           string `json:"source"`
	Expected         int    `json:"expected"`
	Held             int    `json:"held"`
	LastEnumeratedAt string `json:"lastEnumeratedAt"`
}

func (h *Handler) coverage(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.Coverage(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]coverageResponse, 0, len(rows))
	for _, c := range rows {
		out = append(out, coverageResponse{
			Riksmote: c.Riksmote, Source: c.Source, Expected: c.Expected, Held: c.Held,
			LastEnumeratedAt: c.LastEnumeratedAt.UTC().Format(time.RFC3339),
		})
	}
	jsonOK(w, map[string]any{"data": out})
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func toDecision(d domain.YrkandeDecision, outcome string) decisionResponse {
	mode := domain.ModeOf(d.Punkt.Beslutstyp, d.Punkt.VoteringID)
	label := mode.Label()
	if mode == domain.ModeAnnat {
		label = d.Punkt.Beslutstyp // Riksdagen's own word
	}
	out := decisionResponse{
		BetankandeRiksmote: d.Riksmote, BetankandeBeteckning: d.Betankande, BetankandeTitle: d.BetTitle,
		Punkt: d.Punkt.Number, Rubrik: d.Punkt.Rubrik, LinkedBy: d.LinkedBy, Beslutstyp: d.Punkt.Beslutstyp,
		Mode: string(mode), ModeLabel: label, Alternativ: d.Punkt.Alternativ, Outcome: outcome,
		VoteringStored: d.VoteringStored, Parties: []partyResponse{},
	}
	if d.BetDokID != "" {
		out.BetankandeSourceURL = sourceURLBase + d.BetDokID + ".html"
	}
	for _, t := range d.Parties {
		tally := votes.Tally{Ja: t.Ja, Nej: t.Nej, Avstar: t.Avstar, Franvarande: t.Franvarande}
		out.Parties = append(out.Parties, partyResponse{
			Party: t.Party, Partilinje: string(votes.PartilinjeFor(t.Party, tally)),
			Ja: t.Ja, Nej: t.Nej, Avstar: t.Avstar, Franvarande: t.Franvarande,
		})
	}
	return out
}
