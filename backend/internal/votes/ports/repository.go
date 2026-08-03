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

// VoteSummary is a lightweight projection for matching (distinct beteckning/punkt pairs).
type VoteSummary struct {
	Beteckning      string `json:"beteckning"`
	Forslagspunkt   string `json:"forslagspunkt"`
	DocumentTitle   string `json:"documentTitle"`
	ProposedByParty string `json:"proposedByParty,omitempty"`
	ProposalType    string `json:"proposalType,omitempty"`
}

// VotePoint identifies one decided förslagspunkt. Proposal origin is a
// property of the vote point, not of each member's ballot: every MP voting on
// AU9:1 shares one origin, so it must be resolved once and applied to all.
type VotePoint struct {
	Beteckning    string
	Forslagspunkt string
	Session       string
	DokID         string
}

type ListDistinctVotesFilter struct {
	Page     int
	PageSize int
}

type ListDistinctVotesResult struct {
	Votes []VoteSummary
	Total int
}

type VoteRepository interface {
	ListByPolitician(ctx context.Context, f ListVotesFilter) (ListVotesResult, error)
	ListByBeteckning(ctx context.Context, beteckning, punkt string) ([]*domain.Vote, error)
	ListDistinctVotes(ctx context.Context, f ListDistinctVotesFilter) (ListDistinctVotesResult, error)
	UpsertMany(ctx context.Context, vv []*domain.Vote) error
	UpdateProposalOrigin(ctx context.Context, voteringID, politicianID string, o domain.ProposalOrigin) error
	ListWithoutOrigin(ctx context.Context, limit int) ([]*domain.Vote, error)
	// ListVotePointsWithoutOrigin returns distinct vote points still needing
	// origin resolution — one row per point, not per ballot.
	ListVotePointsWithoutOrigin(ctx context.Context, limit int) ([]VotePoint, error)
	// UpdateProposalOriginForPoint applies one resolved origin to every ballot
	// on that vote point in a single statement.
	UpdateProposalOriginForPoint(ctx context.Context, beteckning, forslagspunkt string, o domain.ProposalOrigin) (int64, error)
	ListDistinctByCommitteePrefix(ctx context.Context, prefix string) ([]VoteSummary, error)
}
