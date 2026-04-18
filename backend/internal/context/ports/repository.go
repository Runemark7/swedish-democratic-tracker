package ports

import (
	"context"

	"riksdagskollen/internal/context/domain"
)

// ContextRepository provides budget data scoped to a topic's UO codes.
type ContextRepository interface {
	// GetAreaTrend returns multi-year allocation trend for the given UO codes.
	GetAreaTrend(ctx context.Context, uoCodes []string) ([]domain.RelatedBudgetArea, error)
	// GetFundingContext returns year-over-year budget deltas for the given year.
	GetFundingContext(ctx context.Context, year int) (*domain.FundingContext, error)
	// GetLatestBudgetYear returns the most recent year with budget data.
	GetLatestBudgetYear(ctx context.Context) (int, error)
}
