package postgres

import (
	"context"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/context/domain"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetLatestBudgetYear returns the most recent year that has decided budget data.
func (r *Repository) GetLatestBudgetYear(ctx context.Context) (int, error) {
	const q = `SELECT COALESCE(MAX(year), 0) FROM budget_years WHERE status = 'decided'`
	var year int
	if err := r.db.QueryRow(ctx, q).Scan(&year); err != nil {
		return 0, err
	}
	return year, nil
}

// GetAreaTrend fetches multi-year allocation data for the given UO codes,
// grouped into RelatedBudgetArea values with computed trend metrics.
func (r *Repository) GetAreaTrend(ctx context.Context, uoCodes []string) ([]domain.RelatedBudgetArea, error) {
	if len(uoCodes) == 0 {
		return nil, nil
	}

	const q = `
		SELECT ea.code, ea.name, byr.year, ba.amount_ksek, COALESCE(byr.total_ksek, 0)
		FROM budget_allocations ba
		JOIN expenditure_areas ea ON ea.id = ba.expenditure_area_id
		JOIN budget_years byr ON byr.id = ba.budget_year_id
		WHERE ea.code = ANY($1::text[])
		  AND byr.status = 'decided'
		  AND ba.source = 'government'
		ORDER BY ea.code, byr.year`

	rows, err := r.db.Query(ctx, q, uoCodes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Collect rows grouped by code
	type rawRow struct {
		code       string
		name       string
		year       int
		amountKsek int64
		totalKsek  int64
	}
	byCode := map[string][]rawRow{}
	codeOrder := []string{}

	for rows.Next() {
		var rw rawRow
		if err := rows.Scan(&rw.code, &rw.name, &rw.year, &rw.amountKsek, &rw.totalKsek); err != nil {
			return nil, err
		}
		if _, seen := byCode[rw.code]; !seen {
			codeOrder = append(codeOrder, rw.code)
		}
		byCode[rw.code] = append(byCode[rw.code], rw)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	areas := make([]domain.RelatedBudgetArea, 0, len(codeOrder))
	for _, code := range codeOrder {
		rawRows := byCode[code]
		if len(rawRows) == 0 {
			continue
		}

		// Build trend entries
		trend := make([]domain.BudgetTrendEntry, 0, len(rawRows))
		for _, rw := range rawRows {
			trend = append(trend, domain.BudgetTrendEntry{
				Year:            rw.year,
				AmountKsek:      rw.amountKsek,
				AmountFormatted: domain.FormatKsek(rw.amountKsek),
			})
		}

		// Latest delta %
		var latestDeltaPct float64
		if len(rawRows) >= 2 {
			last := rawRows[len(rawRows)-1]
			prev := rawRows[len(rawRows)-2]
			if prev.amountKsek > 0 {
				latestDeltaPct = float64(last.amountKsek-prev.amountKsek) / float64(prev.amountKsek) * 100
			}
		}

		// Share of total budget (latest year)
		var shareOfBudgetPct float64
		latest := rawRows[len(rawRows)-1]
		if latest.totalKsek > 0 {
			shareOfBudgetPct = float64(latest.amountKsek) / float64(latest.totalKsek) * 100
		}

		areas = append(areas, domain.RelatedBudgetArea{
			Code:             code,
			Name:             rawRows[0].name,
			Trend:            trend,
			LatestDeltaPct:   round2(latestDeltaPct),
			ShareOfBudgetPct: round2(shareOfBudgetPct),
		})
	}
	return areas, nil
}

// GetFundingContext returns year-over-year budget deltas, showing which
// areas grew and which shrank in the given budget year.
func (r *Repository) GetFundingContext(ctx context.Context, year int) (*domain.FundingContext, error) {
	// Query all areas with their delta vs prior year
	const q = `
		SELECT ea.code, ea.name,
		       curr.amount_ksek - prev.amount_ksek AS delta_ksek,
		       prev.amount_ksek
		FROM budget_allocations curr
		JOIN budget_allocations prev ON prev.expenditure_area_id = curr.expenditure_area_id
		JOIN budget_years cy  ON cy.id  = curr.budget_year_id AND cy.year = $1   AND cy.status = 'decided'
		JOIN budget_years py  ON py.id  = prev.budget_year_id AND py.year = $1-1 AND py.status = 'decided'
		JOIN expenditure_areas ea ON ea.id = curr.expenditure_area_id
		WHERE curr.source = 'government' AND prev.source = 'government'
		ORDER BY delta_ksek DESC`

	rows, err := r.db.Query(ctx, q, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allRows []domain.FundingRow
	for rows.Next() {
		var fr domain.FundingRow
		var prevAmount int64
		if err := rows.Scan(&fr.Code, &fr.Name, &fr.DeltaKsek, &prevAmount); err != nil {
			return nil, err
		}
		if prevAmount > 0 {
			fr.DeltaPct = round2(float64(fr.DeltaKsek) / float64(prevAmount) * 100)
		}
		allRows = append(allRows, fr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(allRows) == 0 {
		return nil, nil
	}

	// allRows is already sorted DESC by delta_ksek — take top 3 each side
	const top = 3
	var increases, decreases []domain.FundingRow
	for _, fr := range allRows {
		if fr.DeltaKsek > 0 && len(increases) < top {
			increases = append(increases, fr)
		}
	}
	// Sort ascending to get largest decreases first
	sorted := make([]domain.FundingRow, len(allRows))
	copy(sorted, allRows)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].DeltaKsek < sorted[j].DeltaKsek })
	for _, fr := range sorted {
		if fr.DeltaKsek < 0 && len(decreases) < top {
			decreases = append(decreases, fr)
		}
	}

	// Total budget delta
	const totalQ = `
		SELECT COALESCE(cy.total_ksek, 0) - COALESCE(py.total_ksek, 0),
		       COALESCE(py.total_ksek, 0)
		FROM budget_years cy
		JOIN budget_years py ON py.year = $1-1 AND py.status = 'decided'
		WHERE cy.year = $1 AND cy.status = 'decided'`
	var totalDelta, prevTotal int64
	if err := r.db.QueryRow(ctx, totalQ, year).Scan(&totalDelta, &prevTotal); err != nil {
		// Non-fatal: return what we have without total
		return &domain.FundingContext{
			Year:         year,
			TopIncreases: increases,
			TopDecreases: decreases,
		}, nil
	}

	var totalDeltaPct float64
	if prevTotal > 0 {
		totalDeltaPct = round2(float64(totalDelta) / float64(prevTotal) * 100)
	}

	return &domain.FundingContext{
		Year:                 year,
		TopIncreases:         increases,
		TopDecreases:         decreases,
		TotalBudgetDeltaKsek: totalDelta,
		TotalBudgetDeltaPct:  totalDeltaPct,
	}, nil
}

func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
