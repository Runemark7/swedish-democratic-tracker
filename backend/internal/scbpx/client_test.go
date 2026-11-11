package scbpx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// Fixtures in testdata/ were recorded from the live SCB APIs on 2026-11-11:
//   befolkning_v2_tab638.json        POST v2 /tables/TAB638/data (json-px), Region 0114,0115,0117,01,09, Alder tot, BE0101N1, 2023
//   befolkning_v1_befolkningny.json  POST v1 BE/BE0101/BE0101A/BefolkningNy (json), same selection
//   befolkning_trend_v2_tab638.json  POST v2 TAB638, Region 0180, Alder tot, BE0101N1, 2021-2024
//   befolkning_meta_v2_tab638.json   GET  v2 /tables/TAB638/metadata (json-stat2)
//   befolkning_meta_v1_befolkningny.json GET v1 BE/BE0101/BE0101A/BefolkningNy

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// fakeSCB serves v2 under /v2 and v1 under /v1. Each route answers with the
// queued failure statuses first, then the fixture.
type fakeSCB struct {
	mu     sync.Mutex
	fails  map[string][]int
	bodies map[string][]byte
	calls  map[string]int
	seen   map[string][]byte // last request body per route
}

func newFake() *fakeSCB {
	return &fakeSCB{fails: map[string][]int{}, bodies: map[string][]byte{}, calls: map[string]int{}, seen: map[string][]byte{}}
}

func (f *fakeSCB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	route := r.Method + " " + r.URL.Path
	f.calls[route]++
	if r.Body != nil {
		f.seen[route], _ = io.ReadAll(r.Body)
	}
	if q := f.fails[route]; len(q) > 0 {
		f.fails[route] = q[1:]
		w.WriteHeader(q[0])
		return
	}
	b, ok := f.bodies[route]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(b)
}

const (
	v2Data = "POST /v2/tables/TAB638/data"
	v2Meta = "GET /v2/tables/TAB638/metadata"
	v1Path = "/v1/BE/BE0101/BE0101A/BefolkningNy"
	v1Data = "POST " + v1Path
	v1Meta = "GET " + v1Path
)

func newTestClient(t *testing.T, f *fakeSCB) *Client {
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	c := NewWithBaseURLs(srv.URL+"/v2", srv.URL+"/v1", srv.Client())
	c.SetRetryDelay(time.Millisecond)
	return c
}

var popSel = []Select{
	{Code: "Region", Values: []string{"0114", "0115", "0117", "01", "09"}},
	{Code: "Alder", Values: []string{"tot"}},
	{Code: "ContentsCode", Values: []string{"BE0101N1"}},
	{Code: "Tid", Values: []string{"2023"}},
}

func popByRegion(rows []Row) map[string]string {
	out := map[string]string{}
	for _, r := range rows {
		out[r.Key[0]] = r.Values[0]
	}
	return out
}

func checkPops(t *testing.T, rows []Row) {
	t.Helper()
	got := popByRegion(rows)
	want := map[string]string{"0114": "50110", "0115": "35076", "0117": "49282", "01": "2454821", "09": "61029"}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("population %s = %q, want %q", k, got[k], v)
		}
	}
	if len(rows) > 0 && len(rows[0].Key) != 3 {
		t.Errorf("key %v, want [Region Alder Tid]", rows[0].Key)
	}
}

func TestData_V2RetriesFlaky500(t *testing.T) {
	f := newFake()
	f.bodies[v2Data] = fixture(t, "befolkning_v2_tab638.json")
	f.fails[v2Data] = []int{500, 500, 500}
	c := newTestClient(t, f)
	rows, err := c.Data(context.Background(), BefolkningNy, popSel)
	if err != nil {
		t.Fatal(err)
	}
	checkPops(t, rows)
	if f.calls[v2Data] != 4 || f.calls[v1Data] != 0 {
		t.Errorf("calls v2=%d v1=%d, want 4/0", f.calls[v2Data], f.calls[v1Data])
	}
	var body struct {
		Selection []struct {
			VariableCode string   `json:"variableCode"`
			ValueCodes   []string `json:"valueCodes"`
		} `json:"selection"`
	}
	if err := json.Unmarshal(f.seen[v2Data], &body); err != nil || len(body.Selection) != 4 || body.Selection[1].VariableCode != "Alder" {
		t.Errorf("v2 body = %s (%v)", f.seen[v2Data], err)
	}
}

// The election seeder error on 2026-11-11 ("SCB population returned 400 for
// batch starting 0114"): v1 answered 400 to a valid batch that succeeded on
// retry. With v2 down as well, the client must retry v1's 400s.
func TestData_FallsBackToV1AndRetriesFlaky400(t *testing.T) {
	f := newFake()
	f.fails[v2Data] = []int{500, 500, 500, 500, 500}
	f.bodies[v1Data] = fixture(t, "befolkning_v1_befolkningny.json")
	f.fails[v1Data] = []int{400, 400}
	c := newTestClient(t, f)
	rows, err := c.Data(context.Background(), BefolkningNy, popSel)
	if err != nil {
		t.Fatal(err)
	}
	checkPops(t, rows)
	if f.calls[v2Data] != attempts || f.calls[v1Data] != 3 {
		t.Errorf("calls v2=%d v1=%d, want %d/3", f.calls[v2Data], f.calls[v1Data], attempts)
	}
	if !strings.Contains(string(f.seen[v1Data]), `"filter":"item"`) {
		t.Errorf("v1 body = %s", f.seen[v1Data])
	}
}

