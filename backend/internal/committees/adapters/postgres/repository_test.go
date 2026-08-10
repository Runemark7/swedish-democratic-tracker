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

// The list must come from the record, so UFöU has to appear: it decided 11
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
