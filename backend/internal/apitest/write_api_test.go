package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testWriteToken = "test-write-token-not-a-secret"

func startWrite(t *testing.T) *Harness {
	t.Helper()
	return StartWithWriteToken(t, chainDir(t, "beslut-list"), testWriteToken)
}

func writeReq(h *Harness, method, path, token, key, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	return h.Do(req)
}

func manifestBody(url, state string) string {
	return `{"party":"M","electionYear":2030,"title":"Moderaternas valmanifest 2030","url":"` + url + `","state":"` + state + `","chatRef":"chat-1#42","onBehalfOf":"redaktor"}`
}

type logEntry struct {
	ID        int64           `json:"id"`
	Action    string          `json:"action"`
	Entity    string          `json:"entity"`
	EntityID  int64           `json:"entityId"`
	Actor     string          `json:"actor"`
	OnBehalf  string          `json:"onBehalfOf"`
	ChatRef   string          `json:"chatRef"`
	Old       json.RawMessage `json:"oldValue"`
	New       json.RawMessage `json:"newValue"`
	Undoes    *int64          `json:"undoes"`
	CreatedAt string          `json:"createdAt"`
}

type writeResult struct {
	Changed bool      `json:"changed"`
	Entry   *logEntry `json:"entry"`
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return v
}

