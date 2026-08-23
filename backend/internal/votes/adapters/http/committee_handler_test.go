package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/votes"
	votehttp "riksdagskollen/internal/votes/adapters/http"
	"riksdagskollen/internal/votes/domain"
	"riksdagskollen/internal/votes/ports"
)

// fakeCommitteeRepo satisfies ports.VoteRepository. Only PeriodExists and
// ListByCommitteeWithPositions are exercised by the committee-votes handler
// under test; every other method is a no-op stub, following the same
// faking style already used for this interface in
// internal/votes/enrich_test.go (enrichRepo) and
// internal/ingestion/workers/votes_worker_test.go (fakeVoteRepo) — no
// mocking library.
type fakeCommitteeRepo struct {
	knownPeriods map[string]bool
	page         []ports.CommitteeVotering
	total        int

	// Recorded call arguments, so a test can assert what the handler actually
	// passed down the stack, not only what came back.
	calls     int
	gotPeriod string
	gotCode   string
	gotLimit  int
	gotOffset int
}

func (r *fakeCommitteeRepo) PeriodExists(_ context.Context, period string) (bool, error) {
	return r.knownPeriods[period], nil
}

func (r *fakeCommitteeRepo) ListByCommitteeWithPositions(_ context.Context, period, code string, limit, offset int) ([]ports.CommitteeVotering, int, error) {
	r.calls++
	r.gotPeriod, r.gotCode, r.gotLimit, r.gotOffset = period, code, limit, offset
	return r.page, r.total, nil
}

func (r *fakeCommitteeRepo) ListByPolitician(context.Context, ports.ListVotesFilter) (ports.ListVotesResult, error) {
	return ports.ListVotesResult{}, nil
}
func (r *fakeCommitteeRepo) ListByBeteckning(context.Context, string, string) ([]*domain.Vote, error) {
	return nil, nil
}
func (r *fakeCommitteeRepo) ListDistinctVotes(context.Context, ports.ListDistinctVotesFilter) (ports.ListDistinctVotesResult, error) {
	return ports.ListDistinctVotesResult{}, nil
}
func (r *fakeCommitteeRepo) UpsertMany(context.Context, []*domain.Vote) error { return nil }
func (r *fakeCommitteeRepo) ListVotePointsWithoutOrigin(context.Context, int) ([]ports.VotePoint, error) {
	return nil, nil
}
func (r *fakeCommitteeRepo) UpdateProposalOriginForPoint(context.Context, string, string, domain.ProposalOrigin) (int64, error) {
	return 0, nil
}
func (r *fakeCommitteeRepo) ListDistinctByCommitteePrefix(context.Context, string) ([]ports.VoteSummary, error) {
	return nil, nil
}

// newTestRouter wires the real Handler and Service on top of the fake
// repository — exactly the production wiring, minus Postgres — so these
// tests exercise the actual routing, canonicalisation and error-mapping
// code in handler.go, not a re-implementation of it.
func newTestRouter(repo ports.VoteRepository) *chi.Mux {
	svc := votes.NewService(repo, nil)
	h := votehttp.NewHandler(svc)
	r := chi.NewRouter()
	h.Routes(r)
	return r
}

func TestListByCommittee_MissingPeriod(t *testing.T) {
	repo := &fakeCommitteeRepo{knownPeriods: map[string]bool{"2022-2026": true}}
	r := newTestRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/committees/AU/votes", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (body: %s)", rec.Code, rec.Body.String())
	}
	if repo.calls != 0 {
		t.Errorf("repository was queried despite a missing period — the 400 must be returned "+
			"before ListByCommitteeWithPositions is ever called; calls = %d", repo.calls)
	}
}

func TestListByCommittee_UnknownPeriod(t *testing.T) {
	repo := &fakeCommitteeRepo{knownPeriods: map[string]bool{"2022-2026": true}}
	r := newTestRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/committees/AU/votes?period=2018-2022", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (body: %s)", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err == nil {
		if body["error"] != "unknown mandate period" {
			t.Errorf(`error = %q, want "unknown mandate period" — same message as GET /committees/{code}`, body["error"])
		}
	}
}

// A raw path segment like "sou" must reach the repository as "SoU": the
// committee's own precedent (GET /committees/{code}/goals) canonicalises
// before querying, and this handler must not silently regress to a
// case-sensitive lookup.
func TestListByCommittee_CanonicalisesCode(t *testing.T) {
	repo := &fakeCommitteeRepo{knownPeriods: map[string]bool{"2022-2026": true}}
	r := newTestRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/committees/sou/votes?period=2022-2026", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	if repo.gotCode != "SoU" {
		t.Errorf("repository received code %q, want canonicalised \"SoU\"", repo.gotCode)
	}
}

func TestListByCommittee_ReturnsPageShape(t *testing.T) {
	want := []ports.CommitteeVotering{
		{
			VoteringID:    "V1",
			Beteckning:    "AU5",
			Forslagspunkt: "1",
			DocumentTitle: "Test document",
			Riksmote:      "2022/23",
			PartyPositions: []ports.PartyPosition{
				{Party: "C", Position: "Ja"},
			},
		},
	}
	repo := &fakeCommitteeRepo{knownPeriods: map[string]bool{"2022-2026": true}, page: want, total: 42}
	r := newTestRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/committees/AU/votes?period=2022-2026&limit=10&offset=5", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	if repo.gotPeriod != "2022-2026" || repo.gotLimit != 10 || repo.gotOffset != 5 {
		t.Errorf("repository received period=%q limit=%d offset=%d, want 2022-2026/10/5",
			repo.gotPeriod, repo.gotLimit, repo.gotOffset)
	}

	var body struct {
		Items []ports.CommitteeVotering `json:"items"`
		Total int                       `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Total != 42 {
		t.Errorf("total = %d, want 42", body.Total)
	}
	if len(body.Items) != 1 || body.Items[0].VoteringID != "V1" {
		t.Errorf("items = %+v, want one item with voteringId \"V1\"", body.Items)
	}
}
