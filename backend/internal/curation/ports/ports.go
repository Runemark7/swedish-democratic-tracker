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

	// VoteringForLoefte resolves a votering by Riksdagen's votering_id and
	// says whether it lies inside the mandate period of Löfte l. nil: no such votering.
	VoteringForLoefte(ctx context.Context, voteringID string, l *domain.Loefte) (*domain.Votering, error)
	// KopplingByPair locks the row; nil when absent.
	KopplingByPair(ctx context.Context, loefteID, voteringRef int64) (*domain.Koppling, error)
	InsertKoppling(ctx context.Context, loefteID, voteringRef int64, k domain.Koppling) (*domain.Koppling, error)

	// Veckobatch (issue #160).
	// LockBatches serialises batch creation.
	LockBatches(ctx context.Context) error
	// OpenSuggestions lists foreslagen kopplingar, rolled-over ones first
	// (oldest batch first), then strongest first; at most max.
	OpenSuggestions(ctx context.Context, max int) ([]domain.OpenSuggestion, error)
	CloseOpenBatches(ctx context.Context) error
	InsertBatch(ctx context.Context) (int64, error)
	// InsertBatchItem numbers o in the batch; o names a Löfte or a motion koppling.
	InsertBatchItem(ctx context.Context, batchID int64, nr int, o domain.OpenSuggestion) error
	// BatchByID with the batch row and its kopplingar locked; nil when absent.
	BatchByID(ctx context.Context, id int64) (*domain.Batch, error)
	// DecideItem sets the koppling's state/direction (decided_at follows the
	// state) and the item's answer; answer "" clears it.
	DecideItem(ctx context.Context, batchID int64, nr int, kopplingID int64, state, direction, answer string) error
	CloseBatch(ctx context.Context, id int64) error
	// Motion kopplingar (issue #163).
	// ResolveMotionSuggestion turns the named question/motion/yrkanden into a
	// draft MotionKoppling with the stored dok ids. A reference that does not
	// exist is a ValidationError.
	ResolveMotionSuggestion(ctx context.Context, s domain.MotionSuggestion) (*domain.MotionKoppling, error)
	// MotionKopplingByPair locks the row of the same link; nil when absent.
	MotionKopplingByPair(ctx context.Context, k domain.MotionKoppling) (*domain.MotionKoppling, error)
	InsertMotionKoppling(ctx context.Context, k domain.MotionKoppling) (*domain.MotionKoppling, error)
	// MotionKopplingByID locks the row and includes its latest batch item; nil when absent.
	MotionKopplingByID(ctx context.Context, id int64) (*domain.MotionKoppling, error)
	// DecideMotionItem sets the link's state (decided_at follows it) and the
	// item's answer; answer "" clears it.
	DecideMotionItem(ctx context.Context, batchID int64, nr int, id int64, state, answer string) error

	// KopplingByID locks the row and includes its batch item; nil when absent.
	KopplingByID(ctx context.Context, id int64) (*domain.Koppling, error)
}

type Repository interface {
	BatchByID(ctx context.Context, id int64) (*domain.Batch, error)
	// LatestBatch is the newest Veckobatch; nil when there is none.
	LatestBatch(ctx context.Context) (*domain.Batch, error)
	// InTx runs fn in one transaction: committed if fn returns nil.
	InTx(ctx context.Context, fn func(Store) error) error
	// Candidates lists decided voteringar new since the last closed Veckobatch,
	// inside a mandate period, each with the approved Löften it might test.
	Candidates(ctx context.Context, limit int, afterRef int64, mandatePeriod string) (*domain.Candidates, error)
	ListEntries(ctx context.Context, limit int, beforeID int64) ([]*domain.Entry, error)
}
