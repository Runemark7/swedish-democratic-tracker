package ports

import (
	"context"

	"riksdagskollen/internal/datacoverage/domain"
)

// Repository reads the ingestion coverage records.
type Repository interface {
	// List returns one row per source and riksmöte that has a coverage record,
	// newest riksmöte first.
	List(ctx context.Context) ([]domain.Row, error)
}
