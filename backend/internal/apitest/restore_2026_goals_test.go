package apitest

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Migration 000066 restores the seven 2026 party goals dropped by 000064 as
// proposed (never approved) Löften with their original ids.

func TestRestored2026GoalsAreProposedAndNotPublic(t *testing.T) {
	h := startWrite(t)
	ctx := context.Background()

	if n := count(t, h, `SELECT count(*) FROM loften WHERE id BETWEEN 45 AND 51 AND party='S'
		AND visibility='proposed' AND approved_at IS NULL AND mandate_period='2026-2030'
		AND valmanifest_id = (SELECT id FROM valmanifest WHERE party='S' AND election_year=2026)
		AND coalesce(quote,'')<>'' AND coalesce(source_url,'')<>''`); n != 7 {
		t.Fatalf("restored proposed Löften = %d, want 7", n)
	}
	if n := count(t, h, `SELECT count(*) FROM loften WHERE approved_at IS NOT NULL OR visibility='approved'`); n != 0 {
		t.Fatalf("%d approved Löften after migrate; the restore must approve nothing", n)
	}
	if n := count(t, h, `SELECT count(*) FROM aendringslogg WHERE entity='loefte' AND entity_id BETWEEN 45 AND 51
		AND actor='migration' AND new_value->>'visibility'='proposed'`); n != 7 {
		t.Fatalf("ändringslogg entries = %d, want 7", n)
	}
	var seq int
	if err := h.DB.QueryRow(ctx, `SELECT last_value FROM loften_id_seq`).Scan(&seq); err != nil {
		t.Fatal(err)
	}
	if seq < 51 {
		t.Fatalf("loften_id_seq = %d, want >= 51", seq)
	}

	// Proposed Löften are not served publicly: list, detail, party list, log.
	list := publicLoefteIDs(t, h)
	for _, s := range []string{"maffialag", "Höja pensionerna", "Pensionen ska bli bättre"} {
		if strings.Contains(list, s) {
			t.Errorf("public list leaks %q", s)
		}
	}
	for id := 45; id <= 51; id++ {
		if rec := writeReq(h, "GET", "/api/loften/"+itoa(id), "", "", ""); rec.Code != 404 {
			t.Errorf("GET /api/loften/%d = %d, want 404", id, rec.Code)
		}
	}
	if body := writeReq(h, "GET", "/api/parties/S/loften", "", "", "").Body.String(); strings.Contains(body, "maffialag") {
		t.Error("party list leaks a restored Löfte")
	}
	if log := writeReq(h, "GET", "/api/aendringslogg", "", "", "").Body.String(); strings.Contains(log, "maffialag") || strings.Contains(log, "Pensionen ska bli") {
		t.Errorf("public log leaks restored Löften: %s", log)
	}

	// Review path: a proposed Löfte can be approved through the write API, and
	// only then becomes public (checked on one, then reset is not needed: test DB is private).
	rec := writeReq(h, "POST", "/api/write/loften/50/approve", testWriteToken, "ap-50", prov("chat-1#1"))
	if rec.Code != 200 {
		t.Fatalf("approve 50: %d %s", rec.Code, rec.Body)
	}
	if rec := writeReq(h, "GET", "/api/loften/50", "", "", ""); rec.Code != 200 {
		t.Fatalf("approved 50 public: %d", rec.Code)
	}
}

func TestRestore2026GoalsDownRemovesOnlyUntouchedRestoredRows(t *testing.T) {
	h := startWrite(t)
	ctx := context.Background()

	// 45 gets approved, 46 gets edited, 47 gets a koppling-free change of
	// visibility; 48-51 stay as restored. Another proposed row must survive.
	if _, err := h.DB.Exec(ctx, `UPDATE loften SET visibility='approved', approved_at=now() WHERE id=45`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `UPDATE loften SET text = text || ' (ändrad)' WHERE id=46`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `UPDATE loften SET visibility='struck' WHERE id=47`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `INSERT INTO loften (id, party, valmanifest_id, text, quote, source_url, mandate_period, visibility)
		VALUES (900,'S',(SELECT id FROM valmanifest WHERE party='S' AND election_year=2026),'Annat','Annat citat','https://example.org/x','2026-2030','proposed')`); err != nil {
		t.Fatal(err)
	}
	before := count(t, h, `SELECT count(*) FROM loften`)

	_, file, _, _ := runtime.Caller(0)
	down, err := os.ReadFile(filepath.Join(filepath.Dir(file), "..", "..", "migrations", "000066_restore_2026_goals.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, string(down)); err != nil {
		t.Fatalf("down: %v", err)
	}

	if n := count(t, h, `SELECT count(*) FROM loften WHERE id BETWEEN 48 AND 51`); n != 0 {
		t.Errorf("untouched restored rows left: %d", n)
	}
	if n := count(t, h, `SELECT count(*) FROM loften WHERE id IN (45,46,47,900)`); n != 4 {
		t.Errorf("approved/edited/struck/other rows must survive, got %d of 4", n)
	}
	if after := count(t, h, `SELECT count(*) FROM loften`); after != before-4 {
		t.Errorf("loften rows %d -> %d, want exactly 4 removed", before, after)
	}
}
