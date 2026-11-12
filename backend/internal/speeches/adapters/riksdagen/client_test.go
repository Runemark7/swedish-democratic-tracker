package riksdagen

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFetchSpeechTextJSON(t *testing.T) {
	body := readTestdata(t, "H309127.json")
	srv, got := startSpeechServer(t, http.StatusOK, "application/json", body)
	text, err := NewClientWithBaseURL(srv.URL).FetchSpeechText(context.Background(), "H309127", "93")
	if err != nil {
		t.Fatalf("FetchSpeechText: %v", err)
	}
	if !strings.Contains(text, "övergödning och döda bottnar i Östersjön") {
		t.Fatalf("expected recorded JSON speech text, got %q", trunc(text))
	}
	if *got != "/anforande/H309127-93.json" {
		t.Fatalf("path %s", *got)
	}
}

func TestFetchSpeechTextXMLBodyExtractsText(t *testing.T) {
	// Recorded 2026-11-12 from https://data.riksdagen.se/anforande/HD09100-8.json
	// which returned 200 with an <anforande> XML document instead of JSON.
	body := readTestdata(t, "HD09100-8.xml")
	srv, _ := startSpeechServer(t, http.StatusOK, "text/xml", body)
	text, err := NewClientWithBaseURL(srv.URL).FetchSpeechText(context.Background(), "HD09100", "8")
	if err != nil {
		t.Fatalf("FetchSpeechText: %v", err)
	}
	if !strings.Contains(text, "Azra Muranovic har frågat mig") {
		t.Fatalf("expected text extracted from recorded XML, got %q", trunc(text))
	}
	if !strings.Contains(text, "<p>") {
		t.Fatalf("XML entities should be unescaped, got %q", trunc(text))
	}
}

func TestFetchSpeechTextHTMLUnavailable(t *testing.T) {
	// Synthetic error page (not a live recording): a 200 HTML body with no speech.
	html := []byte("<!DOCTYPE html><html><head><title>Error</title></head><body><h1>Not found</h1></body></html>")
	srv, _ := startSpeechServer(t, http.StatusOK, "text/html", html)
	text, err := NewClientWithBaseURL(srv.URL).FetchSpeechText(context.Background(), "HD09100", "8")
	if err != nil {
		t.Fatalf("HTML error page must be unavailable, not a retryable error: %v", err)
	}
	if text != "" {
		t.Fatalf("expected empty text for unavailable HTML, got %q", text)
	}
}

func TestFetchSpeechTextNotFound(t *testing.T) {
	srv, _ := startSpeechServer(t, http.StatusNotFound, "text/plain", []byte("nope"))
	text, err := NewClientWithBaseURL(srv.URL).FetchSpeechText(context.Background(), "X", "1")
	if err != nil {
		t.Fatalf("404: %v", err)
	}
	if text != "" {
		t.Fatalf("expected empty, got %q", text)
	}
}

func TestFetchSpeechTextTransientStatus(t *testing.T) {
	srv, _ := startSpeechServer(t, http.StatusInternalServerError, "text/plain", []byte("boom"))
	_, err := NewClientWithBaseURL(srv.URL).FetchSpeechText(context.Background(), "X", "1")
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("error should mention status, got %v", err)
	}
}

func TestFetchSpeechTextGarbageIsError(t *testing.T) {
	srv, _ := startSpeechServer(t, http.StatusOK, "text/plain", []byte("not json and not markup"))
	_, err := NewClientWithBaseURL(srv.URL).FetchSpeechText(context.Background(), "X", "1")
	if err == nil {
		t.Fatal("expected decode error for garbage")
	}
	if !strings.Contains(err.Error(), "decode speech text") {
		t.Fatalf("got %v", err)
	}
}

func startSpeechServer(t *testing.T, status int, contentType string, body []byte) (*httptest.Server, *string) {
	t.Helper()
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Path
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv, &got
}

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func trunc(s string) string {
	if len(s) > 180 {
		return s[:180] + "…"
	}
	return s
}

