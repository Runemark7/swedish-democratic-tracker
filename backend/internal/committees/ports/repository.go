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
}
