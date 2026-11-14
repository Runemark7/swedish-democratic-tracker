package apitest

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

const lQuote = "Vi ska införa en ny regel i testet."

func vmID(t *testing.T, h *Harness, party string) int {
	t.Helper()
	var id int
	if err := h.DB.QueryRow(context.Background(),
		`SELECT id FROM valmanifest WHERE party=$1 ORDER BY id LIMIT 1`, party).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func proposeBody(vm int, quote, url string) string {
	return fmt.Sprintf(`{"party":"S","valmanifestId":%d,"text":"Regeln införs","quote":%q,"locator":"s. 12","sourceUrl":%q,"topic":"test","specificity":"concrete","utskott":["SkU"],"chatRef":"chat-9#1","onBehalfOf":"redaktor"}`, vm, quote, url)
}

func prov(ref string) string {
	return `{"chatRef":"` + ref + `","onBehalfOf":"redaktor"}`
}

func publicLoefteIDs(t *testing.T, h *Harness) string {
	t.Helper()
	rec := writeReq(h, "GET", "/api/loften?party=S", "", "", "")
	if rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	return rec.Body.String()
}

func TestLoefteProposeRejectsMissingQuoteOrSource(t *testing.T) {
	h := startWrite(t)
	vm := vmID(t, h, "S")
	good := proposeBody(vm, lQuote, "https://example.org/val.pdf#page=12")
	cases := map[string]string{
		"no quote":            proposeBody(vm, "  ", "https://example.org/x"),
		"no source":           proposeBody(vm, lQuote, ""),
		"bad source":          proposeBody(vm, lQuote, "javascript:alert(1)"),
		"no valmanifest":      proposeBody(0, lQuote, "https://example.org/x"),
		"unknown valmanifest": proposeBody(99999999, lQuote, "https://example.org/x"),
		"wrong party":         strings.Replace(proposeBody(vmID(t, h, "M"), lQuote, "https://example.org/x"), `"M"`, `"M"`, 1),
		"no chat ref":         strings.Replace(good, `"chatRef":"chat-9#1"`, `"chatRef":""`, 1),
	}
	// "wrong party": the body says S but the Valmanifest is M's.
	cases["wrong party"] = proposeBody(vmID(t, h, "M"), lQuote, "https://example.org/x")
	for name, b := range cases {
		rec := writeReq(h, "POST", "/api/write/loften", testWriteToken, "k-"+name, b)
		if rec.Code != 422 {
			t.Errorf("%s: %d %s", name, rec.Code, rec.Body)
		}
	}
	if n := count(t, h, `SELECT count(*) FROM loften WHERE visibility='proposed' AND id NOT BETWEEN 45 AND 51`); n != 0 {
		t.Fatalf("rejected proposals stored: %d", n)
	}
	if n := count(t, h, `SELECT count(*) FROM aendringslogg WHERE actor <> 'migration'`); n != 0 {
		t.Fatalf("rejected proposals logged: %d", n)
	}
	if rec := writeReq(h, "POST", "/api/write/loften", "", "k", good); rec.Code != 401 {
		t.Fatalf("no token: %d", rec.Code)
	}
}

func TestLoefteProposedInvisibleUntilApproved(t *testing.T) {
	h := startWrite(t)
	vm := vmID(t, h, "S")
	rec := writeReq(h, "POST", "/api/write/loften", testWriteToken, "p1", proposeBody(vm, lQuote, "https://example.org/val.pdf#page=12"))
	if rec.Code != 200 {
		t.Fatalf("propose: %d %s", rec.Code, rec.Body)
	}
	pr := decode[writeResult](t, rec)
	id := pr.Entry.EntityID
	idPath := "/api/loften/" + itoa(int(id))

	// Proposed: invisible in list, detail, party list and the public log.
	if strings.Contains(publicLoefteIDs(t, h), lQuote) {
		t.Fatal("proposed Löfte in public list")
	}
	if rec := writeReq(h, "GET", idPath, "", "", ""); rec.Code != 404 {
		t.Fatalf("proposed detail: %d", rec.Code)
	}
	if rec := writeReq(h, "GET", "/api/parties/S/loften", "", "", ""); strings.Contains(rec.Body.String(), lQuote) {
		t.Fatal("proposed in party list")
	}
	log := writeReq(h, "GET", "/api/aendringslogg", "", "", "").Body.String()
	if strings.Contains(log, lQuote) || strings.Contains(log, "Regeln införs") || strings.Contains(log, "val.pdf") {
		t.Fatalf("public log leaks a proposed Löfte: %s", log)
	}

	// An edit while proposed is also not public.
	rec = writeReq(h, "POST", "/api/write/loften/"+itoa(int(id)), testWriteToken, "e1",
		`{"text":"Regeln införs snabbt","quote":"`+lQuote+` Snabbt.","locator":"s. 13","sourceUrl":"https://example.org/val.pdf#page=13","topic":"test","specificity":"concrete","utskott":["SkU"],"chatRef":"chat-9#2","onBehalfOf":"redaktor"}`)
	if rec.Code != 200 || !decode[writeResult](t, rec).Changed {
		t.Fatalf("edit: %d %s", rec.Code, rec.Body)
	}
	if strings.Contains(writeReq(h, "GET", "/api/aendringslogg", "", "", "").Body.String(), "Snabbt") {
		t.Fatal("log leaks an edit of a proposed Löfte")
	}

	// Approve.
	rec = writeReq(h, "POST", "/api/write/loften/"+itoa(int(id))+"/approve", testWriteToken, "a1", prov("chat-9#3"))
	if rec.Code != 200 {
		t.Fatalf("approve: %d %s", rec.Code, rec.Body)
	}
	ap := decode[writeResult](t, rec)
	if !ap.Changed || ap.Entry.ChatRef != "chat-9#3" || ap.Entry.OnBehalf != "redaktor" || ap.Entry.Action != "upsert" {
		t.Fatalf("approval entry: %+v", ap.Entry)
	}

	det := writeReq(h, "GET", idPath, "", "", "")
	if det.Code != 200 {
		t.Fatalf("approved detail: %d", det.Code)
	}
	l := decode[struct {
		Quote, SourceURL, Locator, Visibility string
		ValmanifestID                         *int
	}](t, det)
	if l.Visibility != "approved" || !strings.Contains(l.Quote, "Snabbt") || l.SourceURL != "https://example.org/val.pdf#page=13" ||
		l.Locator != "s. 13" || l.ValmanifestID == nil || *l.ValmanifestID != vm {
		t.Fatalf("approved Löfte: %+v", l)
	}
	if !strings.Contains(publicLoefteIDs(t, h), "Snabbt") {
		t.Fatal("approved Löfte missing from list")
	}
	if n := count(t, h, `SELECT count(*) FROM loften WHERE id=$1 AND approved_at IS NOT NULL`, id); n != 1 {
		t.Fatal("approved_at not set")
	}

	// The approval is in the public Ändringslogg with the chat reference; the
	// earlier propose/edit entries are not.
	lg := decode[struct{ Entries []logEntry }](t, writeReq(h, "GET", "/api/aendringslogg", "", "", ""))
	if len(lg.Entries) != 1 {
		t.Fatalf("public log: %+v", lg.Entries)
	}
	e := lg.Entries[0]
	if e.Entity != "loefte" || e.EntityID != id || e.ChatRef != "chat-9#3" || e.ID != ap.Entry.ID {
		t.Fatalf("public entry: %+v", e)
	}
	if len(e.Old) != 0 && string(e.Old) != "null" {
		t.Fatalf("pre-approval value leaked: %s", e.Old)
	}
	if !strings.Contains(string(e.New), "approved") {
		t.Fatalf("new value: %s", e.New)
	}
	// Approve again: no-op, nothing logged.
	rec = writeReq(h, "POST", "/api/write/loften/"+itoa(int(id))+"/approve", testWriteToken, "a2", prov("chat-9#4"))
	if rec.Code != 200 || decode[writeResult](t, rec).Changed {
		t.Fatalf("re-approve: %d %s", rec.Code, rec.Body)
	}
	if n := count(t, h, `SELECT count(*) FROM aendringslogg WHERE actor <> 'migration'`); n != 3 {
		t.Fatalf("log rows: %d", n)
	}
}

func TestLoefteStrikeAndEditRules(t *testing.T) {
	h := startWrite(t)
	vm := vmID(t, h, "S")
	mk := func(key, quote string) int64 {
		rec := writeReq(h, "POST", "/api/write/loften", testWriteToken, key, proposeBody(vm, quote, "https://example.org/v"))
		if rec.Code != 200 {
			t.Fatalf("propose: %d %s", rec.Code, rec.Body)
		}
		return decode[writeResult](t, rec).Entry.EntityID
	}
	a, b := mk("s1", "Citat A"), mk("s2", "Citat B")

	// Same quote again is refused while the first is not struck.
	if rec := writeReq(h, "POST", "/api/write/loften", testWriteToken, "s3", proposeBody(vm, "Citat A", "https://example.org/v")); rec.Code != 409 {
		t.Fatalf("duplicate: %d", rec.Code)
	}

	// Strike a proposed Löfte: stays invisible, no public log entry.
	if rec := writeReq(h, "POST", "/api/write/loften/"+itoa(int(a))+"/strike", testWriteToken, "s4", prov("chat-9#5")); rec.Code != 200 {
		t.Fatalf("strike: %d %s", rec.Code, rec.Body)
	}
	if rec := writeReq(h, "GET", "/api/loften/"+itoa(int(a)), "", "", ""); rec.Code != 404 {
		t.Fatalf("struck is public: %d", rec.Code)
	}
	if lg := decode[struct{ Entries []logEntry }](t, writeReq(h, "GET", "/api/aendringslogg", "", "", "")); len(lg.Entries) != 0 {
		t.Fatalf("public log shows non-public changes: %+v", lg.Entries)
	}
	// A struck Löfte cannot be approved or edited.
	if rec := writeReq(h, "POST", "/api/write/loften/"+itoa(int(a))+"/approve", testWriteToken, "s5", prov("c")); rec.Code != 409 {
		t.Fatalf("approve struck: %d", rec.Code)
	}
	// After a strike the quote may be proposed again.
	if rec := writeReq(h, "POST", "/api/write/loften", testWriteToken, "s6", proposeBody(vm, "Citat A", "https://example.org/v")); rec.Code != 200 {
		t.Fatalf("re-propose after strike: %d %s", rec.Code, rec.Body)
	}

	// Edit cannot remove quote or source, and unknown ids are 404.
	if rec := writeReq(h, "POST", "/api/write/loften/"+itoa(int(b)), testWriteToken, "s7",
		`{"text":"x","quote":"","sourceUrl":"https://e.org","chatRef":"c","onBehalfOf":"r"}`); rec.Code != 422 {
		t.Fatalf("edit without quote: %d", rec.Code)
	}
	if rec := writeReq(h, "POST", "/api/write/loften/987654/approve", testWriteToken, "s8", prov("c")); rec.Code != 404 {
		t.Fatalf("unknown: %d", rec.Code)
	}

	// Strike an approved Löfte: it leaves the public list; the log shows the
	// approval and the strike, the strike with the struck value withheld.
	writeReq(h, "POST", "/api/write/loften/"+itoa(int(b))+"/approve", testWriteToken, "s9", prov("chat-9#6"))
	if rec := writeReq(h, "POST", "/api/write/loften/"+itoa(int(b))+"/strike", testWriteToken, "s10", prov("chat-9#7")); rec.Code != 200 {
		t.Fatalf("strike approved: %d", rec.Code)
	}
	if strings.Contains(publicLoefteIDs(t, h), "Citat B") {
		t.Fatal("struck Löfte still public")
	}
	lg := decode[struct{ Entries []logEntry }](t, writeReq(h, "GET", "/api/aendringslogg", "", "", ""))
	if len(lg.Entries) != 2 {
		t.Fatalf("public log: %+v", lg.Entries)
	}
	if strings.Contains(string(lg.Entries[0].New), "struck") && len(lg.Entries[0].New) > 4 {
		t.Fatalf("struck value not redacted: %s", lg.Entries[0].New)
	}
}

func TestLoefteUndoAndIdempotency(t *testing.T) {
	h := startWrite(t)
	vm := vmID(t, h, "S")
	b := proposeBody(vm, "Citat U", "https://example.org/v")
	r1 := decode[writeResult](t, writeReq(h, "POST", "/api/write/loften", testWriteToken, "i1", b))
	r2 := decode[writeResult](t, writeReq(h, "POST", "/api/write/loften", testWriteToken, "i1", b))
	if r1.Entry.ID != r2.Entry.ID || count(t, h, `SELECT count(*) FROM loften WHERE visibility='proposed' AND id NOT BETWEEN 45 AND 51`) != 1 {
		t.Fatal("propose not idempotent")
	}
	id := r1.Entry.EntityID
	ap := decode[writeResult](t, writeReq(h, "POST", "/api/write/loften/"+itoa(int(id))+"/approve", testWriteToken, "i2", prov("c1")))
	undo := "/api/write/aendringslogg/" + itoa(int(ap.Entry.ID)) + "/undo"
	if rec := writeReq(h, "POST", undo, testWriteToken, "i3", prov("c2")); rec.Code != 200 {
		t.Fatalf("undo approve: %d %s", rec.Code, rec.Body)
	}
	if rec := writeReq(h, "GET", "/api/loften/"+itoa(int(id)), "", "", ""); rec.Code != 404 {
		t.Fatalf("undone approval still public: %d", rec.Code)
	}
	if n := count(t, h, `SELECT count(*) FROM loften WHERE id=$1 AND approved_at IS NULL AND visibility='proposed'`, id); n != 1 {
		t.Fatal("not back to proposed")
	}
	// Undo of the propose removes the row.
	undoP := "/api/write/aendringslogg/" + itoa(int(r1.Entry.ID)) + "/undo"
	if rec := writeReq(h, "POST", undoP, testWriteToken, "i4", prov("c3")); rec.Code != 200 {
		t.Fatalf("undo propose: %d %s", rec.Code, rec.Body)
	}
	if n := count(t, h, `SELECT count(*) FROM loften WHERE id=$1`, id); n != 0 {
		t.Fatal("proposal not removed")
	}
}

// Archived Löften: only quote, locator (källdokument) and sourceUrl may change.
func TestArchivedLoefteEditsQuoteAndSourceOnly(t *testing.T) {
	h := startWrite(t)
	vm := vmID(t, h, "S")
	rec := writeReq(h, "POST", "/api/write/loften", testWriteToken, "a1", proposeBody(vm, "Gammalt citat", "https://example.org/v"))
	if rec.Code != 200 {
		t.Fatalf("propose: %d %s", rec.Code, rec.Body)
	}
	id := decode[writeResult](t, rec).Entry.EntityID
	path := "/api/write/loften/" + itoa(int(id))
	writeReq(h, "POST", path+"/approve", testWriteToken, "a2", prov("c#1"))
	if _, err := h.DB.Exec(context.Background(), `UPDATE loften SET visibility='archived' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	edit := func(key, text, quote, locator, url string) int {
		return writeReq(h, "POST", path, testWriteToken, key, fmt.Sprintf(
			`{"text":%q,"quote":%q,"locator":%q,"sourceUrl":%q,"topic":"test","specificity":"concrete","utskott":["SkU"],"chatRef":"c#2","onBehalfOf":"redaktor"}`,
			text, quote, locator, url)).Code
	}
	before := len(decode[struct{ Entries []logEntry }](t, writeReq(h, "GET", "/api/aendringslogg", "", "", "")).Entries)

	if c := edit("a3", "Regeln införs", "Nytt citat", "s. 99", "https://example.org/ny"); c != 200 {
		t.Fatalf("archived edit of quote/source: %d", c)
	}
	var quote, loc, url, vis string
	if err := h.DB.QueryRow(context.Background(), `SELECT quote, coalesce(locator,''), source_url, visibility FROM loften WHERE id=$1`, id).
		Scan(&quote, &loc, &url, &vis); err != nil {
		t.Fatal(err)
	}
	if quote != "Nytt citat" || loc != "s. 99" || url != "https://example.org/ny" || vis != "archived" {
		t.Fatalf("row: %q %q %q %q", quote, loc, url, vis)
	}
	after := decode[struct{ Entries []logEntry }](t, writeReq(h, "GET", "/api/aendringslogg", "", "", "")).Entries
	if len(after) != before+1 {
		t.Fatalf("public log: %d -> %d", before, len(after))
	}

	// Validation as for other edits.
	if c := edit("a4", "Regeln införs", "", "s. 99", "https://example.org/ny"); c != 422 {
		t.Fatalf("empty quote: %d", c)
	}
	if c := edit("a5", "Regeln införs", "Q", "s. 99", "not-a-url"); c != 422 {
		t.Fatalf("bad url: %d", c)
	}
	// Text stays locked.
	if c := edit("a6", "Annan text", "Nytt citat", "s. 99", "https://example.org/ny"); c != 409 {
		t.Fatalf("archived text edit: %d", c)
	}
	// Party/status are not editable fields; visibility is only moved by approve/strike.
	for _, p := range []string{"/approve", "/strike"} {
		if c := writeReq(h, "POST", path+p, testWriteToken, "a7"+p, prov("c")).Code; c != 409 {
			t.Fatalf("archived %s: %d", p, c)
		}
	}
	// Non-archived Löften keep full edit.
	rec = writeReq(h, "POST", "/api/write/loften", testWriteToken, "a8", proposeBody(vm, "Annat citat", "https://example.org/v"))
	id2 := decode[writeResult](t, rec).Entry.EntityID
	if c := writeReq(h, "POST", "/api/write/loften/"+itoa(int(id2)), testWriteToken, "a9",
		`{"text":"Ny text","quote":"Annat citat","sourceUrl":"https://example.org/v","chatRef":"c","onBehalfOf":"r"}`).Code; c != 200 {
		t.Fatalf("proposed edit: %d", c)
	}
}
