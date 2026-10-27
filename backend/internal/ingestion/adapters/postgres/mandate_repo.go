package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/ingestion/ports"
)

// MandateRepository reads mandate_periods.
type MandateRepository struct {
	db *pgxpool.Pool
}

func NewMandateRepository(db *pgxpool.Pool) *MandateRepository {
	return &MandateRepository{db: db}
}

func (r *MandateRepository) ListMandatePeriods(ctx context.Context) ([]ports.MandatePeriod, error) {
	rows, err := r.db.Query(ctx, `
		SELECT code, start_date, end_date, riksmoten
		FROM mandate_periods
		ORDER BY start_date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ports.MandatePeriod
	for rows.Next() {
		var p ports.MandatePeriod
		if err := rows.Scan(&p.Code, &p.Start, &p.End, &p.Riksmoten); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
