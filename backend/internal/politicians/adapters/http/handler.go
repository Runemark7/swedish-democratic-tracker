package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/politicians"
	"riksdagskollen/internal/politicians/ports"
)

type Handler struct {
	svc *politicians.Service
}

func NewHandler(svc *politicians.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/politicians", h.list)
	r.Get("/politicians/{id}", h.getByID)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	f := ports.ListFilter{
		Party:      r.URL.Query().Get("party"),
		ActiveOnly: r.URL.Query().Get("active") != "false",
		Page:       queryInt(r, "page", 1),
		PageSize:   queryInt(r, "pageSize", 50),
	}

	result, err := h.svc.List(r.Context(), f)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]any{
		"data":     result.Politicians,
		"total":    result.Total,
		"page":     f.Page,
		"pageSize": f.PageSize,
	})
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if p == nil {
		jsonError(w, "politician not found", http.StatusNotFound)
		return
	}
	jsonOK(w, p)
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

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return def
	}
	return n
}
