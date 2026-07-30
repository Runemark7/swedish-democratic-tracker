# Goal Alignment Scoring Fix — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make party-goal alignment scoring truthful — score only vote points whose direction can be determined from the record, and report "not yet determinable" instead of a false 0%.

**Architecture:** Rewrite the `party_scorecards` materialized view to collapse per-MP ballots into one party position per vote point (majority of cast ballots, absences excluded), then score only matches whose `aligned_direction` is not `unclear`. Surface `scoredVotes` and `alignedVotes` as raw integers so the UI stops reconstructing counts from `pct × votes`, and let `alignmentPct` be NULL when nothing is scorable.

**Tech Stack:** PostgreSQL 17 (materialized view, `mode() WITHIN GROUP`, `FILTER`), Go 1.26 (pgx/v5, chi), React 19 + TypeScript, golang-migrate.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-07-29-goal-alignment-scoring-design.md`
- Migrations must be idempotent (`IF EXISTS` / `IF NOT EXISTS`) — lesson from `000024` (#75).
- `alignment_pct` must be **NULL** when `scored_votes = 0`. Never `COALESCE` it to `0` anywhere in the stack.
- `inferDirection` in `keyword_matcher.go` is **out of scope** — do not modify it.
- Raw SQL only, parameterized. No ORM, no query builder.
- Swedish UI strings; English code, comments, and commit messages.
- Verification gates (CLAUDE.md rule 4): `go build -C backend ./...` and `cd frontend && npx tsc --noEmit` must both pass before any task is considered done.
- DB integration tests follow the existing pattern in `backend/internal/regions/adapters/postgres/budget_repo_test.go`: read `TEST_DATABASE_URL`, `t.Skip` when unset, use a `ZZ` sentinel party for fixtures and clean up in `t.Cleanup`.

---

### Task 1: Rewrite the `party_scorecards` view

**Files:**
- Create: `backend/migrations/000029_fix_goal_alignment_scoring.up.sql`
- Create: `backend/migrations/000029_fix_goal_alignment_scoring.down.sql`
- Test: `backend/internal/matching/adapters/postgres/scorecard_view_test.go`

**Interfaces:**
- Consumes: existing tables `party_goals`, `goal_vote_matches`, `votes`.
- Produces: materialized view `party_scorecards` with columns
  `party TEXT, goal_id INT, goal_text TEXT, topic TEXT, relevant_votes BIGINT, scored_votes BIGINT, aligned_votes BIGINT, alignment_pct NUMERIC NULL`.
  Task 2 reads exactly these column names in this order.

- [ ] **Step 1: Write the failing integration test**

Create `backend/internal/matching/adapters/postgres/scorecard_view_test.go`:

```go
package postgres_test

import (
	"context"
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
	} {
		if _, err := pool.Exec(ctx, q); err != nil {
			t.Fatalf("clean (%s): %v", q, err)
		}
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

// seedBallots inserts one votes row per result in results, each with its own
// synthetic politician, for the ZZ party on the given vote point.
func seedBallots(t *testing.T, pool *pgxpool.Pool, bet, punkt string, results []string) {
	t.Helper()
	ctx := context.Background()
	for i, res := range results {
		pid := "ZZ-" + bet + punkt + "-" + string(rune('a'+i))
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
			bet+punkt+pid, pid, res, bet, punkt); err != nil {
			t.Fatalf("seed vote: %v", err)
		}
	}
}

type scorecard struct {
	relevantVotes int
	scoredVotes   int
	alignedVotes  int
	alignmentPct  *float64
}

func readScorecard(t *testing.T, pool *pgxpool.Pool, goalID int) scorecard {
	t.Helper()
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `REFRESH MATERIALIZED VIEW party_scorecards`); err != nil {
		t.Fatalf("refresh view: %v", err)
	}
	var sc scorecard
	err := pool.QueryRow(ctx, `
		SELECT relevant_votes, scored_votes, aligned_votes, alignment_pct
		FROM party_scorecards WHERE goal_id = $1`, goalID).
		Scan(&sc.relevantVotes, &sc.scoredVotes, &sc.alignedVotes, &sc.alignmentPct)
	if err != nil {
		t.Fatalf("read scorecard: %v", err)
	}
	return sc
}

