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

const voteringListPage = `{"dokumentlista":{"@traffar":"759","@sidor":"4","dokument":[
 {"dok_id":"HD19UbU31p2","beteckning":"UbU31","rm":"2025/26","organ":"UbU",
  "datum":"2026-06-17","systemdatum":"2026-06-17 14:12:03"},
 {"dok_id":"HD19TU20p9","beteckning":"TU20","rm":"2025/26","organ":"TU",
  "datum":"2026-06-17","systemdatum":"2026-06-17 13:58:01"}]}}`

// Unlike /voteringlista, /dokumentlista honours `p` and reports a total, so it
// is the only usable way to enumerate a riksmöte. Results must come back
// newest-first so the incremental worker can stop at the cursor.
func TestListVoteringar_ParsesRefsAndTotal(t *testing.T) {
	var q url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(voteringListPage))
	}))
	defer srv.Close()

	refs, total, err := newTestClient(srv.URL).ListVoteringar(context.Background(), "2025/26", 2, 200)
	if err != nil {
		t.Fatalf("ListVoteringar: %v", err)
	}
	if total != 759 {
		t.Errorf("total = %d, want 759", total)
	}
	if len(refs) != 2 {
		t.Fatalf("got %d refs, want 2", len(refs))
	}
	if refs[0].Beteckning != "UbU31" || refs[0].Organ != "UbU" {
		t.Errorf("ref[0] = %+v", refs[0])
	}
	want := time.Date(2026, 6, 17, 14, 12, 3, 0, time.UTC)
	if !refs[0].SystemDatum.Equal(want) {
		t.Errorf("SystemDatum = %v, want %v", refs[0].SystemDatum, want)
	}
	if got := q.Get("p"); got != "2" {
		t.Errorf("p = %q, want \"2\"", got)
	}
	if got := q.Get("sortorder"); got != "desc" {
		t.Errorf("sortorder = %q, want \"desc\"", got)
	}
	if got := q.Get("doktyp"); got != "votering" {
		t.Errorf("doktyp = %q, want \"votering\"", got)
	}
}

// The organ parameter must not be sent at all.
//
// Riksdagen ignores `organ` on /dokumentlista — organ=SoU returns betänkanden
// from AU, JuU and UbU — so no assertion about the organs that come *back* can
// detect whether we filtered. The only thing that can fail is what we send.
// A planned betänkande must not come back looking decided.
//
// /dokumentlista mixes betänkanden the chamber has decided with ones merely
// scheduled, and sorts both on `datum` — so the planned ones arrive at the top.
// On 2026-08-14 that was 15 of 40, five in the first eight rows, two with a
// justeringsdag four months out. A caller that cannot tell them apart publishes
// a future event as a past one, which is what the front page did.
//
// `datum` is deliberately not the signal: it is populated on both, and on every
// decided betänkande checked it fell one to six days before `beslutsdag`.
func TestFetchRecentBetankanden_DistinguishesPlannedFromDecided(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dokumentlista":{"dokument":[
			{"titel":"Decided","organ":"UbU","datum":"2026-08-12","beteckning":"UbU30",
			 "beslutsdag":"2026-08-13","status":"Webbpublicering"},
			{"titel":"Planned","organ":"AU","datum":"2026-08-12","beteckning":"AU9",
			 "status":"planerat"}]}}`))
	}))
	defer srv.Close()

	docs, err := newTestClient(srv.URL).FetchRecentBetankanden(context.Background(), 40)
	if err != nil {
		t.Fatalf("FetchRecentBetankanden: %v", err)
	}
	if len(docs) != 2 {
		t.Fatalf("got %d documents, want 2", len(docs))
	}

	decided, planned := docs[0], docs[1]
	if !decided.Decided() {
		t.Error("UbU30 has a beslutsdag but reports not decided")
	}
	if decided.DecisionDate != "2026-08-13" {
		t.Errorf("decision date = %q, want \"2026-08-13\"", decided.DecisionDate)
	}
	if decided.DecisionDate == decided.Date {
		t.Error("decision date equals datum; the two are distinct fields and must stay so")
	}
	if decided.Status != "Webbpublicering" {
		t.Errorf("status = %q, want \"Webbpublicering\"", decided.Status)
	}

	if planned.Decided() {
		t.Error("AU9 has no beslutsdag but reports decided — a planned betänkande " +
			"would render as a decision that has already happened")
	}
	if planned.DecisionDate != "" {
		t.Errorf("planned decision date = %q, want empty", planned.DecisionDate)
	}
	if planned.Status != "planerat" {
		t.Errorf("planned status = %q, want \"planerat\"", planned.Status)
	}
}

func TestFetchRecentBetankanden_SendsNoOrganFilter(t *testing.T) {
	var q url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dokumentlista":{"dokument":[
			{"titel":"T","organ":"SoU","datum":"2026-08-12","beteckning":"SoU1"}]}}`))
	}))
	defer srv.Close()

	docs, err := newTestClient(srv.URL).FetchRecentBetankanden(context.Background(), 40)
	if err != nil {
		t.Fatalf("FetchRecentBetankanden: %v", err)
	}
	if _, ok := q["organ"]; ok {
		t.Errorf("client sent organ=%q; a complete feed must send no organ filter", q.Get("organ"))
	}
	if got := q.Get("typ"); got != "bet" {
		t.Errorf("typ = %q, want \"bet\"", got)
	}
	if got := q.Get("sz"); got != "40" {
		t.Errorf("sz = %q, want \"40\"", got)
	}
	if len(docs) != 1 || docs[0].Organ != "SoU" || docs[0].Date != "2026-08-12" {
		t.Errorf("parsed %+v, want one SoU doc dated 2026-08-12", docs)
	}
}

// FetchDocuments must keep sending its organ parameter unchanged, even though
// Riksdagen ignores it: this test exists so the refactor that adds
// FetchRecentBetankanden cannot silently change the older call's behaviour.
func TestFetchDocuments_StillSendsOrgan(t *testing.T) {
	var q url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dokumentlista":{"dokument":[]}}`))
	}))
	defer srv.Close()

	if _, err := newTestClient(srv.URL).FetchDocuments(context.Background(),
		[]string{"SoU", "TU"}, 5); err != nil {
		t.Fatalf("FetchDocuments: %v", err)
	}
	if got := q.Get("organ"); got != "SoU,TU" {
		t.Errorf("organ = %q, want \"SoU,TU\"", got)
	}
}
