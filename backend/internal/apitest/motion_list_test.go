package apitest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// The list reads the same recordings as the decision chain: 2017/18:3293,
// 2017/18:4202, 2018/19:2900 and 2021/22:3638.

type listRow struct {
	Riksmote    string   `json:"riksmote"`
	Beteckning  string   `json:"beteckning"`
	Title       string   `json:"title"`
	Kind        string   `json:"kind"`
	Parties     []string `json:"parties"`
	Date        *string  `json:"date"`
	Utskott     string   `json:"utskott"`
	HandledIn   []string `json:"handledIn"`
	UtskottName string   `json:"utskottName"`
	Outcome     struct {
		Word   string         `json:"word"`
		Counts map[string]int `json:"counts"`
		Total  int            `json:"total"`
	} `json:"outcome"`
}

type listBody struct {
	Data       []listRow `json:"data"`
	NextCursor *string   `json:"nextCursor"`
}

func getList(t *testing.T, h *Harness, query string) listBody {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/motions?"+query, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/motions?%s = %d: %s", query, rec.Code, rec.Body.String())
	}
	var b listBody
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("decode: %v: %s", err, rec.Body.String())
	}
	return b
}

func ids(b listBody) []string {
	out := []string{}
	for _, r := range b.Data {
		out = append(out, r.Riksmote+":"+r.Beteckning)
	}
	return out
}

func sameSet(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	m := map[string]bool{}
	for _, g := range got {
		m[g] = true
	}
	for _, w := range want {
		if !m[w] {
			return false
		}
	}
	return true
}

func listHarness(t *testing.T) *Harness {
	h := startChain(t, chainDir(t, "decision-chain"), "{2017/18,2018/19,2021/22}")
	runChain(h)
	return h
}

func TestListMsAvslagnaKommittemotionerInSkatteutskottetReturnsTheRepeats(t *testing.T) {
	h := listHarness(t)
	b := getList(t, h, "party=M&utskott=SkU&outcome=Avslagen&kind=Kommittémotion")
	want := []string{"2017/18:3293", "2018/19:2900", "2021/22:3638"}
	if !sameSet(ids(b), want) {
		t.Fatalf("got %v, want %v", ids(b), want)
	}
	for _, r := range b.Data {
		if r.Kind != "Kommittémotion" || r.Outcome.Word != "Avslagen" {
			t.Errorf("row = %+v", r)
		}
		if r.Beteckning == "2900" && (r.Utskott != "FiU" || len(r.HandledIn) != 1 || r.HandledIn[0] != "2018/19:SkU11") {
			t.Errorf("2900 is filed to FiU and decided in SkU11: %+v", r)
		}
		if len(r.Parties) == 0 || r.Parties[0] != "M" || r.Date == nil || r.Title == "" {
			t.Errorf("row = %+v", r)
		}
	}
}

func TestListSeveralRiksmotenAtOnceWithOutcomes(t *testing.T) {
	h := listHarness(t)
	b := getList(t, h, "party=M&utskott=SkU&riksmote="+url.QueryEscape("2017/18,2021/22"))
	if !sameSet(ids(b), []string{"2017/18:3293", "2021/22:3638"}) {
		t.Fatalf("got %v", ids(b))
	}
	for _, r := range b.Data {
		if r.Outcome.Word == "" || r.Outcome.Total == 0 {
			t.Errorf("no outcome on %+v", r)
		}
	}
}

func TestListKindFlerpartimotionReturns4202(t *testing.T) {
	h := listHarness(t)
	b := getList(t, h, "kind=Flerpartimotion")
	if !sameSet(ids(b), []string{"2017/18:4202"}) {
		t.Fatalf("got %v", ids(b))
	}
	if b.Data[0].Outcome.Word != "Bifallen" || len(b.Data[0].Parties) != 3 {
		t.Errorf("row = %+v", b.Data[0])
	}
	if e := getList(t, h, "kind=Enskild+motion"); len(e.Data) != 0 {
		t.Errorf("4202 is not an Enskild motion: %v", ids(e))
	}
}

func TestListSearchFindsVirtuellaVarorInTitleOrYrkande(t *testing.T) {
	h := listHarness(t)
	b := getList(t, h, "q="+url.QueryEscape("virtuella varor"))
	got := ids(b)
	found := false
	for _, g := range got {
		if g == "2017/18:3293" {
			found = true
		}
		if g == "2017/18:4202" {
			t.Errorf("4202 is not about virtuella varor: %v", got)
		}
	}
	if !found {
		t.Fatalf("2017/18:3293 not found: %v", got)
	}
	if n := getList(t, h, "q=zzzxqj"); len(n.Data) != 0 {
		t.Errorf("nonsense matched %v", ids(n))
	}
	if n := getList(t, h, "q="+url.QueryEscape("100%")); len(n.Data) != 0 {
		t.Errorf("percent must be literal, matched %v", ids(n))
	}
}

func TestListIsSortedMostRecentlyDecidedAndPagedWithCursors(t *testing.T) {
	h := listHarness(t)
	all := getList(t, h, "limit=100")
	if len(all.Data) < 4 || all.NextCursor != nil {
		t.Fatalf("all = %v cursor %v", ids(all), all.NextCursor)
	}
	var seen []string
	cursor := ""
	for i := 0; i < 10; i++ {
		q := "limit=1"
		if cursor != "" {
			q += "&cursor=" + url.QueryEscape(cursor)
		}
		p := getList(t, h, q)
		seen = append(seen, ids(p)...)
		if p.NextCursor == nil {
			break
		}
		cursor = *p.NextCursor
	}
	want := ids(all)
	if len(seen) != len(want) {
		t.Fatalf("paged %v, all %v", seen, want)
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("paged %v, all %v", seen, want)
		}
	}
}

func TestListPagesThroughAnOutcomeFilterWithoutSkippingRows(t *testing.T) {
	h := listHarness(t)
	var seen []string
	cursor := ""
	for i := 0; i < 10; i++ {
		q := "outcome=Avslagen&limit=1"
		if cursor != "" {
			q += "&cursor=" + url.QueryEscape(cursor)
		}
		p := getList(t, h, q)
		seen = append(seen, ids(p)...)
		if p.NextCursor == nil {
			break
		}
		cursor = *p.NextCursor
	}
	if !sameSet(seen, ids(getList(t, h, "outcome=Avslagen"))) || len(seen) < 3 {
		t.Fatalf("paged %v", seen)
	}
}

func TestListRejectsBadFilters(t *testing.T) {
	h := listHarness(t)
	for _, q := range []string{"kind=Bogus", "riksmote=2017-18", "limit=0", "limit=abc", "cursor=!!!", "outcome=Bogus"} {
		if rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/motions?"+q, nil)); rec.Code != http.StatusBadRequest {
			t.Errorf("GET ?%s = %d, want 400", q, rec.Code)
		}
	}
}
