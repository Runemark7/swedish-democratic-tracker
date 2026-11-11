package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// fixtures/beslut-list: Riksdagen's betänkande listing for 2026/27 recorded on
// 2026-11-11 and reduced to seven rows (see its note), the UbU30 row of 2025/26, plus the one-votering
// recording of 2025/26:UbU30.

type beslutRow struct {
	Riksmote   string  `json:"riksmote"`
	Beteckning string  `json:"beteckning"`
	Title      string  `json:"title"`
	Utskott    string  `json:"utskott"`
	Date       *string `json:"date"`
	Decided    bool    `json:"decided"`
	Outcome    string  `json:"outcome"`
	Parties    []struct {
		Party      string `json:"party"`
		Partilinje string `json:"partilinje"`
	} `json:"parties"`
}

type beslutBody struct {
	Data       []beslutRow `json:"data"`
	NextCursor *string     `json:"nextCursor"`
}

func startBeslut(t *testing.T) *Harness {
	t.Helper()
	h := Start(t, chainDir(t, "beslut-list"))
	ctx := context.Background()
	if _, err := h.DB.Exec(ctx, `DELETE FROM mandate_periods`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
		VALUES ('test', 'test', '2022-09-26', '2026-09-13', '{2025/26,2026/27}')`); err != nil {
		t.Fatal(err)
	}
	// The recorded votering belongs to 2025/26; read it while that was current.
	h.App.SetNow(func() time.Time { return time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC) })
	h.RunWorker(ctx, "fetch-politicians")
	h.RunWorker(ctx, "fetch-votes")
	h.App.SetNow(func() time.Time { return time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC) })
	h.RunWorker(ctx, "sync-beslut-list")
	return h
}

func getBeslut(t *testing.T, h *Harness, query string) (beslutBody, int) {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/beslut?"+query, nil))
	var b beslutBody
	_ = json.Unmarshal(rec.Body.Bytes(), &b)
	return b, rec.Code
}

func keys(rows []beslutRow) string {
	var s []string
	for _, r := range rows {
		s = append(s, r.Beteckning)
	}
	return strings.Join(s, ",")
}

// A planned betänkande is never listed as decided, even when Riksdagen's
// status is Webbpublicering and its beslutsdag is today (KU9: beslutad 0).
func TestBeslutListNeverListsAPlannedBetankandeAsDecided(t *testing.T) {
	h := startBeslut(t)
	dec, code := getBeslut(t, h, "state=decided&limit=100")
	if code != 200 {
		t.Fatalf("status %d", code)
	}
	if got := keys(dec.Data); got != "KU4,KU3,JuU2,UbU30" && got != "KU3,KU4,JuU2,UbU30" {
		t.Errorf("decided = %s", got)
	}
	for _, r := range dec.Data {
		if !r.Decided || r.Outcome == "Ej avgjord än" {
			t.Errorf("%s decided=%v outcome=%q", r.Beteckning, r.Decided, r.Outcome)
		}
	}
	up, _ := getBeslut(t, h, "state=upcoming&limit=100")
	if len(up.Data) != 4 {
		t.Fatalf("upcoming = %s", keys(up.Data))
	}
	var first string
	for i, r := range up.Data {
		if r.Decided || r.Outcome != "Ej avgjord än" || len(r.Parties) != 0 {
			t.Errorf("%s decided=%v outcome=%q parties=%v", r.Beteckning, r.Decided, r.Outcome, r.Parties)
		}
		if i == 0 {
			first = r.Beteckning
		}
	}
	// The soonest planned day comes first: KU9 is today.
	if first != "KU9" {
		t.Errorf("first upcoming = %s, want KU9", first)
	}
	for _, r := range up.Data {
		if r.Beteckning == "SkU8" && (r.Date == nil || *r.Date != "2026-11-25") {
			t.Errorf("SkU8 date = %v", r.Date)
		}
	}
}

// The SQL used for paging and Decided() must agree: the service logs an error
// when they differ, so assert on the data instead of the log: both rows in
// a state filter and the unfiltered list agree.
func TestBeslutListUpcomingFirstThenDecidedAndPagesByCursor(t *testing.T) {
	h := startBeslut(t)
	all, _ := getBeslut(t, h, "limit=100")
	if len(all.Data) != 8 || all.NextCursor != nil {
		t.Fatalf("all = %s next=%v", keys(all.Data), all.NextCursor)
	}
	seenDecided := false
	for _, r := range all.Data {
		if r.Decided {
			seenDecided = true
		} else if seenDecided {
			t.Errorf("upcoming %s after a decided row", r.Beteckning)
		}
	}
	// Walk the list two rows at a time: same rows, same order, no repeats.
	var walked []beslutRow
	q := "limit=2"
	for i := 0; i < 10; i++ {
		rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/beslut?"+q, nil))
		if rec.Code != 200 {
			t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
		}
		var page beslutBody
		_ = json.Unmarshal(rec.Body.Bytes(), &page)
		walked = append(walked, page.Data...)
		if page.NextCursor == nil {
			break
		}
		q = "limit=2&cursor=" + url.QueryEscape(*page.NextCursor)
	}
	if keys(walked) != keys(all.Data) {
		t.Errorf("walked %s, want %s", keys(walked), keys(all.Data))
	}
	if _, code := getBeslut(t, h, "cursor=garbage"); code != 400 {
		t.Errorf("bad cursor: %d", code)
	}
	if _, code := getBeslut(t, h, "state=planned"); code != 400 {
		t.Errorf("bad state: %d", code)
	}
}

func TestBeslutListFiltersByUtskottAndParty(t *testing.T) {
	h := startBeslut(t)
	ku, _ := getBeslut(t, h, "utskott=KU&limit=100")
	if len(ku.Data) != 3 {
		t.Errorf("KU = %s", keys(ku.Data))
	}
	for _, r := range ku.Data {
		if r.Utskott != "KU" {
			t.Errorf("%s utskott %q", r.Beteckning, r.Utskott)
		}
	}
	both, _ := getBeslut(t, h, "utskott=KU,SkU&state=upcoming&limit=100")
	if len(both.Data) != 3 { // KU9, SkU8, SkU1
		t.Errorf("KU,SkU upcoming = %s", keys(both.Data))
	}

	// Party: a reservation of V in KU4 puts KU4 under V and nowhere else.
	ctx := context.Background()
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO betankanden (riksmote, beteckning, dok_id, found) VALUES ('2026/27', 'KU4', 'x', true)`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO betankande_statements (riksmote, beteckning, kind, nummer, parties)
		VALUES ('2026/27', 'KU4', 'reservation', 1, '{V}')`); err != nil {
		t.Fatal(err)
	}
	v, _ := getBeslut(t, h, "party=V&limit=100")
	if keys(v.Data) != "KU4" {
		t.Errorf("party V = %s", keys(v.Data))
	}
	if m, _ := getBeslut(t, h, "party=M&limit=100"); len(m.Data) != 0 {
		t.Errorf("party M = %s", keys(m.Data))
	}
}

func TestBeslutListSyncIsIdempotent(t *testing.T) {
	h := startBeslut(t)
	h.RunWorker(context.Background(), "sync-beslut-list")
	var n int
	if err := h.DB.QueryRow(context.Background(), `SELECT count(*) FROM beslut_listing`).Scan(&n); err != nil || n != 8 {
		t.Errorf("rows = %d err=%v", n, err)
	}
}

// The stored-column SQL that pages the list and RiksdagDocument.Decided must
// give the same answer for the cases the Decided unit test covers.
func TestBeslutListSQLDecidedRuleMatchesDecided(t *testing.T) {
	h := startBeslut(t)
	ctx := context.Background()
	cases := []struct {
		bet, status, day string
		flag             *bool
		want             bool
	}{
		{"T1", "planerat", "2026-12-09", ptrBool(false), false},
		{"T2", "Webbpublicering", "2026-11-04", ptrBool(false), false},
		{"T3", "Webbpublicering", "2026-10-21", ptrBool(true), true},
		{"T4", "planerat", "2026-12-09", nil, false},
		{"T5", "Webbpublicering", "2026-12-09", nil, false},
		{"T6", "Webbpublicering", "2026-11-11", nil, true},
		{"T7", "Webbpublicering", "", nil, false},
		{"T8", "Planerat", "2026-10-01", nil, false},
	}
	for _, c := range cases {
		var day any
		if c.day != "" {
			day = c.day
		}
		if _, err := h.DB.Exec(ctx, `
			INSERT INTO beslut_listing (riksmote, beteckning, dok_id, organ, status, beslutsdag, beslutad, sort_date)
			VALUES ('2026/27', $1, $1, 'ZZ', $2, $3::date, $4, '2026-11-01')`, c.bet, c.status, day, c.flag); err != nil {
			t.Fatal(err)
		}
	}
	for _, state := range []string{"decided", "upcoming"} {
		b, _ := getBeslut(t, h, "utskott=ZZ&limit=100&state="+state)
		got := map[string]bool{}
		for _, r := range b.Data {
			got[r.Beteckning] = true
			if r.Decided != (state == "decided") {
				t.Errorf("%s listed under %s but decided=%v", r.Beteckning, state, r.Decided)
			}
		}
		for _, c := range cases {
			if got[c.bet] != (c.want == (state == "decided")) {
				t.Errorf("%s (%s %s %v) under %s = %v", c.bet, c.status, c.day, c.flag, state, got[c.bet])
			}
		}
	}
}

func ptrBool(v bool) *bool { return &v }

// UbU30 is decided and its votering is stored: each party shows its Partilinje
// for it (S, MP, V Nej; C, KD, L, M, SD Ja), with the page of the same
// betänkande as the one shared source of the rule.
func TestBeslutListRowCarriesEachPartysPartilinje(t *testing.T) {
	h := startBeslut(t)
	b, _ := getBeslut(t, h, "utskott=UbU&limit=100")
	if len(b.Data) != 1 {
		t.Fatalf("UbU = %s", keys(b.Data))
	}
	r := b.Data[0]
	if !r.Decided || r.Outcome != "Avgjord genom omröstning" || r.Date == nil || *r.Date != "2026-08-13" {
		t.Errorf("row = %+v date=%v", r, r.Date)
	}
	want := map[string]string{"C": "Ja", "KD": "Ja", "L": "Ja", "M": "Ja", "SD": "Ja", "S": "Nej", "MP": "Nej", "V": "Nej"}
	if len(r.Parties) != 8 {
		t.Fatalf("parties = %+v", r.Parties)
	}
	for _, p := range r.Parties {
		if want[p.Party] != p.Partilinje {
			t.Errorf("%s = %q, want %q", p.Party, p.Partilinje, want[p.Party])
		}
	}
}
