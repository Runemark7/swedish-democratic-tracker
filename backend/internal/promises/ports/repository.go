package ports

import (
	"context"

	"riksdagskollen/internal/promises/domain"
)

type PromiseRepository interface {
	Create(ctx context.Context, p *domain.Promise) error
	ListByPolitician(ctx context.Context, politicianID, topic string) ([]*domain.Promise, error)
	GetByID(ctx context.Context, id int) (*domain.Promise, error)
}
