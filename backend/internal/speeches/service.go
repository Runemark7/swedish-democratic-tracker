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

// EnrichMissingText fetches the prose body for up to `limit` speeches
// that currently have no text. Each successful fetch is upserted; rows
// where Riksdagen returns no text get a single-space placeholder so the
// next pass doesn't keep retrying the same dead anförande.
func (s *Service) EnrichMissingText(ctx context.Context, limit int) (enriched, skipped int, err error) {
	rows, err := s.repo.ListMissingText(ctx, limit)
	if err != nil {
		return 0, 0, err
	}
	for _, sp := range rows {
		if ctx.Err() != nil {
			return enriched, skipped, ctx.Err()
		}
		text, err := s.riksdagen.FetchSpeechText(ctx, sp.DokID, sp.AnforandeNummer)
		if err != nil {
			skipped++
			continue
		}
		if text == "" {
			// Mark with a single space so we stop retrying. The handler
			// trims this when serving, so users never see it.
			text = " "
		}
		if err := s.repo.UpdateText(ctx, sp.ID, text); err != nil {
			skipped++
			continue
		}
		enriched++
	}
	return enriched, skipped, nil
}

func (s *Service) ListUnprocessed(ctx context.Context, limit int) ([]*domain.Speech, error) {
	return s.repo.ListUnprocessed(ctx, limit)
}

func (s *Service) MarkProcessed(ctx context.Context, id int) error {
	return s.repo.MarkProcessed(ctx, id)
}
