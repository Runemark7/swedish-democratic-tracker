package ports

import (
	"context"

	"riksdagskollen/internal/committees/domain"
)

// Repository reads committees as the record shows them.
type Repository interface {
	// ListForPeriod returns every committee that decided at least one votering
	// in the mandate period, alphabetically by code. Derived from the record:
	// a committee that existed but decided nothing does not appear, and one
	// abolished since does.
	ListForPeriod(ctx context.Context, periodCode string) ([]domain.Committee, error)

	// AreasFor returns the utgiftsområden a committee bereder, with the amount
	// allocated in the given budget year. Never a share of the total.
	AreasFor(ctx context.Context, utskottCode string, budgetYear int) ([]domain.ExpenditureArea, error)

	// NewestDecidedBudgetYear returns the most recent budget year with decided
	// figures, used when a caller does not name a year. Selecting on
	// status = 'decided' matches how the budget feature picks a year.
	// Returns 0, nil when no budget year has been decided yet — a real state,
	// not an error.
	NewestDecidedBudgetYear(ctx context.Context) (int, error)
}
