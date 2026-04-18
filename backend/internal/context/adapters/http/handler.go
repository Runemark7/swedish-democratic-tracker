package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/context"
)

type Handler struct {
	svc *context.Service
}

func NewHandler(svc *context.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/context/topic/{topic}", h.getTopicContext)
}

func (h *Handler) getTopicContext(w http.ResponseWriter, r *http.Request) {
	topic := chi.URLParam(r, "topic")

	result, err := h.svc.GetTopicContext(r.Context(), topic)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result == nil {
		jsonError(w, "unknown topic", http.StatusNotFound)
		return
	}
	jsonOK(w, result)
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
