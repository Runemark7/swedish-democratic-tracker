package apitest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"riksdagskollen/internal/ingestion/workers"
)

// Seam tests for curated motion links through the Veckobatch (issue #163).
// The data is the recorded fixtures of fixtures/questions (fråga 2022/23:339,
// motion 2017/18:3293) and fixtures/recurring-krav (M's virtual-goods demand
// and Rickard Nordin's C motions). The suggestions, the batch and the
// Redaktör's answers are explicit request bodies; nothing answers for anyone.

const curProv = `"chatRef":"chat-163#1","onBehalfOf":"redaktor"`

func startCuratedQuestions(t *testing.T) *Harness {
	t.Helper()
	h := StartWithWriteToken(t, chainDir(t, "questions"), testWriteToken, WithNow(time.Date(2023, 7, 1, 12, 0, 0, 0, time.UTC)))
	ctx := context.Background()
	workers.FetchDelay = 0
	if _, err := h.DB.Exec(ctx, `DELETE FROM mandate_periods`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
		VALUES ('test', 'test', '2022-09-26', '2026-09-27', '{2022/23}')`); err != nil {
		t.Fatal(err)
	}
	h.Backfill(ctx, workers.BackfillOptions{Riksmoten: []string{"2017/18"}, Sources: []string{"motioner"}, SkipRelink: true})
	h.RunWorker(ctx, "fetch-questions")
	return h
}

func startCuratedKrav(t *testing.T) *Harness {
	t.Helper()
	h := StartWithWriteToken(t, chainDir(t, "recurring-krav"), testWriteToken, WithNow(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)))
	ctx := context.Background()
	if _, err := h.DB.Exec(ctx, `DELETE FROM mandate_periods`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
		VALUES ('test', 'test', '2014-09-14', '2026-09-13', '{2017/18,2018/19,2019/20,2021/22,2022/23,2023/24,2024/25}')`); err != nil {
		t.Fatal(err)
	}
	h.Backfill(ctx, workers.BackfillOptions{Sources: []string{"motioner"}})
	return h
}

func submitMotionLinks(t *testing.T, h *Harness, key, suggestions string, want int) {
	t.Helper()
	rec := writeReq(h, "POST", "/api/write/motion-kopplingar", testWriteToken, key, `{`+curProv+`,"suggestions":[`+suggestions+`]}`)
	if rec.Code != want {
		t.Fatalf("submit: %d %s (want %d)", rec.Code, rec.Body, want)
	}
}

type curBatch struct {
	ID    int64 `json:"id"`
	Items []struct {
		Nr         int    `json:"nr"`
		Kind       string `json:"kind"`
		State      string `json:"state"`
		RolledOver bool   `json:"rolledOver"`
		Motion     *struct {
			Rule     string `json:"rule"`
			Question *struct {
				Beteckning string `json:"beteckning"`
				Asker      string `json:"asker"`
			} `json:"question"`
			Motion struct {
				Beteckning string `json:"beteckning"`
				Wording    string `json:"wording"`
			} `json:"motion"`
			Other *struct {
				Beteckning string `json:"beteckning"`
			} `json:"other"`
		} `json:"motion"`
	} `json:"items"`
}

func curCreateBatch(t *testing.T, h *Harness, key string) curBatch {
	t.Helper()
	rec := writeReq(h, "POST", "/api/write/veckobatch", testWriteToken, key, `{`+curProv+`}`)
	if rec.Code != 200 {
		t.Fatalf("create batch: %d %s", rec.Code, rec.Body)
	}
	return decode[curBatch](t, rec)
}

func curAnswer(t *testing.T, h *Harness, id int64, key, text string, want int) answerView {
	t.Helper()
	rec := writeReq(h, "POST", fmt.Sprintf("/api/write/veckobatch/%d/answers", id), testWriteToken, key,
		fmt.Sprintf(`{"text":%q,%s}`, text, curProv))
	if rec.Code != want {
		t.Fatalf("answer %q: %d %s (want %d)", text, rec.Code, rec.Body, want)
	}
	if want != 200 {
		return answerView{}
	}
	return decode[answerView](t, rec)
}

