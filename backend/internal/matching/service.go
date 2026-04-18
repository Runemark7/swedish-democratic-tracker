package matching

import (
	"context"

	"riksdagskollen/internal/matching/domain"
	"riksdagskollen/internal/matching/ports"
)

type Service struct {
	repo ports.MatchRepository
	ai   ports.AIService
}

func NewService(repo ports.MatchRepository, ai ports.AIService) *Service {
	return &Service{repo: repo, ai: ai}
}

func (s *Service) UpsertGoalVoteMatch(ctx context.Context, m *domain.GoalVoteMatch) error {
	return s.repo.UpsertGoalVoteMatch(ctx, m)
}

func (s *Service) GetPartyScorecard(ctx context.Context, party string) ([]*domain.ScorecardRow, error) {
	return s.repo.GetPartyScorecard(ctx, party)
}

func (s *Service) GetAllPartyScorecards(ctx context.Context) ([]*domain.ScorecardRow, error) {
	return s.repo.GetAllPartyScorecards(ctx)
}

func (s *Service) ListMatchesByGoal(ctx context.Context, goalID int) ([]*domain.GoalVoteMatch, error) {
	return s.repo.ListMatchesByGoal(ctx, goalID)
}

func (s *Service) ListMatchesByPromise(ctx context.Context, promiseID int) ([]*domain.PromiseVoteMatch, error) {
	return s.repo.ListMatchesByPromise(ctx, promiseID)
}

func (s *Service) RefreshScorecards(ctx context.Context) error {
	return s.repo.RefreshScorecards(ctx)
}
