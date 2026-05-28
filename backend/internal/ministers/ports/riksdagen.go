package ports

import (
	"context"
	"riksdagskollen/internal/ministers/domain"
)

type ProposalsClient interface {
	FetchProposals(ctx context.Context, deptCode string, year string) ([]*domain.Proposal, error)
}