func curUndo(t *testing.T, h *Harness, entry int64, key string) {
	t.Helper()
	rec := writeReq(h, "POST", fmt.Sprintf("/api/write/aendringslogg/%d/undo", entry), testWriteToken, key, `{`+curProv+`}`)
	if rec.Code != 200 {
		t.Fatalf("undo %d: %d %s", entry, rec.Code, rec.Body)
	}
}

const fraga339 = `"kind":"fraga","question":{"kind":"fraga","riksmote":"2022/23","beteckning":"339"},"motion":{"riksmote":"2017/18","beteckning":"3293"}`

func motion3293Questions(t *testing.T, h *Harness) []questionItem {
	t.Helper()
	var b struct {
		Questions []questionItem `json:"questions"`
	}
	getJSON(t, h, "/api/motions/2017%2F18/3293/questions", &b)
	return b.Questions
}

func assertNoCuratedInPublicLog(t *testing.T, h *Harness) {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/aendringslogg?limit=200", nil))
	if rec.Code != 200 {
		t.Fatalf("log: %d", rec.Code)
	}
	for _, bad := range []string{"motionkoppling", `"koppling"`, "veckobatch", "prose-citation", "similar-wording"} {
		if strings.Contains(rec.Body.String(), bad) {
			t.Fatalf("public Ändringslogg leaks %q: %s", bad, rec.Body)
		}
	}
}

func TestFraga339AppearsOnMotion3293OnlyAfterConfirmationAndLeavesAfterUndo(t *testing.T) {
	h := startCuratedQuestions(t)
	if got := motion3293Questions(t, h); len(got) != 0 {
		t.Fatalf("before any suggestion: %+v", got)
	}
	submitMotionLinks(t, h, "s1", `{`+fraga339+`,"confidence":0.9,"reasoning":"Frågan citerar 'en kommittémotion från Moderaterna redan 2017'"}`, 200)
	// Suggested: not public, neither on the motion page nor on the question.
	if got := motion3293Questions(t, h); len(got) != 0 {
		t.Fatalf("a suggestion is public: %+v", got)
	}
	// Retrying the same key gives the same answer; the same pair is not duplicated.
	submitMotionLinks(t, h, "s1", `{`+fraga339+`,"confidence":0.9,"reasoning":"Frågan citerar 'en kommittémotion från Moderaterna redan 2017'"}`, 200)
	if n := count(t, h, `SELECT count(*) FROM motion_kopplingar`); n != 1 {
		t.Fatalf("motion_kopplingar = %d", n)
	}

	b := curCreateBatch(t, h, "b1")
	if len(b.Items) != 1 || b.Items[0].Kind != "fraga" || b.Items[0].State != "foreslagen" {
		t.Fatalf("batch = %+v", b)
	}
	m := b.Items[0].Motion
	if m == nil || m.Rule != "prose-citation" || m.Question == nil || m.Question.Beteckning != "339" ||
		!strings.Contains(m.Question.Asker, "Nordin") || m.Motion.Beteckning != "3293" {
		t.Fatalf("the chat needs the question and the motion: %+v", m)
	}
	// In a batch but unanswered: still not public.
	if got := motion3293Questions(t, h); len(got) != 0 {
		t.Fatalf("an unanswered batch item is public: %+v", got)
	}

	// "ja men omvänt" has no meaning for a motion link: nothing is stored.
	curAnswer(t, h, b.ID, "a0", "1 ja men omvänt", 422)
	if got := motion3293Questions(t, h); len(got) != 0 {
		t.Fatalf("a refused answer changed something: %+v", got)
	}

	r := curAnswer(t, h, b.ID, "a1", "1 ja", 200)
	if len(r.Decided) != 1 || r.Decided[0].State != "bekraftad" {
		t.Fatalf("answer = %+v", r)
	}
	got := motion3293Questions(t, h)
	if len(got) != 1 || got[0].Beteckning != "339" || got[0].Kind != "fraga" || got[0].Rule != "redaktor-confirmed" ||
		got[0].Asker.Name != "Rickard Nordin" {
		t.Fatalf("after confirmation motion 3293 lists %+v", got)
	}
	var d struct {
		Motions []struct {
			Beteckning string `json:"beteckning"`
			Rule       string `json:"rule"`
		} `json:"motions"`
	}
	getJSON(t, h, "/api/questions/fraga/2022%2F23/339", &d)
	if len(d.Motions) != 1 || d.Motions[0].Beteckning != "3293" || d.Motions[0].Rule != "redaktor-confirmed" {
		t.Fatalf("question page motions = %+v", d.Motions)
	}
	assertNoCuratedInPublicLog(t, h)

	// A re-run of the questions ingestion does not touch a curated link.
	h.RunWorker(context.Background(), "fetch-questions")
	if got := motion3293Questions(t, h); len(got) != 1 {
		t.Fatalf("ingestion removed the confirmed link: %+v", got)
	}

	curUndo(t, h, r.Decided[0].EntryID, "u1")
	if got := motion3293Questions(t, h); len(got) != 0 {
		t.Fatalf("after undo motion 3293 still lists %+v", got)
	}
	getJSON(t, h, "/api/questions/fraga/2022%2F23/339", &d)
	if len(d.Motions) != 0 {
		t.Fatalf("after undo the question still links %+v", d.Motions)
	}
	// The undone suggestion is open again and rolls over into the next batch.
	b2 := curCreateBatch(t, h, "b2")
	if len(b2.Items) != 1 || !b2.Items[0].RolledOver || b2.Items[0].State != "foreslagen" {
		t.Fatalf("batch 2 = %+v", b2)
	}
	assertNoCuratedInPublicLog(t, h)
}

