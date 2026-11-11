package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"riksdagskollen/internal/ingestion/workers"
	"runtime"
	"testing"
	"time"
)

// Fixtures in fixtures/decision-chain were recorded from data.riksdagen.se on
// 2026-11-10: motions 2017/18:3293, 2017/18:4202, 2018/19:2900 and
// 2021/22:3638, the betänkanden they are handled in, and the votering of
// 2021/22:SkU16 förslagspunkt 6. fixtures/decision-chain-missing holds the
// 2018/19:SkU11 recording with the punkt that cites 2018/19:2900 left out.

func chainDir(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "fixtures", name)
}

type decisionBody struct {
	BetankandeRiksmote   string `json:"betankandeRiksmote"`
	BetankandeBeteckning string `json:"betankandeBeteckning"`
	BetankandeSourceURL  string `json:"betankandeSourceUrl"`
	Punkt                string `json:"punkt"`
	Rubrik               string `json:"rubrik"`
	LinkedBy             string `json:"linkedBy"`
	Mode                 string `json:"mode"`
	ModeLabel            string `json:"modeLabel"`
	Alternativ           string `json:"alternativ"`
	Outcome              string `json:"outcome"`
	VoteringStored       bool   `json:"voteringStored"`
	Parties              []struct {
		Party      string `json:"party"`
		Partilinje string `json:"partilinje"`
	} `json:"parties"`
}

type chainMotion struct {
	Outcome struct {
		Word   string         `json:"word"`
		Counts map[string]int `json:"counts"`
		Total  int            `json:"total"`
	} `json:"outcome"`
	Yrkanden []struct {
		Number    int            `json:"number"`
		Outcome   string         `json:"outcome"`
		Decisions []decisionBody `json:"decisions"`
	} `json:"yrkanden"`
}

