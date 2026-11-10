package votes

import (
	"context"
	"errors"

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

// ProposedByFor derives "proposed by" for the given förslagspunkter from the
// Motioner link. See domain.ProposedBy.
func (s *Service) ProposedByFor(ctx context.Context, keys []ports.PointKey) (map[ports.PointKey]domain.ProposedBy, error) {
	return s.repo.ProposedByForPoints(ctx, keys)
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
	ss, err := s.repo.ListDistinctByCommitteePrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}
	return ss, s.attachProposedBy(ctx, ss)
}

// attachProposedBy fills each summary's ProposedBy from the Motioner link.
func (s *Service) attachProposedBy(ctx context.Context, ss []ports.VoteSummary) error {
	keys := make([]ports.PointKey, len(ss))
	for i, v := range ss {
		keys[i] = ports.PointKey{Riksmote: v.Riksmote, Beteckning: v.Beteckning, Punkt: v.Forslagspunkt}
	}
	pb, err := s.repo.ProposedByForPoints(ctx, keys)
	if err != nil {
		return err
	}
	for i := range ss {
		ss[i].ProposedBy = pb[keys[i]]
	}
	return nil
}

// ListByCommitteeWithPositions returns the committee's voteringar — the RÖSTAT
// row on the committee page — in one mandate period, with each party's
// dominant position, plus the total count for pagination.
//
// Refuses an unrecognised period with ErrPeriodNotFound rather than answering
// it with an empty page: 200 {"items":[],"total":0} for ?period=1994-1998 is
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
	res, err := s.repo.ListDistinctVotes(ctx, f)
	if err != nil {
		return res, err
	}
	return res, s.attachProposedBy(ctx, res.Votes)
}

func (s *Service) ListByBeteckning(ctx context.Context, beteckning, punkt, session string) ([]*domain.Vote, error) {
	return s.repo.ListByBeteckning(ctx, beteckning, punkt, session)
}

func (s *Service) GetBetankandeInfo(ctx context.Context, beteckning, session string) (*ports.BetankandeInfo, error) {
	return s.riksdagen.FetchBetankandeByBeteckning(ctx, beteckning, session)
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
