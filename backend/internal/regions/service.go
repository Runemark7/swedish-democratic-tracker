package regions

import (
	"context"

	"riksdagskollen/internal/regions/domain"
	"riksdagskollen/internal/regions/ports"
)

var defaultKPIs = []string{
	// Verksamhet
	"N00900", "N11037", "N15027", "N20043", "N03010",
	// Budget & ekonomi
	"N03007", "N03102", "N03106", "N03040", "N03132",
}
var defaultKPIYears = []int{2019, 2020, 2021, 2022, 2023}
var defaultPopYears = []int{2019, 2020, 2021, 2022, 2023}

var spendingKPIs  = []string{"N11004", "N15028", "N17014", "N20014", "N30005", "N07037", "N09022", "N05011"}
var spendingYears = []int{2019, 2020, 2021, 2022, 2023}

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

func (s *Service) GetMunicipalitySpending(ctx context.Context, munCode string) ([]ports.KPIValue, error) {
	return s.kolada.FetchKPIs(ctx, munCode, spendingKPIs, spendingYears)
}
