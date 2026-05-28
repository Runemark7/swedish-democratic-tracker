package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/parties"
)

type Handler struct {
	svc *parties.Service
}

func NewHandler(svc *parties.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/party-meta", h.listAll)
	r.Get("/party-meta/{code}", h.getByCode)
}

func (h *Handler) listAll(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListAll(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, list)
}

func (h *Handler) getByCode(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	p, err := h.svc.GetByCode(r.Context(), code)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if p == nil {
		jsonError(w, "party not found", http.StatusNotFound)
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
