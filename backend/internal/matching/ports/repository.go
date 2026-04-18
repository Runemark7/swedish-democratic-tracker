package ports

import (
	"context"

	"riksdagskollen/internal/matching/domain"
)

type MatchRepository interface {
	UpsertGoalVoteMatch(ctx context.Context, m *domain.GoalVoteMatch) error
	ListMatchesByGoal(ctx context.Context, goalID int) ([]*domain.GoalVoteMatch, error)
	UpsertPromiseVoteMatch(ctx context.Context, m *domain.PromiseVoteMatch) error
	ListMatchesByPromise(ctx context.Context, promiseID int) ([]*domain.PromiseVoteMatch, error)
	GetPartyScorecard(ctx context.Context, party string) ([]*domain.ScorecardRow, error)
	GetAllPartyScorecards(ctx context.Context) ([]*domain.ScorecardRow, error)
	RefreshScorecards(ctx context.Context) error
}
