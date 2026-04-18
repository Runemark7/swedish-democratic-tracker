package regions

import (
	"context"

	"riksdagskollen/internal/regions/domain"
	"riksdagskollen/internal/regions/ports"
)

var defaultKPIs = []string{"N00902", "N03005", "N00941", "N15033", "N07402"}
var defaultKPIYears = []int{2022, 2023}
var defaultPopYears = []int{2019, 2020, 2021, 2022, 2023}

type Service struct {
	repo   ports.RegionRepository
	kolada ports.KoladaClient
	scb    ports.SCBClient
}

func NewService(repo ports.RegionRepository, kolada ports.KoladaClient, scb ports.SCBClient) *Service {
	return &Service{repo: repo, kolada: kolada, scb: scb}
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

func (s *Service) GetMunicipalityKPIs(ctx context.Context, munCode string) ([]ports.KPIValue, error) {
	return s.kolada.FetchKPIs(ctx, munCode, defaultKPIs, defaultKPIYears)
}

func (s *Service) GetPopulationTrend(ctx context.Context, munCode string) ([]ports.PopulationEntry, error) {
	return s.scb.FetchPopulationTrend(ctx, munCode, defaultPopYears)
}