func startChain(t *testing.T, dir string, riksmoten string) *Harness {
	t.Helper()
	h := Start(t, dir, WithNow(time.Date(2022, 6, 1, 12, 0, 0, 0, time.UTC)))
	ctx := context.Background()
	if _, err := h.DB.Exec(ctx, `DELETE FROM mandate_periods`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
		VALUES ('test', 'test', '2014-09-14', '2022-09-11', $1::text[])`, riksmoten); err != nil {
		t.Fatal(err)
	}
	return h
}

func runChain(h *Harness) {
	ctx := context.Background()
	h.RunWorker(ctx, "fetch-politicians")
	h.RunWorker(ctx, "fetch-votes")
	h.Backfill(ctx, workers.BackfillOptions{})
}

func getChain(t *testing.T, h *Harness, rm, bet string) chainMotion {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/motions/"+rm+"/"+bet, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET motion %s/%s = %d: %s", rm, bet, rec.Code, rec.Body.String())
	}
	var m chainMotion
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestMotion3293Yrkande1IsDecidedByAcklamationInSkU11Punkt6(t *testing.T) {
	h := startChain(t, chainDir(t, "decision-chain"), "{2017/18,2018/19,2021/22}")
	runChain(h)
	m := getChain(t, h, "2017%2F18", "3293")
	if len(m.Yrkanden) != 1 || len(m.Yrkanden[0].Decisions) != 1 {
		t.Fatalf("yrkanden = %+v", m.Yrkanden)
	}
	y := m.Yrkanden[0]
	d := y.Decisions[0]
	if d.BetankandeRiksmote != "2017/18" || d.BetankandeBeteckning != "SkU11" || d.Punkt != "6" || d.Rubrik != "Övriga frågor" {
		t.Errorf("decision = %+v", d)
	}
	if d.LinkedBy != "forslag_text" {
		t.Errorf("linkedBy = %q", d.LinkedBy)
	}
	if y.Outcome != "Avslagen" || m.Outcome.Word != "Avslagen" {
		t.Errorf("outcome = %q / %q", y.Outcome, m.Outcome.Word)
	}
	if d.Mode != "acklamation" || d.ModeLabel != "Avgjord utan omröstning" {
		t.Errorf("mode = %q / %q", d.Mode, d.ModeLabel)
	}
	if len(d.Parties) != 0 {
		t.Errorf("an acklamation has no party table, got %+v", d.Parties)
	}
	if d.BetankandeSourceURL != "https://data.riksdagen.se/dokument/H501SkU11.html" {
		t.Errorf("source = %q", d.BetankandeSourceURL)
	}
}

func TestMotion3638Yrkande9IsDecidedByVoteringWithPartilinjePerParty(t *testing.T) {
	h := startChain(t, chainDir(t, "decision-chain"), "{2017/18,2018/19,2021/22}")
	runChain(h)
	m := getChain(t, h, "2021%2F22", "3638")
	if len(m.Yrkanden) != 1 || len(m.Yrkanden[0].Decisions) != 1 {
		t.Fatalf("yrkanden = %+v", m.Yrkanden)
	}
	d := m.Yrkanden[0].Decisions[0]
	if d.BetankandeBeteckning != "SkU16" || d.Punkt != "6" || d.Mode != "votering" || !d.VoteringStored {
		t.Fatalf("decision = %+v", d)
	}
	if d.Outcome != "Avslagen" {
		t.Errorf("outcome = %q", d.Outcome)
	}
	if d.Alternativ != "Utskottets förslag mot reservation 12 (M)" {
		t.Errorf("alternativ = %q", d.Alternativ)
	}
	want := map[string]string{"M": "Nej", "S": "Ja", "SD": "Ja", "V": "Ja", "L": "Ja", "MP": "Ja", "C": "Avstår", "KD": "Avstår"}
	got := map[string]string{}
	for _, p := range d.Parties {
		if p.Party == "-" {
			if p.Partilinje != "" {
				t.Errorf("party-less ballots got a Partilinje %q", p.Partilinje)
			}
			continue
		}
		got[p.Party] = p.Partilinje
	}
	for party, line := range want {
		if got[party] != line {
			t.Errorf("%s Partilinje = %q, want %q (all: %v)", party, got[party], line, got)
		}
	}
	if len(got) != len(want) {
		t.Errorf("parties = %v", got)
	}
}

func TestMotion2900Yrkande26WithEmptyUtskottetAndKammarenGetsOutcomeFromThePunkt(t *testing.T) {
	h := startChain(t, chainDir(t, "decision-chain"), "{2017/18,2018/19,2021/22}")
	runChain(h)
	m := getChain(t, h, "2018%2F19", "2900")
	if len(m.Yrkanden) != 1 || m.Yrkanden[0].Number != 26 {
		t.Fatalf("yrkanden = %+v", m.Yrkanden)
	}
	y := m.Yrkanden[0]
	if y.Outcome != "Avslagen" {
		t.Errorf("outcome = %q", y.Outcome)
	}
	if len(y.Decisions) != 1 || y.Decisions[0].BetankandeBeteckning != "SkU11" || y.Decisions[0].Punkt != "1" || y.Decisions[0].LinkedBy != "forslag_text" {
		t.Errorf("decisions = %+v", y.Decisions)
	}
	if y.Decisions[0].Mode != "votering" {
		t.Errorf("mode = %q", y.Decisions[0].Mode)
	}
}

func TestMotion4202IsBifallenByAcklamationInTheNextRiksmote(t *testing.T) {
	h := startChain(t, chainDir(t, "decision-chain"), "{2017/18,2018/19,2021/22}")
	runChain(h)
	m := getChain(t, h, "2017%2F18", "4202")
	y := m.Yrkanden[0]
	if y.Outcome != "Bifallen" || m.Outcome.Word != "Bifallen" {
		t.Errorf("outcome = %q / %q", y.Outcome, m.Outcome.Word)
	}
	d := y.Decisions[0]
	if d.BetankandeRiksmote != "2018/19" || d.BetankandeBeteckning != "CU2" || d.Punkt != "2" || d.Mode != "acklamation" {
		t.Errorf("decision = %+v", d)
	}
}

func TestMotionMissingFromEveryForslagspunktIsEjAvgjordAn(t *testing.T) {
	h := startChain(t, chainDir(t, "decision-chain-missing"), "{2018/19}")
	h.Backfill(context.Background(), workers.BackfillOptions{})
	if h.Served("/dokumentstatus/H601SkU11.json") == 0 {
		t.Fatal("the betänkande was not read")
	}
	m := getChain(t, h, "2018%2F19", "2900")
	y := m.Yrkanden[0]
	if y.Outcome != "Ej avgjord än" || m.Outcome.Word != "Ej avgjord än" {
		t.Errorf("outcome = %q / %q", y.Outcome, m.Outcome.Word)
	}
	if len(y.Decisions) != 0 {
		t.Errorf("must not guess a punkt, got %+v", y.Decisions)
	}
}

func TestBetankandeIngestionIsIdempotentAndResumable(t *testing.T) {
	h := startChain(t, chainDir(t, "decision-chain"), "{2017/18,2018/19,2021/22}")
	ctx := context.Background()
	runChain(h)
	first, _ := json.Marshal(getChain(t, h, "2021%2F22", "3638"))
	reads := h.Served("/dokumentstatus/H901SkU16.json") + h.Served("/dokumentstatus/H501SkU11.json") + h.Served("/dokumentstatus/H601SkU11.json") + h.Served("/dokumentstatus/H601CU2.json")
	if reads != 4 {
		t.Fatalf("first run read %d betänkanden, want 4", reads)
	}
	h.RunWorker(ctx, "fetch-betankanden")
	if got := h.Served("/dokumentstatus/H901SkU16.json") + h.Served("/dokumentstatus/H501SkU11.json") + h.Served("/dokumentstatus/H601SkU11.json") + h.Served("/dokumentstatus/H601CU2.json"); got != reads {
		t.Errorf("re-run read betänkanden again: %d, was %d", got, reads)
	}
	again, _ := json.Marshal(getChain(t, h, "2021%2F22", "3638"))
	if string(first) != string(again) {
		t.Errorf("re-run changed the motion:\n%s\n%s", first, again)
	}
	// One link per yrkande × förslagspunkt, and the row says how it was made.
	var n int
	if err := h.DB.QueryRow(ctx, `SELECT count(*) FROM yrkande_forslagspunkt WHERE linked_by NOT IN ('behandlas_i_punkt','forslag_text')`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("linked_by: %d %v", n, err)
	}
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO yrkande_forslagspunkt (motion_dok_id, yrkande_nummer, riksmote, beteckning, punkt, linked_by)
		VALUES ('H9023638', 9, '2021/22', 'SkU16', '6', 'forslag_text')`); err == nil {
		t.Error("a second link for the same yrkande and punkt must be rejected")
	}
}