func TestRejectedAndBelowThresholdFragaLinksAreNeverPublic(t *testing.T) {
	h := startCuratedQuestions(t)
	// Below the threshold: kept as bortsorterad, never offered.
	submitMotionLinks(t, h, "s1", `{`+fraga339+`,"confidence":0.3,"reasoning":"svag"}`, 200)
	if st := fmt.Sprint(count(t, h, `SELECT count(*) FROM motion_kopplingar WHERE state='bortsorterad'`)); st != "1" {
		t.Fatalf("bortsorterad = %s", st)
	}
	rec := writeReq(h, "POST", "/api/write/veckobatch", testWriteToken, "b0", `{`+curProv+`}`)
	if rec.Code != 422 {
		t.Fatalf("a bortsorterad link was offered: %d %s", rec.Code, rec.Body)
	}
	if got := motion3293Questions(t, h); len(got) != 0 {
		t.Fatalf("bortsorterad is public: %+v", got)
	}
	// The same pair suggested again is left as it is.
	submitMotionLinks(t, h, "s2", `{`+fraga339+`,"confidence":0.95,"reasoning":"igen"}`, 200)
	if n := count(t, h, `SELECT count(*) FROM motion_kopplingar WHERE state='bortsorterad'`); n != 1 {
		t.Fatalf("an existing pair was re-judged")
	}
	// Reject a fresh one through the batch.
	if _, err := h.DB.Exec(context.Background(), `UPDATE motion_kopplingar SET state='foreslagen', confidence=0.8`); err != nil {
		t.Fatal(err)
	}
	b := curCreateBatch(t, h, "b1")
	curAnswer(t, h, b.ID, "a1", "1 nej", 200)
	if got := motion3293Questions(t, h); len(got) != 0 {
		t.Fatalf("a rejected link is public: %+v", got)
	}
	if st := count(t, h, `SELECT count(*) FROM motion_kopplingar WHERE state='avvisad'`); st != 1 {
		t.Fatalf("avvisad = %d", st)
	}
	assertNoCuratedInPublicLog(t, h)
}

func TestSuggestionsRejectUnknownAndMalformedReferences(t *testing.T) {
	h := startCuratedQuestions(t)
	submitMotionLinks(t, h, "s1", `{"kind":"fraga","question":{"kind":"fraga","riksmote":"2022/23","beteckning":"99999"},"motion":{"riksmote":"2017/18","beteckning":"3293"},"confidence":0.9,"reasoning":"x"}`, 422)
	submitMotionLinks(t, h, "s2", `{"kind":"fraga","question":{"kind":"fraga","riksmote":"2022/23","beteckning":"339"},"motion":{"riksmote":"2017/18","beteckning":"1"},"confidence":0.9,"reasoning":"x"}`, 422)
	submitMotionLinks(t, h, "s3", `{"kind":"krav","motion":{"riksmote":"2017/18","beteckning":"3293","yrkande":1},"confidence":0.9,"reasoning":"x"}`, 422)
	if rec := writeReq(h, "POST", "/api/write/motion-kopplingar", "", "s4", `{}`); rec.Code != 401 {
		t.Fatalf("no token: %d", rec.Code)
	}
	if n := count(t, h, `SELECT count(*) FROM motion_kopplingar`); n != 0 {
		t.Fatalf("a rejected request stored %d rows", n)
	}
}

