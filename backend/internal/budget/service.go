package budget

import (
	"context"

	"riksdagskollen/internal/budget/domain"
	"riksdagskollen/internal/budget/ports"
)

type Service struct {
	repo ports.BudgetRepository
}

func NewService(repo ports.BudgetRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListYears(ctx context.Context) ([]*domain.BudgetYear, error) {
	return s.repo.ListYears(ctx)
}

func (s *Service) GetYearDetail(ctx context.Context, year int) (*domain.BudgetYearDetail, error) {
	detail, err := s.repo.GetYearDetail(ctx, year)
	if err != nil || detail == nil {
		return detail, err
	}
	detail.Documents = domain.DocumentRefsForYear(year)
	return detail, nil
}

func (s *Service) CompareYears(ctx context.Context, baseYear, compareYear int) (*domain.BudgetComparison, error) {
	return s.repo.CompareYears(ctx, baseYear, compareYear)
}

func (s *Service) ListAreas(ctx context.Context) ([]*domain.ExpenditureArea, error) {
	return s.repo.ListAreas(ctx)
}

func (s *Service) GetAreaTimeSeries(ctx context.Context, code string) (*domain.AreaTimeSeries, error) {
	ts, err := s.repo.GetAreaTimeSeries(ctx, code)
	if err != nil || ts == nil {
		return ts, err
	}
	// Populate document refs for each year entry
	for i := range ts.Entries {
		ts.Entries[i].Documents = domain.DocumentRefsForArea(ts.Entries[i].Year, ts.Area.SortOrder)
	}
	return ts, nil
}
