// Package motions ingests motions from Riksdagen and serves them.
package motions

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"riksdagskollen/internal/motions/domain"
	"riksdagskollen/internal/motions/ports"
)

// Service ties the Riksdagen client and the repository together.
type Service struct {
	repo ports.Repository
	rd   ports.RiksdagenClient
	now  func() time.Time
}

// SetNow pins the clock used to tell a decided betänkande from a planned one.
func (s *Service) SetNow(now func() time.Time) { s.now = now }

func (s *Service) clock() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
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

// Pending lists betänkanden to read, at most limit. A non-empty rm keeps only
// those that motions of that riksmöte point at.
func (s *Service) Pending(ctx context.Context, rm string, limit int) ([]ports.BetRef, error) {
	return s.repo.PendingBetankanden(ctx, rm, limit)
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
	b.DebateDate = hit.DebateDate
	if err := s.repo.SaveBetankande(ctx, b); err != nil {
		return false, err
	}
	s.syncHTML(ctx, ref, b.DokID)
	return true, s.Link(ctx, ref.Riksmote, ref.Beteckning)
}

// syncHTML reads the betänkande's HTML for reservations and reasoning. A
// failure here does not fail the sync: the JSON part is already stored, the
// page shows the HTML parts as not available, and the read is retried after a
// day. A document Riksdagen serves no HTML for is recorded as read-and-empty.
func (s *Service) syncHTML(ctx context.Context, ref ports.BetRef, dokID string) {
	h, err := s.rd.FetchBetankandeHTML(ctx, dokID)
	if errors.Is(err, ports.ErrHTMLUnavailable) {
		h, err = &domain.BetankandeHTML{}, nil
	}
	if err == nil {
		err = s.repo.SaveBetankandeHTML(ctx, ref.Riksmote, ref.Beteckning, h)
	}
	if err != nil {
		slog.Warn("betankanden: html not stored", "rm", ref.Riksmote, "bet", ref.Beteckning, "error", err)
		if merr := s.repo.MarkHTMLAttempt(ctx, ref.Riksmote, ref.Beteckning); merr != nil {
			slog.Warn("betankanden: html attempt not recorded", "error", merr)
		}
	}
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

// ListItem is one row of the motion list with its computed outcome.
type ListItem struct {
	Motion  domain.Motion
	Outcome domain.Aggregate
	Cursor  domain.ListCursor
	// HandledIn lists the betänkanden that decided its yrkanden.
	HandledIn []string
}

// ListResult is one page. Next is nil when there is nothing after it.
type ListResult struct {
	Items []ListItem
	Next  *domain.ListCursor
}

const (
	listBatch = 200
	// listMaxBatches bounds one request when an outcome filter is rare; the
	// page then comes back short with a cursor to continue from.
	listMaxBatches = 25
)

// List returns up to limit motions, most recently decided first. The outcome
// is computed from yrkanden and decisions exactly as on the motion page, so an
// outcome filter reads batches in order until the page is full.
func (s *Service) List(ctx context.Context, f domain.ListFilter, after *domain.ListCursor, limit int) (*ListResult, error) {
	want := map[string]bool{}
	for _, o := range f.Outcomes {
		want[o] = true
	}
	res := &ListResult{Items: []ListItem{}}
	size := limit + 1
	if len(want) > 0 {
		size = listBatch
	}
	cur := after
	for batch := 0; batch < listMaxBatches; batch++ {
		rows, err := s.repo.ListPage(ctx, f, cur, size)
		if err != nil {
			return nil, err
		}
		for i, row := range rows {
			agg := row.Outcome()
			if len(want) > 0 && !want[agg.Word] {
				cur = &rows[i].Cursor
				continue
			}
			if len(res.Items) == limit {
				// A further match exists: continue after the last returned row.
				c := res.Items[len(res.Items)-1].Cursor
				res.Next = &c
				return res, nil
			}
			res.Items = append(res.Items, ListItem{Motion: row.Motion, Outcome: agg, Cursor: row.Cursor, HandledIn: row.HandledIn})
			cur = &rows[i].Cursor
		}
		if len(rows) < size {
			return res, nil
		}
	}
	// Scan budget spent with the page not yet full: hand back where we stopped.
	if cur != nil {
		c := *cur
		res.Next = &c
	}
	return res, nil
}

// BackfillWordingHashes fills the wording hash of yrkanden stored before the
// hash existed, batch rows at a time, and returns how many it wrote. It is
// resumable and idempotent: it only touches rows without a hash.
func (s *Service) BackfillWordingHashes(ctx context.Context, batch int) (int, error) {
	total := 0
	var cur ports.YrkandeCursor
	for {
		written, scanned, last, err := s.repo.BackfillWordingHashes(ctx, cur, batch)
		total += written
		if err != nil {
			return total, err
		}
		if scanned == 0 {
			return total, nil
		}
		cur = last
		if err := ctx.Err(); err != nil {
			return total, err
		}
	}
}

// Recurring returns the Återkommande krav a yrkande belongs to: every stored
// filing with identical normalised wording, earliest first, each with its
// parties and outcome. It returns nil when the wording was filed in only one
// riksmöte (or in none we hold besides this one). A hash that is missing on
// the yrkande is computed and stored here, so old rows heal when read.
func (s *Service) Recurring(ctx context.Context, motionDokID string, y domain.Yrkande) (*domain.Recurring, error) {
	hash := y.WordingHash
	if hash == "" {
		hash = domain.WordingHash(y.Wording)
		if hash == "" {
			return nil, nil
		}
		if err := s.repo.SetWordingHash(ctx, motionDokID, y.Number, hash); err != nil {
			return nil, err
		}
	}
	members, err := s.repo.RecurringMembers(ctx, hash)
	if err != nil {
		return nil, err
	}
	riksmoten := map[string]bool{}
	for _, m := range members {
		riksmoten[m.Riksmote] = true
	}
	if len(riksmoten) < 2 {
		return nil, nil
	}
	decisions := map[string][]domain.YrkandeDecision{}
	for i := range members {
		m := &members[i]
		m.Rule = domain.RuleIdenticalWording
		ds, ok := decisions[m.DokID]
		if !ok {
			if ds, err = s.repo.Decisions(ctx, m.DokID); err != nil {
				return nil, err
			}
			decisions[m.DokID] = ds
		}
		var mine []string
		for _, d := range ds {
			if d.YrkandeNumber == m.Number {
				mine = append(mine, d.Decision)
			}
		}
		m.Outcome = domain.YrkandeWord(m.ChamberDecision, mine)
	}
	return &domain.Recurring{Rule: domain.RuleIdenticalWording, Members: members}, nil
}

// BackfillRiksmoten lists, newest first, the riksmöten with voteringar or
// declared in a mandate period.
func (s *Service) BackfillRiksmoten(ctx context.Context) ([]string, error) {
	return s.repo.BackfillRiksmoten(ctx)
}

// BackfillProgress lists the operator backfill's progress per riksmöte and source.
func (s *Service) BackfillProgress(ctx context.Context) ([]domain.BackfillProgress, error) {
	return s.repo.BackfillProgress(ctx)
}

// SaveBackfillProgress records progress. A completion already recorded stays.
func (s *Service) SaveBackfillProgress(ctx context.Context, p domain.BackfillProgress) error {
	return s.repo.SaveBackfillProgress(ctx, p)
}

// MotionCount counts the motions held for a riksmöte.
func (s *Service) MotionCount(ctx context.Context, rm string) (int, error) {
	return s.repo.MotionCount(ctx, rm)
}

// BetankandeCounts counts the betänkanden a riksmöte's motions point at, and
// how many of them are held.
func (s *Service) BetankandeCounts(ctx context.Context, rm string) (expected, held int, err error) {
	return s.repo.BetankandeCounts(ctx, rm)
}

// DecisionVocabulary counts every distinct decision value stored.
func (s *Service) DecisionVocabulary(ctx context.Context) ([]domain.DecisionValue, error) {
	return s.repo.DecisionVocabulary(ctx)
}

// OriginCoverage counts voteringar and those with a decided yrkande linked.
func (s *Service) OriginCoverage(ctx context.Context) (voteringar, withOrigin int, err error) {
	return s.repo.OriginCoverage(ctx)
}

// RelinkAll recomputes the yrkande links of every betänkande held, from the
// stored förslagspunkter. It makes no request to Riksdagen. A votering's
// proposed-by is derived from these links when it is read, so after this the
// origin of every votering reflects every motion and betänkande now stored.
func (s *Service) RelinkAll(ctx context.Context) (int, error) {
	refs, err := s.repo.HeldBetankanden(ctx)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, ref := range refs {
		if err := ctx.Err(); err != nil {
			return n, err
		}
		if err := s.Link(ctx, ref.Riksmote, ref.Beteckning); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
