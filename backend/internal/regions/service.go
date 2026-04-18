package regions

import (
	"context"

	"riksdagskollen/internal/regions/domain"
	"riksdagskollen/internal/regions/ports"
)

type Service struct {
	repo ports.RegionRepository
}

func NewService(repo ports.RegionRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListRegions(ctx context.Context) ([]*domain.Region, error) {
	return s.repo.ListRegions(ctx)
}

func (s *Service) GetRegion(ctx context.Context, code string) (*domain.RegionDetail, error) {
	return s.repo.GetRegion(ctx, code)
}

func (s *Service) ListMunicipalities(ctx context.Context, regionCode string) ([]*domain.Municipality, error) {
	return s.repo.ListMunicipalities(ctx, regionCode)
}

func (s *Service) GetMunicipality(ctx context.Context, code string) (*domain.MunicipalityDetail, error) {
	return s.repo.GetMunicipality(ctx, code)
}
