package riksdagen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// testdata/bet-list-2002-03-UBU15.json is the /dokumentlista row of
// 2002/03:UBU15 as Riksdagen served it on 2026-11-11 (Debatt 2003-05-28).
func TestFindBetankandeReadsTheDebateDateFromTheStatusrad(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "bet-list-2002-03-UBU15.json"))
	if err != nil {
		t.Fatal(err)
	}
	var body json.RawMessage = raw
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	ref, found, err := NewClientWithBaseURL(srv.URL).FindBetankande(context.Background(), "2002/03", "UBU15")
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	if want := time.Date(2003, 5, 28, 0, 0, 0, 0, time.UTC); !ref.DebateDate.Equal(want) {
		t.Errorf("debate date = %v, want %v", ref.DebateDate, want)
	}
	if ref.Title != "Fristående skolor" {
		t.Errorf("title = %q", ref.Title)
	}
}

func TestParseDebateDateWithoutADebateIsZero(t *testing.T) {
	if got := parseDebateDate("<dl><dt>Justering</dt> <dd>2018-03-08</dd></dl>"); !got.IsZero() {
		t.Errorf("got %v", got)
	}
}
