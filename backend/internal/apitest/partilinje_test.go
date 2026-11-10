package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

type detailBody struct {
	PartyBreakdown []struct {
		Party      string `json:"party"`
		Partilinje string `json:"partilinje"`
	} `json:"partyBreakdown"`
	Avvikare []struct {
		VoteringID   string `json:"voteringId"`
		PoliticianID string `json:"politicianId"`
		Party        string `json:"party"`
		VoteResult   string `json:"voteResult"`
		Partilinje   string `json:"partilinje"`
	} `json:"avvikare"`
}

func getDetail(t *testing.T, h *Harness) detailBody {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/votes/"+recordedBeteckning+"/"+recordedPunkt+"?session=2025/26", nil)
	rec := h.Do(req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET vote detail status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body detailBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}
	return body
}

// TestRecordedVoteringExposesPartilinjeAndAvvikare asserts the votes API's
// Partilinje per party and Avvikare per votering on the recorded UbU30
// förslagspunkt 2. In the recording every party is unanimous among its cast
// ballots (S and C each have one Frånvarande, which is excluded), so there is
// no Avvikare. The second half then flips one recorded S ballot to prove the
// API names that member and nobody else.
func TestRecordedVoteringExposesPartilinjeAndAvvikare(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(file), "fixtures", "one-votering")
	h := Start(t, dir, WithNow(time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)))
	ctx := context.Background()
	h.RunWorker(ctx, "fetch-politicians")
	h.RunWorker(ctx, "fetch-votes")

	body := getDetail(t, h)
	want := map[string]string{
		"C": "Ja", "KD": "Ja", "L": "Ja", "M": "Ja", "SD": "Ja",
		"S": "Nej", "MP": "Nej", "V": "Nej",
		"-": "", // no party affiliation: counted in tallies, never a Partilinje
	}
	if len(body.PartyBreakdown) != len(want) {
		t.Fatalf("partyBreakdown has %d parties, want %d: %+v", len(body.PartyBreakdown), len(want), body.PartyBreakdown)
	}
	for _, p := range body.PartyBreakdown {
		if w, ok := want[p.Party]; !ok || p.Partilinje != w {
			t.Errorf("party %s partilinje = %q, want %q", p.Party, p.Partilinje, w)
		}
	}
	if body.Avvikare == nil || len(body.Avvikare) != 0 {
		t.Fatalf("recorded votering has no Avvikare; got %+v", body.Avvikare)
	}

	// 0257612529618 (S) voted Nej in the recording. Make that ballot a Ja:
	// S's Partilinje stays Nej (105 of 106), so exactly this member deviates.
	if _, err := h.DB.Exec(ctx,
		`UPDATE ballots SET vote_result = 'Ja' WHERE politician_id = $1 AND votering_ref IN
		   (SELECT id FROM voteringar WHERE votering_id = $2)`,
		recordedPoliticianID, recordedVoteringID); err != nil {
		t.Fatalf("flip ballot: %v", err)
	}
	body = getDetail(t, h)
	if len(body.Avvikare) != 1 {
		t.Fatalf("want exactly one Avvikare, got %+v", body.Avvikare)
	}
	a := body.Avvikare[0]
	if a.PoliticianID != recordedPoliticianID || a.Party != "S" || a.VoteResult != "Ja" ||
		a.Partilinje != "Nej" || a.VoteringID != recordedVoteringID {
		t.Fatalf("unexpected Avvikare: %+v", a)
	}
}
