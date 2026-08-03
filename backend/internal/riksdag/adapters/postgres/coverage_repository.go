package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/riksdag/domain"
)

type CoverageRepository struct {
	db *pgxpool.Pool
}

func NewCoverageRepository(db *pgxpool.Pool) *CoverageRepository {
	return &CoverageRepository{db: db}
}

// GetRecordCoverage reports how much of the given mandate period we hold.
//
// Only the riksmöten belonging to the period are counted, so the figure cannot
// drift once the next parliament starts filing votes.
func (r *CoverageRepository) GetRecordCoverage(ctx context.Context, code string) (*domain.RecordCoverage, error) {
	var (
		mandate   domain.MandatePeriod
		riksmoten []string
	)
	err := r.db.QueryRow(ctx, `
		SELECT code, label, end_date < CURRENT_DATE, riksmoten
		FROM mandate_periods WHERE code = $1`, code).
		Scan(&mandate.Code, &mandate.Label, &mandate.Ended, &riksmoten)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT riksmote, COALESCE(expected_voteringar, 0), ingested_voteringar,
		       unreachable_voteringar, to_char(last_vote_date, 'YYYY-MM-DD')
		FROM ingestion_coverage
		WHERE riksmote = ANY($1)
		ORDER BY riksmote`, riksmoten)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := &domain.RecordCoverage{Mandate: mandate, ByRiksmote: []domain.RiksmoteRecord{}}
	for rows.Next() {
		var rec domain.RiksmoteRecord
		if err := rows.Scan(&rec.Riksmote, &rec.Expected, &rec.Ingested,
			&rec.Unreachable, &rec.LastDecisionDate); err != nil {
			return nil, err
		}
		out.Expected += rec.Expected
		out.Ingested += rec.Ingested
		out.Unreachable += rec.Unreachable
		// The newest decision across the period. Riksmöten sort chronologically,
		// so the last non-null date wins.
		if rec.LastDecisionDate != nil {
			out.LastDecisionDate = rec.LastDecisionDate
		}
		out.ByRiksmote = append(out.ByRiksmote, rec)
	}
	return out, rows.Err()
}
