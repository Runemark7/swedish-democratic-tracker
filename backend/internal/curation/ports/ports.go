package ports

import (
	"context"

	"riksdagskollen/internal/curation/domain"
)

// Store is the transactional view of the data a write touches. Every method
// runs inside the transaction Repository.InTx opened.
type Store interface {
	// LockRequest serialises requests that share an idempotency key.
	LockRequest(ctx context.Context, key string) error
	EntryByKey(ctx context.Context, key string) (*domain.Entry, error)
	EntryByID(ctx context.Context, id int64) (*domain.Entry, error)
	AppendEntry(ctx context.Context, e *domain.Entry) (*domain.Entry, error)

	// ValmanifestByKey / ValmanifestByID lock the row; nil when absent.
	ValmanifestByKey(ctx context.Context, party string, year int, title string) (*domain.Valmanifest, error)
	ValmanifestByID(ctx context.Context, id int64) (*domain.Valmanifest, error)
	// InsertValmanifest inserts v (with v.ID when non-zero) and returns it with its id.
	// Unknown party → domain.ErrNotFound-wrapped ValidationError; duplicate key → ErrConflict.
	InsertValmanifest(ctx context.Context, v domain.Valmanifest) (*domain.Valmanifest, error)
	UpdateValmanifest(ctx context.Context, v domain.Valmanifest) error
	// DeleteValmanifest returns ErrConflict when other data still refers to it.
	DeleteValmanifest(ctx context.Context, id int64) error

	// LoefteByID locks the row; nil when absent.
	// LoefteByQuote finds a non-struck Löfte with the same party, Valmanifest and quote.
	LoefteByQuote(ctx context.Context, party string, valmanifestID int64, quote string) (*domain.Loefte, error)
	LoefteByID(ctx context.Context, id int64) (*domain.Loefte, error)
	// InsertLoefte inserts l (with l.ID when non-zero). Unknown party or
	// Valmanifest → ValidationError. A proposed Löfte gets no approved_at.
	InsertLoefte(ctx context.Context, l domain.Loefte) (*domain.Loefte, error)
	UpdateLoefte(ctx context.Context, l domain.Loefte) error
	// DeleteLoefte removes a row (undo of a propose); ErrConflict when referenced.
	DeleteLoefte(ctx context.Context, id int64) error
}

type Repository interface {
	// InTx runs fn in one transaction: committed if fn returns nil.
	InTx(ctx context.Context, fn func(Store) error) error
	ListEntries(ctx context.Context, limit int, beforeID int64) ([]*domain.Entry, error)
}
