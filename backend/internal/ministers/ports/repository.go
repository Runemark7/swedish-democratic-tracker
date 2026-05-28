package ports

import (
	"context"
	"riksdagskollen/internal/ministers/domain"
)

type MinistersRepository interface {
	ListActive(ctx context.Context) ([]*domain.Minister, error)
	GetByID(ctx context.Context, id string) (*domain.Minister, error)
	ListProposalsByDept(ctx context.Context, deptCode string, limit int) ([]*domain.Proposal, error)
	UpsertProposal(ctx context.Context, p *domain.Proposal) error
}
