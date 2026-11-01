package apitest

import (
	"context"
	"testing"
)

func TestHarnessGivesEachCallACleanMigratedDatabase(t *testing.T) {
	ctx := context.Background()
	a := OpenDB(t)
	if _, err := a.Exec(ctx, `
		INSERT INTO politicians (intressent_id, first_name, last_name, party)
		VALUES ('iso-seam', 'Iso', 'Seam', 'S')`); err != nil {
		t.Fatalf("insert: %v", err)
	}

	b := OpenDB(t)
	var n int
	if err := b.QueryRow(ctx, `SELECT count(*) FROM politicians WHERE intressent_id = 'iso-seam'`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Fatalf("second database saw %d rows from the first", n)
	}

	var table *string
	if err := b.QueryRow(ctx, `SELECT to_regclass('public.voteringar')::text`).Scan(&table); err != nil {
		t.Fatalf("voteringar lookup: %v", err)
	}
	if table == nil || *table == "" {
		t.Fatal("migrations did not create voteringar")
	}
}
