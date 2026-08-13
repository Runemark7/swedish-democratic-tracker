package postgres_test

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/votes/adapters/postgres"
)

// The only mandate period seeded in the dev database. Every test below is
// scoped to it, the same way GET /committees/{code} requires ?period=.
const testPeriod = "2022-2026"

func connectVotesCommitteeDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration tests")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect test DB: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// TestPeriodExists covers a known-good period and a period the record does
// not hold. "2018-2022" is the specific example committees.Service's own
// doc comment names for why an unrecognised period must be refused rather
// than answered with an empty result.
func TestPeriodExists(t *testing.T) {
	pool := connectVotesCommitteeDB(t)
	repo := postgres.NewRepository(pool)

	ok, err := repo.PeriodExists(context.Background(), testPeriod)
	if err != nil {
		t.Fatalf("PeriodExists(%q): %v", testPeriod, err)
	}
	if !ok {
		t.Errorf("PeriodExists(%q) = false, want true — this period is seeded", testPeriod)
	}

	ok, err = repo.PeriodExists(context.Background(), "2018-2022")
	if err != nil {
		t.Fatalf("PeriodExists(2018-2022): %v", err)
	}
	if ok {
		t.Errorf("PeriodExists(2018-2022) = true, want false — this period predates the record")
	}
}

func TestListByCommitteeWithPositions(t *testing.T) {
	pool := connectVotesCommitteeDB(t)
	repo := postgres.NewRepository(pool)

	// limit 200 covers the whole of AU's 116 voteringar in one page. This
	// matters beyond just exercising pagination: only 2 of AU's 24
	// 2022/23 voteringar carry a "-" (no-party) ballot, both past position
	// 20 in beteckning order, so a limit of 20 would let the "-" exclusion
	// check below pass vacuously even with the exclusion removed.
	got, total, err := repo.ListByCommitteeWithPositions(context.Background(), testPeriod, "AU", 200, 0)
	if err != nil {
		t.Fatalf("ListByCommitteeWithPositions: %v", err)
	}
	// AU holds 116 voteringar in 2022-2026, counted as count(DISTINCT votering_id)
	// against the dev database. Two separate bugs each shrink this number, so an
	// exact assertion is used rather than a floor — a floor would pass on both:
	//   - gating on origin_enriched drops unenriched voteringar
	//   - keying on beteckning+forslagspunkt without session collapses riksmöten
	//     into each other and yields 76
	if total != 116 {
		t.Errorf("total = %d, want 116. 76 means the query keys on beteckning+forslagspunkt "+
			"without session and has collapsed riksmöten; a value below 76 means it gates "+
			"on origin_enriched", total)
	}
	if len(got) == 0 {
		t.Fatal("no voteringar returned for AU")
	}
	for _, v := range got {
		if v.VoteringID == "" {
			t.Errorf("%s:%s has no voteringId", v.Beteckning, v.Forslagspunkt)
		}
		if len(v.PartyPositions) == 0 {
			t.Errorf("%s:%s has no party positions", v.Beteckning, v.Forslagspunkt)
		}
		for _, p := range v.PartyPositions {
			// "-" marks a ballot cast under no party affiliation (vacant or
			// independent seat). AU alone holds hundreds of these raw ballots
			// in the dev database, so this is exercised on every run, not a
			// hypothetical: RÖSTAT is a per-party list and "-" is not a party.
			if p.Party == "-" {
				t.Errorf("%s:%s has a position for party %q, which must be excluded, not reported",
					v.Beteckning, v.Forslagspunkt, p.Party)
			}
			switch p.Position {
			case "Ja", "Nej", "Avstår", "Frånvarande", "Delad":
			default:
				t.Errorf("unexpected position %q for %s", p.Position, p.Party)
			}
		}
	}
}

// AU10 must not sort before AU9. Ordering by beteckning as text does exactly
// that, which puts the record in an order no reader expects.
func TestListByCommittee_NumericOrdering(t *testing.T) {
	pool := connectVotesCommitteeDB(t)
	repo := postgres.NewRepository(pool)

	got, _, err := repo.ListByCommitteeWithPositions(context.Background(), testPeriod, "AU", 200, 0)
	if err != nil {
		t.Fatalf("ListByCommitteeWithPositions: %v", err)
	}
	prev := -1
	prevRm := ""
	for _, v := range got {
		n, err := strconv.Atoi(numericSuffix(v.Beteckning))
		if err != nil {
			continue
		}
		if v.Riksmote == prevRm && n < prev {
			t.Errorf("beteckning out of numeric order in %s: %d after %d", v.Riksmote, n, prev)
		}
		prev, prevRm = n, v.Riksmote
	}
}

