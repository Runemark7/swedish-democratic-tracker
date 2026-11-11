package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"riksdagskollen/internal/ingestion/workers"
)

// Fixtures: fixtures/speech-backfill. The speeches of riksmöte 2017/18 (the
// 2018-03-22 debate on 2017/18:SkU11) and 2021/22 (the 2022-09-21 debate on
// MJU33 and one unrelated speech), with the person data of each speaker and
// the text of each anförande, recorded 2026-11-11 and reduced to those rows.
// No politician is seeded: the backfill creates every speaker.

type speechCoverage struct {
	Riksmote           string `json:"riksmote"`
	Expected           int    `json:"expected"`
	Stored             int    `json:"stored"`
	Linked             int    `json:"linked"`
	Unlinked           int    `json:"unlinked"`
	TextBytes          int64  `json:"textBytes"`
	RowsAdded          int    `json:"rowsAdded"`
	TableBytesAdded    int64  `json:"tableBytesAdded"`
	PoliticiansCreated int    `json:"politiciansCreated"`
	Complete           bool   `json:"complete"`
}

func getSpeechCoverage(t *testing.T, h *Harness) []speechCoverage {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/speeches/backfill", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/speeches/backfill = %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data []speechCoverage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body.Data
}

func startSpeechBackfill(t *testing.T) *Harness {
	t.Helper()
	return startChain(t, chainDir(t, "speech-backfill"), "{2017/18,2021/22}")
}

