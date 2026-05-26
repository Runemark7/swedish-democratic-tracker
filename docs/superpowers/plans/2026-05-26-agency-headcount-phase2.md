# Agency Headcount + Department — Phase 2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enrich the `authorities` table (Phase 1) with **headcount (årsarbetskrafter)** and **department** per agency, joined on `org_number`, sourced from the Statskontoret Myndighetsförteckning Excel (CC0, annual, 2007–2025). Adds a yearly history JSONB.

**Architecture:** New `myndighetsforteckning` adapter (Excel parser via `github.com/xuri/excelize/v2`) + a separate `authority-headcount` ingestion worker that runs `@weekly` after the register worker. The repository gains a non-destructive `UpdateEnrichment` method that only touches the new columns — Phase 1 register data is never wiped.

**Tech Stack:** Go 1.26.1, pgx/v5 raw SQL, `github.com/xuri/excelize/v2` (new dep). Spec: `docs/superpowers/specs/2026-05-25-myndigheter-register-pipeline-design.md`.

**Order change rationale:** The spec listed expenditure as Phase 2 and headcount as Phase 3. Reordered after planning-time verification: Statskontoret's per-agency Myndighetsförteckning is one clean Excel covering headcount + department + governance, while per-agency *expenditure* needs an additional `anslag → myndighet` mapping (Statsliggaren) — more work, deferred to Phase 3.

**Out of scope (Phase 3):** ESV anslag-based expenditure + Statsliggaren mapping. **Phase 4:** card switches to DB top-N (by headcount once expenditure is missing for many), new searchable list page, openapi + data-source docs.

---

### Task 1: Add `excelize/v2` dependency

**Files:**
- Modify: `backend/go.mod`, `backend/go.sum`

- [ ] **Step 1: Add the dep**

Run from repo root:
```bash
cd backend && go get github.com/xuri/excelize/v2@latest && cd ..
go mod tidy -C backend
```

- [ ] **Step 2: Verify it builds**

Run: `go build -C backend ./...`
Expected: success (the dep is downloaded; nothing imports it yet — that's fine, `go mod tidy` will keep it as a `go get`-pinned requirement until something imports it; if `tidy` removes it, that's OK too — Task 4 re-introduces it via import).

Actually safer to skip `tidy` here and let Task 4's import drive the dep. Run only `go get`:
```bash
cd backend && go get github.com/xuri/excelize/v2@latest && cd ..
go build -C backend ./...
```

- [ ] **Step 3: Commit**

```bash
git add backend/go.mod backend/go.sum
git commit -m "build(backend): add excelize/v2 for xlsx parsing"
```

---

### Task 2: Extend ports — `HeadcountEntry`, `HeadcountClient`, `UpdateEnrichment`

**Files:**
- Modify: `backend/internal/riksdag/ports/register.go`

- [ ] **Step 1: Append to the existing file**

After the existing `AuthorityRepository` interface in `backend/internal/riksdag/ports/register.go`, append:
```go

// YearlyHeadcountSCB is one year's headcount for an agency from
// Statskontoret's Myndighetsförteckning.
type YearlyHeadcountSCB struct {
	Year         int
	HeadcountInt int
}

// HeadcountEntry is one agency's enrichment record from Myndighetsförteckning.
type HeadcountEntry struct {
	OrgNumber  string // canonical "NNNNNN-NNNN"
	Name       string // for logging / unmatched-row diagnostics
	Department string
	Year       int                  // latest year
	Headcount  int                  // latest value
	History    []YearlyHeadcountSCB // chronological, ascending
}

type HeadcountClient interface {
	FetchHeadcounts(ctx context.Context) ([]HeadcountEntry, error)
}

// Enrichment is the update payload for non-register columns. Phase 1's
// UpsertAuthorities never writes these; this method only updates rows whose
// org_number already exists (no INSERT — pre-existing register row required).
type Enrichment struct {
	OrgNumber    string
	Department   string
	HeadcountInt int
	Year         int
	History      []YearlyHeadcountSCB
}
```

