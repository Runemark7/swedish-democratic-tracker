package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/loften"
	"riksdagskollen/internal/loften/domain"
)

var utskottPattern = regexp.MustCompile(`^[A-Za-zÅÄÖåäö]{2,5}$`)
var betankandePattern = regexp.MustCompile(`^\d{4}/\d{2}:[0-9A-Za-zÅÄÖåäö]{1,12}$`)

// mandatperiodPattern is the shape of mandate_periods.code, e.g. 2026-2030.
var mandatperiodPattern = regexp.MustCompile(`^\d{4}-\d{4}$`)

var partyPattern = regexp.MustCompile(`^[A-Z]{1,4}$`)

type Handler struct{ svc *loften.Service }

func NewHandler(svc *loften.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes(r chi.Router) {
	r.Get("/loften", h.list)
	r.Get("/loften.csv", h.listCSV)
	r.Get("/loften/summary", h.summary)
	r.Get("/loften/{id}", h.get)
	r.Get("/loften/{id}/kopplingar", h.loefteKopplingar)
	r.Get("/kopplingar/{id}", h.koppling)
	r.Get("/valmanifest", h.valmanifest)
	r.Get("/parties/{party}/loften", h.list)
	r.Get("/parties/{party}/loften.csv", h.listCSV)
	r.Get("/parties/{party}/valmanifest", h.valmanifest)
}

// partyParam reads the party from the path or ?party=. ok is false (and a 400
// written) when it is present but malformed.
func partyParam(w http.ResponseWriter, r *http.Request) (string, bool) {
	p := chi.URLParam(r, "party")
	if p == "" {
		p = r.URL.Query().Get("party")
	}
	p = strings.ToUpper(p)
	if p != "" && !partyPattern.MatchString(p) {
		jsonError(w, "invalid party", http.StatusBadRequest)
		return "", false
	}
	return p, true
}

// mandatperiodParam reads ?mandatperiod= ("" = every period). ok is false (and
// a 400 written) when it is present but malformed.
func mandatperiodParam(w http.ResponseWriter, r *http.Request) (string, bool) {
	p := r.URL.Query().Get("mandatperiod")
	if p != "" && !mandatperiodPattern.MatchString(p) {
		jsonError(w, "invalid mandatperiod; write it 2026-2030", http.StatusBadRequest)
		return "", false
	}
	return p, true
}

// selection reads the list filters and returns the matching public Löften.
// ok is false (and an error response written) when a filter is malformed or
// the read fails. The JSON list and the CSV download share it, so a file
// always holds exactly the selection the list shows.
func (h *Handler) selection(w http.ResponseWriter, r *http.Request) ([]*domain.LoefteView, bool) {
	party, ok := partyParam(w, r)
	if !ok {
		return nil, false
	}
	q := r.URL.Query()
	utskott := q.Get("utskott")
	if utskott != "" && !utskottPattern.MatchString(utskott) {
		jsonError(w, "invalid utskott", http.StatusBadRequest)
		return nil, false
	}
	bet := q.Get("betankande")
	if bet != "" && !betankandePattern.MatchString(bet) {
		jsonError(w, "invalid betankande; write it 2026/27:SkU2", http.StatusBadRequest)
		return nil, false
	}
	period, ok := mandatperiodParam(w, r)
	if !ok {
		return nil, false
	}
	out, err := h.svc.ListWithStatus(r.Context(), loften.Filter{Party: party, Utskott: utskott, Status: q.Get("status"), Betankande: bet, Mandatperiod: period})
	if errors.Is(err, loften.ErrBadStatus) {
		jsonError(w, "invalid status", http.StatusBadRequest)
		return nil, false
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return nil, false
	}
	return out, true
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	if out, ok := h.selection(w, r); ok {
		jsonOK(w, out)
	}
}

func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	period, ok := mandatperiodParam(w, r)
	if !ok {
		return
	}
	out, err := h.svc.Summary(r.Context(), period)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, out)
}

func (h *Handler) loefteKopplingar(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	l, err := h.svc.GetWithStatus(r.Context(), id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if l == nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	jsonOK(w, l.Kopplingar)
}

func (h *Handler) koppling(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	k, err := h.svc.Koppling(r.Context(), id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if k == nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	jsonOK(w, k)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	l, err := h.svc.GetWithStatus(r.Context(), id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if l == nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	jsonOK(w, l)
}

func (h *Handler) valmanifest(w http.ResponseWriter, r *http.Request) {
	party, ok := partyParam(w, r)
	if !ok {
		return
	}
	out, err := h.svc.Valmanifest(r.Context(), party)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, out)
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
