package ports

import (
	"context"

	"riksdagskollen/internal/politicians/domain"
)

type RiksdagenMemberClient interface {
	FetchMembers(ctx context.Context, party, status string) ([]*domain.Politician, error)
	// FetchPerson reads one person by intressent_id, former members included.
	// A person Riksdagen does not know returns nil, nil.
	FetchPerson(ctx context.Context, intressentID string) (*domain.Politician, error)
}