type recurMembers struct {
	Yrkanden []struct {
		Number    int `json:"number"`
		Recurring *struct {
			Members []struct {
				Riksmote   string   `json:"riksmote"`
				Beteckning string   `json:"beteckning"`
				Number     int      `json:"number"`
				Parties    []string `json:"parties"`
				Rule       string   `json:"rule"`
			} `json:"members"`
		} `json:"recurring"`
	} `json:"yrkanden"`
}

// membersOf returns "party riksmote:beteckning/yrkande" for every member of
// the Återkommande krav of yrkande n on the page.
func membersOf(t *testing.T, h *Harness, enc, bet string, n int) []string {
	t.Helper()
	var b recurMembers
	getJSON(t, h, "/api/motions/"+enc+"/"+bet, &b)
	var out []string
	for _, y := range b.Yrkanden {
		if y.Number != n || y.Recurring == nil {
			continue
		}
		for _, m := range y.Recurring.Members {
			out = append(out, fmt.Sprintf("%s %s:%s/%d %s", strings.Join(m.Parties, "+"), m.Riksmote, m.Beteckning, m.Number, m.Rule))
		}
	}
	return out
}

const kravLink = `"kind":"krav","motion":{"riksmote":"2021/22","beteckning":"3638","yrkande":9},"other":{"riksmote":"2022/23","beteckning":"1675","yrkande":1}`

func TestNordinsCYrkandenJoinMsAterkommandeKravOnlyAfterConfirmationAndLeaveAfterUndo(t *testing.T) {
	h := startCuratedKrav(t)
	hasC := func(ms []string) bool {
		for _, m := range ms {
			if strings.HasPrefix(m, "C ") {
				return true
			}
		}
		return false
	}
	if ms := membersOf(t, h, "2021%2F22", "3638", 9); len(ms) != 4 || hasC(ms) {
		t.Fatalf("before: M's group = %v", ms)
	}
	submitMotionLinks(t, h, "s1", `{`+kravLink+`,"confidence":0.85,"reasoning":"C:s yrkande om vinstmarginalbeskattning på digitala varor är samma krav som M:s"}`, 200)
	if ms := membersOf(t, h, "2021%2F22", "3638", 9); len(ms) != 4 || hasC(ms) {
		t.Fatalf("a suggestion joined the group: %v", ms)
	}

	b := curCreateBatch(t, h, "b1")
	if len(b.Items) != 1 || b.Items[0].Kind != "krav" || b.Items[0].Motion == nil || b.Items[0].Motion.Other == nil ||
		b.Items[0].Motion.Rule != "similar-wording" || b.Items[0].Motion.Motion.Wording == "" {
		t.Fatalf("batch = %+v", b)
	}
	if ms := membersOf(t, h, "2022%2F23", "1675", 1); hasC(ms) && len(ms) != 3 {
		t.Fatalf("an unanswered item changed Nordin's group: %v", ms)
	}

	r := curAnswer(t, h, b.ID, "a1", "1 ja", 200)
	// Both sides now show the same, merged group: M's four and Nordin's three.
	for _, from := range []struct {
		enc, bet string
		n        int
	}{{"2021%2F22", "3638", 9}, {"2022%2F23", "1675", 1}} {
		ms := membersOf(t, h, from.enc, from.bet, from.n)
		if len(ms) != 7 {
			t.Fatalf("after confirmation %s:%s group = %v, want 7 members", from.enc, from.bet, ms)
		}
		var c, m int
		for _, s := range ms {
			switch {
			case strings.HasPrefix(s, "C "):
				c++
				if !strings.HasSuffix(s, "redaktor-confirmed") && from.bet == "3638" {
					t.Errorf("Nordin's filing in M's group lacks the confirmed rule: %s", s)
				}
			case strings.HasPrefix(s, "M "):
				m++
			}
		}
		if c != 3 || m != 4 {
			t.Fatalf("%s: C=%d M=%d in %v", from.bet, c, m, ms)
		}
	}
	// 2017/18:3293 is in M's group, so Nordin's motions list it too.
	if ms := membersOf(t, h, "2024%2F25", "989", 1); len(ms) != 7 {
		t.Fatalf("Nordin's 2024/25 filing group = %v", ms)
	}
	assertNoCuratedInPublicLog(t, h)

	curUndo(t, h, r.Decided[0].EntryID, "u1")
	if ms := membersOf(t, h, "2021%2F22", "3638", 9); len(ms) != 4 || hasC(ms) {
		t.Fatalf("after undo M's group = %v", ms)
	}
	if ms := membersOf(t, h, "2022%2F23", "1675", 1); len(ms) != 3 {
		t.Fatalf("after undo Nordin's group = %v", ms)
	}
}

