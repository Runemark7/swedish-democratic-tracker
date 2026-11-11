package scbmacro

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Recorded fixtures (testdata/, captured from SCB on 2026-11-11):
//   kpi_v2_tab6596.json   PxWebApi v2 GET TAB6596 ContentsCode=00000804 Tid=top(2)
//   kpi_v1_kpi2020m.json  PxWebApi v1 POST KPI2020M, same selection
//   aku_v1_akurlbefk.json PxWebApi v1 POST AKURLBefK incl. ContentsCode=000007V3
//   aku_v2_tab6516.json   PxWebApi v2 GET TAB6516, same selection (after 500s)
// aku_v2_unpublished.synthetic.json is the TAB6516 shape with a third,
// unpublished (null / "..") quarter, which SCB only serves around releases.

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// fakeSCB emulates both PxWebApi versions. It answers 400 when a variable we
// rely on is not selected, and can fail the AKU table the way SCB has since
// November 2026: v2 500 / v1 400 for a number of calls, then success.
type fakeSCB struct {
	t           *testing.T
	akuV2Status int // status while failing; 0 = never fail
	akuV2Fails  int // number of failing calls before success; <0 = always
	akuV2Body   string
	akuV1Fails  int // v1 AKU calls answered 400 before success; <0 = always
	v1Down      bool
	akuV2Calls  int
	akuV1Calls  int
}

var v1Mandatory = map[string][]string{
	"/v1/AM/AM0401/AM0401A/AKURLBefK": {"Arbetskraftstillh", "TypData", "ContentsCode", "Tid"},
	"/v1/PR/PR0101/PR0101A/KPI2020M":  {"ContentsCode", "Tid"},
}