// A code arrives from a URL path. Under a LIKE predicate "%" matches every
// beteckning in the table, so /committees/%/votes would serve the entire record
// as one committee's. starts_with treats "%", "_" and "%%" as literals and
// matches nothing.
//
// "" and "U" cover a second way the same failure mode can reappear even with
// starts_with in place:
//   - "" is what domain.Canonical returns for any code it does not recognise
//     (e.g. the wildcards above never reach the repository with an HTTP
//     caller in front — but the repository must refuse it on its own too),
//     and starts_with(beteckning, "") is true for every row.
//   - "U" is not a committee code — no entry in canonicalCodes is exactly
//     "U" — but it IS a prefix of three real ones (UU, UbU, UFöU), so
//     starts_with("U") alone would serve all three as one committee's
//     record. The authoritative filter is committeedomain.CommitteeCode
//     compared against the requested code, applied in the Go fold; starts_with
//     is only ever a narrowing predicate.
func TestListByCommittee_WildcardIsNotSmuggled(t *testing.T) {
	pool := connectVotesCommitteeDB(t)
	repo := postgres.NewRepository(pool)

	for _, code := range []string{"%", "_", "%%", "", "U"} {
		got, total, err := repo.ListByCommitteeWithPositions(context.Background(), testPeriod, code, 20, 0)
		if err != nil {
			t.Fatalf("ListByCommitteeWithPositions(%q): %v", code, err)
		}
		if total != 0 || len(got) != 0 {
			t.Errorf("code %q returned total=%d rows=%d, want 0 — the predicate is "+
				"treating it as a wildcard instead of a literal, or the authoritative "+
				"CommitteeCode filter is missing", code, total, len(got))
		}
	}
}

// TestListByCommittee_DistinctVoteringsKeepDistinctIDs pins one of the 18
// förslagspunkter in 2022-2026 that Riksdagen genuinely decided twice:
// NU1 punkt 2 in 2023/24 has two separate voteringar, verified against the
// dev database. Folding on (beteckning, förslagspunkt) instead of votering_id
// would merge them into one row and blend their party positions; folding on
// votering_id, as this repository does, keeps them apart and the response
// must still let a caller tell them apart.
func TestListByCommittee_DistinctVoteringsKeepDistinctIDs(t *testing.T) {
	pool := connectVotesCommitteeDB(t)
	repo := postgres.NewRepository(pool)

	got, _, err := repo.ListByCommitteeWithPositions(context.Background(), testPeriod, "NU", 500, 0)
	if err != nil {
		t.Fatalf("ListByCommitteeWithPositions: %v", err)
	}

	wantIDs := map[string]bool{
		"35B5E914-F3FD-45C4-AE5F-A9539273E361": false,
		"6B741FC0-B732-4CBA-90FF-A16AD5AA1044": false,
	}
	matches := 0
	for _, v := range got {
		if v.Beteckning == "NU1" && v.Forslagspunkt == "2" && v.Riksmote == "2023/24" {
			matches++
			if _, ok := wantIDs[v.VoteringID]; !ok {
				t.Errorf("NU1:2 2023/24 returned unexpected voteringId %q", v.VoteringID)
				continue
			}
			wantIDs[v.VoteringID] = true
		}
	}
	if matches != 2 {
		t.Fatalf("NU1:2 2023/24 appeared %d times, want 2 distinct voteringar", matches)
	}
	for id, seen := range wantIDs {
		if !seen {
			t.Errorf("expected voteringId %q not returned for NU1:2 2023/24", id)
		}
	}
}

