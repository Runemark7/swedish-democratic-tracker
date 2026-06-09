package kolada

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestClient points the client at a stub server.
func newTestClient(url string) *Client {
	return &Client{http: &http.Client{Timeout: 5 * time.Second}, baseURL: url}
}

// koladaAllResponse is the shape Kolada's /data/kpi/{kpi}/year endpoint returns:
// one real value and one null value (a region that has not reported yet).
const koladaAllResponse = `{
  "values": [
    {"kpi":"N79173","municipality":"0014","period":2025,"values":[
      {"gender":"T","value":93.05,"status":"","isdeleted":false}]},
    {"kpi":"N79173","municipality":"0013","period":2025,"values":[
      {"gender":"T","value":null,"status":"","isdeleted":false}]}
  ]
}`

func TestFetchKPIAllMunicipalities_SkipsNullValues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(koladaAllResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	got, err := c.FetchKPIAllMunicipalities(context.Background(), "N79173", []int{2025})
	if err != nil {
		t.Fatalf("FetchKPIAllMunicipalities: %v", err)
	}

	// The null-valued region (0013) must NOT appear as a 0.0 row — a missing
	// value is "no data", not zero.
	for _, v := range got {
		if v.MunCode == "0013" {
			t.Fatalf("region 0013 has a null Kolada value but was returned as %v (should be skipped)", v.Value)
		}
	}
	if len(got) != 1 || got[0].MunCode != "0014" || got[0].Value != 93.05 {
		t.Fatalf("got %+v, want only the real 0014=93.05 row", got)
	}
}

// fetchOneKPI (the strip/card path) must skip null values too.
const koladaOneResponse = `{
  "values": [
    {"kpi":"N79173","period":2025,"values":[
      {"gender":"T","value":null,"status":"","isdeleted":false}]},
    {"kpi":"N79173","period":2024,"values":[
      {"gender":"T","value":87.65,"status":"","isdeleted":false}]}
  ]
}`

func TestFetchKPIs_SkipsNullValues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(koladaOneResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	got, err := c.FetchKPIs(context.Background(), "0013", []string{"N79173"}, []int{2025, 2024})
	if err != nil {
		t.Fatalf("FetchKPIs: %v", err)
	}

	if len(got) != 1 || got[0].Year != 2024 || got[0].Value != 87.65 {
		t.Fatalf("got %+v, want only the real 2024=87.65 row (the 2025 null skipped)", got)
	}
}
