package apitest

import (
	"context"
	"testing"

	"riksdagskollen/internal/motions/adapters/postgres"
	"riksdagskollen/internal/motions/domain"
)

// A motion whose source lists the same yrkande number twice is still stored,
// with one row for that number, instead of failing on motion_yrkanden_pkey.
func TestMotionUpsertStoresAMotionWithARepeatedYrkandeNumber(t *testing.T) {
	db := OpenDB(t)
	ctx := context.Background()
	repo := postgres.NewRepository(db)
	m := &domain.Motion{
		DokID: "GS02MJ7", Riksmote: "2004/05", Beteckning: "MJ7", Title: "t",
		Yrkanden: []domain.Yrkande{
			{Number: 5, Wording: "Riksdagen tillkännager", BehandlasI: "2004/05:TU17"},
			{Number: 5, Wording: "Riksdagen tillkännager", BehandlasI: "2004/05:AU6"},
		},
	}
	if err := repo.Upsert(ctx, m); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var n int
	var bet string
	if err := db.QueryRow(ctx, `SELECT count(*), min(behandlas_i) FROM motion_yrkanden WHERE dok_id = 'GS02MJ7'`).Scan(&n, &bet); err != nil {
		t.Fatal(err)
	}
	if n != 1 || bet != "2004/05:TU17" {
		t.Errorf("stored %d rows (behandlas_i %q), want 1 row from the first listing", n, bet)
	}
}

// Two signers with the same ordning must not fail the motion on motion_authors_pkey.
func TestMotionUpsertStoresAMotionWithARepeatedAuthorOrdning(t *testing.T) {
	db := OpenDB(t)
	ctx := context.Background()
	m := &domain.Motion{
		DokID: "H102Sk12", Riksmote: "2013/14", Beteckning: "Sk12", Title: "t",
		Authors: []domain.Author{{Order: 1, Name: "A"}, {Order: 1, Name: "B"}},
	}
	if err := postgres.NewRepository(db).Upsert(ctx, m); err != nil {
		t.Fatalf("upsert: %v", err)
	}
}
