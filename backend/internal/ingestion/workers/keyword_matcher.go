package workers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"riksdagskollen/internal/goals"
	goalsdomain "riksdagskollen/internal/goals/domain"
	"riksdagskollen/internal/matching"
	matchdomain "riksdagskollen/internal/matching/domain"
	"riksdagskollen/internal/votes"
)

type KeywordMatcherWorker struct {
	goals    *goals.Service
	votes    *votes.Service
	matching *matching.Service
}

func NewKeywordMatcherWorker(g *goals.Service, v *votes.Service, m *matching.Service) KeywordMatcherWorker {
	return KeywordMatcherWorker{goals: g, votes: v, matching: m}
}

func (w *KeywordMatcherWorker) Name() string { return "keyword-matcher" }

func (w *KeywordMatcherWorker) Run(ctx context.Context) error {
	allGoals, err := w.goals.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("list goals: %w", err)
	}

	matched := 0
	for _, goal := range allGoals {
		n, err := w.matchGoal(ctx, goal)
		if err != nil {
			slog.Warn("keyword matcher: goal failed", "goalId", goal.ID, "error", err)
			continue
		}
		matched += n
	}

	slog.Info("keyword matcher done", "goals", len(allGoals), "matchesCreated", matched)

	if matched > 0 {
		if err := w.matching.RefreshScorecards(ctx); err != nil {
			return fmt.Errorf("refresh scorecards: %w", err)
		}
	}
	return nil
}

func (w *KeywordMatcherWorker) matchGoal(ctx context.Context, goal *goalsdomain.Goal) (int, error) {
	if len(goal.RelevantCommittees) == 0 || len(goal.Keywords) == 0 {
		return 0, nil
	}

	matched := 0
	for _, committee := range goal.RelevantCommittees {
		summaries, err := w.votes.ListDistinctByCommitteePrefix(ctx, committee)
		if err != nil {
			return matched, fmt.Errorf("list votes for committee %s: %w", committee, err)
		}

		for _, vs := range summaries {
			hits := countKeywordHits(goal.Keywords, vs.DocumentTitle)
			if hits == 0 {
				continue
			}

			relevance := scoreRelevance(hits)
			direction := inferDirection(goal.Party, vs.ProposedByParty)

			matchedKeywords := findMatchedKeywords(goal.Keywords, vs.DocumentTitle)

			m := &matchdomain.GoalVoteMatch{
				GoalID:           goal.ID,
				Beteckning:       vs.Beteckning,
				Forslagspunkt:    vs.Forslagspunkt,
				RelevanceScore:   relevance,
				AlignedDirection: direction,
				Explanation:      fmt.Sprintf("Matchat via utskott %s och nyckelord: %s", committee, strings.Join(matchedKeywords, ", ")),
				ProposedByParty:  vs.ProposedByParty,
				ProposalType:     vs.ProposalType,
				DocumentTitle:    vs.DocumentTitle,
			}

			if err := w.matching.UpsertGoalVoteMatch(ctx, m); err != nil {
				slog.Warn("keyword matcher: upsert failed", "goalId", goal.ID, "bet", vs.Beteckning, "error", err)
				continue
			}
			matched++
		}
	}
	return matched, nil
}

func countKeywordHits(keywords []string, title string) int {
	lower := strings.ToLower(title)
	hits := 0
	for _, kw := range keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			hits++
		}
	}
	return hits
}

func findMatchedKeywords(keywords []string, title string) []string {
	lower := strings.ToLower(title)
	var matched []string
	for _, kw := range keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			matched = append(matched, kw)
		}
	}
	return matched
}

func scoreRelevance(hits int) float64 {
	switch {
	case hits >= 3:
		return 0.85
	case hits == 2:
		return 0.75
	default:
		return 0.60
	}
}

// inferDirection determines which vote direction aligns with the goal.
// If the proposal came from the same party or from "Regeringen", voting Ja aligns.
// For opposition motions, the goal's party would typically vote Nej.
func inferDirection(goalParty, proposedBy string) matchdomain.AlignmentDirection {
	if proposedBy == "" {
		return matchdomain.AlignmentUnclear
	}

	// Government parties in the Tidöavtalet coalition (2022–)
	govParties := map[string]bool{"M": true, "KD": true, "L": true, "SD": true}

	if proposedBy == goalParty {
		return matchdomain.AlignmentJa
	}
	if proposedBy == "Regeringen" {
		if govParties[goalParty] {
			return matchdomain.AlignmentJa
		}
		// Opposition parties typically vote Nej on government proposals
		return matchdomain.AlignmentNej
	}
	// Motion from another party — the goal's party may vote Nej
	return matchdomain.AlignmentUnclear
}
