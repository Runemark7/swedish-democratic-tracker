package ports

import (
	"context"
	"riksdagskollen/internal/parties/domain"
)

type PartiesRepository interface {
	ListAll(ctx context.Context) ([]*domain.Party, error)
	GetByCode(ctx context.Context, code string) (*domain.Party, error)
}
