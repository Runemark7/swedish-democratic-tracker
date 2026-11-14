package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/politicians"
	"riksdagskollen/internal/speeches"
	"riksdagskollen/internal/speeches/domain"
)

// PoliticianLookup resolves intressent_id → display name. The speeches
// handler depends on this minimal surface (not the full politicians
// service) to stay decoupled.
type PoliticianLookup interface {
	NameByID(ctx context.Context, intressentID string) (string, error)
	ImageURLByID(ctx context.Context, intressentID string) (string, error)
}

// BatchPoliticianLookup resolves many ids in one query. A PoliticianLookup
// that also implements it avoids one lookup per speech in the list endpoints.
type BatchPoliticianLookup interface {
	IdentitiesByID(ctx context.Context, ids []string) (map[string]politicians.Identity, error)
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
	r.Get("/speeches/backfill", h.backfill)
	r.Get("/speeches/{id}", h.getByID)
	r.Get("/speeches/by-document/{dokId}", h.listByDocument)
	r.Get("/speeches/by-politician/{intressentId}", h.listByPolitician)
	r.Get("/speeches/by-party/{party}", h.listByParty)
}

// SpeechDTO is the JSON shape returned to the frontend.
type SpeechDTO struct {
	ID                 int       `json:"id"`
	DokID              string    `json:"dokId"`
	AnforandeNummer    string    `json:"anforandeNummer"`
	PoliticianID       string    `json:"politicianId"`
	PoliticianName     string    `json:"politicianName"`
	PoliticianImageURL string    `json:"politicianImageUrl,omitempty"`
	Party              string    `json:"party"`
	Date               time.Time `json:"date"`
	TopicHeading       string    `json:"topicHeading"`
	Snippet            string    `json:"snippet"`
	SpeechText         string    `json:"speechText,omitempty"`
	RelatedDokID       string    `json:"relatedDokId,omitempty"`
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
	jsonOK(w, h.toDTOs(r.Context(), ss))
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

func (h *Handler) listByDocument(w http.ResponseWriter, r *http.Request) {
	dokID := chi.URLParam(r, "dokId")
	ss, err := h.svc.ListByDocument(r.Context(), dokID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, h.toDTOs(r.Context(), ss))
}

func (h *Handler) listByPolitician(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "intressentId")
	ss, err := h.svc.ListByPolitician(r.Context(), id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	limit := 20
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if len(ss) > limit {
		ss = ss[:limit]
	}
	jsonOK(w, h.toDTOs(r.Context(), ss))
}

func (h *Handler) listByParty(w http.ResponseWriter, r *http.Request) {
	party := chi.URLParam(r, "party")
	limit := 50
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			limit = n
		}
	}
	ss, err := h.svc.ListByParty(r.Context(), party, limit)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, h.toDTOs(r.Context(), ss))
}

// toDTOs builds list DTOs, resolving every politician in one query when the
// lookup supports it (a failed batch leaves names empty, as a failed single
// lookup does).
func (h *Handler) toDTOs(ctx context.Context, ss []*domain.Speech) []SpeechDTO {
	out := make([]SpeechDTO, 0, len(ss))
	b, ok := h.pol.(BatchPoliticianLookup)
	if !ok {
		for _, s := range ss {
			out = append(out, h.toDTO(ctx, s, false))
		}
		return out
	}
	seen := map[string]bool{}
	var ids []string
	for _, s := range ss {
		if !seen[s.PoliticianID] {
			seen[s.PoliticianID] = true
			ids = append(ids, s.PoliticianID)
		}
	}
	idents, _ := b.IdentitiesByID(ctx, ids)
	for _, s := range ss {
		id := idents[s.PoliticianID]
		out = append(out, buildDTO(s, id.Name, id.ImageURL, false))
	}
	return out
}

func (h *Handler) toDTO(ctx context.Context, s *domain.Speech, includeFullText bool) SpeechDTO {
	name, _ := h.pol.NameByID(ctx, s.PoliticianID)
	imageURL, _ := h.pol.ImageURLByID(ctx, s.PoliticianID)
	return buildDTO(s, name, imageURL, includeFullText)
}

func buildDTO(s *domain.Speech, name, imageURL string, includeFullText bool) SpeechDTO {
	text := s.SpeechText
	if strings.TrimSpace(text) == "" {
		text = ""
	}
	dto := SpeechDTO{
		ID:                 s.ID,
		DokID:              s.DokID,
		AnforandeNummer:    s.AnforandeNummer,
		PoliticianID:       s.PoliticianID,
		PoliticianName:     name,
		PoliticianImageURL: imageURL,
		Party:              s.Party,
		Date:               s.Date,
		TopicHeading:       s.TopicHeading,
		Snippet:            snippetFrom(stripHTML(text), 180),
	}
	dto.RelatedDokID = s.RelatedDokID
	if includeFullText {
		dto.SpeechText = text
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

// stripHTML removes <tag>...</tag> markup so the result reads as plain
// prose. Mirrors the helper in votes/adapters/riksdagen/client.go;
// kept inline to avoid cross-package coupling on a 15-line utility.
func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
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

type backfillResponse struct {
	Riksmote            string  `json:"riksmote"`
	Expected            int     `json:"expected"`
	Stored              int     `json:"stored"`
	Linked              int     `json:"linked"`
	Unlinked            int     `json:"unlinked"`
	TextBytes           int64   `json:"textBytes"`
	RowsAdded           int     `json:"rowsAdded"`
	TableBytesAdded     int64   `json:"tableBytesAdded"`
	IndexBytesAdded     int64   `json:"indexBytesAdded"`
	PoliticiansCreated  int     `json:"politiciansCreated"`
	UnresolvedSpeakers  int     `json:"unresolvedSpeakers"`
	SpeakerlessSpeeches int     `json:"speakerlessSpeeches"`
	Complete            bool    `json:"complete"`
	CompletedAt         *string `json:"completedAt"`
}

// backfill serves the coverage of the operator's speech backfill per riksmöte,
// so the Data section can show what is read in: speeches the lists report,
// stored, linked to a betänkande, and not linked.
func (h *Handler) backfill(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.ListBackfill(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]backfillResponse, 0, len(rows))
	for _, b := range rows {
		resp := backfillResponse{
			Riksmote: b.Riksmote, Expected: b.Expected, Stored: b.Stored, Linked: b.Linked, Unlinked: b.Unlinked,
			TextBytes: b.TextBytes, RowsAdded: b.RowsAdded, TableBytesAdded: b.TableBytesAdded,
			IndexBytesAdded: b.IndexBytesAdded, PoliticiansCreated: b.PoliticiansCreated, UnresolvedSpeakers: b.Unresolved, SpeakerlessSpeeches: b.Speakerless,
		}
		if b.CompletedAt != nil {
			resp.Complete = true
			t := b.CompletedAt.UTC().Format(time.RFC3339)
			resp.CompletedAt = &t
		}
		out = append(out, resp)
	}
	jsonOK(w, map[string]any{"data": out})
}
