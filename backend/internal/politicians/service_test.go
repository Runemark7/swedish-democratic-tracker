package politicians_test

import (
	"context"
	"testing"

	"riksdagskollen/internal/politicians"
	"riksdagskollen/internal/politicians/domain"
	"riksdagskollen/internal/politicians/ports"
)

// fakeClient records the party filter it was called with and returns a fixed
// roster that includes a party-less member ("-", a politisk vilde).
type fakeClient struct {
	calls   []string // party args, in order
	members []*domain.Politician
}

func (f *fakeClient) FetchMembers(_ context.Context, party, _ string) ([]*domain.Politician, error) {
	f.calls = append(f.calls, party)
	return f.members, nil
}

// fakeRepo records everything passed to UpsertMany.
type fakeRepo struct {
	upserted []*domain.Politician
}

func (r *fakeRepo) UpsertMany(_ context.Context, pp []*domain.Politician) error {
	r.upserted = append(r.upserted, pp...)
	return nil
}
func (r *fakeRepo) GetByID(context.Context, string) (*domain.Politician, error) { return nil, nil }
func (r *fakeRepo) List(context.Context, ports.ListFilter) (ports.ListResult, error) {
	return ports.ListResult{}, nil
}

func TestSyncAll_FetchesEveryPartyIncludingVildar(t *testing.T) {
	client := &fakeClient{members: []*domain.Politician{
		{IntressentID: "1", Party: "S", IsActive: true},
		{IntressentID: "2", Party: "M", IsActive: true},
		{IntressentID: "3", Party: "-", IsActive: true}, // politisk vilde
	}}
	repo := &fakeRepo{}
	svc := politicians.NewService(repo, client)

	if err := svc.SyncAll(context.Background()); err != nil {
		t.Fatalf("SyncAll: %v", err)
	}

	// Must query with an empty party filter (no per-party loop), which is how
	// the riksdagen personlista endpoint returns all members — including the
	// party-less "-" members the old per-party sync missed.
	if len(client.calls) != 1 || client.calls[0] != "" {
		t.Fatalf("FetchMembers party args = %v, want exactly one call with \"\"", client.calls)
	}
	if len(repo.upserted) != 3 {
		t.Fatalf("upserted %d members, want 3 (incl the \"-\" vilde)", len(repo.upserted))
	}
	var hasVilde bool
	for _, p := range repo.upserted {
		if p.Party == "-" {
			hasVilde = true
		}
	}
	if !hasVilde {
		t.Fatal("the party-less (\"-\") member was not upserted")
	}
}