func count(t *testing.T, h *Harness, q string, args ...any) int {
	t.Helper()
	var n int
	if err := h.DB.QueryRow(context.Background(), q, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestWriteAPIAuth(t *testing.T) {
	h := startWrite(t)
	const p = "/api/write/valmanifest"
	b := manifestBody("https://example.org/a", "draft")

	if rec := writeReq(h, "POST", p, "", "k1", b); rec.Code != 401 {
		t.Fatalf("no token: %d", rec.Code)
	}
	if rec := writeReq(h, "POST", p, "wrong", "k1", b); rec.Code != 401 {
		t.Fatalf("wrong token: %d", rec.Code)
	}
	if n := count(t, h, `SELECT count(*) FROM aendringslogg`); n != 0 {
		t.Fatalf("unauthorised calls logged %d entries", n)
	}
	if rec := writeReq(h, "POST", p, testWriteToken, "k1", b); rec.Code != 200 {
		t.Fatalf("right token: %d %s", rec.Code, rec.Body)
	}
}

func TestWriteAPIUnconfiguredIs404(t *testing.T) {
	h := Start(t, chainDir(t, "beslut-list"))
	for _, tok := range []string{"", "anything", testWriteToken} {
		for _, path := range []string{"/api/write/valmanifest", "/api/write/aendringslogg/1/undo"} {
			if rec := writeReq(h, "POST", path, tok, "k", manifestBody("https://e.org", "draft")); rec.Code != 404 {
				t.Fatalf("%s token=%q: %d, want 404", path, tok, rec.Code)
			}
		}
	}
	if n := count(t, h, `SELECT count(*) FROM aendringslogg`); n != 0 {
		t.Fatal("log written")
	}
	// The log is still readable.
	if rec := writeReq(h, "GET", "/api/aendringslogg", "", "", ""); rec.Code != 200 {
		t.Fatalf("read log: %d", rec.Code)
	}
}

func TestWriteAPIValidation(t *testing.T) {
	h := startWrite(t)
	const p = "/api/write/valmanifest"
	good := manifestBody("https://example.org/a", "draft")
	cases := map[string]struct{ key, body string }{
		"no key":        {"", good},
		"bad state":     {"k", manifestBody("https://example.org/a", "final2")},
		"bad url":       {"k", manifestBody("javascript:alert(1)", "draft")},
		"unknown party": {"k", strings.Replace(good, `"M"`, `"ZZ"`, 1)},
		"no chat ref":   {"k", strings.Replace(good, `"chatRef":"chat-1#42"`, `"chatRef":""`, 1)},
		"unknown field": {"k", strings.Replace(good, `{`, `{"bogus":1,`, 1)},
		"not json":      {"k", `nope`},
	}
	for name, c := range cases {
		rec := writeReq(h, "POST", p, testWriteToken, c.key, c.body)
		if rec.Code != 400 && rec.Code != 422 {
			t.Errorf("%s: %d %s", name, rec.Code, rec.Body)
		}
		var e struct{ Error, Code string }
		_ = json.Unmarshal(rec.Body.Bytes(), &e)
		if e.Error == "" || e.Code == "" {
			t.Errorf("%s: error body %q", name, rec.Body)
		}
	}
	if n := count(t, h, `SELECT count(*) FROM aendringslogg`); n != 0 {
		t.Fatal("invalid requests logged")
	}
}

func TestWriteAPIIdempotentAndLogged(t *testing.T) {
	h := startWrite(t)
	const p = "/api/write/valmanifest"
	b := manifestBody("https://example.org/a", "draft")

	first := writeReq(h, "POST", p, testWriteToken, "req-1", b)
	if first.Code != 200 {
		t.Fatalf("%d %s", first.Code, first.Body)
	}
	r1 := decode[writeResult](t, first)
	if !r1.Changed || r1.Entry == nil || r1.Entry.Old != nil && string(r1.Entry.Old) != "null" {
		t.Fatalf("first result %+v", r1)
	}
	if r1.Entry.Actor == "" || r1.Entry.OnBehalf != "redaktor" || r1.Entry.ChatRef != "chat-1#42" || r1.Entry.CreatedAt == "" {
		t.Fatalf("entry lacks who/when/chat: %+v", r1.Entry)
	}

	again := writeReq(h, "POST", p, testWriteToken, "req-1", b)
	if again.Code != 200 {
		t.Fatalf("replay: %d", again.Code)
	}
	r2 := decode[writeResult](t, again)
	if r2.Entry == nil || r2.Entry.ID != r1.Entry.ID {
		t.Fatalf("replay returned a different entry: %+v", r2)
	}
	if n := count(t, h, `SELECT count(*) FROM valmanifest WHERE party='M' AND election_year=2030`); n != 1 {
		t.Fatalf("manifests: %d", n)
	}
	if n := count(t, h, `SELECT count(*) FROM aendringslogg`); n != 1 {
		t.Fatalf("log entries after replay: %d", n)
	}

	// Same key, different content: refused, nothing changes.
	rec := writeReq(h, "POST", p, testWriteToken, "req-1", manifestBody("https://example.org/other", "draft"))
	if rec.Code != 409 {
		t.Fatalf("key reuse: %d", rec.Code)
	}

	// A new key with identical content changes nothing and logs nothing.
	rec = writeReq(h, "POST", p, testWriteToken, "req-2", b)
	if rec.Code != 200 || decode[writeResult](t, rec).Changed {
		t.Fatalf("no-op: %d %s", rec.Code, rec.Body)
	}
	if n := count(t, h, `SELECT count(*) FROM aendringslogg`); n != 1 {
		t.Fatalf("no-op logged: %d", n)
	}

	// An update records old and new.
	rec = writeReq(h, "POST", p, testWriteToken, "req-3", manifestBody("https://example.org/b", "final"))
	r3 := decode[writeResult](t, rec)
	if !r3.Changed || !strings.Contains(string(r3.Entry.Old), "example.org/a") || !strings.Contains(string(r3.Entry.New), "example.org/b") {
		t.Fatalf("update entry: %+v", r3.Entry)
	}
}

func TestWriteAPILogIsAtomicAndAppendOnly(t *testing.T) {
	h := startWrite(t)
	ctx := context.Background()
	rec := writeReq(h, "POST", "/api/write/valmanifest", testWriteToken, "k", manifestBody("https://example.org/a", "draft"))
	if rec.Code != 200 {
		t.Fatal(rec.Body)
	}
	if _, err := h.DB.Exec(ctx, `UPDATE aendringslogg SET actor='x'`); err == nil {
		t.Fatal("update of the log succeeded")
	}
	if _, err := h.DB.Exec(ctx, `DELETE FROM aendringslogg`); err == nil {
		t.Fatal("delete from the log succeeded")
	}
	if _, err := h.DB.Exec(ctx, `TRUNCATE aendringslogg`); err == nil {
		t.Fatal("truncate of the log succeeded")
	}
	// Every change has its entry: a valmanifest the API wrote is in the log.
	if n := count(t, h, `SELECT count(*) FROM aendringslogg WHERE entity='valmanifest' AND new_value->>'title'='Moderaternas valmanifest 2030'`); n != 1 {
		t.Fatalf("entries: %d", n)
	}
}

func TestWriteAPIUndo(t *testing.T) {
	h := startWrite(t)
	const p = "/api/write/valmanifest"
	created := decode[writeResult](t, writeReq(h, "POST", p, testWriteToken, "u1", manifestBody("https://example.org/a", "draft")))
	updated := decode[writeResult](t, writeReq(h, "POST", p, testWriteToken, "u2", manifestBody("https://example.org/b", "final")))

	undoPath := func(id int64) string { return "/api/write/aendringslogg/" + itoa(int(id)) + "/undo" }
	ub := `{"chatRef":"chat-1#50","onBehalfOf":"redaktor"}`

	// Undoing the first (create) now conflicts: the value changed since.
	if rec := writeReq(h, "POST", undoPath(created.Entry.ID), testWriteToken, "u3", ub); rec.Code != 409 {
		t.Fatalf("conflicting undo: %d %s", rec.Code, rec.Body)
	}
	if n := count(t, h, `SELECT count(*) FROM aendringslogg`); n != 2 {
		t.Fatalf("conflicting undo logged: %d", n)
	}

	// Undo of the update restores the old value and is itself logged.
	rec := writeReq(h, "POST", undoPath(updated.Entry.ID), testWriteToken, "u4", ub)
	if rec.Code != 200 {
		t.Fatalf("undo: %d %s", rec.Code, rec.Body)
	}
	un := decode[writeResult](t, rec)
	if un.Entry == nil || un.Entry.Action != "undo" || un.Entry.Undoes == nil || *un.Entry.Undoes != updated.Entry.ID {
		t.Fatalf("undo entry: %+v", un.Entry)
	}
	var url, state string
	if err := h.DB.QueryRow(context.Background(), `SELECT url, state FROM valmanifest WHERE party='M' AND election_year=2030`).Scan(&url, &state); err != nil {
		t.Fatal(err)
	}
	if url != "https://example.org/a" || state != "draft" {
		t.Fatalf("not restored: %s %s", url, state)
	}
	// Repeating the undo with the same key has no extra effect.
	if rec := writeReq(h, "POST", undoPath(updated.Entry.ID), testWriteToken, "u4", ub); rec.Code != 200 {
		t.Fatalf("undo replay: %d", rec.Code)
	}
	if n := count(t, h, `SELECT count(*) FROM aendringslogg`); n != 3 {
		t.Fatalf("log entries: %d", n)
	}
	// Now the create can be undone: the row goes away, and that is logged.
	if rec := writeReq(h, "POST", undoPath(created.Entry.ID), testWriteToken, "u5", ub); rec.Code != 200 {
		t.Fatalf("undo create: %d %s", rec.Code, rec.Body)
	}
	if n := count(t, h, `SELECT count(*) FROM valmanifest WHERE party='M' AND election_year=2030`); n != 0 {
		t.Fatal("create not undone")
	}
	// Unknown entry.
	if rec := writeReq(h, "POST", undoPath(99999), testWriteToken, "u6", ub); rec.Code != 404 {
		t.Fatalf("unknown entry: %d", rec.Code)
	}
}

func TestAendringsloggReadable(t *testing.T) {
	h := startWrite(t)
	const p = "/api/write/valmanifest"
	writeReq(h, "POST", p, testWriteToken, "r1", manifestBody("https://example.org/a", "draft"))
	writeReq(h, "POST", p, testWriteToken, "r2", manifestBody("https://example.org/b", "final"))

	rec := writeReq(h, "GET", "/api/aendringslogg", "", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("%d", rec.Code)
	}
	out := decode[struct{ Entries []logEntry }](t, rec)
	if len(out.Entries) != 2 || out.Entries[0].ID < out.Entries[1].ID {
		t.Fatalf("want newest first, got %+v", out.Entries)
	}
	if strings.Contains(rec.Body.String(), testWriteToken) || strings.Contains(rec.Body.String(), "r1") {
		t.Fatal("log leaks token or request keys")
	}
	rec = writeReq(h, "GET", "/api/aendringslogg?limit=1", "", "", "")
	if got := decode[struct{ Entries []logEntry }](t, rec); len(got.Entries) != 1 {
		t.Fatalf("limit: %d", len(got.Entries))
	}
}
