# Utskott Spine — Phase 2: Committee Pages and Record Surfaces Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the committee page with its three rows (LOVAT / RÖSTAT / KOSTAR), and collapse the three party-goal surfaces into one party record page grouped by committee, so a goal is rendered by exactly one component.

**Architecture:** Two new backend queries — party goals by committee (in the `goals` feature, which owns `party_goals`) and voteringar by committee with per-party positions (in the `votes` feature). Frontend gains `CommitteePage` and a shared `GoalCard`; `PartyGoalsPage` merges into `PartyDetailPage` and `/parties/:party/goals` redirects. No navigation or front-page changes — those are Phase 3.

**Tech Stack:** Go 1.26.1, chi v5, pgx/v5, PostgreSQL 17, React 19, Vite 6, TanStack Query v5, Tailwind v4.

**Depends on:** Phase 1 (`docs/superpowers/plans/2026-08-05-utskott-spine-phase1-data.md`) — `GET /api/committees` and `GET /api/committees/{code}` must exist, and `domain.CommitteeCode` must be the single place a code is parsed.

> **Amended 2026-08-11, before execution.** A pre-flight check against the dev database
> found three defects in this plan's own code. Task 2 keyed a votering on
> beteckning+förslagspunkt without `session`, which reproduces the 27% undercount from
> Phase 1 — 76 for AU against a true 116 — and would have merged two riksmöten's party
> positions into one row; its test asserted a floor of 76, so it would have passed on the
> wrong number. Task 1's SQL omitted `recordViewFilter`, which every other goals query
> applies. Task 4's redirect resolved to `/parties` rather than the party.
>
> A second pass before Task 2 found three more. The query matched with `LIKE`, so a code of
> `%` arriving from the URL path would have served all 2 576 voteringar as a single
> committee's record. `total` came from a second COUNT query carrying its own copy of the
> predicate, free to drift from the rows beneath it. And Task 4's verification expected MP
> to hold goals in 3 committees when it holds 5. All six are corrected below.

## Global Constraints

- **Go 1.26.1, Node 25.9.0.** Parameterised SQL only. Never import an adapter into domain. `cmd/api/main.go` is the only wiring point.
- **`api/openapi.yaml` first**, then `cd frontend && npm run generate:api`. Never hand-edit `api-contract.ts`.
- **`frontend/src/shared/types.ts` hand-mirrors the generated contract.** New API fields must be added there too or `tsc` fails. This duplication is known fog on map #82 — do not fix it here, but do not forget it either.
- **No derived comparison across parties or committees.** No shares, no ranks, no ordering that implies precedence. Alphabetical unless the data has an intrinsic order.
- **Verification gate:** `go build -C backend ./...`, `cd frontend && npx tsc --noEmit`, `npm run lint`, `npm run build` all clean. Lint is enforced in CI.
- **There is no frontend component test runner.** Only `test:scripts` (`node --test scripts/__tests__/*.test.mjs`) exists. Frontend tasks are verified by `tsc` + `lint` + `build` + a scripted browser check against a real API, never by a test step that cannot run.
- **Copy is Swedish; code, comments, commits and docs are English.**

## Decisions this plan implements

