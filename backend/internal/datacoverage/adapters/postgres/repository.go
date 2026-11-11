package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/datacoverage/domain"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// List unions the per-source coverage records. "Held" is counted live from the
// stored rows where the table exists for it, so the numerator is never a stale
// snapshot; "expected" is Riksdagen's own figure. A row is complete when the
// operator's backfill marked it finished, or when what we hold reaches what the
// listing reports. last_update is the newest of the record's own timestamps.
func (r *Repository) List(ctx context.Context) ([]domain.Row, error) {
	rows, err := r.db.Query(ctx, `
		WITH motioner AS (
			SELECT k.riksmote,
			       coalesce(c.expected, b.expected, 0) AS expected,
			       (SELECT count(*) FROM motions m WHERE m.riksmote = k.riksmote) AS held,
			       b.completed_at,
			       greatest(c.last_enumerated_at, b.updated_at) AS updated
			FROM (SELECT riksmote FROM motion_coverage
			      UNION SELECT riksmote FROM motion_backfill WHERE source = 'motioner') k
			LEFT JOIN motion_coverage c ON c.riksmote = k.riksmote
			LEFT JOIN motion_backfill b ON b.riksmote = k.riksmote AND b.source = 'motioner'
		),
		voteringar_held AS (
			SELECT session AS riksmote, count(DISTINCT beteckning || ':' || forslagspunkt) AS n
			FROM voteringar GROUP BY session
		),
		u AS (
			SELECT 'motioner' AS source, riksmote, expected, held, completed_at, updated FROM motioner
			UNION ALL
			SELECT 'betankanden', riksmote, expected, held, completed_at, updated_at
			FROM motion_backfill WHERE source = 'betankanden'
			UNION ALL
			SELECT 'anforanden', b.riksmote, b.expected,
			       (SELECT count(*) FROM speeches s WHERE s.riksmote = b.riksmote),
			       b.completed_at, b.updated_at
			FROM speech_backfill b
			UNION ALL
			SELECT CASE c.kind WHEN 'fraga' THEN 'fragor' ELSE 'interpellationer' END,
			       c.riksmote, c.expected,
			       (SELECT count(*) FROM questions q WHERE q.riksmote = c.riksmote AND q.kind = c.kind),
			       NULL::timestamptz, c.last_enumerated_at
			FROM question_coverage c
			UNION ALL
			SELECT 'voteringar', c.riksmote, coalesce(c.expected_voteringar, 0),
			       coalesce(h.n, 0), NULL::timestamptz, c.checked_at
			FROM ingestion_coverage c LEFT JOIN voteringar_held h ON h.riksmote = c.riksmote
		)
		SELECT source, riksmote,
		       CASE WHEN completed_at IS NOT NULL OR (expected > 0 AND held >= expected)
		            THEN 'complete' ELSE 'partial' END,
		       expected, held, coalesce(updated, completed_at)
		FROM u
		WHERE coalesce(updated, completed_at) IS NOT NULL
		ORDER BY riksmote DESC, source`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Row{}
	for rows.Next() {
		var d domain.Row
		if err := rows.Scan(&d.Source, &d.Riksmote, &d.Status, &d.Expected, &d.Held, &d.LastUpdate); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
