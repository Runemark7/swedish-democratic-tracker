package seeder

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"riksdagskollen/internal/scbpx"
)

// Reproduces the prod error of 2026-11-11 ("fetch populations: SCB population
// returned 400 for batch starting 0114"): every region code in the batch was
// valid; SCB answered the valid query with 400 (v1) / 500 (v2) intermittently.
// The fetch now goes to PxWebApi v2 with retries and falls back to v1.
func TestFetchPopulations_SurvivesFlakySCB(t *testing.T) {
	v1, err := os.ReadFile("../../scbpx/testdata/befolkning_v1_befolkningny.json")
	if err != nil {
		t.Fatal(err)
	}
	calls := map[string]int{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls[r.URL.Path]++
		switch r.URL.Path {
		case "/v2/tables/TAB638/data":
			// v2 down: first batch exhausts its retries, then falls back to
			// v1 (400 once, then OK); the second batch goes to v1 first.
			w.WriteHeader(http.StatusInternalServerError)
		case "/v1/BE/BE0101/BE0101A/BefolkningNy":
			if calls[r.URL.Path] == 1 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Write(v1)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c := scbpx.NewWithBaseURLs(srv.URL+"/v2", srv.URL+"/v1", srv.Client())
	c.SetRetryDelay(time.Millisecond)

	// populationBatchSize+5 codes -> two batches.
	codes := []string{"0114", "0115", "0117", "01", "09"}
	for len(codes) < populationBatchSize+5 {
		codes = append(codes, "0114")
	}
	got, err := fetchPopulations(context.Background(), c, codes)
	if err != nil {
		t.Fatalf("fetchPopulations: %v", err)
	}
	want := map[string]int{"0114": 50110, "0115": 35076, "0117": 49282, "01": 2454821, "09": 61029}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %d, want %d", k, got[k], v)
		}
	}
	if calls["/v1/BE/BE0101/BE0101A/BefolkningNy"] != 3 || calls["/v2/tables/TAB638/data"] != 5 {
		t.Errorf("calls = %v", calls)
	}
}
