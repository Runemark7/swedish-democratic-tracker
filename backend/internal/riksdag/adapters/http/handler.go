package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/riksdag"
)

type Handler struct {
	svc *riksdag.Service
}

func NewHandler(svc *riksdag.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/riksdag/authorities", h.getAuthorities)
}

func (h *Handler) getAuthorities(w http.ResponseWriter, r *http.Request) {
	authorities, err := h.svc.GetAuthorities(r.Context())
	if err != nil {
		jsonError(w, "failed to fetch authorities", http.StatusBadGateway)
		return
	}
	jsonOK(w, authorities)
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
