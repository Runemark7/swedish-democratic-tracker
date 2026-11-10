package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

// Fixtures in fixtures/proposed-by were recorded from data.riksdagen.se on
// 2026-11-11: the motions that 2021/22:SkU16 punkt 6 (3312, 3637, 3638, 3685,
// 4217) and 2025/26:SkU16 punkt 6 (2783) decided, those two betänkanden cut to
// punkt 6, and the two voteringar of punkt 6.

type proposedByBody struct {
	Session    string `json:"session"`
	ProposedBy struct {
		Motions []struct {
			DokID      string   `json:"dokId"`
			Riksmote   string   `json:"riksmote"`
			Beteckning string   `json:"beteckning"`
			Parties    []string `json:"parties"`
			Yrkanden   []int    `json:"yrkanden"`
		} `json:"motions"`
		Propositions []struct {
			Riksmote   string `json:"riksmote"`
			Beteckning string `json:"beteckning"`
		} `json:"propositions"`
		Alternativ string `json:"alternativ"`
	} `json:"proposedBy"`
}

func startProposedBy(t *testing.T) *Harness {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(file), "fixtures", "proposed-by")
	h := Start(t, dir)
	ctx := context.Background()
	if _, err := h.DB.Exec(ctx, `DELETE FROM mandate_periods`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
		VALUES ('test', 'test', '2021-09-01', '2026-09-01', ARRAY['2021/22','2025/26'])`); err != nil {
		t.Fatal(err)
	}
	for _, rm := range []string{"2021_22", "2025_26"} {
		seedBallotMembers(t, h, filepath.Join(dir, "votes-voteringlista-"+rm+".json"))
	}
	for _, year := range []int{2022, 2026} {
		y := year
		h.App.SetNow(func() time.Time { return time.Date(y, 3, 1, 12, 0, 0, 0, time.UTC) })
		h.RunWorker(ctx, "fetch-votes")
	}
	h.RunWorker(ctx, "fetch-motions")
	h.RunWorker(ctx, "fetch-betankanden")
	return h
}

func getProposedBy(t *testing.T, h *Harness, path string) proposedByBody {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("%s: status %d: %s", path, rec.Code, rec.Body.String())
	}
	var b proposedByBody
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("%v: %s", err, rec.Body.String())
	}
	return b
}

func motionIDs(b proposedByBody) []string {
	var ids []string
	for _, m := range b.ProposedBy.Motions {
		ids = append(ids, m.Riksmote+":"+m.Beteckning)
	}
	sort.Strings(ids)
	return ids
}

// The same beteckning and punkt in two riksmöten lists each riksmöte's own
// motions; a 2021/22 votering shows no 2025/26 motion.
func TestVoteringProposedByIsKeyedOnRiksmote(t *testing.T) {
	h := startProposedBy(t)

	old := getProposedBy(t, h, "/api/votes/SkU16/6?session=2021/22")
	got := strings.Join(motionIDs(old), ",")
	if !strings.Contains(got, "2021/22:3638") {
		t.Errorf("2021/22 SkU16 p6 motions = %s, want 2021/22:3638 among them", got)
	}
	for _, m := range old.ProposedBy.Motions {
		if m.Riksmote != "2021/22" {
			t.Errorf("2021/22 SkU16 p6 shows %s:%s from another riksmöte", m.Riksmote, m.Beteckning)
		}
	}

	cur := getProposedBy(t, h, "/api/votes/SkU16/6?session=2025/26")
	if got := strings.Join(motionIDs(cur), ","); got != "2025/26:2783" {
		t.Errorf("2025/26 SkU16 p6 motions = %s, want only 2025/26:2783", got)
	}
	if len(cur.ProposedBy.Motions) == 1 && strings.Join(cur.ProposedBy.Motions[0].Parties, ",") != "V" {
		t.Errorf("2025/26:2783 parties = %v, want [V]", cur.ProposedBy.Motions[0].Parties)
	}
}

// A förslagspunkt that decided motions from several parties lists them all,
// each with its own party and the yrkanden decided.
func TestVoteringListsEveryDecidedMotionWithItsParties(t *testing.T) {
	h := startProposedBy(t)
	b := getProposedBy(t, h, "/api/votes/SkU16/6?session=2021/22")

	want := map[string]string{
		"2021/22:3312": "M", "2021/22:3637": "M", "2021/22:3638": "M",
		"2021/22:3685": "C", "2021/22:4217": "KD",
	}
	if len(b.ProposedBy.Motions) != len(want) {
		t.Fatalf("motions = %v, want %d", motionIDs(b), len(want))
	}
	yr := map[string][]int{"2021/22:3637": {6}, "2021/22:3638": {9}, "2021/22:3685": {17}, "2021/22:4217": {4}}
	for _, m := range b.ProposedBy.Motions {
		id := m.Riksmote + ":" + m.Beteckning
		if p, ok := want[id]; !ok || strings.Join(m.Parties, ",") != p {
			t.Errorf("%s parties = %v, want %s", id, m.Parties, p)
		}
		if y, ok := yr[id]; ok && len(m.Yrkanden) != 1 || ok && m.Yrkanden[0] != y[0] {
			t.Errorf("%s yrkanden = %v, want %v", id, m.Yrkanden, y)
		}
	}
	if !strings.Contains(b.ProposedBy.Alternativ, "reservation 12") {
		t.Errorf("alternativ = %q, want it to name reservation 12", b.ProposedBy.Alternativ)
	}
}

// The decision feed carries the same derived proposed-by, per riksmöte, and
// lists voteringar whether or not any origin step ever ran.
func TestDecisionFeedShowsDerivedProposedBy(t *testing.T) {
	h := startProposedBy(t)
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/votes", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data []struct {
			Riksmote   string `json:"riksmote"`
			Beteckning string `json:"beteckning"`
			Punkt      string `json:"forslagspunkt"`
			ProposedBy struct {
				Motions []struct {
					Beteckning string   `json:"beteckning"`
					Parties    []string `json:"parties"`
				} `json:"motions"`
			} `json:"proposedBy"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	n := map[string]int{}
	for _, d := range body.Data {
		n[d.Riksmote] = len(d.ProposedBy.Motions)
	}
	if len(body.Data) != 2 || n["2021/22"] != 5 || n["2025/26"] != 1 {
		t.Errorf("feed = %s", rec.Body.String())
	}
}

// Nothing writes the retired single-motion origin any more.
func TestNoOriginStepWritesTheOldColumns(t *testing.T) {
	h := startProposedBy(t)
	var n int
	if err := h.DB.QueryRow(context.Background(),
		`SELECT count(*) FROM voteringar WHERE proposed_by_party IS NOT NULL OR proposal_type IS NOT NULL OR proposal_dok_id IS NOT NULL OR origin_enriched`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("%d voteringar carry a stored origin", n)
	}
}

func seedBallotMembers(t *testing.T, h *Harness, path string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Body struct {
			Voteringlista struct {
				Votering []struct {
					IntressentID string `json:"intressent_id"`
					Parti        string `json:"parti"`
				} `json:"votering"`
			} `json:"voteringlista"`
		} `json:"body"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	for _, v := range f.Body.Voteringlista.Votering {
		if _, err := h.DB.Exec(context.Background(),
			`INSERT INTO politicians (intressent_id, first_name, last_name, party) VALUES ($1,'Test','Ledamot',$2) ON CONFLICT DO NOTHING`,
			v.IntressentID, v.Parti); err != nil {
			t.Fatal(err)
		}
	}
}
