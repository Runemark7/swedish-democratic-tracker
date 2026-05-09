package ports

import (
	"context"

	"riksdagskollen/internal/speeches/domain"
)

type SpeechRepository interface {
	GetByID(ctx context.Context, id int) (*domain.Speech, error)
	ListByPolitician(ctx context.Context, politicianID string) ([]*domain.Speech, error)
	ListRecent(ctx context.Context, limit int) ([]*domain.Speech, error)
	UpsertMany(ctx context.Context, ss []*domain.Speech) error
	ListUnprocessed(ctx context.Context, limit int) ([]*domain.Speech, error)
	MarkProcessed(ctx context.Context, id int) error
	// ListMissingText returns speeches with empty speech_text, ordered
	// newest first so user-facing pages get backfilled first.
	ListMissingText(ctx context.Context, limit int) ([]*domain.Speech, error)
	// UpdateText overwrites speech_text for one row.
	UpdateText(ctx context.Context, id int, text string) error
	// ListByDocument returns speeches whose rel_dok_id matches the
	// given betänkande document id, ordered by anforande_nummer ASC
	// (chronological within the debate).
	ListByDocument(ctx context.Context, dokID string) ([]*domain.Speech, error)
	// ListByParty returns speeches by all members of one party,
	// newest first, capped at limit (default 50, max 200).
	ListByParty(ctx context.Context, party string, limit int) ([]*domain.Speech, error)
}
