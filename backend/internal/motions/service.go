// Package motions ingests motions from Riksdagen and serves them.
package motions

import (
	"context"
	"time"

	"riksdagskollen/internal/motions/domain"
	"riksdagskollen/internal/motions/ports"
)

// Service ties the Riksdagen client and the repository together.
type Service struct {
	repo ports.Repository
	rd   ports.RiksdagenClient
}

func NewService(repo ports.Repository, rd ports.RiksdagenClient) *Service {
	return &Service{repo: repo, rd: rd}
}

func (s *Service) Get(ctx context.Context, riksmote, beteckning string) (*domain.Motion, error) {
	return s.repo.Get(ctx, riksmote, beteckning)
}

func (s *Service) Coverage(ctx context.Context) ([]domain.Coverage, error) {
	return s.repo.Coverage(ctx)
}

func (s *Service) ListMotions(ctx context.Context, rm string, page, size int) ([]ports.MotionRef, int, error) {
	return s.rd.ListMotions(ctx, rm, page, size)
}

func (s *Service) StoredVersions(ctx context.Context, rm string) (map[string]time.Time, error) {
	return s.repo.StoredVersions(ctx, rm)
}

// FetchAndStore reads one motion and writes it. Writing the same version twice
// leaves the same rows.
func (s *Service) FetchAndStore(ctx context.Context, dokID string) error {
	m, err := s.rd.FetchMotion(ctx, dokID)
	if err != nil {
		return err
	}
	return s.repo.Upsert(ctx, m)
}

func (s *Service) RecordExpected(ctx context.Context, rm string, expected int) error {
	return s.repo.UpsertCoverage(ctx, rm, domain.SourceRiksdagenDokumentlista, expected)
}
