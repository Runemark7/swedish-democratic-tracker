package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/goals/adapters/postgres"
)

func connectGoalsCommitteeDB(t *testing.T) *pgxpool.Pool {
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

// MJU is the most-referenced committee in the seed data: it appears in the
// goals of every party. Goals must come back ordered by party alphabetically,
// since ordering them any other way would rank the parties.
func TestListByCommittee(t *testing.T) {
	pool := connectGoalsCommitteeDB(t)
	repo := postgres.NewRepository(pool)

	got, err := repo.ListByCommittee(context.Background(), "MJU")
	if err != nil {
		t.Fatalf("ListByCommittee: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("no goals for MJU — the array containment query is wrong")
	}
	for i, g := range got {
		var found bool
		for _, c := range g.RelevantCommittees {
			if c == "MJU" {
				found = true
			}
		}
		if !found {
			t.Errorf("goal %d returned for MJU but lists %v", g.ID, g.RelevantCommittees)
		}
		if i > 0 && got[i-1].Party > g.Party {
			t.Errorf("not alphabetical by party: %q before %q", got[i-1].Party, g.Party)
		}
	}

	// KrU is the one committee with no seeded goals. It must return an empty
	// slice and no error: the caller renders the row and attributes the gap to
	// our seed data, so "no goals" must be a normal answer, not a failure.
	empty, err := repo.ListByCommittee(context.Background(), "KrU")
	if err != nil {
		t.Fatalf("ListByCommittee(KrU): %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("KrU returned %d goals, expected 0 from seed data", len(empty))
	}
}

// TestListByCommittee_ExcludesElectionCycle2026 guards recordViewFilter on
// this specific query. CU is where the leak is verified live: S has two
// goals referencing CU, id 4 (election_cycle 2022) and id 49 (election_cycle
// 2026). Only id 4 belongs in the record view — the 2022-2026 voting record
// cannot speak to a promise made in 2026. Without recordViewFilter,
// ListByCommittee("CU") returns both, so S appears twice instead of once.
// The same defect is present against SoU, AU, FiU, JuU, UbU and SfU; CU is
// picked because it is the smallest committee that exhibits it.
func TestListByCommittee_ExcludesElectionCycle2026(t *testing.T) {
	pool := connectGoalsCommitteeDB(t)
	ctx := context.Background()
	repo := postgres.NewRepository(pool)

	got, err := repo.ListByCommittee(ctx, "CU")
	if err != nil {
		t.Fatalf("ListByCommittee(CU): %v", err)
	}

	// domain.Goal does not expose election_cycle, so ask the database
	// directly which ids are 2026-tagged rather than hardcoding one — that
	// keeps the assertion correct even if seed data is renumbered.
	rows, err := pool.Query(ctx, `SELECT id FROM party_goals WHERE election_cycle = '2026'`)
	if err != nil {
		t.Fatalf("query 2026-tagged goals: %v", err)
	}
	is2026 := map[int]bool{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		is2026[id] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate 2026-tagged goals: %v", err)
	}

	sCount := 0
	for _, g := range got {
		if is2026[g.ID] {
			t.Errorf("goal %d has election_cycle = 2026 but was returned by ListByCommittee(CU) — recordViewFilter missing", g.ID)
		}
		if g.Party == "S" {
			sCount++
		}
	}

	// Pinned to the specific count that changes when the filter is dropped:
	// S has one 2022 CU goal (id 4) and one 2026 CU goal (id 49). Only the
	// former belongs here.
	if sCount != 1 {
		t.Errorf("S appeared %d times in ListByCommittee(CU), want 1 — the 2026-tagged S goal (id 49) was not excluded", sCount)
	}
}
