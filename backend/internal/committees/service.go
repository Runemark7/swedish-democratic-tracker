package committees

import (
	"context"
	"errors"

	"riksdagskollen/internal/committees/domain"
	"riksdagskollen/internal/committees/ports"
)

// Errors the caller must be able to distinguish. Each one names a request the
// record cannot answer, as opposed to an answer that happens to be empty:
// rendering "nothing here" for a question we never understood is how a site
// ends up stating something false with a straight face.
var (
	// ErrPeriodNotFound: the mandate period is not one we hold.
	ErrPeriodNotFound = errors.New("unknown mandate period")
	// ErrCommitteeNotFound: no committee with that code decided anything in the
	// period.
	ErrCommitteeNotFound = errors.New("committee not found in that period")
	// ErrBudgetYearNotDecided: the budget year has no decided figures, so any
	// amount we showed for it would be an invented zero.
	ErrBudgetYearNotDecided = errors.New("no decided budget figures for that year")
)

type Service struct {
	repo ports.Repository
}

func NewService(repo ports.Repository) *Service {
	return &Service{repo: repo}
}

// List returns the committees of a mandate period, alphabetically.
//
// An unknown period is ErrPeriodNotFound and never an empty list: 200 [] for
// ?period=2018-2022 is indistinguishable from a parliament that decided
// nothing, which is a claim about the record we are in no position to make.
func (s *Service) List(ctx context.Context, periodCode string) ([]domain.Committee, error) {
	if err := s.requirePeriod(ctx, periodCode); err != nil {
		return nil, err
	}
	return s.repo.ListForPeriod(ctx, periodCode)
}

// Get returns one committee with its utgiftsområden for the given budget
// year. Pass budgetYear 0 to resolve it to the newest decided budget year,
// rather than a hardcoded guess that would go stale the moment a newer year
// is seeded.
//
// Returns ErrCommitteeNotFound when the committee decided nothing in the
// period: a page that invented a committee out of a URL would claim something
// the record does not.
func (s *Service) Get(ctx context.Context, periodCode, code string, budgetYear int) (*domain.Committee, error) {
	if err := s.requirePeriod(ctx, periodCode); err != nil {
		return nil, err
	}

	year, err := s.resolveBudgetYear(ctx, budgetYear)
	if err != nil {
		return nil, err
	}

	c, err := s.repo.FindInPeriod(ctx, periodCode, code)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCommitteeNotFound
	}

	areas, err := s.repo.AreasFor(ctx, c.Code, year)
	if err != nil {
		return nil, err
	}
	c.ExpenditureAreas = areas
	return c, nil
}

func (s *Service) requirePeriod(ctx context.Context, periodCode string) error {
	ok, err := s.repo.PeriodExists(ctx, periodCode)
	if err != nil {
		return err
	}
	if !ok {
		return ErrPeriodNotFound
	}
	return nil
}

// resolveBudgetYear turns the requested year into one we actually hold decided
// figures for, or refuses. A year outside our allocations would otherwise be
// answered with every area at 0 tkr, which reads as a fact about the budget
// rather than as an absence of data.
func (s *Service) resolveBudgetYear(ctx context.Context, budgetYear int) (int, error) {
	if budgetYear != 0 {
		ok, err := s.repo.DecidedBudgetYearExists(ctx, budgetYear)
		if err != nil {
			return 0, err
		}
		if !ok {
			return 0, ErrBudgetYearNotDecided
		}
		return budgetYear, nil
	}

	year, err := s.repo.NewestDecidedBudgetYear(ctx)
	if err != nil {
		return 0, err
	}
	if year == 0 {
		// No decided budget year exists at all. This is a real state and must
		// not be papered over with a guessed year.
		return 0, errors.New("no decided budget year available")
	}
	return year, nil
}
