package committees

import (
	"context"

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

// Get returns one committee with its utgiftsområden for the given budget year.
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
		areas, err := s.repo.AreasFor(ctx, code, budgetYear)
		if err != nil {
			return nil, err
		}
		all[i].ExpenditureAreas = areas
		return &all[i], nil
	}
	return nil, nil
}
