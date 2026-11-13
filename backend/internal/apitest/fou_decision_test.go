package apitest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"
)

// fixtures/fou-decision: Försvarsutskottet's 2026/27:FöU2 recorded on
// 2026-11-13: its listing row, its one votering (förslagspunkt 2, three
// ballots per party) and a five-row betänkande listing of 2026/27 holding
// FöU2, FöU3, UFöU1, KU4 and SfU5.
//
// Gap decision-page-broken-for-fou-committee: a beteckning or utskott code
// with ö was answered 400. Gap decision-url-case-sensitive: a lower-case
// beteckning found the listing but none of the stored rows.

func startFoU(t *testing.T) *Harness {
	t.Helper()
	dir := chainDir(t, "fou-decision")
	h := Start(t, dir)
	ctx := context.Background()
	if _, err := h.DB.Exec(ctx, `DELETE FROM mandate_periods`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
		VALUES ('test', 'test', '2026-09-14', '2030-09-08', '{2026/27}')`); err != nil {
		t.Fatal(err)
	}
	seedBallotMembers(t, h, filepath.Join(dir, "votes-voteringlista-2026_27.json"))
	h.App.SetNow(func() time.Time { return time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC) })
	h.RunWorker(ctx, "fetch-votes")
	h.RunWorker(ctx, "sync-beslut-list")
	return h
}

// The page of FöU2 must open however the ö reaches the server: raw UTF-8,
// percent-encoded, and in any letter case.
func TestBetankandePageOpensForFoUInEveryEncodingAndCase(t *testing.T) {
	h := startFoU(t)
	for _, bet := range []string{"FöU2", "F%C3%B6U2", "F%c3%b6U2", "föu2", "F%C3%96U2", "FÖU2"} {
		b, code := getPage(t, h, "/api/betankanden/2026%2F27/"+bet)
		if code != 200 {
			t.Errorf("%s: status %d", bet, code)
			continue
		}
		if b.Riksmote != "2026/27" || b.Beteckning != "FöU2" || b.Title != "Personalförsörjning av det militära försvaret" {
			t.Errorf("%s: %s:%s %q", bet, b.Riksmote, b.Beteckning, b.Title)
		}
		if b.Decided == nil || !*b.Decided || b.DecisionDate != "2026-11-04" {
			t.Errorf("%s: decided %v date %q", bet, b.Decided, b.DecisionDate)
		}
		if len(b.Punkter) != 1 || b.Punkter[0].Number != "2" || b.Punkter[0].Mode != "votering" || len(b.Punkter[0].Parties) != 8 {
			t.Errorf("%s: punkter = %+v", bet, b.Punkter)
		}
	}
	if _, code := getPage(t, h, "/api/betankanden/2026%2F27/F%C3%B6U99"); code != 404 {
		t.Errorf("unknown FöU99: status %d, want 404", code)
	}
}

func TestBeslutListFiltersByAnUtskottWithDiacritics(t *testing.T) {
	h := startFoU(t)
	for _, code := range []string{"FöU", "F%C3%B6U", "föu", "FÖU"} {
		b, status := getBeslut(t, h, "utskott="+code+"&limit=100")
		if status != 200 {
			t.Errorf("utskott=%s: status %d", code, status)
			continue
		}
		if got := keys(b.Data); got != "FöU3,FöU2" {
			t.Errorf("utskott=%s: %s", code, got)
		}
		for _, r := range b.Data {
			if r.Utskott != "FöU" {
				t.Errorf("%s utskott %q", r.Beteckning, r.Utskott)
			}
		}
	}
	// The joint committee is its own code, not a FöU match.
	if b, status := getBeslut(t, h, "utskott="+url.QueryEscape("UFöU")+"&limit=100"); status != 200 || keys(b.Data) != "UFöU1" {
		t.Errorf("UFöU: %d %s", status, keys(b.Data))
	}
}

// A cursor carries the last row's beteckning, so paging must survive an ö.
func TestBeslutListPagesPastABeteckningWithDiacritics(t *testing.T) {
	h := startFoU(t)
	first, status := getBeslut(t, h, "utskott=F%C3%B6U&limit=1")
	if status != 200 || len(first.Data) != 1 || first.NextCursor == nil {
		t.Fatalf("first page: %d %s next=%v", status, keys(first.Data), first.NextCursor)
	}
	next, status := getBeslut(t, h, "utskott=F%C3%B6U&limit=1&cursor="+url.QueryEscape(*first.NextCursor))
	if status != 200 || keys(next.Data) != "FöU2" {
		t.Errorf("second page: %d %s", status, keys(next.Data))
	}
}

func TestResolveFindsABeteckningWithDiacritics(t *testing.T) {
	h := startFoU(t)
	for _, bet := range []string{"F%C3%B6U2", "f%C3%B6u2"} {
		rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/betankanden/resolve?beteckning="+bet, nil))
		if rec.Code != 200 {
			t.Errorf("%s: status %d: %s", bet, rec.Code, rec.Body.String())
		}
	}
}

// The motion list shares the utskott filter.
func TestMotionListAcceptsAnUtskottWithDiacritics(t *testing.T) {
	h := startFoU(t)
	if rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/motions?utskott=F%C3%B6U", nil)); rec.Code != 200 {
		t.Errorf("status %d: %s", rec.Code, rec.Body.String())
	}
}

// Löften are filtered by the betänkande that tests them; that key carries the ö too.
func TestLoftenListAcceptsABetankandeWithDiacritics(t *testing.T) {
	h := startFoU(t)
	if rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/loften?betankande=2026%2F27:F%C3%B6U2", nil)); rec.Code != 200 {
		t.Errorf("status %d: %s", rec.Code, rec.Body.String())
	}
}