And extend the `AuthorityRepository` interface in the same file by adding one method line. Locate:
```go
type AuthorityRepository interface {
	UpsertAuthorities(ctx context.Context, items []domain.RegisteredAuthority) (int, error)
	List(ctx context.Context, f AuthorityFilter) ([]domain.RegisteredAuthority, error)
	Count(ctx context.Context, f AuthorityFilter) (int, error)
}
```
Add a fourth line so it becomes:
```go
type AuthorityRepository interface {
	UpsertAuthorities(ctx context.Context, items []domain.RegisteredAuthority) (int, error)
	List(ctx context.Context, f AuthorityFilter) ([]domain.RegisteredAuthority, error)
	Count(ctx context.Context, f AuthorityFilter) (int, error)
	UpdateEnrichment(ctx context.Context, items []Enrichment) (matched, unmatched int, err error)
}
```

- [ ] **Step 2: Verify build (will fail in postgres until Task 3 implements `UpdateEnrichment`)**

Run: `go build -C backend ./internal/riksdag/ports/...`
Expected: success (ports is an interface declaration — implementations come later).

`go build -C backend ./...` will FAIL at this point because `*postgres.AuthorityRepository` no longer satisfies the interface. That is expected; Task 3 fixes it. Do NOT commit yet — combine ports + repo in one commit at the end of Task 3 to keep the tree compilable.

---

### Task 3: Implement `UpdateEnrichment` in postgres + commit Tasks 2+3 together

**Files:**
- Modify: `backend/internal/riksdag/adapters/postgres/authority_repository.go`

- [ ] **Step 1: Add the method to the repository**

At the bottom of `backend/internal/riksdag/adapters/postgres/authority_repository.go`, append (do not remove anything):
```go

func (r *AuthorityRepository) UpdateEnrichment(ctx context.Context, items []ports.Enrichment) (int, int, error) {
	matched, unmatched := 0, 0
	for _, e := range items {
		historyJSON, err := json.Marshal(e.History)
		if err != nil {
			return matched, unmatched, err
		}
		tag, err := r.pool.Exec(ctx, `
			UPDATE authorities
			SET department        = CASE WHEN $2 <> '' THEN $2 ELSE department END,
			    headcount_int     = $3,
			    year              = GREATEST(year, $4),
			    headcount_history = $5,
			    updated_at        = now()
			WHERE org_number = $1
		`, e.OrgNumber, e.Department, e.HeadcountInt, e.Year, historyJSON)
		if err != nil {
			return matched, unmatched, err
		}
		if tag.RowsAffected() == 1 {
			matched++
		} else {
			unmatched++
		}
	}
	return matched, unmatched, nil
}
```

Add `"encoding/json"` to the import block at the top of the same file (alongside `"context"`, `"strconv"`, `"strings"`).

- [ ] **Step 2: Verify the full build is green**

Run: `go build -C backend ./...`
Expected: success. `*AuthorityRepository` now satisfies the extended interface.

- [ ] **Step 3: Commit Tasks 2 + 3 together (single compilable commit)**

```bash
git add backend/internal/riksdag/ports/register.go \
        backend/internal/riksdag/adapters/postgres/authority_repository.go
git commit -m "feat(riksdag): UpdateEnrichment + HeadcountClient/Entry ports"
```

---

### Task 4: Myndighetsförteckning Excel client (TDD on the normalizer)

**Files:**
- Create: `backend/internal/riksdag/adapters/myndighetsforteckning/client.go`
- Create: `backend/internal/riksdag/adapters/myndighetsforteckning/client_test.go`

**About the data (verified 2026-05-26):**
The Excel `statskontorets-myndighetsforteckning-<year>.xlsx` (CC0, ~600 KB) has multiple sheets; the longitudinal one (`Förteckning 2007-2025`) has one row per agency-year with columns including `orgnr` (10-digit no-dash, e.g. `2021004284`), `myndighet` (name), `år` (year), `årsarbetskrafter` (headcount as a number, possibly empty), `departement`. The current-year sheet (`Förteckning 2025`) holds the latest snapshot.

- [ ] **Step 1: Write the failing test (org-number normalizer is pure logic, easy to TDD)**

