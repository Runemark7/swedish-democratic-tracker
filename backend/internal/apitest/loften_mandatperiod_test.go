package apitest

import (
	"context"
	"sort"
	"strings"
	"testing"
)

// Seam tests for gap promise-list-no-term-filter: Mandatperiod is its own
// filter on the Löften list and summary, free to combine with party, utskott
// and Löftesstatus. Without it every period is returned, as before.

type periodLoefte struct {
	ID            int64    `json:"id"`
	Party         string   `json:"party"`
	Text          string   `json:"text"`
	MandatePeriod string   `json:"mandatePeriod"`
	Utskott       []string `json:"utskott"`
	Status        *string  `json:"status"`
}

type periodSeed struct {
	*kSeed
	oldSkU, oldAU, oldM int64
}

// startPeriods: S has two approved 2026–2030 Löften (SkU, AU) from startK, and
// two archived 2022–2026 ones (SkU, AU); M has one archived 2022–2026 Löfte.
// The archive the migrations seed is struck so the counts below are exact.
func startPeriods(t *testing.T) *periodSeed {
	t.Helper()
	s := &periodSeed{kSeed: startK(t)}
	s.oldSkU = seedLoefte(t, s.h, "S", "Gammal skatt", "archived", "SkU")
	s.oldAU = seedLoefte(t, s.h, "S", "Gammal arbetsrätt", "archived", "AU")
	s.oldM = seedLoefte(t, s.h, "M", "Gammalt M-löfte", "archived", "SkU")
	if _, err := s.h.DB.Exec(context.Background(),
		`UPDATE loften SET visibility = 'struck' WHERE visibility = 'archived' AND id <> ALL($1)`,
		[]int64{s.oldSkU, s.oldAU, s.oldM}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.h.DB.Exec(context.Background(),
		`UPDATE loften SET mandate_period = '2022-2026' WHERE id = ANY($1)`,
		[]int64{s.oldSkU, s.oldAU, s.oldM}); err != nil {
		t.Fatal(err)
	}
	return s
}

func listTexts(t *testing.T, h *Harness, path string) string {
	t.Helper()
	ls := decode[[]periodLoefte](t, publicGet(t, h, path, 200))
	texts := make([]string, 0, len(ls))
	for _, l := range ls {
		texts = append(texts, l.Text)
	}
	sort.Strings(texts)
	return strings.Join(texts, "|")
}

func TestLoftenListFiltersOnMandatperiod(t *testing.T) {
	s := startPeriods(t)
	for _, c := range []struct{ name, path, want string }{
		{"no period keeps every period", "/api/loften?party=S",
			"Arbetsrätt|Gammal arbetsrätt|Gammal skatt|Skatt sänks"},
		{"current period only", "/api/loften?mandatperiod=2026-2030", "Arbetsrätt|Skatt sänks"},
		{"previous period only", "/api/loften?mandatperiod=2022-2026",
			"Gammal arbetsrätt|Gammal skatt|Gammalt M-löfte"},
		{"period and party", "/api/loften?mandatperiod=2022-2026&party=S", "Gammal arbetsrätt|Gammal skatt"},
		{"period and party in the path", "/api/parties/S/loften?mandatperiod=2022-2026", "Gammal arbetsrätt|Gammal skatt"},
		{"period and utskott", "/api/loften?mandatperiod=2022-2026&utskott=SkU", "Gammal skatt|Gammalt M-löfte"},
		{"period, party and utskott", "/api/loften?mandatperiod=2026-2030&party=S&utskott=AU", "Arbetsrätt"},
		{"period and status", "/api/loften?mandatperiod=2026-2030&status=inte-provat", "Arbetsrätt|Skatt sänks"},
		{"period and a status it has none of", "/api/loften?mandatperiod=2022-2026&status=inte-provat", ""},
		{"period and arkiv", "/api/loften?mandatperiod=2026-2030&status=arkiv", ""},
		{"a period without Löften is empty, not an error", "/api/loften?mandatperiod=2018-2022", ""},
	} {
		if got := listTexts(t, s.h, c.path); got != c.want {
			t.Errorf("%s: GET %s = %q, want %q", c.name, c.path, got, c.want)
		}
	}
	// Every row says which period it belongs to, so the reader can check the filter.
	for _, l := range decode[[]periodLoefte](t, publicGet(t, s.h, "/api/loften?mandatperiod=2022-2026", 200)) {
		if l.MandatePeriod != "2022-2026" {
			t.Errorf("Löfte %d has mandatePeriod %q under the 2022-2026 filter", l.ID, l.MandatePeriod)
		}
	}
}

func TestLoftenMandatperiodMustBeWellFormed(t *testing.T) {
	s := startPeriods(t)
	for _, bad := range []string{"2026", "2026–2030", "nu", "2026-2030;DROP"} {
		publicGet(t, s.h, "/api/loften?mandatperiod="+strings.ReplaceAll(bad, ";", "%3B"), 400)
	}
	publicGet(t, s.h, "/api/loften/summary?mandatperiod=nu", 400)
}

func TestLoftenSummaryFollowsMandatperiod(t *testing.T) {
	s := startPeriods(t)
	type row struct {
		Party                    string
		Total, InteProvat, Arkiv int
	}
	get := func(path string) map[string]row {
		out := map[string]row{}
		for _, r := range decode[[]row](t, publicGet(t, s.h, path, 200)) {
			out[r.Party] = r
		}
		return out
	}
	all := get("/api/loften/summary")
	if all["S"] != (row{"S", 4, 2, 2}) || all["M"] != (row{"M", 1, 0, 1}) {
		t.Errorf("all periods: %+v", all)
	}
	now := get("/api/loften/summary?mandatperiod=2026-2030")
	if now["S"] != (row{"S", 2, 2, 0}) {
		t.Errorf("2026-2030: %+v", now)
	}
	if _, ok := now["M"]; ok {
		t.Errorf("M has no 2026-2030 Löften but is in the summary: %+v", now)
	}
	old := get("/api/loften/summary?mandatperiod=2022-2026")
	if old["S"] != (row{"S", 2, 0, 2}) || old["M"] != (row{"M", 1, 0, 1}) {
		t.Errorf("2022-2026: %+v", old)
	}
}
