package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/goals"
	goaldomain "riksdagskollen/internal/goals/domain"
	"riksdagskollen/internal/matching"
)

type Handler struct {
	goalsSvc    *goals.Service
	matchingSvc *matching.Service
}

func NewHandler(goalsSvc *goals.Service, matchingSvc *matching.Service) *Handler {
	return &Handler{goalsSvc: goalsSvc, matchingSvc: matchingSvc}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/parties", h.listParties)
	r.Get("/parties/{party}/goals", h.listGoals)
	r.Get("/parties/{party}/goals/{goalId}/votes", h.goalVotes)
}

func (h *Handler) listParties(w http.ResponseWriter, r *http.Request) {
	// Aggregate by party
	// No alignment figure: the computed percentage was retired (migration
	// 000034) because it systematically disadvantaged opposition parties.
	type topicEntry struct {
		Topic     string `json:"topic"`
		GoalCount int    `json:"goalCount"`
	}
	type partySummary struct {
		Party          string       `json:"party"`
		TotalGoals     int          `json:"totalGoals"`
		TopicBreakdown []topicEntry `json:"topicBreakdown"`
	}

	allGoals, err := h.goalsSvc.ListAll(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parties := map[string]*partySummary{}
	topicCounts := map[string]map[string]int{} // party -> topic -> goals

	// Every goal contributes exactly once.
	for _, g := range allGoals {
		if _, ok := parties[g.Party]; !ok {
			parties[g.Party] = &partySummary{Party: g.Party}
			topicCounts[g.Party] = map[string]int{}
		}
		parties[g.Party].TotalGoals++
		topicCounts[g.Party][g.Topic]++
	}

	result := make([]*partySummary, 0, len(parties))
	for party, ps := range parties {
		for topic, count := range topicCounts[party] {
			ps.TopicBreakdown = append(ps.TopicBreakdown, topicEntry{Topic: topic, GoalCount: count})
		}
		result = append(result, ps)
	}
	jsonOK(w, result)
}

type goalWithAlignment struct {
	ID                 int      `json:"id"`
	Party              string   `json:"party"`
	GoalText           string   `json:"goalText"`
	Topic              string   `json:"topic"`
	Specificity        string   `json:"specificity"`
	SourceDocument     string   `json:"sourceDocument"`
	SourceURL          string   `json:"sourceUrl,omitempty"`
	SourceQuote        string   `json:"sourceQuote,omitempty"`
	Keywords           []string `json:"keywords,omitempty"`
	RelevantCommittees []string `json:"relevantCommittees,omitempty"`
	// RelevantVotes is how many vote points matched this goal — a fact about
	// our matching. The alignment percentage was retired (migration 000034).
	RelevantVotes int `json:"relevantVotes"`
}

func (h *Handler) listGoals(w http.ResponseWriter, r *http.Request) {
	party := chi.URLParam(r, "party")
	topic := r.URL.Query().Get("topic")

	var goals any
	var err error
	if topic != "" {
		goals, err = h.goalsSvc.ListByTopic(r.Context(), party, topic)
	} else {
		goals, err = h.goalsSvc.ListByParty(r.Context(), party)
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Enrich goals with scorecard alignment data
	rows, err := h.matchingSvc.GetPartyScorecard(r.Context(), party)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	relevantByGoal := map[int]int{}
	for _, row := range rows {
		relevantByGoal[row.GoalID] = row.RelevantVotes
	}

	// Convert to response type with alignment info
	goalsSlice, ok := goals.([]*goaldomain.Goal)
	if !ok {
		jsonOK(w, goals)
		return
	}
	result := make([]goalWithAlignment, 0, len(goalsSlice))
	for _, g := range goalsSlice {
		result = append(result, goalWithAlignment{
			ID:                 g.ID,
			Party:              g.Party,
			GoalText:           g.GoalText,
			Topic:              g.Topic,
			Specificity:        string(g.Specificity),
			SourceDocument:     g.SourceDocument,
			SourceURL:          g.SourceURL,
			SourceQuote:        g.SourceQuote,
			Keywords:           g.Keywords,
			RelevantCommittees: g.RelevantCommittees,
			RelevantVotes:      relevantByGoal[g.ID],
		})
	}
	jsonOK(w, result)
}

func (h *Handler) goalVotes(w http.ResponseWriter, r *http.Request) {
	goalID, err := strconv.Atoi(chi.URLParam(r, "goalId"))
	if err != nil {
		jsonError(w, "invalid goal id", http.StatusBadRequest)
		return
	}

	goal, err := h.goalsSvc.GetByID(r.Context(), goalID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if goal == nil {
		jsonError(w, "goal not found", http.StatusNotFound)
		return
	}

	matches, err := h.matchingSvc.ListMatchesByGoal(r.Context(), goalID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]any{"goal": goal, "matches": matches})
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
