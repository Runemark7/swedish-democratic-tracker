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

// Gap recent-decisions-missing-proposal-text (#267). 2026/27:KU9 is a vilande
// grundlagsförslag: decided by votering on 2026-11-11, and no motion points at
// it. The betänkanden worker used to read only what stored motions point at,
// so its page had the ballots but no förslagstext, winner or reservation.
//
// fixtures/betankande-no-motions, recorded 2026-11-14: the 2026/27 listing
// reduced to KU9 (decided), KrU2 (beslutsdag ahead, beslutad 0) and NU4
// (planerat); KU9's own listing row, dokumentstatus and HTML, unchanged.
// Nothing is recorded for KrU2 and NU4, so reading them would fail the test.

type noMotionsPage struct {
	Held    bool `json:"held"`
	Punkter []struct {
		Number                string `json:"number"`
		Rubrik                string `json:"rubrik"`
		Text                  string `json:"text"`
		Vinnare               string `json:"vinnare"`
		ReservationsAvailable bool   `json:"reservationsAvailable"`
		Reservations          []struct {
			Number  int      `json:"number"`
			Parties []string `json:"parties"`
		} `json:"reservations"`
	} `json:"punkter"`
}

func startNoMotions(t *testing.T) *Harness {
	t.Helper()
	h := Start(t, chainDir(t, "betankande-no-motions"))
	ctx := context.Background()
	if _, err := h.DB.Exec(ctx, `DELETE FROM mandate_periods`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
		VALUES ('test', 'test', '2026-09-13', '2030-09-08', '{2026/27}')`); err != nil {
		t.Fatal(err)
	}
	h.App.SetNow(func() time.Time { return time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC) })
	h.RunWorker(ctx, "sync-beslut-list")
	return h
}

func getNoMotionsPage(t *testing.T, h *Harness, bet string) noMotionsPage {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/betankanden/2026%2F27/"+bet, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d: %s", bet, rec.Code, rec.Body.String())
	}
	var p noMotionsPage
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDecidedBetankandeNoMotionPointsAtGetsItsForslagstextAndReservation(t *testing.T) {
	h := startNoMotions(t)
	ctx := context.Background()
	h.RunWorker(ctx, "fetch-betankanden")

	p := getNoMotionsPage(t, h, "KU9")
	if !p.Held || len(p.Punkter) != 1 {
		t.Fatalf("held = %v, punkter = %d; want the stored förslagspunkt", p.Held, len(p.Punkter))
	}
	k := p.Punkter[0]
	if k.Number != "1" || k.Rubrik != "Vilande förslag" {
		t.Errorf("punkt = %q rubrik = %q", k.Number, k.Rubrik)
	}
	if !strings.Contains(k.Text, "Riksdagen antar slutligt det vilande förslaget") {
		t.Errorf("text = %q", k.Text)
	}
	if k.Vinnare != "utskottet" {
		t.Errorf("vinnare = %q", k.Vinnare)
	}
	if !k.ReservationsAvailable || len(k.Reservations) != 1 || strings.Join(k.Reservations[0].Parties, ",") != "V,MP" {
		t.Errorf("reservations (available %v) = %+v, want one by V and MP", k.ReservationsAvailable, k.Reservations)
	}

	// Not decided, so not read: nothing is recorded for them, and they stay
	// without punkter.
	for _, bet := range []string{"KrU2", "NU4"} {
		var n int
		if err := h.DB.QueryRow(ctx, `SELECT count(*) FROM betankanden WHERE riksmote = '2026/27' AND beteckning = $1`, bet).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Errorf("%s is not decided but was read", bet)
		}
	}
}

func TestDecidedBetankandeNoMotionPointsAtIsReadOnce(t *testing.T) {
	h := startNoMotions(t)
	ctx := context.Background()
	h.RunWorker(ctx, "fetch-betankanden")
	h.RunWorker(ctx, "fetch-betankanden")
	if n := h.Served("/dokumentstatus/HE01KU9.json"); n != 1 {
		t.Errorf("dokumentstatus read %d times over two runs, want 1", n)
	}
}

// A betänkande read on the morning of its beslutsdag has no winner yet. It is
// read again once that day has passed.
func TestBetankandeReadBeforeItsDecisionDayWasOverIsReadAgain(t *testing.T) {
	h := startNoMotions(t)
	ctx := context.Background()
	h.RunWorker(ctx, "fetch-betankanden")
	if _, err := h.DB.Exec(ctx, `
		UPDATE betankanden SET fetched_at = '2026-11-11 05:00:00+00' WHERE riksmote = '2026/27' AND beteckning = 'KU9'`); err != nil {
		t.Fatal(err)
	}
	h.RunWorker(ctx, "fetch-betankanden")
	if n := h.Served("/dokumentstatus/HE01KU9.json"); n != 2 {
		t.Errorf("dokumentstatus read %d times, want 2", n)
	}
	h.RunWorker(ctx, "fetch-betankanden")
	if n := h.Served("/dokumentstatus/HE01KU9.json"); n != 2 {
		t.Errorf("dokumentstatus read %d times after the re-read, want 2", n)
	}
}
