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
		SELECT id, label, description, raw, target, worse_higher, unit, trend, delta,
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
			&k.ID, &k.Label, &k.Description, &k.Raw, &k.Target,
			&k.WorseHigher, &k.Unit, &k.Trend, &k.Delta,
			&k.Note, &k.SourceURL, &k.Year,
		); err != nil {
			return nil, err
		}
		kpis = append(kpis, k)
	}
	return kpis, rows.Err()
}
