// Package stub provides a no-op AIService implementation.
// Replace this with the Claude-backed adapter when ready.
package stub

import (
	"context"

	goalsdomain "riksdagskollen/internal/goals/domain"
	matchdomain "riksdagskollen/internal/matching/domain"
	speechdomain "riksdagskollen/internal/speeches/domain"
	votesdomain "riksdagskollen/internal/votes/domain"
)

type AIService struct{}

func NewAIService() *AIService { return &AIService{} }

func (s *AIService) ExtractPromises(_ context.Context, _ speechdomain.Speech) ([]matchdomain.ExtractedPromise, error) {
	// TODO: implement Claude API call for promise extraction
	return nil, nil
}

func (s *AIService) ScoreVoteRelevance(_ context.Context, _ goalsdomain.Goal, _ votesdomain.Vote) (*matchdomain.RelevanceScore, error) {
	// TODO: implement Claude API call for vote-relevance scoring
	return &matchdomain.RelevanceScore{
		Relevant:         false,
		AlignedDirection: matchdomain.AlignmentUnclear,
		Confidence:       0,
		ContextNote:      "",
	}, nil
}
