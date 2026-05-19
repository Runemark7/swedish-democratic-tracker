package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// IngestionRunRepository implements ports.IngestionRunRepository via PostgreSQL.
type IngestionRunRepository struct {
	db *pgxpool.Pool
}

func NewIngestionRunRepository(db *pgxpool.Pool) *IngestionRunRepository {
	return &IngestionRunRepository{db: db}
}

func (r *IngestionRunRepository) StartRun(ctx context.Context, workerName string) (int64, error) {
	const q = `
		INSERT INTO ingestion_runs (worker_name, started_at, status)
		VALUES ($1, NOW(), 'running')
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, q, workerName).Scan(&id)
	return id, err
}

func (r *IngestionRunRepository) FinishRun(ctx context.Context, id int64, rowsAffected int, runErr error) error {
	status := "success"
	var errMsg *string
	if runErr != nil {
		status = "error"
		msg := runErr.Error()
		errMsg = &msg
	}
	const q = `
		UPDATE ingestion_runs
		SET finished_at = NOW(),
		    status      = $2,
		    rows_affected = $3,
		    error_msg   = $4
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, q, id, status, rowsAffected, errMsg)
	return err
}
