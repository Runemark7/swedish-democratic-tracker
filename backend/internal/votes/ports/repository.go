package ports

import (
	"context"

	"riksdagskollen/internal/votes/domain"
)

type ListVotesFilter struct {
	PoliticianID string
	Session      string
	Topic        string
	Page         int
	PageSize     int
}

type ListVotesResult struct {
	Votes []*domain.Vote
	Total int
}

// VoteSummary is a lightweight projection for the decision feed and matching
// (distinct riksmöte/beteckning/punkt triples).
type VoteSummary struct {
	Riksmote      string `json:"riksmote"`
	Beteckning    string `json:"beteckning"`
	Forslagspunkt string `json:"forslagspunkt"`
	DocumentTitle string `json:"documentTitle"`
	// ProposedBy is derived from the Motioner link; see domain.ProposedBy.
	ProposedBy domain.ProposedBy `json:"proposedBy"`
}

// PointKey identifies one förslagspunkt: numbering restarts every riksmöte, so
// all three parts are the key.
type PointKey struct {
	Riksmote   string
	Beteckning string
	Punkt      string
}

type ListDistinctVotesFilter struct {
	Page     int
	PageSize int
}

type ListDistinctVotesResult struct {
	Votes []VoteSummary
	Total int
}

// PartyPosition is a party's Partilinje in one votering, as defined once in
// domain.PartilinjeOf. Stated as a fact, never scored against a promise. An
// empty Position means no member of the party cast a ballot. Party never
// carries the record's "-" placeholder: RÖSTAT is a per-party list, and "-" is
// not a party.
type PartyPosition struct {
	Party    string            `json:"party"`
	Position domain.Partilinje `json:"position,omitempty"`
}

// CommitteeVotering is one decided förslagspunkt with each party's position.
//
// Carries no date: votes hold only system_datum, which is when Riksdagen last
// touched the record, so presenting it as a decision date would repeat a bug
// the site already removed.
//
// VoteringID distinguishes the 18 förslagspunkter in 2022-2026 that were
// genuinely decided by two separate voteringar (same beteckning and
// förslagspunkt, different votering_id) — without it, two voteringar with
// contradictory party positions would be indistinguishable in the response.
type CommitteeVotering struct {
	VoteringID    string `json:"voteringId"`
	Beteckning    string `json:"beteckning"`
	Forslagspunkt string `json:"forslagspunkt"`
	DocumentTitle string `json:"documentTitle"`
	Riksmote      string `json:"riksmote"`
	// DecidedByMultipleVoteringar is true when the record holds more than one
	// votering for this (riksmöte, beteckning, förslagspunkt).
	//
	// It belongs to the response and not to the caller because the caller sees
	// one page at a time, and a pair can straddle a page boundary: NU7 punkt 2
	// in 2025/26 sits at indices 149 and 150, so a page-local computation marks
	// neither. Unmarked, those two rows read as four parties reversing their
	// vote on the same question with no explanation.
	DecidedByMultipleVoteringar bool            `json:"decidedByMultipleVoteringar"`
	PartyPositions              []PartyPosition `json:"partyPositions"`
}

type VoteRepository interface {
	ListByPolitician(ctx context.Context, f ListVotesFilter) (ListVotesResult, error)
	// ListByBeteckning returns ballots for one förslagspunkt. session, when
	// non-empty, limits them to that riksmöte — beteckning numbering restarts
	// each year, so an unscoped read mixes two betänkanden into one tally.
	ListByBeteckning(ctx context.Context, beteckning, punkt, session string) ([]*domain.Vote, error)
	ListDistinctVotes(ctx context.Context, f ListDistinctVotesFilter) (ListDistinctVotesResult, error)
	UpsertMany(ctx context.Context, vv []*domain.Vote) error
	// ProposedByForPoints derives "proposed by" for each förslagspunkt from
	// the stored yrkande → förslagspunkt links (motions, their parties and the
	// yrkanden decided), the propositions the punkt text cites and the
	// alternative put to the vote. A point with no links gets an empty value.
	ProposedByForPoints(ctx context.Context, keys []PointKey) (map[PointKey]domain.ProposedBy, error)
	ListDistinctByCommitteePrefix(ctx context.Context, prefix string) ([]VoteSummary, error)
	// PeriodExists reports whether the mandate period is one the record knows,
	// mirroring committees/ports.Repository.PeriodExists — an unrecognised
	// period must be refused, not answered with an empty page that reads as
	// "this committee decided nothing".
	PeriodExists(ctx context.Context, periodCode string) (bool, error)
	// ListByCommitteeWithPositions returns the committee's voteringar in one
	// mandate period, with each party's dominant position, plus the total
	// count for pagination.
	//
	// Must not filter on the deprecated origin_enriched column: it says nothing
	// about whether a votering is real.
	// periodCode is required, scoping the query the same way /committees/{code}
	// scopes by mandate_periods — the record already anticipates holding more
	// than one period.
	ListByCommitteeWithPositions(ctx context.Context, periodCode, code string, limit, offset int) ([]CommitteeVotering, int, error)
}
