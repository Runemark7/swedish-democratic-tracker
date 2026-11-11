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

// FetchAndStore fetches one party-riksmöte and returns the rows stored, so the
// caller can detect truncation and track the newest record seen. SyncSpeeches
// discards both, leaving no way to know how much was actually retrieved.
func (s *Service) FetchAndStore(ctx context.Context, f ports.FetchSpeechesFilter) ([]*domain.Speech, error) {
	ss, err := s.riksdagen.FetchSpeeches(ctx, f)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpsertMany(ctx, ss); err != nil {
		return nil, err
	}
	return ss, nil
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
		if err := s.enrichOne(ctx, sp); err != nil {
			skipped++
			continue
		}
		enriched++
	}
	return enriched, skipped, nil
}

// enrichOne fetches and stores the text of one speech.
func (s *Service) enrichOne(ctx context.Context, sp *domain.Speech) error {
	text, err := s.riksdagen.FetchSpeechText(ctx, sp.DokID, sp.AnforandeNummer)
	if err != nil {
		return err
	}
	if text == "" {
		// Mark with a single space so we stop retrying. The handler
		// trims this when serving, so users never see it.
		text = " "
	}
	return s.repo.UpdateText(ctx, sp.ID, text)
}

func (s *Service) ListByDocument(ctx context.Context, dokID string) ([]*domain.Speech, error) {
	return s.repo.ListByDocument(ctx, dokID)
}

func (s *Service) ListByParty(ctx context.Context, party string, limit int) ([]*domain.Speech, error) {
	return s.repo.ListByParty(ctx, party, limit)
}

func (s *Service) ListUnprocessed(ctx context.Context, limit int) ([]*domain.Speech, error) {
	return s.repo.ListUnprocessed(ctx, limit)
}

func (s *Service) MarkProcessed(ctx context.Context, id int) error {
	return s.repo.MarkProcessed(ctx, id)
}

// RelinkBatch is how many speeches one relink step reads.
const RelinkBatch = 2000

// LinkStats is the outcome of RelinkDebates.
type LinkStats struct {
	Seen, Linked, Unlinked, Changed int
}

// RelinkDebates links every stored speech to the betänkande it debates (see
// domain.LinkSpeech) and stores the rule used. It walks the speeches in id
// order in batches and writes only the rows whose link changed, so running it
// again over unchanged data writes nothing, and an interrupted run is simply
// run again.
func (s *Service) RelinkDebates(ctx context.Context) (LinkStats, error) {
	return s.RelinkDebatesIn(ctx, "")
}

// RelinkDebatesIn is RelinkDebates limited to one riksmöte's speeches; an
// empty riksmöte takes them all.
func (s *Service) RelinkDebatesIn(ctx context.Context, riksmote string) (LinkStats, error) {
	var st LinkStats
	cands, err := s.repo.DebateCandidates(ctx)
	if err != nil {
		return st, err
	}
	byDay := map[string][]domain.DebateCandidate{}
	for _, c := range cands {
		k := c.DebateDate.Format("2006-01-02")
		byDay[k] = append(byDay[k], c)
	}
	after := 0
	for {
		if err := ctx.Err(); err != nil {
			return st, err
		}
		rows, err := s.repo.ListForLinking(ctx, riksmote, after, RelinkBatch)
		if err != nil {
			return st, err
		}
		if len(rows) == 0 {
			return st, nil
		}
		var updates []ports.LinkUpdate
		for _, r := range rows {
			after = r.ID
			st.Seen++
			l, ok := domain.LinkSpeech(domain.SpeechForLinking{
				RelatedDokID: r.RelatedDokID, TopicHeading: r.TopicHeading, Date: r.Date,
			}, byDay[r.Date.Format("2006-01-02")])
			if ok {
				st.Linked++
			} else {
				st.Unlinked++
			}
			if l.BetankandeDokID != r.CurrentDokID || l.Rule != r.CurrentRule {
				updates = append(updates, ports.LinkUpdate{ID: r.ID, DokID: l.BetankandeDokID, Rule: l.Rule})
			}
		}
		if len(updates) > 0 {
			if err := s.repo.SetLinks(ctx, updates); err != nil {
				return st, err
			}
			st.Changed += len(updates)
		}
	}
}
