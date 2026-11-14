package apitest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Seam tests for Löftesstatus (issue #161): confirmations and undo go through
// the real write API; the public read computes the status from Postgres.

type statusView struct {
	ID              int     `json:"id"`
	Status          *string `json:"status"`
	StatusChangedAt *string `json:"statusChangedAt"`
	StatusReason    string  `json:"statusReason"`
	Kopplingar      []struct {
		ID          int    `json:"id"`
		URL         string `json:"url"`
		Effect      string `json:"effect"`
		Partilinje  string `json:"partilinje"`
		Explanation string `json:"explanation"`
		ConfirmedAt string `json:"confirmedAt"`
		Direction   string `json:"direction"`
	} `json:"kopplingar"`
}

func publicGet(t *testing.T, h *Harness, path string, want int) *httptest.ResponseRecorder {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != want {
		t.Fatalf("GET %s = %d %s (want %d)", path, rec.Code, rec.Body, want)
	}
	return rec
}

func statusOf(t *testing.T, h *Harness, id int64) statusView {
	t.Helper()
	return decode[statusView](t, publicGet(t, h, fmt.Sprintf("/api/loften/%d", id), 200))
}

func ballotsFor(t *testing.T, h *Harness, votering int64, party string, ja, nej, avst int) {
	t.Helper()
	n := 0
	add := func(res string, count int) {
		for i := 0; i < count; i++ {
			n++
			pid := fmt.Sprintf("ls-%d-%s-%d", votering, party, n)
			if _, err := h.DB.Exec(context.Background(),
				`INSERT INTO politicians (intressent_id, first_name, last_name, party) VALUES ($1,'A','B',$2)`, pid, party); err != nil {
				t.Fatal(err)
			}
			if _, err := h.DB.Exec(context.Background(),
				`INSERT INTO ballots (votering_ref, politician_id, party, vote_result) VALUES ($1,$2,$3,$4)`,
				votering, pid, party, res); err != nil {
				t.Fatal(err)
			}
		}
	}
	add("Ja", ja)
	add("Nej", nej)
	add("Avstår", avst)
}

func confirmKoppling(t *testing.T, s *kSeed, kid int64, batchKey string) (entry int64) {
	t.Helper()
	b := createBatch(t, s, batchKey)
	nr := 0
	for _, it := range b.Items {
		if it.KopplingID == kid {
			nr = it.Nr
		}
	}
	r := answer(t, s, b.ID, batchKey+"-a", fmt.Sprintf(`{"text":"%d ja",`+vbProv+`}`, nr), 200)
	return r.Decided[0].EntryID
}

func TestConfirmingAndUndoingAKopplingChangesTheLoeftesstatus(t *testing.T) {
	s := startK(t)
	// S voted Ja (90 of 100 cast). The Löfte's direction is Ja.
	k := addKoppling(t, s, "x", 0.9, "Ja")
	var vref int64
	if err := s.h.DB.QueryRow(context.Background(), `SELECT votering_ref FROM kopplingar WHERE id=$1`, k).Scan(&vref); err != nil {
		t.Fatal(err)
	}
	ballotsFor(t, s.h, vref, "S", 9, 1, 0)

	// A suggestion is not a Bekräftad koppling: no status effect, and not public.
	v := statusOf(t, s.h, s.loefteSkU)
	if v.Status == nil || *v.Status != "Inte prövat än" || len(v.Kopplingar) != 0 {
		t.Fatalf("before confirm: %+v", v)
	}
	publicGet(t, s.h, fmt.Sprintf("/api/kopplingar/%d", k), 404)

	entry := confirmKoppling(t, s, k, "ls1")
	v = statusOf(t, s.h, s.loefteSkU)
	if *v.Status != "Röstat för" || v.StatusChangedAt == nil || v.StatusReason == "" || len(v.Kopplingar) != 1 {
		t.Fatalf("after confirm: %+v", v)
	}
	kv := v.Kopplingar[0]
	if kv.Effect != "for" || kv.Partilinje != "Ja" || kv.ConfirmedAt == "" || kv.Explanation == "" || kv.URL != fmt.Sprintf("/kopplingar/%d", k) {
		t.Fatalf("koppling view: %+v", kv)
	}
	// The koppling has its own stable public URL.
	publicGet(t, s.h, fmt.Sprintf("/api/kopplingar/%d", k), 200)
	publicGet(t, s.h, fmt.Sprintf("/api/loften/%d/kopplingar", s.loefteSkU), 200)

	// Filters by party, utskott and status.
	for path, want := range map[string]int{
		"/api/loften?status=for": 1, "/api/loften?status=emot": 0, "/api/loften?status=inte-provat": 1, // the AU Löfte
		"/api/loften?party=S&utskott=SkU&status=for": 1, "/api/loften?utskott=AU&status=for": 0,
	} {
		var got []struct{ ID int }
		got = decode[[]struct{ ID int }](t, publicGet(t, s.h, path, 200))
		if len(got) != want {
			t.Errorf("%s returned %d, want %d", path, len(got), want)
		}
	}
	publicGet(t, s.h, "/api/loften?status=hållet", 400)
	sum := decode[[]struct {
		Party      string
		Total, For int
	}](t, publicGet(t, s.h, "/api/loften/summary", 200))
	for _, p := range sum {
		if p.Party == "S" && p.For != 1 || p.Party != "S" && p.For != 0 {
			t.Fatalf("summary %+v", sum)
		}
	}

	// Undo: back to Inte prövat än, and the koppling is no longer public.
	rec := writeReq(s.h, "POST", fmt.Sprintf("/api/write/aendringslogg/%d/undo", entry), testWriteToken, "ls-undo", `{`+vbProv+`}`)
	if rec.Code != 200 {
		t.Fatalf("undo: %d %s", rec.Code, rec.Body)
	}
	v = statusOf(t, s.h, s.loefteSkU)
	if *v.Status != "Inte prövat än" || len(v.Kopplingar) != 0 || v.StatusChangedAt != nil {
		t.Fatalf("after undo: %+v", v)
	}
	publicGet(t, s.h, fmt.Sprintf("/api/kopplingar/%d", k), 404)
}

func TestMixedAvstarAndArchiveStatus(t *testing.T) {
	s := startK(t)
	k1 := addKoppling(t, s, "m1", 0.9, "Ja")
	k2 := addKoppling(t, s, "m2", 0.8, "Ja")
	var r1, r2 int64
	ctx := context.Background()
	_ = s.h.DB.QueryRow(ctx, `SELECT votering_ref FROM kopplingar WHERE id=$1`, k1).Scan(&r1)
	_ = s.h.DB.QueryRow(ctx, `SELECT votering_ref FROM kopplingar WHERE id=$1`, k2).Scan(&r2)
	ballotsFor(t, s.h, r1, "S", 1, 1, 8) // Avstår line: no direction
	ballotsFor(t, s.h, r2, "S", 1, 6, 0) // Nej line: against the Löfte's Ja direction
	b := createBatch(t, s, "mx")
	answer(t, s, b.ID, "mx-a", `{"text":"1,2 ja",`+vbProv+`}`, 200)
	v := statusOf(t, s.h, s.loefteSkU)
	if *v.Status != "Röstat emot" {
		t.Fatalf("status %v", *v.Status)
	}
	effects := []string{}
	for _, k := range v.Kopplingar {
		effects = append(effects, k.Effect)
	}
	if strings.Join(effects, ",") != "ingen,emot" && strings.Join(effects, ",") != "emot,ingen" {
		t.Fatalf("effects %v (Avstår is listed but not counted)", effects)
	}

	// An archived Löfte has no status without a Bekräftad koppling (#259) ...
	arch := seedLoefte(t, s.h, "S", "Gammalt", "archived", "SkU")
	if av := statusOf(t, s.h, arch); av.Status != nil {
		t.Fatalf("archived without kopplingar has status %q", *av.Status)
	}
	// ... and the same status as any Löfte once it has one.
	if _, err := s.h.DB.Exec(ctx, `UPDATE kopplingar SET loefte_id=$1 WHERE id=$2`, arch, k2); err != nil {
		t.Fatal(err)
	}
	if av := statusOf(t, s.h, arch); av.Status == nil || *av.Status != "Röstat emot" || len(av.Kopplingar) != 1 {
		t.Fatalf("archived with a Bekräftad koppling: %+v", av)
	}
}
