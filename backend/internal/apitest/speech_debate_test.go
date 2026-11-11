package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Fixtures: fixtures/speech-debate. Speeches of riksmöte 2021/22 (the 14
// anföranden of the 2022-09-21 debate carrying rel_dok_id H901MJU33, plus one
// unrelated speech) and of 2002/03 (the 2003-05-28 debate "Fristående skolor",
// where Riksdagen's rel_dok_id is empty), recorded 2026-11-11 and reduced to
// those rows.

type debateBody struct {
	Debate *struct {
		Date     *string `json:"date"`
		Total    int     `json:"total"`
		Speeches []struct {
			SpeechID         int    `json:"speechId"`
			Name             string `json:"name"`
			Party            string `json:"party"`
			Date             string `json:"date"`
			ExcerptAvailable bool   `json:"excerptAvailable"`
			RiksdagenURL     string `json:"riksdagenUrl"`
			LinkRule         string `json:"linkRule"`
		} `json:"speeches"`
	} `json:"debate"`
}

func startDebate(t *testing.T) *Harness {
	t.Helper()
	h := startChain(t, chainDir(t, "speech-debate"), "{2021/22,2002/03}")
	ctx := context.Background()
	h.RunWorker(ctx, "fetch-politicians")
	h.RunWorker(ctx, "fetch-speeches")
	if h.Served("/anforandelista/") == 0 {
		t.Fatal("ingestion did not read the recording")
	}
	return h
}

func getDebate(t *testing.T, h *Harness, path string) debateBody {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d: %s", path, rec.Code, rec.Body.String())
	}
	var b debateBody
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	return b
}

func TestSpeechCarryingRelDokIDH901MJU33LinksToThatBetankande(t *testing.T) {
	h := startDebate(t)
	h.RunWorker(context.Background(), "link-speeches")

	b := getDebate(t, h, "/api/betankanden/2021%2F22/MJU33")
	if b.Debate == nil || b.Debate.Total != 14 || len(b.Debate.Speeches) != 14 {
		t.Fatalf("debate = %+v", b.Debate)
	}
	if b.Debate.Date == nil || *b.Debate.Date != "2022-09-21" {
		t.Errorf("date = %v", b.Debate.Date)
	}
	parties := map[string]int{}
	var gardfjell bool
	for _, s := range b.Debate.Speeches {
		parties[s.Party]++
		if s.LinkRule != "rel_dok_id" || s.Date != "2022-09-21" {
			t.Errorf("speech %+v", s)
		}
		if s.Name == "Maria Gardfjell" && s.Party == "MP" {
			gardfjell = true
		}
		if s.ExcerptAvailable {
			t.Errorf("list carries no text yet, excerpt claimed: %+v", s)
		}
		if !strings.HasPrefix(s.RiksdagenURL, "https://data.riksdagen.se/anforande/H909142-") || !strings.HasSuffix(s.RiksdagenURL, "/html") {
			t.Errorf("riksdagen url = %q", s.RiksdagenURL)
		}
	}
	if !gardfjell || parties["MP"] != 5 || parties["S"] != 3 || parties["SD"] != 3 {
		t.Errorf("speakers: gardfjell=%v parties=%v", gardfjell, parties)
	}
}

func TestSpeechWithoutRelDokIDLinksByTitleAndDebateDate2002(t *testing.T) {
	h := startDebate(t)
	ctx := context.Background()
	// 2002/03:UBU15 as the betänkande worker stores it: title and the Debatt
	// date of Riksdagen's listing (2003-05-28, see motions/adapters/riksdagen).
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO betankanden (riksmote, beteckning, dok_id, found, title, debate_date, debate_checked_at)
		VALUES ('2002/03', 'UBU15', 'GQ01UBU15', true, 'Fristående skolor', '2003-05-28', now())`); err != nil {
		t.Fatal(err)
	}
	h.RunWorker(ctx, "link-speeches")

	b := getDebate(t, h, "/api/betankanden/2002%2F03/UBU15")
	if b.Debate == nil || b.Debate.Total != 47 {
		t.Fatalf("debate = %+v", b.Debate)
	}
	for _, s := range b.Debate.Speeches {
		if s.LinkRule != "avsnittsrubrik_datum" {
			t.Errorf("rule = %q", s.LinkRule)
		}
	}
}

func TestRelinkIsRerunnableAndUnlinkedSpeechesStayOnMemberPages(t *testing.T) {
	h := startDebate(t)
	ctx := context.Background()
	h.RunWorker(ctx, "link-speeches")
	h.RunWorker(ctx, "link-speeches") // idempotent

	// Wipe the links, as for speeches stored before linking existed.
	if _, err := h.DB.Exec(ctx, `UPDATE speeches SET betankande_dok_id = NULL, betankande_link_rule = NULL`); err != nil {
		t.Fatal(err)
	}
	if b := getDebate(t, h, "/api/betankanden/2021%2F22/MJU33"); b.Debate == nil || b.Debate.Total != 0 {
		t.Fatalf("links should be gone: %+v", b.Debate)
	}
	h.RunWorker(ctx, "link-speeches")
	if b := getDebate(t, h, "/api/betankanden/2021%2F22/MJU33"); b.Debate == nil || b.Debate.Total != 14 {
		t.Fatalf("relink should restore the debate: %+v", b.Debate)
	}

	// Annika Hirvonen's speech of 2022-06-14 belongs to no betänkande we know
	// of: it stays unlinked and stays on her member page.
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/speeches/by-politician/0819601620927", nil))
	var ss []struct {
		Date string `json:"date"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &ss); err != nil || len(ss) != 1 || ss[0].Date[:10] != "2022-06-14" {
		t.Fatalf("member page = %d %s (%v)", rec.Code, rec.Body.String(), err)
	}
	var linked int
	if err := h.DB.QueryRow(ctx, `SELECT count(*) FROM speeches WHERE betankande_dok_id IS NOT NULL AND politician_id = '0819601620927'`).Scan(&linked); err != nil || linked != 0 {
		t.Fatalf("unrelated speech linked: %d %v", linked, err)
	}
}

