package committees

import (
	"context"
	"errors"

	"riksdagskollen/internal/committees/domain"
	"riksdagskollen/internal/committees/ports"
)

type Service struct {
	repo ports.Repository
}

func NewService(repo ports.Repository) *Service {
	return &Service{repo: repo}
}

// List returns the committees of a mandate period, alphabetically.
func (s *Service) List(ctx context.Context, periodCode string) ([]domain.Committee, error) {
	return s.repo.ListForPeriod(ctx, periodCode)
}

// Get returns one committee with its utgiftsområden for the given budget
// year. Pass budgetYear 0 to resolve it to the newest decided budget year,
// rather than a hardcoded guess that would go stale the moment a newer year
// is seeded.
//
// Returns nil when the committee decided nothing in the period: a page that
// invented a committee out of a URL would claim something the record does not.
func (s *Service) Get(ctx context.Context, periodCode, code string, budgetYear int) (*domain.Committee, error) {
	all, err := s.repo.ListForPeriod(ctx, periodCode)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].Code != code {
			continue
		}
		year := budgetYear
		if year == 0 {
			year, err = s.repo.NewestDecidedBudgetYear(ctx)
			if err != nil {
				return nil, err
			}
			if year == 0 {
				// No decided budget year exists at all. This is a real state
				// and must not be papered over with a guessed year.
				return nil, errors.New("no decided budget year available")
			}
		}
		areas, err := s.repo.AreasFor(ctx, code, year)
		if err != nil {
			return nil, err
		}
		all[i].ExpenditureAreas = areas
		return &all[i], nil
	}
	return nil, nil
}
