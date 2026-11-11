// Package http serves frågor and interpellationer.
package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/questions"
	"riksdagskollen/internal/questions/ports"
)

type Handler struct{ svc *questions.Service }

func NewHandler(svc *questions.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes(r chi.Router) {
	r.Get("/questions/coverage", h.coverage)
	r.Get("/questions/{kind}/{riksmote}/{beteckning}", h.get)
	r.Get("/politicians/{id}/questions", h.byMember)
	r.Get("/motions/{riksmote}/{beteckning}/questions", h.byMotion)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, err error) {
	if errors.Is(err, ports.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}

func (h *Handler) coverage(w http.ResponseWriter, r *http.Request) {
	c, err := h.svc.Coverage(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"coverage": c})
}

func (h *Handler) byMember(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ByMember(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"questions": items})
}

// riksmote reads the riksmöte path segment; "2022%2F23" arrives escaped.
func riksmote(r *http.Request) string {
	rm, err := url.PathUnescape(chi.URLParam(r, "riksmote"))
	if err != nil {
		return chi.URLParam(r, "riksmote")
	}
	return rm
}

func (h *Handler) byMotion(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ByMotion(r.Context(), riksmote(r), chi.URLParam(r, "beteckning"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"questions": items})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	d, err := h.svc.Get(r.Context(), chi.URLParam(r, "kind"), riksmote(r), chi.URLParam(r, "beteckning"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}
