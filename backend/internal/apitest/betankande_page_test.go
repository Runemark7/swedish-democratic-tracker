package apitest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Fixtures: fixtures/decision-chain and fixtures/proposed-by (see their tests)
// and fixtures/betankande-planned, the SkU8 listing of 2026/27 recorded on
// 2026-11-11 while the betänkande was still planerat.

type pageBody struct {
	Riksmote     string `json:"riksmote"`
	Beteckning   string `json:"beteckning"`
	Title        string `json:"title"`
	Decided      *bool  `json:"decided"`
	DecisionDate string `json:"decisionDate"`
	DecisionText string `json:"decisionText"`
	Punkter      []struct {
		Number  string `json:"number"`
		Mode    string `json:"mode"`
		Parties []struct {
			Party      string `json:"party"`
			Partilinje string `json:"partilinje"`
		} `json:"parties"`
		Avvikare []struct {
			PoliticianID string `json:"politicianId"`
			Name         string `json:"name"`
			Party        string `json:"party"`
		} `json:"avvikare"`
		Reservations []struct {
			Number int `json:"number"`
		} `json:"reservations"`
		Reasoning string `json:"reasoning"`
		Motions   []struct {
			Riksmote   string   `json:"riksmote"`
			Beteckning string   `json:"beteckning"`
			Parties    []string `json:"parties"`
			Yrkanden   []int    `json:"yrkanden"`
		} `json:"motions"`
	} `json:"punkter"`
}

func getPage(t *testing.T, h *Harness, path string) (pageBody, int) {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, path, nil))
	var b pageBody
	_ = json.Unmarshal(rec.Body.Bytes(), &b)
	return b, rec.Code
}

func TestMotionsDecidedInAForslagspunktReturns3293ForSkU11Punkt6(t *testing.T) {
	h := startChain(t, chainDir(t, "decision-chain"), "{2017/18,2018/19,2021/22}")
	runChain(h)
	b, code := getPage(t, h, "/api/betankanden/2017%2F18/SkU11")
	if code != 200 {
		t.Fatalf("status %d", code)
	}
	for _, p := range b.Punkter {
		if p.Number != "6" {
			continue
		}
		if len(p.Motions) != 1 || p.Motions[0].Riksmote != "2017/18" || p.Motions[0].Beteckning != "3293" || len(p.Motions[0].Yrkanden) != 1 || p.Motions[0].Yrkanden[0] != 1 {
			t.Errorf("motions = %+v", p.Motions)
		}
		if p.Mode != "acklamation" {
			t.Errorf("mode = %q", p.Mode)
		}
		if len(p.Reservations) == 0 || p.Reasoning == "" {
			t.Errorf("reservations/reasoning missing: %+v", p)
		}
		if b.Decided == nil || !*b.Decided || b.DecisionDate != "2018-03-28" {
			t.Errorf("decided = %v date = %q", b.Decided, b.DecisionDate)
		}
		return
	}
	t.Fatal("punkt 6 not on the page")
}

func TestSameBeteckningInTwoRiksmotenIsTwoPagesWithAvvikareByName(t *testing.T) {
	h := startProposedBy(t)
	old, code := getPage(t, h, "/api/betankanden/2021%2F22/SkU16")
	if code != 200 {
		t.Fatalf("2021/22: %d", code)
	}
	cur, code := getPage(t, h, "/api/betankanden/2025%2F26/SkU16")
	if code != 200 {
		t.Fatalf("2025/26: %d", code)
	}
	if old.Riksmote != "2021/22" || cur.Riksmote != "2025/26" {
		t.Fatalf("riksmöten %q %q", old.Riksmote, cur.Riksmote)
	}
	ids := func(b pageBody) string {
		var s []string
		for _, p := range b.Punkter {
			for _, m := range p.Motions {
				s = append(s, m.Riksmote+":"+m.Beteckning)
			}
		}
		return strings.Join(s, ",")
	}
	if !strings.Contains(ids(old), "2021/22:3638") || strings.Contains(ids(old), "2025/26") {
		t.Errorf("2021/22 motions = %s", ids(old))
	}
	if ids(cur) != "2025/26:2783" {
		t.Errorf("2025/26 motions = %s", ids(cur))
	}
	for _, p := range old.Punkter {
		for _, a := range p.Avvikare {
			if a.Name == "" || a.Party == "" || a.PoliticianID == "" {
				t.Errorf("avvikare lacks name/party/id: %+v", a)
			}
		}
	}
}

func TestPlannedBetankandeIsEjAvgjortAnWithPlannedDate(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	h := Start(t, filepath.Join(filepath.Dir(file), "fixtures", "betankande-planned"))
	h.App.SetNow(func() time.Time { return time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC) })
	b, code := getPage(t, h, "/api/betankanden/2026%2F27/SkU8")
	if code != 200 {
		t.Fatalf("status %d", code)
	}
	if b.Decided == nil || *b.Decided {
		t.Errorf("decided = %v", b.Decided)
	}
	if b.DecisionDate != "2026-11-25" || b.DecisionText != "Ej avgjort än" {
		t.Errorf("date %q text %q", b.DecisionDate, b.DecisionText)
	}
}

func TestResolveRedirectTargets(t *testing.T) {
	h := startProposedBy(t)
	type res struct{ Riksmote, Beteckning, Punkt string }
	get := func(q string) (res, int) {
		rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/betankanden/resolve?"+q, nil))
		var r res
		_ = json.Unmarshal(rec.Body.Bytes(), &r)
		return r, rec.Code
	}
	// No riksmöte: the most recent one with that beteckning.
	if r, c := get("beteckning=SkU16"); c != 200 || r.Riksmote != "2025/26" {
		t.Errorf("no rm: %d %+v", c, r)
	}
	// A votering URL: the riksmöte the votering belongs to, with its punkt.
	if r, c := get("beteckning=SkU16&punkt=6&rm=2021/22"); c != 200 || r.Riksmote != "2021/22" || r.Punkt != "6" {
		t.Errorf("rm given: %d %+v", c, r)
	}
	if _, c := get("beteckning=NoSuch1"); c != 404 {
		t.Errorf("unknown: %d", c)
	}
}
