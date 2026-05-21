package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/regions/domain"
	"riksdagskollen/internal/regions/ports"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListRegions(ctx context.Context) ([]*domain.Region, error) {
	rows, err := r.db.Query(ctx, `
		SELECT code, name, capital, population, governing_parties, election_year, total_mandates
		FROM regions
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []*domain.Region
	for rows.Next() {
		reg := &domain.Region{}
		if err := rows.Scan(
			&reg.Code, &reg.Name, &reg.Capital, &reg.Population,
			&reg.GoverningParties, &reg.ElectionYear, &reg.TotalMandates,
		); err != nil {
			return nil, err
		}
		regions = append(regions, reg)
	}
	return regions, rows.Err()
}

func (r *Repository) GetRegion(ctx context.Context, code string) (*domain.RegionDetail, error) {
	detail := &domain.RegionDetail{}
	detail.Region.GoverningParties = []string{}
	detail.ElectionResults = []domain.ElectionResult{}
	err := r.db.QueryRow(ctx, `
		SELECT code, name, capital, population, governing_parties, election_year, total_mandates
		FROM regions WHERE code = $1
	`, code).Scan(
		&detail.Code, &detail.Name, &detail.Capital, &detail.Population,
		&detail.GoverningParties, &detail.ElectionYear, &detail.TotalMandates,
	)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT party, mandates, vote_pct, total_mandates
		FROM regional_election_results
		WHERE region_code = $1
		ORDER BY mandates DESC
	`, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		er := domain.ElectionResult{}
		if err := rows.Scan(&er.Party, &er.Mandates, &er.VotePct, &er.TotalMandates); err != nil {
			return nil, err
		}
		detail.ElectionResults = append(detail.ElectionResults, er)
	}
	return detail, rows.Err()
}

func (r *Repository) ListMunicipalities(ctx context.Context, regionCode string) ([]*domain.Municipality, error) {
	if regionCode != "" {
		return r.listMunicipalitiesByRegion(ctx, regionCode)
	}
	return r.listAllMunicipalities(ctx)
}

