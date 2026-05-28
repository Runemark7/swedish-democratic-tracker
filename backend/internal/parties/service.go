package parties

import (
	"context"

	"riksdagskollen/internal/parties/domain"
	"riksdagskollen/internal/parties/ports"
)

type Service struct {
	repo ports.PartiesRepository
}

func NewService(repo ports.PartiesRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListAll(ctx context.Context) ([]*domain.Party, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) GetByCode(ctx context.Context, code string) (*domain.Party, error) {
	return s.repo.GetByCode(ctx, code)
}
