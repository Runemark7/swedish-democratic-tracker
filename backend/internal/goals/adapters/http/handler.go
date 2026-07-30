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
	type topicEntry struct {
		Topic        string   `json:"topic"`
		GoalCount    int      `json:"goalCount"`
		AlignmentPct *float64 `json:"alignmentPct"`
	}
	type partySummary struct {
		Party          string       `json:"party"`
		TotalGoals     int          `json:"totalGoals"`
		AvgAlignment   *float64     `json:"avgAlignmentPct"`
		TopicBreakdown []topicEntry `json:"topicBreakdown"`
	}

	allGoals, err := h.goalsSvc.ListAll(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	scorecards, err := h.matchingSvc.GetAllPartyScorecards(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Alignment per goal, present only where the direction was determinable.
	pctByGoal := make(map[int]float64, len(scorecards))
	for _, row := range scorecards {
		if row.AlignmentPct != nil {
			pctByGoal[row.GoalID] = *row.AlignmentPct
		}
	}

	parties := map[string]*partySummary{}
	topicCounts := map[string]map[string]int{}     // party -> topic -> goals
	topicPcts := map[string]map[string][]float64{} // party -> topic -> scored pcts

	// Every goal contributes exactly once. Goals whose direction could not be
	// determined count toward GoalCount but are excluded from the averages —
	// averaging them in as 0 would state "not in line" where we simply do not know.
	for _, g := range allGoals {
		if _, ok := parties[g.Party]; !ok {
			parties[g.Party] = &partySummary{Party: g.Party}
			topicCounts[g.Party] = map[string]int{}
			topicPcts[g.Party] = map[string][]float64{}
		}
		parties[g.Party].TotalGoals++
		topicCounts[g.Party][g.Topic]++
		if pct, ok := pctByGoal[g.ID]; ok {
			topicPcts[g.Party][g.Topic] = append(topicPcts[g.Party][g.Topic], pct)
		}
	}

	result := make([]*partySummary, 0, len(parties))
	for party, ps := range parties {
		var partySum float64
		var partyScored int
		for topic, count := range topicCounts[party] {
			entry := topicEntry{Topic: topic, GoalCount: count}
			if pcts := topicPcts[party][topic]; len(pcts) > 0 {
				var sum float64
				for _, p := range pcts {
					sum += p
				}
				avg := sum / float64(len(pcts))
				entry.AlignmentPct = &avg
				partySum += sum
				partyScored += len(pcts)
			}
			ps.TopicBreakdown = append(ps.TopicBreakdown, entry)
		}
		if partyScored > 0 {
			avg := partySum / float64(partyScored)
			ps.AvgAlignment = &avg
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
	RelevantVotes      int      `json:"relevantVotes"`
	ScoredVotes        int      `json:"scoredVotes"`
	AlignedVotes       int      `json:"alignedVotes"`
	AlignmentPct       *float64 `json:"alignmentPct"`
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
	type goalScore struct {
		relevant int
		scored   int
		aligned  int
		pct      *float64
	}
	scoreMap := map[int]goalScore{}
	for _, row := range rows {
		scoreMap[row.GoalID] = goalScore{
			relevant: row.RelevantVotes,
			scored:   row.ScoredVotes,
			aligned:  row.AlignedVotes,
			pct:      row.AlignmentPct,
		}
	}

	// Convert to response type with alignment info
	goalsSlice, ok := goals.([]*goaldomain.Goal)
	if !ok {
		jsonOK(w, goals)
		return
	}
	result := make([]goalWithAlignment, 0, len(goalsSlice))
	for _, g := range goalsSlice {
		sc := scoreMap[g.ID]
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
			RelevantVotes:      sc.relevant,
			ScoredVotes:        sc.scored,
			AlignedVotes:       sc.aligned,
			AlignmentPct:       sc.pct,
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
