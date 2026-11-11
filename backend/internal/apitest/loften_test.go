package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type loftePublic struct {
	ID            int    `json:"id"`
	Party         string `json:"party"`
	Text          string `json:"text"`
	Quote         string `json:"quote"`
	MandatePeriod string `json:"mandatePeriod"`
	Visibility    string `json:"visibility"`
	Archived      bool   `json:"archived"`
	ValmanifestID *int   `json:"valmanifestId"`
}

// startLoften builds an API whose party_goals hold three 2022 goals and one
// 2026 goal, then runs the real 000057 migration over them (down, then up), so
// the archive step is exercised on existing rows rather than an empty table.
func startLoften(t *testing.T) *Harness {
	t.Helper()
	// Löften read only Postgres; any recorded set satisfies the replay server.
	h := Start(t, chainDir(t, "beslut-list"))
	ctx := context.Background()
	dir := filepath.Join(migrationsDirForTest(t))
	down, err := os.ReadFile(filepath.Join(dir, "000057_loften.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	up, err := os.ReadFile(filepath.Join(dir, "000057_loften.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	// 000059 (kopplingar) refers to loften, so it goes down first and up again last.
	down59, err := os.ReadFile(filepath.Join(dir, "000059_kopplingar.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	up59, err := os.ReadFile(filepath.Join(dir, "000059_kopplingar.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	down60, err := os.ReadFile(filepath.Join(dir, "000060_veckobatch_items.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	up60, err := os.ReadFile(filepath.Join(dir, "000060_veckobatch_items.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, string(down60)); err != nil {
		t.Fatalf("down 60: %v", err)
	}
	if _, err := h.DB.Exec(ctx, string(down59)); err != nil {
		t.Fatalf("down 59: %v", err)
	}
	if _, err := h.DB.Exec(ctx, string(down)); err != nil {
		t.Fatalf("down: %v", err)
	}
	// The seed migrations already filled party_goals; this is a throwaway DB.
	if _, err := h.DB.Exec(ctx, `DELETE FROM party_goals`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO party_goals (id, party, goal_text, topic, specificity, source_document, election_cycle) VALUES
		 (7,  'M', 'Fler poliser',        'trygghet', 'concrete',   'Valmanifest 2022', '2022'),
		 (22, 'C', 'Värna integriteten',  'frihet',   'directional','Partiprogram',     NULL),
		 (1,  'S', 'Kortare vårdköer',    'sjukvard', 'concrete',   'Valmanifest 2022', '2022'),
		 (45, 'S', 'Plan för Sverige',    'ekonomi',  'concrete',   'Valplattform 2026','2026')`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, string(up)); err != nil {
		t.Fatalf("up: %v", err)
	}
	if _, err := h.DB.Exec(ctx, string(up59)); err != nil {
		t.Fatalf("up 59: %v", err)
	}
	if _, err := h.DB.Exec(ctx, string(up60)); err != nil {
		t.Fatalf("up 60: %v", err)
	}
	// A proposed and an approved Löfte, created by the test the way a later
	// write API would. Nothing in the migration approves anything.
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO loften (party, text, quote, locator, source_url, mandate_period, visibility) VALUES
		 ('M', 'Förslag som väntar', 'citat', 's. 3', 'https://x.example/m', '2026-2030', 'proposed'),
		 ('M', 'Struken',            'citat', 's. 4', 'https://x.example/m', '2026-2030', 'struck')`); err != nil {
		t.Fatal(err)
	}
	return h
}

func migrationsDirForTest(t *testing.T) string {
	t.Helper()
	return strings.TrimPrefix(migrationsURL(), "file://")
}

func loftenGet(t *testing.T, h *Harness, path string, v any) int {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, path, nil))
	if v != nil && rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
			t.Fatalf("%s: %v: %s", path, err, rec.Body.String())
		}
	}
	return rec.Code
}

// Archived 2022 Löften keep the ids of the goals they came from and carry no
// Löftesstatus; the 2026 campaign goal is not archived.
func TestArchivedLoftenKeepTheirOldIdsAndShowNoLoftesstatus(t *testing.T) {
	h := startLoften(t)
	var all []loftePublic
	if code := loftenGet(t, h, "/api/loften", &all); code != 200 {
		t.Fatalf("status %d", code)
	}
	ids := map[int]loftePublic{}
	for _, l := range all {
		ids[l.ID] = l
	}
	for _, id := range []int{7, 22, 1} {
		l, ok := ids[id]
		if !ok {
			t.Fatalf("goal %d missing from archive: %+v", id, all)
		}
		if !l.Archived || l.Visibility != "archived" || l.MandatePeriod != "2022-2026" {
			t.Errorf("%d = %+v", id, l)
		}
	}
	if _, ok := ids[45]; ok {
		t.Error("2026 goal 45 was archived")
	}
	if ids[7].ValmanifestID == nil {
		t.Error("goal 7 (M, valmanifest 2022) is not linked to M's 2022 Valmanifest")
	}
	if ids[22].ValmanifestID != nil {
		t.Error("a standing programme goal was linked to a Valmanifest")
	}

	var one map[string]any
	if code := loftenGet(t, h, "/api/loften/7", &one); code != 200 {
		t.Fatalf("detail status %d", code)
	}
	if one["text"] != "Fler poliser" {
		t.Errorf("detail = %v", one)
	}
	raw := h.Do(httptest.NewRequest(http.MethodGet, "/api/loften", nil)).Body.String()
	for _, banned := range []string{"status", "Status", "Röstat", "Inte prövat"} {
		if strings.Contains(raw, banned) {
			t.Errorf("response mentions %q: archived Löften must not show a Löftesstatus", banned)
		}
	}
	// The old tables are untouched.
	var n int
	if err := h.DB.QueryRow(context.Background(), `SELECT count(*) FROM party_goals`).Scan(&n); err != nil || n != 4 {
		t.Errorf("party_goals = %d, %v", n, err)
	}
}

// A Löfte that is proposed (not approved), or struck, is never public: not in
// the overview, the party list or by id.
func TestProposedLoftenNeverAppearInPublicResponses(t *testing.T) {
	h := startLoften(t)
	var hidden []int
	rows, err := h.DB.Query(context.Background(), `SELECT id FROM loften WHERE visibility IN ('proposed','struck')`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var id int
		_ = rows.Scan(&id)
		hidden = append(hidden, id)
	}
	rows.Close()
	if len(hidden) != 2 {
		t.Fatalf("hidden = %v", hidden)
	}
	for _, path := range []string{"/api/loften", "/api/loften?party=M", "/api/parties/M/loften"} {
		body := h.Do(httptest.NewRequest(http.MethodGet, path, nil)).Body.String()
		if strings.Contains(body, "Förslag som väntar") || strings.Contains(body, "Struken") {
			t.Errorf("%s leaks a non-public Löfte: %s", path, body)
		}
	}
	for _, id := range hidden {
		rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/loften/"+itoa(id), nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET /api/loften/%d = %d", id, rec.Code)
		}
	}
	if code := loftenGet(t, h, "/api/loften/9999", nil); code != http.StatusNotFound {
		t.Errorf("unknown id = %d", code)
	}
}

