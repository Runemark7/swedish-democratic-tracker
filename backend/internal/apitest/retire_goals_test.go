package apitest

import (
	"context"
	"testing"
)

// Issue #164: goals, keyword matching, scorecards and the old promises feature
// are gone.

// The #156 archive must exist and survive: the drop migration may only remove
// the retired objects, never loften.
func TestRetireGoalsKeepsTheArchiveAndDropsTheRest(t *testing.T) {
	h := Start(t, chainDir(t, "beslut-list"))
	ctx := context.Background()

	var archived int
	if err := h.DB.QueryRow(ctx, `SELECT count(*) FROM loften WHERE visibility = 'archived'`).Scan(&archived); err != nil {
		t.Fatal(err)
	}
	if archived == 0 {
		t.Fatal("no archived Löften: the #156 archive must exist before the retired goals are dropped")
	}

	for _, rel := range []string{"party_goals", "goal_vote_matches", "promises", "promise_vote_matches", "party_scorecards",
		"party_goals_id_seq", "goal_vote_matches_id_seq", "promises_id_seq", "promise_vote_matches_id_seq"} {
		var reg *string
		if err := h.DB.QueryRow(ctx, `SELECT to_regclass($1)::text`, rel).Scan(&reg); err != nil {
			t.Fatal(err)
		}
		if reg != nil {
			t.Errorf("%s still exists after migration 000064", rel)
		}
	}
}
