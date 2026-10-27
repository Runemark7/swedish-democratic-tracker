package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/ingestion"
)

// FreshnessRepository reads the newest speech and vote we hold and the latest
// recorded outcome of each ingestion worker.
type FreshnessRepository struct {
	db *pgxpool.Pool
}

func NewFreshnessRepository(db *pgxpool.Pool) *FreshnessRepository {
	return &FreshnessRepository{db: db}
}

func (r *FreshnessRepository) Snapshot(ctx context.Context) (ingestion.Snapshot, error) {
	var snap ingestion.Snapshot
	var err error
	if snap.SpeechesLatest, err = scanTime(ctx, r.db,
		`SELECT MAX(date)::timestamptz FROM speeches`); err != nil {
		return ingestion.Snapshot{}, err
	}
	if snap.VotesLatest, err = scanTime(ctx, r.db,
		`SELECT MAX(system_datum) FROM voteringar`); err != nil {
		return ingestion.Snapshot{}, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT ON (g.worker_name)
			g.worker_name, g.last_success, g.last_finished, r.status
		FROM (
			SELECT worker_name,
				MAX(finished_at) FILTER (WHERE status = 'success') AS last_success,
				MAX(finished_at) AS last_finished
			FROM ingestion_runs
			WHERE finished_at IS NOT NULL
			GROUP BY worker_name
		) g
		JOIN ingestion_runs r
			ON r.worker_name = g.worker_name AND r.finished_at = g.last_finished
		ORDER BY g.worker_name, r.id DESC`)
	if err != nil {
		return ingestion.Snapshot{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var job ingestion.JobFreshness
		if err := rows.Scan(&job.Name, &job.LastSuccessAt, &job.LastFinishedAt, &job.LastStatus); err != nil {
			return ingestion.Snapshot{}, err
		}
		snap.Jobs = append(snap.Jobs, job)
	}
	return snap, rows.Err()
}

func scanTime(ctx context.Context, db *pgxpool.Pool, query string) (*time.Time, error) {
	var t *time.Time
	if err := db.QueryRow(ctx, query).Scan(&t); err != nil {
		return nil, err
	}
	return t, nil
}