// A v2 400 is SCB rejecting the query (e.g. an unpublished year): no retry and
// no v1 fallback.
func TestData_V2BadRequestIsFinal(t *testing.T) {
	f := newFake()
	f.fails[v2Data] = []int{400}
	f.bodies[v1Data] = fixture(t, "befolkning_v1_befolkningny.json")
	c := newTestClient(t, f)
	_, err := c.Data(context.Background(), BefolkningNy, popSel)
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Fatalf("err = %v, want a 400 error", err)
	}
	if f.calls[v2Data] != 1 || f.calls[v1Data] != 0 {
		t.Errorf("calls v2=%d v1=%d, want 1/0", f.calls[v2Data], f.calls[v1Data])
	}
}

func TestData_BothDown(t *testing.T) {
	f := newFake()
	f.fails[v2Data] = []int{500, 500, 500, 500, 500}
	f.fails[v1Data] = []int{400, 400, 400, 400, 400}
	c := newTestClient(t, f)
	_, err := c.Data(context.Background(), BefolkningNy, popSel)
	if err == nil || !strings.Contains(err.Error(), "v1 fallback") {
		t.Fatalf("err = %v", err)
	}
}

func TestMeta_V2ParsesAndCaches(t *testing.T) {
	f := newFake()
	f.bodies[v2Meta] = fixture(t, "befolkning_meta_v2_tab638.json")
	f.fails[v2Meta] = []int{503}
	c := newTestClient(t, f)
	for i := 0; i < 2; i++ {
		m, err := c.Meta(context.Background(), BefolkningNy)
		if err != nil {
			t.Fatal(err)
		}
		checkMeta(t, m)
	}
	if f.calls[v2Meta] != 2 {
		t.Errorf("v2 meta calls = %d, want 2 (one retry, then cached)", f.calls[v2Meta])
	}
}

func TestMeta_V1Fallback(t *testing.T) {
	f := newFake()
	f.fails[v2Meta] = []int{500, 500, 500, 500, 500}
	f.bodies[v1Meta] = fixture(t, "befolkning_meta_v1_befolkningny.json")
	c := newTestClient(t, f)
	m, err := c.Meta(context.Background(), BefolkningNy)
	if err != nil {
		t.Fatal(err)
	}
	checkMeta(t, m)
}

func checkMeta(t *testing.T, m *Meta) {
	t.Helper()
	var codes []string
	for _, v := range m.Variables {
		codes = append(codes, v.Code)
	}
	if strings.Join(codes, ",") != "Region,Civilstand,Alder,Kon,ContentsCode,Tid" {
		t.Errorf("variables = %v", codes)
	}
	r := m.Variable("Region")
	if r == nil || len(r.Values) != 312 || r.Values[2] != "0114" || r.ValueTexts[2] != "Upplands Väsby" {
		t.Errorf("Region = %d values, [2]=%v", len(r.Values), r.Values[2]+"/"+r.ValueTexts[2])
	}
}

func TestHas_FiltersUnpublishedYears(t *testing.T) {
	f := newFake()
	f.bodies[v2Meta] = fixture(t, "befolkning_meta_v2_tab638.json")
	c := newTestClient(t, f)
	got, err := c.Has(context.Background(), BefolkningNy, "Tid", []string{"2025", "2024", "2023", "2022", "2021"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, ",") != "2024,2023,2022,2021" {
		t.Errorf("Has = %v", got)
	}
}

// After v2 exhausts its retries, the next call goes to v1 first.
func TestData_V2CooldownPrefersV1(t *testing.T) {
	f := newFake()
	f.fails[v2Data] = []int{500, 500, 500, 500, 500}
	f.bodies[v2Data] = fixture(t, "befolkning_v2_tab638.json")
	f.bodies[v1Data] = fixture(t, "befolkning_v1_befolkningny.json")
	c := newTestClient(t, f)
	for i := 0; i < 2; i++ {
		rows, err := c.Data(context.Background(), BefolkningNy, popSel)
		if err != nil {
			t.Fatal(err)
		}
		checkPops(t, rows)
	}
	if f.calls[v2Data] != attempts || f.calls[v1Data] != 2 {
		t.Errorf("calls v2=%d v1=%d, want %d/2", f.calls[v2Data], f.calls[v1Data], attempts)
	}
	// v1 failing during the cooldown still reaches v2.
	f.fails[v1Data] = []int{400, 400, 400, 400, 400}
	rows, err := c.Data(context.Background(), BefolkningNy, popSel)
	if err != nil {
		t.Fatal(err)
	}
	checkPops(t, rows)
	if c.preferV1(BefolkningNy) {
		t.Error("v2 success should end the cooldown")
	}
}
