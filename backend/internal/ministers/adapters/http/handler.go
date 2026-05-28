package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/ministers"
	"riksdagskollen/internal/ministers/domain"
)

type Handler struct {
	svc *ministers.Service
}

func NewHandler(svc *ministers.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/ministers", h.list)
	r.Get("/ministers/{id}", h.getDetail)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListActive(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type deptGroup struct {
		Department     string             `json:"department"`
		DepartmentCode string             `json:"departmentCode"`
		Ministers      []*domain.Minister `json:"ministers"`
	}
	order := []string{}
	groups := map[string]*deptGroup{}
	for _, m := range list {
		if _, ok := groups[m.DepartmentCode]; !ok {
			groups[m.DepartmentCode] = &deptGroup{
				Department:     m.Department,
				DepartmentCode: m.DepartmentCode,
			}
			order = append(order, m.DepartmentCode)
		}
		groups[m.DepartmentCode].Ministers = append(groups[m.DepartmentCode].Ministers, m)
	}

	result := make([]*deptGroup, 0, len(order))
	for _, code := range order {
		result = append(result, groups[code])
	}
	jsonOK(w, result)
}

func (h *Handler) getDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	m, proposals, err := h.svc.GetDetail(r.Context(), id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if m == nil {
		jsonError(w, "minister not found", http.StatusNotFound)
		return
	}
	jsonOK(w, map[string]any{
		"minister":  m,
		"proposals": proposals,
	})
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
