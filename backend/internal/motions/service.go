// Package motions ingests motions from Riksdagen and serves them.
package motions

import (
	"context"
	"strings"
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

// Pending lists betänkanden to read, at most limit.
func (s *Service) Pending(ctx context.Context, limit int) ([]ports.BetRef, error) {
	return s.repo.PendingBetankanden(ctx, limit)
}

// SyncBetankande reads one betänkande, stores it with its förslagspunkter and
// links every yrkande that points at it. Running it again leaves the same rows.
// found is false when Riksdagen lists no such betänkande.
func (s *Service) SyncBetankande(ctx context.Context, ref ports.BetRef) (found bool, err error) {
	hit, ok, err := s.rd.FindBetankande(ctx, ref.Riksmote, ref.Beteckning)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, s.repo.MarkBetankandeMissing(ctx, ref.Riksmote, ref.Beteckning)
	}
	b, err := s.rd.FetchBetankande(ctx, hit.DokID)
	if err != nil {
		return false, err
	}
	// Store under the key the yrkanden use, whatever case Riksdagen answered.
	b.Riksmote, b.Beteckning = ref.Riksmote, ref.Beteckning
	if err := s.repo.SaveBetankande(ctx, b); err != nil {
		return false, err
	}
	return true, s.Link(ctx, ref.Riksmote, ref.Beteckning)
}

// Link recomputes the links of every yrkande handled in rm:bet from the stored
// förslagspunkter.
func (s *Service) Link(ctx context.Context, rm, bet string) error {
	keys, err := s.repo.YrkandenHandledIn(ctx, rm, bet)
	if err != nil {
		return err
	}
	pts, err := s.repo.Punkter(ctx, rm, bet)
	if err != nil {
		return err
	}
	punkter := make([]domain.Punkt, 0, len(pts))
	byNumber := map[string]domain.Forslagspunkt{}
	for _, p := range pts {
		punkter = append(punkter, domain.Punkt{Number: p.Number, Text: p.Text})
		byNumber[p.Number] = p
	}
	for _, k := range keys {
		links := domain.LinkYrkande(domain.YrkandeRef{
			Riksmote: k.MotionRiksmote, Beteckning: k.MotionBeteckn, Number: k.Number, BehandlasIPunkt: k.BehandlasIPunkt,
		}, punkter)
		for i := range links {
			if links[i].Decision == "" {
				p := byNumber[links[i].Punkt]
				links[i].Decision = domain.PunktDecisionOf(p.Text, p.Vinnare)
			} else if v := strings.TrimSpace(byNumber[links[i].Punkt].Vinnare); v != "" && !strings.EqualFold(v, "utskottet") {
				// The utskott's proposal did not win; the text's verb is not the outcome.
				links[i].Decision = ""
			}
		}
		if err := s.repo.ReplaceLinks(ctx, k, rm, bet, links); err != nil {
			return err
		}
	}
	return nil
}

// Decisions returns the stored decisions of a motion's yrkanden.
func (s *Service) Decisions(ctx context.Context, motionDokID string) ([]domain.YrkandeDecision, error) {
	return s.repo.Decisions(ctx, motionDokID)
}
