package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/budget"
)

type Handler struct {
	svc *budget.Service
}

func NewHandler(svc *budget.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/budget/years", h.listYears)
	r.Get("/budget/years/{year}", h.getYear)
	r.Get("/budget/years/{year}/compare/{compareYear}", h.compareYears)
	r.Get("/budget/areas", h.listAreas)
	r.Get("/budget/areas/{code}", h.getAreaTimeSeries)
}

func (h *Handler) listYears(w http.ResponseWriter, r *http.Request) {
	years, err := h.svc.ListYears(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, years)
}

func (h *Handler) getYear(w http.ResponseWriter, r *http.Request) {
	year, err := strconv.Atoi(chi.URLParam(r, "year"))
	if err != nil {
		jsonError(w, "invalid year", http.StatusBadRequest)
		return
	}

	detail, err := h.svc.GetYearDetail(r.Context(), year)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if detail == nil {
		jsonError(w, "budget year not found", http.StatusNotFound)
		return
	}
	jsonOK(w, detail)
}

func (h *Handler) compareYears(w http.ResponseWriter, r *http.Request) {
	baseYear, err := strconv.Atoi(chi.URLParam(r, "year"))
	if err != nil {
		jsonError(w, "invalid base year", http.StatusBadRequest)
		return
	}
	compareYear, err := strconv.Atoi(chi.URLParam(r, "compareYear"))
	if err != nil {
		jsonError(w, "invalid compare year", http.StatusBadRequest)
		return
	}

	comp, err := h.svc.CompareYears(r.Context(), baseYear, compareYear)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if comp == nil {
		jsonError(w, "no budget data for the requested years", http.StatusNotFound)
		return
	}
	jsonOK(w, comp)
}

func (h *Handler) listAreas(w http.ResponseWriter, r *http.Request) {
	areas, err := h.svc.ListAreas(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, areas)
}

func (h *Handler) getAreaTimeSeries(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	ts, err := h.svc.GetAreaTimeSeries(r.Context(), code)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ts == nil {
		jsonError(w, "expenditure area not found", http.StatusNotFound)
		return
	}
	jsonOK(w, ts)
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
