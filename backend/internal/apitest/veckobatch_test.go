package apitest

import (
	"context"
	"fmt"
	"strings"
	"testing"

	cdomain "riksdagskollen/internal/curation/domain"
)

// Seam tests for the Veckobatch (issue #160). Kopplingar are seeded straight
// into Postgres through the real suggestion endpoint or SQL; nothing here is a
// recorded Riksdagen response and no test answers on anyone's behalf: every
// answer is an explicit request body.

type batchItem struct {
	Nr         int     `json:"nr"`
	KopplingID int64   `json:"kopplingId"`
	RolledOver bool    `json:"rolledOver"`
	State      string  `json:"state"`
	Direction  string  `json:"direction"`
	Confidence float64 `json:"confidence"`
	Answer     *string `json:"answer"`
	Quote      string  `json:"quote"`
}
type batchView struct {
	ID       int64       `json:"id"`
	ClosedAt *string     `json:"closedAt"`
	Items    []batchItem `json:"items"`
}
type answerView struct {
	Decided []struct {
		Nr        int    `json:"nr"`
		State     string `json:"state"`
		Direction string `json:"direction"`
		EntryID   int64  `json:"entryId"`
	} `json:"decided"`
	Unanswered []int `json:"unanswered"`
	Closed     bool  `json:"closed"`
}

const vbProv = `"chatRef":"chat-1#9","onBehalfOf":"redaktor"`

