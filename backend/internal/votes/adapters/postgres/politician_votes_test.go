package postgres_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"riksdagskollen/internal/votes/adapters/postgres"
	"riksdagskollen/internal/votes/ports"
)

// ListByPolitician picks a query strategy from the member's vote count (see
// ballotsFirstMax). Both strategies must return exactly what the plain join
// returns, for every page, with and without a session filter.
func TestListByPoliticianStrategiesAgree(t *testing.T) {
	pool := connectVotesCommitteeDB(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	var pid, session string
	if err := pool.QueryRow(ctx, `SELECT b.politician_id, vg.session FROM ballots b JOIN voteringar vg ON vg.id = b.votering_ref
		GROUP BY 1, 2 ORDER BY count(*) DESC, 1 LIMIT 1`).Scan(&pid, &session); err != nil {
		t.Skipf("no ballots seeded: %v", err)
	}

	reference := func(session string, limit, offset int) []string {
		q := `SELECT vg.votering_id FROM ballots b JOIN voteringar vg ON vg.id = b.votering_ref
			WHERE b.politician_id = $1 AND ($2 = '' OR vg.session = $2)
			ORDER BY vg.created_at DESC, vg.id DESC LIMIT $3 OFFSET $4`
		rows, err := pool.Query(ctx, q, pid, session, limit, offset)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var ids []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				t.Fatal(err)
			}
			ids = append(ids, id)
		}
		return ids
	}
	var refTotal int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM ballots b JOIN voteringar vg ON vg.id = b.votering_ref WHERE b.politician_id = $1`, pid).Scan(&refTotal); err != nil {
		t.Fatal(err)
	}

	for _, strat := range []struct {
		name string
		max  int
	}{{"ballots-first", 1 << 30}, {"date-walk", 0}} {
		restore := postgres.SetBallotsFirstMax(strat.max)
		for _, sess := range []string{"", session} {
			for _, page := range []int{1, 2, 1000} {
				name := fmt.Sprintf("%s/session=%q/page=%d", strat.name, sess, page)
				res, err := repo.ListByPolitician(ctx, ports.ListVotesFilter{PoliticianID: pid, Session: sess, Page: page, PageSize: 7})
				if err != nil {
					t.Fatalf("%s: %v", name, err)
				}
				var got []string
				for _, v := range res.Votes {
					if v.PoliticianID != pid {
						t.Fatalf("%s: ballot of %s", name, v.PoliticianID)
					}
					got = append(got, v.VoteringID)
				}
				if want := reference(sess, 7, (page-1)*7); !reflect.DeepEqual(got, want) {
					t.Errorf("%s: got %v, want %v", name, got, want)
				}
				if sess == "" && res.Total != refTotal {
					t.Errorf("%s: total %d, want %d", name, res.Total, refTotal)
				}
			}
		}
		restore()
	}

	res, err := repo.ListByPolitician(ctx, ports.ListVotesFilter{PoliticianID: "no-such-member", Page: 1, PageSize: 50})
	if err != nil || res.Total != 0 || len(res.Votes) != 0 {
		t.Fatalf("unknown member: %+v, %v", res, err)
	}
}
