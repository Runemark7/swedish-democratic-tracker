package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/regions/domain"
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

func (r *Repository) GetMunicipality(ctx context.Context, code string) (*domain.MunicipalityDetail, error) {
	detail := &domain.MunicipalityDetail{}
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
