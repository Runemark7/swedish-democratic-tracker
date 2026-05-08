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
}