// Fixtures: fixtures/speech-debate-motion. Motion 2017/18:3293, its betänkande
// 2017/18:SkU11 (as in decision-chain) and the 15 anföranden of the 2018-03-22
// debate on it (rel_dok_id H501SkU11), recorded 2026-11-11.
func startMotionDebate(t *testing.T) *Harness {
	t.Helper()
	h := Start(t, chainDir(t, "speech-debate-motion"), WithNow(time.Date(2018, 6, 1, 12, 0, 0, 0, time.UTC)))
	ctx := context.Background()
	if _, err := h.DB.Exec(ctx, `DELETE FROM mandate_periods`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
		VALUES ('test', 'test', '2014-09-14', '2018-09-09', '{2017/18}')`); err != nil {
		t.Fatal(err)
	}
	// Read the riksmöte from the start of March, so the previous one is not asked for.
	if _, err := h.DB.Exec(ctx, `INSERT INTO ingestion_cursors (data_type, last_date) VALUES ('speeches', '2018-03-01')`); err != nil {
		t.Fatal(err)
	}
	for _, w := range []string{"fetch-politicians", "fetch-speeches", "fetch-motions", "fetch-betankanden", "link-speeches"} {
		h.RunWorker(ctx, w)
	}
	return h
}

type motionDebateBody struct {
	Yrkanden []struct {
		Decisions []debateBody `json:"decisions"`
	} `json:"yrkanden"`
}

func TestMotionAndBetankandePagesShowTheDebateOfTheDecidingBetankande(t *testing.T) {
	h := startMotionDebate(t)

	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/motions/2017%2F18/3293", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("motion %d: %s", rec.Code, rec.Body.String())
	}
	var m motionDebateBody
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if len(m.Yrkanden) != 1 || len(m.Yrkanden[0].Decisions) != 1 {
		t.Fatalf("yrkanden = %s", rec.Body.String())
	}
	d := m.Yrkanden[0].Decisions[0].Debate
	if d == nil || d.Total != 15 || len(d.Speeches) != 15 || d.Date == nil || *d.Date != "2018-03-22" {
		t.Fatalf("motion decision debate = %+v", d)
	}
	b := getDebate(t, h, "/api/betankanden/2017%2F18/SkU11")
	if b.Debate == nil || b.Debate.Total != 15 {
		t.Fatalf("betänkande debate = %+v", b.Debate)
	}
	names := map[string]bool{}
	for _, s := range b.Debate.Speeches {
		names[s.Name+" "+s.Party] = true
	}
	if !names["Jörgen Warborn M"] || !names["Malcolm Momodou Jallow V"] {
		t.Errorf("speakers = %v", names)
	}
}

// The debate date read from Riksdagen's listing ("Debatt 2018-03-22") carries
// the fallback: with rel_dok_id gone the same speeches still link, by the
// rubrik "Mervärdesskatt" and the protocol date.
func TestDebateLinksByTitleAndDateWhenRelDokIDIsMissing(t *testing.T) {
	h := startMotionDebate(t)
	ctx := context.Background()
	if _, err := h.DB.Exec(ctx, `UPDATE speeches SET related_dok_id = NULL`); err != nil {
		t.Fatal(err)
	}
	h.RunWorker(ctx, "link-speeches")
	b := getDebate(t, h, "/api/betankanden/2017%2F18/SkU11")
	if b.Debate == nil || b.Debate.Total != 15 {
		t.Fatalf("debate = %+v", b.Debate)
	}
	for _, s := range b.Debate.Speeches {
		if s.LinkRule != "avsnittsrubrik_datum" {
			t.Errorf("rule = %q", s.LinkRule)
		}
	}
}
