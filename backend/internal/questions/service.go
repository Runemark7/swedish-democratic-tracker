// Package questions ingests skriftliga frågor, svar and interpellationer and
// links them to motions by rule.
package questions

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"riksdagskollen/internal/questions/domain"
	"riksdagskollen/internal/questions/ports"
)

type Service struct {
	repo ports.Repository
	rd   ports.RiksdagenClient
}

func NewService(repo ports.Repository, rd ports.RiksdagenClient) *Service {
	return &Service{repo: repo, rd: rd}
}

var Kinds = []string{domain.KindFraga, domain.KindInterpellation}

func (s *Service) StoredVersions(ctx context.Context, kind, rm string) (map[string]time.Time, error) {
	return s.repo.StoredVersions(ctx, kind, rm)
}

func (s *Service) List(ctx context.Context, kind, rm string, page, size int) ([]ports.Ref, int, error) {
	return s.rd.List(ctx, kind, rm, page, size)
}

func (s *Service) RecordExpected(ctx context.Context, rm, kind string, expected int) error {
	return s.repo.UpsertCoverage(ctx, rm, kind, expected)
}

// FetchAndStore reads one document and stores it.
func (s *Service) FetchAndStore(ctx context.Context, kind, dokID string) error {
	q, err := s.rd.Fetch(ctx, kind, dokID)
	if err != nil {
		return err
	}
	return s.repo.Upsert(ctx, q)
}

// Relink recomputes the motion links of every question in a riksmöte from the
// stored text and references. It writes only the rules' result, so repeating it
// changes nothing, and it picks up motions ingested after the question.
func (s *Service) Relink(ctx context.Context, rm string) (int, error) {
	qs, err := s.repo.ForLinking(ctx, rm)
	if err != nil {
		return 0, err
	}
	linked := 0
	held := func(ref string) (string, bool) {
		rmq, bet := splitRef(ref)
		id, ok, err := s.repo.MotionDokID(ctx, rmq, bet)
		if err != nil {
			slog.Warn("questions: motion lookup failed", "ref", ref, "error", err)
			return "", false
		}
		return id, ok
	}
	for _, q := range qs {
		links := domain.Resolve(q, held)
		if err := s.repo.ReplaceLinks(ctx, q.DokID, links); err != nil {
			return linked, fmt.Errorf("link %s: %w", q.DokID, err)
		}
		linked += len(links)
	}
	return linked, nil
}

func splitRef(ref string) (rm, bet string) {
	for i := len(ref) - 1; i >= 0; i-- {
		if ref[i] == ':' {
			return ref[:i], ref[i+1:]
		}
	}
	return ref, ""
}

func (s *Service) Coverage(ctx context.Context) ([]domain.Coverage, error) {
	return s.repo.Coverage(ctx)
}

func (s *Service) ByMember(ctx context.Context, id string) ([]domain.Item, error) {
	return s.repo.ByMember(ctx, id, 500)
}

func (s *Service) ByMotion(ctx context.Context, rm, bet string) ([]domain.Item, error) {
	return s.repo.ByMotion(ctx, rm, bet, 200)
}

func (s *Service) Get(ctx context.Context, kind, rm, bet string) (*domain.Detail, error) {
	d, err := s.repo.Get(ctx, kind, rm, bet)
	if err != nil {
		return nil, err
	}
	if d.Kind == domain.KindInterpellation {
		d.Debate, err = s.repo.Debate(ctx, d)
		if err != nil {
			return nil, err
		}
	}
	return d, nil
}
