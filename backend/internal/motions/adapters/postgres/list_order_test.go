package postgres_test

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"riksdagskollen/internal/apitest"
	"riksdagskollen/internal/motions/adapters/postgres"
	"riksdagskollen/internal/motions/domain"
)

// referenceOrder is the ordering the list had before it was split into a
// decided and an undecided read: one query ranking every motion.
const referenceOrder = `
	SELECT m.dok_id FROM motions m
	LEFT JOIN LATERAL (
		SELECT max(b.systemdatum) AS at
		FROM yrkande_forslagspunkt l
		JOIN forslagspunkter p ON p.riksmote = l.riksmote AND p.beteckning = l.beteckning AND p.punkt = l.punkt
		JOIN betankanden b ON b.riksmote = l.riksmote AND b.beteckning = l.beteckning
		WHERE l.motion_dok_id = m.dok_id AND p.beslutstyp <> ''
	) dec ON true
	ORDER BY (dec.at IS NOT NULL) DESC, COALESCE(dec.at, m.motion_date::timestamp, 'epoch'::timestamp) DESC, m.dok_id DESC`

func TestListPageKeepsTheDecidedFirstOrderAcrossCursorsAndPageSizes(t *testing.T) {
	pool := apitest.OpenDB(t)
	ctx := context.Background()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	// Betänkanden: A decided on 2024-03-01, B decided on 2024-05-01, C has no
	// date (a decision that does not count), D has an undecided punkt only.
	for _, b := range []struct {
		bet   string
		at    *time.Time
		type_ string
	}{
		{"A1", ptr(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)), "utskottet"},
		{"B1", ptr(time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)), "utskottet"},
		{"C1", nil, "utskottet"},
		{"D1", ptr(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)), ""},
	} {
		exec(`INSERT INTO betankanden (riksmote, beteckning, systemdatum) VALUES ('2023/24', $1, $2)`, b.bet, b.at)
		exec(`INSERT INTO forslagspunkter (riksmote, beteckning, punkt, beslutstyp) VALUES ('2023/24', $1, '1', $2)`, b.bet, b.type_)
	}
	// Motions: m01..m12. Filing dates repeat so dok_id breaks ties; some have none.
	link := map[int]string{1: "A1", 2: "B1", 3: "B1", 4: "A1", 5: "C1", 6: "D1", 7: "A1"}
	for i := 1; i <= 12; i++ {
		id := fmt.Sprintf("H3%02d", i)
		var date any
		if i%4 != 0 {
			date = time.Date(2024, time.Month(1+i%3), 1+i%2, 0, 0, 0, 0, time.UTC)
		}
		exec(`INSERT INTO motions (dok_id, riksmote, beteckning, motion_date, systemdatum) VALUES ($1,'2023/24',$2,$3,now())`, id, fmt.Sprint(i), date)
		if bet, ok := link[i]; ok {
			exec(`INSERT INTO yrkande_forslagspunkt (motion_dok_id, yrkande_nummer, riksmote, beteckning, punkt, linked_by) VALUES ($1,1,'2023/24',$2,'1','forslag_text')`, id, bet)
		}
	}
	rows, err := pool.Query(ctx, referenceOrder)
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		want = append(want, id)
	}
	rows.Close()
	if len(want) != 12 {
		t.Fatalf("reference has %d motions", len(want))
	}

	repo := postgres.NewRepository(pool)
	for _, size := range []int{1, 2, 3, 5, 12, 50} {
		var got []string
		var after *domain.ListCursor
		for guard := 0; guard < 30; guard++ {
			page, err := repo.ListPage(ctx, domain.ListFilter{}, after, size)
			if err != nil {
				t.Fatalf("size %d: %v", size, err)
			}
			for _, r := range page {
				got = append(got, r.Motion.DokID)
			}
			if len(page) < size {
				break
			}
			c := page[len(page)-1].Cursor
			after = &c
		}
		if !slices.Equal(got, want) {
			t.Errorf("page size %d: got %v, want %v", size, got, want)
		}
	}
}

func ptr[T any](v T) *T { return &v }
