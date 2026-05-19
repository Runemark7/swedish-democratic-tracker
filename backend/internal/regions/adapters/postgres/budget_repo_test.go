package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	regionsPG "riksdagskollen/internal/regions/adapters/postgres"
	"riksdagskollen/internal/regions/ports"
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

func cleanBudgetSnapshots(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `DELETE FROM region_budget_snapshots WHERE region_code LIKE 'ZZ%'`)
	if err != nil {
		t.Fatalf("clean snapshots: %v", err)
	}
}

func TestUpsertRegionBudgetSnapshots_Idempotency(t *testing.T) {
	pool := connectTestDB(t)
	cleanBudgetSnapshots(t, pool)
	t.Cleanup(func() { cleanBudgetSnapshots(t, pool) })

	repo := regionsPG.NewRepository(pool)
	ctx := context.Background()

	snapshots := []ports.RegionBudgetSnapshot{
		{RegionCode: "ZZ01", AreaName: "TestArea A", Year: 2022, ValueMnkr: 100, TotalMnkr: 500, Pct: 20.0},
		{RegionCode: "ZZ01", AreaName: "TestArea B", Year: 2022, ValueMnkr: 200, TotalMnkr: 500, Pct: 40.0},
		{RegionCode: "ZZ01", AreaName: "TestArea A", Year: 2023, ValueMnkr: 110, TotalMnkr: 550, Pct: 20.0},
	}

	// First insert
	n, err := repo.UpsertRegionBudgetSnapshots(ctx, snapshots)
	if err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 rows affected on first insert, got %d", n)
	}

	// Second upsert with changed values
	updated := []ports.RegionBudgetSnapshot{
		{RegionCode: "ZZ01", AreaName: "TestArea A", Year: 2022, ValueMnkr: 150, TotalMnkr: 500, Pct: 30.0},
		{RegionCode: "ZZ01", AreaName: "TestArea B", Year: 2022, ValueMnkr: 250, TotalMnkr: 500, Pct: 50.0},
		{RegionCode: "ZZ01", AreaName: "TestArea A", Year: 2023, ValueMnkr: 120, TotalMnkr: 550, Pct: 21.8},
	}
	_, err = repo.UpsertRegionBudgetSnapshots(ctx, updated)
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	// Verify exactly 3 rows in DB (not 6)
	var count int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM region_budget_snapshots WHERE region_code = 'ZZ01'`).Scan(&count)
	if err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 rows after upsert (not 6), got %d", count)
	}

	// Verify values were updated
	var val float64
	err = pool.QueryRow(ctx, `SELECT value_mnkr FROM region_budget_snapshots WHERE region_code='ZZ01' AND area_name='TestArea A' AND year=2022`).Scan(&val)
	if err != nil {
		t.Fatalf("read updated value: %v", err)
	}
	if val != 150 {
		t.Errorf("expected updated value_mnkr=150, got %f", val)
	}
}

func TestGetRegionBudgetHistory(t *testing.T) {
	pool := connectTestDB(t)
	cleanBudgetSnapshots(t, pool)
	t.Cleanup(func() { cleanBudgetSnapshots(t, pool) })

	repo := regionsPG.NewRepository(pool)
	ctx := context.Background()

	// Seed: 2 regions × 3 years
	snapshots := []ports.RegionBudgetSnapshot{
		{RegionCode: "ZZ02", AreaName: "AreaX", Year: 2021, ValueMnkr: 10, TotalMnkr: 100, Pct: 10},
		{RegionCode: "ZZ02", AreaName: "AreaX", Year: 2022, ValueMnkr: 11, TotalMnkr: 100, Pct: 11},
		{RegionCode: "ZZ02", AreaName: "AreaX", Year: 2023, ValueMnkr: 12, TotalMnkr: 100, Pct: 12},
		{RegionCode: "ZZ03", AreaName: "AreaX", Year: 2021, ValueMnkr: 20, TotalMnkr: 200, Pct: 10},
		{RegionCode: "ZZ03", AreaName: "AreaX", Year: 2022, ValueMnkr: 21, TotalMnkr: 200, Pct: 10.5},
		{RegionCode: "ZZ03", AreaName: "AreaX", Year: 2023, ValueMnkr: 22, TotalMnkr: 200, Pct: 11},
	}
	if _, err := repo.UpsertRegionBudgetSnapshots(ctx, snapshots); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Query ZZ02 for only 2 years
	history, err := repo.GetRegionBudgetHistory(ctx, "ZZ02", []int{2021, 2023})
	if err != nil {
		t.Fatalf("GetRegionBudgetHistory: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(history))
	}
	for _, s := range history {
		if s.RegionCode != "ZZ02" {
			t.Errorf("unexpected region code: %s", s.RegionCode)
		}
		if s.Year != 2021 && s.Year != 2023 {
			t.Errorf("unexpected year: %d", s.Year)
		}
	}
}

func TestGetAreaAcrossRegions(t *testing.T) {
	pool := connectTestDB(t)
	cleanBudgetSnapshots(t, pool)
	t.Cleanup(func() { cleanBudgetSnapshots(t, pool) })

	repo := regionsPG.NewRepository(pool)
	ctx := context.Background()

	snapshots := []ports.RegionBudgetSnapshot{
		{RegionCode: "ZZ04", AreaName: "AreaY", Year: 2023, ValueMnkr: 50, TotalMnkr: 500, Pct: 10},
		{RegionCode: "ZZ05", AreaName: "AreaY", Year: 2023, ValueMnkr: 60, TotalMnkr: 600, Pct: 10},
		{RegionCode: "ZZ06", AreaName: "AreaY", Year: 2023, ValueMnkr: 70, TotalMnkr: 700, Pct: 10},
		// Different year — should not be returned
		{RegionCode: "ZZ04", AreaName: "AreaY", Year: 2022, ValueMnkr: 40, TotalMnkr: 400, Pct: 10},
	}
	if _, err := repo.UpsertRegionBudgetSnapshots(ctx, snapshots); err != nil {
		t.Fatalf("seed: %v", err)
	}

	datapoints, err := repo.GetAreaAcrossRegions(ctx, "AreaY", 2023)
	if err != nil {
		t.Fatalf("GetAreaAcrossRegions: %v", err)
	}

	found := 0
	for _, d := range datapoints {
		switch d.RegionCode {
		case "ZZ04", "ZZ05", "ZZ06":
			found++
		}
	}
	if found != 3 {
		t.Errorf("expected 3 datapoints for 2023, got %d (total returned: %d)", found, len(datapoints))
	}
}