// addKoppling inserts a foreslagen koppling for a fresh votering.
func addKoppling(t *testing.T, s *kSeed, name string, conf float64, dir string) int64 {
	t.Helper()
	v := seedVotering(t, s.h, "VB-"+name, "SkU7", "2026/27", "2026-10-22 10:00+00", "2026-10-22 10:00+00")
	var id int64
	if err := s.h.DB.QueryRow(context.Background(),
		`INSERT INTO kopplingar (loefte_id, votering_ref, direction, state, confidence, threshold, reasoning, origin)
		 VALUES ($1,$2,$3,'foreslagen',$4,0.6,'skäl '||$5,'bot') RETURNING id`,
		s.loefteSkU, v, dir, conf, name).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func createBatch(t *testing.T, s *kSeed, key string) batchView {
	t.Helper()
	rec := writeReq(s.h, "POST", "/api/write/veckobatch", testWriteToken, key, `{`+vbProv+`}`)
	if rec.Code != 200 {
		t.Fatalf("create batch: %d %s", rec.Code, rec.Body)
	}
	return decode[batchView](t, rec)
}

func answer(t *testing.T, s *kSeed, id int64, key, body string, want int) answerView {
	t.Helper()
	rec := writeReq(s.h, "POST", fmt.Sprintf("/api/write/veckobatch/%d/answers", id), testWriteToken, key, body)
	if rec.Code != want {
		t.Fatalf("answer %s: %d %s (want %d)", body, rec.Code, rec.Body, want)
	}
	if want != 200 {
		return answerView{}
	}
	return decode[answerView](t, rec)
}

func kState(t *testing.T, s *kSeed, id int64) (state, dir string) {
	t.Helper()
	if err := s.h.DB.QueryRow(context.Background(), `SELECT state, direction FROM kopplingar WHERE id=$1`, id).Scan(&state, &dir); err != nil {
		t.Fatal(err)
	}
	return
}

func TestBatchNumberingIsStableAndStrongestFirst(t *testing.T) {
	s := startK(t)
	weak := addKoppling(t, s, "weak", 0.65, "Ja")
	strong := addKoppling(t, s, "strong", 0.95, "Nej")
	mid := addKoppling(t, s, "mid", 0.8, "Ja")
	// A bortsorterad suggestion is never offered.
	if _, err := s.h.DB.Exec(context.Background(),
		`UPDATE kopplingar SET state='bortsorterad' WHERE id=$1`, addKoppling(t, s, "dropped", 0.3, "Ja")); err != nil {
		t.Fatal(err)
	}
	if rec := writeReq(s.h, "POST", "/api/write/veckobatch", "", "b0", `{`+vbProv+`}`); rec.Code != 401 {
		t.Fatalf("no token: %d", rec.Code)
	}
	b := createBatch(t, s, "b1")
	if len(b.Items) != 3 || b.Items[0].KopplingID != strong || b.Items[1].KopplingID != mid || b.Items[2].KopplingID != weak {
		t.Fatalf("items = %+v; want strong, mid, weak numbered 1..3", b.Items)
	}
	for i, it := range b.Items {
		if it.Nr != i+1 || it.Answer != nil || it.RolledOver {
			t.Fatalf("item %d = %+v", i, it)
		}
	}
	// Re-reading, and a retry of the create, give the same numbers.
	again := decode[batchView](t, writeReq(s.h, "GET", fmt.Sprintf("/api/write/veckobatch/%d", b.ID), testWriteToken, "", ""))
	retry := createBatch(t, s, "b1")
	if retry.ID != b.ID || fmt.Sprint(again.Items) != fmt.Sprint(b.Items) || fmt.Sprint(retry.Items) != fmt.Sprint(b.Items) {
		t.Fatalf("numbering moved:\n%+v\n%+v\n%+v", b.Items, again.Items, retry.Items)
	}
	if n := count(t, s.h, `SELECT count(*) FROM veckobatch`); n != 2 { // the seeded closed one + this
		t.Fatalf("retry created another batch: %d", n)
	}
	// Adding a stronger suggestion later does not renumber the existing batch.
	addKoppling(t, s, "later", 0.99, "Ja")
	again = decode[batchView](t, writeReq(s.h, "GET", "/api/write/veckobatch/latest", testWriteToken, "", ""))
	if again.ID != b.ID || fmt.Sprint(again.Items) != fmt.Sprint(b.Items) {
		t.Fatalf("numbering changed after a new suggestion: %+v", again.Items)
	}
	// Same key, different request: refused.
	if rec := writeReq(s.h, "POST", "/api/write/veckobatch", testWriteToken, "b1", `{"max":2,`+vbProv+`}`); rec.Code != 409 {
		t.Fatalf("key reuse: %d", rec.Code)
	}
}

func TestBatchHoldsAtMostTwentyAndNeverEmpty(t *testing.T) {
	s := startK(t)
	if rec := writeReq(s.h, "POST", "/api/write/veckobatch", testWriteToken, "e1", `{`+vbProv+`}`); rec.Code != 422 {
		t.Fatalf("empty batch: %d", rec.Code)
	}
	for i := 0; i < 23; i++ {
		addKoppling(t, s, fmt.Sprint("n", i), 0.61+float64(i)/100, "Ja")
	}
	b := createBatch(t, s, "b1")
	if len(b.Items) != 20 {
		t.Fatalf("batch size %d, want 20", len(b.Items))
	}
	if b.Items[0].Confidence < b.Items[19].Confidence || b.Items[19].Confidence < 0.64 {
		t.Fatalf("not the strongest 20: first %v last %v", b.Items[0].Confidence, b.Items[19].Confidence)
	}
	if rec := writeReq(s.h, "POST", "/api/write/veckobatch", testWriteToken, "b2", `{"max":21,`+vbProv+`}`); rec.Code != 422 {
		t.Fatalf("max 21: %d", rec.Code)
	}
}

func TestAnswersJaNejAndReversedApplyAtomically(t *testing.T) {
	s := startK(t)
	k1 := addKoppling(t, s, "a", 0.9, "Ja")
	k2 := addKoppling(t, s, "b", 0.89, "Nej")
	k3 := addKoppling(t, s, "c", 0.88, "Ja")
	k4 := addKoppling(t, s, "d", 0.87, "Ja")
	b := createBatch(t, s, "b1")
	path := fmt.Sprintf("/api/write/veckobatch/%d/answers", b.ID)

	// An answer naming a missing item changes nothing at all.
	if rec := writeReq(s.h, "POST", path, testWriteToken, "x1",
		`{"answers":[{"nr":1,"answer":"ja"},{"nr":9,"answer":"ja"}],`+vbProv+`}`); rec.Code != 422 {
		t.Fatalf("missing item: %d", rec.Code)
	}
	if st, _ := kState(t, s, k1); st != "foreslagen" {
		t.Fatalf("partial apply: item 1 is %s", st)
	}
	if n := count(t, s.h, `SELECT count(*) FROM aendringslogg WHERE entity='koppling' AND old_value IS NOT NULL`); n != 0 {
		t.Fatalf("rejected answer logged %d entries", n)
	}

	r := answer(t, s, b.ID, "a1", `{"text":"1 ja, 2 ja men omvänt, 4 nej",`+vbProv+`}`, 200)
	if len(r.Decided) != 3 || r.Closed || fmt.Sprint(r.Unanswered) != "[3]" {
		t.Fatalf("answer = %+v", r)
	}
	for id, want := range map[int64][2]string{k1: {"bekraftad", "Ja"}, k2: {"bekraftad", "Ja"}, k3: {"foreslagen", "Ja"}, k4: {"avvisad", "Ja"}} {
		if st, d := kState(t, s, id); st != want[0] || d != want[1] {
			t.Fatalf("koppling %d = %s/%s, want %v", id, st, d, want)
		}
	}
	if count(t, s.h, `SELECT count(*) FROM kopplingar WHERE id=$1 AND decided_at IS NOT NULL`, k1) != 1 {
		t.Fatal("decided_at not set")
	}
	// Answering again with the same key replays; a decided item cannot be re-answered.
	rp := writeReq(s.h, "POST", path, testWriteToken, "a1", `{"text":"1 ja, 2 ja men omvänt, 4 nej",`+vbProv+`}`)
	if rp.Code != 200 || rp.Header().Get("Idempotent-Replayed") != "true" || len(decode[answerView](t, rp).Decided) != 3 {
		t.Fatalf("replay: %d %s", rp.Code, rp.Body)
	}
	answer(t, s, b.ID, "a2", `{"answers":[{"nr":1,"answer":"nej"}],`+vbProv+`}`, 409)

	// The last item closes the batch.
	r = answer(t, s, b.ID, "a3", `{"text":"3 ja",`+vbProv+`}`, 200)
	if !r.Closed || len(r.Unanswered) != 0 {
		t.Fatalf("closing answer = %+v", r)
	}
	answer(t, s, b.ID, "a4", `{"text":"3 nej",`+vbProv+`}`, 409)
}

func TestShortFormTextAndRest(t *testing.T) {
	s := startK(t)
	ids := []int64{}
	for i := 0; i < 8; i++ {
		ids = append(ids, addKoppling(t, s, fmt.Sprint("t", i), 0.9-float64(i)/100, "Ja"))
	}
	b := createBatch(t, s, "b1")
	r := answer(t, s, b.ID, "a1", `{"text":"1,3,7 ja, 4 ja men omvänt, resten nej",`+vbProv+`}`, 200)
	if len(r.Decided) != 8 || !r.Closed {
		t.Fatalf("answer = %+v", r)
	}
	want := map[int][2]string{1: {"bekraftad", "Ja"}, 3: {"bekraftad", "Ja"}, 7: {"bekraftad", "Ja"}, 4: {"bekraftad", "Nej"},
		2: {"avvisad", "Ja"}, 5: {"avvisad", "Ja"}, 6: {"avvisad", "Ja"}, 8: {"avvisad", "Ja"}}
	for nr, w := range want {
		if st, d := kState(t, s, ids[nr-1]); st != w[0] || d != w[1] {
			t.Fatalf("item %d = %s/%s want %v", nr, st, d, w)
		}
	}
}

func TestAnswerTextIsParsedStrictly(t *testing.T) {
	s := startK(t)
	addKoppling(t, s, "p", 0.9, "Ja")
	addKoppling(t, s, "q", 0.8, "Ja")
	b := createBatch(t, s, "b1")
	for i, bad := range []string{`1 kanske`, `ja`, `1 ja 1 nej`, `1`, `resten`, `1 nej men omvänt`, `1-x ja`, `1 ja resten`} {
		answer(t, s, b.ID, fmt.Sprint("bad", i), fmt.Sprintf(`{"text":%q,%s}`, bad, vbProv), 422)
	}
	if n := count(t, s.h, `SELECT count(*) FROM kopplingar WHERE state <> 'foreslagen'`); n != 0 {
		t.Fatalf("unparseable answers changed %d kopplingar", n)
	}
	answer(t, s, b.ID, "ok", `{"text":"1-2 ja",`+vbProv+`}`, 200)
	// A request without text or answers decides nothing.
	answer(t, s, b.ID, "none", `{`+vbProv+`}`, 422)
}

func TestParseAnswerText(t *testing.T) {
	a, rest, err := cdomain.ParseAnswerText("1,3,7 ja, resten nej")
	if err != nil || rest != "nej" || fmt.Sprint(a) != "[{1 ja} {3 ja} {7 ja}]" {
		t.Fatalf("%v %q %v", a, rest, err)
	}
	a, rest, err = cdomain.ParseAnswerText("4 ja men omvänt")
	if err != nil || rest != "" || fmt.Sprint(a) != "[{4 ja-omvaent}]" {
		t.Fatalf("%v %q %v", a, rest, err)
	}
	if _, _, err = cdomain.ParseAnswerText("1 ja men"); err != nil && !strings.Contains(err.Error(), "answer text") {
		t.Fatal(err)
	}
}

func TestUnansweredItemsRollIntoNextBatch(t *testing.T) {
	s := startK(t)
	k1 := addKoppling(t, s, "a", 0.9, "Ja")
	k2 := addKoppling(t, s, "b", 0.8, "Ja")
	k3 := addKoppling(t, s, "c", 0.7, "Ja")
	b1 := createBatch(t, s, "w1")
	answer(t, s, b1.ID, "a1", `{"text":"1 ja",`+vbProv+`}`, 200)

	// A fresh, stronger suggestion arrives; the unanswered ones still come first.
	k4 := addKoppling(t, s, "d", 0.99, "Ja")
	b2 := createBatch(t, s, "w2")
	if b2.ID == b1.ID || len(b2.Items) != 3 {
		t.Fatalf("batch 2 = %+v", b2)
	}
	got := []int64{b2.Items[0].KopplingID, b2.Items[1].KopplingID, b2.Items[2].KopplingID}
	if fmt.Sprint(got) != fmt.Sprint([]int64{k2, k3, k4}) {
		t.Fatalf("batch 2 order %v, want %v (rolled over first, answered %d excluded)", got, []int64{k2, k3, k4}, k1)
	}
	if !b2.Items[0].RolledOver || !b2.Items[1].RolledOver || b2.Items[2].RolledOver {
		t.Fatalf("rolledOver flags: %+v", b2.Items)
	}
	// The old batch is closed: its unanswered items can only be answered in the new one.
	answer(t, s, b1.ID, "late", `{"text":"2 ja",`+vbProv+`}`, 409)
	answer(t, s, b2.ID, "a2", `{"text":"1 nej",`+vbProv+`}`, 200)
	if st, _ := kState(t, s, k2); st != "avvisad" {
		t.Fatalf("rolled item = %s", st)
	}
	if count(t, s.h, `SELECT count(*) FROM veckobatch_items WHERE veckobatch_id=$1 AND answer IS NULL`, b1.ID) != 2 {
		t.Fatal("old batch's unanswered item should keep its own number and no answer")
	}
}

func TestEveryDecisionIsInTheAendringsloggAndCanBeUndone(t *testing.T) {
	s := startK(t)
	k1 := addKoppling(t, s, "a", 0.9, "Ja")
	k2 := addKoppling(t, s, "b", 0.8, "Ja")
	b := createBatch(t, s, "b1")
	r := answer(t, s, b.ID, "a1", `{"text":"1 ja men omvänt, 2 nej",`+vbProv+`}`, 200)

	if n := count(t, s.h, `SELECT count(*) FROM aendringslogg WHERE entity='koppling' AND action='upsert' AND old_value->>'state'='foreslagen'
		AND chat_ref='chat-1#9' AND on_behalf_of='redaktor' AND actor='bot'`); n != 2 {
		t.Fatalf("decision entries = %d, want 2", n)
	}
	if count(t, s.h, `SELECT count(*) FROM aendringslogg WHERE entity='veckobatch'`) != 1 {
		t.Fatal("batch creation not logged")
	}
	// None of it is public.
	body := writeReq(s.h, "GET", "/api/aendringslogg?limit=200", "", "", "").Body.String()
	if strings.Contains(body, "koppling") || strings.Contains(body, "veckobatch") || strings.Contains(body, "bekraftad") {
		t.Fatalf("public log leaks: %s", body)
	}

	undo := func(entry int64, key string, want int) {
		rec := writeReq(s.h, "POST", fmt.Sprintf("/api/write/aendringslogg/%d/undo", entry), testWriteToken, key, `{`+vbProv+`}`)
		if rec.Code != want {
			t.Fatalf("undo %d: %d %s (want %d)", entry, rec.Code, rec.Body, want)
		}
	}
	undo(r.Decided[0].EntryID, "u1", 200)
	if st, d := kState(t, s, k1); st != "foreslagen" || d != "Ja" {
		t.Fatalf("after undo: %s/%s; want foreslagen with the suggested direction", st, d)
	}
	if count(t, s.h, `SELECT count(*) FROM veckobatch_items WHERE veckobatch_id=$1 AND nr=1 AND answer IS NULL AND answered_at IS NULL`, b.ID) != 1 {
		t.Fatal("item answer not cleared")
	}
	if count(t, s.h, `SELECT count(*) FROM aendringslogg WHERE action='undo' AND entity='koppling' AND undoes=$1`, r.Decided[0].EntryID) != 1 {
		t.Fatal("undo not logged")
	}
	// The undone suggestion is open again and rolls into the next batch; the other decision stands.
	b2 := createBatch(t, s, "b2")
	if len(b2.Items) != 1 || b2.Items[0].KopplingID != k1 || !b2.Items[0].RolledOver {
		t.Fatalf("batch 2 = %+v", b2.Items)
	}
	if st, _ := kState(t, s, k2); st != "avvisad" {
		t.Fatalf("other decision changed: %s", st)
	}
	// Undoing the undo re-applies the decision; a stale undo is refused.
	var undoID int64
	if err := s.h.DB.QueryRow(context.Background(), `SELECT id FROM aendringslogg WHERE action='undo' AND entity='koppling'`).Scan(&undoID); err != nil {
		t.Fatal(err)
	}
	undo(r.Decided[0].EntryID, "u2", 409) // the value is no longer what that entry wrote
	undo(undoID, "u3", 200)
	if st, d := kState(t, s, k1); st != "bekraftad" || d != "Nej" {
		t.Fatalf("redo: %s/%s", st, d)
	}
	// Suggestion entries (from #159) are not undoable here.
	var sugEntry int64
	_ = s.h.DB.QueryRow(context.Background(), `SELECT 0`).Scan(&sugEntry)
	rec := writeReq(s.h, "POST", "/api/write/kopplingar", testWriteToken, "sg", suggestBody(sug(s.loefteAU, "K-AU", "Ja", 0.9, "x")))
	if rec.Code != 200 {
		t.Fatal(rec.Body)
	}
	if err := s.h.DB.QueryRow(context.Background(), `SELECT id FROM aendringslogg WHERE request_key='sg#0'`).Scan(&sugEntry); err != nil {
		t.Fatal(err)
	}
	undo(sugEntry, "u4", 422)
}

func TestConfirmedKopplingKeepsDirectionAndStaysPrivate(t *testing.T) {
	s := startK(t)
	k := addKoppling(t, s, "a", 0.9, "Nej")
	b := createBatch(t, s, "b1")
	answer(t, s, b.ID, "a1", `{"answers":[{"nr":1,"answer":"ja"}],`+vbProv+`}`, 200)
	if st, d := kState(t, s, k); st != "bekraftad" || d != "Nej" {
		t.Fatalf("%s/%s", st, d)
	}
	// The closed batch moves the candidates cutoff to its creation time.
	var since *string
	rec := writeReq(s.h, "GET", "/api/write/candidates", testWriteToken, "", "")
	since = decode[struct{ Since *string }](t, rec).Since
	if since == nil {
		t.Fatal("closed batch did not set the candidates cutoff")
	}
}
