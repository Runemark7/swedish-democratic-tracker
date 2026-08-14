package votes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"riksdagskollen/internal/votes/domain"
	"riksdagskollen/internal/votes/ports"
)

// ErrPeriodNotFound mirrors committees.ErrPeriodNotFound's message exactly —
// same status, same words — for a mandate period the record does not hold.
// Kept as votes' own sentinel rather than importing the committees package:
// the votes repository already queries mandate_periods directly, so no
// cross-feature dependency is needed to answer this.
var ErrPeriodNotFound = errors.New("unknown mandate period")

type Service struct {
	repo      ports.VoteRepository
	riksdagen ports.RiksdagenVoteClient
}

func NewService(repo ports.VoteRepository, riksdagen ports.RiksdagenVoteClient) *Service {
	return &Service{repo: repo, riksdagen: riksdagen}
}

// ListVoteringar enumerates voteringar for a riksmöte, newest first, together
// with the total the API reports.
func (s *Service) ListVoteringar(ctx context.Context, rm string, page, size int) ([]ports.VoteringRef, int, error) {
	return s.riksdagen.ListVoteringar(ctx, rm, page, size)
}

// FetchAndStore fetches one betänkande's ballots and returns what was stored,
// so callers can track the newest record actually retrieved. SyncVotes discards
// that, which leaves a caller no way to know how far it really got.
func (s *Service) FetchAndStore(ctx context.Context, f ports.FetchVotesFilter) ([]*domain.Vote, error) {
	vv, err := s.riksdagen.FetchVotes(ctx, f)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpsertMany(ctx, vv); err != nil {
		return nil, err
	}
	return vv, nil
}

func (s *Service) SyncVotes(ctx context.Context, f ports.FetchVotesFilter) error {
	vv, err := s.riksdagen.FetchVotes(ctx, f)
	if err != nil {
		return err
	}
	return s.repo.UpsertMany(ctx, vv)
}

// EnrichOrigins resolves proposal origin for vote points that lack it, and
// returns how many points were enriched.
//
// Origin is a property of the vote point, not of an individual ballot: every
// member voting on AU9:1 shares one proposal. Resolving per ballot re-fetched
// the same document once per member — roughly 349 times per point — which made
// the full 2022-2026 record (~895 000 ballots) a multi-day job instead of a
// few minutes. One request per point, applied to every ballot on it.
func (s *Service) EnrichOrigins(ctx context.Context, batchSize int) (int, error) {
	points, err := s.repo.ListVotePointsWithoutOrigin(ctx, batchSize)
	if err != nil {
		return 0, err
	}

	enriched := 0
	for _, p := range points {
		origin, err := s.resolveOrigin(ctx, &domain.Vote{
			Beteckning:    p.Beteckning,
			Forslagspunkt: p.Forslagspunkt,
			Session:       p.Session,
			DokID:         p.DokID,
		})
		if err != nil {
			// Log and continue — don't abort the whole batch
			continue
		}
		if _, err := s.repo.UpdateProposalOriginForPoint(ctx, p.Beteckning, p.Forslagspunkt, origin); err != nil {
			return enriched, err
		}
		enriched++
		// Rate limit: be gentle on the Riksdagen API
		time.Sleep(100 * time.Millisecond)
	}
	return enriched, nil
}

