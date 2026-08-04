# Pre-election record — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the site hold and honestly present the complete 2022–2026 parliamentary record before the 13 September 2026 election.

**Architecture:** Three ingestion defects currently cap the site at ~1.2% of the record (32 of 2 562 voteringar). Fix them, backfill four riksmöten, then make the site state plainly what it holds — freshness, coverage, and mandate scope — instead of implying completeness it does not have. No new information architecture: the utskott spine is post-election work.

**Tech Stack:** Go 1.26 (pgx/v5, chi), PostgreSQL 17, golang-migrate, React 19 + TypeScript, Riksdagen Open Data API.

## Global Constraints

- Governing decisions: wayfinder ticket [#83](https://github.com/Runemark7/swedish-democratic-tracker/issues/83). Do not re-litigate them.
- **Election: 13 September 2026.** The Riksdag is in recess — no new votes until after it. Backfill matters; live-catch-up does not.
- Never present absence as evidence. Partial coverage ships **only** if it states its own incompleteness.
- Strictly separate FACT from INTERPRETATION (`CONTEXT.md`). No verdicts, scores or rankings.
- Migrations idempotent (`IF EXISTS` / `IF NOT EXISTS`) — lesson from `000024` (#75).
- Raw parameterised SQL only. No ORM.
- Swedish UI strings; English code, comments, commits.
- Verification gates (CLAUDE.md rule 4): `go build -C backend ./...` and `cd frontend && npx tsc --noEmit` must pass before any task is done.
- DB integration tests follow `backend/internal/regions/adapters/postgres/budget_repo_test.go`: read `TEST_DATABASE_URL`, `t.Skip` when unset, `ZZ` sentinel fixtures cleaned in `t.Cleanup`.
- Client tests follow `backend/internal/regions/adapters/kolada/*_test.go`: `httptest.NewServer` plus a `newTestClient(url)` helper injecting `baseURL`.
- **The frontend has no test runner.** Only `frontend/scripts/__tests__/*.test.mjs` exists, run ad hoc via `node --test`. Frontend tasks below verify by explicit command or browser check, and say so. Do not add a test framework as part of this plan.

## ⚠️ Correction (2026-08-01, during Task 3)

The fetch strategy first written into this plan does not work. Verified against
the live API:

- **`/voteringlista` ignores `p`.** `p=1`, `p=2` and `p=3` return byte-identical
  rows (intersection 100 of 100). The earlier observation that page 20 still
  returned 500 rows was not deep pagination — it was the same first page every
  time.
- **`sz` is capped at 10 000.** `sz=50000` and `sz=200000` both return exactly
  10 000 rows — 95 of 759 vote points for one party-riksmöte.

So a whole party-riksmöte (~81 000 ballots) cannot be retrieved from that
endpoint at all, and a page-until-short loop would never terminate.

**Corrected strategy — partition by betänkande:**

1. Enumerate voteringar per riksmöte via `/dokumentlista?doktyp=votering&rm=X`,
   which *does* paginate correctly (`@traffar: 759`, `@sidor: 4`, zero overlap
   between pages) and supplies the coverage denominator.
2. Take the distinct `beteckning` values from that list.
3. Fetch ballots per betänkande via `/voteringlista?rm=X&bet=Y&sz=10000` with
   **no party filter** — one request returns every party and stays far under the
   cap (`AU9` = 1 047 rows, 3 vote points, 9 parties).

This is both correct and cheaper: roughly **620 requests** rather than the 1 788
originally planned, and no per-party loop at all. Betänkanden decided without a
vote return 0 rows, which is expected (`SoU1`).

**Tasks affected:** Task 3 drops the `Page` field entirely (shipping a parameter
the API ignores would be the same class of silent failure this plan exists to
fix). Tasks 4 and 6 must be rewritten around beteckning partitioning before they
are executed.

## Mandate period

`2022–2026` comprises exactly four riksmöten: **`2022/23`, `2023/24`, `2024/25`, `2025/26`** — 562 + 589 + 652 + 759 = **2 562 voteringar** (counts from `data.riksdagen.se/dokumentlista/?doktyp=votering&rm=<rm>`, retrieved 2026-08-01).

## File structure

| File | Responsibility |
|---|---|
| `backend/migrations/000030_party_goal_election_cycle.{up,down}.sql` | vintage column + backfill |
| `backend/migrations/000031_ingestion_coverage.{up,down}.sql` | per-riksmöte coverage facts |
| `backend/migrations/000032_mandate_periods.{up,down}.sql` | mandate definition seed |
| `backend/internal/votes/ports/riksdagen.go` | `FetchVotesFilter.Page` |
| `backend/internal/votes/adapters/riksdagen/client.go` | injectable `baseURL`, paging |
| `backend/internal/ingestion/workers/votes.go` | session iteration, cursor correctness |
| `backend/internal/ingestion/workers/speeches.go` | same, for speeches |
| `backend/cmd/backfill/main.go` | one-shot backfill runner |
| `frontend/src/shared/design.ts` | committee regex + map |
| `frontend/src/App.tsx` | freshness badge |

---

### Task 1: Separate promise vintages

Fixes a **live neutrality defect**: S is the only party showing 2026 campaign material, six weeks before an election. Do this first — it is small, independent, and currently visible in production.

**Files:**
- Create: `backend/migrations/000030_party_goal_election_cycle.{up,down}.sql`
- Modify: `backend/internal/goals/domain/goal.go`, `backend/internal/goals/adapters/postgres/repository.go`
- Test: `backend/internal/goals/adapters/postgres/election_cycle_test.go`

**Interfaces:**
- Produces: `party_goals.election_cycle TEXT NULL` — `'2022'`, `'2026'`, or NULL for material not tied to an election. Repository list methods exclude `'2026'`.

- [ ] **Step 1: Write the failing test**

```go
package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func connectGoalsTestDB(t *testing.T) *pgxpool.Pool {
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

// Every seeded goal must carry a vintage derived from its source document,
// and no 2026 campaign material may leak into the record view.
func TestElectionCycle_BackfilledFromSourceDocument(t *testing.T) {
	pool := connectGoalsTestDB(t)
	ctx := context.Background()

	cases := map[string]*string{
		"Valmanifest 2022":   strptr("2022"),
		"Tidöavtalet 2022":   strptr("2022"),
		"Valplattform 2022":  strptr("2022"),
		"Valplattform 2026":  strptr("2026"),
		"Partiprogram":       nil,
	}
	for doc, want := range cases {
		rows, err := pool.Query(ctx,
			`SELECT DISTINCT election_cycle FROM party_goals WHERE source_document = $1`, doc)
		if err != nil {
			t.Fatalf("query %s: %v", doc, err)
		}
		var got []*string
		for rows.Next() {
			var v *string
			if err := rows.Scan(&v); err != nil {
				t.Fatal(err)
			}
			got = append(got, v)
		}
		rows.Close()
		if len(got) == 0 {
			continue // that source document is not present in this DB
		}
		if len(got) != 1 {
			t.Errorf("%s: expected one cycle value, got %d", doc, len(got))
			continue
		}
		if (got[0] == nil) != (want == nil) || (got[0] != nil && want != nil && *got[0] != *want) {
			t.Errorf("%s: election_cycle = %v, want %v", doc, deref(got[0]), deref(want))
		}
	}
}

func strptr(s string) *string { return &s }
func deref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
```

- [ ] **Step 2: Run it to verify it fails**

```bash
cd backend && TEST_DATABASE_URL='postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable' \
  go test ./internal/goals/adapters/postgres/ -run TestElectionCycle -v
```

Expected: FAIL — `column "election_cycle" does not exist`.

If it SKIPs, start postgres: `docker compose -f docker-compose.dev.yml up -d postgres`.

- [ ] **Step 3: Write the migration**

`000030_party_goal_election_cycle.up.sql`:

```sql
-- Promise vintage. The 2022-2026 voting record can only speak to promises made
-- at or before the 2022 election; 2026 campaign material is a different question
-- ("what do they promise now") and must not sit beside a four-year voting record.
--
-- NULL means the goal is not tied to an election cycle (standing party programme).

ALTER TABLE party_goals ADD COLUMN IF NOT EXISTS election_cycle TEXT;

ALTER TABLE party_goals DROP CONSTRAINT IF EXISTS party_goals_election_cycle_check;
ALTER TABLE party_goals
  ADD CONSTRAINT party_goals_election_cycle_check
  CHECK (election_cycle IS NULL OR election_cycle IN ('2022', '2026'));

UPDATE party_goals SET election_cycle = '2026'
  WHERE source_document ILIKE '%2026%' AND election_cycle IS DISTINCT FROM '2026';

UPDATE party_goals SET election_cycle = '2022'
  WHERE source_document ILIKE '%2022%' AND election_cycle IS NULL;

CREATE INDEX IF NOT EXISTS idx_party_goals_election_cycle
  ON party_goals (election_cycle);
```

`000030_party_goal_election_cycle.down.sql`:

```sql
DROP INDEX IF EXISTS idx_party_goals_election_cycle;
ALTER TABLE party_goals DROP CONSTRAINT IF EXISTS party_goals_election_cycle_check;
ALTER TABLE party_goals DROP COLUMN IF EXISTS election_cycle;
```

- [ ] **Step 4: Apply and re-run the test**

```bash
make migrate
cd backend && TEST_DATABASE_URL='postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable' \
  go test ./internal/goals/adapters/postgres/ -run TestElectionCycle -v
```

Expected: PASS.

- [ ] **Step 5: Exclude 2026 material from the record view**

In `backend/internal/goals/adapters/postgres/repository.go`, every query that lists goals for display (`ListByParty`, `ListByTopic`, `ListAll`) gains:

```sql
AND (election_cycle IS NULL OR election_cycle <> '2026')
```

Add this comment above the shared WHERE fragment:

```go
// 2026 campaign material is excluded from the record view: the 2022-2026 votes
// cannot speak to a promise made in 2026. A 2026 platform view ships only when
// all eight parties are represented (wayfinder #83), and does not exist yet.
```

- [ ] **Step 6: Verify parity**

```bash
go build -C backend ./... && echo "GO OK"
# restart backend, then:
for p in S M SD C V KD L MP; do
  echo -n "$p: "
  curl -s "localhost:8080/api/parties/$p/goals" | python3 -c "
import json,sys,collections
print(collections.Counter(g['sourceDocument'] for g in json.load(sys.stdin)))"
done
```

Expected: **no party shows any 2026 document.** S drops from 13 goals to 6.

- [ ] **Step 7: Commit**

```bash
git add backend/migrations/000030_* backend/internal/goals/ 
git commit -m "fix(goals): exclude 2026 campaign material from the 2022-2026 record view"
```

---

### Task 2: Fix committee resolution (MJU, KrU)

Two committees are invisible today. Independent of everything else.

**Files:**
- Modify: `frontend/src/shared/design.ts:85-106`

**Interfaces:**
- Produces: `committeeFromBeteckning(b: string): string | undefined` resolving all 15 committees.

- [ ] **Step 1: Reproduce both bugs**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
node -e '
const re = /^([A-ZÅÄÖ][a-zåäöÅÄÖ]*U)/;
for (const b of ["MJU12","KrU5","AU9"]) {
  const m = b.match(re);
  console.log(b, "->", m ? m[1] : "NULL");
}'
```

Expected now: `MJU12 -> NULL` (the class excludes uppercase `J`), `KrU5 -> KrU` (extracts, but is absent from `COMMITTEES`).

- [ ] **Step 2: Fix the regex and add the missing committee**

In `frontend/src/shared/design.ts`, add to `COMMITTEES`:

```ts
  KrU: "Kulturutskottet",
```

and replace the extractor:

```ts
/** Extract committee name from beteckning, e.g. "SoU12" → "Socialutskottet".
 *  The code is the leading letters up to and including the final "U"; it may
 *  contain interior uppercase (MJU) and Swedish vowels (FöU). */
export function committeeFromBeteckning(beteckning: string): string | undefined {
  const match = beteckning.match(/^([A-ZÅÄÖa-zåäö]*U)(?=\d|$)/);
  return match ? COMMITTEES[match[1]] : undefined;
}
```

- [ ] **Step 3: Verify all 15 resolve**

```bash
node -e '
const re = /^([A-ZÅÄÖa-zåäö]*U)(?=\d|$)/;
const C = ["AU","CU","FiU","FöU","JuU","KU","KrU","MJU","NU","SfU","SkU","SoU","TU","UbU","UU"];
let bad = 0;
for (const c of C) { const m = (c+"12").match(re); if (!m || m[1] !== c) { console.log("FAIL", c, m && m[1]); bad++; } }
console.log(bad === 0 ? "all 15 resolve" : bad + " failing");'
```

Expected: `all 15 resolve`.

- [ ] **Step 4: Typecheck and commit**

```bash
cd frontend && npx tsc --noEmit && echo "TSC OK"
git add frontend/src/shared/design.ts
git commit -m "fix(committees): resolve MJU and add missing KrU"
```

---

### Task 3: Make the votes client injectable and preserve SystemDatum

**Status: done** (commit on `feat/votes-client-pagination`). Shipped without the `Page` field
originally specified — see the Correction above. The client gains an injectable `baseURL`
(matching the Kolada adapter, so it can be exercised with `httptest`), keeps Riksdagen's
`systemdatum` on `domain.Vote` for the ingestion cursor, and defaults `sz` to the API's
10 000-row ceiling. The endpoint's real constraints are documented on `FetchVotesFilter`
so they are not rediscovered the hard way.

**Files:**
- Modify: `backend/internal/votes/ports/riksdagen.go:10-17`, `backend/internal/votes/adapters/riksdagen/client.go:15-31`
- Test: `backend/internal/votes/adapters/riksdagen/client_test.go`

**Interfaces:**
- Produces: `FetchVotesFilter.Page int` (1-based; 0 treated as 1). `Client.baseURL` becomes a struct field. `newTestClient(url string) *Client` for tests.

- [ ] **Step 1: Write the failing test**

```go
package riksdagen

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"riksdagskollen/internal/votes/ports"
)

func newTestClient(url string) *Client {
	return &Client{http: &http.Client{Timeout: 5 * time.Second}, baseURL: url}
}

func votePage(n int) string {
	out := `{"voteringlista":{"votering":[`
	for i := 0; i < n; i++ {
		if i > 0 {
			out += ","
		}
		out += fmt.Sprintf(`{"votering_id":"v%d","intressent_id":"p%d","namn":"T","parti":"S",`+
			`"rost":"Ja","beteckning":"AU9","punkt":"1","rm":"2025/26","dok_id":"d",`+
			`"systemdatum":"2026-06-17 10:00:00"}`, i, i)
	}
	return out + `]}}`
}

func TestFetchVotes_SendsRequestedPage(t *testing.T) {
	var mu sync.Mutex
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.URL.Query().Get("p"))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(votePage(2)))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	if _, err := c.FetchVotes(context.Background(), ports.FetchVotesFilter{
		Session: "2025/26", Party: "S", Size: 500, Page: 3,
	}); err != nil {
		t.Fatalf("FetchVotes: %v", err)
	}
	if len(seen) != 1 || seen[0] != "3" {
		t.Errorf("requested pages = %v, want [3]", seen)
	}
}

