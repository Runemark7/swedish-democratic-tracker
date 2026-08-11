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

	// FindInPeriod returns the one committee with that canonical code as the
	// record shows it in the period, or nil, nil when it decided nothing there.
	// Scoped to one code so the detail path does not re-derive the whole period.
	FindInPeriod(ctx context.Context, periodCode, code string) (*domain.Committee, error)

	// PeriodExists reports whether the mandate period is one we hold at all.
	// "No such period" and "this parliament decided nothing" are different
	// facts and must not share a response.
	PeriodExists(ctx context.Context, periodCode string) (bool, error)

	// DecidedBudgetYearExists reports whether the year has decided figures. A
	// year we hold nothing for must be refused, not rendered as 0 tkr per area.
	DecidedBudgetYearExists(ctx context.Context, year int) (bool, error)

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