`backend/internal/riksdag/adapters/myndighetsforteckning/client_test.go`:
```go
package myndighetsforteckning

import "testing"

func TestNormalizeOrgNumber(t *testing.T) {
	cases := map[string]string{
		"2021004284":   "202100-4284", // 10-digit flat → canonical
		"202100-4284":  "202100-4284", // already canonical
		"857209-0606": "857209-0606", // already canonical (AP-fond style)
		"8572090606":   "857209-0606", // 10-digit flat AP-fond
		" 2021004284 ": "202100-4284", // whitespace tolerated
		"":             "",            // empty stays empty
		"abc":          "",            // garbage rejected
	}
	for in, want := range cases {
		if got := normalizeOrgNumber(in); got != want {
			t.Errorf("normalizeOrgNumber(%q) = %q, want %q", in, got, want)
		}
	}
}
```

- [ ] **Step 2: Run, verify it FAILS**

Run: `go test -C backend ./internal/riksdag/adapters/myndighetsforteckning/ -run TestNormalizeOrgNumber -v`
Expected: FAIL — undefined `normalizeOrgNumber`.

- [ ] **Step 3: Write the client**

`backend/internal/riksdag/adapters/myndighetsforteckning/client.go`:
```go
package myndighetsforteckning

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xuri/excelize/v2"

	"riksdagskollen/internal/riksdag/ports"
)

// Latest year's published file. Update this annually (or derive from the
// landing page if dynamic discovery becomes worthwhile). v= query string is
// the page's cache-buster — harmless if it ages, the server ignores unknown
// version strings.
const datasetURL = "https://www.statskontoret.se/contentassets/bbd19bc969054c86bbc1aa757e7d5c85/statskontorets-myndighetsforteckning-2025.xlsx"

// Sheet that contains the longitudinal "one row per agency-year" data.
const sheetName = "Förteckning 2007-2025"

var (
	digitsRe   = regexp.MustCompile(`^\d{10}$`)
	canonicalRe = regexp.MustCompile(`^\d{6}-\d{4}$`)
)

// normalizeOrgNumber accepts the two forms SCB uses (flat 10-digit and
// canonical NNNNNN-NNNN) and returns the canonical form. Empty/invalid → "".
func normalizeOrgNumber(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if canonicalRe.MatchString(s) {
		return s
	}
	if digitsRe.MatchString(s) {
		return s[:6] + "-" + s[6:]
	}
	return ""
}

type Client struct {
	http     *http.Client
	mu       sync.Mutex
	cached   []ports.HeadcountEntry
	cachedAt time.Time
	cacheTTL time.Duration
}

func NewClient() *Client {
	return &Client{
		http:     &http.Client{Timeout: 60 * time.Second},
		cacheTTL: 24 * time.Hour,
	}
}

func (c *Client) FetchHeadcounts(ctx context.Context) ([]ports.HeadcountEntry, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.cached) > 0 && time.Since(c.cachedAt) < c.cacheTTL {
		return c.cached, nil
	}
	raw, err := c.download(ctx)
	if err != nil {
		return nil, err
	}
	entries, err := parseWorkbook(raw)
	if err != nil {
		return nil, err
	}
	c.cached = entries
	c.cachedAt = time.Now()
	return entries, nil
}

func (c *Client) download(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, datasetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "riksdagskollen/1.0")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("myndighetsforteckning download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("myndighetsforteckning download: HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// parseWorkbook reads the longitudinal sheet and folds rows into one
// HeadcountEntry per org-number with a chronological history. The latest year
// with a non-empty headcount becomes the snapshot fields.
func parseWorkbook(raw []byte) ([]ports.HeadcountEntry, error) {
	f, err := excelize.OpenReader(strings.NewReader(string(raw)))
	if err != nil {
		return nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("read sheet %q: %w", sheetName, err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("sheet %q has %d rows, expected >1", sheetName, len(rows))
	}

	// Resolve column indexes from the header row by name (the workbook's
	// column order has shifted across years; bind by name to stay robust).
	header := rows[0]
	idx := func(name string) int {
		for i, h := range header {
			if strings.EqualFold(strings.TrimSpace(h), name) {
				return i
			}
		}
		return -1
	}
	iOrg := idx("orgnr")
	iName := idx("myndighet")
	iDept := idx("departement")
	iYear := idx("år")
	iFTE := idx("årsarbetskrafter")
	if iOrg < 0 || iName < 0 || iYear < 0 || iFTE < 0 {
		return nil, fmt.Errorf("missing required columns; header=%v", header)
	}

	type acc struct {
		name    string
		dept    string
		history []ports.YearlyHeadcountSCB
	}
	byOrg := map[string]*acc{}

	cell := func(row []string, i int) string {
		if i < 0 || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	for _, row := range rows[1:] {
		org := normalizeOrgNumber(cell(row, iOrg))
		if org == "" {
			continue
		}
		year, err := strconv.Atoi(cell(row, iYear))
		if err != nil {
			continue
		}
		fte := 0
		if v := cell(row, iFTE); v != "" {
			if n, err := strconv.Atoi(strings.ReplaceAll(v, " ", "")); err == nil {
				fte = n
			} else if f, err := strconv.ParseFloat(strings.ReplaceAll(v, ",", "."), 64); err == nil {
				fte = int(f + 0.5)
			}
		}

		a := byOrg[org]
		if a == nil {
			a = &acc{}
			byOrg[org] = a
		}
		if n := cell(row, iName); n != "" {
			a.name = n
		}
		if iDept >= 0 {
			if d := cell(row, iDept); d != "" {
				a.dept = d // last non-empty wins (latest sample tends to be last)
			}
		}
		a.history = append(a.history, ports.YearlyHeadcountSCB{Year: year, HeadcountInt: fte})
	}

	out := make([]ports.HeadcountEntry, 0, len(byOrg))
	for org, a := range byOrg {
		// Sort history ascending and pick the latest non-zero as snapshot.
		sortAscByYear(a.history)
		snapYear, snapFTE := latestNonZero(a.history)
		out = append(out, ports.HeadcountEntry{
			OrgNumber:  org,
			Name:       a.name,
			Department: a.dept,
			Year:       snapYear,
			Headcount:  snapFTE,
			History:    a.history,
		})
	}
	return out, nil
}

func sortAscByYear(h []ports.YearlyHeadcountSCB) {
	for i := 1; i < len(h); i++ {
		for j := i; j > 0 && h[j-1].Year > h[j].Year; j-- {
			h[j-1], h[j] = h[j], h[j-1]
		}
	}
}

func latestNonZero(h []ports.YearlyHeadcountSCB) (int, int) {
	for i := len(h) - 1; i >= 0; i-- {
		if h[i].HeadcountInt > 0 {
			return h[i].Year, h[i].HeadcountInt
		}
	}
	if len(h) > 0 {
		return h[len(h)-1].Year, 0
	}
	return 0, 0
}
```

