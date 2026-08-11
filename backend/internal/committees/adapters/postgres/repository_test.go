package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/committees/adapters/postgres"
)

func connectTestDB(t *testing.T) *pgxpool.Pool {
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

// The list must come from the record, so UFöU has to appear: it decided 17
// voteringar in 2022-2026 and is absent from any hardcoded set of "the 15
// standing committees".
func TestListForPeriod_IncludesJointCommittee(t *testing.T) {
	pool := connectTestDB(t)
	repo := postgres.NewRepository(pool)

	got, err := repo.ListForPeriod(context.Background(), "2022-2026")
	if err != nil {
		t.Fatalf("ListForPeriod: %v", err)
	}
	if len(got) < 16 {
		t.Errorf("got %d committees, want at least 16 (15 standing + UFöU)", len(got))
	}

	var sawUFoU bool
	for i, c := range got {
		if c.Code == "UFöU" {
			sawUFoU = true
			if c.Name != "Sammansatta utrikes- och försvarsutskottet" {
				t.Errorf("UFöU name = %q", c.Name)
			}
			if c.Voteringar == 0 {
				t.Error("UFöU has 0 voteringar but was derived from the record")
			}
		}
		if i > 0 && got[i-1].Code > c.Code {
			t.Errorf("not alphabetical: %q before %q", got[i-1].Code, c.Code)
		}
	}
	if !sawUFoU {
		t.Error("UFöU missing — the list is gating on a hardcoded set")
	}
}

// The published voteringar figure has to equal the number of voteringar the
// record holds. It previously did not: keying on (beteckning, förslagspunkt)
// collapsed 2022/23 FiU1 with 2023/24 FiU1, because betänkande numbering
// restarts each riksmöte, and reported 1 870 where the record holds 2 576 — a
// 27 % undercount published as a fact, and one that contradicted the feature's
// own stated ground truth of 2 558.
//
// Asserted two ways. First against the count computed independently in SQL,
// which is the definition of correctness and cannot go stale. Second against
// 2 576 as a floor rather than an exact figure: 2022-2026 is still an open
// mandate period, so the record grows with every ingestion run, and pinning
// the exact number would turn this test red for a reason that is not a defect.
// Either assertion fails loudly on the 1 870 behaviour.
func TestListForPeriod_CountsEveryVoteringInTheRecord(t *testing.T) {
	pool := connectTestDB(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	got, err := repo.ListForPeriod(ctx, "2022-2026")
	if err != nil {
		t.Fatalf("ListForPeriod: %v", err)
	}
	sum := 0
	for _, c := range got {
		sum += c.Voteringar
	}

	var inRecord int
	err = pool.QueryRow(ctx, `
		SELECT count(DISTINCT v.votering_id)
		FROM votes v
		JOIN mandate_periods p ON v.session = ANY(p.riksmoten)
		WHERE p.code = '2022-2026'
		  AND v.beteckning <> ''`).Scan(&inRecord)
	if err != nil {
		t.Fatalf("count voteringar in record: %v", err)
	}

	if sum != inRecord {
		t.Errorf("committee voteringar sum to %d, but the record holds %d distinct voteringar", sum, inRecord)
	}
	// 2 576 was measured on 2026-08-10; the period is open, so it can only grow.
	if sum < 2576 {
		t.Errorf("voteringar sum = %d, want at least 2576 (measured ground truth for 2022-2026)", sum)
	}
}

// The record holds förslagspunkter that were decided by more than one votering
// — 2023/24 NU1 punkt 2 among them. Counting per förslagspunkt discards those,
// so this pins the fact that they are counted, since votering_id is the
// record's own identity for a votering.
func TestListForPeriod_KeepsMultipleVoteringarOnOneForslagspunkt(t *testing.T) {
	pool := connectTestDB(t)
	ctx := context.Background()

	var doubled int
	err := pool.QueryRow(ctx, `
		SELECT count(*) FROM (
			SELECT v.session, v.beteckning, v.forslagspunkt
			FROM votes v
			JOIN mandate_periods p ON v.session = ANY(p.riksmoten)
			WHERE p.code = '2022-2026'
			  AND v.beteckning <> ''
			GROUP BY 1, 2, 3
			HAVING count(DISTINCT v.votering_id) > 1
		) x`).Scan(&doubled)
	if err != nil {
		t.Fatalf("count multi-votering förslagspunkter: %v", err)
	}
	if doubled == 0 {
		t.Skip("no förslagspunkt in this dataset carries two voteringar — nothing to distinguish")
	}

	repo := postgres.NewRepository(pool)
	got, err := repo.ListForPeriod(ctx, "2022-2026")
	if err != nil {
		t.Fatalf("ListForPeriod: %v", err)
	}
	sum := 0
	for _, c := range got {
		sum += c.Voteringar
	}

	var byForslagspunkt int
	err = pool.QueryRow(ctx, `
		SELECT count(DISTINCT v.session || ':' || v.beteckning || ':' || v.forslagspunkt)
		FROM votes v
		JOIN mandate_periods p ON v.session = ANY(p.riksmoten)
		WHERE p.code = '2022-2026'
		  AND v.beteckning <> ''`).Scan(&byForslagspunkt)
	if err != nil {
		t.Fatalf("count förslagspunkter: %v", err)
	}
	if sum != byForslagspunkt+doubled {
		t.Errorf("voteringar sum = %d, want %d (%d förslagspunkter + %d extra voteringar)",
			sum, byForslagspunkt+doubled, byForslagspunkt, doubled)
	}
}

// FindInPeriod must resolve the same committee the list does, with the same
// count — the detail path narrows the query and must not narrow the answer.
func TestFindInPeriod_MatchesTheList(t *testing.T) {
	pool := connectTestDB(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	all, err := repo.ListForPeriod(ctx, "2022-2026")
	if err != nil {
		t.Fatalf("ListForPeriod: %v", err)
	}
	for _, want := range all {
		got, err := repo.FindInPeriod(ctx, "2022-2026", want.Code)
		if err != nil {
			t.Fatalf("FindInPeriod(%q): %v", want.Code, err)
		}
		if got == nil {
			t.Errorf("FindInPeriod(%q) = nil, but the list contains it", want.Code)
			continue
		}
		if got.Voteringar != want.Voteringar {
			t.Errorf("FindInPeriod(%q).Voteringar = %d, list says %d", want.Code, got.Voteringar, want.Voteringar)
		}
	}

	// A code that decided nothing is nil, not a zero-count committee.
	got, err := repo.FindInPeriod(ctx, "2022-2026", "ZZU")
	if err != nil {
		t.Fatalf("FindInPeriod(ZZU): %v", err)
	}
	if got != nil {
		t.Errorf("FindInPeriod(ZZU) = %+v, want nil", got)
	}
}

func TestPeriodExists(t *testing.T) {
	pool := connectTestDB(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	ok, err := repo.PeriodExists(ctx, "2022-2026")
	if err != nil {
		t.Fatalf("PeriodExists: %v", err)
	}
	if !ok {
		t.Error("PeriodExists(2022-2026) = false, but the period is seeded")
	}

	ok, err = repo.PeriodExists(ctx, "2018-2022")
	if err != nil {
		t.Fatalf("PeriodExists: %v", err)
	}
	if ok {
		t.Error("PeriodExists(2018-2022) = true, but we hold no such period")
	}
}

func TestDecidedBudgetYearExists(t *testing.T) {
	pool := connectTestDB(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	newest, err := repo.NewestDecidedBudgetYear(ctx)
	if err != nil {
		t.Fatalf("NewestDecidedBudgetYear: %v", err)
	}
	ok, err := repo.DecidedBudgetYearExists(ctx, newest)
	if err != nil {
		t.Fatalf("DecidedBudgetYearExists: %v", err)
	}
	if !ok {
		t.Errorf("DecidedBudgetYearExists(%d) = false for the newest decided year", newest)
	}

	// 1999 predates every allocation we hold. Answering for it would publish
	// "FiU bereder UO2: 0 tkr", which is false rather than merely absent.
	ok, err = repo.DecidedBudgetYearExists(ctx, 1999)
	if err != nil {
		t.Fatalf("DecidedBudgetYearExists: %v", err)
	}
	if ok {
		t.Error("DecidedBudgetYearExists(1999) = true, but we hold no allocations for it")
	}
}

func TestAreasFor_NoShareOfTotal(t *testing.T) {
	pool := connectTestDB(t)
	repo := postgres.NewRepository(pool)

	got, err := repo.AreasFor(context.Background(), "FiU", 2026)
	if err != nil {
		t.Fatalf("AreasFor: %v", err)
	}
	// FiU bereder UO2, UO25, UO26, UO27.
	if len(got) != 4 {
		t.Errorf("FiU areas = %d, want 4", len(got))
	}
	for _, a := range got {
		if a.Code == "" || a.Name == "" {
			t.Errorf("area missing code or name: %+v", a)
		}
	}
}

// A revision of the Bilaga adds one row for the utgiftsområde that moved and
// leaves the other 26 untouched — that is exactly what
// docs/data-sources/utskott-utgiftsomrade.md tells the maintainer to do, and
// SFS 2026:1349 takes effect 2026-09-01.
//
// Selecting the rows carrying the global max(in_force_from) breaks on that path:
// the newest date belongs to one row only, so the whole allocation collapses to
// that single row and every other committee page shows zero utgiftsområden,
// silently and with no error. The allocation therefore has to be resolved per
// uo_code. This inserts a one-row partial revision with an already-effective
// date and asserts both halves: the moved UO follows its new utskott, and the
// other 26 survive.
func TestAreasFor_SurvivesAPartialRevision(t *testing.T) {
	pool := connectTestDB(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	const (
		movedUO     = "UO13" // AU per the 2014 allocation.
		newUtskott  = "CU"
		oldUtskott  = "AU"
		revisedFrom = "2026-01-01" // Already in force, so it must take effect.
	)

	// Self-healing: a killed run would otherwise leave the row behind and make
	// the next insert fail on the primary key before cleanup is registered.
	del := func() error {
		_, err := pool.Exec(ctx,
			`DELETE FROM utskott_utgiftsomrade WHERE uo_code = $1 AND in_force_from = $2`,
			movedUO, revisedFrom)
		return err
	}
	if err := del(); err != nil {
		t.Fatalf("pre-clean revision row: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO utskott_utgiftsomrade (utskott_code, uo_code, in_force_from) VALUES ($1, $2, $3)`,
		newUtskott, movedUO, revisedFrom,
	); err != nil {
		t.Fatalf("insert partial revision row: %v", err)
	}
	t.Cleanup(func() {
		if err := del(); err != nil {
			t.Errorf("cleanup: delete revision row: %v", err)
		}
	})

	year, err := repo.NewestDecidedBudgetYear(ctx)
	if err != nil {
		t.Fatalf("NewestDecidedBudgetYear: %v", err)
	}

	rows, err := pool.Query(ctx, `SELECT DISTINCT utskott_code FROM utskott_utgiftsomrade`)
	if err != nil {
		t.Fatalf("list utskott: %v", err)
	}
	var utskott []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			t.Fatalf("scan utskott: %v", err)
		}
		utskott = append(utskott, c)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("list utskott: %v", err)
	}

	seen := map[string]string{} // uo_code -> utskott it was returned under
	for _, u := range utskott {
		areas, err := repo.AreasFor(ctx, u, year)
		if err != nil {
			t.Fatalf("AreasFor(%q): %v", u, err)
		}
		for _, a := range areas {
			if prev, dup := seen[a.Code]; dup {
				t.Errorf("%s returned under both %s and %s — the allocation is a partition", a.Code, prev, u)
			}
			seen[a.Code] = u
		}
	}

	if len(seen) != 27 {
		t.Errorf("a one-row revision left %d utgiftsområden allocated, want 27 — the other 26 must survive", len(seen))
	}
	if got := seen[movedUO]; got != newUtskott {
		t.Errorf("%s is allocated to %q, want %q — the newest in-force row for that UO must win", movedUO, got, newUtskott)
	}

	// And the UO must be gone from its previous utskott, not present in both.
	old, err := repo.AreasFor(ctx, oldUtskott, year)
	if err != nil {
		t.Fatalf("AreasFor(%q): %v", oldUtskott, err)
	}
	for _, a := range old {
		if a.Code == movedUO {
			t.Errorf("%s still allocated to %s after the revision moved it", movedUO, oldUtskott)
		}
	}
}

// budget_years permits more than one row per year (UNIQUE (year, status),
// status in {'decided', 'proposed'}). Without pinning to status = 'decided',
// AreasFor's join on budget_years fans out one row per matching budget_years
// row and doubles every area's amount — an authoritative-looking wrong
// number. This inserts a second, 'proposed' row for 2026 to prove the
// repository does not do that, then deletes exactly the row it inserted.
func TestAreasFor_IgnoresProposedBudgetYear(t *testing.T) {
	pool := connectTestDB(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	// Self-healing: UNIQUE (year, status) means a leftover row from a killed run
	// would fail the insert *before* cleanup is registered, leaving this test
	// permanently red for a reason unrelated to the code under test.
	if _, err := pool.Exec(ctx,
		`DELETE FROM budget_years WHERE year = $1 AND status = 'proposed'`, 2026,
	); err != nil {
		t.Fatalf("pre-clean proposed budget_years row: %v", err)
	}

	var proposedID int
	err := pool.QueryRow(ctx,
		`INSERT INTO budget_years (year, status) VALUES ($1, 'proposed') RETURNING id`,
		2026,
	).Scan(&proposedID)
	if err != nil {
		t.Fatalf("insert proposed budget_years row: %v", err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(ctx, `DELETE FROM budget_years WHERE id = $1`, proposedID); err != nil {
			t.Errorf("cleanup: delete budget_years id=%d: %v", proposedID, err)
		}
	})

	got, err := repo.AreasFor(ctx, "FiU", 2026)
	if err != nil {
		t.Fatalf("AreasFor: %v", err)
	}
	if len(got) != 4 {
		t.Errorf("FiU areas = %d, want 4 (a second budget_years row for 2026 must not double the result)", len(got))
	}
}

// NewestDecidedBudgetYear must skip a newer 'proposed' row and report the
// newest 'decided' one — a proposal is not yet the record, so a caller that
// defaults to it would show a figure that could still change.
func TestNewestDecidedBudgetYear_SkipsProposedYear(t *testing.T) {
	pool := connectTestDB(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	before, err := repo.NewestDecidedBudgetYear(ctx)
	if err != nil {
		t.Fatalf("NewestDecidedBudgetYear (baseline): %v", err)
	}

	// Self-healing, for the same reason as above: a leftover row from a killed
	// run must not make the insert fail before cleanup is registered.
	if _, err := pool.Exec(ctx,
		`DELETE FROM budget_years WHERE year = $1 AND status = 'proposed'`, before+1,
	); err != nil {
		t.Fatalf("pre-clean proposed budget_years row: %v", err)
	}

	var proposedID int
	err = pool.QueryRow(ctx,
		`INSERT INTO budget_years (year, status) VALUES ($1, 'proposed') RETURNING id`,
		before+1,
	).Scan(&proposedID)
	if err != nil {
		t.Fatalf("insert proposed budget_years row: %v", err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(ctx, `DELETE FROM budget_years WHERE id = $1`, proposedID); err != nil {
			t.Errorf("cleanup: delete budget_years id=%d: %v", proposedID, err)
		}
	})

	got, err := repo.NewestDecidedBudgetYear(ctx)
	if err != nil {
		t.Fatalf("NewestDecidedBudgetYear: %v", err)
	}
	if got != before {
		t.Errorf("NewestDecidedBudgetYear = %d, want %d (a proposed year must not count as decided)", got, before)
	}
}
