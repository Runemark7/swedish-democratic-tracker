package ports

import (
	"context"

	"riksdagskollen/internal/riksdag/domain"
)

// RegisterEntry is one row scraped from SCB Myndighetsregistret.
type RegisterEntry struct {
	OrgNumber       string
	Slug            string
	Name            string
	Type            string
	PrincipalBody   string
	UnderGovernment bool
	Website         string
	SFS             string
}

type RegisterClient interface {
	FetchRegister(ctx context.Context) ([]RegisterEntry, error)
}

type AuthorityFilter struct {
	Query           string
	UnderGovernment *bool
	Limit           int
	Offset          int
}

type AuthorityRepository interface {
	UpsertAuthorities(ctx context.Context, items []domain.RegisteredAuthority) (int, error)
	List(ctx context.Context, f AuthorityFilter) ([]domain.RegisteredAuthority, error)
	Count(ctx context.Context, f AuthorityFilter) (int, error)
}
