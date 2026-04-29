package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
)

type AgencyIntelRepository struct {
	pool *pgxpool.Pool
}

func NewAgencyIntelRepository(pool *pgxpool.Pool) *AgencyIntelRepository {
	return &AgencyIntelRepository{pool: pool}
}

func (r *AgencyIntelRepository) UpsertRegleringsbrev(ctx context.Context, slug string, docs []ports.RiksdagenDoc) error {
	for _, d := range docs {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO agency_regleringsbrev (agency_slug, year, dok_id, date, title, summary, url)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (agency_slug, year) DO UPDATE SET
				dok_id  = EXCLUDED.dok_id,
				date    = EXCLUDED.date,
				title   = EXCLUDED.title,
				summary = EXCLUDED.summary,
				url     = EXCLUDED.url
		`, slug, d.Year, d.DokID, d.Date, d.Title, d.Summary, d.URL)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *AgencyIntelRepository) UpsertDecisions(ctx context.Context, slug string, docs []ports.RiksdagenDoc) error {
	for _, d := range docs {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO agency_decisions (agency_slug, dok_id, date, title, doc_type, summary, url)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (agency_slug, dok_id) DO UPDATE SET
				date     = EXCLUDED.date,
				title    = EXCLUDED.title,
				doc_type = EXCLUDED.doc_type,
				summary  = EXCLUDED.summary,
				url      = EXCLUDED.url
		`, slug, d.DokID, d.Date, d.Title, d.DocType, d.Summary, d.URL)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *AgencyIntelRepository) GetRegleringsbrev(ctx context.Context, slug string) ([]domain.Regleringsbrev, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT year, date, title, summary, url
		FROM agency_regleringsbrev
		WHERE agency_slug = $1
		ORDER BY year DESC
	`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Regleringsbrev
	for rows.Next() {
		var rb domain.Regleringsbrev
		var date time.Time
		if err := rows.Scan(&rb.Year, &date, &rb.Title, &rb.Summary, &rb.URL); err != nil {
			return nil, err
		}
		rb.Date = date.Format("2006-01-02")
		result = append(result, rb)
	}
	return result, rows.Err()
}

func (r *AgencyIntelRepository) GetDecisions(ctx context.Context, slug string) ([]domain.AgencyDecision, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT date, title, doc_type, summary, url
		FROM agency_decisions
		WHERE agency_slug = $1
		ORDER BY date DESC
		LIMIT 5
	`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.AgencyDecision
	for rows.Next() {
		var d domain.AgencyDecision
		var date time.Time
		if err := rows.Scan(&date, &d.Title, &d.DocType, &d.Summary, &d.URL); err != nil {
			return nil, err
		}
		d.Date = date.Format("2006-01-02")
		result = append(result, d)
	}
	return result, rows.Err()
}
