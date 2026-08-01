package riksdagen

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"riksdagskollen/internal/votes/ports"
)

// newTestClient points the client at a stub server.
func newTestClient(url string) *Client {
	return &Client{http: &http.Client{Timeout: 5 * time.Second}, baseURL: url}
}

// votePage renders n ballot rows in the shape /voteringlista returns.
func votePage(n int) string {
	out := `{"voteringlista":{"votering":[`
	for i := 0; i < n; i++ {
		if i > 0 {
			out += ","
		}
		out += fmt.Sprintf(`{"votering_id":"v%d","intressent_id":"p%d","namn":"Test Person",`+
			`"parti":"S","rost":"Ja","beteckning":"AU9","punkt":"1","rm":"2025/26",`+
			`"dok_id":"HC01AU9","systemdatum":"2026-06-17 10:00:00"}`, i, i)
	}
	return out + `]}}`
}

// Work is partitioned by betänkande, so the beteckning filter must reach the
// API — it is the only thing keeping a request under the 10 000-row cap.
func TestFetchVotes_SendsBeteckningAndDefaultsToTheCap(t *testing.T) {
	var q url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(votePage(1)))
	}))
	defer srv.Close()

	if _, err := newTestClient(srv.URL).FetchVotes(context.Background(),
		ports.FetchVotesFilter{Session: "2025/26", Beteckning: "AU9"}); err != nil {
		t.Fatalf("FetchVotes: %v", err)
	}
	if got := q.Get("bet"); got != "AU9" {
		t.Errorf("bet = %q, want \"AU9\"", got)
	}
	// Default size must be the cap: the endpoint ignores `p`, so a smaller
	// default would silently truncate a betänkande with no way to detect it.
	if got := q.Get("sz"); got != "10000" {
		t.Errorf("sz = %q, want \"10000\"", got)
	}
	if _, ok := q["p"]; ok {
		t.Error("client sent a `p` parameter, which /voteringlista ignores")
	}
}

// Every row must survive the round trip: with no pagination available, a
// dropped row is simply lost.
func TestFetchVotes_ReturnsEveryRow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(votePage(7)))
	}))
	defer srv.Close()

	vv, err := newTestClient(srv.URL).FetchVotes(context.Background(),
		ports.FetchVotesFilter{Session: "2025/26", Party: "S"})
	if err != nil {
		t.Fatalf("FetchVotes: %v", err)
	}
	if len(vv) != 7 {
		t.Errorf("got %d votes, want 7", len(vv))
	}
}

// SystemDatum drives the ingestion cursor. If it is dropped, the worker has no
// high-water mark and falls back to advancing blindly — the defect that left
// the site months behind.
func TestFetchVotes_PreservesSystemDatum(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(votePage(1)))
	}))
	defer srv.Close()

	vv, err := newTestClient(srv.URL).FetchVotes(context.Background(),
		ports.FetchVotesFilter{Session: "2025/26", Party: "S"})
	if err != nil {
		t.Fatalf("FetchVotes: %v", err)
	}
	if len(vv) != 1 {
		t.Fatalf("got %d votes, want 1", len(vv))
	}
	want := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)
	if !vv[0].SystemDatum.Equal(want) {
		t.Errorf("SystemDatum = %v, want %v", vv[0].SystemDatum, want)
	}
}
