package speeches

import (
	"context"

	"riksdagskollen/internal/speeches/domain"
	"riksdagskollen/internal/speeches/ports"
)

// MissingTextIn lists up to limit speeches of the riksmöte with no text.
func (s *Service) MissingTextIn(ctx context.Context, riksmote string, limit int) ([]*domain.Speech, error) {
	return s.repo.MissingTextIn(ctx, riksmote, limit)
}

// EnrichOne fetches and stores the body of one speech.
func (s *Service) EnrichOne(ctx context.Context, sp *domain.Speech) error {
	return s.enrichOne(ctx, sp)
}

// Store upserts speeches fetched elsewhere. A speech whose politician has no
// row is not stored (the caller creates the speakers first).
func (s *Service) Store(ctx context.Context, ss []*domain.Speech) error {
	return s.repo.UpsertMany(ctx, ss)
}

func (s *Service) RiksmoteCounts(ctx context.Context, riksmote string) (domain.RiksmoteCounts, error) {
	return s.repo.RiksmoteCounts(ctx, riksmote)
}

func (s *Service) StorageSizes(ctx context.Context) (table, index int64, err error) {
	return s.repo.StorageSizes(ctx)
}

func (s *Service) SaveBackfill(ctx context.Context, run domain.BackfillRun) error {
	return s.repo.SaveBackfill(ctx, run)
}

func (s *Service) ListBackfill(ctx context.Context) ([]domain.Backfill, error) {
	return s.repo.ListBackfill(ctx)
}

// Fetch reads one party-riksmöte from Riksdagen without storing it.
func (s *Service) Fetch(ctx context.Context, f ports.FetchSpeechesFilter) ([]*domain.Speech, error) {
	return s.riksdagen.FetchSpeeches(ctx, f)
}
