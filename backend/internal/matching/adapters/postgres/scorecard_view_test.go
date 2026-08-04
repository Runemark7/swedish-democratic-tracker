package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func connectScorecardTestDB(t *testing.T) *pgxpool.Pool {
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

// cleanScorecardFixtures removes the ZZ sentinel party rows. goal_vote_matches
// rows cascade from party_goals.
func cleanScorecardFixtures(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	for _, q := range []string{
		`DELETE FROM votes WHERE party = 'ZZ'`,
		`DELETE FROM party_goals WHERE party = 'ZZ'`,
		`DELETE FROM politicians WHERE intressent_id LIKE 'ZZ-%'`,
		`DELETE FROM parties WHERE code = 'ZZ'`,
	} {
		if _, err := pool.Exec(ctx, q); err != nil {
			t.Fatalf("clean (%s): %v", q, err)
		}
	}
}

// seedParty inserts the ZZ sentinel party. party_goals.party is a FK to
// parties.code (migration 000021), so this must exist before seeding goals.
func seedParty(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO parties (code, name, color_hex, text_hex)
		VALUES ('ZZ', 'Testpartiet', '#000000', '#ffffff')
		ON CONFLICT (code) DO NOTHING`)
	if err != nil {
		t.Fatalf("seed party: %v", err)
	}
}

// seedGoal inserts a ZZ-party goal and returns its id.
func seedGoal(t *testing.T, pool *pgxpool.Pool, text string) int {
	t.Helper()
	var id int
	err := pool.QueryRow(context.Background(), `
		INSERT INTO party_goals (party, goal_text, topic, specificity, source_document)
		VALUES ('ZZ', $1, 'test', 'directional', 'test')
		RETURNING id`, text).Scan(&id)
	if err != nil {
		t.Fatalf("seed goal: %v", err)
	}
	return id
}

func seedMatch(t *testing.T, pool *pgxpool.Pool, goalID int, bet, punkt, direction string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO goal_vote_matches
			(goal_id, beteckning, forslagspunkt, relevance_score, aligned_direction)
		VALUES ($1, $2, $3, 0.75, $4)`, goalID, bet, punkt, direction)
	if err != nil {
		t.Fatalf("seed match: %v", err)
	}
}

// seedBallots inserts one votes row per result, each with its own synthetic
// politician, for the ZZ party on the given vote point.
func seedBallots(t *testing.T, pool *pgxpool.Pool, bet, punkt string, results []string) {
	t.Helper()
	ctx := context.Background()
	for i, res := range results {
		pid := fmt.Sprintf("ZZ-%s%s-%d", bet, punkt, i)
		if _, err := pool.Exec(ctx, `
			INSERT INTO politicians (intressent_id, first_name, last_name, party)
			VALUES ($1, 'Test', 'Person', 'ZZ')
			ON CONFLICT (intressent_id) DO NOTHING`, pid); err != nil {
			t.Fatalf("seed politician: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO votes
				(votering_id, politician_id, party, vote_result, beteckning, forslagspunkt, session)
			VALUES ($1, $2, 'ZZ', $3, $4, $5, '2025/26')`,
			pid+"-v", pid, res, bet, punkt); err != nil {
			t.Fatalf("seed vote: %v", err)
		}
	}
}

type scorecard struct {
	relevantVotes int
}

func readScorecard(t *testing.T, pool *pgxpool.Pool, goalID int) scorecard {
	t.Helper()
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `REFRESH MATERIALIZED VIEW party_scorecards`); err != nil {
		t.Fatalf("refresh view: %v", err)
	}
	var sc scorecard
	err := pool.QueryRow(ctx, `
		SELECT relevant_votes
		FROM party_scorecards WHERE goal_id = $1`, goalID).
		Scan(&sc.relevantVotes)
	if err != nil {
		t.Fatalf("read scorecard: %v", err)
	}
	return sc
}

// relevant_votes survives the retirement of the alignment percentage: it counts
// vote points matched to a goal, which is a fact about our matching rather than
// a judgement about the party. Direction is deliberately not consulted.
func TestScorecard_CountsMatchedVotePointsRegardlessOfDirection(t *testing.T) {
	pool := connectScorecardTestDB(t)
	cleanScorecardFixtures(t, pool)
	t.Cleanup(func() { cleanScorecardFixtures(t, pool) })
	seedParty(t, pool)

	goalID := seedGoal(t, pool, "counts matches")
	seedMatch(t, pool, goalID, "ZZ1", "1", "unclear")
	seedMatch(t, pool, goalID, "ZZ1", "2", "Ja")
	seedMatch(t, pool, goalID, "ZZ2", "1", "Nej")
	seedBallots(t, pool, "ZZ1", "1", []string{"Ja", "Ja"})

	sc := readScorecard(t, pool, goalID)

	if sc.relevantVotes != 3 {
		t.Errorf("relevantVotes = %d, want 3 — every matched vote point counts, "+
			"whatever direction was inferred", sc.relevantVotes)
	}
}
