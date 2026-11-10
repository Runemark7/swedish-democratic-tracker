package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"riksdagskollen/internal/votes"
	votepg "riksdagskollen/internal/votes/adapters/postgres"
	voterd "riksdagskollen/internal/votes/adapters/riksdagen"
)

// Fixtures in fixtures/origin-per-riksmote were recorded from data.riksdagen.se
// on 2026-11-10 (see each file's recordedFrom). One förslagspunkt per case:
//
//	2021/22 SkU16 p6   H901SkU16  first motion H9021211 (C)
//	2022/23 SkU15 p4   HA01SkU15  first motion HA021022 (SD)
//	2024/25 SkU14 p3   HC01SkU14  first motion HC021150 (M)
//	2025/26 SkU16 p6   HD01SkU16  first motion HD021141 (C)
type originCase struct {
	rm, bet, punkt, wantMotion, wantParty string
}

var originCases = []originCase{
	{"2021/22", "SkU16", "6", "H9021211", "C"},
	{"2022/23", "SkU15", "4", "HA021022", "SD"},
	{"2024/25", "SkU14", "3", "HC021150", "M"},
	{"2025/26", "SkU16", "6", "HD021141", "C"},
}

// ingestOrigins stores each case's votering (oldest riksmöte first, so the
// votes cursor only moves forward) and returns the harness.
func ingestOrigins(t *testing.T) *Harness {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(file), "fixtures", "origin-per-riksmote")
	h := Start(t, dir)
	ctx := context.Background()
	for _, c := range originCases {
		h.seedFixtureMembers(t, "voteringlista-"+strings.ReplaceAll(c.rm, "/", "_")+".json")
	}
	for _, c := range originCases {
		// The votes worker picks the riksmöte from the clock: pin it inside rm.
		year := 2000 + int((c.rm[2]-'0')*10+(c.rm[3]-'0'))
		h.App.SetNow(func() time.Time { return time.Date(year, 12, 1, 12, 0, 0, 0, time.UTC) })
		h.RunWorker(ctx, "fetch-votes")
	}
	return h
}

type originRow struct {
	Session, Bet, Punkt string
	Party, Type, DokID  *string
	Enriched            bool
}