func (r *Repository) listMunicipalitiesByRegion(ctx context.Context, regionCode string) ([]*domain.Municipality, error) {
	rows, err := r.db.Query(ctx, `
		SELECT m.code, m.name, m.region_code, reg.name, m.population,
		       m.governing_parties, m.election_year, m.total_mandates
		FROM municipalities m
		JOIN regions reg ON reg.code = m.region_code
		WHERE m.region_code = $1
		ORDER BY m.name
	`, regionCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMunicipalities(rows)
}

func (r *Repository) listAllMunicipalities(ctx context.Context) ([]*domain.Municipality, error) {
	rows, err := r.db.Query(ctx, `
		SELECT m.code, m.name, m.region_code, reg.name, m.population,
		       m.governing_parties, m.election_year, m.total_mandates
		FROM municipalities m
		JOIN regions reg ON reg.code = m.region_code
		ORDER BY m.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMunicipalities(rows)
}

func scanMunicipalities(rows pgx.Rows) ([]*domain.Municipality, error) {
	var municipalities []*domain.Municipality
	for rows.Next() {
		mun := &domain.Municipality{}
		if err := rows.Scan(
			&mun.Code, &mun.Name, &mun.RegionCode, &mun.RegionName,
			&mun.Population, &mun.GoverningParties, &mun.ElectionYear, &mun.TotalMandates,
		); err != nil {
			return nil, err
		}
		municipalities = append(municipalities, mun)
	}
	return municipalities, rows.Err()
}

func (r *Repository) UpsertRegionBudgetSnapshots(ctx context.Context, snapshots []ports.RegionBudgetSnapshot) (int, error) {
	if len(snapshots) == 0 {
		return 0, nil
	}
	const q = `
		INSERT INTO region_budget_snapshots (region_code, area_name, year, value_mnkr, total_mnkr, pct, fetched_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (region_code, area_name, year) DO UPDATE SET
			value_mnkr = EXCLUDED.value_mnkr,
			total_mnkr = EXCLUDED.total_mnkr,
			pct        = EXCLUDED.pct,
			fetched_at = NOW()
	`
	var count int
	for _, s := range snapshots {
		tag, err := r.db.Exec(ctx, q, s.RegionCode, s.AreaName, s.Year, s.ValueMnkr, s.TotalMnkr, s.Pct)
		if err != nil {
			return count, err
		}
		count += int(tag.RowsAffected())
	}
	return count, nil
}

func (r *Repository) GetRegionBudgetHistory(ctx context.Context, regionCode string, years []int) ([]ports.RegionBudgetSnapshot, error) {
	var rows pgx.Rows
	var err error
	if len(years) == 0 {
		rows, err = r.db.Query(ctx, `
			SELECT region_code, area_name, year, value_mnkr, total_mnkr, pct
			FROM region_budget_snapshots
			WHERE region_code = $1
			ORDER BY year, area_name
		`, regionCode)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT region_code, area_name, year, value_mnkr, total_mnkr, pct
			FROM region_budget_snapshots
			WHERE region_code = $1 AND year = ANY($2)
			ORDER BY year, area_name
		`, regionCode, years)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ports.RegionBudgetSnapshot
	for rows.Next() {
		var s ports.RegionBudgetSnapshot
		if err := rows.Scan(&s.RegionCode, &s.AreaName, &s.Year, &s.ValueMnkr, &s.TotalMnkr, &s.Pct); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (r *Repository) GetAreaAcrossRegions(ctx context.Context, areaName string, year int) ([]ports.RegionAreaDataPoint, error) {
	if year == 0 {
		// Pick the year with the most regions for this area (most complete snapshot).
		if err := r.db.QueryRow(ctx, `
			SELECT year FROM region_budget_snapshots
			WHERE area_name = $1
			GROUP BY year ORDER BY COUNT(*) DESC, year DESC LIMIT 1
		`, areaName).Scan(&year); err != nil {
			return nil, err
		}
	}
	rows, err := r.db.Query(ctx, `
		SELECT region_code, value_mnkr, total_mnkr, pct
		FROM region_budget_snapshots
		WHERE area_name = $1 AND year = $2
		ORDER BY region_code
	`, areaName, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ports.RegionAreaDataPoint
	for rows.Next() {
		var d ports.RegionAreaDataPoint
		if err := rows.Scan(&d.RegionCode, &d.ValueMnkr, &d.TotalMnkr, &d.Pct); err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

func (r *Repository) UpsertMunicipalityBudgetSnapshots(ctx context.Context, snapshots []ports.MunicipalityBudgetSnapshot) (int, error) {
	if len(snapshots) == 0 {
		return 0, nil
	}
	const q = `
		INSERT INTO municipality_budget_snapshots (mun_code, area_name, year, value_mnkr, total_mnkr, pct, fetched_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (mun_code, area_name, year) DO UPDATE SET
			value_mnkr = EXCLUDED.value_mnkr,
			total_mnkr = EXCLUDED.total_mnkr,
			pct        = EXCLUDED.pct,
			fetched_at = NOW()
	`
	var count int
	for _, s := range snapshots {
		tag, err := r.db.Exec(ctx, q, s.MunCode, s.AreaName, s.Year, s.ValueMnkr, s.TotalMnkr, s.Pct)
		if err != nil {
			return count, err
		}
		count += int(tag.RowsAffected())
	}
	return count, nil
}

func (r *Repository) GetMunicipalityBudgetHistory(ctx context.Context, munCode string, years []int) ([]ports.MunicipalityBudgetSnapshot, error) {
	var rows pgx.Rows
	var err error
	if len(years) == 0 {
		rows, err = r.db.Query(ctx, `
			SELECT mun_code, area_name, year, value_mnkr, total_mnkr, pct
			FROM municipality_budget_snapshots
			WHERE mun_code = $1
			ORDER BY year, area_name
		`, munCode)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT mun_code, area_name, year, value_mnkr, total_mnkr, pct
			FROM municipality_budget_snapshots
			WHERE mun_code = $1 AND year = ANY($2)
			ORDER BY year, area_name
		`, munCode, years)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ports.MunicipalityBudgetSnapshot
	for rows.Next() {
		var s ports.MunicipalityBudgetSnapshot
		if err := rows.Scan(&s.MunCode, &s.AreaName, &s.Year, &s.ValueMnkr, &s.TotalMnkr, &s.Pct); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (r *Repository) GetMunicipality(ctx context.Context, code string) (*domain.MunicipalityDetail, error) {
	detail := &domain.MunicipalityDetail{}
	detail.Municipality.GoverningParties = []string{}
	detail.ElectionResults = []domain.ElectionResult{}
	err := r.db.QueryRow(ctx, `
		SELECT m.code, m.name, m.region_code, reg.name, m.population,
		       m.governing_parties, m.election_year, m.total_mandates
		FROM municipalities m
		JOIN regions reg ON reg.code = m.region_code
		WHERE m.code = $1
	`, code).Scan(
		&detail.Code, &detail.Name, &detail.RegionCode, &detail.RegionName,
		&detail.Population, &detail.GoverningParties, &detail.ElectionYear, &detail.TotalMandates,
	)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT party, mandates, vote_pct, total_mandates
		FROM municipal_election_results
		WHERE municipality_code = $1
		ORDER BY mandates DESC
	`, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		er := domain.ElectionResult{}
		if err := rows.Scan(&er.Party, &er.Mandates, &er.VotePct, &er.TotalMandates); err != nil {
			return nil, err
		}
		detail.ElectionResults = append(detail.ElectionResults, er)
	}
	return detail, rows.Err()
}