func TestFetchVotes_DefaultsToPageOne(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query().Get("p")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(votePage(1)))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	if _, err := c.FetchVotes(context.Background(), ports.FetchVotesFilter{
		Session: "2025/26", Party: "S",
	}); err != nil {
		t.Fatalf("FetchVotes: %v", err)
	}
	if got != "1" {
		t.Errorf("page = %q, want \"1\"", got)
	}
}

func TestFetchVotes_ReturnsAllRowsOnPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(votePage(7)))
	}))
	defer srv.Close()

	vv, err := newTestClient(srv.URL).FetchVotes(context.Background(),
		ports.FetchVotesFilter{Session: "2025/26", Party: "S"})
	if err != nil {
		t.Fatalf("FetchVotes: %v", err)
	}
	if len(vv) != 7 {
		t.Errorf("got %d votes, want 7", len(vv))
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

```bash
cd backend && go test ./internal/votes/adapters/riksdagen/ -run TestFetchVotes -v
```

Expected: compile failure — `Page` is not a field of `FetchVotesFilter`, and `Client` has no `baseURL` field.

- [ ] **Step 3: Add the field and the page parameter**

In `ports/riksdagen.go`, add to `FetchVotesFilter`:

```go
	Page         int       // 1-based page; 0 is treated as 1. The API reports no
	                       // total, so callers page until a short page comes back.
```

In `adapters/riksdagen/client.go`, replace the const with a field and thread the page through:

```go
const defaultBaseURL = "https://data.riksdagen.se"

type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient() *Client {
	return &Client{
		http:    &http.Client{Timeout: 30 * time.Second},
		baseURL: defaultBaseURL,
	}
}
```

and in `FetchVotes`:

```go
	size := f.Size
	if size == 0 {
		size = 500
	}
	page := f.Page
	if page == 0 {
		page = 1
	}
	url := fmt.Sprintf("%s/voteringlista/?rm=%s&parti=%s&iid=%s&bet=%s&sz=%d&p=%d&utformat=json",
		c.baseURL, f.Session, f.Party, f.PoliticianID, f.Beteckning, size, page)
```

Replace every other `baseURL` reference in this file with `c.baseURL`.

- [ ] **Step 4: Run the tests**

```bash
cd backend && go test ./internal/votes/adapters/riksdagen/ -run TestFetchVotes -v
```

Expected: all three PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/votes/ports/riksdagen.go backend/internal/votes/adapters/riksdagen/
git commit -m "feat(votes): paginate the Riksdagen vote client"
```

---

### Task 4: Enumerate voteringar, then fetch by betänkande

Rewritten 2026-08-01: the original per-party page loop cannot work (see Correction). Two defects remain to fix in the worker itself — `Session: currentSession` is the hardcoded constant `"2024/25"` (`speeches.go:13`), so it fetches the session before last; and the cursor advances to `now()` unconditionally (`votes.go:47`), permanently skipping any window it fails to retrieve.

**Files:**
- Modify: `backend/internal/votes/ports/riksdagen.go`, `backend/internal/votes/adapters/riksdagen/client.go`
- Modify: `backend/internal/ingestion/workers/votes.go:24-56`
- Test: `backend/internal/ingestion/workers/votes_worker_test.go`, `backend/internal/votes/adapters/riksdagen/client_test.go`

**Interfaces:**
- Consumes: `FetchVotesFilter.Beteckning`, `domain.Vote.SystemDatum` (Task 3).
- Produces:
  - `ports.VoteringRef{Beteckning, Organ, DokID, Date string; SystemDatum time.Time}`
  - `ListVoteringar(ctx, rm string, page, size int) ([]VoteringRef, int, error)` on `RiksdagenVoteClient` — returns one page newest-first plus the API's `@traffar` total.
  - `workers.MandateRiksmoten = []string{"2022/23","2023/24","2024/25","2025/26"}`
  - `workers.CurrentRiksmote(t time.Time) string`

- [ ] **Step 1: Write the failing client test**

Append to `backend/internal/votes/adapters/riksdagen/client_test.go`:

```go
const voteringListPage = `{"dokumentlista":{"@traffar":"759","@sidor":"4","dokument":[
 {"dok_id":"HD19UbU31p2","beteckning":"UbU31","rm":"2025/26","organ":"UbU",
  "datum":"2026-06-17","systemdatum":"2026-06-17 14:12:03"},
 {"dok_id":"HD19TU20p9","beteckning":"TU20","rm":"2025/26","organ":"TU",
  "datum":"2026-06-17","systemdatum":"2026-06-17 13:58:01"}]}}`

func TestListVoteringar_ParsesRefsAndTotal(t *testing.T) {
	var q url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(voteringListPage))
	}))
	defer srv.Close()

	refs, total, err := newTestClient(srv.URL).ListVoteringar(context.Background(), "2025/26", 2, 200)
	if err != nil {
		t.Fatalf("ListVoteringar: %v", err)
	}
	if total != 759 {
		t.Errorf("total = %d, want 759", total)
	}
	if len(refs) != 2 {
		t.Fatalf("got %d refs, want 2", len(refs))
	}
	if refs[0].Beteckning != "UbU31" || refs[0].Organ != "UbU" {
		t.Errorf("ref[0] = %+v", refs[0])
	}
	want := time.Date(2026, 6, 17, 14, 12, 3, 0, time.UTC)
	if !refs[0].SystemDatum.Equal(want) {
		t.Errorf("SystemDatum = %v, want %v", refs[0].SystemDatum, want)
	}
	// Unlike /voteringlista, /dokumentlista honours p — it must be sent, and
	// results must come back newest-first so the worker can stop at the cursor.
	if q.Get("p") != "2" {
		t.Errorf("p = %q, want \"2\"", q.Get("p"))
	}
	if q.Get("sortorder") != "desc" {
		t.Errorf("sortorder = %q, want \"desc\"", q.Get("sortorder"))
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

```bash
cd backend && go test ./internal/votes/adapters/riksdagen/ -run TestListVoteringar -v
```

Expected: compile failure — `ListVoteringar` undefined.

- [ ] **Step 3: Add the port type and method**

In `backend/internal/votes/ports/riksdagen.go`:

```go
// VoteringRef identifies one votering (a single förslagspunkt decided by vote).
// Enumerated from /dokumentlista, which — unlike /voteringlista — paginates
// correctly and reports a total.
type VoteringRef struct {
	Beteckning  string // "UbU31"
	Organ       string // "UbU"
	DokID       string
	Date        string // "2026-06-17"
	SystemDatum time.Time
}
```

and on the `RiksdagenVoteClient` interface:

```go
	// ListVoteringar enumerates voteringar for a riksmöte, newest first.
	// Returns one page and the total the API reports (@traffar), which is also
	// the coverage denominator. page is 1-based.
	ListVoteringar(ctx context.Context, rm string, page, size int) ([]VoteringRef, int, error)
```

- [ ] **Step 4: Implement it in the client**

In `backend/internal/votes/adapters/riksdagen/client.go`:

```go
func (c *Client) ListVoteringar(ctx context.Context, rm string, page, size int) ([]ports.VoteringRef, int, error) {
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 200
	}
	url := fmt.Sprintf(
		"%s/dokumentlista/?doktyp=votering&rm=%s&utformat=json&sz=%d&p=%d&sort=datum&sortorder=desc",
		c.baseURL, rm, size, page)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("riksdagen dokumentlista returned %d", resp.StatusCode)
	}

	var payload struct {
		Dokumentlista struct {
			Traffar  string `json:"@traffar"`
			Dokument []struct {
				DokID       string `json:"dok_id"`
				Beteckning  string `json:"beteckning"`
				Organ       string `json:"organ"`
				Datum       string `json:"datum"`
				Systemdatum string `json:"systemdatum"`
			} `json:"dokument"`
		} `json:"dokumentlista"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, 0, fmt.Errorf("decode dokumentlista: %w", err)
	}

	total := 0
	_, _ = fmt.Sscanf(payload.Dokumentlista.Traffar, "%d", &total)

	refs := make([]ports.VoteringRef, 0, len(payload.Dokumentlista.Dokument))
	for _, d := range payload.Dokumentlista.Dokument {
		var sd time.Time
		if t, err := time.Parse("2006-01-02 15:04:05", d.Systemdatum); err == nil {
			sd = t
		}
		refs = append(refs, ports.VoteringRef{
			Beteckning:  d.Beteckning,
			Organ:       d.Organ,
			DokID:       d.DokID,
			Date:        d.Datum,
			SystemDatum: sd,
		})
	}
	return refs, total, nil
}
```

Add the same method to any stub or fake implementing `RiksdagenVoteClient`, or the build will break.

- [ ] **Step 5: Run the client test**

```bash
cd backend && go test ./internal/votes/adapters/riksdagen/ -run TestListVoteringar -v
```

Expected: PASS.

- [ ] **Step 6: Write the failing worker test**

Create `backend/internal/ingestion/workers/votes_worker_test.go`:

```go
package workers_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	ingPorts "riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/ingestion/workers"
	"riksdagskollen/internal/votes"
	votedomain "riksdagskollen/internal/votes/domain"
	voteports "riksdagskollen/internal/votes/ports"
)