func TestScorecard_UnclearMatchesAreNotScored(t *testing.T) {
	pool := connectScorecardTestDB(t)
	cleanScorecardFixtures(t, pool)
	t.Cleanup(func() { cleanScorecardFixtures(t, pool) })

	goalID := seedGoal(t, pool, "unclear only")
	seedMatch(t, pool, goalID, "ZZ1", "1", "unclear")
	seedBallots(t, pool, "ZZ1", "1", []string{"Ja", "Ja", "Nej"})

	sc := readScorecard(t, pool, goalID)

	if sc.relevantVotes != 1 {
		t.Errorf("relevantVotes = %d, want 1", sc.relevantVotes)
	}
	if sc.scoredVotes != 0 {
		t.Errorf("scoredVotes = %d, want 0", sc.scoredVotes)
	}
	if sc.alignmentPct != nil {
		t.Errorf("alignmentPct = %v, want nil (undeterminable must not read as 0)", *sc.alignmentPct)
	}
}

func TestScorecard_PartyPositionIsMajorityOfCastBallots(t *testing.T) {
	pool := connectScorecardTestDB(t)
	cleanScorecardFixtures(t, pool)
	t.Cleanup(func() { cleanScorecardFixtures(t, pool) })

	goalID := seedGoal(t, pool, "majority position")
	seedMatch(t, pool, goalID, "ZZ2", "1", "Ja")
	// Majority Ja despite one dissenter; absences must not count against it.
	seedBallots(t, pool, "ZZ2", "1", []string{"Ja", "Ja", "Nej", "Frånvarande", "Frånvarande"})

	sc := readScorecard(t, pool, goalID)

	if sc.scoredVotes != 1 {
		t.Errorf("scoredVotes = %d, want 1", sc.scoredVotes)
	}
	if sc.alignedVotes != 1 {
		t.Errorf("alignedVotes = %d, want 1", sc.alignedVotes)
	}
	if sc.alignmentPct == nil || *sc.alignmentPct != 100 {
		t.Errorf("alignmentPct = %v, want 100", sc.alignmentPct)
	}
}

func TestScorecard_MixedScoredAndUnclear(t *testing.T) {
	pool := connectScorecardTestDB(t)
	cleanScorecardFixtures(t, pool)
	t.Cleanup(func() { cleanScorecardFixtures(t, pool) })

	goalID := seedGoal(t, pool, "mixed")
	seedMatch(t, pool, goalID, "ZZ3", "1", "Ja") // party votes Ja  -> aligned
	seedMatch(t, pool, goalID, "ZZ3", "2", "Nej") // party votes Ja -> not aligned
	seedMatch(t, pool, goalID, "ZZ3", "3", "unclear") // excluded entirely
	seedBallots(t, pool, "ZZ3", "1", []string{"Ja", "Ja"})
	seedBallots(t, pool, "ZZ3", "2", []string{"Ja", "Ja"})
	seedBallots(t, pool, "ZZ3", "3", []string{"Ja", "Ja"})

	sc := readScorecard(t, pool, goalID)

	if sc.relevantVotes != 3 {
		t.Errorf("relevantVotes = %d, want 3", sc.relevantVotes)
	}
	if sc.scoredVotes != 2 {
		t.Errorf("scoredVotes = %d, want 2 (unclear excluded)", sc.scoredVotes)
	}
	if sc.alignedVotes != 1 {
		t.Errorf("alignedVotes = %d, want 1", sc.alignedVotes)
	}
	if sc.alignmentPct == nil || *sc.alignmentPct != 50 {
		t.Errorf("alignmentPct = %v, want 50", sc.alignmentPct)
	}
}