func itoa(i int) string { b, _ := json.Marshal(i); return string(b) }

// Each party's Valmanifest is listed as a source: title, URL, draft or final.
func TestPartyListsItsValmanifestAsSources(t *testing.T) {
	h := startLoften(t)
	var list []struct {
		Party        string  `json:"party"`
		ElectionYear int     `json:"electionYear"`
		Title        string  `json:"title"`
		URL          string  `json:"url"`
		State        string  `json:"state"`
		PublishedOn  *string `json:"publishedOn"`
	}
	if code := loftenGet(t, h, "/api/parties/MP/valmanifest", &list); code != 200 {
		t.Fatalf("status %d", code)
	}
	if len(list) == 0 {
		t.Fatal("no Valmanifest for MP")
	}
	sawDraft := false
	for _, v := range list {
		if v.Party != "MP" || v.URL == "" || v.Title == "" {
			t.Errorf("row %+v", v)
		}
		if v.State != "draft" && v.State != "final" {
			t.Errorf("state %q", v.State)
		}
		if v.ElectionYear == 2026 && v.State == "draft" {
			sawDraft = true
		}
	}
	if !sawDraft {
		t.Error("MP's 2026 utkast is not marked draft")
	}
	if list[0].ElectionYear < list[len(list)-1].ElectionYear {
		t.Error("newest election first")
	}
	var bad any
	if code := loftenGet(t, h, "/api/parties/M;DROP/valmanifest", &bad); code != http.StatusBadRequest {
		t.Errorf("bad party = %d", code)
	}
}