type fakeVoteClient struct {
	mu        sync.Mutex
	refs      []voteports.VoteringRef
	total     int
	byBet     map[string][]*votedomain.Vote
	fetched   []string // beteckningar actually fetched
	listErr   error
	fetchErr  error
}

func (f *fakeVoteClient) ListVoteringar(_ context.Context, _ string, page, _ int) ([]voteports.VoteringRef, int, error) {
	if f.listErr != nil {
		return nil, 0, f.listErr
	}
	if page > 1 {
		return nil, f.total, nil // single page of results
	}
	return f.refs, f.total, nil
}

func (f *fakeVoteClient) FetchVotes(_ context.Context, flt voteports.FetchVotesFilter) ([]*votedomain.Vote, error) {
	if f.fetchErr != nil {
		return nil, f.fetchErr
	}
	f.mu.Lock()
	f.fetched = append(f.fetched, flt.Beteckning)
	f.mu.Unlock()
	return f.byBet[flt.Beteckning], nil
}

func (f *fakeVoteClient) FetchDocumentStatus(context.Context, string) (*votedomain.DocumentStatus, error) {
	return nil, nil
}
func (f *fakeVoteClient) FetchDocuments(context.Context, []string, int) ([]voteports.RiksdagDocument, error) {
	return nil, nil
}
func (f *fakeVoteClient) FetchBetankandeByBeteckning(context.Context, string) (*voteports.BetankandeInfo, error) {
	return nil, nil
}

