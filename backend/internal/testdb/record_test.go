package testdb

import (
	"strconv"
	"testing"

	"riksdagskollen/internal/committees/domain"
)

// The committee and vote integration tests pin this shape. The plan is what
// an empty migrated database is filled with, so a drift here fails before
// Postgres is involved.
func TestVotePlan_MatchesThePinnedRecord(t *testing.T) {
	rows := votePlan()
	if len(rows) != 2576 {
		t.Fatalf("rows = %d, want 2576", len(rows))
	}

	ids := map[string]bool{}
	codes := map[string]int{}
	auBallots := 0
	for _, row := range rows {
		if ids[row.id] {
			t.Errorf("duplicate votering id %s", row.id)
		}
		ids[row.id] = true
		code := domain.CommitteeCode(row.beteckning)
		if code == "" {
			t.Errorf("%s has no committee code", row.beteckning)
		}
		codes[code]++
		if code == "AU" && row.ballot {
			auBallots++
		}
	}
	if auBallots != 116 {
		t.Errorf("AU ballots = %d, want 116", auBallots)
	}
	if len(codes) < 16 {
		t.Errorf("committees = %d, want at least 16", len(codes))
	}
	if codes["UFöU"] == 0 {
		t.Error("UFöU missing")
	}

	nu := nuBallotRows(rows)
	for i := 1; i < len(nu); i++ {
		if voteLess(nu[i], nu[i-1]) {
			t.Fatalf("NU not in read order at %d: %s %s before %s %s",
				i, nu[i-1].session, nu[i-1].id, nu[i].session, nu[i].id)
		}
	}
	if len(nu) < 200 {
		t.Fatalf("NU voteringar = %d, want at least 200", len(nu))
	}
	if nu[149].id != "A971F760-DB68-4099-AB44-5E947B2A326D" {
		t.Errorf("NU[149] = %s, want the first NU7 votering", nu[149].id)
	}
	if nu[150].id != "BD92BDCE-1D0F-4748-A5DF-B17F18A66DE7" {
		t.Errorf("NU[150] = %s, want the second NU7 votering", nu[150].id)
	}

	wantDup := map[string]bool{
		"2023/24|NU1|2":  true,
		"2023/24|NU5|2":  true,
		"2023/24|NU14|1": true,
		"2024/25|NU1|2":  true,
		"2025/26|NU1|1":  true,
		"2025/26|NU7|2":  true,
		"2025/26|NU16|8": true,
		"2025/26|NU22|1": true,
		"2025/26|NU27|5": true,
	}
	counts := map[string]int{}
	for _, row := range nu {
		counts[row.session+"|"+row.beteckning+"|"+row.punkt]++
	}
	for key, n := range counts {
		dup := n > 1
		if wantDup[key] != dup {
			t.Errorf("%s count = %d, dup = %v, want dup %v", key, n, dup, wantDup[key])
		}
	}
	for key := range wantDup {
		if counts[key] < 2 {
			t.Errorf("%s missing from NU", key)
		}
	}

	var sawDistinct bool
	for _, row := range nu {
		if row.session == "2023/24" && row.beteckning == "NU1" && row.punkt == "2" {
			sawDistinct = true
			switch row.id {
			case "35B5E914-F3FD-45C4-AE5F-A9539273E361", "6B741FC0-B732-4CBA-90FF-A16AD5AA1044":
			default:
				t.Errorf("NU1:2 2023/24 id = %s", row.id)
			}
		}
	}
	if !sawDistinct {
		t.Error("NU1:2 2023/24 pair missing")
	}
}

func nuBallotRows(rows []voteRow) []voteRow {
	var nu []voteRow
	for _, row := range rows {
		if row.ballot && domain.CommitteeCode(row.beteckning) == "NU" {
			nu = append(nu, row)
		}
	}
	return nu
}

func voteLess(a, b voteRow) bool {
	if a.session != b.session {
		return a.session < b.session
	}
	an, bn := digits(a.beteckning), digits(b.beteckning)
	if an != bn {
		return an < bn
	}
	ap, bp := digits(a.punkt), digits(b.punkt)
	if ap != bp {
		return ap < bp
	}
	return a.id < b.id
}

func digits(s string) int {
	n := ""
	for _, r := range s {
		if r >= '0' && r <= '9' {
			n += string(r)
		}
	}
	v, _ := strconv.Atoi(n)
	return v
}
