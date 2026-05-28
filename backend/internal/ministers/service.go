package ministers

import (
	"context"

	"riksdagskollen/internal/ministers/domain"
	"riksdagskollen/internal/ministers/ports"
)

type Service struct {
	repo     ports.MinistersRepository
	rdClient ports.ProposalsClient
}

func NewService(repo ports.MinistersRepository, rdClient ports.ProposalsClient) *Service {
	return &Service{repo: repo, rdClient: rdClient}
}

func (s *Service) ListActive(ctx context.Context) ([]*domain.Minister, error) {
	return s.repo.ListActive(ctx)
}

func (s *Service) GetDetail(ctx context.Context, id string) (*domain.Minister, []*domain.Proposal, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil || m == nil {
		return m, nil, err
	}
	props, err := s.repo.ListProposalsByDept(ctx, m.DepartmentCode, 20)
	return m, props, err
}

func (s *Service) SyncProposals(ctx context.Context, deptCode string, years []string) error {
	for _, year := range years {
		proposals, err := s.rdClient.FetchProposals(ctx, deptCode, year)
		if err != nil {
			return err
		}
		for _, p := range proposals {
			if err := s.repo.UpsertProposal(ctx, p); err != nil {
				return err
			}
		}
	}
	return nil
}