type fakeVoteRepo struct{ upserted int }

func (r *fakeVoteRepo) UpsertMany(_ context.Context, vv []*votedomain.Vote) error {
	r.upserted += len(vv)
	return nil
}

type fakeCursors struct{ cur *ingPorts.Cursor }

func (c *fakeCursors) Get(context.Context, string) (*ingPorts.Cursor, error) { return c.cur, nil }
func (c *fakeCursors) Upsert(_ context.Context, x ingPorts.Cursor) error     { c.cur = &x; return nil }

func ref(bet string, sd time.Time) voteports.VoteringRef {
	return voteports.VoteringRef{Beteckning: bet, SystemDatum: sd}
}

func vote(bet string, sd time.Time) *votedomain.Vote {
	return &votedomain.Vote{Beteckning: bet, SystemDatum: sd}
}

// Work is partitioned by betänkande: each distinct beteckning is fetched once,
// with no party filter, because one request returns every party's ballots.
func TestVotesWorker_FetchesEachBeteckningOnce(t *testing.T) {
	newer := time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC)
	client := &fakeVoteClient{
		refs: []voteports.VoteringRef{
			ref("UbU31", newer), ref("TU20", newer), ref("UbU31", newer), // duplicate
		},
		total: 3,
		byBet: map[string][]*votedomain.Vote{
			"UbU31": {vote("UbU31", newer)},
			"TU20":  {vote("TU20", newer)},
		},
	}
	cur := &fakeCursors{}
	w := workers.NewVotesWorker(votes.NewService(&fakeVoteRepo{}, client), cur)

	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(client.fetched) != 2 {
		t.Errorf("fetched %v, want each beteckning exactly once", client.fetched)
	}
}

// A failed fetch must not move the cursor: that is how the site fell months
// behind while logging success every day.
func TestVotesWorker_CursorUnchangedOnFetchError(t *testing.T) {
	before := time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC)
	cur := &fakeCursors{cur: &ingPorts.Cursor{DataType: "votes", LastDate: &before}}
	client := &fakeVoteClient{
		refs:     []voteports.VoteringRef{ref("UbU31", time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC))},
		total:    1,
		fetchErr: errors.New("upstream down"),
	}
	w := workers.NewVotesWorker(votes.NewService(&fakeVoteRepo{}, client), cur)
	_ = w.Run(context.Background())

	if !cur.cur.LastDate.Equal(before) {
		t.Errorf("cursor moved to %v on error, want unchanged %v", cur.cur.LastDate, before)
	}
}

