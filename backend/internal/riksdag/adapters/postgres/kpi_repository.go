package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/riksdag/domain"
)

type KpiRepository struct {
	db *pgxpool.Pool
}

func NewKpiRepository(db *pgxpool.Pool) *KpiRepository {
	return &KpiRepository{db: db}
}

func (r *KpiRepository) ListKpis(ctx context.Context) ([]domain.Kpi, error) {
	const q = `
		SELECT id, label, description, raw, worse_higher, unit, trend, delta,
		       note, COALESCE(source_url, ''), year
		FROM riksdag_kpis
		WHERE year = (SELECT MAX(year) FROM riksdag_kpis)
		ORDER BY sort_order
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var kpis []domain.Kpi
	for rows.Next() {
		var k domain.Kpi
		if err := rows.Scan(
			&k.ID, &k.Label, &k.Description, &k.Raw,
			&k.WorseHigher, &k.Unit, &k.Trend, &k.Delta,
			&k.Note, &k.SourceURL, &k.Year,
		); err != nil {
			return nil, err
		}
		kpis = append(kpis, k)
	}
	return kpis, rows.Err()
}

func (r *KpiRepository) UpsertKpi(ctx context.Context, k domain.Kpi, sortOrder int) error {
	const q = `
		INSERT INTO riksdag_kpis
			(label, description, raw, worse_higher, unit, trend, delta, note, source_url, year, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (label, year) DO UPDATE SET
			description = EXCLUDED.description,
			raw         = EXCLUDED.raw,
			worse_higher = EXCLUDED.worse_higher,
			unit        = EXCLUDED.unit,
			trend       = EXCLUDED.trend,
			delta       = EXCLUDED.delta,
			note        = EXCLUDED.note,
			source_url  = EXCLUDED.source_url,
			sort_order  = EXCLUDED.sort_order
	`
	_, err := r.db.Exec(ctx, q,
		k.Label, k.Description, k.Raw, k.WorseHigher, k.Unit,
		k.Trend, k.Delta, k.Note, k.SourceURL, k.Year, sortOrder,
	)
	return err
}
