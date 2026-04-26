package http

import (
	"encoding/json"
	"errors"
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
	r.Get("/riksdag/authorities/{slug}", h.getAuthority)
}

func (h *Handler) getAuthorities(w http.ResponseWriter, r *http.Request) {
	authorities, err := h.svc.GetAuthorities(r.Context())
	if err != nil {
		jsonError(w, "failed to fetch authorities", http.StatusBadGateway)
		return
	}
	jsonOK(w, authorities)
}

func (h *Handler) getAuthority(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	detail, err := h.svc.GetAuthority(r.Context(), slug)
	if err != nil {
		if errors.Is(err, riksdag.ErrNotFound) {
			jsonError(w, "authority not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to fetch authority", http.StatusBadGateway)
		return
	}
	jsonOK(w, detail)
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
