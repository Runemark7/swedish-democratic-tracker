package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/ingestion/ports"
)

type CoverageRepo struct {
	db *pgxpool.Pool
}

func NewCoverageRepo(db *pgxpool.Pool) *CoverageRepo {
	return &CoverageRepo{db: db}
}

// RiksmotenNeedingDenominator returns riksmöten of any period still running,
// plus any riksmöte whose denominator has never been established.
//
// The mandate periods are read from the database rather than hardcoded so that
// the set follows the election instead of needing a deploy on 14 September 2026.
func (r *CoverageRepo) RiksmotenNeedingDenominator(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		WITH all_rm AS (
			SELECT unnest(riksmoten) AS riksmote, end_date >= CURRENT_DATE AS active
			FROM mandate_periods
		)
		SELECT DISTINCT a.riksmote
		FROM all_rm a
		LEFT JOIN ingestion_coverage c ON c.riksmote = a.riksmote
		WHERE a.active OR c.expected_voteringar IS NULL
		ORDER BY a.riksmote`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var rm string
		if err := rows.Scan(&rm); err != nil {
			return nil, err
		}
		out = append(out, rm)
	}
	return out, rows.Err()
}

// UpsertDenominator writes one riksmöte's upstream counts and stamps checked_at.
//
// ingested_voteringar is deliberately not written. It is counted from the votes
// table when coverage is read, and keeping a second copy here is what allowed
// the published numerator to freeze between manual backfill runs.
func (r *CoverageRepo) UpsertDenominator(ctx context.Context, d ports.Denominator) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO ingestion_coverage
			(riksmote, expected_voteringar, unreachable_voteringar, last_vote_date, checked_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (riksmote) DO UPDATE SET
			expected_voteringar    = EXCLUDED.expected_voteringar,
			unreachable_voteringar = EXCLUDED.unreachable_voteringar,
			last_vote_date         = EXCLUDED.last_vote_date,
			checked_at             = now()`,
		d.Riksmote, d.Expected, d.Unreachable, d.LastVoteDate)
	return err
}
