package speeches

import (
	"context"

	"riksdagskollen/internal/speeches/domain"
	"riksdagskollen/internal/speeches/ports"
)

type Service struct {
	repo      ports.SpeechRepository
	riksdagen ports.RiksdagenSpeechClient
}

func NewService(repo ports.SpeechRepository, riksdagen ports.RiksdagenSpeechClient) *Service {
	return &Service{repo: repo, riksdagen: riksdagen}
}

func (s *Service) SyncSpeeches(ctx context.Context, f ports.FetchSpeechesFilter) error {
	speeches, err := s.riksdagen.FetchSpeeches(ctx, f)
	if err != nil {
		return err
	}
	return s.repo.UpsertMany(ctx, speeches)
}

func (s *Service) GetByID(ctx context.Context, id int) (*domain.Speech, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByPolitician(ctx context.Context, politicianID string) ([]*domain.Speech, error) {
	return s.repo.ListByPolitician(ctx, politicianID)
}

func (s *Service) ListRecent(ctx context.Context, limit int) ([]*domain.Speech, error) {
	return s.repo.ListRecent(ctx, limit)
}

func (s *Service) ListUnprocessed(ctx context.Context, limit int) ([]*domain.Speech, error) {
	return s.repo.ListUnprocessed(ctx, limit)
}

func (s *Service) MarkProcessed(ctx context.Context, id int) error {
	return s.repo.MarkProcessed(ctx, id)
}