// An empty result is not evidence of freshness.
func TestVotesWorker_CursorUnchangedWhenNothingFetched(t *testing.T) {
	before := time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC)
	cur := &fakeCursors{cur: &ingPorts.Cursor{DataType: "votes", LastDate: &before}}
	w := workers.NewVotesWorker(
		votes.NewService(&fakeVoteRepo{}, &fakeVoteClient{}), cur)
	_ = w.Run(context.Background())

	if !cur.cur.LastDate.Equal(before) {
		t.Errorf("cursor moved to %v with no rows, want unchanged %v", cur.cur.LastDate, before)
	}
}

// The cursor advances only as far as data actually retrieved.
func TestVotesWorker_CursorAdvancesToNewestFetched(t *testing.T) {
	newest := time.Date(2026, 6, 17, 14, 12, 3, 0, time.UTC)
	client := &fakeVoteClient{
		refs:  []voteports.VoteringRef{ref("UbU31", newest)},
		total: 1,
		byBet: map[string][]*votedomain.Vote{"UbU31": {vote("UbU31", newest)}},
	}
	cur := &fakeCursors{}
	w := workers.NewVotesWorker(votes.NewService(&fakeVoteRepo{}, client), cur)
	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cur.cur == nil || !cur.cur.LastDate.Equal(newest) {
		t.Errorf("cursor = %v, want %v", cur.cur, newest)
	}
}

func TestCurrentRiksmote(t *testing.T) {
	cases := map[string]string{
		"2026-08-01": "2025/26", // before the new riksmöte opens in September
		"2026-09-20": "2026/27",
		"2026-01-15": "2025/26",
	}
	for in, want := range cases {
		d, _ := time.Parse("2006-01-02", in)
		if got := workers.CurrentRiksmote(d); got != want {
			t.Errorf("CurrentRiksmote(%s) = %q, want %q", in, got, want)
		}
	}
}
```

- [ ] **Step 7: Run it to verify it fails**

```bash
cd backend && go test ./internal/ingestion/workers/ -run 'TestVotesWorker|TestCurrentRiksmote' -v
```

Expected: FAIL — `CurrentRiksmote` undefined and the worker still loops parties.

- [ ] **Step 8: Rewrite the worker**

Replace `Run` in `backend/internal/ingestion/workers/votes.go`:

```go
// MandateRiksmoten is the 2022-2026 mandate period: four riksmöten,
// 2 562 voteringar in total.
var MandateRiksmoten = []string{"2022/23", "2023/24", "2024/25", "2025/26"}

const voteringPageSize = 200

// CurrentRiksmote returns the riksmöte label for a date. A riksmöte runs from
// September to September, so anything before September belongs to the one that
// opened the previous year.
func CurrentRiksmote(t time.Time) string {
	y := t.Year()
	if t.Month() < time.September {
		y--
	}
	return fmt.Sprintf("%d/%02d", y, (y+1)%100)
}

func (w *VotesWorker) Run(ctx context.Context) error {
	var since time.Time
	if cur, err := w.cursors.Get(ctx, "votes"); err == nil && cur != nil && cur.LastDate != nil {
		since = *cur.LastDate
	}
	rm := CurrentRiksmote(time.Now().UTC())
	slog.Info("votes: incremental sync", "riksmote", rm, "since", since.Format("2006-01-02"))

	// Walk the enumeration newest-first and stop at the cursor. /dokumentlista
	// paginates correctly, unlike /voteringlista.
	seen := map[string]bool{}
	var order []string
	stop := false
	for page := 1; !stop; page++ {
		refs, _, err := w.svc.ListVoteringar(ctx, rm, page, voteringPageSize)
		if err != nil {
			slog.Warn("votes: enumeration failed", "rm", rm, "page", page, "error", err)
			return nil // cursor untouched
		}
		if len(refs) == 0 {
			break
		}
		for _, r := range refs {
			if !since.IsZero() && !r.SystemDatum.After(since) {
				stop = true
				break
			}
			if !seen[r.Beteckning] {
				seen[r.Beteckning] = true
				order = append(order, r.Beteckning)
			}
		}
		if len(refs) < voteringPageSize {
			break
		}
	}

	// One request per betänkande returns every party's ballots for it.
	var newest time.Time
	for _, bet := range order {
		vv, err := w.svc.FetchAndStore(ctx, ports.FetchVotesFilter{
			Session: rm, Beteckning: bet, Since: since,
		})
		if err != nil {
			slog.Warn("votes: fetch failed", "rm", rm, "bet", bet, "error", err)
			continue
		}
		for _, v := range vv {
			if v.SystemDatum.After(newest) {
				newest = v.SystemDatum
			}
		}
		time.Sleep(200 * time.Millisecond)
	}

	if newest.IsZero() {
		slog.Info("votes: nothing fetched, cursor unchanged", "betankanden", len(order))
		return nil
	}
	return w.cursors.Upsert(ctx, ingPorts.Cursor{DataType: "votes", LastDate: &newest})
}
```

Add the two service passthroughs to `backend/internal/votes/service.go`:

```go
// ListVoteringar enumerates voteringar for a riksmöte, newest first.
func (s *Service) ListVoteringar(ctx context.Context, rm string, page, size int) ([]ports.VoteringRef, int, error) {
	return s.riksdagen.ListVoteringar(ctx, rm, page, size)
}

// FetchAndStore fetches one betänkande's ballots and returns what was stored,
// so callers can track the newest record actually retrieved.
func (s *Service) FetchAndStore(ctx context.Context, f ports.FetchVotesFilter) ([]*domain.Vote, error) {
	vv, err := s.riksdagen.FetchVotes(ctx, f)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpsertMany(ctx, vv); err != nil {
		return nil, err
	}
	return vv, nil
}
```

- [ ] **Step 9: Run the tests**

```bash
cd backend && go test ./internal/ingestion/workers/ -run 'TestVotesWorker|TestCurrentRiksmote' -v
go build ./...
```

Expected: all five PASS, build clean.

- [ ] **Step 10: Commit**

```bash
git add backend/internal/votes/ backend/internal/ingestion/workers/votes.go backend/internal/ingestion/workers/votes_worker_test.go
git commit -m "fix(votes): enumerate voteringar then fetch by betankande, correct the cursor"
```

---

### Task 5: Speeches — one request per party-riksmöte, with a truncation guard

Rewritten 2026-08-01. `/anforandelista` shares `/voteringlista`'s defect — `p` is ignored (p=1 and p=2 return identical rows) and there is no `@traffar`. But it is **not capped in practice**: S's whole 2025/26 riksmöte is 2 849 rows, far below the 10 000 ceiling. So speeches needs one large request per party-riksmöte — 8 × 4 = **32 requests** — not pagination.

The danger is silent truncation: if a party-riksmöte ever reaches 10 000 there is no way to page for the remainder, and the shortfall would be invisible. The worker must fail loudly rather than store a truncated riksmöte.

**Files:**
- Modify: `backend/internal/ingestion/workers/speeches.go`, `backend/internal/speeches/adapters/riksdagen/client.go`, `backend/internal/speeches/service.go`
- Test: `backend/internal/ingestion/workers/speeches_worker_test.go`

**Interfaces:**
- Consumes: `workers.MandateRiksmoten` (Task 4).
- Produces: `speeches.Service.FetchAndStore(ctx, ports.FetchSpeechesFilter) ([]*domain.Speech, error)`.

- [ ] **Step 1: Write the failing test**

Create `backend/internal/ingestion/workers/speeches_worker_test.go`:

```go
package workers_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	ingPorts "riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/ingestion/workers"
	"riksdagskollen/internal/speeches"
	speechdomain "riksdagskollen/internal/speeches/domain"
	speechports "riksdagskollen/internal/speeches/ports"
)

