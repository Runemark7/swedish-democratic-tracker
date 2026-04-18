package ports

import (
	"context"

	goalsdomain "riksdagskollen/internal/goals/domain"
	matchdomain "riksdagskollen/internal/matching/domain"
	speechdomain "riksdagskollen/internal/speeches/domain"
	votesdomain "riksdagskollen/internal/votes/domain"
)

// AIService is the port for AI-powered extraction and scoring.
// The stub implementation returns empty results; the real Claude-backed
// implementation will be wired in a later iteration.
type AIService interface {
	// ExtractPromises analyses a speech and returns concrete commitments found in it.
	ExtractPromises(ctx context.Context, speech speechdomain.Speech) ([]matchdomain.ExtractedPromise, error)

	// ScoreVoteRelevance determines whether a vote is relevant to a goal and
	// which direction of voting aligns with that goal.
	ScoreVoteRelevance(ctx context.Context, goal goalsdomain.Goal, vote votesdomain.Vote) (*matchdomain.RelevanceScore, error)
}
