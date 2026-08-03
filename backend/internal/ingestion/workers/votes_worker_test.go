package workers_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	ingPorts "riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/ingestion/workers"
	"riksdagskollen/internal/votes"
	votedomain "riksdagskollen/internal/votes/domain"
	voteports "riksdagskollen/internal/votes/ports"
)

type fakeVoteClient struct {
	mu       sync.Mutex
	refs     []voteports.VoteringRef
	total    int
	byBet    map[string][]*votedomain.Vote
	fetched  []string
	listErr  error
	fetchErr error
}

func (f *fakeVoteClient) ListVoteringar(_ context.Context, _ string, page, _ int) ([]voteports.VoteringRef, int, error) {
	if f.listErr != nil {
		return nil, 0, f.listErr
	}
	if page > 1 {
		return nil, f.total, nil
	}
	return f.refs, f.total, nil
}

func (f *fakeVoteClient) FetchVotes(_ context.Context, flt voteports.FetchVotesFilter) ([]*votedomain.Vote, error) {
	if f.fetchErr != nil {
		return nil, f.fetchErr
	}
	f.mu.Lock()
	f.fetched = append(f.fetched, flt.Beteckning)
	f.mu.Unlock()
	return f.byBet[flt.Beteckning], nil
}

func (f *fakeVoteClient) FetchDocumentStatus(context.Context, string) (*votedomain.DocumentStatus, error) {
	return nil, nil
}

func (f *fakeVoteClient) FetchDocuments(context.Context, []string, int) ([]voteports.RiksdagDocument, error) {
	return nil, nil
}

func (f *fakeVoteClient) FetchBetankandeByBeteckning(context.Context, string) (*voteports.BetankandeInfo, error) {
	return nil, nil
}

// fakeVoteRepo satisfies ports.VoteRepository. Only UpsertMany is exercised by
// the worker; the rest are read paths the worker never touches.
type fakeVoteRepo struct{ upserted int }

func (r *fakeVoteRepo) UpsertMany(_ context.Context, vv []*votedomain.Vote) error {
	r.upserted += len(vv)
	return nil
}

func (r *fakeVoteRepo) ListByPolitician(context.Context, voteports.ListVotesFilter) (voteports.ListVotesResult, error) {
	return voteports.ListVotesResult{}, nil
}

func (r *fakeVoteRepo) ListByBeteckning(context.Context, string, string) ([]*votedomain.Vote, error) {
	return nil, nil
}

func (r *fakeVoteRepo) ListDistinctVotes(context.Context, voteports.ListDistinctVotesFilter) (voteports.ListDistinctVotesResult, error) {
	return voteports.ListDistinctVotesResult{}, nil
}

func (r *fakeVoteRepo) UpdateProposalOrigin(context.Context, string, string, votedomain.ProposalOrigin) error {
	return nil
}

func (r *fakeVoteRepo) ListWithoutOrigin(context.Context, int) ([]*votedomain.Vote, error) {
	return nil, nil
}

func (r *fakeVoteRepo) ListDistinctByCommitteePrefix(context.Context, string) ([]voteports.VoteSummary, error) {
	return nil, nil
}

func (r *fakeVoteRepo) ListVotePointsWithoutOrigin(context.Context, int) ([]voteports.VotePoint, error) {
	return nil, nil
}

func (r *fakeVoteRepo) UpdateProposalOriginForPoint(context.Context, string, string, votedomain.ProposalOrigin) (int64, error) {
	return 0, nil
}

type fakeCursors struct{ cur *ingPorts.Cursor }

func (c *fakeCursors) Get(context.Context, string) (*ingPorts.Cursor, error) { return c.cur, nil }
func (c *fakeCursors) Upsert(_ context.Context, x ingPorts.Cursor) error     { c.cur = &x; return nil }

// TestMain removes the API throttle: these tests use fakes, so the sleep buys
// nothing and costs seconds per test.
func TestMain(m *testing.M) {
	workers.FetchDelay = 0
	os.Exit(m.Run())
}

func ref(bet string, sd time.Time) voteports.VoteringRef {
	return voteports.VoteringRef{Beteckning: bet, SystemDatum: sd}
}

func vote(bet string, sd time.Time) *votedomain.Vote {
	return &votedomain.Vote{Beteckning: bet, SystemDatum: sd}
}

