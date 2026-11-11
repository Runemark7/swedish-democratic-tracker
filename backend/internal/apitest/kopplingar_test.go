package apitest

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// These tests seed voteringar, betänkanden and Löften directly in the test
// database. They exercise the candidates/suggestions seam, which reads only
// Postgres; nothing here is a recorded Riksdagen response.

type kSeed struct {
	h                         *Harness
	loefteSkU, loefteAU       int64
	loefteProposed            int64
	vNew, vOld, vOutside, vAU int64
	vNewID                    string
}

func seedVotering(t *testing.T, h *Harness, voteringID, bet, session string, datum string, created string) int64 {
	t.Helper()
	var id int64
	if err := h.DB.QueryRow(context.Background(),
		`INSERT INTO voteringar (votering_id, beteckning, forslagspunkt, session, system_datum, created_at)
		 VALUES ($1,$2,'3',$3,$4::timestamptz,$5::timestamptz) RETURNING id`,
		voteringID, bet, session, datum, created).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func seedLoefte(t *testing.T, h *Harness, party, text, vis string, utskott string) int64 {
	t.Helper()
	var id int64
	if err := h.DB.QueryRow(context.Background(),
		`INSERT INTO loften (party, valmanifest_id, text, quote, source_url, utskott, mandate_period, visibility, approved_at)
		 VALUES ($1, (SELECT id FROM valmanifest WHERE party=$1 ORDER BY id LIMIT 1), $2, 'citat '||$2, 'https://example.org/v',
		         string_to_array($3, ','), '2026-2030', $4, CASE WHEN $4='approved' THEN now() END) RETURNING id`,
		party, text, utskott, vis).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func startK(t *testing.T) *kSeed {
	t.Helper()
	h := startWrite(t)
	s := &kSeed{h: h}
	s.loefteSkU = seedLoefte(t, h, "S", "Skatt sänks", "approved", "SkU")
	s.loefteAU = seedLoefte(t, h, "S", "Arbetsrätt", "approved", "AU")
	s.loefteProposed = seedLoefte(t, h, "S", "Bara förslag", "proposed", "SkU")
	// Closed Veckobatch: voteringar ingested before it are not new.
	if _, err := h.DB.Exec(context.Background(),
		`INSERT INTO veckobatch (created_at, closed_at) VALUES ('2026-10-04 06:00+00','2026-10-05 06:00+00')`); err != nil {
		t.Fatal(err)
	}
	s.vOld = seedVotering(t, h, "K-OLD", "SkU1", "2026/27", "2026-10-01 10:00+00", "2026-10-02 10:00+00")
	s.vNewID = "K-NEW"
	s.vNew = seedVotering(t, h, s.vNewID, "SkU2", "2026/27", "2026-10-20 10:00+00", "2026-10-21 10:00+00")
	s.vAU = seedVotering(t, h, "K-AU", "AU3", "2026/27", "2026-10-20 11:00+00", "2026-10-21 10:00+00")
	s.vOutside = seedVotering(t, h, "K-OUT", "SkU9", "2023/24", "2024-03-20 10:00+00", "2026-10-21 10:00+00")
	// Betänkande, förslagspunkt and the motion decided there, for the candidate payload.
	ctx := context.Background()
	for _, q := range []string{
		`INSERT INTO betankanden (riksmote, beteckning, dok_id, found, title, organ) VALUES ('2026/27','SkU2','D2',true,'Skatt på kapital','SkU')`,
		`INSERT INTO forslagspunkter (riksmote, beteckning, punkt, rubrik, forslag_text, votering_id) VALUES ('2026/27','SkU2','3','Kapitalskatt','Riksdagen avslår motionerna','K-NEW')`,
		`INSERT INTO motions (dok_id, riksmote, beteckning, title, subtyp, systemdatum) VALUES ('M1','2026/27','2900','Sänk skatten','kommittémotion', now())`,
		`INSERT INTO motion_authors (dok_id, ordning, party) VALUES ('M1',1,'M')`,
		`INSERT INTO motion_yrkanden (dok_id, nummer, wording) VALUES ('M1',1,'Riksdagen ska sänka skatten')`,
		`INSERT INTO yrkande_forslagspunkt (motion_dok_id, yrkande_nummer, riksmote, beteckning, punkt, linked_by) VALUES ('M1',1,'2026/27','SkU2','3','behandlas_i_punkt')`,
	} {
		if _, err := h.DB.Exec(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	return s
}

func (s *kSeed) candidates(t *testing.T) []struct {
	VoteringID string                 `json:"voteringId"`
	Betankande struct{ Title string } `json:"betankande"`
	Motions    []struct {
		Beteckning string
		Wording    string
		Party      string
	} `json:"motions"`
	Loften []struct {
		ID    int64
		Quote string
	} `json:"loften"`
} {
	t.Helper()
	rec := writeReq(s.h, "GET", "/api/write/candidates", testWriteToken, "", "")
	if rec.Code != 200 {
		t.Fatalf("candidates: %d %s", rec.Code, rec.Body)
	}
	return decode[struct {
		Voteringar []struct {
			VoteringID string                 `json:"voteringId"`
			Betankande struct{ Title string } `json:"betankande"`
			Motions    []struct {
				Beteckning string
				Wording    string
				Party      string
			} `json:"motions"`
			Loften []struct {
				ID    int64
				Quote string
			} `json:"loften"`
		} `json:"voteringar"`
	}](t, rec).Voteringar
}

func suggestBody(items ...string) string {
	return `{"threshold":0.6,"chatRef":"bot-run-1","onBehalfOf":"redaktor","suggestions":[` + strings.Join(items, ",") + `]}`
}

func sug(loefte int64, votering string, dir string, conf float64, reasoning string) string {
	return fmt.Sprintf(`{"loefteId":%d,"voteringId":%q,"direction":%q,"confidence":%v,"reasoning":%q}`, loefte, votering, dir, conf, reasoning)
}

func TestCandidatesOnlyNewVoteringarAndApprovedLoften(t *testing.T) {
	s := startK(t)
	if rec := writeReq(s.h, "GET", "/api/write/candidates", "", "", ""); rec.Code != 401 {
		t.Fatalf("no token: %d", rec.Code)
	}
	c := s.candidates(t)
	if len(c) != 2 || c[0].VoteringID != "K-NEW" || c[1].VoteringID != "K-AU" {
		t.Fatalf("candidates = %+v; want K-NEW and K-AU (K-OLD predates the last batch, K-OUT is outside the period)", c)
	}
	if len(c[1].Loften) != 1 || c[1].Loften[0].ID != s.loefteAU {
		t.Fatalf("K-AU Löften = %+v; want only the AU Löfte", c[1].Loften)
	}
	if len(c[0].Loften) != 1 || c[0].Loften[0].ID != s.loefteSkU {
		t.Fatalf("Löften = %+v; want only the approved SkU Löfte", c[0].Loften)
	}
	if !strings.Contains(c[0].Loften[0].Quote, "citat") {
		t.Fatal("candidate Löfte carries no quote")
	}
	if c[0].Betankande.Title != "Skatt på kapital" || len(c[0].Motions) != 1 ||
		c[0].Motions[0].Wording != "Riksdagen ska sänka skatten" || c[0].Motions[0].Party != "M" {
		t.Fatalf("payload lacks betänkande/motion context: %+v", c[0])
	}

	// With no closed batch at all, K-OLD is new too.
	if _, err := s.h.DB.Exec(context.Background(), `DELETE FROM veckobatch`); err != nil {
		t.Fatal(err)
	}
	if got := len(s.candidates(t)); got != 3 {
		t.Fatalf("without a batch: %d candidates, want 3 (K-OLD, K-NEW, K-AU)", got)
	}
	// An open (unclosed) batch is not "the last Veckobatch".
	if _, err := s.h.DB.Exec(context.Background(), `INSERT INTO veckobatch (created_at) VALUES (now() + interval '1 day')`); err != nil {
		t.Fatal(err)
	}
	if got := len(s.candidates(t)); got != 3 {
		t.Fatalf("open batch narrowed candidates: %d", got)
	}
}

func TestSuggestionsBelowThresholdAreBortsorteradAndNeverPublic(t *testing.T) {
	s := startK(t)
	const secretLow, secretHigh = "HEMLIG-RESONEMANG-LAG", "HEMLIG-RESONEMANG-HOG"
	rec := writeReq(s.h, "POST", "/api/write/kopplingar", testWriteToken, "k1", suggestBody(
		sug(s.loefteSkU, "K-NEW", "Nej", 0.59, secretLow),
		sug(s.loefteAU, "K-AU", "Ja", 0.6, secretHigh),
	))
	if rec.Code != 200 {
		t.Fatalf("submit: %d %s", rec.Code, rec.Body)
	}
	n := func(state string) int {
		return count(t, s.h, `SELECT count(*) FROM kopplingar WHERE state=$1`, state)
	}
	if n("bortsorterad") != 1 || n("foreslagen") != 1 {
		t.Fatalf("states: bortsorterad=%d foreslagen=%d", n("bortsorterad"), n("foreslagen"))
	}
	if count(t, s.h, `SELECT count(*) FROM kopplingar WHERE state='bortsorterad' AND confidence=0.59 AND reasoning=$1`, secretLow) != 1 {
		t.Fatal("dropped suggestion not stored with its confidence and reasoning")
	}

	// A bortsorterad pair is not a candidate again; the föreslagen one neither.
	for _, c := range s.candidates(t) {
		if len(c.Loften) != 0 && c.VoteringID == "K-NEW" {
			t.Fatalf("suggested pair offered again: %+v", c)
		}
	}

	// Nothing about either suggestion is public.
	for _, p := range []string{
		"/api/aendringslogg?limit=200",
		fmt.Sprintf("/api/loften/%d", s.loefteSkU), fmt.Sprintf("/api/loften/%d", s.loefteAU),
		"/api/loften?party=S", "/api/parties/S/loften",
		"/api/beslut", "/api/motions", "/api/votes",
	} {
		body := writeReq(s.h, "GET", p, "", "", "").Body.String()
		for _, leak := range []string{secretLow, secretHigh, "bortsorterad", "foreslagen", "kopplingar", "K-NEW"} {
			if leak == "K-NEW" && !strings.HasPrefix(p, "/api/beslut") && !strings.HasPrefix(p, "/api/votes") {
				continue
			}
			if strings.Contains(body, leak) {
				t.Errorf("%s leaks %q", p, leak)
			}
		}
	}
	if lg := decode[struct{ Entries []logEntry }](t, writeReq(s.h, "GET", "/api/aendringslogg?limit=200", "", "", "")); len(lg.Entries) != 0 {
		t.Fatalf("public Ändringslogg shows suggestion entries: %+v", lg.Entries)
	}
	// ...but they are in the audit log itself.
	if count(t, s.h, `SELECT count(*) FROM aendringslogg WHERE entity='koppling'`) != 2 {
		t.Fatal("suggestions not audit-logged")
	}
}

func TestSuggestionOutsideMandatePeriodRejected(t *testing.T) {
	s := startK(t)
	rec := writeReq(s.h, "POST", "/api/write/kopplingar", testWriteToken, "k2", suggestBody(
		sug(s.loefteSkU, "K-NEW", "Ja", 0.9, "giltig"),
		sug(s.loefteSkU, "K-OUT", "Ja", 0.9, "före mandatperioden"),
	))
	if rec.Code != 422 || !strings.Contains(rec.Body.String(), "mandate period") {
		t.Fatalf("outside period: %d %s", rec.Code, rec.Body)
	}
	if count(t, s.h, `SELECT count(*) FROM kopplingar`) != 0 || count(t, s.h, `SELECT count(*) FROM aendringslogg WHERE actor <> 'migration'`) != 0 {
		t.Fatal("a rejected batch stored something (must be all-or-nothing)")
	}
}

func TestSuggestionValidationAndIdempotency(t *testing.T) {
	s := startK(t)
	bad := map[string]string{
		"proposed Löfte":    suggestBody(sug(s.loefteProposed, "K-NEW", "Ja", 0.9, "x")),
		"unknown Löfte":     suggestBody(sug(98765432, "K-NEW", "Ja", 0.9, "x")),
		"unknown votering":  suggestBody(sug(s.loefteSkU, "NOPE", "Ja", 0.9, "x")),
		"bad direction":     suggestBody(sug(s.loefteSkU, "K-NEW", "Avstår", 0.9, "x")),
		"confidence > 1":    suggestBody(sug(s.loefteSkU, "K-NEW", "Ja", 1.5, "x")),
		"no reasoning":      suggestBody(sug(s.loefteSkU, "K-NEW", "Ja", 0.9, " ")),
		"no suggestions":    suggestBody(),
		"duplicate in body": suggestBody(sug(s.loefteSkU, "K-NEW", "Ja", 0.9, "x"), sug(s.loefteSkU, "K-NEW", "Nej", 0.8, "y")),
	}
	i := 0
	for name, b := range bad {
		i++
		if rec := writeReq(s.h, "POST", "/api/write/kopplingar", testWriteToken, fmt.Sprintf("b%d", i), b); rec.Code != 422 && rec.Code != 404 {
			t.Errorf("%s: %d %s", name, rec.Code, rec.Body)
		}
	}
	if rec := writeReq(s.h, "POST", "/api/write/kopplingar", "", "z", suggestBody()); rec.Code != 401 {
		t.Fatalf("no token: %d", rec.Code)
	}
	if count(t, s.h, `SELECT count(*) FROM kopplingar`) != 0 {
		t.Fatal("rejected suggestions stored")
	}

	body := suggestBody(sug(s.loefteSkU, "K-NEW", "Ja", 0.8, "ok"))
	r1 := writeReq(s.h, "POST", "/api/write/kopplingar", testWriteToken, "same", body)
	r2 := writeReq(s.h, "POST", "/api/write/kopplingar", testWriteToken, "same", body)
	if r1.Code != 200 || r2.Code != 200 || r2.Header().Get("Idempotent-Replayed") != "true" ||
		count(t, s.h, `SELECT count(*) FROM kopplingar`) != 1 {
		t.Fatalf("replay: %d %d %s", r1.Code, r2.Code, r2.Body)
	}
	if rec := writeReq(s.h, "POST", "/api/write/kopplingar", testWriteToken, "same",
		suggestBody(sug(s.loefteSkU, "K-NEW", "Nej", 0.8, "other"))); rec.Code != 409 {
		t.Fatalf("key reuse: %d", rec.Code)
	}
	// Re-suggesting the same pair under a new key changes nothing.
	r3 := writeReq(s.h, "POST", "/api/write/kopplingar", testWriteToken, "new-key", body)
	if r3.Code != 200 || count(t, s.h, `SELECT count(*) FROM kopplingar`) != 1 || !strings.Contains(r3.Body.String(), "existing") {
		t.Fatalf("existing pair: %d %s", r3.Code, r3.Body)
	}
}

func TestCandidatesPaging(t *testing.T) {
	s := startK(t)
	type page struct {
		Next       *int64 `json:"next"`
		Voteringar []struct {
			VoteringID string `json:"voteringId"`
		} `json:"voteringar"`
	}
	p1 := decode[page](t, writeReq(s.h, "GET", "/api/write/candidates?limit=1", testWriteToken, "", ""))
	if len(p1.Voteringar) != 1 || p1.Next == nil || p1.Voteringar[0].VoteringID != "K-NEW" {
		t.Fatalf("page 1: %+v", p1)
	}
	p2 := decode[page](t, writeReq(s.h, "GET", fmt.Sprintf("/api/write/candidates?limit=1&after=%d", *p1.Next), testWriteToken, "", ""))
	if len(p2.Voteringar) != 1 || p2.Voteringar[0].VoteringID != "K-AU" {
		t.Fatalf("page 2: %+v", p2)
	}
	p3 := decode[page](t, writeReq(s.h, "GET", fmt.Sprintf("/api/write/candidates?limit=1&after=%d", *p2.Next), testWriteToken, "", ""))
	if len(p3.Voteringar) != 0 || p3.Next != nil {
		t.Fatalf("page 3: %+v", p3)
	}
}
