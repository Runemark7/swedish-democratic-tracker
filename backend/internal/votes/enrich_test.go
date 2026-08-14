package votes_test

import (
	"context"
	"testing"

	"riksdagskollen/internal/votes"
	"riksdagskollen/internal/votes/domain"
	"riksdagskollen/internal/votes/ports"
)

type enrichRepo struct {
	points  []ports.VotePoint
	updates []string // "bet:punkt" per update
}

func (r *enrichRepo) ListVotePointsWithoutOrigin(context.Context, int) ([]ports.VotePoint, error) {
	return r.points, nil
}

func (r *enrichRepo) UpdateProposalOriginForPoint(_ context.Context, bet, punkt string, _ domain.ProposalOrigin) (int64, error) {
	r.updates = append(r.updates, bet+":"+punkt)
	return 349, nil // one statement covers every ballot on the point
}

func (r *enrichRepo) ListWithoutOrigin(context.Context, int) ([]*domain.Vote, error) { return nil, nil }
func (r *enrichRepo) UpdateProposalOrigin(context.Context, string, string, domain.ProposalOrigin) error {
	return nil
}
func (r *enrichRepo) UpsertMany(context.Context, []*domain.Vote) error { return nil }
func (r *enrichRepo) ListByPolitician(context.Context, ports.ListVotesFilter) (ports.ListVotesResult, error) {
	return ports.ListVotesResult{}, nil
}
func (r *enrichRepo) ListByBeteckning(context.Context, string, string) ([]*domain.Vote, error) {
	return nil, nil
}
func (r *enrichRepo) ListDistinctVotes(context.Context, ports.ListDistinctVotesFilter) (ports.ListDistinctVotesResult, error) {
	return ports.ListDistinctVotesResult{}, nil
}
func (r *enrichRepo) ListDistinctByCommitteePrefix(context.Context, string) ([]ports.VoteSummary, error) {
	return nil, nil
}
func (r *enrichRepo) ListByCommitteeWithPositions(context.Context, string, string, int, int) ([]ports.CommitteeVotering, int, error) {
	return nil, 0, nil
}
func (r *enrichRepo) PeriodExists(context.Context, string) (bool, error) {
	return false, nil
}

type enrichClient struct{ calls int }

func (c *enrichClient) FetchDocumentStatus(context.Context, string) (*domain.DocumentStatus, error) {
	c.calls++
	return &domain.DocumentStatus{Title: "T"}, nil
}
func (c *enrichClient) FetchVotes(context.Context, ports.FetchVotesFilter) ([]*domain.Vote, error) {
	return nil, nil
}
func (c *enrichClient) ListVoteringar(context.Context, string, int, int) ([]ports.VoteringRef, int, error) {
	return nil, 0, nil
}
func (c *enrichClient) FetchDocuments(context.Context, []string, int) ([]ports.RiksdagDocument, error) {
	return nil, nil
}
func (c *enrichClient) FetchRecentBetankanden(context.Context, int) ([]ports.RiksdagDocument, error) {
	return nil, nil
}
func (c *enrichClient) FetchBetankandeByBeteckning(context.Context, string) (*ports.BetankandeInfo, error) {
	return nil, nil
}

// Origin belongs to the vote point, so it must be resolved once per point and
// applied to every ballot. Resolving per ballot re-fetched the same document
// ~349 times, which turned the full record into a multi-day job.
func TestEnrichOrigins_OneRequestPerVotePoint(t *testing.T) {
	repo := &enrichRepo{points: []ports.VotePoint{
		{Beteckning: "AU9", Forslagspunkt: "1", Session: "2025/26", DokID: "HC01AU9"},
		{Beteckning: "AU9", Forslagspunkt: "2", Session: "2025/26", DokID: "HC01AU9"},
		{Beteckning: "TU20", Forslagspunkt: "9", Session: "2025/26", DokID: "HC01TU20"},
	}}
	client := &enrichClient{}

	n, err := votes.NewService(repo, client).EnrichOrigins(context.Background(), 100)
	if err != nil {
		t.Fatalf("EnrichOrigins: %v", err)
	}
	if n != 3 {
		t.Errorf("enriched %d points, want 3", n)
	}
	if client.calls != 3 {
		t.Errorf("made %d API calls, want 3 — one per vote point", client.calls)
	}
	if len(repo.updates) != 3 {
		t.Errorf("issued %d updates, want 3", len(repo.updates))
	}
}
