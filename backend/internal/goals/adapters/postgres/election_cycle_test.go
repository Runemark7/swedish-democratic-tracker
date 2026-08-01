package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func connectGoalsTestDB(t *testing.T) *pgxpool.Pool {
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

func deref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

// Every seeded goal must carry a vintage derived from its source document.
// The 2022-2026 voting record can only speak to promises made at or before the
// 2022 election; 2026 campaign material is a different question entirely.
func TestElectionCycle_BackfilledFromSourceDocument(t *testing.T) {
	pool := connectGoalsTestDB(t)
	ctx := context.Background()

	cases := map[string]*string{
		"Valmanifest 2022":  new("2022"),
		"Tidöavtalet 2022":  new("2022"),
		"Valplattform 2022": new("2022"),
		"Valplattform 2026": new("2026"),
		"Partiprogram":      nil,
	}

	for doc, want := range cases {
		rows, err := pool.Query(ctx,
			`SELECT DISTINCT election_cycle FROM party_goals WHERE source_document = $1`, doc)
		if err != nil {
			t.Fatalf("query %s: %v", doc, err)
		}
		var got []*string
		for rows.Next() {
			var v *string
			if err := rows.Scan(&v); err != nil {
				rows.Close()
				t.Fatal(err)
			}
			got = append(got, v)
		}
		rows.Close()

		if len(got) == 0 {
			continue // that source document is not present in this database
		}
		if len(got) != 1 {
			t.Errorf("%s: expected one cycle value, got %d", doc, len(got))
			continue
		}
		if (got[0] == nil) != (want == nil) ||
			(got[0] != nil && want != nil && *got[0] != *want) {
			t.Errorf("%s: election_cycle = %s, want %s", doc, deref(got[0]), deref(want))
		}
	}
}

// No 2026 campaign material may leak into the record view. This is the
// neutrality guarantee: before the fix, S was the only party showing a current
// platform while every other party showed four-year-old material.
func TestElectionCycle_NoCampaignMaterialInRecordView(t *testing.T) {
	pool := connectGoalsTestDB(t)
	ctx := context.Background()

	var n int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM party_goals
		WHERE election_cycle = '2026'
		  AND source_document NOT ILIKE '%2026%'`).Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 0 {
		t.Errorf("%d goals marked 2026 without a 2026 source document", n)
	}

	// Every goal must be classified: either tied to an election cycle, or
	// explicitly not election material (a standing party programme).
	var unclassified int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM party_goals
		WHERE election_cycle IS NULL
		  AND (source_document ILIKE '%2022%' OR source_document ILIKE '%2026%')`).
		Scan(&unclassified); err != nil {
		t.Fatalf("query: %v", err)
	}
	if unclassified != 0 {
		t.Errorf("%d goals have a dated source document but no election_cycle", unclassified)
	}
}
