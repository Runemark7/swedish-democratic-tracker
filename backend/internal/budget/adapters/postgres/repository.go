package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/budget/domain"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListYears(ctx context.Context) ([]*domain.BudgetYear, error) {
	const q = `SELECT year, status, COALESCE(total_ksek, 0)
		FROM budget_years ORDER BY year DESC`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var years []*domain.BudgetYear
	for rows.Next() {
		var y domain.BudgetYear
		if err := rows.Scan(&y.Year, &y.Status, &y.TotalKsek); err != nil {
			return nil, err
		}
		years = append(years, &y)
	}
	return years, rows.Err()
}

func (r *Repository) GetYearDetail(ctx context.Context, year int) (*domain.BudgetYearDetail, error) {
	// Fetch budget year
	const yearQ = `SELECT year, status, COALESCE(total_ksek, 0)
		FROM budget_years WHERE year = $1 AND status = 'decided'`
	var detail domain.BudgetYearDetail
	err := r.db.QueryRow(ctx, yearQ, year).Scan(&detail.Year, &detail.Status, &detail.TotalKsek)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Fetch allocations
	const allocQ = `SELECT ea.code, ea.name, COALESCE(ea.description, ''), ea.sort_order,
		ba.amount_ksek
		FROM budget_allocations ba
		JOIN expenditure_areas ea ON ea.id = ba.expenditure_area_id
		JOIN budget_years by2 ON by2.id = ba.budget_year_id
		WHERE by2.year = $1 AND by2.status = 'decided' AND ba.source = 'government'
		ORDER BY ea.sort_order`
	rows, err := r.db.Query(ctx, allocQ, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var a domain.BudgetAllocation
		if err := rows.Scan(&a.Area.Code, &a.Area.Name, &a.Area.Description,
			&a.Area.SortOrder, &a.AmountKsek); err != nil {
			return nil, err
		}
		a.AmountFormatted = domain.FormatKsek(a.AmountKsek)
		detail.Allocations = append(detail.Allocations, a)
	}
	return &detail, rows.Err()
}

func (r *Repository) CompareYears(ctx context.Context, baseYear, compareYear int) (*domain.BudgetComparison, error) {
	const q = `SELECT
		ea.code, ea.name, COALESCE(ea.description, ''), ea.sort_order,
		COALESCE(base_alloc.amount_ksek, 0) AS base_amount,
		COALESCE(comp_alloc.amount_ksek, 0) AS comp_amount,
		COALESCE(comp_alloc.amount_ksek, 0) - COALESCE(base_alloc.amount_ksek, 0) AS delta_ksek,
		CASE WHEN COALESCE(base_alloc.amount_ksek, 0) = 0 THEN 0
		     ELSE ROUND((COALESCE(comp_alloc.amount_ksek, 0) - COALESCE(base_alloc.amount_ksek, 0))::numeric
		          / base_alloc.amount_ksek * 100, 1)
		END AS delta_pct
	FROM expenditure_areas ea
	LEFT JOIN budget_allocations base_alloc
		ON base_alloc.expenditure_area_id = ea.id
		AND base_alloc.budget_year_id = (SELECT id FROM budget_years WHERE year = $1 AND status = 'decided')
		AND base_alloc.source = 'government'
	LEFT JOIN budget_allocations comp_alloc
		ON comp_alloc.expenditure_area_id = ea.id
		AND comp_alloc.budget_year_id = (SELECT id FROM budget_years WHERE year = $2 AND status = 'decided')
		AND comp_alloc.source = 'government'
	ORDER BY ea.sort_order`

	rows, err := r.db.Query(ctx, q, baseYear, compareYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comp := &domain.BudgetComparison{
		BaseYear:    baseYear,
		CompareYear: compareYear,
	}

	for rows.Next() {
		var row domain.ComparisonRow
		if err := rows.Scan(&row.Area.Code, &row.Area.Name, &row.Area.Description,
			&row.Area.SortOrder, &row.BaseAmountKsek, &row.CompareAmountKsek,
			&row.DeltaKsek, &row.DeltaPct); err != nil {
			return nil, err
		}
		comp.BaseTotalKsek += row.BaseAmountKsek
		comp.CompareTotalKsek += row.CompareAmountKsek
		comp.Rows = append(comp.Rows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	comp.TotalDeltaKsek = comp.CompareTotalKsek - comp.BaseTotalKsek
	if comp.BaseTotalKsek > 0 {
		comp.TotalDeltaPct = float64(comp.TotalDeltaKsek) / float64(comp.BaseTotalKsek) * 100
	}

	// Return nil if no data found (both years missing)
	if len(comp.Rows) == 0 {
		return nil, nil
	}
	return comp, nil
}

func (r *Repository) ListAreas(ctx context.Context) ([]*domain.ExpenditureArea, error) {
	const q = `SELECT code, name, COALESCE(description, ''), sort_order
		FROM expenditure_areas ORDER BY sort_order`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []*domain.ExpenditureArea
	for rows.Next() {
		var a domain.ExpenditureArea
		if err := rows.Scan(&a.Code, &a.Name, &a.Description, &a.SortOrder); err != nil {
			return nil, err
		}
		areas = append(areas, &a)
	}
	return areas, rows.Err()
}

func (r *Repository) GetAreaTimeSeries(ctx context.Context, code string) (*domain.AreaTimeSeries, error) {
	// Fetch area info
	const areaQ = `SELECT code, name, COALESCE(description, ''), sort_order
		FROM expenditure_areas WHERE code = $1`
	var ts domain.AreaTimeSeries
	err := r.db.QueryRow(ctx, areaQ, code).Scan(
		&ts.Area.Code, &ts.Area.Name, &ts.Area.Description, &ts.Area.SortOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Fetch yearly amounts
	const seriesQ = `SELECT by2.year, ba.amount_ksek
		FROM budget_allocations ba
		JOIN budget_years by2 ON by2.id = ba.budget_year_id
		JOIN expenditure_areas ea ON ea.id = ba.expenditure_area_id
		WHERE ea.code = $1 AND ba.source = 'government' AND by2.status = 'decided'
		ORDER BY by2.year`
	rows, err := r.db.Query(ctx, seriesQ, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e domain.AreaTimeSeriesEntry
		if err := rows.Scan(&e.Year, &e.AmountKsek); err != nil {
			return nil, err
		}
		ts.Entries = append(ts.Entries, e)
	}
	return &ts, rows.Err()
}
