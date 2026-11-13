package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"riksdagskollen/internal/votes/adapters/postgres"
	"riksdagskollen/internal/votes/ports"
)

// The summary on a member's vote list counts every ballot in scope, not the
// page that was fetched. Gap member-vote-stats-capped-at-100.
func TestListByPoliticianSummaryCoversWholeHistory(t *testing.T) {
	pool := connectVotesCommitteeDB(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()
	const pid = "test-cap100-member"
	cleanup := func() {
		_, _ = pool.Exec(ctx, `DELETE FROM politicians WHERE intressent_id = $1`, pid)
		_, _ = pool.Exec(ctx, `DELETE FROM voteringar WHERE votering_id LIKE 'CAP100-%'`)
	}
	cleanup()
	t.Cleanup(cleanup)
	if _, err := pool.Exec(ctx, `INSERT INTO politicians (intressent_id, first_name, last_name, party) VALUES ($1,'Test','Cap','S')`, pid); err != nil {
		t.Fatal(err)
	}
	// 130 ballots: 60 in 2023/24 (30 Ja, 10 Nej, 5 Avstår, 15 Frånvarande), 70 in 2024/25 (all Ja).
	for i := 0; i < 130; i++ {
		sess, res := "2024/25", "Ja"
		if i < 60 {
			sess = "2023/24"
			switch {
			case i < 30:
				res = "Ja"
			case i < 40:
				res = "Nej"
			case i < 45:
				res = "Avstår"
			default:
				res = "Frånvarande"
			}
		}
		var ref int
		if err := pool.QueryRow(ctx, `INSERT INTO voteringar (votering_id, beteckning, forslagspunkt, session) VALUES ($1,$2,'1',$3) RETURNING id`,
			fmt.Sprintf("CAP100-%03d", i), fmt.Sprintf("CAP%d", i), sess).Scan(&ref); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO ballots (votering_ref, politician_id, party, vote_result) VALUES ($1,$2,'S',$3)`, ref, pid, res); err != nil {
			t.Fatal(err)
		}
	}

	res, err := repo.ListByPolitician(ctx, ports.ListVotesFilter{PoliticianID: pid, Page: 1, PageSize: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Votes) != 100 {
		t.Fatalf("page size: %d", len(res.Votes))
	}
	want := ports.PoliticianVoteSummary{Ja: 100, Nej: 10, Avstar: 5, Franvarande: 15, Total: 130}
	if res.Summary != want {
		t.Errorf("all: %+v, want %+v", res.Summary, want)
	}

	res, err = repo.ListByPolitician(ctx, ports.ListVotesFilter{PoliticianID: pid, Session: "2023/24", Page: 1, PageSize: 5})
	if err != nil {
		t.Fatal(err)
	}
	want = ports.PoliticianVoteSummary{Ja: 30, Nej: 10, Avstar: 5, Franvarande: 15, Total: 60}
	if res.Summary != want || res.Total != 60 {
		t.Errorf("session: %+v total %d, want %+v", res.Summary, res.Total, want)
	}

	// A page past the end still reports the summary.
	res, err = repo.ListByPolitician(ctx, ports.ListVotesFilter{PoliticianID: pid, Page: 99, PageSize: 100})
	if err != nil {
		t.Fatal(err)
	}
	if res.Summary.Total != 130 {
		t.Errorf("past-end summary: %+v", res.Summary)
	}
}