type fakeSpeechClient struct {
	mu       sync.Mutex
	sessions map[string]bool
	rows     []*speechdomain.Speech
	err      error
}

func (f *fakeSpeechClient) FetchSpeeches(_ context.Context, flt speechports.FetchSpeechesFilter) ([]*speechdomain.Speech, error) {
	f.mu.Lock()
	if f.sessions == nil {
		f.sessions = map[string]bool{}
	}
	f.sessions[flt.Session] = true
	f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	return f.rows, nil
}

func (f *fakeSpeechClient) FetchSpeechText(context.Context, string, string) (string, error) {
	return "", nil
}

type fakeSpeechRepo struct{ upserted int }

func (r *fakeSpeechRepo) UpsertMany(_ context.Context, ss []*speechdomain.Speech) error {
	r.upserted += len(ss)
	return nil
}

type fakeSpeechCursors struct{ cur *ingPorts.Cursor }

func (c *fakeSpeechCursors) Get(context.Context, string) (*ingPorts.Cursor, error) { return c.cur, nil }
func (c *fakeSpeechCursors) Upsert(_ context.Context, x ingPorts.Cursor) error {
	c.cur = &x
	return nil
}

// The worker must cover the whole mandate, not the hardcoded 2024/25 session.
func TestSpeechesWorker_UsesMandateRiksmoten(t *testing.T) {
	client := &fakeSpeechClient{}
	w := workers.NewSpeechesWorker(
		speeches.NewService(&fakeSpeechRepo{}, client), &fakeSpeechCursors{})
	_ = w.Run(context.Background())

	for _, rm := range workers.MandateRiksmoten {
		if !client.sessions[rm] {
			t.Errorf("riksmöte %s was never requested", rm)
		}
	}
}

// A failed fetch must not move the cursor.
func TestSpeechesWorker_CursorUnchangedOnFetchError(t *testing.T) {
	before := time.Date(2025, 9, 4, 0, 0, 0, 0, time.UTC)
	cur := &fakeSpeechCursors{cur: &ingPorts.Cursor{DataType: "speeches", LastDate: &before}}
	w := workers.NewSpeechesWorker(
		speeches.NewService(&fakeSpeechRepo{}, &fakeSpeechClient{err: errors.New("down")}), cur)
	_ = w.Run(context.Background())

	if !cur.cur.LastDate.Equal(before) {
		t.Errorf("cursor moved to %v on error, want unchanged %v", cur.cur.LastDate, before)
	}
}

// An empty result is not evidence of freshness.
func TestSpeechesWorker_CursorUnchangedWhenNothingFetched(t *testing.T) {
	before := time.Date(2025, 9, 4, 0, 0, 0, 0, time.UTC)
	cur := &fakeSpeechCursors{cur: &ingPorts.Cursor{DataType: "speeches", LastDate: &before}}
	w := workers.NewSpeechesWorker(
		speeches.NewService(&fakeSpeechRepo{}, &fakeSpeechClient{}), cur)
	_ = w.Run(context.Background())

	if !cur.cur.LastDate.Equal(before) {
		t.Errorf("cursor moved to %v with no rows, want unchanged %v", cur.cur.LastDate, before)
	}
}

// /anforandelista cannot paginate, so a full page means the riksmöte was
// truncated and the shortfall is undetectable downstream. Fail loudly.
func TestSpeechesWorker_ErrorsOnPossibleTruncation(t *testing.T) {
	rows := make([]*speechdomain.Speech, workers.SpeechFetchLimit)
	for i := range rows {
		rows[i] = &speechdomain.Speech{}
	}
	w := workers.NewSpeechesWorker(
		speeches.NewService(&fakeSpeechRepo{}, &fakeSpeechClient{rows: rows}),
		&fakeSpeechCursors{})

	if err := w.Run(context.Background()); err == nil {
		t.Error("expected an error when a party-riksmöte fills the fetch limit")
	}
}
```

If `speeches.NewService` or the repository interface differs in shape, adjust the fakes to satisfy the real interfaces — do not weaken the assertions.

- [ ] **Step 2: Run it to verify it fails**

```bash
cd backend && go test ./internal/ingestion/workers/ -run TestSpeechesWorker -v
```

Expected: FAIL — `SpeechFetchLimit` undefined; only `2024/25` requested; cursor advances regardless.

- [ ] **Step 3: Remove the hardcoded session**

```bash
grep -rn "currentSession" backend/
```

Delete `const currentSession = "2024/25"` from `speeches.go:13`. Task 4 already removed the votes worker's use of it, so nothing else should reference it.

- [ ] **Step 4: Raise the client's page size**

In `backend/internal/speeches/adapters/riksdagen/client.go`, move `baseURL` onto the struct exactly as the votes client now does:

```go
const defaultBaseURL = "https://data.riksdagen.se"

type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient() *Client {
	return &Client{
		http:    &http.Client{Timeout: 30 * time.Second},
		baseURL: defaultBaseURL,
	}
}
```

Replace every bare `baseURL` in the file with `c.baseURL`. Do **not** add a `p` parameter — the endpoint ignores it.

- [ ] **Step 5: Rewrite the worker**

```go
// SpeechFetchLimit is the largest page /anforandelista will return. The
// endpoint ignores `p`, so a response of exactly this size means the
// party-riksmöte was truncated with no way to retrieve the remainder.
const SpeechFetchLimit = 10000

func (w *SpeechesWorker) Run(ctx context.Context) error {
	var since time.Time
	if cur, err := w.cursors.Get(ctx, "speeches"); err == nil && cur != nil && cur.LastDate != nil {
		since = *cur.LastDate
		slog.Info("speeches: incremental sync", "since", since.Format("2006-01-02"))
	}

	var newest time.Time
	for _, rm := range MandateRiksmoten {
		for _, party := range ActiveParties {
			ss, err := w.svc.FetchAndStore(ctx, ports.FetchSpeechesFilter{
				Session: rm, Party: party, Size: SpeechFetchLimit, Since: since,
			})
			if err != nil {
				slog.Warn("speeches sync failed", "rm", rm, "party", party, "error", err)
				continue
			}
			if len(ss) >= SpeechFetchLimit {
				// Storing a silently truncated riksmöte would understate the
				// record without any way to detect the gap.
				return fmt.Errorf(
					"speeches: %s/%s returned the fetch limit (%d); the riksmöte is truncated and cannot be paged",
					rm, party, SpeechFetchLimit)
			}
			for _, s := range ss {
				if s.Date.After(newest) {
					newest = s.Date
				}
			}
			time.Sleep(200 * time.Millisecond)
		}
	}

	if newest.IsZero() {
		slog.Info("speeches: nothing fetched, cursor unchanged")
		return nil
	}
	return w.cursors.Upsert(ctx, ingPorts.Cursor{DataType: "speeches", LastDate: &newest})
}
```

`s.Date` above stands for whichever field on `domain.Speech` carries Riksdagen's own date — open the struct and use that field. Never use an insert timestamp.

Add `FetchAndStore` to `backend/internal/speeches/service.go`:

```go
// FetchAndStore fetches one party-riksmöte and returns the rows stored, so the
// caller can detect truncation and track the newest record seen.
func (s *Service) FetchAndStore(ctx context.Context, f ports.FetchSpeechesFilter) ([]*domain.Speech, error) {
	ss, err := s.riksdagen.FetchSpeeches(ctx, f)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpsertMany(ctx, ss); err != nil {
		return nil, err
	}
	return ss, nil
}
```

- [ ] **Step 6: Run the tests**

```bash
cd backend && go test ./internal/ingestion/workers/ -run TestSpeechesWorker -v
go build ./...
```

Expected: all four PASS, build clean.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/speeches/ backend/internal/ingestion/workers/speeches.go backend/internal/ingestion/workers/speeches_worker_test.go
git commit -m "fix(speeches): cover the mandate, correct the cursor, fail on truncation"
```

