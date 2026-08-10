package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/committees"
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
		jsonError(w, err.Error(), http.StatusInternalServerError)
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
	// decided budget year rather than us guessing one here.
	year := 0
	if y := r.URL.Query().Get("year"); y != "" {
		parsed, err := strconv.Atoi(y)
		if err != nil {
			jsonError(w, "year must be an integer", http.StatusBadRequest)
			return
		}
		year = parsed
	}
	c, err := h.svc.Get(r.Context(), period, chi.URLParam(r, "code"), year)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if c == nil {
		jsonError(w, "committee not found in that period", http.StatusNotFound)
		return
	}
	jsonOK(w, c)
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
