package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
	r.Get("/riksdag/kpis", h.getKpis)
	r.Get("/riksdag/government", h.getGovernment)
	r.Get("/riksdag/agenda", h.getAgenda)
	r.Get("/riksdag/agenda/{id}", h.getAgendaItem)
	r.Get("/riksdag/live-votes", h.getLiveVotes)
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

func (h *Handler) getKpis(w http.ResponseWriter, r *http.Request) {
	kpis, err := h.svc.ListKpis(r.Context())
	if err != nil {
		jsonError(w, "failed to fetch kpis", http.StatusInternalServerError)
		return
	}
	jsonOK(w, kpis)
}

func (h *Handler) getGovernment(w http.ResponseWriter, r *http.Request) {
	gov, err := h.svc.GetGovernment(r.Context())
	if err != nil {
		jsonError(w, "failed to fetch government", http.StatusInternalServerError)
		return
	}
	jsonOK(w, gov)
}

func (h *Handler) getAgenda(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListAgenda(r.Context())
	if err != nil {
		jsonError(w, "failed to fetch agenda", http.StatusInternalServerError)
		return
	}
	jsonOK(w, items)
}

func (h *Handler) getAgendaItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		jsonError(w, "invalid agenda id", http.StatusBadRequest)
		return
	}

	item, err := h.svc.GetAgendaItem(r.Context(), id)
	if err != nil {
		if errors.Is(err, riksdag.ErrNotFound) {
			jsonError(w, "agenda item not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to fetch agenda item", http.StatusInternalServerError)
		return
	}
	jsonOK(w, item)
}

func (h *Handler) getLiveVotes(w http.ResponseWriter, r *http.Request) {
	votes, err := h.svc.ListLiveVotes(r.Context(), 10)
	if err != nil {
		jsonError(w, "failed to fetch live votes", http.StatusInternalServerError)
		return
	}
	jsonOK(w, votes)
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
