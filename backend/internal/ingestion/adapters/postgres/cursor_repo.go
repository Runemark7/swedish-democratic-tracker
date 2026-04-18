package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/ingestion/ports"
)

type CursorRepository struct {
	db *pgxpool.Pool
}

func NewCursorRepository(db *pgxpool.Pool) *CursorRepository {
	return &CursorRepository{db: db}
}

func (r *CursorRepository) Get(ctx context.Context, dataType string) (*ports.Cursor, error) {
	const q = `SELECT data_type, last_date, last_id, updated_at
		FROM ingestion_cursors WHERE data_type = $1`

	var c ports.Cursor
	var lastDate *time.Time
	var lastID *string
	err := r.db.QueryRow(ctx, q, dataType).Scan(&c.DataType, &lastDate, &lastID, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // no cursor yet — first run
	}
	if err != nil {
		return nil, err
	}
	c.LastDate = lastDate
	if lastID != nil {
		c.LastID = *lastID
	}
	return &c, nil
}

func (r *CursorRepository) Upsert(ctx context.Context, c ports.Cursor) error {
	const q = `INSERT INTO ingestion_cursors (data_type, last_date, last_id, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (data_type) DO UPDATE SET
			last_date  = EXCLUDED.last_date,
			last_id    = EXCLUDED.last_id,
			updated_at = now()`

	_, err := r.db.Exec(ctx, q, c.DataType, c.LastDate, c.LastID)
	return err
}
