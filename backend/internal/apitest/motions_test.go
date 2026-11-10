package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Recorded from data.riksdagen.se on 2026-11-10. Motion 2017/18:3293 is a
// Kommittémotion by Maria Malmer Stenergard m.fl. (M); 2017/18:4202 is typed
// Enskild motion by Riksdagen and signed by C, KD and M.
func motionFixtureDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "fixtures", "motions-2017-18")
}

type motionBody struct {
	DokID           string  `json:"dokId"`
	Riksmote        string  `json:"riksmote"`
	Beteckning      string  `json:"beteckning"`
	Title           string  `json:"title"`
	Subtyp          string  `json:"subtyp"`
	Kind            string  `json:"kind"`
	KindExplanation string  `json:"kindExplanation"`
	Date            *string `json:"date"`
	Utskott         string  `json:"utskott"`
	UtskottName     string  `json:"utskottName"`
	BehandlasI      string  `json:"behandlasI"`
	SourceURL       string  `json:"sourceUrl"`
	Authors         []struct {
		Name  string `json:"name"`
		Party string `json:"party"`
	} `json:"authors"`
	Yrkanden []struct {
		Number          int    `json:"number"`
		Wording         string `json:"wording"`
		UtskottProposal string `json:"utskottProposal"`
		ChamberDecision string `json:"chamberDecision"`
		BehandlasI      string `json:"behandlasI"`
	} `json:"yrkanden"`
}