func TestScorecard_NoVotesRecordedIsNotScored(t *testing.T) {
	pool := connectScorecardTestDB(t)
	cleanScorecardFixtures(t, pool)
	t.Cleanup(func() { cleanScorecardFixtures(t, pool) })

	goalID := seedGoal(t, pool, "matched but no ballots")
	seedMatch(t, pool, goalID, "ZZ4", "1", "Ja") // no votes rows seeded

	sc := readScorecard(t, pool, goalID)

	if sc.scoredVotes != 0 {
		t.Errorf("scoredVotes = %d, want 0", sc.scoredVotes)
	}
	if sc.alignmentPct != nil {
		t.Errorf("alignmentPct = %v, want nil", *sc.alignmentPct)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

```bash
cd backend && TEST_DATABASE_URL='postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable' \
  go test ./internal/matching/adapters/postgres/ -run TestScorecard -v
```

Expected: FAIL — `column "scored_votes" does not exist`.

If the tests SKIP, postgres is not up. Start it first with `docker compose -f docker-compose.dev.yml up -d postgres` and re-run.

- [ ] **Step 3: Write the up migration**

Create `backend/migrations/000029_fix_goal_alignment_scoring.up.sql`:

```sql
-- The original party_scorecards view produced alignment_pct = 0 for every goal.
-- Three defects compounded:
--   1. Matches with aligned_direction = 'unclear' sat in the denominator, but
--      'unclear' can never equal a vote_result, so they scored 0 by construction.
--   2. votes holds one row per politician, so the join fanned out to individual
--      MP ballots rather than a single party position per vote point.
--   3. 'Frånvarande' ballots inflated the denominator; absence is not a position.
--
-- This view scores only vote points whose direction is determinable, and
-- collapses ballots to the party's majority position among cast votes.
-- alignment_pct is NULL — never 0 — when nothing is scorable.

DROP MATERIALIZED VIEW IF EXISTS party_scorecards;

CREATE MATERIALIZED VIEW party_scorecards AS
WITH party_positions AS (
  SELECT
    beteckning,
    forslagspunkt,
    party,
    mode() WITHIN GROUP (ORDER BY vote_result) AS position
  FROM votes
  WHERE vote_result <> 'Frånvarande'
  GROUP BY beteckning, forslagspunkt, party
)
SELECT
  pg.party,
  pg.id                                                       AS goal_id,
  pg.goal_text,
  pg.topic,
  COUNT(DISTINCT gvm.beteckning || ':' || gvm.forslagspunkt)  AS relevant_votes,
  COUNT(*) FILTER (
    WHERE gvm.aligned_direction <> 'unclear'
      AND pp.position IS NOT NULL
  )                                                           AS scored_votes,
  COUNT(*) FILTER (
    WHERE pp.position = gvm.aligned_direction
  )                                                           AS aligned_votes,
  ROUND(
    COUNT(*) FILTER (WHERE pp.position = gvm.aligned_direction)::numeric /
    NULLIF(
      COUNT(*) FILTER (
        WHERE gvm.aligned_direction <> 'unclear'
          AND pp.position IS NOT NULL
      ), 0
    ) * 100,
    1
  )                                                           AS alignment_pct
FROM party_goals pg
LEFT JOIN goal_vote_matches gvm ON gvm.goal_id = pg.id
LEFT JOIN party_positions pp
  ON  pp.beteckning    = gvm.beteckning
  AND pp.forslagspunkt = gvm.forslagspunkt
  AND pp.party         = pg.party
GROUP BY pg.party, pg.id, pg.goal_text, pg.topic;

-- UNIQUE index is required for REFRESH MATERIALIZED VIEW CONCURRENTLY,
-- which RefreshScorecards uses.
CREATE UNIQUE INDEX IF NOT EXISTS idx_party_scorecards_pk
  ON party_scorecards (party, goal_id);
CREATE INDEX IF NOT EXISTS idx_party_scorecards_party
  ON party_scorecards (party);
```

- [ ] **Step 4: Write the down migration**

Create `backend/migrations/000029_fix_goal_alignment_scoring.down.sql`:

```sql
-- Restore the original (buggy) view definition from 000001.

DROP MATERIALIZED VIEW IF EXISTS party_scorecards;

CREATE MATERIALIZED VIEW party_scorecards AS
SELECT
  pg.party,
  pg.id                                                         AS goal_id,
  pg.goal_text,
  pg.topic,
  COUNT(DISTINCT gvm.beteckning || ':' || gvm.forslagspunkt)   AS relevant_votes,
  ROUND(
    COUNT(CASE WHEN v.vote_result = gvm.aligned_direction THEN 1 END)::numeric /
    NULLIF(COUNT(v.id), 0) * 100,
    1
  )                                                             AS alignment_pct
FROM party_goals pg
LEFT JOIN goal_vote_matches gvm ON gvm.goal_id = pg.id
LEFT JOIN votes v
  ON  v.beteckning     = gvm.beteckning
  AND v.forslagspunkt  = gvm.forslagspunkt
  AND v.party          = pg.party
GROUP BY pg.party, pg.id, pg.goal_text, pg.topic;

CREATE UNIQUE INDEX IF NOT EXISTS idx_party_scorecards_pk
  ON party_scorecards (party, goal_id);
CREATE INDEX IF NOT EXISTS idx_party_scorecards_party
  ON party_scorecards (party);
```

- [ ] **Step 5: Apply the migration and re-run the tests**

```bash
make migrate
cd backend && TEST_DATABASE_URL='postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable' \
  go test ./internal/matching/adapters/postgres/ -run TestScorecard -v
```

Expected: all four tests PASS.

- [ ] **Step 6: Commit**

```bash
git add backend/migrations/000029_fix_goal_alignment_scoring.up.sql \
        backend/migrations/000029_fix_goal_alignment_scoring.down.sql \
        backend/internal/matching/adapters/postgres/scorecard_view_test.go
git commit -m "fix(scorecards): score only determinable vote points"
```

---

### Task 2: Carry the new columns through domain and repository

**Files:**
- Modify: `backend/internal/matching/domain/match.go:53-61`
- Modify: `backend/internal/matching/adapters/postgres/repository.go:133-167`

**Interfaces:**
- Consumes: the view columns from Task 1.
- Produces: `domain.ScorecardRow` with fields
  `Party string`, `GoalID int`, `GoalText string`, `Topic string`,
  `RelevantVotes int`, `ScoredVotes int`, `AlignedVotes int`, `AlignmentPct *float64`.
  Task 3 reads exactly these field names.

- [ ] **Step 1: Widen `ScorecardRow`**

Replace the struct at `backend/internal/matching/domain/match.go:53-61`:

```go
// ScorecardRow is a single row from the party_scorecards materialized view.
//
// RelevantVotes counts every vote point matched to the goal. ScoredVotes counts
// the subset whose aligned_direction is determinable and for which the party has
// a recorded position. AlignmentPct is nil when ScoredVotes is 0 — "not yet
// determinable" must never render as 0%.
type ScorecardRow struct {
	Party         string   `json:"party"`
	GoalID        int      `json:"goalId"`
	GoalText      string   `json:"goalText"`
	Topic         string   `json:"topic"`
	RelevantVotes int      `json:"relevantVotes"`
	ScoredVotes   int      `json:"scoredVotes"`
	AlignedVotes  int      `json:"alignedVotes"`
	AlignmentPct  *float64 `json:"alignmentPct"`
}
```

- [ ] **Step 2: Update both queries and the scanner**

In `backend/internal/matching/adapters/postgres/repository.go`, replace the two
query constants and the scan. Note `COALESCE(alignment_pct, 0)` is deliberately
gone — NULL must survive to the API.

```go
const scorecardColumns = `party, goal_id, goal_text, topic,
	relevant_votes, scored_votes, aligned_votes, alignment_pct`

func (r *Repository) GetPartyScorecard(ctx context.Context, party string) ([]*domain.ScorecardRow, error) {
	const q = `SELECT ` + scorecardColumns + `
		FROM party_scorecards WHERE party = $1 ORDER BY topic, goal_id`
	return r.queryScorecards(ctx, q, party)
}

func (r *Repository) GetAllPartyScorecards(ctx context.Context) ([]*domain.ScorecardRow, error) {
	const q = `SELECT ` + scorecardColumns + `
		FROM party_scorecards ORDER BY party, topic, goal_id`
	return r.queryScorecards(ctx, q)
}
```

And in `queryScorecards`, widen the `Scan`:

```go
		if err := rows.Scan(&row.Party, &row.GoalID, &row.GoalText, &row.Topic,
			&row.RelevantVotes, &row.ScoredVotes, &row.AlignedVotes,
			&row.AlignmentPct); err != nil {
			return nil, err
		}
```

- [ ] **Step 3: Build**

```bash
go build -C backend ./...
```

Expected: success. If `AlignmentPct` is used as a bare `float64` anywhere the
compiler will point at it — Task 3 fixes the handler; fix nothing else here.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/matching/domain/match.go \
        backend/internal/matching/adapters/postgres/repository.go
git commit -m "feat(scorecards): carry scoredVotes/alignedVotes through repository"
```

---

### Task 3: Expose the new fields on the goals API

**Files:**
- Modify: `backend/internal/goals/adapters/http/handler.go:29-95` (topic aggregate) and `:101-170` (`goalWithAlignment` + `listGoals`)
- Modify: `api/openapi.yaml`
- Regenerate: `frontend/src/shared/api-contract.ts`

**Interfaces:**
- Consumes: `domain.ScorecardRow` from Task 2.
- Produces: JSON fields `relevantVotes: int`, `scoredVotes: int`,
  `alignedVotes: int`, `alignmentPct: number | null` on each goal in
  `GET /api/parties/{party}/goals`. Task 4 reads exactly these.

- [ ] **Step 1: Widen the response struct**

In `backend/internal/goals/adapters/http/handler.go`, update `goalWithAlignment`
(currently at line 101). Keep every existing field; add three and change one:

```go
type goalWithAlignment struct {
	ID                 int      `json:"id"`
	Party              string   `json:"party"`
	GoalText           string   `json:"goalText"`
	Topic              string   `json:"topic"`
	Specificity        string   `json:"specificity"`
	SourceDocument     string   `json:"sourceDocument"`
	SourceURL          string   `json:"sourceUrl,omitempty"`
	SourceQuote        string   `json:"sourceQuote,omitempty"`
	Keywords           []string `json:"keywords,omitempty"`
	RelevantCommittees []string `json:"relevantCommittees,omitempty"`
	RelevantVotes      int      `json:"relevantVotes"`
	ScoredVotes        int      `json:"scoredVotes"`
	AlignedVotes       int      `json:"alignedVotes"`
	AlignmentPct       *float64 `json:"alignmentPct"`
}
```

- [ ] **Step 2: Update the score map in `listGoals`**

Replace the `scoreMap` block and the struct literal that fills `RelevantVotes` /
`AlignmentPct` (currently lines 136-167):

```go
	type goalScore struct {
		relevant int
		scored   int
		aligned  int
		pct      *float64
	}
	scoreMap := map[int]goalScore{}
	for _, row := range rows {
		scoreMap[row.GoalID] = goalScore{
			relevant: row.RelevantVotes,
			scored:   row.ScoredVotes,
			aligned:  row.AlignedVotes,
			pct:      row.AlignmentPct,
		}
	}
```

and in the append loop:

```go
			RelevantVotes:      sc.relevant,
			ScoredVotes:        sc.scored,
			AlignedVotes:       sc.aligned,
			AlignmentPct:       sc.pct,
```

- [ ] **Step 3: Fix the topic aggregate**

The topic aggregate at lines 29-95 averages `row.AlignmentPct` into
`AvgAlignment`. It must skip unscored goals rather than averaging nils as zero.
Change the accumulator at line 73 and the average at line 90:

```go
	// Only goals with a determinable direction contribute to the average.
	for _, row := range rows {
		if row.AlignmentPct == nil {
			continue
		}
		topicGoals[row.Party][row.Topic] = append(topicGoals[row.Party][row.Topic], *row.AlignmentPct)
	}
```

```go
			avg := 0.0
			if len(pcts) > 0 {
				sum := 0.0
				for _, p := range pcts {
					sum += p
				}
				avg = sum / float64(len(pcts))
			}
			// ... AlignmentPct: avg
```

Read the surrounding code before editing — the exact variable names in that
block must be preserved.

- [ ] **Step 4: Update the OpenAPI spec**

In `api/openapi.yaml`, find the goal schema that already carries `relevantVotes`
and `alignmentPct` (the response schema for `/parties/{party}/goals`). Add the
two new integers and make `alignmentPct` nullable:

```yaml
        relevantVotes:
          type: integer
          description: Vote points matched to this goal.
        scoredVotes:
          type: integer
          description: >-
            Subset of relevantVotes whose direction is determinable and for
            which the party has a recorded position. Only these are scored.
        alignedVotes:
          type: integer
          description: Subset of scoredVotes where the party's position matched the goal.
        alignmentPct:
          type: number
          nullable: true
          description: >-
            alignedVotes / scoredVotes as a percentage. Null when scoredVotes is
            0 — the direction could not be determined from the record.
```

- [ ] **Step 5: Regenerate the frontend contract**

```bash
cd frontend && npm run generate:api
```

Never hand-edit `src/shared/api-contract.ts`.

- [ ] **Step 6: Verify both gates**

```bash
go build -C backend ./...
cd frontend && npx tsc --noEmit
```

Expected: backend passes. `tsc` may now fail inside `PartyGoalsPage.tsx` because
`alignmentPct` became nullable — that is the expected signal, and Task 4 fixes it.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/goals/adapters/http/handler.go api/openapi.yaml frontend/src/shared/api-contract.ts
git commit -m "feat(goals): expose scoredVotes/alignedVotes, nullable alignmentPct"
```

---

### Task 4: Render "not yet determinable" instead of a false 0%

**Files:**
- Modify: `frontend/src/features/parties/PartyGoalsPage.tsx`
- Modify: `frontend/src/shared/types.ts` (if `GoalWithAlignment` is hand-written there rather than re-exported from the contract — check first)

**Interfaces:**
- Consumes: `relevantVotes`, `scoredVotes`, `alignedVotes`, `alignmentPct` from Task 3.
- Produces: no downstream consumers.

- [ ] **Step 1: Check where `GoalWithAlignment` comes from**

```bash
grep -rn "GoalWithAlignment" frontend/src/shared/types.ts frontend/src/features/parties/
```

If it is hand-written in `types.ts`, add `scoredVotes: number`,
`alignedVotes: number`, and change `alignmentPct` to `number | null`. If it is
derived from `api-contract.ts`, Task 3's regeneration already covered it.

- [ ] **Step 2: Update the per-goal card**

In `GoalCard`, replace the derived-count line and the vote summary. The current
code multiplies `pct` by `votes`, which is wrong now that they have different
denominators:

```tsx
  const scored = goal.scoredVotes ?? 0;
  const aligned = goal.alignedVotes ?? 0;
  const matched = goal.relevantVotes ?? 0;
```

and the summary span:

```tsx
              {scored > 0 ? (
                <span>{aligned} av {scored} relevanta röster i linje <SourceMarker sourceId="riksdagen" /></span>
              ) : matched > 0 ? (
                <span>{matched} omröstningar matchade, riktning ej fastställd <SourceMarker sourceId="riksdagen" /></span>
              ) : (
                <span>Ej prövad i någon omröstning ännu <SourceMarker sourceId="riksdagen" /></span>
              )}
```

- [ ] **Step 3: Suppress the ring when nothing is scored**

The `AlignmentRing` must not render a `0%` ring for an unscored goal. Where the
card renders the ring, guard it:

```tsx
            {scored > 0 ? (
              <AlignmentRing pct={goal.alignmentPct ?? 0} size={44} />
            ) : (
              <span className="text-xs text-on-surface-variant">—</span>
            )}
```

Read the existing ring usage in the card first and preserve its existing props
(size, colour) — only add the guard.

- [ ] **Step 4: Fix the header aggregate**

Replace the `avgAlignment` and `totalVotes` computation:

```tsx
  // Neutral aggregates only — raw facts, no good/bad bucketing.
  const scoredGoals = goals.filter((g) => (g.scoredVotes ?? 0) > 0);
  const totalScored = goals.reduce((s, g) => s + (g.scoredVotes ?? 0), 0);
  const avgAlignment = scoredGoals.length
    ? Math.round(
        scoredGoals.reduce((s, g) => s + (g.alignmentPct ?? 0), 0) / scoredGoals.length,
      )
    : null;
```

and in the header block, show `—` rather than `0%` when nothing is scored:

```tsx
            <div className="text-center">
              {avgAlignment !== null ? (
                <AlignmentRing pct={avgAlignment} size={64} color={pc?.bg} />
              ) : (
                <div className="text-2xl text-on-surface-variant">—</div>
              )}
              <div className="text-[10px] text-on-surface-variant mt-1 uppercase tracking-wider">Snitt i linje</div>
            </div>
            <StatBlock value={goals.length} label="Mål" small />
            <StatBlock value={totalScored} label="Bedömda röster" small />
```

- [ ] **Step 5: Verify**

```bash
cd frontend && npx tsc --noEmit && npm run lint
```

Expected: both pass.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/features/parties/PartyGoalsPage.tsx frontend/src/shared/types.ts
git commit -m "feat(goals): show 'riktning ej fastställd' instead of false 0%"
```

---

### Task 5: End-to-end verification against the running stack

**Files:** none (verification only)

- [ ] **Step 1: Bring the stack up**

```bash
make run
```

Wait for the backend to report migrations applied.

- [ ] **Step 2: Confirm the API returns null, not 0**

```bash
curl -s localhost:8080/api/parties/S/goals | \
  python3 -c "import json,sys; [print(g['goalText'][:40], g['relevantVotes'], g['scoredVotes'], g['alignedVotes'], g['alignmentPct']) for g in json.load(sys.stdin)]"
```

Expected: goals whose matches are all `unclear` show `scoredVotes = 0` and
`alignmentPct = None`. No goal shows `alignmentPct = 0` while `scoredVotes = 0`.

- [ ] **Step 3: Confirm the UI**

Open `http://localhost:5173/parties/S/goals`. Expected: unscored goals read
`"N omröstningar matchade, riktning ej fastställd"` with a `—` instead of a `0%`
ring. The header shows `—` for Snitt i linje when no goal is scored.

- [ ] **Step 4: Final gates**

```bash
go build -C backend ./...
cd frontend && npx tsc --noEmit
```

- [ ] **Step 5: Squash and push**

Squash every commit on the branch into one (project convention), then push and
open a PR against `main`.
