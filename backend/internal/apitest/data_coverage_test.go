package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

type dataCoverageRow struct {
	Source     string  `json:"source"`
	Riksmote   string  `json:"riksmote"`
	Status     string  `json:"status"`
	Expected   int     `json:"expected"`
	Held       int     `json:"held"`
	LastUpdate *string `json:"lastUpdate"`
}

func TestDataCoverageReportsStatusAndLastUpdatePerSourceAndRiksmote(t *testing.T) {
	h := Start(t, filepath.Join(chainDir(t, "questions")))
	ctx := context.Background()
	for _, q := range []string{
		`INSERT INTO motion_backfill (riksmote, source, expected, held, completed_at)
		   VALUES ('2021/22','motioner',5,5,now()), ('2021/22','betankanden',7,7,now()),
		          ('2020/21','motioner',9,3,NULL)`,
		`INSERT INTO motion_coverage (riksmote, source, expected) VALUES ('2022/23','riksdagen-dokumentlista',12)`,
		`INSERT INTO speech_backfill (riksmote, expected, completed_at) VALUES ('2021/22',40,now()), ('2020/21',30,NULL)`,
		`INSERT INTO question_coverage (riksmote, kind, expected) VALUES ('2022/23','fraga',4), ('2022/23','interpellation',2)`,
		`INSERT INTO ingestion_coverage (riksmote, expected_voteringar) VALUES ('2022/23',100)`,
	} {
		if _, err := h.DB.Exec(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/data/coverage", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/data/coverage = %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data []dataCoverageRow `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	got := map[string]dataCoverageRow{}
	for _, r := range body.Data {
		got[r.Source+"|"+r.Riksmote] = r
	}
	want := map[string]string{
		"motioner|2021/22":         "complete",
		"betankanden|2021/22":      "complete",
		"motioner|2020/21":         "partial",
		"motioner|2022/23":         "partial",
		"anforanden|2021/22":       "complete",
		"anforanden|2020/21":       "partial",
		"fragor|2022/23":           "partial",
		"interpellationer|2022/23": "partial",
		"voteringar|2022/23":       "partial",
	}
	for k, status := range want {
		r, ok := got[k]
		if !ok {
			t.Errorf("missing %s", k)
			continue
		}
		if r.Status != status {
			t.Errorf("%s status = %q, want %q", k, r.Status, status)
		}
		if r.LastUpdate == nil || *r.LastUpdate == "" {
			t.Errorf("%s has no lastUpdate", k)
		}
	}
	if len(body.Data) != len(want) {
		t.Errorf("got %d rows, want %d: %s", len(body.Data), len(want), rec.Body.String())
	}
	if r := got["motioner|2020/21"]; r.Expected != 9 || r.Held != 0 {
		t.Errorf("motioner 2020/21 = %+v, want expected 9 and held 0 (counted live from stored motions)", r)
	}
	// Newest riksmöte first.
	if len(body.Data) > 0 && body.Data[0].Riksmote != "2022/23" {
		t.Errorf("first row riksmote = %s, want newest first", body.Data[0].Riksmote)
	}
}
