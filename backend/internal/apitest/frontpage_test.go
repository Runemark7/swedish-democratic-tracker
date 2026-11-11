package apitest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type frontPageBody struct {
	Decided  []beslutRow `json:"decided"`
	Upcoming []beslutRow `json:"upcoming"`
	Motions  []struct {
		listRow
		DecidedAt string `json:"decidedAt"`
	} `json:"motions"`
	Coverage struct {
		Oldest *string `json:"oldest"`
		Newest *string `json:"newest"`
		Stale  bool    `json:"stale"`
	} `json:"coverage"`
}

func getFrontPage(t *testing.T, h *Harness) frontPageBody {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/frontpage", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/frontpage = %d: %s", rec.Code, rec.Body.String())
	}
	var b frontPageBody
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	return b
}

// The feed holds the newest decided betänkanden and the next planned ones,
// and a planned betänkande (KU9: Webbpublicering, beslutsdag today, beslutad
// 0) is never among the decided.
func TestFrontPageNeverMixesUpcomingIntoDecided(t *testing.T) {
	h := startBeslut(t)
	b := getFrontPage(t, h)
	if len(b.Decided) != 4 || len(b.Decided) > 5 {
		t.Fatalf("decided = %s", keys(b.Decided))
	}
	if len(b.Upcoming) != 3 {
		t.Fatalf("upcoming = %s", keys(b.Upcoming))
	}
	if b.Upcoming[0].Beteckning != "KU9" {
		t.Errorf("first upcoming = %s, want KU9", b.Upcoming[0].Beteckning)
	}
	seen := map[string]bool{}
	for _, r := range b.Decided {
		seen[r.Riksmote+":"+r.Beteckning] = true
		if !r.Decided || r.Outcome == "Ej avgjord än" {
			t.Errorf("decided row %s: decided=%v outcome=%q", r.Beteckning, r.Decided, r.Outcome)
		}
	}
	for _, r := range b.Upcoming {
		if r.Decided || seen[r.Riksmote+":"+r.Beteckning] {
			t.Errorf("upcoming row %s is decided or also listed as decided", r.Beteckning)
		}
		if len(r.Parties) != 0 {
			t.Errorf("upcoming %s carries a Partilinje", r.Beteckning)
		}
	}
	// UbU30 is decided with a votering: all eight parties are marked.
	for _, r := range b.Decided {
		if r.Beteckning == "UbU30" && len(r.Parties) != 8 {
			t.Errorf("UbU30 parties = %d", len(r.Parties))
		}
	}
	if b.Motions == nil {
		t.Error("motions must be an empty list, not null")
	}
}

// The coverage sentence is driven by `stale`: the newest decision is older
// than the current week.
func TestFrontPageCoverageIsStaleWhenNewestDecisionIsBeforeThisWeek(t *testing.T) {
	h := startBeslut(t) // now = Wednesday 2026-11-11
	b := getFrontPage(t, h)
	if b.Coverage.Newest == nil || b.Coverage.Oldest == nil {
		t.Fatalf("coverage = %+v", b.Coverage)
	}
	var newest string
	for _, r := range b.Decided {
		if r.Date != nil && *r.Date > newest {
			newest = *r.Date
		}
	}
	if *b.Coverage.Newest != newest {
		t.Errorf("newest = %s, want %s", *b.Coverage.Newest, newest)
	}
	wantStale := newest < "2026-11-09"
	if b.Coverage.Stale != wantStale {
		t.Errorf("stale = %v for newest %s", b.Coverage.Stale, newest)
	}
	// Move the clock far ahead: the same decisions are now certainly older
	// than the current week.
	h.App.SetNow(func() time.Time { return time.Date(2027, 3, 3, 12, 0, 0, 0, time.UTC) })
	if later := getFrontPage(t, h); !later.Coverage.Stale {
		t.Errorf("not stale in March 2027: %+v", later.Coverage)
	}
	// And back in the week of the newest decision it is not stale.
	nd, _ := time.Parse("2006-01-02", newest)
	h.App.SetNow(func() time.Time { return nd.Add(12 * time.Hour) })
	if cur := getFrontPage(t, h); cur.Coverage.Stale {
		t.Errorf("stale in the week of the newest decision: %+v", cur.Coverage)
	}
}

// Recently decided motions come from the stored decision chain: decided
// only, each with party, kind, outcome and the betänkande that decided it.
func TestFrontPageListsRecentlyDecidedMotions(t *testing.T) {
	h := listHarness(t)
	b := getFrontPage(t, h)
	if len(b.Motions) == 0 || len(b.Motions) > 5 {
		t.Fatalf("motions = %d", len(b.Motions))
	}
	for _, m := range b.Motions {
		if m.Outcome.Word == "" || m.Outcome.Word == "Ej avgjord än" || m.Kind == "" || len(m.Parties) == 0 || len(m.HandledIn) == 0 || m.DecidedAt == "" {
			t.Errorf("motion %s:%s = %+v", m.Riksmote, m.Beteckning, m)
		}
	}
	// Newest decision first.
	for i := 1; i < len(b.Motions); i++ {
		if b.Motions[i].DecidedAt > b.Motions[i-1].DecidedAt {
			t.Errorf("motions not newest first: %s after %s", b.Motions[i].DecidedAt, b.Motions[i-1].DecidedAt)
		}
	}
}
