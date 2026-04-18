package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"riksdagskollen/internal/regions"
	"riksdagskollen/internal/regions/ports"
)

type Handler struct {
	svc *regions.Service
}

func NewHandler(svc *regions.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/regions", h.listRegions)
	r.Get("/regions/{code}", h.getRegion)
	r.Get("/municipalities", h.listMunicipalities)
	r.Get("/municipalities/{code}", h.getMunicipality)
	r.Get("/municipalities/{code}/kpi", h.getMunicipalityKPI)
	r.Get("/municipalities/{code}/spending", h.getMunicipalitySpending)
	r.Get("/municipalities/{code}/population-trend", h.getPopulationTrend)
}

func (h *Handler) listRegions(w http.ResponseWriter, r *http.Request) {
	regs, err := h.svc.ListRegions(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if regs == nil {
		jsonOK(w, []any{})
		return
	}
	jsonOK(w, regs)
}

func (h *Handler) getRegion(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	detail, err := h.svc.GetRegion(r.Context(), code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			jsonError(w, "region not found", http.StatusNotFound)
			return
		}
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, detail)
}

func (h *Handler) listMunicipalities(w http.ResponseWriter, r *http.Request) {
	regionCode := r.URL.Query().Get("regionCode")
	muns, err := h.svc.ListMunicipalities(r.Context(), regionCode)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if muns == nil {
		jsonOK(w, []any{})
		return
	}
	jsonOK(w, muns)
}

func (h *Handler) getMunicipality(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	detail, err := h.svc.GetMunicipality(r.Context(), code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			jsonError(w, "municipality not found", http.StatusNotFound)
			return
		}
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, detail)
}

func (h *Handler) getMunicipalityKPI(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	values, err := h.svc.GetMunicipalityKPIs(r.Context(), code)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	if values == nil {
		values = []ports.KPIValue{}
	}
	jsonOK(w, values)
}

func (h *Handler) getMunicipalitySpending(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	values, err := h.svc.GetMunicipalitySpending(r.Context(), code)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	if values == nil {
		values = []ports.KPIValue{}
	}
	jsonOK(w, values)
}

func (h *Handler) getPopulationTrend(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	entries, err := h.svc.GetPopulationTrend(r.Context(), code)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	if entries == nil {
		entries = []ports.PopulationEntry{}
	}
	jsonOK(w, entries)
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}
