package politicians

import (
	"context"

	"riksdagskollen/internal/politicians/domain"
	"riksdagskollen/internal/politicians/ports"
)

type Service struct {
	repo      ports.PoliticianRepository
	riksdagen ports.RiksdagenMemberClient
}

func NewService(repo ports.PoliticianRepository, riksdagen ports.RiksdagenMemberClient) *Service {
	return &Service{repo: repo, riksdagen: riksdagen}
}

// SyncParty fetches all active members for a party and upserts them.
func (s *Service) SyncParty(ctx context.Context, party string) error {
	members, err := s.riksdagen.FetchMembers(ctx, party, "tjanstgorande")
	if err != nil {
		return err
	}
	return s.repo.UpsertMany(ctx, members)
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.Politician, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, f ports.ListFilter) (ports.ListResult, error) {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.PageSize == 0 {
		f.PageSize = 50
	}
	return s.repo.List(ctx, f)
}
