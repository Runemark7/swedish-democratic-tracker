package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"riksdagskollen/internal/ingestion/workers"
)

// Fixtures: fixtures/questions. Recorded 2026-11-11 from data.riksdagen.se:
// skriftlig fråga 2022/23:339 (HA11339) and its svar (HA12339), interpellation
// 2022/23:397 (HA10397), the listings that hold them, and motion 2017/18:3293.
// Where a listing was reduced the fixture's note says so.

const (
	nordinID     = "0132818093422" // Rickard Nordin (C)
	svantessonID = "0899112324519" // Finansminister Elisabeth Svantesson (M)
)

type person struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Party string `json:"party"`
}

type questionItem struct {
	DokID       string  `json:"dokId"`
	Kind        string  `json:"kind"`
	Riksmote    string  `json:"riksmote"`
	Beteckning  string  `json:"beteckning"`
	Title       string  `json:"title"`
	AnswerDate  *string `json:"answerDate"`
	AnswerDokID string  `json:"answerDokId"`
	Asker       person  `json:"asker"`
	Answerer    person  `json:"answerer"`
	Role        string  `json:"role"`
	Rule        string  `json:"rule"`
	SourceURL   string  `json:"sourceUrl"`
}

func startQuestions(t *testing.T, dir string) *Harness {
	t.Helper()
	h := Start(t, dir, WithNow(time.Date(2023, 7, 1, 12, 0, 0, 0, time.UTC)))
	ctx := context.Background()
	workers.FetchDelay = 0
	if _, err := h.DB.Exec(ctx, `DELETE FROM mandate_periods`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
		VALUES ('test', 'test', '2022-09-26', '2026-09-27', '{2022/23}')`); err != nil {
		t.Fatal(err)
	}
	// Motion 2017/18:3293 is held before the questions are read.
	h.Backfill(ctx, workers.BackfillOptions{Riksmoten: []string{"2017/18"}, Sources: []string{"motioner"}, SkipRelink: true})
	h.RunWorker(ctx, "fetch-questions")
	if h.Served("/dokumentstatus/HA11339.json") == 0 {
		t.Fatal("ingestion did not read the recording")
	}
	return h
}

func getJSON(t *testing.T, h *Harness, path string, out any) {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d: %s", path, rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
		t.Fatal(err)
	}
}

func memberQuestions(t *testing.T, h *Harness, id string) []questionItem {
	t.Helper()
	var b struct {
		Questions []questionItem `json:"questions"`
	}
	getJSON(t, h, "/api/politicians/"+id+"/questions", &b)
	return b.Questions
}

func find(items []questionItem, kind, bet string) *questionItem {
	for i := range items {
		if items[i].Kind == kind && items[i].Beteckning == bet {
			return &items[i]
		}
	}
	return nil
}

func TestFraga339IsPairedWithItsSvarAskedByNordinAndAnsweredBySvantesson(t *testing.T) {
	h := startQuestions(t, chainDir(t, "questions"))
	var d struct {
		questionItem
		Text string `json:"text"`
	}
	getJSON(t, h, "/api/questions/fraga/2022%2F23/339", &d)
	if d.AnswerDokID != "HA12339" || d.AnswerDate == nil || (*d.AnswerDate)[:10] != "2023-02-22" {
		t.Fatalf("svar = %q on %v, want HA12339 on 2023-02-22", d.AnswerDokID, d.AnswerDate)
	}
	if d.Asker.Name != "Rickard Nordin" || d.Asker.Party != "C" || d.Asker.ID != nordinID {
		t.Fatalf("asker = %+v", d.Asker)
	}
	if d.Answerer.Name != "Finansminister Elisabeth Svantesson" || d.Answerer.Party != "M" || d.Answerer.ID != svantessonID {
		t.Fatalf("answerer = %+v", d.Answerer)
	}
	if d.SourceURL == "" || d.Text == "" {
		t.Fatalf("no source link or text: %+v", d)
	}
}

func TestFraga339AppearsForNordinAndSvantessonButIsNotLinkedToMotion3293(t *testing.T) {
	h := startQuestions(t, chainDir(t, "questions"))
	if q := find(memberQuestions(t, h, nordinID), "fraga", "339"); q == nil || q.Role != "asker" {
		t.Fatalf("Nordin's list lacks 339 as asker: %+v", q)
	}
	if q := find(memberQuestions(t, h, svantessonID), "fraga", "339"); q == nil || q.Role != "answerer" {
		t.Fatalf("Svantesson's list lacks 339 as answerer: %+v", q)
	}
	var b struct {
		Questions []questionItem `json:"questions"`
	}
	getJSON(t, h, "/api/motions/2017%2F18/3293/questions", &b)
	if len(b.Questions) != 0 {
		t.Fatalf("339 cites motion 3293 only in prose, but motion 3293 lists %+v", b.Questions)
	}
	var d struct {
		Motions []any `json:"motions"`
	}
	getJSON(t, h, "/api/questions/fraga/2022%2F23/339", &d)
	if len(d.Motions) != 0 {
		t.Fatalf("339 has motion links: %+v", d.Motions)
	}
}

func TestInterpellationIsListedForAskerAndMinisterWithItsDebate(t *testing.T) {
	h := startQuestions(t, chainDir(t, "questions"))
	h.RunWorker(context.Background(), "fetch-politicians")
	h.RunWorker(context.Background(), "fetch-speeches")
	const erikssonID, tenjeID = "0296384583518", "0476336341120"
	if q := find(memberQuestions(t, h, erikssonID), "interpellation", "401"); q == nil || q.Role != "asker" || q.Asker.Party != "S" {
		t.Fatalf("Eriksson's list lacks ip 401 as asker: %+v", q)
	}
	q := find(memberQuestions(t, h, tenjeID), "interpellation", "401")
	if q == nil || q.Role != "answerer" || q.AnswerDate == nil || (*q.AnswerDate)[:10] != "2023-08-31" {
		t.Fatalf("Tenje's list lacks ip 401 as answerer on 2023-08-31: %+v", q)
	}
	var d struct {
		Debate []struct {
			Name string `json:"name"`
			Date string `json:"date"`
		} `json:"debate"`
	}
	getJSON(t, h, "/api/questions/interpellation/2022%2F23/401", &d)
	if len(d.Debate) != 7 {
		t.Fatalf("debate has %d anföranden, want the 3 + 4 recorded", len(d.Debate))
	}
	for _, s := range d.Debate {
		if s.Date[:10] != "2023-08-31" {
			t.Fatalf("anförande on %s", s.Date)
		}
	}
}

func TestIngestionIsIdempotentAndRecordsCoverage(t *testing.T) {
	h := startQuestions(t, chainDir(t, "questions"))
	before := h.Served("/dokumentstatus/HA11339.json")
	h.RunWorker(context.Background(), "fetch-questions")
	if after := h.Served("/dokumentstatus/HA11339.json"); after != before {
		t.Fatalf("unchanged question read again: %d -> %d requests", before, after)
	}
	if n := len(memberQuestions(t, h, nordinID)); n != 1 {
		t.Fatalf("Nordin has %d questions after a second run, want 1", n)
	}
	var cov struct {
		Coverage []struct {
			Riksmote string `json:"riksmote"`
			Kind     string `json:"kind"`
			Expected int    `json:"expected"`
			Held     int    `json:"held"`
		} `json:"coverage"`
	}
	getJSON(t, h, "/api/questions/coverage", &cov)
	got := map[string][2]int{}
	for _, c := range cov.Coverage {
		got[c.Riksmote+" "+c.Kind] = [2]int{c.Expected, c.Held}
	}
	if got["2022/23 fraga"] != [2]int{979, 1} || got["2022/23 interpellation"] != [2]int{432, 1} {
		t.Fatalf("coverage = %v, want fraga 979 listed / 1 held and interpellation 432 / 1", got)
	}
}

// citedFixtures copies fixtures/questions to a temp dir and edits ONE recorded
// document: fråga HA11339's html gets two added sentences. This is synthetic.
// No recorded fråga or interpellation in ~800 read from Riksdagen on
// 2026-11-11 cites a motion by beteckning, so the rule is exercised on the
// recorded document plus a sentence added here, not on a recording.
func citedFixtures(t *testing.T, added string) string {
	t.Helper()
	src := chainDir(t, "questions")
	dst := t.TempDir()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if e.Name() == "status-HA11339.json" {
			var fx map[string]any
			if err := json.Unmarshal(b, &fx); err != nil {
				t.Fatal(err)
			}
			doc := fx["body"].(map[string]any)["dokumentstatus"].(map[string]any)["dokument"].(map[string]any)
			doc["html"] = doc["html"].(string) + "<p>" + added + "</p>"
			fx["note"] = "SYNTHETIC: recorded HA11339 with a sentence added to its html by the test."
			if b, err = json.Marshal(fx); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(dst, e.Name()), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dst
}

func TestQuestionCitingAMotionBeteckningIsLinkedWithRuleCitedBeteckning(t *testing.T) {
	dir := citedFixtures(t, "Se kommittémotion 2017/18:3293 och regeringens proposition 2017/18:4202 samt motion 2017/18:9999.")
	h := startQuestions(t, dir)
	var b struct {
		Questions []questionItem `json:"questions"`
	}
	getJSON(t, h, "/api/motions/2017%2F18/3293/questions", &b)
	if len(b.Questions) != 1 || b.Questions[0].Beteckning != "339" || b.Questions[0].Rule != "cited-beteckning" {
		t.Fatalf("motion 3293 lists %+v, want fråga 339 with rule cited-beteckning", b.Questions)
	}
	// 4202 is cited as a proposition and 9999 is a motion we do not hold: neither
	// is shown, and a second ingestion keeps it that way.
	h.RunWorker(context.Background(), "fetch-questions")
	var d struct {
		Motions []struct {
			Beteckning string `json:"beteckning"`
			Rule       string `json:"rule"`
		} `json:"motions"`
	}
	getJSON(t, h, "/api/questions/fraga/2022%2F23/339", &d)
	if len(d.Motions) != 1 || d.Motions[0].Beteckning != "3293" || d.Motions[0].Rule != "cited-beteckning" {
		t.Fatalf("linked motions = %+v, want only 3293", d.Motions)
	}
	var n int
	if err := h.DB.QueryRow(context.Background(), `SELECT count(*) FROM question_motion_links`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("stored links = %d (%v), want 1", n, err)
	}
	getJSON(t, h, "/api/motions/2017%2F18/4202/questions", &b)
	if len(b.Questions) != 0 {
		t.Fatalf("motion 4202 is cited as a proposition but lists %+v", b.Questions)
	}
}
