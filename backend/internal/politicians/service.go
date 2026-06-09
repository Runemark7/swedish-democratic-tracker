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

// SyncAll fetches every currently-serving member and upserts them. It queries
// with an empty party filter so the riksdagen personlista endpoint returns all
// 349 members, including party-less ones (status "-", politiska vildar) that a
// per-party sync would miss.
func (s *Service) SyncAll(ctx context.Context) error {
	members, err := s.riksdagen.FetchMembers(ctx, "", "tjanstgorande")
	if err != nil {
		return err
	}
	return s.repo.UpsertMany(ctx, members)
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.Politician, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) NameByID(ctx context.Context, intressentID string) (string, error) {
	p, err := s.repo.GetByID(ctx, intressentID)
	if err != nil {
		return "", err
	}
	if p == nil {
		return "", nil
	}
	return p.FullName(), nil
}

func (s *Service) ImageURLByID(ctx context.Context, intressentID string) (string, error) {
	p, err := s.repo.GetByID(ctx, intressentID)
	if err != nil {
		return "", err
	}
	if p == nil {
		return "", nil
	}
	return p.ImageURL, nil
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