- [ ] **Step 4: Run the test, verify it PASSES**

Run: `go test -C backend ./internal/riksdag/adapters/myndighetsforteckning/ -v`
Expected: `TestNormalizeOrgNumber` PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/riksdag/adapters/myndighetsforteckning/ backend/go.sum
git commit -m "feat(riksdag): Myndighetsforteckning xlsx client (headcount + dept)"
```

---

### Task 5: Authority headcount worker (TDD)

**Files:**
- Create: `backend/internal/ingestion/workers/authority_headcount.go`
- Create: `backend/internal/ingestion/workers/authority_headcount_test.go`

- [ ] **Step 1: Write the failing test**

`backend/internal/ingestion/workers/authority_headcount_test.go`:
```go
package workers_test

import (
	"context"
	"errors"
	"testing"

	"riksdagskollen/internal/ingestion/workers"
	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
)

type mockHeadcountClient struct {
	entries []ports.HeadcountEntry
	err     error
}

func (m *mockHeadcountClient) FetchHeadcounts(_ context.Context) ([]ports.HeadcountEntry, error) {
	return m.entries, m.err
}

type mockEnrichRepo struct {
	captured []ports.Enrichment
	matched  int
	updateErr error
}

func (m *mockEnrichRepo) UpsertAuthorities(_ context.Context, _ []domain.RegisteredAuthority) (int, error) {
	return 0, nil
}
func (m *mockEnrichRepo) List(_ context.Context, _ ports.AuthorityFilter) ([]domain.RegisteredAuthority, error) {
	return nil, nil
}
func (m *mockEnrichRepo) Count(_ context.Context, _ ports.AuthorityFilter) (int, error) {
	return 0, nil
}
func (m *mockEnrichRepo) UpdateEnrichment(_ context.Context, items []ports.Enrichment) (int, int, error) {
	if m.updateErr != nil {
		return 0, 0, m.updateErr
	}
	m.captured = append(m.captured, items...)
	return m.matched, len(items) - m.matched, nil
}

