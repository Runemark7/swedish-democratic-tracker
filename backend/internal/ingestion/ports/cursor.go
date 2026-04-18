package ports

import (
	"context"
	"time"
)

// Cursor represents the last-fetched position for a data type.
type Cursor struct {
	DataType  string
	LastDate  *time.Time
	LastID    string
	UpdatedAt time.Time
}

// CursorRepository reads and writes ingestion cursors.
type CursorRepository interface {
	Get(ctx context.Context, dataType string) (*Cursor, error)
	Upsert(ctx context.Context, c Cursor) error
}
