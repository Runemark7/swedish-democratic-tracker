# Goal alignment scoring — design

**Date:** 2026-07-29
**Status:** approved
**Area:** `matching` / `goals` — `party_scorecards` materialized view

## Problem

Every party goal on the live site reports `alignmentPct: 0`. All 80 goals,
all 8 parties. The promise-vs-vote feature — the core of the project — is
silently dead in production.

Worse than the zero: the UI renders it as `"0 av 5 relevanta röster i linje"`.
A reader parses that as *"this party voted against its own promise, five times
out of five."* We never determined how those five votes relate to the goal. The
page presents **unknown as broken promise**. That is a false claim stated as
fact, and it is exactly the bias-through-framing failure `CONTEXT.md` warns
about.

## Root cause

Three defects compound in the `party_scorecards` view
(`backend/migrations/000001_init_schema.up.sql:141-159`):

```sql
COUNT(CASE WHEN v.vote_result = gvm.aligned_direction THEN 1 END)::numeric /
NULLIF(COUNT(v.id), 0) * 100
```

### 1. `unclear` matches counted in the denominator

`goal_vote_matches.aligned_direction` allows `('Ja','Nej','Avstår','unclear')`.
`inferDirection` (`keyword_matcher.go:135`) returns `unclear` for any motion
proposed by a party other than the goal's own party or `"Regeringen"` — the
dominant real-world case. `unclear` can never equal a `vote_result`
(`Ja`/`Nej`/`Avstår`/`Frånvarande`), so those matches contribute 0 to the
numerator while still inflating the denominator.

Verified on live data: every sampled match returns
`"alignedDirection":"unclear"` (e.g. goal 47, party S, matched to an SD motion).

### 2. The join fans out to MP ballots, not party positions

`votes` holds **one row per politician**, not per party:

```
votes(politician_id, party, vote_result, beteckning, forslagspunkt, …)
```

So `LEFT JOIN votes v ON … AND v.party = pg.party` matches every MP of that
party for the vote point. `COUNT(v.id)` therefore counts individual ballots
(~100 per point), while the displayed `relevantVotes` counts *distinct vote
points* (4). The two units never reconcile, so the frontend's
`alignedVotes = pct/100 × votes` is wrong by construction.

### 3. Absent ballots sit in the denominator

`Frånvarande` rows join like any other and inflate `COUNT(v.id)`. Absence is
not a position; counting it drags every percentage down on its own.

## Design

Aggregate to a party position per vote point, and score only matches whose
direction can actually be determined from the record.

### View

```sql
WITH party_positions AS (
  -- Collapse MP ballots to one position per (vote point, party).
  -- Absence is not a position, so it is excluded before the vote is taken.
  SELECT beteckning, forslagspunkt, party,
         mode() WITHIN GROUP (ORDER BY vote_result) AS position
  FROM votes
  WHERE vote_result <> 'Frånvarande'
  GROUP BY beteckning, forslagspunkt, party
)
SELECT
  pg.party,
  pg.id AS goal_id,
  pg.goal_text,
  pg.topic,
  COUNT(DISTINCT gvm.beteckning || ':' || gvm.forslagspunkt) AS relevant_votes,
  COUNT(*) FILTER (
    WHERE gvm.aligned_direction <> 'unclear' AND pp.position IS NOT NULL
  ) AS scored_votes,
  COUNT(*) FILTER (WHERE pp.position = gvm.aligned_direction) AS aligned_votes,
  ROUND(
    COUNT(*) FILTER (WHERE pp.position = gvm.aligned_direction)::numeric /
    NULLIF(COUNT(*) FILTER (
      WHERE gvm.aligned_direction <> 'unclear' AND pp.position IS NOT NULL
    ), 0) * 100,
    1
  ) AS alignment_pct
FROM party_goals pg
LEFT JOIN goal_vote_matches gvm ON gvm.goal_id = pg.id
LEFT JOIN party_positions pp
  ON  pp.beteckning    = gvm.beteckning
  AND pp.forslagspunkt = gvm.forslagspunkt
  AND pp.party         = pg.party
GROUP BY pg.party, pg.id, pg.goal_text, pg.topic;
```

`alignment_pct` is **NULL** when `scored_votes = 0` — the honest value for "we
cannot determine this". It must not be coalesced to 0 anywhere in the stack.

**Party position rule:** majority of cast ballots. Riksdag party-line
discipline is high, so this is near-unanimous in practice. `mode()` breaks an
exact tie by sort order; a perfectly split party delegation is rare enough to
accept the arbitrary pick rather than add a tie-breaking rule.

### Expose raw counts, not just the percentage

The frontend currently reconstructs the count as `pct × votes`, which is lossy
and — given defect 2 — wrong. The counts are the fact; the percentage is
derived. Send both.

| Field | Meaning |
|---|---|
| `relevantVotes` | vote points matched to this goal (unchanged) |
| `scoredVotes` | of those, how many have a determinable direction **(new)** |
| `alignedVotes` | of those, how many the party's position matched **(new)** |
| `alignmentPct` | `alignedVotes / scoredVotes`; **null** when `scoredVotes = 0` |

### Frontend

- `scoredVotes === 0` → `"Ej prövad i någon omröstning ännu"`. When
  `relevantVotes > 0`, add `"{relevantVotes} omröstningar matchade, riktning ej
  fastställd"` so the reader can see matching ran and still returned no
  determinable direction.
- otherwise → `"{alignedVotes} av {scoredVotes} relevanta röster i linje"`
- header average skips goals with `scoredVotes === 0` rather than averaging
  zeros into the figure. If no goal is scored, show `—`, not `0%`.

## Explicitly out of scope

`inferDirection` is **not** touched. Widening it (bloc heuristics, inferring a
party's intended position from who proposed a bill) would make the site guess a
verdict from party identity rather than proposal content — the interpretation
layer `CONTEXT.md` reserves for the reader. Coverage stays honestly low until
direction can be established from the record. That is the intended outcome, not
a shortfall: the fix converts a false "0 av 5" into a truthful "not yet
determinable".

## Migration

New migration `000029`. A materialized view definition cannot be altered, so:

1. `DROP MATERIALIZED VIEW IF EXISTS party_scorecards`
2. recreate with the definition above (populated on create)
3. recreate `idx_party_scorecards_pk` UNIQUE (party, goal_id) — required for
   `REFRESH … CONCURRENTLY`, which `RefreshScorecards` uses
4. recreate `idx_party_scorecards_party`

Idempotent throughout (`IF EXISTS` / `IF NOT EXISTS`), per the lesson from
`000024` (#75). The `.down.sql` restores the original view verbatim.

## Files

- `backend/migrations/000029_fix_goal_alignment_scoring.{up,down}.sql` — new
- `backend/internal/matching/domain/match.go` — `ScoredVotes`, `AlignedVotes`;
  `AlignmentPct` becomes nullable
- `backend/internal/matching/adapters/postgres/repository.go` — select the new
  columns; drop `COALESCE(alignment_pct, 0)`
- `backend/internal/goals/adapters/http/handler.go` — expose new fields
- `api/openapi.yaml` + regenerated `frontend/src/shared/api-contract.ts`
- `frontend/src/features/parties/PartyGoalsPage.tsx` — display rules above

## Verification

- `go build -C backend ./...`
- `cd frontend && npx tsc --noEmit`
- Local stack via `make run`; confirm a goal with only `unclear` matches renders
  "Ej prövad", and that no goal shows a `0%` ring while `scoredVotes = 0`
- Post-deploy: re-run the live API probe across all 8 parties and confirm
  `alignmentPct` is `null` (not `0`) wherever `scoredVotes = 0`
