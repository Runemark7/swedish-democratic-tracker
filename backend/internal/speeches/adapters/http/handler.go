package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/speeches"
	"riksdagskollen/internal/speeches/domain"
)

// PoliticianLookup resolves intressent_id → display name. The speeches
// handler depends on this minimal surface (not the full politicians
// service) to stay decoupled.
type PoliticianLookup interface {
	NameByID(ctx context.Context, intressentID string) (string, error)
}

type Handler struct {
	svc *speeches.Service
	pol PoliticianLookup
}

func NewHandler(svc *speeches.Service, pol PoliticianLookup) *Handler {
	return &Handler{svc: svc, pol: pol}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/speeches/recent", h.listRecent)
	r.Get("/speeches/{id}", h.getByID)
}

// SpeechDTO is the JSON shape returned to the frontend.
type SpeechDTO struct {
	ID              int       `json:"id"`
	DokID           string    `json:"dokId"`
	AnforandeNummer string    `json:"anforandeNummer"`
	PoliticianID    string    `json:"politicianId"`
	PoliticianName  string    `json:"politicianName"`
	Party           string    `json:"party"`
	Date            time.Time `json:"date"`
	TopicHeading    string    `json:"topicHeading"`
	Snippet         string    `json:"snippet"`
	SpeechText      string    `json:"speechText,omitempty"`
}

func (h *Handler) listRecent(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			limit = n
		}
	}
	ss, err := h.svc.ListRecent(r.Context(), limit)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]SpeechDTO, 0, len(ss))
	for _, s := range ss {
		out = append(out, h.toDTO(r.Context(), s, false))
	}
	jsonOK(w, out)
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	s, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s == nil {
		jsonError(w, "speech not found", http.StatusNotFound)
		return
	}
	jsonOK(w, h.toDTO(r.Context(), s, true))
}

func (h *Handler) toDTO(ctx context.Context, s *domain.Speech, includeFullText bool) SpeechDTO {
	name, _ := h.pol.NameByID(ctx, s.PoliticianID)
	dto := SpeechDTO{
		ID:              s.ID,
		DokID:           s.DokID,
		AnforandeNummer: s.AnforandeNummer,
		PoliticianID:    s.PoliticianID,
		PoliticianName:  name,
		Party:           s.Party,
		Date:            s.Date,
		TopicHeading:    s.TopicHeading,
		Snippet:         snippetFrom(s.SpeechText, 180),
	}
	if includeFullText {
		dto.SpeechText = s.SpeechText
	}
	return dto
}

func snippetFrom(text string, maxRunes int) string {
	if text == "" {
		return ""
	}
	r := []rune(text)
	if len(r) <= maxRunes {
		return text
	}
	return string(r[:maxRunes]) + "…"
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
