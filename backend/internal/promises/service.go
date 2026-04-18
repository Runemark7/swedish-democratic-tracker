package promises

import (
	"context"

	"riksdagskollen/internal/promises/domain"
	"riksdagskollen/internal/promises/ports"
)

type Service struct {
	repo ports.PromiseRepository
}

func NewService(repo ports.PromiseRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, p *domain.Promise) error {
	return s.repo.Create(ctx, p)
}

func (s *Service) ListByPolitician(ctx context.Context, politicianID, topic string) ([]*domain.Promise, error) {
	return s.repo.ListByPolitician(ctx, politicianID, topic)
}

func (s *Service) GetByID(ctx context.Context, id int) (*domain.Promise, error) {
	return s.repo.GetByID(ctx, id)
}
