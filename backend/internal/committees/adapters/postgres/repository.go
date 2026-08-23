package postgres

import (
	"context"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/committees/domain"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ListForPeriod derives the committees from the beteckningar in the record.
//
// The prefix rule lives in domain.CommitteeCode and nowhere else, so the
// ALL-CAPS vintage in 2002-2006 collapses onto the same committee rather than
// appearing as a second one.
func (r *Repository) ListForPeriod(ctx context.Context, periodCode string) ([]domain.Committee, error) {
	counts, err := r.countVoteringar(ctx, periodCode, "")
	if err != nil {
		return nil, err
	}

	out := make([]domain.Committee, 0, len(counts))
	for code, n := range counts {
		out = append(out, committee(code, n))
	}
	// Alphabetical by code. Any other order — by volume, by "importance" — is a
	// ranking, and ranking is an editorial act.
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out, nil
}

// FindInPeriod returns one committee as the record shows it in the period, or
// nil when it decided nothing there.
//
// Scoped to a single code so the detail page does not re-derive every
// committee in the period just to read one of them.
func (r *Repository) FindInPeriod(ctx context.Context, periodCode, code string) (*domain.Committee, error) {
	if code == "" {
		return nil, nil
	}
	counts, err := r.countVoteringar(ctx, periodCode, code)
	if err != nil {
		return nil, err
	}
	n, ok := counts[code]
	if !ok {
		return nil, nil
	}
	c := committee(code, n)
	return &c, nil
}

// PeriodExists reports whether the mandate period is one the record knows.
//
// Needed because "no committees" and "no such period" are different facts: a
// typo'd or pre-record period must not be answered with an empty list, which
// reads as "this parliament decided nothing".
func (r *Repository) PeriodExists(ctx context.Context, periodCode string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM mandate_periods WHERE code = $1)`, periodCode,
	).Scan(&exists)
	return exists, err
}

// DecidedBudgetYearExists reports whether the year has decided figures.
//
// A year we hold no allocations for would otherwise render every area at 0 tkr
// — "FiU bereder UO2: 0 tkr" — which is a false statement, not a missing one.
func (r *Repository) DecidedBudgetYearExists(ctx context.Context, year int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM budget_years WHERE year = $1 AND status = 'decided')`, year,
	).Scan(&exists)
	return exists, err
}

// countVoteringar folds the record's distinct voteringar into a count per
// canonical committee code.
//
// Counted per votering_id — the record's own identity for a votering. The
// obvious alternative, (beteckning, förslagspunkt), is wrong twice over:
// betänkande numbering restarts each riksmöte, so 2022/23 FiU1 and 2023/24 FiU1
// collapse into one, and it also discards the 18 förslagspunkter in 2022-2026
// on which Riksdagen genuinely held two voteringar (e.g. 2023/24 NU1 punkt 2).
// Both are facts of the record, and both were being deleted from the total.
//
// codeFilter, when non-empty, narrows the scan to one committee. It is only a
// narrowing predicate: the authoritative code is always derived in Go by
// domain.CommitteeCode, so a beteckning that merely starts with the same
// letters can never be counted under the wrong committee.
func (r *Repository) countVoteringar(ctx context.Context, periodCode, codeFilter string) (map[string]int, error) {
	// One row per votering already, so the DISTINCT this needed against the old
	// table — where each votering appeared ~350 times — is gone, and the scan
	// is over 15 835 rows rather than 5.5 million.
	query := `
		SELECT v.beteckning, v.votering_id
		FROM voteringar v
		JOIN mandate_periods p ON v.session = ANY(p.riksmoten)
		WHERE p.code = $1
		  AND v.beteckning <> ''`
	args := []any{periodCode}
	if codeFilter != "" {
		// starts_with, not LIKE: the code comes from a URL path and must not be
		// able to smuggle in wildcards.
		query += `
		  AND starts_with(upper(v.beteckning), upper($2))`
		args = append(args, codeFilter)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	merged := map[string]int{}
	for rows.Next() {
		var beteckning, voteringID string
		if err := rows.Scan(&beteckning, &voteringID); err != nil {
			return nil, err
		}
		code := domain.CommitteeCode(beteckning)
		if code == "" {
			continue
		}
		// One row is one distinct votering, so this is an increment and never a
		// sum of pre-aggregated counts.
		merged[code]++
	}
	return merged, rows.Err()
}

func committee(code string, voteringar int) domain.Committee {
	return domain.Committee{
		Code:             code,
		Name:             domain.DisplayName(code),
		Voteringar:       voteringar,
		ExpenditureAreas: []domain.ExpenditureArea{},
	}
}

// AreasFor returns the utgiftsområden a committee bereder with their amounts.
//
// The allocation in force is resolved per utgiftsområde: the newest row for
// that UO whose in_force_from has passed. A global max(in_force_from) would be
// correct only if every future revision re-inserted all 27 rows, and
// docs/data-sources/utskott-utgiftsomrade.md prescribes the opposite — add one
// row for the UO that moved and leave the rest untouched. On that documented
// path a global max would return only the single new row, so every other
// committee page would show zero utgiftsområden with no error anywhere.
func (r *Repository) AreasFor(ctx context.Context, utskottCode string, budgetYear int) ([]domain.ExpenditureArea, error) {
	rows, err := r.db.Query(ctx, `
		WITH in_force AS (
			SELECT DISTINCT ON (uo_code) uo_code, utskott_code
			FROM utskott_utgiftsomrade
			WHERE in_force_from <= CURRENT_DATE
			ORDER BY uo_code, in_force_from DESC
		)
		SELECT ea.code, ea.name, COALESCE(ba.amount_ksek, 0)
		FROM in_force f
		JOIN expenditure_areas ea ON ea.code = f.uo_code
		-- budget_years permits more than one row per year (decided vs. proposed,
		-- UNIQUE (year, status)). The rest of the codebase always pins to the
		-- decided one (see internal/budget/adapters/postgres/repository.go);
		-- omitting the filter here would join both rows for a dual-status year
		-- and silently double every area's amount.
		LEFT JOIN budget_years by2 ON by2.year = $2 AND by2.status = 'decided'
		LEFT JOIN budget_allocations ba
		       ON ba.expenditure_area_id = ea.id
		      AND ba.budget_year_id = by2.id
		      AND ba.source = 'government'
		      AND ba.party IS NULL
		WHERE f.utskott_code = $1
		-- sort_order, as every other query over this table does (see
		-- internal/budget/adapters/postgres/repository.go). Casting the digits
		-- out of ea.code gives the same order today but *errors* on a code that
		-- holds no digits, turning a display detail into a failed request.
		ORDER BY ea.sort_order`, utskottCode, budgetYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.ExpenditureArea{}
	for rows.Next() {
		var a domain.ExpenditureArea
		if err := rows.Scan(&a.Code, &a.Name, &a.AmountKsek); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// NewestDecidedBudgetYear returns the most recent budget year with decided
// figures, matching how the budget feature pins to status = 'decided' (see
// GetYearDetail in internal/budget/adapters/postgres/repository.go).
//
// max() over an empty set is a single row holding NULL, not zero rows, so
// COALESCE folds "no decided year yet" to 0 rather than requiring a
// pgx.ErrNoRows branch. 0 is a real state the caller must check for, not an
// error: the query itself did not fail.
func (r *Repository) NewestDecidedBudgetYear(ctx context.Context) (int, error) {
	var year int
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(max(year), 0) FROM budget_years WHERE status = 'decided'`,
	).Scan(&year)
	if err != nil {
		return 0, err
	}
	return year, nil
}
