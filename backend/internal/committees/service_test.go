package committees_test

import (
	"context"
	"errors"
	"testing"

	"riksdagskollen/internal/committees"
	"riksdagskollen/internal/committees/domain"
)

// stubRepo answers from fixed data so the service's refusals can be pinned
// without a database.
type stubRepo struct {
	periods      map[string]bool
	decidedYears map[int]bool
	newestYear   int
	inPeriod     map[string]int
	areas        []domain.ExpenditureArea
	listCalls    int
}

func (s *stubRepo) ListForPeriod(_ context.Context, _ string) ([]domain.Committee, error) {
	s.listCalls++
	out := []domain.Committee{}
	for code, n := range s.inPeriod {
		out = append(out, domain.Committee{Code: code, Name: code, Voteringar: n})
	}
	return out, nil
}

func (s *stubRepo) FindInPeriod(_ context.Context, _, code string) (*domain.Committee, error) {
	n, ok := s.inPeriod[code]
	if !ok {
		return nil, nil
	}
	return &domain.Committee{Code: code, Name: code, Voteringar: n}, nil
}

func (s *stubRepo) PeriodExists(_ context.Context, code string) (bool, error) {
	return s.periods[code], nil
}

func (s *stubRepo) DecidedBudgetYearExists(_ context.Context, year int) (bool, error) {
	return s.decidedYears[year], nil
}

func (s *stubRepo) AreasFor(_ context.Context, _ string, _ int) ([]domain.ExpenditureArea, error) {
	return s.areas, nil
}

func (s *stubRepo) NewestDecidedBudgetYear(_ context.Context) (int, error) {
	return s.newestYear, nil
}

func newStub() *stubRepo {
	return &stubRepo{
		periods:      map[string]bool{"2022-2026": true},
		decidedYears: map[int]bool{2026: true},
		newestYear:   2026,
		inPeriod:     map[string]int{"FiU": 104},
		areas:        []domain.ExpenditureArea{{Code: "UO2", Name: "Samhällsekonomi", AmountKsek: 1}},
	}
}

// An unknown period must be refused, not answered with an empty list: 200 []
// asserts that the parliament decided nothing, which is a claim about the
// record and not a report of a bad request.
func TestList_UnknownPeriodIsRefused(t *testing.T) {
	svc := committees.NewService(newStub())

	if _, err := svc.List(context.Background(), "2018-2022"); !errors.Is(err, committees.ErrPeriodNotFound) {
		t.Errorf("List(2018-2022) err = %v, want ErrPeriodNotFound", err)
	}
	if _, err := svc.List(context.Background(), "2022-2026"); err != nil {
		t.Errorf("List(2022-2026) err = %v, want nil", err)
	}
}

func TestGet_UnknownPeriodIsRefused(t *testing.T) {
	svc := committees.NewService(newStub())

	if _, err := svc.Get(context.Background(), "2018-2022", "FiU", 0); !errors.Is(err, committees.ErrPeriodNotFound) {
		t.Errorf("Get err = %v, want ErrPeriodNotFound", err)
	}
}

// A year we hold no decided figures for must be refused rather than rendered
// as every area at 0 tkr, which reads as a budget fact instead of an absence.
func TestGet_YearWithoutDecidedFiguresIsRefused(t *testing.T) {
	svc := committees.NewService(newStub())

	if _, err := svc.Get(context.Background(), "2022-2026", "FiU", 1999); !errors.Is(err, committees.ErrBudgetYearNotDecided) {
		t.Errorf("Get(year=1999) err = %v, want ErrBudgetYearNotDecided", err)
	}
}

func TestGet_UnspecifiedYearResolvesToNewestDecided(t *testing.T) {
	svc := committees.NewService(newStub())

	c, err := svc.Get(context.Background(), "2022-2026", "FiU", 0)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(c.ExpenditureAreas) != 1 {
		t.Errorf("areas = %d, want 1", len(c.ExpenditureAreas))
	}
}

func TestGet_UnknownCommitteeIsNotFound(t *testing.T) {
	svc := committees.NewService(newStub())

	if _, err := svc.Get(context.Background(), "2022-2026", "ZZU", 0); !errors.Is(err, committees.ErrCommitteeNotFound) {
		t.Errorf("Get(ZZU) err = %v, want ErrCommitteeNotFound", err)
	}
}

// The detail path must read one committee, not re-derive the whole period.
func TestGet_DoesNotListThePeriod(t *testing.T) {
	repo := newStub()
	svc := committees.NewService(repo)

	if _, err := svc.Get(context.Background(), "2022-2026", "FiU", 2026); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if repo.listCalls != 0 {
		t.Errorf("ListForPeriod called %d times on the detail path, want 0", repo.listCalls)
	}
}
