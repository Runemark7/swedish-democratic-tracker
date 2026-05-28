package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

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
	r.Get("/regions/budget/area/{areaName}", h.getAreaAcrossRegions)
	r.Get("/regions/{code}", h.getRegion)
	r.Get("/regions/{code}/budget", h.getRegionBudget)
	r.Get("/regions/{code}/budget/history", h.getRegionBudgetHistory)
	r.Get("/regions/{code}/kpi", h.getRegionKPI)
	r.Get("/municipalities", h.listMunicipalities)
	r.Get("/municipalities/budget/area/{areaName}", h.getAreaAcrossMunicipalities)
	r.Get("/municipalities/kpi/{kpiCode}/ranking", h.getMunicipalityKPIRanking)
	r.Get("/municipalities/{code}", h.getMunicipality)
	r.Get("/municipalities/{code}/kpi", h.getMunicipalityKPI)
	r.Get("/municipalities/{code}/kpi-ranks", h.getMunicipalityKPIRanks)
	r.Get("/municipalities/{code}/spending", h.getMunicipalitySpending)
	r.Get("/municipalities/{code}/budget/history", h.getMunicipalityBudgetHistory)
	r.Get("/municipalities/{code}/population-trend", h.getPopulationTrend)
	r.Get("/municipalities/{code}/procurement", h.getMunicipalityProcurement)
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

func (h *Handler) getRegionKPI(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	values, err := h.svc.GetRegionKPIs(r.Context(), code)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	if values == nil {
		values = []ports.KPIValue{}
	}
	jsonOK(w, values)
}

func (h *Handler) getRegionBudget(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	// SCB publishes the previous year's data; try year-1, then walk back
	// up to 4 years. Some regions lag (e.g. preliminary numbers ".." for
	// recent years).
	startYear := time.Now().Year() - 1
	var (
		areas   []ports.RegionBudgetArea
		lastErr error
	)
	for y := startYear; y >= startYear-4; y-- {
		areas, lastErr = h.svc.GetRegionBudget(r.Context(), code, y)
		if lastErr == nil && len(areas) > 0 {
			break
		}
	}
	if lastErr != nil && len(areas) == 0 {
		jsonError(w, lastErr.Error(), http.StatusBadGateway)
		return
	}
	if areas == nil {
		areas = []ports.RegionBudgetArea{}
	}
	jsonOK(w, areas)
}

func (h *Handler) getRegionBudgetHistory(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	yearsParam := r.URL.Query().Get("years")
	var years []int
	if yearsParam != "" {
		if n, err := strconv.Atoi(yearsParam); err == nil && n > 0 {
			current := time.Now().Year()
			years = make([]int, n)
			for i := range years {
				years[i] = current - 1 - i
			}
		}
	}
	snapshots, err := h.svc.GetRegionBudgetHistory(r.Context(), code, years)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if snapshots == nil {
		snapshots = []ports.RegionBudgetSnapshot{}
	}
	jsonOK(w, snapshots)
}

func (h *Handler) getAreaAcrossRegions(w http.ResponseWriter, r *http.Request) {
	areaName := chi.URLParam(r, "areaName")
	yearParam := r.URL.Query().Get("year")
	year := 0 // 0 = let repo pick the most complete year
	if yearParam != "" {
		if parsed, err := strconv.Atoi(yearParam); err == nil {
			year = parsed
		}
	}
	points, err := h.svc.GetAreaAcrossRegions(r.Context(), areaName, year)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if points == nil {
		points = []ports.RegionAreaDataPoint{}
	}
	jsonOK(w, points)
}

func (h *Handler) getAreaAcrossMunicipalities(w http.ResponseWriter, r *http.Request) {
	areaName := chi.URLParam(r, "areaName")
	yearParam := r.URL.Query().Get("year")
	year := 0
	if yearParam != "" {
		if parsed, err := strconv.Atoi(yearParam); err == nil {
			year = parsed
		}
	}
	points, err := h.svc.GetAreaAcrossMunicipalities(r.Context(), areaName, year)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if points == nil {
		points = []ports.MunicipalityAreaDataPoint{}
	}
	jsonOK(w, points)
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

func (h *Handler) getMunicipalityProcurement(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	result, err := h.svc.GetMunicipalityProcurement(r.Context(), code)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	if result == nil {
		result = []ports.ProcurementCategorySummary{}
	}
	jsonOK(w, result)
}

func (h *Handler) getMunicipalityBudgetHistory(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	yearsParam := r.URL.Query().Get("years")
	var years []int
	if yearsParam != "" {
		if n, err := strconv.Atoi(yearsParam); err == nil && n > 0 {
			current := time.Now().Year()
			years = make([]int, n)
			for i := range years {
				years[i] = current - 1 - i
			}
		}
	}
	snapshots, err := h.svc.GetMunicipalityBudgetHistory(r.Context(), code, years)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if snapshots == nil {
		snapshots = []ports.MunicipalityBudgetSnapshot{}
	}
	jsonOK(w, snapshots)
}

func (h *Handler) getMunicipalityKPIRanking(w http.ResponseWriter, r *http.Request) {
	kpiCode := chi.URLParam(r, "kpiCode")
	entries, err := h.svc.GetKPIRanking(r.Context(), kpiCode)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	if entries == nil {
		entries = []ports.KPIRankEntry{}
	}
	jsonOK(w, entries)
}

func (h *Handler) getMunicipalityKPIRanks(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	ranks, err := h.svc.GetMunicipalityKPIRanks(r.Context(), code)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	if ranks == nil {
		ranks = []ports.KPIRank{}
	}
	jsonOK(w, ranks)
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