func TestAuthorityHeadcountWorker_MapsAndUpdates(t *testing.T) {
	client := &mockHeadcountClient{entries: []ports.HeadcountEntry{
		{OrgNumber: "202100-2114", Name: "Arbetsförmedlingen", Department: "Arbetsmarknadsdepartementet", Year: 2024, Headcount: 9400, History: []ports.YearlyHeadcountSCB{{2023, 9300}, {2024, 9400}}},
	}}
	repo := &mockEnrichRepo{matched: 1}
	w := workers.NewAuthorityHeadcountWorker(client, repo)

	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.captured) != 1 {
		t.Fatalf("expected 1 enrichment, got %d", len(repo.captured))
	}
	got := repo.captured[0]
	if got.OrgNumber != "202100-2114" || got.HeadcountInt != 9400 || got.Department != "Arbetsmarknadsdepartementet" || got.Year != 2024 {
		t.Errorf("payload mapped wrong: %+v", got)
	}
	if len(got.History) != 2 {
		t.Errorf("history not propagated: %+v", got.History)
	}
}

func TestAuthorityHeadcountWorker_FetchErrorNoUpdate(t *testing.T) {
	client := &mockHeadcountClient{err: errors.New("xlsx down")}
	repo := &mockEnrichRepo{}
	w := workers.NewAuthorityHeadcountWorker(client, repo)

	if err := w.Run(context.Background()); err == nil {
		t.Error("expected error when fetch fails")
	}
	if len(repo.captured) != 0 {
		t.Error("must not update on fetch failure (keep last good data)")
	}
}
```

- [ ] **Step 2: Run, verify it FAILS**

Run: `go test -C backend ./internal/ingestion/workers/ -run TestAuthorityHeadcountWorker -v`
Expected: undefined `workers.NewAuthorityHeadcountWorker`.

- [ ] **Step 3: Write the worker**

`backend/internal/ingestion/workers/authority_headcount.go`:
```go
package workers

import (
	"context"
	"log/slog"

	"riksdagskollen/internal/riksdag/ports"
)

type AuthorityHeadcountWorker struct {
	client ports.HeadcountClient
	repo   ports.AuthorityRepository
}

func NewAuthorityHeadcountWorker(client ports.HeadcountClient, repo ports.AuthorityRepository) AuthorityHeadcountWorker {
	return AuthorityHeadcountWorker{client: client, repo: repo}
}

func (w *AuthorityHeadcountWorker) Name() string { return "authority-headcount" }

func (w *AuthorityHeadcountWorker) Run(ctx context.Context) error {
	entries, err := w.client.FetchHeadcounts(ctx)
	if err != nil {
		// Keep last good enrichment — do not touch the table on fetch failure.
		return err
	}

	items := make([]ports.Enrichment, 0, len(entries))
	for _, e := range entries {
		items = append(items, ports.Enrichment{
			OrgNumber:    e.OrgNumber,
			Department:   e.Department,
			HeadcountInt: e.Headcount,
			Year:         e.Year,
			History:      e.History,
		})
	}

	matched, unmatched, err := w.repo.UpdateEnrichment(ctx, items)
	if err != nil {
		return err
	}
	slog.Info("authority-headcount: enrichment applied", "matched", matched, "unmatched", unmatched)
	return nil
}
```

- [ ] **Step 4: Run, verify PASSES**

Run: `go test -C backend ./internal/ingestion/workers/ -run TestAuthorityHeadcountWorker -v`
Expected: both cases PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/ingestion/workers/authority_headcount.go backend/internal/ingestion/workers/authority_headcount_test.go
git commit -m "feat(ingestion): authority-headcount worker enriches via Myndighetsforteckning"
```

---

### Task 6: Wire into main.go

**Files:**
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: Import the new adapter**

Add to the import block (alongside `riksdagRegistret`):
```go
riksdagMF "riksdagskollen/internal/riksdag/adapters/myndighetsforteckning"
```

- [ ] **Step 2: Construct the worker**

