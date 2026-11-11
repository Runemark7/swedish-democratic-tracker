package apitest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"riksdagskollen/internal/ingestion/workers"
)

// The backfill runs over the fixtures in fixtures/decision-chain:
// motions and betänkanden of 2017/18, 2018/19 and 2021/22.

// completeListings copies a fixture directory and sets @traffar of each motion
// listing to the number of rows kept in the recording. The recordings keep a
// few rows of a listing that reports thousands, so a riksmöte could never be
// "complete" against them. Only the count is changed; the rows are as recorded.
func completeListings(t *testing.T, dir string) string {
	t.Helper()
	root := t.TempDir()
	out := filepath.Join(root, filepath.Base(dir))
	// The fixtures point at ../html for the recorded betänkande pages.
	for _, d := range []string{dir, filepath.Join(filepath.Dir(dir), "html")} {
		dst := out
		if d != dir {
			dst = filepath.Join(root, "html")
		}
		if err := os.MkdirAll(dst, 0o700); err != nil {
			t.Fatal(err)
		}
		if d == dir {
			continue
		}
		pages, err := os.ReadDir(d)
		if err != nil {
			t.Fatal(err)
		}
		for _, pg := range pages {
			b, err := os.ReadFile(filepath.Join(d, pg.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dst, pg.Name()), b, 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(e.Name(), "mot-list-") {
			var f map[string]any
			if err := json.Unmarshal(raw, &f); err != nil {
				t.Fatal(err)
			}
			dl := f["body"].(map[string]any)["dokumentlista"].(map[string]any)
			rows := 0
			switch d := dl["dokument"].(type) {
			case []any:
				rows = len(d)
			case map[string]any:
				rows = 1
			}
			dl["@traffar"] = strconv.Itoa(rows)
			if raw, err = json.Marshal(f); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(out, e.Name()), raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return out
}

type progressRow struct {
	Riksmote string `json:"riksmote"`
	Source   string `json:"source"`
	Expected int    `json:"expected"`
	Held     int    `json:"held"`
	Complete bool   `json:"complete"`
}

func getProgress(t *testing.T, h *Harness) []progressRow {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/motions/backfill", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/motions/backfill = %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data []progressRow `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body.Data
}

func TestBackfillWalksNewestFirstAndRecordsCoveragePerRiksmoteAndSource(t *testing.T) {
	h := startChain(t, completeListings(t, chainDir(t, "decision-chain")), "{2017/18,2018/19,2021/22}")
	rep := h.Backfill(context.Background(), workers.BackfillOptions{})

	var order []string
	for _, r := range rep.Rows {
		order = append(order, r.Riksmote+"/"+r.Source)
		if !r.Complete {
			t.Errorf("%s %s not complete: %+v", r.Riksmote, r.Source, r)
		}
	}
	want := "2021/22/motioner 2021/22/betankanden 2018/19/motioner 2018/19/betankanden 2017/18/motioner 2017/18/betankanden"
	if strings.Join(order, " ") != want {
		t.Errorf("order = %v", order)
	}
	prog := getProgress(t, h)
	if len(prog) != 6 {
		t.Fatalf("progress rows = %+v", prog)
	}
	for _, p := range prog {
		if !p.Complete || p.Held == 0 {
			t.Errorf("progress = %+v", p)
		}
	}
}

func TestBackfillResumesAndReRunsAreIdempotent(t *testing.T) {
	h := startChain(t, completeListings(t, chainDir(t, "decision-chain")), "{2017/18,2018/19,2021/22}")
	ctx := context.Background()

	// An interrupted run: only the newest riksmöte got done.
	h.Backfill(ctx, workers.BackfillOptions{Riksmoten: []string{"2021/22"}})
	if len(getProgress(t, h)) != 2 {
		t.Fatal("first run should record 2021/22 only")
	}
	listed := h.Served("/dokumentlista/")

	rep := h.Backfill(ctx, workers.BackfillOptions{})
	skipped := 0
	for _, r := range rep.Rows {
		if r.Skipped {
			skipped++
			if r.Riksmote != "2021/22" {
				t.Errorf("skipped %s", r.Riksmote)
			}
		}
	}
	if skipped != 2 {
		t.Errorf("skipped %d rows, want the 2 of 2021/22", skipped)
	}
	if h.Served("/dokumentlista/") == listed {
		t.Error("the resumed run listed nothing for the older riksmöten")
	}

	// A third run changes nothing and reads no motion again.
	first, _ := json.Marshal(getChain(t, h, "2021%2F22", "3638"))
	reads := h.Served("/dokumentstatus/H9023638.json")
	rep = h.Backfill(ctx, workers.BackfillOptions{})
	for _, r := range rep.Rows {
		if !r.Skipped {
			t.Errorf("third run redid %+v", r)
		}
	}
	// Forcing redoes the work, but stored motions are not read again.
	h.Backfill(ctx, workers.BackfillOptions{Force: true})
	if got := h.Served("/dokumentstatus/H9023638.json"); got != reads {
		t.Errorf("forced re-run read the motion again: %d, was %d", got, reads)
	}
	again, _ := json.Marshal(getChain(t, h, "2021%2F22", "3638"))
	if string(first) != string(again) {
		t.Errorf("re-run changed the motion:\n%s\n%s", first, again)
	}
	if n := len(getProgress(t, h)); n != 6 {
		t.Errorf("progress rows = %d", n)
	}
}

func TestBackfillDoesNotCallARiksmoteCompleteWhileTheListingReportsMore(t *testing.T) {
	// The recordings hold a few rows of listings that report thousands.
	h := startChain(t, chainDir(t, "decision-chain"), "{2021/22}")
	rep := h.Backfill(context.Background(), workers.BackfillOptions{})
	for _, r := range rep.Rows {
		if r.Source == "motioner" && (r.Complete || r.Expected != 4777 || r.Held != 1) {
			t.Errorf("row = %+v", r)
		}
	}
}

func TestDailyChainTakesOnlyTheCurrentRiksmote(t *testing.T) {
	// The clock is in 2021/22; the older riksmöten are declared but are the
	// operator's backfill, not the daily chain's.
	h := startChain(t, chainDir(t, "decision-chain"), "{2017/18,2018/19,2021/22}")
	ctx := context.Background()
	h.RunWorker(ctx, "fetch-motions")
	h.RunWorker(ctx, "fetch-betankanden")
	if h.Served("/dokumentstatus/H9023638.json") == 0 {
		t.Fatal("the current riksmöte's motion was not read")
	}
	for _, id := range []string{"H5023293", "H5024202", "H6022900"} {
		if n := h.Served("/dokumentstatus/" + id + ".json"); n != 0 {
			t.Errorf("daily chain read older motion %s (%d)", id, n)
		}
	}
	if n := len(getProgress(t, h)); n != 0 {
		t.Errorf("the daily chain wrote backfill progress: %d rows", n)
	}
}

func TestBackfillReportListsDecisionVocabularyAndShowsUnmappedValuesVerbatim(t *testing.T) {
	h := startChain(t, completeListings(t, chainDir(t, "decision-chain")), "{2017/18,2018/19,2021/22}")
	ctx := context.Background()
	h.Backfill(ctx, workers.BackfillOptions{})

	// A decision word Outcome does not know, as Riksdagen might one day write it.
	if _, err := h.DB.Exec(ctx, `
		UPDATE motion_yrkanden SET kammaren = 'Återförvisas till utskottet'
		WHERE dok_id = 'H9023638'`); err != nil {
		t.Fatal(err)
	}

	rep := h.Backfill(ctx, workers.BackfillOptions{})
	var known, unknown *workers.VocabularyRow
	for i, v := range rep.Vocabulary {
		if v.Field != "kammaren" {
			continue
		}
		if v.Mapped && known == nil {
			known = &rep.Vocabulary[i]
		}
		if !v.Mapped {
			unknown = &rep.Vocabulary[i]
		}
	}
	if known == nil || known.Count == 0 {
		t.Errorf("no mapped kammaren value counted: %+v", rep.Vocabulary)
	}
	if unknown == nil || unknown.Value != "Återförvisas till utskottet" || unknown.Count != 1 {
		t.Fatalf("unmapped value not reported: %+v", rep.Vocabulary)
	}
	var out bytes.Buffer
	rep.Write(&out)
	text := out.String()
	for _, want := range []string{"Kammarens beslutsvärden", "EJ MAPPAT", `"Återförvisas till utskottet"`, "länkar omräknade", "voteringar med"} {
		if !strings.Contains(text, want) {
			t.Errorf("report lacks %q:\n%s", want, text)
		}
	}

	// The site shows the word exactly as Riksdagen wrote it.
	m := getChain(t, h, "2021%2F22", "3638")
	if m.Yrkanden[0].Outcome != "Återförvisas till utskottet" {
		t.Errorf("site outcome = %q, want Riksdagen's own word", m.Yrkanden[0].Outcome)
	}
}
