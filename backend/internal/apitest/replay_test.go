package apitest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFixture(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReplayServesRecordedResponseByPathAndQuery(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "voteringlista.json", `{
		"method": "GET",
		"path": "/voteringlista/",
		"query": {"rm": "2025/26", "bet": "UbU30", "sz": "10000", "utformat": "json"},
		"status": 200,
		"body": {"voteringlista": {"votering": [{"votering_id": "v1"}]}}
	}`)

	var missed []string
	h := mustHandler(t, dir, func(format string, args ...any) {
		missed = append(missed, fmt.Sprintf(format, args...))
	})

	// Parameter order differs from the fixture. The key is path + query, not
	// the order the client happened to write.
	req := httptest.NewRequest(http.MethodGet, "/voteringlista/?utformat=json&sz=10000&bet=UbU30&rm=2025/26", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if len(missed) != 0 {
		t.Fatalf("recorded request was treated as a miss: %v", missed)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"votering_id":"v1"`) && !strings.Contains(rec.Body.String(), `"votering_id": "v1"`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q", got)
	}
}

func TestReplayFailsLoudlyOnUnrecordedRequest(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "one.json", `{
		"method": "GET",
		"path": "/voteringlista/",
		"query": {"rm": "2025/26"},
		"body": {"ok": true}
	}`)

	var missed []string
	h := mustHandler(t, dir, func(format string, args ...any) {
		missed = append(missed, fmt.Sprintf(format, args...))
	})

	req := httptest.NewRequest(http.MethodGet, "/voteringlista/?rm=2024/25&bet=AU1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		t.Fatal("unrecorded request returned 200")
	}
	if len(missed) != 1 {
		t.Fatalf("miss reports = %d, want 1 (%v)", len(missed), missed)
	}
	if !strings.Contains(missed[0], "/voteringlista/") || !strings.Contains(missed[0], "rm=2024") {
		t.Fatalf("miss report %q does not name the request", missed[0])
	}
	if !strings.Contains(rec.Body.String(), "/voteringlista/") {
		t.Fatalf("response body %q does not name the request", rec.Body.String())
	}
}

func TestReplayRejectsDuplicateFixtureKeys(t *testing.T) {
	dir := t.TempDir()
	body := `{
		"method": "GET",
		"path": "/personlista/",
		"query": {"parti": ""},
		"body": {"personlista": {}}
	}`
	writeFixture(t, dir, "a.json", body)
	writeFixture(t, dir, "b.json", body)

	if _, err := loadReplay(dir, func(string, ...any) {}); err == nil {
		t.Fatal("duplicate fixture key was accepted")
	}
}

func mustHandler(t *testing.T, dir string, onMiss func(string, ...any)) http.Handler {
	t.Helper()
	h, err := loadReplay(dir, onMiss)
	if err != nil {
		t.Fatal(err)
	}
	return h
}
