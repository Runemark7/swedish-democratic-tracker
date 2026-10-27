package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"riksdagskollen/internal/ingestion"
)

type fakeSource struct {
	snap ingestion.Snapshot
	err  error
}

func (f fakeSource) Snapshot(context.Context) (ingestion.Snapshot, error) {
	return f.snap, f.err
}

func TestServeFreshness_ReportsStaleSpeechAndNeverRunJob(t *testing.T) {
	now := time.Date(2026, 10, 27, 6, 0, 0, 0, time.UTC)
	speech := time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)
	vote := time.Date(2026, 10, 15, 12, 52, 14, 0, time.UTC)
	h := NewHandler(fakeSource{snap: ingestion.Snapshot{
		SpeechesLatest: &speech,
		VotesLatest:    &vote,
	}}, []ingestion.ScheduledJob{{Name: "fetch-speeches", Spec: "@daily"}})
	h.now = func() time.Time { return now }

	rec := httptest.NewRecorder()
	h.ServeFreshness(rec, httptest.NewRequest(http.MethodGet, "/api/ingestion/freshness", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}

	var body freshnessJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Speeches.LatestDate == nil || *body.Speeches.LatestDate != "2026-08-13" {
		t.Errorf("speeches = %+v", body.Speeches)
	}
	if body.Votes.LatestDate == nil || *body.Votes.LatestDate != "2026-10-15" {
		t.Errorf("votes = %+v", body.Votes)
	}
	if len(body.Jobs) != 1 || body.Jobs[0].Name != "fetch-speeches" || body.Jobs[0].LastSuccessAt != nil {
		t.Fatalf("jobs = %+v", body.Jobs)
	}
	if len(body.Alerts) < 2 {
		t.Fatalf("alerts = %v, want the stale speech and the job that never succeeded", body.Alerts)
	}
}

func TestServeFreshness_SourceErrorIsJSON(t *testing.T) {
	h := NewHandler(fakeSource{err: context.DeadlineExceeded}, nil)
	rec := httptest.NewRecorder()
	h.ServeFreshness(rec, httptest.NewRequest(http.MethodGet, "/api/ingestion/freshness", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"] == "" {
		t.Fatalf("body = %v", body)
	}
}
