// Package http serves the Data section's coverage.
package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/datacoverage"
)

type Handler struct{ svc *datacoverage.Service }

func NewHandler(svc *datacoverage.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes(r chi.Router) { r.Get("/data/coverage", h.coverage) }

type row struct {
	Source     string `json:"source"`
	Riksmote   string `json:"riksmote"`
	Status     string `json:"status"`
	Expected   int    `json:"expected"`
	Held       int    `json:"held"`
	LastUpdate string `json:"lastUpdate"`
}

func (h *Handler) coverage(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.Coverage(r.Context())
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
		return
	}
	out := make([]row, 0, len(rows))
	for _, c := range rows {
		out = append(out, row{c.Source, c.Riksmote, c.Status, c.Expected, c.Held, c.LastUpdate.UTC().Format(time.RFC3339)})
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
}