// startMotions declares a single mandate period holding the given riksmöten,
// so the worker's riksmöten come from the table and the fixture covers them.
func startMotions(t *testing.T, riksmoten string) *Harness {
	t.Helper()
	h := Start(t, motionFixtureDir(t), WithNow(time.Date(2018, 3, 29, 12, 0, 0, 0, time.UTC)))
	ctx := context.Background()
	if _, err := h.DB.Exec(ctx, `DELETE FROM mandate_periods`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
		VALUES ('test', 'test', '2014-09-14', '2018-09-09', $1::text[])`, riksmoten); err != nil {
		t.Fatal(err)
	}
	return h
}

func getMotion(t *testing.T, h *Harness, path string) motionBody {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d: %s", path, rec.Code, rec.Body.String())
	}
	var m motionBody
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode: %v: %s", err, rec.Body.String())
	}
	return m
}

func TestMotion3293IsAKommittemotionWithItsYrkande(t *testing.T) {
	h := startMotions(t, "{2017/18}")
	h.RunWorker(context.Background(), "fetch-motions")
	if h.Served("/dokumentstatus/H5023293.json") == 0 {
		t.Fatal("ingestion did not read the recorded motion")
	}

	m := getMotion(t, h, "/api/motions/2017%2F18/3293")
	if m.Riksmote != "2017/18" || m.Beteckning != "3293" || m.DokID != "H5023293" {
		t.Fatalf("identity = %+v", m)
	}
	if m.Title != "Skattemässig hantering av virtuella varor" {
		t.Errorf("title = %q", m.Title)
	}
	if m.Subtyp != "Kommittémotion" || m.Kind != "Kommittémotion" {
		t.Errorf("subtyp/kind = %q/%q", m.Subtyp, m.Kind)
	}
	if m.KindExplanation == "" {
		t.Error("kind has no one-line explanation")
	}
	if m.Date == nil || *m.Date != "2017-10-05" {
		t.Errorf("date = %v", m.Date)
	}
	if m.Utskott != "SkU" || m.UtskottName != "Skatteutskottet" {
		t.Errorf("utskott = %q / %q", m.Utskott, m.UtskottName)
	}
	if m.BehandlasI != "2017/18:SkU11" {
		t.Errorf("behandlas_i = %q", m.BehandlasI)
	}
	if m.SourceURL != "https://data.riksdagen.se/dokument/H5023293.html" {
		t.Errorf("sourceUrl = %q", m.SourceURL)
	}
	if len(m.Authors) != 5 || m.Authors[0].Name != "Maria Malmer Stenergard" || m.Authors[0].Party != "M" {
		t.Errorf("authors = %+v", m.Authors)
	}
	if len(m.Yrkanden) != 1 {
		t.Fatalf("yrkanden = %+v", m.Yrkanden)
	}
	y := m.Yrkanden[0]
	const wording = "Riksdagen ställer sig bakom det som anförs i motionen om att tillsätta en utredning för att se över den skatterättsliga bedömningen av handel med virtuella varor och tillkännager detta för regeringen."
	if y.Number != 1 || y.Wording != wording {
		t.Errorf("yrkande 1 = %+v", y)
	}
	if y.UtskottProposal != "Avslag" || y.ChamberDecision != "Avslag" || y.BehandlasI != "2017/18:SkU11" {
		t.Errorf("decision as given = %+v", y)
	}
}

func TestMotion4202IsTypedEnskildByRiksdagenAndDerivedFlerpartimotion(t *testing.T) {
	h := startMotions(t, "{2017/18}")
	h.RunWorker(context.Background(), "fetch-motions")

	m := getMotion(t, h, "/api/motions/2017%2F18/4202")
	if m.Subtyp != "Enskild motion" {
		t.Errorf("subtyp as given = %q", m.Subtyp)
	}
	if m.Kind != "Flerpartimotion" {
		t.Errorf("kind = %q", m.Kind)
	}
	parties := map[string]bool{}
	for _, a := range m.Authors {
		parties[a.Party] = true
	}
	if !parties["C"] || !parties["KD"] || !parties["M"] || len(parties) != 3 {
		t.Errorf("signer parties = %v", parties)
	}
	if m.BehandlasI != "2018/19:CU2" {
		t.Errorf("behandlas_i = %q", m.BehandlasI)
	}
}

func TestMotionIngestionIsIdempotentAndRecordsCoverage(t *testing.T) {
	h := startMotions(t, "{2017/18}")
	ctx := context.Background()
	h.RunWorker(ctx, "fetch-motions")
	first := getMotion(t, h, "/api/motions/2017%2F18/3293")
	reads := h.Served("/dokumentstatus/H5023293.json") + h.Served("/dokumentstatus/H5024202.json")
	if reads != 2 {
		t.Fatalf("first run read %d motions, want 2", reads)
	}

	h.RunWorker(ctx, "fetch-motions")
	if got := h.Served("/dokumentstatus/H5023293.json") + h.Served("/dokumentstatus/H5024202.json"); got != reads {
		t.Errorf("re-run read motions again: %d reads, was %d", got, reads)
	}
	again := getMotion(t, h, "/api/motions/2017%2F18/3293")
	a, _ := json.Marshal(first)
	b, _ := json.Marshal(again)
	if string(a) != string(b) {
		t.Errorf("re-run changed the motion:\n%s\n%s", a, b)
	}

	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/motions/coverage", nil))
	var cov struct {
		Data []struct {
			Riksmote string `json:"riksmote"`
			Source   string `json:"source"`
			Expected int    `json:"expected"`
			Held     int    `json:"held"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &cov); err != nil {
		t.Fatal(err)
	}
	if len(cov.Data) != 1 {
		t.Fatalf("coverage = %s", rec.Body.String())
	}
	c := cov.Data[0]
	if c.Riksmote != "2017/18" || c.Source != "riksdagen-dokumentlista" || c.Expected != 4203 || c.Held != 2 {
		t.Errorf("coverage row = %+v", c)
	}
}

func TestMotionRiksmotenComeFromMandatePeriodsAndSkipUnopenedOnes(t *testing.T) {
	// 2018/19 has not opened at the pinned clock (29 March 2018). Asking for it
	// would be an unrecorded request and fail the test.
	h := startMotions(t, "{2017/18,2018/19}")
	h.RunWorker(context.Background(), "fetch-motions")
	getMotion(t, h, "/api/motions/2017%2F18/3293")
}

func TestMotionAPIRejectsTheDashFormAndUnknownMotions(t *testing.T) {
	h := startMotions(t, "{2017/18}")
	h.RunWorker(context.Background(), "fetch-motions")
	for path, want := range map[string]int{
		"/api/motions/2017-18/3293":     http.StatusBadRequest,
		"/api/motions/2017%2F18/999999": http.StatusNotFound,
	} {
		if rec := h.Do(httptest.NewRequest(http.MethodGet, path, nil)); rec.Code != want {
			t.Errorf("GET %s = %d, want %d", path, rec.Code, want)
		}
	}
}
