package riksdagen

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// testdata/mot-2004-05-MJ7-GS02MJ7.json is /dokumentstatus/GS02MJ7.json as
// Riksdagen served it on 2026-11-11, with the wordings cut to 200 characters.
// Yrkande 5 is listed five times, once per betänkande it was referred to
// (TU17, AU6, SfU10, FiU21, MJU9). Storing all five rows broke the
// (dok_id, nummer) key and the backfill skipped the whole motion.
func TestFetchMotionKeepsOneYrkandePerNumber(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "mot-2004-05-MJ7-GS02MJ7.json"))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(raw)
	}))
	defer srv.Close()

	m, err := NewClientWithBaseURL(srv.URL).FetchMotion(context.Background(), "GS02MJ7")
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Yrkanden) != 6 {
		t.Fatalf("got %d yrkanden, want 6 (numbers 1-6)", len(m.Yrkanden))
	}
	for i, y := range m.Yrkanden {
		if y.Number != i+1 {
			t.Errorf("yrkande %d has number %d, want %d", i, y.Number, i+1)
		}
	}
	if got := m.Yrkanden[4].BehandlasI; got != "2004/05:TU17" {
		t.Errorf("yrkande 5 behandlas i %q, want the first listed, 2004/05:TU17", got)
	}
	if m.Yrkanden[4].Wording == "" || m.Yrkanden[4].ChamberDecision != "Avslag" {
		t.Errorf("yrkande 5 = %+v, want its wording and the chamber's Avslag", m.Yrkanden[4])
	}
}

func TestYrkandenFromFillsEmptyFieldsFromALaterDuplicate(t *testing.T) {
	got := yrkandenFrom([]forslag{
		{Nummer: "1", Lydelse: "Riksdagen begär", BehandlasI: "2004/05:AU6"},
		{Nummer: "1", Lydelse: "Riksdagen begär", Kammaren: "Avslag", BehandlasI: "2004/05:SfU10"},
		{Nummer: "2", Lydelse: "Andra"},
	})
	if len(got) != 2 {
		t.Fatalf("got %d yrkanden, want 2", len(got))
	}
	if got[0].BehandlasI != "2004/05:AU6" || got[0].ChamberDecision != "Avslag" {
		t.Errorf("yrkande 1 = %+v, want behandlas_i of the first row and the decision of the second", got[0])
	}
}