---

### Task 6: Backfill runner and coverage record

Rewritten 2026-08-01 around betänkande partitioning. Roughly **620 requests**: four riksmöten × (about 4 enumeration pages + ~150 betänkanden).

**Files:**
- Create: `backend/migrations/000031_ingestion_coverage.{up,down}.sql`, `backend/cmd/backfill/main.go`

**Interfaces:**
- Consumes: `ListVoteringar` and `FetchAndStore` (Task 4).
- Produces: `ingestion_coverage(riksmote TEXT PRIMARY KEY, expected_voteringar INT, ingested_voteringar INT, last_vote_date DATE, checked_at TIMESTAMPTZ)`.

- [ ] **Step 1: Write the migration**

`000031_ingestion_coverage.up.sql`:

```sql
-- What we actually hold, per riksmöte, so the site can state its own coverage
-- rather than implying completeness. expected_voteringar is Riksdagen's own
-- count (@traffar from /dokumentlista); ingested_voteringar is what we stored.
CREATE TABLE IF NOT EXISTS ingestion_coverage (
  riksmote            TEXT PRIMARY KEY,
  expected_voteringar INTEGER,
  ingested_voteringar INTEGER NOT NULL DEFAULT 0,
  last_vote_date      DATE,
  checked_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

`000031_ingestion_coverage.down.sql`:

```sql
DROP TABLE IF EXISTS ingestion_coverage;
```

- [ ] **Step 2: Write the backfill command**

`backend/cmd/backfill/main.go`:

```go
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/ingestion/workers"
	"riksdagskollen/internal/votes"
	votepg "riksdagskollen/internal/votes/adapters/postgres"
	voterd "riksdagskollen/internal/votes/adapters/riksdagen"
	voteports "riksdagskollen/internal/votes/ports"
)

const enumPageSize = 200

