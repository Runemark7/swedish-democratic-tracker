package ports

import (
	"context"

	"riksdagskollen/internal/politicians/domain"
)

type RiksdagenMemberClient interface {
	FetchMembers(ctx context.Context, party, status string) ([]*domain.Politician, error)
}
