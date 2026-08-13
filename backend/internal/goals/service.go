package goals

import (
	"context"

	"riksdagskollen/internal/goals/domain"
	"riksdagskollen/internal/goals/ports"
)

type Service struct {
	repo ports.GoalRepository
}

func NewService(repo ports.GoalRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, g *domain.Goal) error {
	return s.repo.Create(ctx, g)
}

func (s *Service) ListAll(ctx context.Context) ([]*domain.Goal, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int) (*domain.Goal, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByParty(ctx context.Context, party string) ([]*domain.Goal, error) {
	return s.repo.ListByParty(ctx, party)
}

func (s *Service) ListByTopic(ctx context.Context, party, topic string) ([]*domain.Goal, error) {
	return s.repo.ListByTopic(ctx, party, topic)
}

func (s *Service) ListByCommittee(ctx context.Context, code string) ([]*domain.Goal, error) {
	return s.repo.ListByCommittee(ctx, code)
}