func TestRejectedKravLinkNeverJoinsTheGroup(t *testing.T) {
	h := startCuratedKrav(t)
	// The same pair given in the other order is one link, not two.
	submitMotionLinks(t, h, "s1", `{`+kravLink+`,"confidence":0.85,"reasoning":"x"}`, 200)
	submitMotionLinks(t, h, "s2", `{"kind":"krav","motion":{"riksmote":"2022/23","beteckning":"1675","yrkande":1},"other":{"riksmote":"2021/22","beteckning":"3638","yrkande":9},"confidence":0.7,"reasoning":"x"}`, 200)
	if n := count(t, h, `SELECT count(*) FROM motion_kopplingar`); n != 1 {
		t.Fatalf("the same pair in the other order made %d links", n)
	}
	b := curCreateBatch(t, h, "b1")
	curAnswer(t, h, b.ID, "a1", "1 nej", 200)
	if ms := membersOf(t, h, "2021%2F22", "3638", 9); len(ms) != 4 {
		t.Fatalf("a rejected link joined the group: %v", ms)
	}
	if ms := membersOf(t, h, "2022%2F23", "1675", 1); len(ms) != 3 {
		t.Fatalf("a rejected link joined Nordin's group: %v", ms)
	}
	assertNoCuratedInPublicLog(t, h)
}

func TestQuestionAndKravLinksShareOneNumberedBatch(t *testing.T) {
	// Both kinds in one Veckobatch need one database holding a fråga and
	// yrkanden; seed the minimum by SQL, as the Löfte batch tests do.
	h := startCuratedKrav(t)
	ctx := context.Background()
	for _, q := range []string{
		`INSERT INTO questions (dok_id, kind, riksmote, beteckning, title, systemdatum) VALUES ('HQ1','fraga','2022/23','339','t', now())`,
	} {
		if _, err := h.DB.Exec(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	submitMotionLinks(t, h, "s1", `{`+kravLink+`,"confidence":0.7,"reasoning":"x"},
		{"kind":"fraga","question":{"kind":"fraga","riksmote":"2022/23","beteckning":"339"},"motion":{"riksmote":"2021/22","beteckning":"3638"},"confidence":0.9,"reasoning":"y"}`, 200)
	b := curCreateBatch(t, h, "b1")
	if len(b.Items) != 2 || b.Items[0].Kind != "fraga" || b.Items[1].Kind != "krav" || b.Items[0].Nr != 1 || b.Items[1].Nr != 2 {
		t.Fatalf("batch = %+v; want fraga (0.9) as 1, krav (0.7) as 2", b)
	}
	// Answering one leaves the other open to roll over.
	r := curAnswer(t, h, b.ID, "a1", "2 ja", 200)
	if len(r.Unanswered) != 1 || r.Unanswered[0] != 1 || r.Closed {
		t.Fatalf("answer = %+v", r)
	}
	if ms := membersOf(t, h, "2021%2F22", "3638", 9); len(ms) != 7 {
		t.Fatalf("krav not confirmed: %v", ms)
	}
	var qs struct {
		Questions []questionItem `json:"questions"`
	}
	getJSON(t, h, "/api/motions/2021%2F22/3638/questions", &qs)
	if len(qs.Questions) != 0 {
		t.Fatalf("the unanswered fråga is public: %+v", qs.Questions)
	}
}