// Work is partitioned by betänkande: each distinct beteckning is fetched once,
// with no party filter, because one request returns every party's ballots --
// including party-less members, which the old per-party loop missed entirely.
func TestVotesWorker_FetchesEachBeteckningOnce(t *testing.T) {
	newer := time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC)
	client := &fakeVoteClient{
		refs:  []voteports.VoteringRef{ref("UbU31", newer), ref("TU20", newer), ref("UbU31", newer)},
		total: 3,
		byBet: map[string][]*votedomain.Vote{
			"UbU31": {vote("UbU31", newer)},
			"TU20":  {vote("TU20", newer)},
		},
	}
	w := workers.NewVotesWorker(votes.NewService(&fakeVoteRepo{}, client), &fakeCursors{})

	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(client.fetched) != 2 {
		t.Errorf("fetched %v, want each beteckning exactly once", client.fetched)
	}
	for _, bet := range client.fetched {
		if bet == "" {
			t.Error("fetched with an empty beteckning — the partition key is required")
		}
	}
}

// A failed fetch must not move the cursor: that is how the site fell months
// behind while logging success every day.
func TestVotesWorker_CursorUnchangedOnFetchError(t *testing.T) {
	before := time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC)
	cur := &fakeCursors{cur: &ingPorts.Cursor{DataType: "votes", LastDate: &before}}
	client := &fakeVoteClient{
		refs:     []voteports.VoteringRef{ref("UbU31", time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC))},
		total:    1,
		fetchErr: errors.New("upstream down"),
	}
	w := workers.NewVotesWorker(votes.NewService(&fakeVoteRepo{}, client), cur)
	_ = w.Run(context.Background())

	if !cur.cur.LastDate.Equal(before) {
		t.Errorf("cursor moved to %v on error, want unchanged %v", cur.cur.LastDate, before)
	}
}

// An empty result is not evidence of freshness.
func TestVotesWorker_CursorUnchangedWhenNothingFetched(t *testing.T) {
	before := time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC)
	cur := &fakeCursors{cur: &ingPorts.Cursor{DataType: "votes", LastDate: &before}}
	w := workers.NewVotesWorker(
		votes.NewService(&fakeVoteRepo{}, &fakeVoteClient{}), cur)
	_ = w.Run(context.Background())

	if !cur.cur.LastDate.Equal(before) {
		t.Errorf("cursor moved to %v with no rows, want unchanged %v", cur.cur.LastDate, before)
	}
}

// A failed enumeration must not move the cursor either.
func TestVotesWorker_CursorUnchangedOnEnumerationError(t *testing.T) {
	before := time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC)
	cur := &fakeCursors{cur: &ingPorts.Cursor{DataType: "votes", LastDate: &before}}
	w := workers.NewVotesWorker(
		votes.NewService(&fakeVoteRepo{}, &fakeVoteClient{listErr: errors.New("down")}), cur)
	_ = w.Run(context.Background())

	if !cur.cur.LastDate.Equal(before) {
		t.Errorf("cursor moved to %v on enumeration error, want unchanged %v", cur.cur.LastDate, before)
	}
}

// The cursor advances only as far as data actually retrieved.
func TestVotesWorker_CursorAdvancesToNewestFetched(t *testing.T) {
	newest := time.Date(2026, 6, 17, 14, 12, 3, 0, time.UTC)
	client := &fakeVoteClient{
		refs:  []voteports.VoteringRef{ref("UbU31", newest)},
		total: 1,
		byBet: map[string][]*votedomain.Vote{"UbU31": {vote("UbU31", newest)}},
	}
	cur := &fakeCursors{}
	w := workers.NewVotesWorker(votes.NewService(&fakeVoteRepo{}, client), cur)
	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cur.cur == nil || cur.cur.LastDate == nil || !cur.cur.LastDate.Equal(newest) {
		t.Errorf("cursor = %v, want %v", cur.cur, newest)
	}
}

// Voteringar at or before the cursor are skipped, so an incremental run does
// not refetch the whole riksmöte.
func TestVotesWorker_StopsAtTheCursor(t *testing.T) {
	before := time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC)
	older := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC)
	client := &fakeVoteClient{
		refs:  []voteports.VoteringRef{ref("NEW1", newer), ref("OLD1", older)},
		total: 2,
		byBet: map[string][]*votedomain.Vote{"NEW1": {vote("NEW1", newer)}},
	}
	cur := &fakeCursors{cur: &ingPorts.Cursor{DataType: "votes", LastDate: &before}}
	w := workers.NewVotesWorker(votes.NewService(&fakeVoteRepo{}, client), cur)
	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, bet := range client.fetched {
		if bet == "OLD1" {
			t.Error("fetched a beteckning older than the cursor")
		}
	}
}

func TestCurrentRiksmote(t *testing.T) {
	cases := map[string]string{
		"2026-08-01": "2025/26", // before the new riksmöte opens in September
		"2026-09-20": "2026/27",
		"2026-01-15": "2025/26",
		"2022-09-15": "2022/23",
	}
	for in, want := range cases {
		d, _ := time.Parse("2006-01-02", in)
		if got := workers.CurrentRiksmote(d); got != want {
			t.Errorf("CurrentRiksmote(%s) = %q, want %q", in, got, want)
		}
	}
}