// resolveOrigin traces beteckning → dokumentstatus → prop/mot → proposing party.
func (s *Service) resolveOrigin(ctx context.Context, v *domain.Vote) (domain.ProposalOrigin, error) {
	// Use the vote's dok_id directly (e.g. "HC01AU10") — this is the correct format
	// for the dokumentstatus API. Falls back to constructing from session+beteckning.
	dokID := v.DokID
	if dokID == "" {
		session := strings.ReplaceAll(v.Session, "/", "")
		dokID = fmt.Sprintf("%s:%s", session, v.Beteckning)
	}

	ds, err := s.riksdagen.FetchDocumentStatus(ctx, dokID)
	if err != nil || ds == nil {
		return domain.ProposalOrigin{}, err
	}

	origin := domain.ProposalOrigin{
		DocumentTitle: ds.Title,
		ProposalDokID: dokID,
	}

	for _, ref := range ds.References {
		switch ref.RefDokTyp {
		case "prop":
			origin.ProposedByParty = "Regeringen"
			origin.ProposalType = domain.ProposalProp
			origin.ProposalDokID = ref.RefDokID
			return origin, nil
		case "mot":
			// Fetch the motion's document to get the proposing party
			motDS, err := s.riksdagen.FetchDocumentStatus(ctx, ref.RefDokID)
			if err == nil && motDS != nil {
				for _, mr := range motDS.References {
					if mr.PartyBet != "" {
						origin.ProposedByParty = mr.PartyBet
						break
					}
				}
			}
			origin.ProposalType = domain.ProposalMot
			origin.ProposalDokID = ref.RefDokID
			return origin, nil
		}
	}
	// Fallback: the betänkande itself
	origin.ProposalType = domain.ProposalBet
	return origin, nil
}

func (s *Service) ListByPolitician(ctx context.Context, f ports.ListVotesFilter) (ports.ListVotesResult, error) {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.PageSize == 0 {
		f.PageSize = 50
	}
	return s.repo.ListByPolitician(ctx, f)
}

func (s *Service) ListDistinctByCommitteePrefix(ctx context.Context, prefix string) ([]ports.VoteSummary, error) {
	return s.repo.ListDistinctByCommitteePrefix(ctx, prefix)
}

// ListByCommitteeWithPositions returns the committee's voteringar — the RÖSTAT
// row on the committee page — in one mandate period, with each party's
// dominant position, plus the total count for pagination.
//
// Refuses an unrecognised period with ErrPeriodNotFound rather than answering
// it with an empty page: 200 {"items":[],"total":0} for ?period=2018-2022 is
// indistinguishable from a committee that decided nothing, which is a claim
// about the record this endpoint is in no position to make — the same
// reasoning committees.Service.Get already applies to GET /committees/{code}.
func (s *Service) ListByCommitteeWithPositions(ctx context.Context, periodCode, code string, limit, offset int) ([]ports.CommitteeVotering, int, error) {
	ok, err := s.repo.PeriodExists(ctx, periodCode)
	if err != nil {
		return nil, 0, err
	}
	if !ok {
		return nil, 0, ErrPeriodNotFound
	}
	return s.repo.ListByCommitteeWithPositions(ctx, periodCode, code, limit, offset)
}

func (s *Service) ListDistinctVotes(ctx context.Context, f ports.ListDistinctVotesFilter) (ports.ListDistinctVotesResult, error) {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.PageSize == 0 {
		f.PageSize = 50
	}
	return s.repo.ListDistinctVotes(ctx, f)
}

func (s *Service) ListByBeteckning(ctx context.Context, beteckning, punkt string) ([]*domain.Vote, error) {
	return s.repo.ListByBeteckning(ctx, beteckning, punkt)
}

func (s *Service) GetBetankandeInfo(ctx context.Context, beteckning string) (*ports.BetankandeInfo, error) {
	return s.riksdagen.FetchBetankandeByBeteckning(ctx, beteckning)
}

func (s *Service) GetDocumentStatus(ctx context.Context, dokID string) (*domain.DocumentStatus, error) {
	return s.riksdagen.FetchDocumentStatus(ctx, dokID)
}

// GetRecentBetankanden returns the most recently published betänkanden across
// every committee, newest first — the complete feed the front page's timeline
// is built on. It applies no committee selection.
func (s *Service) GetRecentBetankanden(ctx context.Context, count int) ([]ports.RiksdagDocument, error) {
	return s.riksdagen.FetchRecentBetankanden(ctx, count)
}