func readOrigins(t *testing.T, h *Harness) map[string]originRow {
	t.Helper()
	rows, err := h.DB.Query(context.Background(),
		`SELECT session, beteckning, forslagspunkt, proposed_by_party, proposal_type, proposal_dok_id, origin_enriched FROM voteringar`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[string]originRow{}
	for rows.Next() {
		var r originRow
		if err := rows.Scan(&r.Session, &r.Bet, &r.Punkt, &r.Party, &r.Type, &r.DokID, &r.Enriched); err != nil {
			t.Fatal(err)
		}
		out[r.Session+" "+r.Bet+" p"+r.Punkt] = r
	}
	return out
}

func str(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

// motionRiksmote reads the riksmöte out of a motion's dok_id: H9 = 2021/22,
// HA = 2022/23, HC = 2024/25, HD = 2025/26.
func motionRiksmote(dokID string) string {
	switch {
	case strings.HasPrefix(dokID, "H9"):
		return "2021/22"
	case strings.HasPrefix(dokID, "HA"):
		return "2022/23"
	case strings.HasPrefix(dokID, "HC"):
		return "2024/25"
	case strings.HasPrefix(dokID, "HD"):
		return "2025/26"
	}
	return "?"
}

func assertOriginsPerRiksmote(t *testing.T, h *Harness) {
	t.Helper()
	got := readOrigins(t, h)
	if len(got) != len(originCases) {
		t.Fatalf("stored %d voteringar, want %d: %+v", len(got), len(originCases), got)
	}
	for _, c := range originCases {
		r, ok := got[c.rm+" "+c.bet+" p"+c.punkt]
		if !ok {
			t.Fatalf("no votering for %s %s p%s", c.rm, c.bet, c.punkt)
		}
		if !r.Enriched || str(r.DokID) != c.wantMotion || str(r.Party) != c.wantParty || str(r.Type) != "mot" {
			t.Errorf("%s %s p%s origin = {enriched:%v motion:%s party:%s type:%s}, want motion %s party %s",
				c.rm, c.bet, c.punkt, r.Enriched, str(r.DokID), str(r.Party), str(r.Type), c.wantMotion, c.wantParty)
		}
		if mr := motionRiksmote(str(r.DokID)); mr != c.rm {
			t.Errorf("%s %s p%s shows motion %s from riksmöte %s", c.rm, c.bet, c.punkt, str(r.DokID), mr)
		}
	}
	// SkU16 p6 exists in two riksmöten and must not share an origin.
	a, b := got["2021/22 SkU16 p6"], got["2025/26 SkU16 p6"]
	if str(a.DokID) == str(b.DokID) {
		t.Errorf("SkU16 p6 in 2021/22 and 2025/26 share origin %s", str(a.DokID))
	}
}

// AC1+AC2+AC3: origin resolved from the votering's own riksmöte and written
// only to it. 2022/23 SkU15 p4 gets an origin where Riksdagen has a reference.
func TestOriginIsResolvedAndWrittenPerRiksmote(t *testing.T) {
	h := ingestOrigins(t)
	h.RunWorker(context.Background(), "enrich-vote-origins")
	assertOriginsPerRiksmote(t, h)
}

// AC4: the recompute re-resolves rows a previous (buggy) run tagged with one
// riksmöte's motion for every year, and can be run again safely.
func TestRecomputeOriginsFixesWrongTagsAndIsRerunnable(t *testing.T) {
	h := ingestOrigins(t)
	ctx := context.Background()
	// The shipped bug: every SkU16 p6 carries the 2021/22 motion; 2022/23 has none.
	if _, err := h.DB.Exec(ctx, `UPDATE voteringar SET proposed_by_party='C', proposal_type='mot',
		proposal_dok_id='H9021211', origin_enriched = true WHERE beteckning='SkU16'`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `UPDATE voteringar SET origin_enriched = true WHERE beteckning IN ('SkU15','SkU14')`); err != nil {
		t.Fatal(err)
	}

	svc := votes.NewService(votepg.NewRepository(h.DB), voterd.NewClientWithBaseURL(h.ReplayURL()))
	svc.SetOriginDelay(0)

	marked, err := svc.MarkOriginsForRecompute(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if marked != int64(len(originCases)) {
		t.Errorf("marked %d, want %d", marked, len(originCases))
	}
	n, err := svc.RecomputeOrigins(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(originCases) {
		t.Errorf("recomputed %d, want %d", n, len(originCases))
	}
	assertOriginsPerRiksmote(t, h)

	// Re-run: mark again and recompute — same result, nothing duplicated.
	if _, err := svc.MarkOriginsForRecompute(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RecomputeOrigins(ctx, 3); err != nil {
		t.Fatal(err)
	}
	assertOriginsPerRiksmote(t, h)

	// Without a mark, a second drain has nothing to do (resume is a no-op).
	n, err = svc.RecomputeOrigins(ctx, 3)
	if err != nil || n != 0 {
		t.Errorf("unmarked recompute = %d, %v; want 0, nil", n, err)
	}
}

// AC5: the vote detail page's data shows the corrected "proposed by" for
// 2024/25 SkU14 p3 — the motion is from 2024/25 and carries its own party.
func TestVoteDetailShowsOwnRiksmoteOrigin(t *testing.T) {
	h := ingestOrigins(t)
	ctx := context.Background()
	h.RunWorker(ctx, "enrich-vote-origins")

	req := httptest.NewRequest(http.MethodGet, "/api/votes/SkU14/3?session=2024/25", nil)
	rec := h.Do(req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		ProposedByParty string `json:"proposedByParty"`
		ProposalType    string `json:"proposalType"`
		Session         string `json:"session"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("%v: %s", err, rec.Body.String())
	}
	if body.ProposedByParty != "M" || body.ProposalType != "mot" || body.Session != "2024/25" {
		t.Errorf("detail origin = %+v (body %s)", body, rec.Body.String())
	}
	// The motion itself is HC021150 (riksmöte 2024/25), not another year's.
	if got := readOrigins(t, h)["2024/25 SkU14 p3"]; str(got.DokID) != "HC021150" {
		t.Errorf("stored origin motion = %s, want HC021150", str(got.DokID))
	}
}

// seedFixtureMembers inserts the members who cast the recorded ballots, since a
// ballot is only stored for a known politician.
func (h *Harness) seedFixtureMembers(t *testing.T, fixture string) {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(file), "fixtures", "origin-per-riksmote", fixture))
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
