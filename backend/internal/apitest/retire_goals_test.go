package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Issue #164: goals, keyword matching, scorecards and the old promises feature
// are gone; the retired endpoints answer 410 with a pointer to the Löften API.

func TestRetiredEndpointsPointToLoften(t *testing.T) {
	h := Start(t, chainDir(t, "beslut-list"))
	for _, path := range []string{
		"/api/parties",
		"/api/parties/S/goals",
		"/api/parties/S/goals/1/votes",
		"/api/committees/FiU/goals",
		"/api/politicians/abc/promises",
	} {
		rec := h.Do(httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusGone {
			t.Errorf("GET %s = %d, want 410", path, rec.Code)
			continue
		}
		var body map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Errorf("GET %s: body not JSON: %v", path, err)
			continue
		}
		if body["loften"] != "/api/loften" {
			t.Errorf("GET %s: pointer = %q, want /api/loften", path, body["loften"])
		}
	}
}

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