func countOf(t *testing.T, h *Harness, q string) int {
	t.Helper()
	var n int
	if err := h.DB.QueryRow(context.Background(), q).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestSpeechBackfillWalksNewestFirstAndCreatesEverySpeaker(t *testing.T) {
	h := startSpeechBackfill(t)
	ctx := context.Background()
	if n := countOf(t, h, `SELECT count(*) FROM politicians`); n != 0 {
		t.Fatalf("test expects no politicians, got %d", n)
	}
	rep := h.SpeechBackfill(ctx, workers.SpeechBackfillOptions{Riksmoten: []string{"2017/18", "2021/22"}})

	var order []string
	for _, r := range rep.Rows {
		order = append(order, r.Riksmote)
		if !r.Complete || r.Err != "" || len(r.Unresolved) != 0 {
			t.Errorf("%s not complete: %+v", r.Riksmote, r)
		}
	}
	if got := strings.Join(order, " "); got != "2021/22 2017/18" {
		t.Errorf("order = %v, want newest first", order)
	}
	cov := map[string]speechCoverage{}
	for _, c := range getSpeechCoverage(t, h) {
		cov[c.Riksmote] = c
	}
	if c := cov["2017/18"]; c.Stored != 15 || c.Expected != 15 || !c.Complete {
		t.Errorf("2017/18 = %+v", c)
	}
	if c := cov["2021/22"]; c.Stored != 15 || c.Expected != 15 || !c.Complete {
		t.Errorf("2021/22 = %+v", c)
	}
	// Speakers were created from the person data, former members as inactive.
	if n := countOf(t, h, `SELECT count(*) FROM politicians`); n != 14 {
		t.Errorf("politicians = %d, want 14", n)
	}
	var first, party string
	var active bool
	if err := h.DB.QueryRow(ctx, `SELECT first_name, party, is_active FROM politicians WHERE intressent_id = '0494054455310'`).Scan(&first, &party, &active); err != nil {
		t.Fatal(err)
	}
	if first != "David" || party != "SD" || active {
		t.Errorf("created politician = %s %s active=%v", first, party, active)
	}
	// Text was fetched for each speech.
	if n := countOf(t, h, `SELECT count(*) FROM speeches WHERE coalesce(speech_text,'') = ''`); n != 0 {
		t.Errorf("%d speeches without text", n)
	}
	if n := countOf(t, h, `SELECT count(*) FROM speeches WHERE riksmote IS NULL`); n != 0 {
		t.Errorf("%d speeches without riksmote", n)
	}
}

func TestSpeechBackfillLinksThroughTheRulesOf146(t *testing.T) {
	h := startSpeechBackfill(t)
	ctx := context.Background()
	h.SpeechBackfill(ctx, workers.SpeechBackfillOptions{Riksmoten: []string{"2017/18", "2021/22"}})

	cov := map[string]speechCoverage{}
	for _, c := range getSpeechCoverage(t, h) {
		cov[c.Riksmote] = c
	}
	// All 15 anföranden of 2017/18 carry rel_dok_id H501SkU11; of 2021/22's 15,
	// 14 carry H901MJU33 and one is unrelated.
	if c := cov["2017/18"]; c.Linked != 15 || c.Unlinked != 0 {
		t.Errorf("2017/18 = %+v", c)
	}
	if c := cov["2021/22"]; c.Linked != 14 || c.Unlinked != 1 {
		t.Errorf("2021/22 = %+v", c)
	}
	if n := countOf(t, h, `SELECT count(*) FROM speeches WHERE betankande_dok_id IS NOT NULL AND betankande_link_rule <> 'rel_dok_id'`); n != 0 {
		t.Errorf("%d speeches linked by another rule", n)
	}
	if n := countOf(t, h, `SELECT count(*) FROM speeches WHERE betankande_dok_id = 'H901MJU33'`); n != 14 {
		t.Errorf("speeches linked to H901MJU33 = %d, want 14", n)
	}
}

func TestSpeechBackfillResumesAndReRunsAreIdempotent(t *testing.T) {
	h := startSpeechBackfill(t)
	ctx := context.Background()
	h.SpeechBackfill(ctx, workers.SpeechBackfillOptions{Riksmoten: []string{"2021/22"}})
	textBefore := countOf(t, h, `SELECT coalesce(sum(octet_length(speech_text)),0)::int FROM speeches`)
	served := textRequests(h)

	// The resumed run skips what is complete and takes the rest.
	rep := h.SpeechBackfill(ctx, workers.SpeechBackfillOptions{Riksmoten: []string{"2021/22", "2017/18"}})
	if len(rep.Rows) != 2 || rep.Rows[0].Riksmote != "2021/22" || !rep.Rows[0].Skipped || rep.Rows[1].Skipped || !rep.Rows[1].Complete {
		t.Fatalf("rows = %+v", rep.Rows)
	}
	if got := textRequests(h) - served; got != 15 {
		t.Errorf("text requests = %d, want 15 (2017/18 only)", got)
	}

	// Redoing everything changes no row, and keeps the text the list does not carry.
	total := countOf(t, h, `SELECT count(*) FROM speeches`)
	people := countOf(t, h, `SELECT count(*) FROM politicians`)
	textAll := countOf(t, h, `SELECT coalesce(sum(octet_length(speech_text)),0)::int FROM speeches`)
	rep = h.SpeechBackfill(ctx, workers.SpeechBackfillOptions{Riksmoten: []string{"2021/22", "2017/18"}, Force: true})
	for _, r := range rep.Rows {
		if !r.Complete || r.RowsAdded != 0 || r.PoliticiansCreated != 0 {
			t.Errorf("redo %s = %+v", r.Riksmote, r)
		}
	}
	if countOf(t, h, `SELECT count(*) FROM speeches`) != total || countOf(t, h, `SELECT count(*) FROM politicians`) != people ||
		countOf(t, h, `SELECT coalesce(sum(octet_length(speech_text)),0)::int FROM speeches`) != textAll {
		t.Error("a redo changed stored rows")
	}
	if textBefore == 0 || textAll <= textBefore {
		t.Errorf("text bytes before %d, after %d", textBefore, textAll)
	}
}

func TestSpeechBackfillReportsStorageGrowthPerRiksmote(t *testing.T) {
	h := startSpeechBackfill(t)
	rep := h.SpeechBackfill(context.Background(), workers.SpeechBackfillOptions{Riksmoten: []string{"2017/18"}})
	if len(rep.Rows) != 1 {
		t.Fatalf("rows = %+v", rep.Rows)
	}
	r := rep.Rows[0]
	if r.RowsAdded != 15 || r.TextBytes <= 0 || r.TableBytesAdded <= 0 || r.IndexBytesAdded <= 0 {
		t.Errorf("growth = %+v", r)
	}
	var out strings.Builder
	rep.Write(&out)
	for _, want := range []string{"2017/18", "+15 rader", "tabell", "index", "snitt per riksmöte"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("report lacks %q:\n%s", want, out.String())
		}
	}
	c := getSpeechCoverage(t, h)
	if len(c) != 1 || c[0].RowsAdded != 15 || c[0].TableBytesAdded <= 0 || c[0].TextBytes <= 0 {
		t.Errorf("coverage = %+v", c)
	}
}

// textRequests counts the speech-text requests the replay server has answered.
func textRequests(h *Harness) int {
	n := 0
	for key, hits := range h.replay.Hits() {
		if strings.Contains(key, " /anforande/") {
			n += hits
		}
	}
	return n
}
