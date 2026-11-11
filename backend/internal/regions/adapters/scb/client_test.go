package scb_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"riksdagskollen/internal/regions/adapters/scb"
)

// scbResponse builds the JSON payload the SCB API returns.
func scbResponse(rows []struct {
	Key    []string
	Values []string
}) []byte {
	type row struct {
		Key    []string `json:"key"`
		Values []string `json:"values"`
	}
	payload := struct {
		Data []row `json:"data"`
	}{}
	for _, r := range rows {
		payload.Data = append(payload.Data, row{Key: r.Key, Values: r.Values})
	}
	b, _ := json.Marshal(payload)
	return b
}

// newTestClient creates a Client pointing at a test server.
// It replaces the unexported http field via the exported constructor + URL swap.
// Since the URL constants are package-level vars we can't swap them here,
// so we expose a constructor that accepts a base URL via a helper factory.
func newTestClient(srv *httptest.Server) *scb.Client {
	return scb.NewClientWithHTTP(srv.URL, &http.Client{})
}

// --- FetchRegionBudgetMultiYear ---

func TestFetchRegionBudgetMultiYear_Basic(t *testing.T) {
	// Fixture: region "06L", 2 years (2022, 2023), 2 areas + total
	rows := []struct {
		Key    []string
		Values []string
	}{
		// year 2022
		{Key: []string{"06L", "0", "2022"}, Values: []string{"1000"}}, // Primärvård
		{Key: []string{"06L", "1", "2022"}, Values: []string{"2000"}}, // Specialiserad somatisk vård
		{Key: []string{"06L", "0-9", "2022"}, Values: []string{"5000"}},
		// year 2023
		{Key: []string{"06L", "0", "2023"}, Values: []string{"1100"}},
		{Key: []string{"06L", "1", "2023"}, Values: []string{"2200"}},
		{Key: []string{"06L", "0-9", "2023"}, Values: []string{"6000"}},
		// ".." should be skipped
		{Key: []string{"06L", "2", "2022"}, Values: []string{".."}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(scbResponse(rows))
	}))
	defer srv.Close()

	cl := newTestClient(srv)
	snapshots, err := cl.FetchRegionBudgetMultiYear(context.Background(), "06", []int{2022, 2023})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect 2 areas × 2 years = 4 snapshots (area "2" is skipped due to "..")
	if len(snapshots) != 4 {
		t.Fatalf("expected 4 snapshots, got %d", len(snapshots))
	}

	// Check one known value
	var found bool
	for _, s := range snapshots {
		if s.Year == 2022 && s.AreaName == "Primärvård" {
			found = true
			if s.ValueMnkr != 1000 {
				t.Errorf("expected ValueMnkr=1000, got %f", s.ValueMnkr)
			}
			if s.TotalMnkr != 5000 {
				t.Errorf("expected TotalMnkr=5000, got %f", s.TotalMnkr)
			}
			if s.Pct != 20.0 {
				t.Errorf("expected Pct=20.0, got %f", s.Pct)
			}
			if s.RegionCode != "06" {
				t.Errorf("expected RegionCode=06, got %s", s.RegionCode)
			}
		}
	}
	if !found {
		t.Error("snapshot for Primärvård 2022 not found")
	}
}

func TestFetchRegionBudgetMultiYear_GotlandSpecialCode(t *testing.T) {
	// Gotland: region code "09" → SCB code "0980L"
	var gotSCBCode string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		// PxWebApi v2 body: {"selection":[{"variableCode":..,"valueCodes":[..]}]}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if q, ok := body["selection"].([]any); ok {
				for _, item := range q {
					m, _ := item.(map[string]any)
					if m["variableCode"] == "Region" {
						vals, _ := m["valueCodes"].([]any)
						if len(vals) > 0 {
							gotSCBCode, _ = vals[0].(string)
						}
					}
				}
			}
		}
		// Return empty data — just testing the code conversion.
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	cl := newTestClient(srv)
	_, _ = cl.FetchRegionBudgetMultiYear(context.Background(), "09", []int{2023})
	if gotSCBCode != "0980L" {
		t.Errorf("expected SCB code '0980L' for Gotland, got %q", gotSCBCode)
	}
}

func TestFetchRegionBudgetMultiYear_SkipDotDot(t *testing.T) {
	// All values are ".." → should return empty slice, no error.
	rows := []struct {
		Key    []string
		Values []string
	}{
		{Key: []string{"06L", "0", "2022"}, Values: []string{".."}},
		{Key: []string{"06L", "0-9", "2022"}, Values: []string{".."}},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(scbResponse(rows))
	}))
	defer srv.Close()

	cl := newTestClient(srv)
	snapshots, err := cl.FetchRegionBudgetMultiYear(context.Background(), "06", []int{2022})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snapshots) != 0 {
		t.Errorf("expected 0 snapshots for all-dotdot data, got %d", len(snapshots))
	}
}

