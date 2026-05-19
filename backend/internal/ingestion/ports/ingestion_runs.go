package ports

import "context"

// IngestionRunRepository records the start and finish of each ingestion worker run.
type IngestionRunRepository interface {
	StartRun(ctx context.Context, workerName string) (int64, error)
	FinishRun(ctx context.Context, id int64, rowsAffected int, runErr error) error
}
