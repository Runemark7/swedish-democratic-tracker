package apitest

import (
	"context"
	"strings"
	"testing"
)

// Issue #259: archived 2022–2026 Löften with a quote and a source URL can get
// suggested and Bekräftade kopplingar, and then a Löftesstatus like 2026 Löften.

type archSeed struct {
	*kSeed
	withQuote, noQuote, noURL int64
	v2023, v2023b, v2026      int64
}

func seedArchived(t *testing.T, h *Harness, party, text, utskott, quote, url string) int64 {
	t.Helper()
	var id int64
	if err := h.DB.QueryRow(context.Background(),
		`INSERT INTO loften (party, valmanifest_id, text, quote, source_url, utskott, mandate_period, visibility)
		 VALUES ($1, (SELECT id FROM valmanifest WHERE party=$1 ORDER BY id LIMIT 1), $2, NULLIF($4,''), NULLIF($5,''),
		         string_to_array($3, ','), '2022-2026', 'archived') RETURNING id`,
		party, text, utskott, quote, url).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func startArch(t *testing.T) *archSeed {
	t.Helper()
	s := &archSeed{kSeed: startK(t)}
	// The archive the migrations seed would add noise to the candidate lists.
	if _, err := s.h.DB.Exec(context.Background(), `UPDATE loften SET visibility='struck' WHERE visibility='archived'`); err != nil {
		t.Fatal(err)
	}
	s.withQuote = seedArchived(t, s.h, "S", "Gammal skatt", "SkU", "citat gammal skatt", "https://example.org/2022")
	s.noQuote = seedArchived(t, s.h, "S", "Utan citat", "SkU", "", "https://example.org/2022")
	s.noURL = seedArchived(t, s.h, "S", "Utan källa", "SkU", "citat", "")
	// Old voteringar, ingested before the last closed Veckobatch.
	s.v2023 = seedVotering(t, s.h, "A-2023", "SkU4", "2023/24", "2024-03-20 10:00+00", "2024-03-21 10:00+00")
	s.v2023b = seedVotering(t, s.h, "A-2023b", "AU4", "2023/24", "2024-03-20 10:00+00", "2024-03-21 10:00+00")
	s.v2026 = seedVotering(t, s.h, "A-2026", "SkU5", "2026/27", "2026-10-20 10:00+00", "2026-10-21 10:00+00")
	return s
}

func (s *archSeed) candidatesAt(t *testing.T, path string) map[string][]int64 {
	t.Helper()
	rec := writeReq(s.h, "GET", path, testWriteToken, "", "")
	if rec.Code != 200 {
		t.Fatalf("%s: %d %s", path, rec.Code, rec.Body)
	}
	v := decode[struct {
		Voteringar []struct {
			VoteringID string `json:"voteringId"`
			Loften     []struct{ ID int64 }
		}
	}](t, rec)
	out := map[string][]int64{}
	for _, c := range v.Voteringar {
		for _, l := range c.Loften {
			out[c.VoteringID] = append(out[c.VoteringID], l.ID)
		}
	}
	return out
}

func TestCandidatesDefaultNeverOffersArchive(t *testing.T) {
	s := startArch(t)
	got := s.candidatesAt(t, "/api/write/candidates")
	for vid, ids := range got {
		for _, id := range ids {
			if id == s.withQuote || id == s.noQuote || id == s.noURL {
				t.Fatalf("archived Löfte %d offered on %s in the default read", id, vid)
			}
		}
	}
	if _, ok := got["A-2023"]; ok {
		t.Fatalf("2023/24 votering offered in the default read: %+v", got)
	}
}

func TestCandidatesForArchivePeriod(t *testing.T) {
	s := startArch(t)
	got := s.candidatesAt(t, "/api/write/candidates?mandatePeriod=2022-2026")
	// The old votering is offered despite the Veckobatch cutoff, with only the complete archived Löfte.
	if ids := got["A-2023"]; len(ids) != 1 || ids[0] != s.withQuote {
		t.Fatalf("A-2023 Löften = %v; want only %d (no quote / no source URL are never offered)", ids, s.withQuote)
	}
	if _, ok := got["A-2023b"]; ok {
		t.Fatalf("AU votering offered for SkU Löfte: %+v", got)
	}
	if _, ok := got["A-2026"]; ok {
		t.Fatalf("2026/27 votering offered for 2022–2026: %+v", got)
	}
	for vid, ids := range got {
		for _, id := range ids {
			if id == s.loefteSkU || id == s.loefteAU {
				t.Fatalf("2026 Löfte %d offered on %s for the archive period", id, vid)
			}
		}
	}
	rec := writeReq(s.h, "GET", "/api/write/candidates?mandatePeriod=nu", testWriteToken, "", "")
	if rec.Code != 400 {
		t.Fatalf("bad period: %d", rec.Code)
	}
	// Already-suggested pairs are omitted.
	if rec := writeReq(s.h, "POST", "/api/write/kopplingar", testWriteToken, "arch-c1",
		suggestBody(sug(s.withQuote, "A-2023", "Ja", 0.9, "skäl"))); rec.Code != 200 {
		t.Fatalf("suggest: %d %s", rec.Code, rec.Body)
	}
	if _, ok := s.candidatesAt(t, "/api/write/candidates?mandatePeriod=2022-2026")["A-2023"]; ok {
		t.Fatal("suggested pair offered again")
	}
}

func TestSuggestionsAcceptCompleteArchivedLoefte(t *testing.T) {
	s := startArch(t)
	post := func(key, body string, want int) {
		t.Helper()
		if rec := writeReq(s.h, "POST", "/api/write/kopplingar", testWriteToken, key, body); rec.Code != want {
			t.Fatalf("%s: %d %s (want %d)", key, rec.Code, rec.Body, want)
		}
	}
	post("arch-s-no-quote", suggestBody(sug(s.noQuote, "A-2023", "Ja", 0.9, "x")), 422)
	post("arch-s-no-url", suggestBody(sug(s.noURL, "A-2023", "Ja", 0.9, "x")), 422)
	post("arch-s-wrong-period", suggestBody(sug(s.withQuote, "A-2026", "Ja", 0.9, "x")), 422)
	post("arch-s-2026-in-2023", suggestBody(sug(s.loefteSkU, "A-2023", "Ja", 0.9, "x")), 422)
	post("arch-s-ok", suggestBody(sug(s.withQuote, "A-2023", "Ja", 0.9, "x")), 200)
	var n int
	_ = s.h.DB.QueryRow(context.Background(), `SELECT count(*) FROM kopplingar WHERE loefte_id=$1`, s.withQuote).Scan(&n)
	if n != 1 {
		t.Fatalf("%d kopplingar on the archived Löfte", n)
	}
}

func TestArchivedLoefteGetsStatusFromBekraftadKoppling(t *testing.T) {
	s := startArch(t)
	if v := statusOf(t, s.h, s.withQuote); v.Status != nil {
		t.Fatalf("no Bekräftad koppling yet, status %q", *v.Status)
	}
	if rec := writeReq(s.h, "POST", "/api/write/kopplingar", testWriteToken, "arch-st",
		suggestBody(sug(s.withQuote, "A-2023", "Ja", 0.9, "x"))); rec.Code != 200 {
		t.Fatalf("suggest: %d %s", rec.Code, rec.Body)
	}
	var kid int64
	_ = s.h.DB.QueryRow(context.Background(), `SELECT id FROM kopplingar WHERE loefte_id=$1`, s.withQuote).Scan(&kid)
	// Suggested only: still no status and no public koppling.
	if v := statusOf(t, s.h, s.withQuote); v.Status != nil || len(v.Kopplingar) != 0 {
		t.Fatalf("föreslagen koppling leaked: %+v", v)
	}
	ballotsFor(t, s.h, s.v2023, "S", 1, 20, 0)
	confirmKoppling(t, s.kSeed, kid, "arch-b")
	v := statusOf(t, s.h, s.withQuote)
	if v.Status == nil || *v.Status != "Röstat emot" {
		t.Fatalf("status %v; want Röstat emot (S voted Nej against the Ja direction)", v.Status)
	}
	if len(v.Kopplingar) != 1 || v.Kopplingar[0].Effect != "emot" {
		t.Fatalf("kopplingar %+v", v.Kopplingar)
	}
	// Siblings without a confirmed koppling still show no status.
	if v := statusOf(t, s.h, s.noQuote); v.Status != nil {
		t.Fatalf("status %q on an archived Löfte without kopplingar", *v.Status)
	}
}

func TestArchiveCountsStaySeparateFrom2026(t *testing.T) {
	s := startArch(t)
	if rec := writeReq(s.h, "POST", "/api/write/kopplingar", testWriteToken, "arch-sum",
		suggestBody(sug(s.withQuote, "A-2023", "Ja", 0.9, "x"))); rec.Code != 200 {
		t.Fatal(rec.Body)
	}
	var kid int64
	_ = s.h.DB.QueryRow(context.Background(), `SELECT id FROM kopplingar WHERE loefte_id=$1`, s.withQuote).Scan(&kid)
	ballotsFor(t, s.h, s.v2023, "S", 20, 1, 0)
	confirmKoppling(t, s.kSeed, kid, "arch-sumb")

	type row struct {
		Party                                        string
		Total, For, Emot, Blandat, InteProvat, Arkiv int
	}
	find := func(rows []row) row {
		for _, r := range rows {
			if r.Party == "S" {
				return r
			}
		}
		return row{}
	}
	old := find(decode[[]row](t, publicGet(t, s.h, "/api/loften/summary?mandatperiod=2022-2026", 200)))
	if old.Total != 3 || old.For != 1 || old.Arkiv != 2 {
		t.Fatalf("archive summary %+v; want 3 total, 1 Röstat för, 2 without status", old)
	}
	now := find(decode[[]row](t, publicGet(t, s.h, "/api/loften/summary?mandatperiod=2026-2030", 200)))
	if now.Total != 2 || now.For != 0 || now.Arkiv != 0 || now.InteProvat != 2 {
		t.Fatalf("2026 summary %+v: the archive must not leak in", now)
	}
	// The front page tiles are the current period only.
	fp := decode[struct {
		PartyTiles []row `json:"partyTiles"`
	}](t, publicGet(t, s.h, "/api/frontpage", 200))
	if got := find(fp.PartyTiles); got.Total != 2 || got.For != 0 || got.Arkiv != 0 {
		t.Fatalf("front-page S tile %+v: must count 2026–2030 only", got)
	}
	_ = strings.TrimSpace
}