// TestListByCommittee_MultipleVoteringarFlagSurvivesPaging pins the pair that
// makes this flag a server-side concern rather than something the page could
// work out for itself.
//
// NU7 punkt 2 in 2025/26 was decided by two separate voteringar
// (A971F760-… and BD92BDCE-…) whose party positions differ: C, MP, S and V
// all read as having reversed themselves between the two rows. In the
// committee's ordering those two voteringar sit at indices 149 and 150, so
// with the page size the committee page uses (50) they land on opposite sides
// of a page boundary — one closes page 3, the other opens page 4. A page-local
// duplicate count therefore sees a single row on each page and marks neither,
// which is precisely the case where the reader most needs the explanation.
//
// Both offsets are asserted rather than just one: the flag must be a property
// of the record, not of which page you happened to request.
func TestListByCommittee_MultipleVoteringarFlagSurvivesPaging(t *testing.T) {
	pool := connectVotesCommitteeDB(t)
	repo := postgres.NewRepository(pool)

	const pageSize = 50

	// Every NU förslagspunkt the dev database holds two voteringar for, verified
	// with GROUP BY session, beteckning, forslagspunkt HAVING count(DISTINCT
	// votering_id) > 1. Listed in full so the assertion below can be exact in
	// both directions: a row is marked if and only if it is one of these.
	// Marking a singly-decided point would put "avgjordes av två separata
	// voteringar" under a row where no second votering exists.
	dupPoints := map[string]bool{
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

	// The straddling pair itself: which page each half lands on.
	wantIDs := map[string]string{
		"A971F760-DB68-4099-AB44-5E947B2A326D": "offset 100 (page 3, last row)",
		"BD92BDCE-1D0F-4748-A5DF-B17F18A66DE7": "offset 150 (page 4, first row)",
	}

	seen := map[string]bool{}
	for _, offset := range []int{2 * pageSize, 3 * pageSize} {
		got, _, err := repo.ListByCommitteeWithPositions(
			context.Background(), testPeriod, "NU", pageSize, offset)
		if err != nil {
			t.Fatalf("ListByCommitteeWithPositions(offset=%d): %v", offset, err)
		}
		if len(got) != pageSize {
			t.Fatalf("offset %d returned %d rows, want %d — NU no longer holds enough "+
				"voteringar for this test's page arithmetic", offset, len(got), pageSize)
		}
		for _, v := range got {
			key := v.Riksmote + "|" + v.Beteckning + "|" + v.Forslagspunkt
			if want := dupPoints[key]; v.DecidedByMultipleVoteringar != want {
				t.Errorf("offset %d: %s (%s) decidedByMultipleVoteringar = %v, want %v",
					offset, key, v.VoteringID, v.DecidedByMultipleVoteringar, want)
			}
			if where, ok := wantIDs[v.VoteringID]; ok {
				seen[v.VoteringID] = true
				if !v.DecidedByMultipleVoteringar {
					t.Errorf("%s: votering %s has decidedByMultipleVoteringar=false, want true. "+
						"Its twin sits on the adjacent page, so the flag must be computed over "+
						"the whole folded slice before paging, not per page", where, v.VoteringID)
				}
			}
		}
	}

	for id, where := range wantIDs {
		if !seen[id] {
			t.Errorf("votering %s never appeared at %s — the ordering this test pins has moved",
				id, where)
		}
	}
}

// TestListByCommittee_TieIsSurfaced pins two real ties in the dev database
// where no single position holds a strict plurality among a party's members.
// The previous tie-break (ORDER BY count(*) DESC, vote_result) resolved these
// alphabetically, which published "Frånvarande" for L on the second votering
// below even though exactly as many L members voted Ja — a flatly wrong
// statement about how half the party voted. Both must come back "Delad".
func TestListByCommittee_TieIsSurfaced(t *testing.T) {
	pool := connectVotesCommitteeDB(t)
	repo := postgres.NewRepository(pool)

	cases := []struct {
		committee  string
		voteringID string
		party      string
	}{
		// KrU7 punkt 5, 2022/23: MP split 7 Avstår / 7 Ja.
		{"KrU", "922CFC33-E4D2-4F3C-A819-E2927BE94B51", "MP"},
		// MJU19 punkt 4, 2025/26: L split 8 Frånvarande / 8 Ja.
		{"MJU", "9D3AD261-F412-4DCB-80E8-6D36D3465864", "L"},
	}

	for _, c := range cases {
		got, _, err := repo.ListByCommitteeWithPositions(context.Background(), testPeriod, c.committee, 500, 0)
		if err != nil {
			t.Fatalf("ListByCommitteeWithPositions(%s): %v", c.committee, err)
		}
		var found *string
		for _, v := range got {
			if v.VoteringID != c.voteringID {
				continue
			}
			for _, p := range v.PartyPositions {
				if p.Party == c.party {
					pos := p.Position
					found = &pos
				}
			}
		}
		if found == nil {
			t.Fatalf("votering %s (%s) party %s not found in %s results",
				c.voteringID, c.committee, c.party, c.committee)
		}
		if *found != "Delad" {
			t.Errorf("votering %s party %s = %q, want \"Delad\" — a tied vote must not be "+
				"resolved to whichever position sorts first alphabetically",
				c.voteringID, c.party, *found)
		}
	}
}

func numericSuffix(s string) string {
	out := ""
	for _, r := range s {
		if r >= '0' && r <= '9' {
			out += string(r)
		} else if out != "" {
			break
		}
	}
	return out
}
