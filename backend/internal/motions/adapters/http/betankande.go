package http

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/motions"
	"riksdagskollen/internal/motions/domain"
)

// decisionTextPlanned is shown for a betänkande the chamber has not decided.
const decisionTextPlanned = "Ej avgjort än"

type betankandeResponse struct {
	Riksmote   string `json:"riksmote"`
	Beteckning string `json:"beteckning"`
	Title      string `json:"title"`
	Organ      string `json:"organ"`
	DokID      string `json:"dokId"`
	SourceURL  string `json:"sourceUrl"`
	// Held is false when the punkter come from the voteringar alone.
	Held bool `json:"held"`
	// Status, DecisionDate and Decided are Riksdagen's (beslutsdag, beslutad,
	// status) read live. Decided is null when that could not be read. When it
	// is false, DecisionDate is the planned date and DecisionText says so.
	Status       string          `json:"status"`
	Decided      *bool           `json:"decided"`
	DecisionDate string          `json:"decisionDate"`
	DecisionText string          `json:"decisionText"`
	Punkter      []punktResponse `json:"punkter"`
}

type avvikareResponse struct {
	PoliticianID string `json:"politicianId"`
	Name         string `json:"name"`
	Party        string `json:"party"`
	VoteResult   string `json:"voteResult"`
	Partilinje   string `json:"partilinje"`
}

type punktMotionResponse struct {
	DokID      string   `json:"dokId"`
	Riksmote   string   `json:"riksmote"`
	Beteckning string   `json:"beteckning"`
	Title      string   `json:"title"`
	Subtyp     string   `json:"subtyp"`
	Parties    []string `json:"parties"`
	Yrkanden   []int    `json:"yrkanden"`
}

type punktResponse struct {
	Number     string `json:"number"`
	Rubrik     string `json:"rubrik"`
	Text       string `json:"text"`
	Beslutstyp string `json:"beslutstyp"`
	Mode       string `json:"mode"`
	ModeLabel  string `json:"modeLabel"`
	Alternativ string `json:"alternativ"`
	Vinnare    string `json:"vinnare"`
	VoteringID string `json:"voteringId"`

	VoteringStored        bool                      `json:"voteringStored"`
	Parties               []partyResponse           `json:"parties"`
	Avvikare              []avvikareResponse        `json:"avvikare"`
	Reasoning             string                    `json:"reasoning"`
	ReasoningAvailable    bool                      `json:"reasoningAvailable"`
	ReservationsAvailable bool                      `json:"reservationsAvailable"`
	Reservations          []statementResponse       `json:"reservations"`
	SarskildaYttranden    []statementResponse       `json:"sarskildaYttranden"`
	VotedReservation      *votedReservationResponse `json:"votedReservation"`
	Motions               []punktMotionResponse     `json:"motions"`
}

func (h *Handler) getBetankande(w http.ResponseWriter, r *http.Request) {
	rm, err := url.PathUnescape(chi.URLParam(r, "riksmote"))
	bet := chi.URLParam(r, "beteckning")
	if err != nil || !riksmotePattern.MatchString(rm) || !beteckningPattern.MatchString(bet) {
		jsonError(w, "invalid riksmöte or beteckning; the riksmöte is written 2025/26 and sent as 2025%2F26", http.StatusBadRequest)
		return
	}
	p, err := h.svc.Betankande(r.Context(), rm, bet)
	if errors.Is(err, motions.ErrBetankandeNotFound) {
		jsonError(w, "betänkande not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := betankandeResponse{
		Riksmote: p.Riksmote, Beteckning: p.Beteckning, Title: p.Title, Organ: p.Organ, DokID: p.DokID,
		Held: p.Held, Status: p.Status, Decided: p.Decided, DecisionDate: p.DecisionDate,
		Punkter: make([]punktResponse, 0, len(p.Punkter)),
	}
	if p.DokID != "" {
		out.SourceURL = sourceURLBase + p.DokID + ".html"
	}
	if p.Decided != nil && !*p.Decided {
		out.DecisionText = decisionTextPlanned
	}
	for _, v := range p.Punkter {
		out.Punkter = append(out.Punkter, toPunkt(p, v))
	}
	jsonOK(w, out)
}

func toPunkt(p *motions.BetankandePage, v motions.PunktView) punktResponse {
	// Reuse the motion page's decision mapping so both pages word it alike.
	d := toDecision(domain.YrkandeDecision{
		Riksmote: p.Riksmote, Betankande: p.Beteckning, BetDokID: p.DokID, Punkt: v.Forslagspunkt,
		VoteringStored: v.VoteringStored, Reservations: v.Reservations, Yttranden: v.Yttranden,
		HTMLRead: p.HTMLRead, HasReasoningSection: p.HasReasoningSection, HasReservationsSection: p.HasReservationsSection,
	}, "")
	out := punktResponse{
		Number: v.Number, Rubrik: v.Rubrik, Text: v.Text, Beslutstyp: v.Beslutstyp, Mode: d.Mode, ModeLabel: d.ModeLabel,
		Alternativ: v.Alternativ, Vinnare: v.Vinnare, VoteringID: v.VoteringID, VoteringStored: v.VoteringStored,
		Reasoning: d.Reasoning, ReasoningAvailable: d.ReasoningAvailable,
		ReservationsAvailable: d.ReservationsAvailable && p.Held, Reservations: d.Reservations,
		SarskildaYttranden: d.SarskildaYttranden, VotedReservation: d.VotedReservation,
		Parties: make([]partyResponse, 0, len(v.Parties)), Avvikare: []avvikareResponse{}, Motions: []punktMotionResponse{},
	}
	for _, t := range v.Parties {
		out.Parties = append(out.Parties, partyResponse{Party: t.Party, Partilinje: string(t.Partilinje), Ja: t.Ja, Nej: t.Nej, Avstar: t.Avstar, Franvarande: t.Franvarande})
	}
	for _, a := range v.Avvikare {
		out.Avvikare = append(out.Avvikare, avvikareResponse{PoliticianID: a.PoliticianID, Name: a.Name, Party: a.Party, VoteResult: a.Result, Partilinje: string(a.Partilinje)})
	}
	for _, m := range v.Motions {
		out.Motions = append(out.Motions, punktMotionResponse{DokID: m.DokID, Riksmote: m.Riksmote, Beteckning: m.Beteckning, Title: m.Title, Subtyp: m.Subtyp, Parties: m.Parties, Yrkanden: m.Yrkanden})
	}
	return out
}

// resolve backs the redirects of old URLs: a vote detail URL (beteckning +
// punkt, optionally the votering's riksmöte) and the betänkande URL without a
// riksmöte (the most recent riksmöte holding the beteckning).
func (h *Handler) resolveBetankande(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	bet, punkt, rm := q.Get("beteckning"), q.Get("punkt"), q.Get("rm")
	if !beteckningPattern.MatchString(bet) || (rm != "" && !riksmotePattern.MatchString(rm)) || len(punkt) > 12 {
		jsonError(w, "invalid beteckning, punkt or rm", http.StatusBadRequest)
		return
	}
	found, ok, err := h.svc.Resolve(r.Context(), bet, punkt, rm)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		jsonError(w, "betänkande not found", http.StatusNotFound)
		return
	}
	jsonOK(w, map[string]string{"riksmote": found, "beteckning": bet, "punkt": punkt})
}
