package ports

import (
	"context"

	"riksdagskollen/internal/politicians/domain"
)

type ListFilter struct {
	Party      string
	ActiveOnly bool
	Page       int
	PageSize   int
}

type ListResult struct {
	Politicians []*domain.Politician
	Total       int
}

type PoliticianRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Politician, error)
	List(ctx context.Context, f ListFilter) (ListResult, error)
	UpsertMany(ctx context.Context, pp []*domain.Politician) error
}
