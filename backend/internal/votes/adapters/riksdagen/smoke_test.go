package riksdagen

import (
	"context"
	"testing"
	"time"

	"riksdagskollen/internal/votes/ports"
)

// TestLiveSmoke exercises the real Riksdagen API. The whole ingestion strategy
// rests on two measured facts — /dokumentlista paginates and /voteringlista
// does not — so those are asserted here rather than assumed. Skipped under
// -short.
func TestLiveSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	c := NewClient()

	// 1. Enumeration must paginate and report a total.
	p1, total, err := c.ListVoteringar(ctx, "2025/26", 1, 50)
	if err != nil {
		t.Fatalf("ListVoteringar p1: %v", err)
	}
	if total < 100 {
		t.Errorf("total = %d, want a full riksmöte (hundreds)", total)
	}
	if len(p1) != 50 {
		t.Fatalf("p1 returned %d refs, want 50", len(p1))
	}

	p2, _, err := c.ListVoteringar(ctx, "2025/26", 2, 50)
	if err != nil {
		t.Fatalf("ListVoteringar p2: %v", err)
	}
	first := map[string]bool{}
	for _, r := range p1 {
		first[r.DokID] = true
	}
	overlap := 0
	for _, r := range p2 {
		if first[r.DokID] {
			overlap++
		}
	}
	if overlap > 0 {
		t.Errorf("page 2 overlaps page 1 by %d — enumeration is not paginating", overlap)
	}

	// Newest-first ordering is what lets the worker stop at the cursor.
	if len(p1) > 1 && p1[0].SystemDatum.Before(p1[len(p1)-1].SystemDatum) {
		t.Error("enumeration is not newest-first")
	}

	// 2. One betänkande must return every party and stay under the row cap.
	vv, err := c.FetchVotes(ctx, ports.FetchVotesFilter{Session: "2025/26", Beteckning: p1[0].Beteckning})
	if err != nil {
		t.Fatalf("FetchVotes %s: %v", p1[0].Beteckning, err)
	}
	if len(vv) == 0 {
		t.Fatalf("no ballots for %s", p1[0].Beteckning)
	}
	if len(vv) >= 10000 {
		t.Errorf("%s returned %d rows — at the cap, so it may be truncated",
			p1[0].Beteckning, len(vv))
	}
	parties := map[string]bool{}
	for _, v := range vv {
		parties[v.Party] = true
	}
	if len(parties) < 8 {
		t.Errorf("%s covered %d parties, want all 8+", p1[0].Beteckning, len(parties))
	}
	if vv[0].SystemDatum.IsZero() {
		t.Error("SystemDatum not populated — the ingestion cursor depends on it")
	}
	t.Logf("total=%d bet=%s ballots=%d parties=%d",
		total, p1[0].Beteckning, len(vv), len(parties))
}
