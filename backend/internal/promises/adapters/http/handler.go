package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/matching"
	"riksdagskollen/internal/promises"
)

type Handler struct {
	promiseSvc  *promises.Service
	matchingSvc *matching.Service
}

func NewHandler(promiseSvc *promises.Service, matchingSvc *matching.Service) *Handler {
	return &Handler{promiseSvc: promiseSvc, matchingSvc: matchingSvc}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/politicians/{id}/promises", h.list)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	topic := r.URL.Query().Get("topic")

	pp, err := h.promiseSvc.ListByPolitician(r.Context(), id, topic)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Enrich each promise with its vote matches
	type promiseWithMatches struct {
		ID           int    `json:"id"`
		PoliticianID string `json:"politicianId"`
		SpeechID     int    `json:"speechId"`
		PromiseText  string `json:"promiseText"`
		Topic        string `json:"topic"`
		Specificity  string `json:"specificity"`
		Keywords     []string `json:"keywords"`
		VoteMatches  any    `json:"voteMatches"`
	}

	result := make([]promiseWithMatches, 0, len(pp))
	for _, p := range pp {
		matches, _ := h.matchingSvc.ListMatchesByPromise(r.Context(), p.ID)
		result = append(result, promiseWithMatches{
			ID:           p.ID,
			PoliticianID: p.PoliticianID,
			SpeechID:     p.SpeechID,
			PromiseText:  p.PromiseText,
			Topic:        p.Topic,
			Specificity:  string(p.Specificity),
			Keywords:     p.Keywords,
			VoteMatches:  matches,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
