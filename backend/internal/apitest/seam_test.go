package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Recorded from data.riksdagen.se on 2026-11-01.
// Votering 8BADA090-C009-495A-98D2-F967E799BA61 is förslagspunkt 2 of
// betänkande 2025/26:UbU30. Kenneth G Forslund (S) voted Nej.
const (
	recordedVoteringID   = "8BADA090-C009-495A-98D2-F967E799BA61"
	recordedBeteckning   = "UbU30"
	recordedPunkt        = "2"
	recordedSession      = "2025/26"
	recordedPoliticianID = "0257612529618"
	recordedParty        = "S"
	recordedVote         = "Nej"
)

// TestRecordedVoteringIsReadableThroughTheVotesAPI is the first seam test:
// ingest one recorded votering and read that ballot back through the public
// votes API. The clock is pinned inside riksmöte 2025/26, the session the
// fixture was recorded from.
func TestRecordedVoteringIsReadableThroughTheVotesAPI(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Join(filepath.Dir(file), "fixtures", "one-votering")

	// 13 August 2026 is before September, so the votes worker asks for 2025/26.
	h := Start(t, dir, WithNow(time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)))
	ctx := context.Background()
	h.RunWorker(ctx, "fetch-politicians")
	h.RunWorker(ctx, "fetch-votes")
	// A live call would also contain this ballot. These counts prove the
	// ingestion read the fixture server, not data.riksdagen.se.
	for _, path := range []string{"/personlista/", "/dokumentlista/", "/voteringlista/"} {
		if h.Served(path) == 0 {
			t.Fatalf("ingestion made no recorded request for %s", path)
		}
	}

	q := url.Values{}
	q.Set("session", recordedSession)
	req := httptest.NewRequest(http.MethodGet, "/api/politicians/"+recordedPoliticianID+"/votes?"+q.Encode(), nil)
	rec := h.Do(req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET votes status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data []struct {
			VoteringID    string `json:"voteringId"`
			PoliticianID  string `json:"politicianId"`
			Party         string `json:"party"`
			VoteResult    string `json:"voteResult"`
			Beteckning    string `json:"beteckning"`
			Forslagspunkt string `json:"forslagspunkt"`
			Session       string `json:"session"`
		} `json:"data"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode votes: %v body=%s", err, rec.Body.String())
	}
	if body.Total < 1 {
		t.Fatalf("votes API returned no ballots for the recorded member; total=%d body=%s", body.Total, rec.Body.String())
	}
	for _, v := range body.Data {
		if v.VoteringID == recordedVoteringID &&
			v.PoliticianID == recordedPoliticianID &&
			v.Party == recordedParty &&
			v.VoteResult == recordedVote &&
			v.Beteckning == recordedBeteckning &&
			v.Forslagspunkt == recordedPunkt &&
			v.Session == recordedSession {
			return
		}
	}
	t.Fatalf("recorded votering %s was not in the votes API response: %+v", recordedVoteringID, body.Data)
}