Immediately after the existing line:
```go
authoritiesWorker := workers.NewAuthoritiesWorker(riksdagRegistret.NewClient(), authorityRepo)
```
add:
```go
headcountWorker := workers.NewAuthorityHeadcountWorker(riksdagMF.NewClient(), authorityRepo)
```

- [ ] **Step 3: Register on the scheduler (after the authorities-worker block)**

Immediately after the existing `RegisterSync("@weekly", &authoritiesWorker)` block, add:
```go
if err := sched.RegisterSync("@weekly", &headcountWorker); err != nil {
    slog.Error("failed to register authority-headcount worker", "error", err)
}
```

- [ ] **Step 4: Verify full build**

Run: `go build -C backend ./...`
Expected: success.

- [ ] **Step 5: Commit**

```bash
git add backend/cmd/api/main.go
git commit -m "feat(riksdag): wire authority-headcount worker on weekly schedule"
```

---

### Task 7: End-to-end verification

**Files:** none (verification only).

- [ ] **Step 1: Create a one-shot runner**

Create `backend/cmd/runenrich/main.go` (will be deleted in Step 4):
```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/ingestion/workers"
	riksdagMF "riksdagskollen/internal/riksdag/adapters/myndighetsforteckning"
	riksdagPG "riksdagskollen/internal/riksdag/adapters/postgres"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil { fmt.Fprintln(os.Stderr, "db:", err); os.Exit(1) }
	defer pool.Close()
	repo := riksdagPG.NewAuthorityRepository(pool)
	w := workers.NewAuthorityHeadcountWorker(riksdagMF.NewClient(), repo)
	if err := w.Run(ctx); err != nil { fmt.Fprintln(os.Stderr, "worker:", err); os.Exit(1) }
	fmt.Println("OK")
}
```

- [ ] **Step 2: Run it**

```bash
go run -C backend ./cmd/runenrich/
```
Expected: log line `authority-headcount: enrichment applied matched=<N> unmatched=<M>`. `matched` should be in the hundreds (most förvaltningsmyndigheter + some others have headcount data).

- [ ] **Step 3: Verify the DB**

```bash
docker compose -f docker-compose.dev.yml exec -T postgres \
  psql -U riksdagskollen -d riksdagskollen \
  -c "SELECT count(*) AS rows_with_headcount FROM authorities WHERE headcount_int IS NOT NULL;
      SELECT count(*) AS rows_with_department FROM authorities WHERE department <> '';
      SELECT name, headcount_int, department, year FROM authorities WHERE name ILIKE '%polismyn%' OR name ILIKE '%forsakringskass%' OR name ILIKE '%arbetsformedling%' OR name ILIKE '%skattever%';"
```
Expected: hundreds of rows have headcount + department populated; the four major agencies show realistic FTE numbers and the right department.

- [ ] **Step 4: Delete the runner + commit nothing**

```bash
rm -rf backend/cmd/runenrich
git status   # should be clean apart from any unrelated working-tree edits
```

- [ ] **Step 5: Confirm no UI regression**

Open `http://localhost:5173/riksdag`. The MYNDIGHETER card + detail pages must look exactly as before (still live-fetch; Phase 4 will switch to DB reads).

---

## Self-Review

**Spec coverage (Phase 2 slice):** headcount per agency ✓ Tasks 4–6; org-number canonical join ✓ Task 4 normalizer; best-effort / keep-last-good ✓ Tasks 3 (UPDATE-only) + 5 (return early on fetch err); department backfill ✓ Task 3 (CASE-WHEN preserves register department if Excel is blank); JSONB history ✓ Task 3.

**Placeholder scan:** No "TBD"/"TODO". Task 7 includes a fully-spelled-out runner with all imports/fields used and is explicitly deleted in Step 4.

**Type consistency:** `ports.HeadcountEntry`, `ports.HeadcountClient`, `ports.Enrichment`, `ports.YearlyHeadcountSCB` introduced in Task 2, used identically in Tasks 3–5. `*AuthorityRepository.UpdateEnrichment` signature `(ctx, []ports.Enrichment) (int, int, error)` matches the interface and the worker's call site. `NewAuthorityHeadcountWorker(client, repo)` matches main.go in Task 6.
