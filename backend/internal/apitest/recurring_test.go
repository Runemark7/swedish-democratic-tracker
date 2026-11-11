package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"riksdagskollen/internal/motions"
	motionsPG "riksdagskollen/internal/motions/adapters/postgres"
)

// Fixtures in fixtures/recurring-krav were recorded from data.riksdagen.se on
// 2026-11-11: M's virtual-goods demand in 2017/18:3293 y1, 2018/19:2900 y26,
// 2019/20:2718 y26 and 2021/22:3638 y9, and Rickard Nordin's (C) enskilda
// motioner 2022/23:1675, 2023/24:1796 and 2024/25:989.

type recurringBody struct {
	Yrkanden []struct {
		Number    int `json:"number"`
		Recurring *struct {
			Rule      string `json:"rule"`
			RuleLabel string `json:"ruleLabel"`
			Members   []struct {
				Riksmote   string   `json:"riksmote"`
				Beteckning string   `json:"beteckning"`
				Number     int      `json:"number"`
				Parties    []string `json:"parties"`
				Outcome    string   `json:"outcome"`
				Rule       string   `json:"rule"`
				Current    bool     `json:"current"`
			} `json:"members"`
		} `json:"recurring"`
	} `json:"yrkanden"`
}

func startRecurring(t *testing.T) *Harness {
	t.Helper()
	h := Start(t, chainDir(t, "recurring-krav"), WithNow(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)))
	ctx := context.Background()
	if _, err := h.DB.Exec(ctx, `DELETE FROM mandate_periods`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
		VALUES ('test', 'test', '2014-09-14', '2026-09-13', '{2017/18,2018/19,2019/20,2021/22,2022/23,2023/24,2024/25}')`); err != nil {
		t.Fatal(err)
	}
	h.RunWorker(ctx, "fetch-motions")
	return h
}

func getRecurring(t *testing.T, h *Harness, rm, bet string) recurringBody {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/motions/"+rm+"/"+bet, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s/%s = %d: %s", rm, bet, rec.Code, rec.Body.String())
	}
	var b recurringBody
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	return b
}

func TestMsVirtualGoodsDemandIsOneAterkommandeKrav(t *testing.T) {
	h := startRecurring(t)
	want := []struct {
		rm, bet string
		n       int
	}{{"2017/18", "3293", 1}, {"2018/19", "2900", 26}, {"2019/20", "2718", 26}, {"2021/22", "3638", 9}}

	// Every member sees the same group, earliest first, with itself marked.
	for i, from := range want {
		enc := map[string]string{"2017/18": "2017%2F18", "2018/19": "2018%2F19", "2019/20": "2019%2F20", "2021/22": "2021%2F22"}[from.rm]
		b := getRecurring(t, h, enc, from.bet)
		var rec = b.Yrkanden[0].Recurring
		for _, y := range b.Yrkanden {
			if y.Number == from.n {
				rec = y.Recurring
			}
		}
		if rec == nil {
			t.Fatalf("%s:%s y%d has no Återkommande krav", from.rm, from.bet, from.n)
		}
		if rec.Rule != "identical_wording" || rec.RuleLabel == "" {
			t.Errorf("rule = %q / %q", rec.Rule, rec.RuleLabel)
		}
		if len(rec.Members) != len(want) {
			t.Fatalf("%s:%s members = %+v, want %d", from.rm, from.bet, rec.Members, len(want))
		}
		for j, m := range rec.Members {
			w := want[j]
			if m.Riksmote != w.rm || m.Beteckning != w.bet || m.Number != w.n {
				t.Errorf("member %d = %+v, want %+v", j, m, w)
			}
			if m.Rule != "identical_wording" {
				t.Errorf("member %d rule = %q", j, m.Rule)
			}
			if len(m.Parties) != 1 || m.Parties[0] != "M" {
				t.Errorf("member %d parties = %v", j, m.Parties)
			}
			if m.Current != (j == i) {
				t.Errorf("member %d current = %v", j, m.Current)
			}
		}
	}
}

func TestNordinsCYrkandenAreNotGroupedWithMs(t *testing.T) {
	h := startRecurring(t)
	for _, c := range []struct{ rm, enc, bet string }{
		{"2022/23", "2022%2F23", "1675"}, {"2023/24", "2023%2F24", "1796"}, {"2024/25", "2024%2F25", "989"},
	} {
		b := getRecurring(t, h, c.enc, c.bet)
		rec := b.Yrkanden[0].Recurring
		if rec == nil {
			t.Fatalf("%s:%s: Nordin's identical filings should group with each other", c.rm, c.bet)
		}
		if len(rec.Members) != 3 {
			t.Errorf("%s:%s members = %+v", c.rm, c.bet, rec.Members)
		}
		for _, m := range rec.Members {
			if len(m.Parties) != 1 || m.Parties[0] != "C" {
				t.Errorf("%s:%s grouped with a non-C filing: %+v", c.rm, c.bet, m)
			}
		}
	}
}

func TestRecurringHashIsHealedOnReadAndBackfilled(t *testing.T) {
	h := startRecurring(t)
	ctx := context.Background()
	// Simulate rows stored before the hash existed.
	if _, err := h.DB.Exec(ctx, `UPDATE motion_yrkanden SET wording_hash = NULL`); err != nil {
		t.Fatal(err)
	}
	// Viewing one heals that yrkande only; the others are found by the backfill.
	getRecurring(t, h, "2017%2F18", "3293")
	var n int
	if err := h.DB.QueryRow(ctx, `SELECT count(*) FROM motion_yrkanden WHERE wording_hash IS NOT NULL`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("hashed after one read = %d, %v", n, err)
	}
	svc := motions.NewService(motionsPG.NewRepository(h.DB), nil)
	got, err := svc.BackfillWordingHashes(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got == 0 {
		t.Fatal("backfill wrote nothing")
	}
	if again, err := svc.BackfillWordingHashes(ctx, 2); err != nil || again != 0 {
		t.Fatalf("second backfill = %d, %v; want 0", again, err)
	}
	b := getRecurring(t, h, "2017%2F18", "3293")
	if r := b.Yrkanden[0].Recurring; r == nil || len(r.Members) != 4 {
		t.Fatalf("after backfill: %+v", r)
	}
}
