package apitest

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Seam tests for the CSV download of the Löften selection (gap
// no-data-export-csv). The file is the same selection as GET /api/loften with
// the same filters, read only from Postgres.

const utf8BOM = "\xEF\xBB\xBF"

// loftenCSV fetches the download and parses it the way a spreadsheet would:
// BOM first, semicolon-separated, one header row.
func loftenCSV(t *testing.T, h *Harness, path string) (header []string, rows []map[string]string) {
	t.Helper()
	rec := publicGet(t, h, path, 200)
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv; charset=utf-8" {
		t.Errorf("%s Content-Type = %q", path, ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.HasPrefix(cd, "attachment;") || !strings.Contains(cd, ".csv") {
		t.Errorf("%s Content-Disposition = %q", path, cd)
	}
	body := rec.Body.String()
	if !strings.HasPrefix(body, utf8BOM) {
		t.Fatalf("%s does not start with a UTF-8 BOM: %q", path, body[:min(len(body), 20)])
	}
	r := csv.NewReader(strings.NewReader(strings.TrimPrefix(body, utf8BOM)))
	r.Comma = ';'
	all, err := r.ReadAll()
	if err != nil {
		t.Fatalf("%s is not valid CSV: %v\n%s", path, err, body)
	}
	if len(all) == 0 {
		t.Fatalf("%s has no header row", path)
	}
	header = all[0]
	for _, rec := range all[1:] {
		row := map[string]string{}
		for i, col := range header {
			row[col] = rec[i]
		}
		rows = append(rows, row)
	}
	return header, rows
}

func TestLoftenCSVHoldsTheFilteredSelection(t *testing.T) {
	s := startK(t)
	ctx := context.Background()
	// One Bekräftad koppling where S voted in the Löfte's direction.
	k := addKoppling(t, s, "x", 0.9, "Ja")
	var vref int64
	if err := s.h.DB.QueryRow(ctx, `SELECT votering_ref FROM kopplingar WHERE id=$1`, k).Scan(&vref); err != nil {
		t.Fatal(err)
	}
	ballotsFor(t, s.h, vref, "S", 9, 1, 0)
	confirmKoppling(t, s, k, "csv1")
	// A Löfte whose text needs every kind of escaping, and one that a
	// spreadsheet would otherwise run as a formula.
	tricky := "Sänk \"skatten\"; höj\nbarnbidraget, å ä ö"
	trickyID := seedLoefte(t, s.h, "M", tricky, "approved", "SkU,FiU")
	formulaID := seedLoefte(t, s.h, "M", "=HYPERLINK(\"https://x.example\")", "approved", "AU")

	header, rows := loftenCSV(t, s.h, "/api/loften.csv")
	want := []string{"Id", "Parti", "Löfte", "Löftesstatus", "Status ändrad", "Mandatperiod", "Utskott", "Ämne", "Citat", "Plats i källan", "Källdokument", "Källa (URL)", "Arkiverat"}
	if strings.Join(header, "|") != strings.Join(want, "|") {
		t.Fatalf("header = %q\nwant     %q", header, want)
	}

	// Every filter combination yields exactly the ids of the JSON list.
	for _, q := range []string{
		"", "?party=S", "?party=M", "?party=V&status=for", "?utskott=SkU", "?status=for", "?status=inte-provat",
		"?status=arkiv", "?party=S&utskott=SkU&status=for", "?utskott=AU&status=for", "?betankande=2026/27:SkU7",
	} {
		js := decode[[]struct{ ID int }](t, publicGet(t, s.h, "/api/loften"+q, 200))
		_, got := loftenCSV(t, s.h, "/api/loften.csv"+q)
		if len(got) != len(js) {
			t.Errorf("%s: CSV has %d rows, JSON list has %d", q, len(got), len(js))
			continue
		}
		for i := range js {
			if got[i]["Id"] != strconv.Itoa(js[i].ID) {
				t.Errorf("%s row %d: CSV id %s, JSON id %d", q, i, got[i]["Id"], js[i].ID)
			}
		}
	}
	// The party path is the same selection as ?party=.
	_, byPath := loftenCSV(t, s.h, "/api/parties/S/loften.csv")
	_, byQuery := loftenCSV(t, s.h, "/api/loften.csv?party=S")
	if len(byPath) < 2 || len(byPath) != len(byQuery) {
		t.Errorf("/api/parties/S/loften.csv has %d rows, ?party=S has %d", len(byPath), len(byQuery))
	}
	for _, r := range byPath {
		if r["Parti"] != "S" {
			t.Errorf("/api/parties/S/loften.csv holds a %s row", r["Parti"])
		}
	}

	by := map[string]map[string]string{}
	for _, r := range rows {
		by[r["Id"]] = r
	}
	sku := by[fmt.Sprint(s.loefteSkU)]
	if sku == nil {
		t.Fatalf("Löfte %d missing: %v", s.loefteSkU, rows)
	}
	for col, want := range map[string]string{
		"Parti": "S", "Löfte": "Skatt sänks", "Löftesstatus": "Röstat för", "Mandatperiod": "2026-2030",
		"Utskott": "SkU", "Citat": "citat Skatt sänks", "Källa (URL)": "https://example.org/v", "Arkiverat": "nej",
	} {
		if sku[col] != want {
			t.Errorf("SkU Löfte %s = %q, want %q", col, sku[col], want)
		}
	}
	if len(sku["Status ändrad"]) != len("2026-11-13") {
		t.Errorf("Status ändrad = %q, want a date", sku["Status ändrad"])
	}
	if au := by[fmt.Sprint(s.loefteAU)]; au["Löftesstatus"] != "Inte prövat än" || au["Status ändrad"] != "" {
		t.Errorf("AU Löfte = %v", au)
	}
	// Quotes, semicolons, line breaks and å ä ö survive the round trip.
	if got := by[fmt.Sprint(trickyID)]; got["Löfte"] != tricky || got["Utskott"] != "SkU, FiU" {
		t.Errorf("tricky Löfte = %q / %q", got["Löfte"], got["Utskott"])
	}
	// A cell that starts like a formula is neutralised with a leading apostrophe.
	if got := by[fmt.Sprint(formulaID)]["Löfte"]; got != `'=HYPERLINK("https://x.example")` {
		t.Errorf("formula cell = %q", got)
	}
	// A proposed Löfte is never public, in any format.
	if _, leaked := by[fmt.Sprint(s.loefteProposed)]; leaked {
		t.Error("the CSV leaks a proposed Löfte")
	}
	if raw := publicGet(t, s.h, "/api/loften.csv", 200).Body.String(); strings.Contains(raw, "Bara förslag") {
		t.Error("the CSV leaks a proposed Löfte's text")
	}

	// An empty selection is still a file with its header row.
	if _, none := loftenCSV(t, s.h, "/api/loften.csv?party=V&status=for"); len(none) != 0 {
		t.Errorf("party=V&status=for rows = %v", none)
	}
	// Bad filters are refused exactly as on the JSON list.
	for _, bad := range []string{"?status=hållet", "?party=1", "?utskott=x", "?betankande=SkU2"} {
		rec := s.h.Do(httptest.NewRequest(http.MethodGet, "/api/loften.csv"+bad, nil))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s = %d, want 400", bad, rec.Code)
		}
	}
}

// Archived Löften have no Löftesstatus; the file leaves the cell empty rather
// than printing a label.
func TestLoftenCSVShowsNoLoftesstatusForArchivedLoften(t *testing.T) {
	h := startLoften(t)
	_, rows := loftenCSV(t, h, "/api/loften.csv?status=arkiv")
	if len(rows) != 3 {
		t.Fatalf("archived rows = %d, want 3: %v", len(rows), rows)
	}
	for _, r := range rows {
		if r["Löftesstatus"] != "" || r["Arkiverat"] != "ja" || r["Mandatperiod"] != "2022-2026" {
			t.Errorf("archived row = %v", r)
		}
	}
	raw := publicGet(t, h, "/api/loften.csv", 200).Body.String()
	for _, banned := range []string{"Röstat", "Inte prövat", "Förslag som väntar", "Struken"} {
		if strings.Contains(raw, banned) {
			t.Errorf("CSV mentions %q", banned)
		}
	}
}

// The Mandatperiod filter of the list applies to the file too, alone and
// combined with the other filters, so the download equals the visible
// selection.
func TestLoftenCSVFollowsMandatperiod(t *testing.T) {
	s := startPeriods(t)
	csvTexts := func(path string) string {
		_, rows := loftenCSV(t, s.h, path)
		texts := make([]string, 0, len(rows))
		for _, r := range rows {
			texts = append(texts, r["Löfte"])
		}
		sort.Strings(texts)
		return strings.Join(texts, "|")
	}
	for _, c := range []struct{ name, query, want string }{
		{"no period keeps every period", "?party=S", "Arbetsrätt|Gammal arbetsrätt|Gammal skatt|Skatt sänks"},
		{"current period only", "?mandatperiod=2026-2030", "Arbetsrätt|Skatt sänks"},
		{"previous period only", "?mandatperiod=2022-2026", "Gammal arbetsrätt|Gammal skatt|Gammalt M-löfte"},
		{"period and party", "?mandatperiod=2022-2026&party=S", "Gammal arbetsrätt|Gammal skatt"},
		{"period and utskott", "?mandatperiod=2022-2026&utskott=SkU", "Gammal skatt|Gammalt M-löfte"},
		{"period, party and utskott", "?mandatperiod=2026-2030&party=S&utskott=AU", "Arbetsrätt"},
		{"period and status", "?mandatperiod=2026-2030&status=inte-provat", "Arbetsrätt|Skatt sänks"},
		{"period and arkiv", "?mandatperiod=2026-2030&status=arkiv", ""},
		{"a period without Löften is a header-only file", "?mandatperiod=2018-2022", ""},
	} {
		got := csvTexts("/api/loften.csv" + c.query)
		if got != c.want {
			t.Errorf("%s: CSV %s = %q, want %q", c.name, c.query, got, c.want)
		}
		// The file is the JSON list for the same query, nothing more or less.
		if js := listTexts(t, s.h, "/api/loften"+c.query); got != js {
			t.Errorf("%s: CSV %s = %q, JSON list = %q", c.name, c.query, got, js)
		}
	}
	if got := csvTexts("/api/parties/S/loften.csv?mandatperiod=2022-2026"); got != "Gammal arbetsrätt|Gammal skatt" {
		t.Errorf("party path with period = %q", got)
	}
	// Every row says which period it belongs to.
	_, rows := loftenCSV(t, s.h, "/api/loften.csv?mandatperiod=2022-2026")
	for _, r := range rows {
		if r["Mandatperiod"] != "2022-2026" {
			t.Errorf("row %s has Mandatperiod %q under the 2022-2026 filter", r["Id"], r["Mandatperiod"])
		}
	}
	// A malformed period is refused, as on the JSON list; never a full file.
	for _, bad := range []string{"2026", "nu", "2026-2030%3BDROP"} {
		rec := s.h.Do(httptest.NewRequest(http.MethodGet, "/api/loften.csv?mandatperiod="+bad, nil))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("mandatperiod=%s = %d, want 400", bad, rec.Code)
		}
	}
}