// --- FetchAreaAcrossRegions ---

func TestFetchAreaAcrossRegions_Basic(t *testing.T) {
	// Fixture: 3 regions, area "Primärvård" (code "0") + total
	rows := []struct {
		Key    []string
		Values []string
	}{
		{Key: []string{"06L", "0", "2023"}, Values: []string{"1000"}},
		{Key: []string{"06L", "0-9", "2023"}, Values: []string{"5000"}},
		{Key: []string{"07L", "0", "2023"}, Values: []string{"800"}},
		{Key: []string{"07L", "0-9", "2023"}, Values: []string{"4000"}},
		{Key: []string{"08L", "0", "2023"}, Values: []string{"600"}},
		{Key: []string{"08L", "0-9", "2023"}, Values: []string{"3000"}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(scbResponse(rows))
	}))
	defer srv.Close()

	cl := newTestClient(srv)
	datapoints, err := cl.FetchAreaAcrossRegions(context.Background(), "Primärvård", []string{"06", "07", "08"}, 2023)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(datapoints) != 3 {
		t.Fatalf("expected 3 datapoints, got %d", len(datapoints))
	}

	byCode := map[string]struct{ v, t, p float64 }{}
	for _, d := range datapoints {
		byCode[d.RegionCode] = struct{ v, t, p float64 }{d.ValueMnkr, d.TotalMnkr, d.Pct}
	}

	if d, ok := byCode["06"]; !ok {
		t.Error("missing datapoint for region 06")
	} else {
		if d.v != 1000 {
			t.Errorf("region 06: expected value 1000, got %f", d.v)
		}
		if d.t != 5000 {
			t.Errorf("region 06: expected total 5000, got %f", d.t)
		}
		if d.p != 20.0 {
			t.Errorf("region 06: expected pct 20.0, got %f", d.p)
		}
	}
}

func TestFetchAreaAcrossRegions_UnknownArea(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	cl := newTestClient(srv)
	_, err := cl.FetchAreaAcrossRegions(context.Background(), "NonExistentArea", []string{"06"}, 2023)
	if err == nil {
		t.Error("expected error for unknown area name, got nil")
	}
}

func TestFetchAreaAcrossRegions_SkipDotDot(t *testing.T) {
	rows := []struct {
		Key    []string
		Values []string
	}{
		{Key: []string{"06L", "0", "2023"}, Values: []string{".."}},
		{Key: []string{"06L", "0-9", "2023"}, Values: []string{"5000"}},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(scbResponse(rows))
	}))
	defer srv.Close()

	cl := newTestClient(srv)
	datapoints, err := cl.FetchAreaAcrossRegions(context.Background(), "Primärvård", []string{"06"}, 2023)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The "0-9" row is parsed but area row is skipped; region ends up with area=0, total=5000, pct=0.
	for _, d := range datapoints {
		if d.RegionCode == "06" && d.Pct != 0.0 {
			t.Errorf("expected pct=0 when area value is skipped, got %f", d.Pct)
		}
	}
}

// --- FetchPopulationTrend (recorded SCB fixtures, see internal/scbpx/testdata) ---

func TestFetchPopulationTrend_SkipsUnpublishedYears(t *testing.T) {
	meta, err := os.ReadFile("../../../scbpx/testdata/befolkning_meta_v2_tab638.json")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../../scbpx/testdata/befolkning_trend_v2_tab638.json")
	if err != nil {
		t.Fatal(err)
	}
	var gotTid []any
	dataCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tables/TAB638/metadata":
			w.Write(meta)
		case "/tables/TAB638/data":
			dataCalls++
			if dataCalls == 1 { // SCB v2's intermittent 500
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			for _, item := range body["selection"].([]any) {
				m := item.(map[string]any)
				if m["variableCode"] == "Tid" {
					gotTid = m["valueCodes"].([]any)
				}
			}
			w.Write(data)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	cl := newTestClient(srv)
	// 2025 is not in BE0101 yet; asking for it would make SCB reject the query.
	entries, err := cl.FetchPopulationTrend(context.Background(), "0180", []int{2025, 2024, 2023, 2022, 2021})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gotTid) != 4 || gotTid[0] != "2024" {
		t.Errorf("requested Tid = %v, want 2024..2021", gotTid)
	}
	want := map[int]int{2021: 978770, 2022: 984748, 2023: 988943, 2024: 995574}
	if len(entries) != len(want) {
		t.Fatalf("entries = %v", entries)
	}
	for _, e := range entries {
		if want[e.Year] != e.Population {
			t.Errorf("%d = %d, want %d", e.Year, e.Population, want[e.Year])
		}
	}
}