func main() {
	only := flag.String("rm", "", "backfill a single riksmöte, e.g. 2025/26")
	flag.Parse()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	svc := votes.NewService(votepg.NewRepository(pool), voterd.NewClient())

	riksmoten := workers.MandateRiksmoten
	if *only != "" {
		riksmoten = []string{*only}
	}

	for _, rm := range riksmoten {
		// 1. Enumerate every votering. /dokumentlista paginates correctly and
		//    reports @traffar, which is the coverage denominator.
		seen := map[string]bool{}
		var order []string
		expected := 0
		for page := 1; ; page++ {
			refs, total, err := svc.ListVoteringar(ctx, rm, page, enumPageSize)
			if err != nil {
				slog.Error("enumeration failed", "rm", rm, "page", page, "error", err)
				break
			}
			if total > 0 {
				expected = total
			}
			for _, r := range refs {
				if r.Beteckning != "" && !seen[r.Beteckning] {
					seen[r.Beteckning] = true
					order = append(order, r.Beteckning)
				}
			}
			if len(refs) < enumPageSize {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		slog.Info("enumerated", "rm", rm, "expected", expected, "betankanden", len(order))

		// 2. One request per betänkande returns every party's ballots.
		//    Since is deliberately zero: a backfill must not apply the
		//    incremental cutoff, or it re-skips the gap it exists to fill.
		for i, bet := range order {
			vv, err := svc.FetchAndStore(ctx, voteports.FetchVotesFilter{
				Session: rm, Beteckning: bet,
			})
			if err != nil {
				slog.Warn("fetch failed", "rm", rm, "bet", bet, "error", err)
				continue
			}
			slog.Info("stored", "rm", rm, "bet", bet, "rows", len(vv),
				"progress", i+1, "of", len(order))
			time.Sleep(200 * time.Millisecond)
		}

		// 3. Record what we actually hold.
		var ingested int
		var lastDate *time.Time
		if err := pool.QueryRow(ctx, `
			SELECT count(DISTINCT beteckning || ':' || forslagspunkt), max(system_datum)::date
			FROM votes WHERE session = $1`, rm).Scan(&ingested, &lastDate); err != nil {
			slog.Warn("coverage query failed", "rm", rm, "error", err)
			continue
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO ingestion_coverage
				(riksmote, expected_voteringar, ingested_voteringar, last_vote_date, checked_at)
			VALUES ($1, $2, $3, $4, now())
			ON CONFLICT (riksmote) DO UPDATE SET
				expected_voteringar = EXCLUDED.expected_voteringar,
				ingested_voteringar = EXCLUDED.ingested_voteringar,
				last_vote_date      = EXCLUDED.last_vote_date,
				checked_at          = now()`,
			rm, expected, ingested, lastDate); err != nil {
			slog.Warn("coverage upsert failed", "rm", rm, "error", err)
		}
		slog.Info("riksmöte done", "rm", rm, "ingested", ingested, "expected", expected)
	}
}
```

The `max(system_datum)` above assumes the `votes` table persists Riksdagen's own date. **Check first**: `\d votes` in psql. If the column does not exist, add it in a migration and populate it from `domain.Vote.SystemDatum` (Task 3 already carries the value through). Do **not** fall back to `created_at` — presenting an ingest date as a decision date is exactly the dishonesty this plan removes.

The command is safely re-runnable: `UpsertMany` is idempotent, so an interrupted run can simply be restarted.

- [ ] **Step 3: Dry-run one riksmöte**

```bash
make migrate
cd backend && DATABASE_URL='postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable' \
  go run ./cmd/backfill -rm 2025/26
```

Expected: enumeration reports `expected=759`; roughly 150 betänkanden fetched; `ingestion_coverage` gains a row with `ingested_voteringar` close to 759.

- [ ] **Step 4: Full run**

```bash
cd backend && DATABASE_URL='...' go run ./cmd/backfill
psql "$DATABASE_URL" -c "SELECT * FROM ingestion_coverage ORDER BY riksmote;"
```

Expected: four rows totalling roughly 2 562 ingested against 2 562 expected. Investigate any riksmöte more than 2% short before proceeding — a systematic shortfall usually means betänkanden whose votering rows are filed under a different beteckning.

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/000031_* backend/cmd/backfill/
git commit -m "feat(ingestion): backfill the 2022-2026 record and record coverage"
```

### Task 7: Freshness honesty

`liveDateStr()` (`App.tsx:75-81`) returns `new Date()` — today's browser date, with no connection to the data. The header renders `LIVE · <today>` over an eleven-month-old feed.

**Files:**
- Modify: `backend/internal/riksdag/adapters/http/handler.go` (add `GET /api/riksdag/freshness`), `frontend/src/App.tsx:75-81,261`

**Interfaces:**
- Produces: `GET /api/riksdag/freshness` → `{ "latestVoteDate": "2026-06-17", "latestSpeechDate": "2026-06-17" }` (either may be `null`).

- [ ] **Step 1: Add the endpoint**

Query `MAX(systemdatum)::date` from `votes` and from `speeches`, return both. Register the route in `cmd/api/main.go` beside the other `/riksdag/*` routes, and add it to `api/openapi.yaml`, then run `cd frontend && npm run generate:api`.

- [ ] **Step 2: Replace the badge**

In `App.tsx`, delete `liveDateStr()` and render the real date:

```tsx
// The header must never claim currency it does not have: show the date of the
// newest record actually held, not today's date.
{freshness?.latestVoteDate
  ? `Senaste beslut · ${formatSwedishDate(freshness.latestVoteDate)}`
  : "Ingen omröstningsdata"}
```

- [ ] **Step 3: Verify**

```bash
curl -s localhost:8080/api/riksdag/freshness
```

Then load `http://localhost:5173/` and confirm the header shows the newest vote date, **not** today's date. With the backfill applied this should read `Senaste beslut · 17 juni 2026`.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/riksdag/ backend/cmd/api/main.go api/openapi.yaml frontend/src/shared/api-contract.ts frontend/src/App.tsx
git commit -m "fix(ui): show real data freshness instead of today's date"
```

---

### Task 8: Coverage at the point of use

**Files:**
- Modify: `backend/internal/riksdag/adapters/http/handler.go` (add `GET /api/riksdag/coverage`), `frontend/src/features/parties/PartyGoalsPage.tsx`

**Interfaces:**
- Consumes: `ingestion_coverage` from Task 6.
- Produces: `GET /api/riksdag/coverage` → `{ "mandate": "2022-2026", "expected": 2562, "ingested": 2562, "byRiksmote": [{ "riksmote": "2022/23", "expected": 562, "ingested": 562 }] }`

- [ ] **Step 1: Add the endpoint**

Sum `expected_voteringar` and `ingested_voteringar` across `ingestion_coverage`, plus the per-riksmöte breakdown. Add to `api/openapi.yaml`; regenerate the contract.

- [ ] **Step 2: Render it on the record view**

Above the goals list on `PartyGoalsPage`:

```tsx
// Completeness is a claim like any other, so it must be checkable. If the
// backfill is short, the reader sees the shortfall rather than silence.
<p className="text-xs text-on-surface-variant">
  Baserat på {coverage.ingested.toLocaleString("sv-SE")} av{" "}
  {coverage.expected.toLocaleString("sv-SE")} omröstningar 2022–2026
  <SourceMarker sourceId="riksdagen" />
  {coverage.ingested < coverage.expected && (
    <Link to="/data" className="ml-2 underline">Vad saknas?</Link>
  )}
</p>
```

- [ ] **Step 3: Verify**

```bash
curl -s localhost:8080/api/riksdag/coverage | python3 -m json.tool
```

Load `/parties/S/goals` and confirm the line renders with real numbers. Temporarily delete one riksmöte's rows and confirm the shortfall and the "Vad saknas?" link appear, then restore.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/riksdag/ api/openapi.yaml frontend/src/shared/api-contract.ts frontend/src/features/parties/PartyGoalsPage.tsx
git commit -m "feat(ui): state record coverage where the record is used"
```

---

### Task 9: Mandate scoping and the 14 September relabel

Without this, votes from the incoming parliament flow into the outgoing one's record and misattribute it.

**Files:**
- Create: `backend/migrations/000032_mandate_periods.{up,down}.sql`
- Modify: `backend/internal/riksdag/adapters/http/handler.go`, `frontend/src/features/parties/PartyGoalsPage.tsx`

**Interfaces:**
- Produces: table `mandate_periods(code TEXT PRIMARY KEY, label TEXT, start_date DATE, end_date DATE, riksmoten TEXT[])`, seeded with `('2022-2026', 'Mandatperioden 2022–2026', '2022-09-11', '2026-09-13', ARRAY['2022/23','2023/24','2024/25','2025/26'])`. Exposed on the coverage endpoint as `mandateLabel` and `mandateEnded bool`.

- [ ] **Step 1: Write the migration**

```sql
CREATE TABLE IF NOT EXISTS mandate_periods (
  code       TEXT PRIMARY KEY,
  label      TEXT NOT NULL,
  start_date DATE NOT NULL,
  end_date   DATE NOT NULL,
  riksmoten  TEXT[] NOT NULL
);

INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
VALUES ('2022-2026', 'Mandatperioden 2022–2026', '2022-09-11', '2026-09-13',
        ARRAY['2022/23','2023/24','2024/25','2025/26'])
ON CONFLICT (code) DO NOTHING;
```

Down: `DROP TABLE IF EXISTS mandate_periods;`

- [ ] **Step 2: Expose label and ended-state**

Extend the coverage handler to read the mandate row and return `mandateLabel` plus `mandateEnded` (`end_date < CURRENT_DATE`). Update `api/openapi.yaml`; regenerate.

- [ ] **Step 3: Render the scope**

Replace the record heading so it always names the period, and appends `(avslutad)` once ended:

```tsx
<h2>{coverage.mandateLabel}{coverage.mandateEnded ? " (avslutad)" : ""}</h2>
```

- [ ] **Step 4: Verify the boundary**

```bash
psql "$DATABASE_URL" -c "UPDATE mandate_periods SET end_date = '2026-07-01' WHERE code = '2022-2026';"
curl -s localhost:8080/api/riksdag/coverage | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['mandateLabel'], d['mandateEnded'])"
```

Expected: `Mandatperioden 2022–2026 True`, and the heading shows `(avslutad)`. Restore:

```bash
psql "$DATABASE_URL" -c "UPDATE mandate_periods SET end_date = '2026-09-13' WHERE code = '2022-2026';"
```

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/000032_* backend/internal/riksdag/ api/openapi.yaml frontend/src/shared/api-contract.ts frontend/src/features/parties/PartyGoalsPage.tsx
git commit -m "feat: scope the record to a mandate period and relabel when it ends"
```

---

## Final verification

```bash
go build -C backend ./...
cd backend && TEST_DATABASE_URL='postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable' go test ./...
cd frontend && npx tsc --noEmit
```

Then against the running stack:

- No party shows any 2026 source document
- `/api/riksdag/coverage` reports ~2 562 of 2 562
- The header shows `Senaste beslut · 17 juni 2026`, not today's date
- `MJU12` and `KrU5` resolve to committee names in the UI
- A goal page states its coverage above the goals list
