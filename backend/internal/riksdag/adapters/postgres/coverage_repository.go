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

	// Ingested is counted from the votes we actually hold rather than read from
	// ingestion_coverage, which only a manual backfill writes. Snapshotting our
	// own row count froze the numerator between backfill runs and hid any
	// riksmöte that had votes but no coverage row: it counted as zero, or
	// vanished from the breakdown entirely.
	//
	// Expected, unreachable and last_vote_date still come from the snapshot,
	// because establishing them requires enumerating Riksdagen's listing and
	// votes carries no decision date of its own — only system_datum, which is
	// when Riksdagen last touched the record. checked_at travels with them so
	// their age is visible instead of implied.
	rows, err := r.db.Query(ctx, `
		WITH period AS (
			SELECT unnest($1::text[]) AS riksmote
		),
		held AS (
			SELECT session, COUNT(DISTINCT beteckning || ':' || forslagspunkt) AS n
			FROM voteringar
			WHERE session = ANY($1)
			GROUP BY session
		)
		SELECT p.riksmote,
		       COALESCE(c.expected_voteringar, 0),
		       COALESCE(h.n, 0),
		       COALESCE(c.unreachable_voteringar, 0),
		       to_char(c.last_vote_date, 'YYYY-MM-DD'),
		       to_char(c.checked_at, 'YYYY-MM-DD')
		FROM period p
		LEFT JOIN ingestion_coverage c ON c.riksmote = p.riksmote
		LEFT JOIN held h ON h.session = p.riksmote
		ORDER BY p.riksmote`, riksmoten)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := &domain.RecordCoverage{Mandate: mandate, ByRiksmote: []domain.RiksmoteRecord{}}
	for rows.Next() {
		var rec domain.RiksmoteRecord
		if err := rows.Scan(&rec.Riksmote, &rec.Expected, &rec.Ingested,
			&rec.Unreachable, &rec.LastDecisionDate, &rec.DenominatorCheckedAt); err != nil {
			return nil, err
		}
		// The period's denominator is only as fresh as its stalest riksmöte, so
		// the oldest check wins. A riksmöte whose denominator was never
		// established contributes nothing to Expected, so it cannot make the
		// stated age look better than it is either.
		if rec.DenominatorCheckedAt != nil &&
			(out.DenominatorCheckedAt == nil || *rec.DenominatorCheckedAt < *out.DenominatorCheckedAt) {
			out.DenominatorCheckedAt = rec.DenominatorCheckedAt
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