| Source | Decision |
|---|---|
| [#100](https://github.com/Runemark7/swedish-democratic-tracker/issues/100) | Spine and record are **transposes**, neither canonical. **One subject taxonomy: committees**; `topic` survives only as a tag. **No vote count beside a goal.** Three party-goal surfaces collapse to **two**, behind **one shared renderer**. **All committees shown per party**, empties captioned. |
| [#86](https://github.com/Runemark7/swedish-democratic-tracker/issues/86) | Show the empty row and **attribute the absence to our data, never to the parties**. |
| [#90](https://github.com/Runemark7/swedish-democratic-tracker/issues/90) | KOSTAR is **amounts per utgiftsområde, never a share and never a ranking**, plus a plain-words note where a committee's remit exceeds its areas. |
| [#102](https://github.com/Runemark7/swedish-democratic-tracker/issues/102) | Anföranden appear on a **decision's page only**, never on a committee page. |
| [#103](https://github.com/Runemark7/swedish-democratic-tracker/issues/103) | **No regional data on a committee page.** There is no fourth row. |
| [#97](https://github.com/Runemark7/swedish-democratic-tracker/issues/97) | Committee display names never gate what exists. |

## What is already done — do not rebuild

Verified in the codebase on 2026-08-06:

- **Anföranden on decision pages already work.** `BeslutDetailPage.tsx:163` calls `useSpeechesByDocument(dokId)` and renders "Debatten · N talare" at `:498`.
- **Its empty state already follows #86.** `debattFallbackLabel` (`:51-62`) returns *"Debatten hölls X · inga anföranden registrerade här"* — the gap attributed to our register — and distinguishes a planned debate from a held one. **Do not touch it.**
- **`relevantCommittees` is already exposed** on the goals API (`goals/adapters/http/handler.go:83`).
- **The alignment percentage is gone**, and `PartiesPage` is already alphabetical.

So #102's anföranden half needs **no work in this phase**. What remains is not adding them to committee pages.

## File Structure

| File | Responsibility |
|---|---|
| `backend/internal/goals/ports/repository.go` | +`ListByCommittee` |
| `backend/internal/goals/adapters/postgres/repository.go` | SQL for goals by committee |
| `backend/internal/goals/adapters/postgres/committee_test.go` | DB-backed test |
| `backend/internal/goals/adapters/http/handler.go` | `GET /committees/{code}/goals` |
| `backend/internal/votes/ports/repository.go` | +`ListByCommitteeWithPositions`, +`CommitteeVotering` |
| `backend/internal/votes/adapters/postgres/committee_repo.go` | SQL, **no enrichment gate**, numeric ordering |
| `backend/internal/votes/adapters/postgres/committee_repo_test.go` | DB-backed test |
| `backend/internal/votes/adapters/http/handler.go` | `GET /committees/{code}/votes` |
| `frontend/src/shared/components/GoalCard.tsx` | **The one** goal renderer |
| `frontend/src/features/committees/CommitteePage.tsx` | The three rows |
| `frontend/src/features/committees/api.ts` | Fetchers |
| `frontend/src/features/parties/PartyDetailPage.tsx` | Absorbs the record view |
| `frontend/src/App.tsx` | `/committees/:code` route; `/parties/:party/goals` → redirect |
| `api/openapi.yaml`, `frontend/src/shared/types.ts` | Contract |

Deleted at the end of Task 4: `frontend/src/features/parties/PartyGoalsPage.tsx`.

---

### Task 1: Party goals by committee (LOVAT data)

`party_goals.relevant_committees` is a `text[]` owned by the `goals` feature, so the query lives there rather than in `committees` reaching across features.

**Files:**
- Modify: `backend/internal/goals/ports/repository.go`
- Modify: `backend/internal/goals/adapters/postgres/repository.go`
- Create: `backend/internal/goals/adapters/postgres/committee_test.go`
- Modify: `backend/internal/goals/adapters/http/handler.go`
- Modify: `api/openapi.yaml`

**Interfaces:**
- Produces: `ListByCommittee(ctx context.Context, code string) ([]*domain.Goal, error)`; `GET /api/committees/{code}/goals` → `Goal[]`.

- [ ] **Step 1: Write the failing test**

Create `backend/internal/goals/adapters/postgres/committee_test.go`:

```go
package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/goals/adapters/postgres"
)

func connectGoalsCommitteeDB(t *testing.T) *pgxpool.Pool {
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

// MJU is the most-referenced committee in the seed data: it appears in the
// goals of every party. Goals must come back ordered by party alphabetically,
// since ordering them any other way would rank the parties.
func TestListByCommittee(t *testing.T) {
	pool := connectGoalsCommitteeDB(t)
	repo := postgres.NewRepository(pool)

	got, err := repo.ListByCommittee(context.Background(), "MJU")
	if err != nil {
		t.Fatalf("ListByCommittee: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("no goals for MJU — the array containment query is wrong")
	}
	for i, g := range got {
		var found bool
		for _, c := range g.RelevantCommittees {
			if c == "MJU" {
				found = true
			}
		}
		if !found {
			t.Errorf("goal %d returned for MJU but lists %v", g.ID, g.RelevantCommittees)
		}
		if i > 0 && got[i-1].Party > g.Party {
			t.Errorf("not alphabetical by party: %q before %q", got[i-1].Party, g.Party)
		}
	}

	// KrU is the one committee with no seeded goals. It must return an empty
	// slice and no error: the caller renders the row and attributes the gap to
	// our seed data, so "no goals" must be a normal answer, not a failure.
	empty, err := repo.ListByCommittee(context.Background(), "KrU")
	if err != nil {
		t.Fatalf("ListByCommittee(KrU): %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("KrU returned %d goals, expected 0 from seed data", len(empty))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && TEST_DATABASE_URL="postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable" go test ./internal/goals/adapters/postgres/ -run TestListByCommittee -v`
Expected: FAIL with `repo.ListByCommittee undefined`.

- [ ] **Step 3: Add the port method**

In `backend/internal/goals/ports/repository.go`, add to the `GoalRepository` interface:

```go
	// ListByCommittee returns the goals whose relevant_committees contains the
	// code, alphabetically by party. An empty result is a normal answer: one
	// committee has no seeded goals, and the caller states that as a gap in our
	// data rather than as the parties having nothing to say.
	ListByCommittee(ctx context.Context, code string) ([]*domain.Goal, error)
```

- [ ] **Step 4: Implement the query**

In `backend/internal/goals/adapters/postgres/repository.go`, add:

```go
// ListByCommittee returns goals whose relevant_committees array contains code.
//
// Uses array containment (@>) rather than unnest+join: relevant_committees is a
// text[] and containment is exact, so a code is never matched as a substring of
// another — "NU" must not match "UFöU".
// Applies recordViewFilter like every other query in this file. Without it the
// committee page would surface goals that the party pages hide, and the two
// surfaces would disagree about what the record contains.
func (r *Repository) ListByCommittee(ctx context.Context, code string) ([]*domain.Goal, error) {
	q := "SELECT " + selectCols + " FROM party_goals" +
		" WHERE relevant_committees @> ARRAY[$1]::text[] AND " + recordViewFilter +
		" ORDER BY party, id"
	rows, err := r.db.Query(ctx, q, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGoals(rows)
}
```

Check the existing file for the scan helper's real name — if it is not `scanGoals`, use whatever `ListByParty` uses, and match its error handling exactly.

- [ ] **Step 5: Run test to verify it passes**

Run: `cd backend && TEST_DATABASE_URL="postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable" go test ./internal/goals/adapters/postgres/ -run TestListByCommittee -v`
Expected: PASS, with MJU returning goals from several parties and KrU returning zero.

- [ ] **Step 6: Expose the endpoint**

Add to `api/openapi.yaml` under `paths:`:

```yaml
  /committees/{code}/goals:
    get:
      summary: Party goals whose relevant_committees contains this committee
      description: >-
        The LOVAT row. Goals are the parties' own stated aims, alphabetical by
        party. An empty result means our seed data holds no goals for this area,
        which the caller must state as our gap and never as the parties' silence.
      parameters:
        - name: code
          in: path
          required: true
          schema: { type: string }
      responses:
        "200":
          description: Goals, alphabetical by party
          content:
            application/json:
              schema:
                type: array
                items: { $ref: "#/components/schemas/PartyGoal" }
```

Reuse the existing goal schema — find its name in the spec (the goals handler DTO at `goals/adapters/http/handler.go:83` shows the field set) and reference that rather than defining a second shape.

In `backend/internal/goals/adapters/http/handler.go`, register `r.Get("/committees/{code}/goals", h.listByCommittee)` and implement it by calling the service, mapping to the existing DTO used by the other goal endpoints.

- [ ] **Step 7: Verify and commit**

```bash
go build -C backend ./...
cd frontend && npm run generate:api && npx tsc --noEmit && npm run lint
```

```bash
git add backend/internal/goals/ api/openapi.yaml frontend/src/shared/api-contract.ts
git commit -m "feat: party goals by committee for the LOVAT row

Array containment on relevant_committees, alphabetical by party. Lives in
the goals feature, which owns party_goals, rather than a committees
repository reaching across features. An empty result is a normal answer:
KrU has no seeded goals and the caller states that as our gap."
```

---

### Task 2: Voteringar by committee with party positions (RÖSTAT data)

**Files:**
- Modify: `backend/internal/votes/ports/repository.go`
- Create: `backend/internal/votes/adapters/postgres/committee_repo.go`
- Create: `backend/internal/votes/adapters/postgres/committee_repo_test.go`
- Modify: `backend/internal/votes/adapters/http/handler.go`, `api/openapi.yaml`

**Interfaces:**
- Produces:
  ```go
  type CommitteeVotering struct {
      Beteckning     string
      Forslagspunkt  string
      DocumentTitle  string
      Riksmote       string
      PartyPositions []PartyPosition // dominant position per party
  }
  type PartyPosition struct {
      Party    string
      Position string // "Ja" | "Nej" | "Avstår" | "Frånvarande"
  }
  ```
  `ListByCommitteeWithPositions(ctx, code string, limit, offset int) ([]CommitteeVotering, int, error)` — returns page plus total.

**Five traps this task must avoid.** The existing `ListDistinctByCommitteePrefix` (`votes/adapters/postgres/repository.go:275`) is **not** reusable as-is:

1. It filters `origin_enriched = true`, which would silently drop voteringar whose origin has not been enriched — RÖSTAT is a record of how parties voted and must not depend on enrichment.
2. It orders by `beteckning` as text, so `AU10` sorts before `AU9`.
3. **The prefix predicate is `starts_with(upper(...), upper($1))`, never `LIKE`.** The code arrives from a URL path; under `LIKE` a code of `%` matches every beteckning in the table, serving all 2 576 voteringar as one committee's record. The committees feature already ruled on this (`committees/adapters/postgres/repository.go:110`, *"starts_with, not LIKE: the code comes from a URL path and must not be able to smuggle in wildcards"*). Verified against the dev database on 2026-08-11.
4. **The total is the folded length, not a second COUNT query.** The function already materialises every row before paging, so a separate count adds a second copy of the predicate that can drift from the rows beneath it.
5. **A votering is identified by `votering_id`, never by `beteckning` + `forslagspunkt`.** Beteckning numbering restarts each riksmöte, so keying without `session` collapses `AU5` in 2022/23 into `AU5` in 2024/25 — which undercounts AU by 34% (76 against a true 116) *and* merges two different decisions' party positions into one row. This is the defect that shipped through five reviews in Phase 1. Verified against the dev database on 2026-08-11: `votering_id` is never null and never spans more than one (beteckning, förslagspunkt, session), so it is the whole key. Use it in all four places — the count query, the window partition, the join, and the Go fold.

Also note: **`votes` carries no decision date** — only `system_datum`, which is when Riksdagen last touched the record and must never be shown as a decision date. So RÖSTAT cannot be ordered chronologically. Order by riksmöte, then the numeric part of the beteckning, then the numeric förslagspunkt.

- [ ] **Step 1: Write the failing test**

Create `backend/internal/votes/adapters/postgres/committee_repo_test.go`:

```go
package postgres_test

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/votes/adapters/postgres"
)

func connectVotesCommitteeDB(t *testing.T) *pgxpool.Pool {
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

func TestListByCommitteeWithPositions(t *testing.T) {
	pool := connectVotesCommitteeDB(t)
	repo := postgres.NewRepository(pool)

	got, total, err := repo.ListByCommitteeWithPositions(context.Background(), "AU", 20, 0)
	if err != nil {
		t.Fatalf("ListByCommitteeWithPositions: %v", err)
	}
	// AU holds 116 voteringar in 2022-2026, counted as count(DISTINCT votering_id)
	// against the dev database. Two separate bugs each shrink this number, so an
	// exact assertion is used rather than a floor — a floor would pass on both:
	//   - gating on origin_enriched drops unenriched voteringar
	//   - keying on beteckning+forslagspunkt without session collapses riksmöten
	//     into each other and yields 76
	if total != 116 {
		t.Errorf("total = %d, want 116. 76 means the query keys on beteckning+forslagspunkt "+
			"without session and has collapsed riksmöten; a value below 76 means it gates "+
			"on origin_enriched", total)
	}
	if len(got) == 0 {
		t.Fatal("no voteringar returned for AU")
	}
	for _, v := range got {
		if len(v.PartyPositions) == 0 {
			t.Errorf("%s:%s has no party positions", v.Beteckning, v.Forslagspunkt)
		}
		for _, p := range v.PartyPositions {
			switch p.Position {
			case "Ja", "Nej", "Avstår", "Frånvarande":
			default:
				t.Errorf("unexpected position %q for %s", p.Position, p.Party)
			}
		}
	}
}

// AU10 must not sort before AU9. Ordering by beteckning as text does exactly
// that, which puts the record in an order no reader expects.
func TestListByCommittee_NumericOrdering(t *testing.T) {
	pool := connectVotesCommitteeDB(t)
	repo := postgres.NewRepository(pool)

	got, _, err := repo.ListByCommitteeWithPositions(context.Background(), "AU", 200, 0)
	if err != nil {
		t.Fatalf("ListByCommitteeWithPositions: %v", err)
	}
	prev := -1
	prevRm := ""
	for _, v := range got {
		n, err := strconv.Atoi(numericSuffix(v.Beteckning))
		if err != nil {
			continue
		}
		if v.Riksmote == prevRm && n < prev {
			t.Errorf("beteckning out of numeric order in %s: %d after %d", v.Riksmote, n, prev)
		}
		prev, prevRm = n, v.Riksmote
	}
}

// A code arrives from a URL path. Under a LIKE predicate "%" matches every
// beteckning in the table, so /committees/%/votes would serve the entire record
// as one committee's. starts_with treats it as a literal and matches nothing.
func TestListByCommittee_WildcardIsNotSmuggled(t *testing.T) {
	pool := connectVotesCommitteeDB(t)
	repo := postgres.NewRepository(pool)

	for _, code := range []string{"%", "_", "%%"} {
		got, total, err := repo.ListByCommitteeWithPositions(context.Background(), code, 20, 0)
		if err != nil {
			t.Fatalf("ListByCommitteeWithPositions(%q): %v", code, err)
		}
		if total != 0 || len(got) != 0 {
			t.Errorf("code %q returned total=%d rows=%d, want 0 — the predicate is "+
				"treating it as a wildcard instead of a literal", code, total, len(got))
		}
	}
}

func numericSuffix(s string) string {
	out := ""
	for _, r := range s {
		if r >= '0' && r <= '9' {
			out += string(r)
		} else if out != "" {
			break
		}
	}
	return out
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && TEST_DATABASE_URL="postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable" go test ./internal/votes/adapters/postgres/ -run TestListByCommittee -v`
Expected: FAIL with `repo.ListByCommitteeWithPositions undefined`.

- [ ] **Step 3: Add the port types and method**

In `backend/internal/votes/ports/repository.go`:

```go
// PartyPosition is how a party voted on one förslagspunkt: the position most of
// its members took. Stated as a fact, never scored against a promise.
type PartyPosition struct {
	Party    string `json:"party"`
	Position string `json:"position"`
}

// CommitteeVotering is one decided förslagspunkt with each party's position.
//
// Carries no date: votes hold only system_datum, which is when Riksdagen last
// touched the record, so presenting it as a decision date would repeat a bug
// the site already removed.
type CommitteeVotering struct {
	Beteckning     string          `json:"beteckning"`
	Forslagspunkt  string          `json:"forslagspunkt"`
	DocumentTitle  string          `json:"documentTitle"`
	Riksmote       string          `json:"riksmote"`
	PartyPositions []PartyPosition `json:"partyPositions"`
}
```

and on the repository interface:

```go
	// ListByCommitteeWithPositions returns the committee's voteringar with each
	// party's dominant position, plus the total count for pagination.
	//
	// Must not filter on origin_enriched: enrichment is about who proposed a
	// bill, and gating on it would remove real voteringar from the record.
	ListByCommitteeWithPositions(ctx context.Context, code string, limit, offset int) ([]CommitteeVotering, int, error)
}
```

- [ ] **Step 4: Implement the query**

Create `backend/internal/votes/adapters/postgres/committee_repo.go`:

```go
package postgres

import (
	"context"

	"riksdagskollen/internal/votes/ports"
)

// The unit of this query is votering_id, which is the whole key: beteckning
// numbering restarts each riksmöte, so keying on beteckning+forslagspunkt
// silently folds AU5/2022-23 into AU5/2024-25 — both undercounting the record
// and blending two decisions' party positions into one row.
//
// Ordering: votes carry no decision date (only system_datum, which is when
// Riksdagen last touched the row), so chronological order is unavailable and
// must not be faked. Order by riksmöte, then the numeric part of the beteckning
// — regexp_replace strips non-digits so "AU10" → 10 sorts after "AU9" → 9,
// which ordering the raw text would get backwards. NULLIF guards the cast: a
// beteckning with no digits would otherwise make ''::int raise.
const committeeVoteringQuery = `
WITH pos AS (
	SELECT votering_id, party, vote_result,
	       row_number() OVER (
	           PARTITION BY votering_id, party
	           ORDER BY count(*) DESC, vote_result
	       ) AS rn
	FROM votes
	WHERE starts_with(upper(beteckning), upper($1))
	GROUP BY votering_id, party, vote_result
),
pts AS (
	SELECT votering_id,
	       min(beteckning)                  AS beteckning,
	       min(forslagspunkt)               AS forslagspunkt,
	       min(session)                     AS session,
	       COALESCE(max(document_title), '') AS document_title
	FROM votes
	WHERE starts_with(upper(beteckning), upper($1))
	GROUP BY votering_id
)
SELECT p.votering_id, p.beteckning, p.forslagspunkt, p.document_title, p.session,
       pos.party, pos.vote_result
FROM pts p
JOIN pos ON pos.votering_id = p.votering_id AND pos.rn = 1
ORDER BY p.session,
         NULLIF(regexp_replace(p.beteckning, '\D', '', 'g'), '')::int,
         NULLIF(regexp_replace(p.forslagspunkt, '\D', '', 'g'), '')::int,
         pos.party`

// ListByCommitteeWithPositions returns the committee's voteringar, each with the
// dominant position per party.
//
// The prefix match is starts_with, never LIKE. The code arrives from a URL path,
// and under LIKE a code of "%" matches every beteckning in the table — 2 576
// voteringar served as though they were one committee's record. The committees
// feature already settled this at
// committees/adapters/postgres/repository.go:110; this query follows it.
//
// No committee code is a prefix of another (AU CU FiU FöU JuU KrU KU MJU NU SfU
// SkU SoU TU UbU UFöU UU), so the prefix needs no further correction. Callers
// still pass a canonical code from domain.Canonical, never a fragment.
func (r *Repository) ListByCommitteeWithPositions(ctx context.Context, code string, limit, offset int) ([]ports.CommitteeVotering, int, error) {
	rows, err := r.db.Query(ctx, committeeVoteringQuery, code)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Rows arrive one per (votering, party); fold them into voteringar. The fold
	// key is votering_id for the same reason the SQL groups on it — folding on
	// beteckning+forslagspunkt would merge riksmöten back together here even if
	// the query kept them apart.
	var out []ports.CommitteeVotering
	idx := map[string]int{}
	for rows.Next() {
		var vid, bet, punkt, title, rm, party, result string
		if err := rows.Scan(&vid, &bet, &punkt, &title, &rm, &party, &result); err != nil {
			return nil, 0, err
		}
		key := vid
		i, ok := idx[key]
		if !ok {
			out = append(out, ports.CommitteeVotering{
				Beteckning:    bet,
				Forslagspunkt: punkt,
				DocumentTitle: title,
				Riksmote:      rm,
			})
			i = len(out) - 1
			idx[key] = i
		}
		out[i].PartyPositions = append(out[i].PartyPositions,
			ports.PartyPosition{Party: party, Position: result})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// The total is the folded length, not a second COUNT query. Two queries with
	// two copies of the predicate can drift apart, and a total that disagrees
	// with the rows beneath it is exactly the kind of authoritative-looking wrong
	// number this site exists to avoid.
	total := len(out)

	// Page after folding, so a page is always whole voteringar.
	if offset >= len(out) {
		return []ports.CommitteeVotering{}, total, nil
	}
	end := offset + limit
	if end > len(out) {
		end = len(out)
	}
	return out[offset:end], total, nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend && TEST_DATABASE_URL="postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable" go test ./internal/votes/adapters/postgres/ -run TestListByCommittee -v`
Expected: PASS. `total` for AU must be exactly **116**. **76** means the query keyed on beteckning+forslagspunkt and collapsed the riksmöten; anything below 76 means an enrichment filter has crept in.

- [ ] **Step 6: Expose the endpoint, verify, commit**

Add `GET /api/committees/{code}/votes` with `limit`/`offset` query params and a `CommitteeVoteringPage` schema (`items`, `total`) to `api/openapi.yaml`, register it on the votes handler, regenerate types, then:

```bash
go build -C backend ./... && cd frontend && npm run generate:api && npx tsc --noEmit && npm run lint
```

```bash
git add backend/internal/votes/ api/openapi.yaml frontend/src/shared/api-contract.ts
git commit -m "feat: committee voteringar with party positions for RÖSTAT

Keys on votering_id throughout — count, window partition, join and fold.
Beteckning numbering restarts each riksmöte, so keying on
beteckning+forslagspunkt collapses AU5/2022-23 into AU5/2024-25: it
undercounts AU by 34% (76 against 116) and merges two decisions' party
positions into one row.

Deliberately does not reuse ListDistinctByCommitteePrefix: that filters
origin_enriched = true, which would drop real voteringar from a record of
how parties voted, and orders beteckning as text so AU10 precedes AU9.
Ordering is riksmöte, then numeric beteckning, then numeric förslagspunkt —
votes carry no decision date, only system_datum, so chronological order is
not available and must not be faked."
```

---

### Task 3: The committee page

**Files:**
- Create: `frontend/src/features/committees/api.ts`
- Create: `frontend/src/features/committees/CommitteePage.tsx`
- Modify: `frontend/src/App.tsx` (add `/committees/:code`)
- Modify: `frontend/src/shared/types.ts`

**Interfaces:**
- Consumes: `/api/committees/{code}`, `/api/committees/{code}/goals`, `/api/committees/{code}/votes`.
- Produces: route `/committees/:code`.

- [ ] **Step 1: Add the mirrored types**

In `frontend/src/shared/types.ts` add `Committee`, `CommitteeExpenditureArea`, `CommitteeVotering`, `PartyPosition` matching the generated contract exactly. `tsc` will fail if a field name or nullability differs — that failure is the check.

- [ ] **Step 2: Write the fetchers**

Create `frontend/src/features/committees/api.ts` following the pattern in `frontend/src/features/parties/api.ts` — same `fetch` wrapper, same error handling. Three functions: `getCommittee(code, period)`, `listGoals(code)`, `listVotes(code, page)`.

- [ ] **Step 3: Build the page with three rows**

Create `frontend/src/features/committees/CommitteePage.tsx`. Requirements, each traceable to a decision:

- **Heading**: committee name from the API (which falls back to the code — never render "undefined").
- **LOVAT** — heading reads *"Partimål inom {name}s område"*, never "Löften om X". `party_goals.relevant_committees` is committee-level, so a narrower claim would overclaim. Renders the shared `GoalCard` from Task 4; if Task 4 has not landed yet, do Task 4 first.
- **RÖSTAT** — the committee's voteringar with each party's position, stated as fact. No direction, no alignment, no "kept/broke". The promise sits beside it and the reader connects them.
- **KOSTAR** — heading *"Utgiftsområden som utskottet bereder"*. Amounts per UO with `formatBudgetAmount` from `@/shared/design`. **No share of the total. No ranking. No implication that one area funds another.**
- **Every empty row renders**, with the absence attributed to our data:

```tsx
{goals.length === 0 && (
  <p className="text-sm text-on-surface-variant">
    Vi har inga inlästa partimål inom detta område.{" "}
    <SourceMarker sourceId="seed-party-goals" />
  </p>
)}
```

Never *"Inga partimål inom detta område"*, which reads as no party caring about the subject. Today only KrU is actually empty.

- **A plain-words note where the remit exceeds the areas.** Hardcode nothing clever: render the note when the committee is `SkU`, whose areas total ~1 % of expenditure while it handles all taxation because revenue is not an utgiftsområde, and when it is `FiU`, whose four areas understate a committee that rams all 27:

```tsx
const REMIT_NOTE: Record<string, string> = {
  SkU: "Skatteutskottet bereder statens inkomster, som inte är ett utgiftsområde. Beloppen nedan visar därför bara en del av utskottets område.",
  FiU: "Finansutskottet bereder budgeten som helhet, inte bara utgiftsområdena nedan.",
};
```

- **No I DIN REGION row and no anföranden row.** #103 and #102 settled both; adding either is a regression.

- [ ] **Step 4: Add the route**

In `frontend/src/App.tsx`, add beside the other routes:

```tsx
<Route path="/committees/:code" element={<CommitteePage />} />
```

Do not touch `navPills` or the sub-tabs — navigation is Phase 3.

- [ ] **Step 5: Verify against a real API**

```bash
cd backend && DATABASE_URL="postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable" PORT=8099 INITIAL_SYNC=false go run ./cmd/api &
cd frontend && BACKEND_URL=http://localhost:8099 npx vite --port 5199 &
```

Then in a browser check `http://localhost:5199/committees/SoU`, `/committees/KrU`, `/committees/SkU`, `/committees/UFöU`:

- SoU shows goals, voteringar and UO9 with an amount.
- **KrU shows the LOVAT row with the our-data caption**, not a missing row.
- SkU shows the remit note.
- **UFöU renders** with its name and voteringar, and an empty KOSTAR row captioned — a joint committee bereder no utgiftsområde.

Then: `npx tsc --noEmit && npm run lint && npm run build`.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/features/committees/ frontend/src/App.tsx frontend/src/shared/types.ts
git commit -m "feat: committee page with LOVAT, RÖSTAT and KOSTAR

Three rows, each complete or explicitly captioned as our gap. RÖSTAT states
positions and draws no conclusion. KOSTAR shows amounts per utgiftsområde
with no share and no ranking, plus a plain-words note for SkU and FiU whose
remit exceeds their areas. No regional row and no anföranden row."
```

---

### Task 4: One shared goal renderer, and three party surfaces become two

Three surfaces render a party goal today — `PartiesPage` (count + topic tags), `PartyDetailPage`'s "Mål" tab (`:222`, `{topic} · {relevantVotes} matchade omröstningar`), and `PartyGoalsPage` (full card). Two of them state the same derived number in **two different wordings**, live. `GoalCard` is local to `PartyGoalsPage`, not shared.

**Files:**
- Create: `frontend/src/shared/components/GoalCard.tsx`
- Modify: `frontend/src/features/parties/PartyDetailPage.tsx`
- Modify: `frontend/src/App.tsx`
- Delete: `frontend/src/features/parties/PartyGoalsPage.tsx`

- [ ] **Step 1: Extract the shared renderer**

Create `frontend/src/shared/components/GoalCard.tsx` from the existing `GoalCard` in `PartyGoalsPage.tsx:14-90`, with three changes:

1. **Drop the vote count entirely.** Remove `matched`, `relevantVotes`, and both `"N matchade omröstningar"` and `"Ej prövad i någon omröstning ännu"`. No denominator beside a goal is honest: the keyword count under-claims because it is our keyword list, and the committee's full count over-claims because a committee deciding 40 things does not mean one promise was tested 40 times.
2. **Keyword matches become a labelled aid, not a measure**: where matches exist, link to them as *"omröstningar vars titel nämner {keywords}"* with the keywords visible and marked as ours.
3. **Never source our own gap to Riksdagen.** The old `<SourceMarker sourceId="riksdagen" />` beside "Ej prövad…" attributed our keyword gap to the record. It is gone with the sentence.

Keep `SpecificityBadge` and `TopicTag` — `topic` survives as a tag.

- [ ] **Step 2: Merge the record view into the party page**

In `PartyDetailPage.tsx`, make the "Mål" tab the full record view:

- **Group by committee, not by topic.** One subject taxonomy; `topic` is a tag on the card.
- **Render all committees from `GET /api/committees?period=…`**, not only those where the party has goals. MP has goals in 3 committees, L in 10 — so MP's page shows 3 populated and the rest captioned.
- Each empty committee row carries **identical wording from the shared renderer**, naming our seed reading as the gap: *"Vi har inga inlästa mål från {parti} inom detta område."* Never the party's silence.
- Keep the mandate-period and coverage line already there (`PartyGoalsPage.tsx:136-157`), including `denominatorCheckedAt`.
- Leave the Anföranden and Politiker tabs alone.

- [ ] **Step 3: Redirect the old route**

In `App.tsx`, replace the `/parties/:party/goals` route with a redirect so existing links and any shared URLs keep working:

The route is declared flat, so a relative `<Navigate to=".." />` resolves against the route hierarchy and lands on `/parties` — the list, not the party. Redirect to the party explicitly:

```tsx
function GoalsRedirect() {
  const { party } = useParams();
  return <Navigate to={`/parties/${party}`} replace />;
}
```

```tsx
<Route path="/parties/:party/goals" element={<GoalsRedirect />} />
```

Import `Navigate` and `useParams` from `react-router-dom`. Keep `/parties/:party/goals/:goalId/votes` working — `GoalVotesPage` is still reachable from the card's keyword-match link, and it must stay declared *after* this redirect route so the more specific path still matches.

- [ ] **Step 4: Delete the old page and confirm nothing references it**

```bash
git rm frontend/src/features/parties/PartyGoalsPage.tsx
grep -rn "PartyGoalsPage" frontend/src || echo "no references remain"
```

- [ ] **Step 5: Verify**

With the API and dev server running as in Task 3 Step 5, check in a browser:

- `/parties/MP` — Mål tab groups by committee, shows **5 populated and the remaining committees captioned**, and no vote count appears anywhere. (The plan originally said 3; verified against the dev database on 2026-08-11, MP has goals in 5 committees. Full distribution: C 8, KD 6, L 10, M 10, MP 5, S 6, SD 8, V 8.)
- `/parties/L` — 10 populated.
- `/parties/S/goals` — redirects to `/parties/S`.
- `grep -rn "matchade omröstningar\|Ej prövad" frontend/src` returns **nothing**.

Then `npx tsc --noEmit && npm run lint && npm run build`.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/shared/components/GoalCard.tsx frontend/src/features/parties/ frontend/src/App.tsx
git commit -m "refactor: one goal renderer, two party surfaces

Three surfaces rendered a goal and two stated the same derived number in
different wordings. One shared GoalCard now renders it everywhere,
including the committee page, so drift is impossible by construction
rather than policed.

The vote count is gone: neither denominator beside a goal is honest. The
keyword match becomes a labelled aid with the keywords visible, and the
SourceMarker that attributed our keyword gap to Riksdagen goes with it.

Goals group by committee, and every committee is listed per party with
empties captioned as our seed reading, never as the party's silence."
```

---

## Self-Review

**Spec coverage.** The spec's Phase 2 was "utskott→UO mapping table + data source doc, backend endpoints, the committee page with its three rows". The mapping and its doc landed in Phase 1 Task 3; the endpoints are Tasks 1–2 here plus Phase 1 Task 5; the page is Task 3. Task 4 is not in the spec — it comes from #100, which post-dates it, and is required because the committee page would otherwise be the *fourth* surface rendering a goal.

**Placeholders.** Tasks 1–2 carry complete Go and SQL. Tasks 3–4 are frontend and deliberately specify *requirements plus the exact copy strings and the `REMIT_NOTE` map* rather than full JSX: the components are large, the codebase has an established style to follow in `PartyGoalsPage`/`PartiesPage`, and inventing 200 lines of speculative JSX here would be less useful than naming every constraint and letting the implementer match the surrounding code. Every sentence the user sees is specified verbatim, which is the part that carries the decisions.

**Type consistency.** `CommitteeVotering` / `PartyPosition` field names are identical in the Go ports (Task 2 Step 3), the OpenAPI schema (Step 6) and `shared/types.ts` (Task 3 Step 1). `ListByCommittee` (goals) and `ListByCommitteeWithPositions` (votes) are distinct names on distinct features — deliberately, since they return different things.

**Ordering risk worth stating.** `TestListByCommittee_NumericOrdering` is written against real seeded data rather than a fixture, so it can pass vacuously if `AU` happens to hold no beteckning above 9. Check the assertion actually ran by adding `-v` and confirming AU returns beteckningar in double digits; the dev database has `AU` voteringar across four riksmöten, so it should.

**A dependency to respect.** Task 3 renders `GoalCard`, which Task 4 creates. Do **Task 4 before Task 3**, or build Task 3's LOVAT row against the shared component's signature and expect a compile error until Task 4 lands. The task order here is by subject, not by build order.
