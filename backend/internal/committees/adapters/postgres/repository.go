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
// The prefix is taken in SQL and then canonicalised in Go, so the ALL-CAPS
// vintage in 2002-2006 collapses onto the same committee rather than appearing
// as a second one.
func (r *Repository) ListForPeriod(ctx context.Context, periodCode string) ([]domain.Committee, error) {
	rows, err := r.db.Query(ctx, `
		SELECT substring(v.beteckning from '^[A-Za-zÅÄÖåäö]+') AS code,
		       count(DISTINCT v.beteckning || ':' || v.forslagspunkt)  AS voteringar
		FROM votes v
		JOIN mandate_periods p ON v.session = ANY(p.riksmoten)
		WHERE p.code = $1
		  AND substring(v.beteckning from '^[A-Za-zÅÄÖåäö]+') <> ''
		GROUP BY 1`, periodCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Canonicalise in Go, then merge: two casings are one committee.
	merged := map[string]int{}
	for rows.Next() {
		var raw string
		var n int
		if err := rows.Scan(&raw, &n); err != nil {
			return nil, err
		}
		code := domain.Canonical(raw)
		if code == "" {
			continue
		}
		merged[code] += n
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]domain.Committee, 0, len(merged))
	for code, n := range merged {
		out = append(out, domain.Committee{
			Code:             code,
			Name:             domain.DisplayName(code),
			Voteringar:       n,
			ExpenditureAreas: []domain.ExpenditureArea{},
		})
	}
	// Alphabetical by code. Any other order — by volume, by "importance" — is a
	// ranking, and ranking is an editorial act.
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out, nil
}

// AreasFor returns the utgiftsområden a committee bereder with their amounts.
//
// The allocation in force is the newest one whose in_force_from has passed.
func (r *Repository) AreasFor(ctx context.Context, utskottCode string, budgetYear int) ([]domain.ExpenditureArea, error) {
	rows, err := r.db.Query(ctx, `
		WITH in_force AS (
			SELECT uo_code, utskott_code
			FROM utskott_utgiftsomrade m
			WHERE in_force_from = (
				SELECT max(in_force_from) FROM utskott_utgiftsomrade
				WHERE in_force_from <= CURRENT_DATE
			)
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
		ORDER BY (regexp_replace(ea.code, '\D', '', 'g'))::int`, utskottCode, budgetYear)
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
