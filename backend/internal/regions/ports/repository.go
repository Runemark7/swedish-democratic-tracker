package ports

import (
	"context"

	"riksdagskollen/internal/regions/domain"
)

type RegionRepository interface {
	ListRegions(ctx context.Context) ([]*domain.Region, error)
	GetRegion(ctx context.Context, code string) (*domain.RegionDetail, error)
	ListMunicipalities(ctx context.Context, regionCode string) ([]*domain.Municipality, error)
	GetMunicipality(ctx context.Context, code string) (*domain.MunicipalityDetail, error)
	UpsertRegionBudgetSnapshots(ctx context.Context, snapshots []RegionBudgetSnapshot) (int, error)
	GetRegionBudgetHistory(ctx context.Context, regionCode string, years []int) ([]RegionBudgetSnapshot, error)
	GetAreaAcrossRegions(ctx context.Context, areaName string, year int) ([]RegionAreaDataPoint, error)
}
