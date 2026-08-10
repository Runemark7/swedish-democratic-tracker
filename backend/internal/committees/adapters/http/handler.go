package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/committees"
	"riksdagskollen/internal/committees/domain"
)

type Handler struct {
	svc *committees.Service
}

func NewHandler(svc *committees.Service) *Handler {
	return &Handler{svc: svc}
}

// Routes registers the committee endpoints on r, matching the convention
// used by every other feature's HTTP adapter.
func (h *Handler) Routes(r chi.Router) {
	r.Get("/committees", h.list)
	r.Get("/committees/{code}", h.get)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		jsonError(w, "period is required", http.StatusBadRequest)
		return
	}
	cs, err := h.svc.List(r.Context(), period)
	if err != nil {
		writeErr(w, err)
		return
	}
	jsonOK(w, cs)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		jsonError(w, "period is required", http.StatusBadRequest)
		return
	}
	// year 0 means "not specified" — the service resolves it to the newest
	// decided budget year rather than us guessing one here. Because 0 carries
	// that meaning internally, an explicit ?year=0 has to be refused here: it
	// is not a budget year, and silently reading it as "unspecified" would
	// answer a question nobody asked.
	year := 0
	if y := r.URL.Query().Get("year"); y != "" {
		parsed, err := strconv.Atoi(y)
		if err != nil {
			jsonError(w, "year must be an integer", http.StatusBadRequest)
			return
		}
		if parsed <= 0 {
			jsonError(w, "year must be a real budget year", http.StatusBadRequest)
			return
		}
		year = parsed
	}

	// Canonicalise the path code before looking it up. /committees/fiu is the
	// same committee as /committees/FiU, and answering "committee not found in
	// that period" for a committee that decided 104 voteringar would be a
	// false statement produced by a casing gate — exactly the kind of gate
	// this feature exists to remove.
	code := domain.Canonical(chi.URLParam(r, "code"))

	c, err := h.svc.Get(r.Context(), period, code, year)
	if err != nil {
		writeErr(w, err)
		return
	}
	jsonOK(w, c)
}

// writeErr maps the service's named refusals onto status codes. A request the
// record cannot answer must say so — 404 for a period or committee we do not
// hold, 400 for a budget year without decided figures — never a 200 carrying
// an empty list or a zero amount.
func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, committees.ErrPeriodNotFound),
		errors.Is(err, committees.ErrCommitteeNotFound):
		jsonError(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, committees.ErrBudgetYearNotDecided):
		jsonError(w, err.Error(), http.StatusBadRequest)
	default:
		jsonError(w, err.Error(), http.StatusInternalServerError)
	}
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
