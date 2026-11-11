package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Fixtures: the decision-chain recordings plus the raw betänkande HTML
// (fixtures/html, recorded 2026-11-11 from data.riksdagen.se/dokument/<id>.html).

type statementBody struct {
	Number   int      `json:"number"`
	Parties  []string `json:"parties"`
	Signers  []string `json:"signers"`
	Proposal string   `json:"proposal"`
	Text     string   `json:"text"`
}

type htmlDecision struct {
	decisionBody
	Reasoning             string          `json:"reasoning"`
	ReasoningAvailable    bool            `json:"reasoningAvailable"`
	ReservationsAvailable bool            `json:"reservationsAvailable"`
	Reservations          []statementBody `json:"reservations"`
	SarskildaYttranden    []statementBody `json:"sarskildaYttranden"`
	VotedReservation      *struct {
		Number  int      `json:"number"`
		Party   string   `json:"party"`
		Parties []string `json:"parties"`
	} `json:"votedReservation"`
}

func decisionOf(t *testing.T, h *Harness, rm, bet string, yrkande int) htmlDecision {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/motions/"+rm+"/"+bet, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET motion %s/%s = %d: %s", rm, bet, rec.Code, rec.Body.String())
	}
	var m struct {
		Yrkanden []struct {
			Number    int            `json:"number"`
			Decisions []htmlDecision `json:"decisions"`
		} `json:"yrkanden"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	for _, y := range m.Yrkanden {
		if y.Number == yrkande && len(y.Decisions) == 1 {
			return y.Decisions[0]
		}
	}
	t.Fatalf("no single decision for yrkande %d in %s", yrkande, rec.Body.String())
	return htmlDecision{}
}

func partiesOf(ss []statementBody) map[int]string {
	out := map[int]string{}
	for _, s := range ss {
		out[s.Number] = strings.Join(s.Parties, ",")
	}
	return out
}

func TestMotion3293ShowsReservations8And9AndTheUtskottsReasoning(t *testing.T) {
	h := startChain(t, chainDir(t, "decision-chain"), "{2017/18,2018/19,2021/22}")
	runChain(h)
	d := decisionOf(t, h, "2017%2F18", "3293", 1)
	if d.BetankandeBeteckning != "SkU11" || d.Punkt != "6" {
		t.Fatalf("decision = %+v", d.decisionBody)
	}
	got := partiesOf(d.Reservations)
	if len(got) != 2 || got[8] != "M" || got[9] != "KD" {
		t.Errorf("reservations on SkU11 p6 = %v, want 8 (M) and 9 (KD)", got)
	}
	if !d.ReservationsAvailable {
		t.Error("reservations should be available")
	}
	if !d.ReasoningAvailable || !strings.Contains(d.Reasoning, "gör ingen annan bedömning om virtuella varor än Skatteverket") {
		t.Errorf("reasoning = %q (available %v)", d.Reasoning, d.ReasoningAvailable)
	}
	for _, r := range d.Reservations {
		if r.Number == 8 && (len(r.Signers) != 4 || !strings.Contains(r.Text, "Vi föreslår en utredning")) {
			t.Errorf("reservation 8 = %+v", r)
		}
	}
	if d.VotedReservation != nil {
		t.Errorf("an acklamation has no votered alternative, got %+v", d.VotedReservation)
	}
}

func TestMotion3638Yrkande9VoteredAlternativeWasReservation12(t *testing.T) {
	h := startChain(t, chainDir(t, "decision-chain"), "{2017/18,2018/19,2021/22}")
	runChain(h)
	d := decisionOf(t, h, "2021%2F22", "3638", 9)
	if d.BetankandeBeteckning != "SkU16" || d.Punkt != "6" {
		t.Fatalf("decision = %+v", d.decisionBody)
	}
	v := d.VotedReservation
	if v == nil || v.Number != 12 || v.Party != "M" || strings.Join(v.Parties, ",") != "M" {
		t.Fatalf("votedReservation = %+v", v)
	}
	if partiesOf(d.Reservations)[12] != "M" {
		t.Errorf("reservation 12 not listed on the punkt: %v", partiesOf(d.Reservations))
	}
	if !d.ReservationsAvailable || !d.ReasoningAvailable {
		t.Errorf("available: reservations %v reasoning %v", d.ReservationsAvailable, d.ReasoningAvailable)
	}
}

func TestMotion2900Yrkande26VoteredAlternativeWasAnotherPartysReservation(t *testing.T) {
	h := startChain(t, chainDir(t, "decision-chain"), "{2017/18,2018/19,2021/22}")
	runChain(h)
	d := decisionOf(t, h, "2018%2F19", "2900", 26)
	v := d.VotedReservation
	if v == nil || v.Number != 2 || v.Party != "SD" {
		t.Fatalf("votedReservation = %+v, want reservation 2 (SD)", v)
	}
	if v.Party == "M" {
		t.Error("the motion is M's, but the votered reservation is not M's own")
	}
	if partiesOf(d.Reservations)[2] != "SD" {
		t.Errorf("reservations on the punkt = %v", partiesOf(d.Reservations))
	}
	// M's own reservation on the punkt (if any) is listed but is not the voted one.
	for _, r := range d.Reservations {
		if r.Number != 2 && r.Number == v.Number {
			t.Errorf("duplicate voted reservation %+v", r)
		}
	}
}

func TestHTMLNotAvailableIsReportedNotInvented(t *testing.T) {
	// Riksdagen answers 404 for the betänkande HTML (a stub error response,
	// not a recording): the page must get "not available", no text.
	src := chainDir(t, "decision-chain")
	dir := t.TempDir()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(e.Name(), "html-H501SkU11") {
			b = []byte(`{"method":"GET","path":"/dokument/H501SkU11.html","query":{},"status":404,"body":{}}`)
		}
		if err := os.WriteFile(filepath.Join(dir, e.Name()), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// bodyFile paths are relative to the fixture directory.
	if err := os.MkdirAll(filepath.Join(filepath.Dir(dir), "html"), 0o755); err == nil {
		defer os.RemoveAll(filepath.Join(filepath.Dir(dir), "html"))
		for _, id := range []string{"H601SkU11", "H901SkU16", "H601CU2"} {
			b, _ := os.ReadFile(filepath.Join(src, "..", "html", id+".html"))
			_ = os.WriteFile(filepath.Join(filepath.Dir(dir), "html", id+".html"), b, 0o644)
		}
	}
	h := startChain(t, dir, "{2017/18,2018/19,2021/22}")
	runChain(h)
	d := decisionOf(t, h, "2017%2F18", "3293", 1)
	if d.ReasoningAvailable || d.Reasoning != "" || d.ReservationsAvailable || len(d.Reservations) != 0 {
		t.Errorf("invented content: %+v", d)
	}
	if d.BetankandeSourceURL != "https://data.riksdagen.se/dokument/H501SkU11.html" {
		t.Errorf("source link = %q", d.BetankandeSourceURL)
	}
	// The other betänkanden are unaffected.
	if e := decisionOf(t, h, "2021%2F22", "3638", 9); !e.ReasoningAvailable {
		t.Error("SkU16 should still have reasoning")
	}
}

func TestHTMLIngestionIsIdempotent(t *testing.T) {
	h := startChain(t, chainDir(t, "decision-chain"), "{2017/18,2018/19,2021/22}")
	runChain(h)
	first, _ := json.Marshal(decisionOf(t, h, "2017%2F18", "3293", 1))
	reads := h.Served("/dokument/H501SkU11.html")
	if reads != 1 {
		t.Fatalf("html read %d times", reads)
	}
	h.RunWorker(context.Background(), "fetch-betankanden")
	if h.Served("/dokument/H501SkU11.html") != reads {
		t.Error("an HTML already read was read again")
	}
	again, _ := json.Marshal(decisionOf(t, h, "2017%2F18", "3293", 1))
	if string(first) != string(again) {
		t.Error("re-run changed the decision")
	}
}
