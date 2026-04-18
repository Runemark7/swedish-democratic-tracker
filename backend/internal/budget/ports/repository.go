package ports

import (
	"context"

	"riksdagskollen/internal/budget/domain"
)

type BudgetRepository interface {
	ListYears(ctx context.Context) ([]*domain.BudgetYear, error)
	GetYearDetail(ctx context.Context, year int) (*domain.BudgetYearDetail, error)
	CompareYears(ctx context.Context, baseYear, compareYear int) (*domain.BudgetComparison, error)
	ListAreas(ctx context.Context) ([]*domain.ExpenditureArea, error)
	GetAreaTimeSeries(ctx context.Context, code string) (*domain.AreaTimeSeries, error)
}