func (f *fakeSCB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/v2/tables/TAB6596/data":
		q := r.URL.Query()
		if q.Get("valueCodes[ContentsCode]") != "00000804" || q.Get("valueCodes[Tid]") != "top(2)" || q.Get("outputFormat") != "json-stat2" {
			http.Error(w, "bad selection "+r.URL.RawQuery, 400)
			return
		}
		w.Write(fixture(f.t, "kpi_v2_tab6596.json"))
	case r.Method == http.MethodGet && r.URL.Path == "/v2/tables/TAB6516/data":
		q := r.URL.Query()
		want := map[string]string{"Arbetskraftstillh": "ALÖSP", "TypData": "SR_DATA", "Kon": "1+2", "Alder": "tot15-74", "ContentsCode": "000007V3", "Tid": "top(2)"}
		for k, v := range want {
			if got := q.Get("valueCodes[" + k + "]"); got != v {
				http.Error(w, "valueCodes["+k+"]="+got, 400)
				return
			}
		}
		f.akuV2Calls++
		if f.akuV2Status != 0 && (f.akuV2Fails < 0 || f.akuV2Calls <= f.akuV2Fails) {
			w.WriteHeader(f.akuV2Status)
			return
		}
		body := f.akuV2Body
		if body == "" {
			body = "aku_v2_tab6516.json"
		}
		w.Write(fixture(f.t, body))
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1/"):
		if f.v1Down {
			http.NotFound(w, r)
			return
		}
		var body struct {
			Query []struct {
				Code string `json:"code"`
			} `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Bad Request", 400)
			return
		}
		have := map[string]bool{}
		for _, q := range body.Query {
			have[q.Code] = true
		}
		for _, m := range v1Mandatory[r.URL.Path] {
			if !have[m] {
				http.Error(w, "Bad Request", 400)
				return
			}
		}
		if strings.HasSuffix(r.URL.Path, "AKURLBefK") {
			f.akuV1Calls++
			if f.akuV1Fails < 0 || f.akuV1Calls <= f.akuV1Fails {
				http.Error(w, "Bad Request", 400)
				return
			}
			w.Write(fixture(f.t, "aku_v1_akurlbefk.json"))
		} else {
			w.Write(fixture(f.t, "kpi_v1_kpi2020m.json"))
		}
	default:
		http.NotFound(w, r)
	}
}

func newFake(t *testing.T, f *fakeSCB) *Client {
	f.t = t
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	c := NewClientWithBaseURLs(srv.URL+"/v2", srv.URL+"/v1")
	c.retryDelay = time.Millisecond
	return c
}

func TestInflationFromV2Recording(t *testing.T) {
	c := newFake(t, &fakeSCB{v1Down: true})
	pts, err := c.FetchInflationRate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 2 || pts[0].Period != "2026M09" || pts[0].Value != 1.1 || pts[1].Period != "2026M08" || pts[1].Value != 0.3 {
		t.Fatalf("got %+v", pts)
	}
}

func TestUnemploymentFromV2Recording(t *testing.T) {
	c := newFake(t, &fakeSCB{v1Down: true})
	pts, err := c.FetchUnemploymentRate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 2 || pts[0].Period != "2026K3" || pts[0].Value != 8.7 || pts[1].Period != "2026K2" || pts[1].Value != 8.7 {
		t.Fatalf("got %+v", pts)
	}
}

func TestUnemploymentV2SkipsUnpublished(t *testing.T) {
	c := newFake(t, &fakeSCB{v1Down: true, akuV2Body: "aku_v2_unpublished.synthetic.json"})
	pts, err := c.FetchUnemploymentRate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// 2026K4 is null (".."), so the latest published quarter is 2026K3.
	if len(pts) != 2 || pts[0].Period != "2026K3" || pts[0].Value != 8.7 || pts[1].Period != "2026K2" {
		t.Fatalf("got %+v", pts)
	}
}

// SCB on 2026-11-11: TAB6516 answers 500 a few times, then 200.
func TestUnemploymentRetriesFlakyV2(t *testing.T) {
	f := &fakeSCB{akuV2Status: 500, akuV2Fails: 2, v1Down: true}
	c := newFake(t, f)
	pts, err := c.FetchUnemploymentRate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if pts[0].Value != 8.7 || f.akuV2Calls != 3 {
		t.Fatalf("got %+v after %d calls", pts, f.akuV2Calls)
	}
}

// v2 keeps failing; v1 answers 400 twice (as it did in prod), then 200.
func TestUnemploymentFallsBackToFlakyV1(t *testing.T) {
	f := &fakeSCB{akuV2Status: 500, akuV2Fails: -1, akuV1Fails: 2}
	c := newFake(t, f)
	pts, err := c.FetchUnemploymentRate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 2 || pts[0].Period != "2026K3" || pts[0].Value != 8.7 || pts[1].Period != "2026K2" {
		t.Fatalf("got %+v", pts)
	}
	if f.akuV2Calls != attempts || f.akuV1Calls != 3 {
		t.Fatalf("calls v2=%d v1=%d", f.akuV2Calls, f.akuV1Calls)
	}
}

func TestInflationV1FallbackRecording(t *testing.T) {
	c := newFake(t, &fakeSCB{})
	c.v2Base += "/gone" // 404: final, no retries
	pts, err := c.FetchInflationRate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if pts[0].Period != "2026M09" || pts[0].Value != 1.1 {
		t.Fatalf("got %+v", pts)
	}
}

func TestBothVersionsFailingReportsBoth(t *testing.T) {
	c := newFake(t, &fakeSCB{akuV2Status: 500, akuV2Fails: -1, v1Down: true})
	_, err := c.FetchUnemploymentRate(context.Background())
	if err == nil || !strings.Contains(err.Error(), "v2: SCB returned 500") || !strings.Contains(err.Error(), "v1 fallback: SCB returned 404") {
		t.Fatalf("err = %v", err)
	}
}

func TestParseJSONStat2RejectsMultiCategoryDims(t *testing.T) {
	_, err := parseJSONStat2([]byte(`{"id":["Kon","Tid"],"size":[2,1],"value":[1,2],"dimension":{"Tid":{"category":{"index":{"2026K1":0}}}}}`))
	if err == nil {
		t.Fatal("want error for a dimension with 2 categories")
	}
}
