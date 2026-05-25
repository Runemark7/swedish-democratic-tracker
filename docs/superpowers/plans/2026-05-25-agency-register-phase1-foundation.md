# Agency Register — Phase 1: Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Scrape the full SCB Myndighetsregistret (~449 agencies across 6 groups) into a new `authorities` DB table via a weekly ingestion worker — the data foundation for later expenditure/headcount enrichment and the searchable list page.

**Architecture:** New, isolated backend slice in `backend/internal/riksdag` + `backend/internal/ingestion`. A `registret` adapter POSTs to `myndighetsregistret.scb.se/Myndighet/HamtaMynd` per group, parses the returned HTML table, and a worker upserts rows into `authorities` keyed by `org_number`. **Nothing reads this table yet** — the existing live-fetch card and detail page are untouched, so there is zero UI regression. Reads + UI come in later phases.

**Tech Stack:** Go 1.26.1, chi, pgx/v5 (raw SQL), golang-migrate, robfig/cron. Stdlib only for parsing (`regexp`, `html`, `net/http`). Spec: `docs/superpowers/specs/2026-05-25-myndigheter-register-pipeline-design.md`.

**Scope (Phase 1 only):** register scrape → DB. Out of scope (later phases): ESV expenditure, Arbetsgivarverket headcount, the `/riksdag/myndigheter` list page, switching the card to read the DB.

---

### Task 1: Migration — `authorities` table

**Files:**
- Create: `backend/migrations/000020_authorities.up.sql`
- Create: `backend/migrations/000020_authorities.down.sql`

- [ ] **Step 1: Write the up migration**

`backend/migrations/000020_authorities.up.sql`:
```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE authorities (
    org_number          TEXT PRIMARY KEY,
    slug                TEXT NOT NULL UNIQUE,
    name                TEXT NOT NULL,
    type                TEXT NOT NULL DEFAULT '',
    principal_body      TEXT NOT NULL DEFAULT '',
    department          TEXT NOT NULL DEFAULT '',
    under_government    BOOLEAN NOT NULL DEFAULT FALSE,
    website             TEXT NOT NULL DEFAULT '',
    sfs                 TEXT NOT NULL DEFAULT '',
    expenditure_mdkr    DOUBLE PRECISION,
    budget_mdkr         DOUBLE PRECISION,
    headcount_int       INTEGER,
    year                INTEGER NOT NULL DEFAULT 0,
    expenditure_history JSONB NOT NULL DEFAULT '[]',
    headcount_history   JSONB NOT NULL DEFAULT '[]',
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_authorities_under_government ON authorities (under_government);
CREATE INDEX idx_authorities_expenditure ON authorities (expenditure_mdkr DESC NULLS LAST);
CREATE INDEX idx_authorities_name_trgm ON authorities USING gin (name gin_trgm_ops);
```

- [ ] **Step 2: Write the down migration**

`backend/migrations/000020_authorities.down.sql`:
```sql
DROP TABLE IF EXISTS authorities;
```

- [ ] **Step 3: Apply and verify**

Run: `make migrate`
Expected: migration `000020` applies clean. Verify:
```bash
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U riksdagskollen -d riksdagskollen -c '\d authorities'
```
Expected: table `authorities` with the columns above and three indexes.

- [ ] **Step 4: Commit**

```bash
git add backend/migrations/000020_authorities.up.sql backend/migrations/000020_authorities.down.sql
git commit -m "feat(riksdag): add authorities table migration"
```

---

### Task 2: Ports + domain types

**Files:**
- Create: `backend/internal/riksdag/ports/register.go`
- Create: `backend/internal/riksdag/domain/registered_authority.go`

- [ ] **Step 1: Write the domain type**

`backend/internal/riksdag/domain/registered_authority.go`:
```go
package domain

import "time"

// RegisteredAuthority is a state agency as stored in the authorities table.
// Expenditure/headcount are pointers: nil means "no source data yet" (rendered
// as "saknas"), distinct from a real zero.
type RegisteredAuthority struct {
	OrgNumber       string
	Slug            string
	Name            string
	Type            string
	PrincipalBody   string
	Department      string
	UnderGovernment bool
	Website         string
	SFS             string
	ExpenditureMdkr *float64
	BudgetMdkr      *float64
	HeadcountInt    *int
	Year            int
	UpdatedAt       time.Time
}
```

- [ ] **Step 2: Write the ports**

`backend/internal/riksdag/ports/register.go`:
```go
package ports

import (
	"context"

	"riksdagskollen/internal/riksdag/domain"
)

// RegisterEntry is one row scraped from SCB Myndighetsregistret.
type RegisterEntry struct {
	OrgNumber       string
	Slug            string
	Name            string
	Type            string
	PrincipalBody   string
	UnderGovernment bool
	Website         string
	SFS             string
}

type RegisterClient interface {
	FetchRegister(ctx context.Context) ([]RegisterEntry, error)
}

type AuthorityFilter struct {
	Query           string
	UnderGovernment *bool
	Limit           int
	Offset          int
}

type AuthorityRepository interface {
	UpsertAuthorities(ctx context.Context, items []domain.RegisteredAuthority) (int, error)
	List(ctx context.Context, f AuthorityFilter) ([]domain.RegisteredAuthority, error)
	Count(ctx context.Context, f AuthorityFilter) (int, error)
}
```

- [ ] **Step 3: Verify it compiles**

Run: `go build -C backend ./internal/riksdag/...`
Expected: no output (success).

- [ ] **Step 4: Commit**

```bash
git add backend/internal/riksdag/ports/register.go backend/internal/riksdag/domain/registered_authority.go
git commit -m "feat(riksdag): register ports + RegisteredAuthority domain type"
```

---

### Task 3: Register scraper — HTML table parser (TDD)

**Files:**
- Create: `backend/internal/riksdag/adapters/registret/client.go`
- Create: `backend/internal/riksdag/adapters/registret/client_test.go`

- [ ] **Step 1: Write the failing test**

`backend/internal/riksdag/adapters/registret/client_test.go`:
```go
package registret

import "testing"

const sampleFragment = `
<table id="ResultTable"><thead><tr><th>Namn</th><th>Organisationsnr</th><th>SFS</th><th>WebbAdress</th></tr></thead>
<tbody>
<tr><td></td><td></td><td></td><td></td></tr>
<tr><td>Alkoholsortimentsn&#228;mnden</td><td>202100-5943</td><td>2007:1216</td><td>www.kammarkollegiet.se/alkoholsortimentsnamnden</td></tr>
<tr><td>Arbetsf&#246;rmedlingen</td><td>202100-2114</td><td>2007:1030</td><td>www.arbetsformedlingen.se</td></tr>
</tbody></table>`

func TestParseAgencyTable(t *testing.T) {
	g := group{typ: "Förvaltningsmyndighet", principal: "Regeringen", underGov: true}
	got := parseAgencyTable(sampleFragment, g)

	if len(got) != 2 {
		t.Fatalf("expected 2 agencies, got %d", len(got))
	}
	first := got[0]
	if first.Name != "Alkoholsortimentsnämnden" {
		t.Errorf("name: got %q", first.Name)
	}
	if first.OrgNumber != "202100-5943" {
		t.Errorf("org: got %q", first.OrgNumber)
	}
	if first.SFS != "2007:1216" {
		t.Errorf("sfs: got %q", first.SFS)
	}
	if first.Website != "www.kammarkollegiet.se/alkoholsortimentsnamnden" {
		t.Errorf("website: got %q", first.Website)
	}
	if first.Slug != "alkoholsortimentsnamnden" {
		t.Errorf("slug: got %q", first.Slug)
	}
	if !first.UnderGovernment || first.Type != "Förvaltningsmyndighet" {
		t.Errorf("group mapping wrong: %+v", first)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Åklagarmyndigheten":  "aklagarmyndigheten",
		"Sveriges Domstolar":  "sveriges-domstolar",
		"Försäkringskassan":   "forsakringskassan",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test -C backend ./internal/riksdag/adapters/registret/ -run TestParseAgencyTable -v`
Expected: FAIL — `undefined: parseAgencyTable` / `group` / `slugify`.

- [ ] **Step 3: Write the parser + helpers**

`backend/internal/riksdag/adapters/registret/client.go` (parser only — the HTTP
`Client` is added in Task 4; every import here is used by the parser/test):
```go
package registret

import (
	"html"
	"regexp"
	"strings"
	"unicode"

	"riksdagskollen/internal/riksdag/ports"
)

// group is a Myndighetsregistret category. The group label is sent verbatim as
// the HamtaMynd "mynd" parameter; typ/principal/underGov classify the rows.
type group struct {
	label     string
	typ       string
	principal string
	underGov  bool
}

// groups enumerates the six categories from the #MyId selector. Iterating all
// of them yields the full register (~449).
var groups = []group{
	{"Statliga förvaltningsmyndigheter", "Förvaltningsmyndighet", "Regeringen", true},
	{"Myndigheter under riksdagen", "Riksdagsmyndighet", "Riksdagen", false},
	{"Statliga affärsverk", "Affärsverk", "Regeringen", true},
	{"AP-fonder", "AP-fond", "Regeringen", true},
	{"Sveriges domstolar samt Domstolsverket", "Domstol", "Riksdagen", false},
	{"Svenska utlandsmyndigheter", "Utlandsmyndighet", "Regeringen", true},
}

var (
	rowRe   = regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)
	cellRe  = regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
	tagRe   = regexp.MustCompile(`<[^>]+>`)
	orgNrRe = regexp.MustCompile(`^\d{6}-\d{4}$`)
)

func cleanCell(s string) string {
	s = tagRe.ReplaceAllString(s, "")
	return strings.TrimSpace(html.UnescapeString(s))
}

// slugify mirrors riksdag.toSlug so detail-page slugs stay consistent.
func slugify(name string) string {
	replacer := strings.NewReplacer(
		"å", "a", "Å", "a",
		"ä", "a", "Ä", "a",
		"ö", "o", "Ö", "o",
		" ", "-",
	)
	s := strings.ToLower(replacer.Replace(name))
	var b strings.Builder
	for _, r := range s {
		if r <= unicode.MaxASCII && (r == '-' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func parseAgencyTable(fragment string, g group) []ports.RegisterEntry {
	var out []ports.RegisterEntry
	for _, rm := range rowRe.FindAllStringSubmatch(fragment, -1) {
		cells := cellRe.FindAllStringSubmatch(rm[1], -1)
		if len(cells) < 2 {
			continue
		}
		name := cleanCell(cells[0][1])
		org := cleanCell(cells[1][1])
		if name == "" || !orgNrRe.MatchString(org) {
			continue
		}
		var sfs, web string
		if len(cells) > 2 {
			sfs = cleanCell(cells[2][1])
		}
		if len(cells) > 3 {
			web = cleanCell(cells[3][1])
		}
		out = append(out, ports.RegisterEntry{
			OrgNumber:       org,
			Slug:            slugify(name),
			Name:            name,
			Type:            g.typ,
			PrincipalBody:   g.principal,
			UnderGovernment: g.underGov,
			Website:         web,
			SFS:             sfs,
		})
	}
	return out
}
```
(The HTTP `Client` + `FetchRegister` are added in Task 4.)

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test -C backend ./internal/riksdag/adapters/registret/ -v`
Expected: PASS (both `TestParseAgencyTable` and `TestSlugify`).

- [ ] **Step 5: Commit**

```bash
git add backend/internal/riksdag/adapters/registret/
git commit -m "feat(riksdag): SCB register HTML table parser + slugify"
```

---

### Task 4: Register scraper — FetchRegister (HTTP per group)

**Files:**
- Modify: `backend/internal/riksdag/adapters/registret/client.go`

- [ ] **Step 1: Append the HTTP client**

Add these imports to `backend/internal/riksdag/adapters/registret/client.go`:
`"context"`, `"fmt"`, `"io"`, `"net/http"`, `"time"` (keep the existing four).
Then append the `Client` implementation to the file:
```go
const hamtaURL = "https://myndighetsregistret.scb.se/Myndighet/HamtaMynd"

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) FetchRegister(ctx context.Context) ([]ports.RegisterEntry, error) {
	byOrg := make(map[string]ports.RegisterEntry)
	var firstErr error

	for i, g := range groups {
		entries, err := c.fetchGroup(ctx, g)
		if err != nil {
			// Best-effort: log via returned error, keep going so one bad group
			// does not lose the others.
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		for _, e := range entries {
			byOrg[e.OrgNumber] = e // dedupe across groups by org number
		}
		if i < len(groups)-1 {
			time.Sleep(400 * time.Millisecond)
		}
	}

	if len(byOrg) == 0 {
		if firstErr != nil {
			return nil, fmt.Errorf("register: all groups failed: %w", firstErr)
		}
		return nil, fmt.Errorf("register: no entries parsed")
	}

	out := make([]ports.RegisterEntry, 0, len(byOrg))
	for _, e := range byOrg {
		out = append(out, e)
	}
	return out, nil
}

func (c *Client) fetchGroup(ctx context.Context, g group) ([]ports.RegisterEntry, error) {
	body := fmt.Sprintf(`{"mynd":%q}`, g.label)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hamtaURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "riksdagskollen/1.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hamtamynd %q: %w", g.label, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hamtamynd %q: HTTP %d", g.label, resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseAgencyTable(string(raw), g), nil
}
```
Add `"io"` to the imports; keep `"fmt"` (now used).

- [ ] **Step 2: Verify it builds and existing tests still pass**

Run: `go build -C backend ./... && go test -C backend ./internal/riksdag/adapters/registret/ -v`
Expected: build success; parser tests PASS.

- [ ] **Step 3 (optional live smoke test, network required):**

```bash
cd backend && go run ./internal/riksdag/adapters/registret/cmd 2>/dev/null || true
```
Skip if no such cmd; the worker (Task 6) exercises it. Live verification happens
in Task 8.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/riksdag/adapters/registret/client.go
git commit -m "feat(riksdag): FetchRegister scrapes all 6 register groups"
```

---

### Task 5: Postgres AuthorityRepository

**Files:**
- Create: `backend/internal/riksdag/adapters/postgres/authority_repository.go`

- [ ] **Step 1: Write the repository**

`backend/internal/riksdag/adapters/postgres/authority_repository.go`:
```go
package postgres

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
)

type AuthorityRepository struct {
	pool *pgxpool.Pool
}

func NewAuthorityRepository(pool *pgxpool.Pool) *AuthorityRepository {
	return &AuthorityRepository{pool: pool}
}

func (r *AuthorityRepository) UpsertAuthorities(ctx context.Context, items []domain.RegisteredAuthority) (int, error) {
	n := 0
	for _, a := range items {
		// Register fields only (Phase 1). Expenditure/headcount columns are left
		// untouched on update so later enrichment is not wiped.
		_, err := r.pool.Exec(ctx, `
			INSERT INTO authorities
				(org_number, slug, name, type, principal_body, department,
				 under_government, website, sfs, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9, now())
			ON CONFLICT (org_number) DO UPDATE SET
				slug             = EXCLUDED.slug,
				name             = EXCLUDED.name,
				type             = EXCLUDED.type,
				principal_body   = EXCLUDED.principal_body,
				under_government = EXCLUDED.under_government,
				website          = EXCLUDED.website,
				sfs              = EXCLUDED.sfs,
				updated_at       = now()
		`, a.OrgNumber, a.Slug, a.Name, a.Type, a.PrincipalBody, a.Department,
			a.UnderGovernment, a.Website, a.SFS)
		if err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (r *AuthorityRepository) List(ctx context.Context, f ports.AuthorityFilter) ([]domain.RegisteredAuthority, error) {
	where, args := buildWhere(f)
	sql := `
		SELECT org_number, slug, name, type, principal_body, department,
		       under_government, website, sfs, expenditure_mdkr, budget_mdkr,
		       headcount_int, year, updated_at
		FROM authorities ` + where + `
		ORDER BY expenditure_mdkr DESC NULLS LAST, name ASC`
	if f.Limit > 0 {
		args = append(args, f.Limit)
		sql += " LIMIT $" + itoa(len(args))
		args = append(args, f.Offset)
		sql += " OFFSET $" + itoa(len(args))
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.RegisteredAuthority
	for rows.Next() {
		var a domain.RegisteredAuthority
		if err := rows.Scan(
			&a.OrgNumber, &a.Slug, &a.Name, &a.Type, &a.PrincipalBody, &a.Department,
			&a.UnderGovernment, &a.Website, &a.SFS, &a.ExpenditureMdkr, &a.BudgetMdkr,
			&a.HeadcountInt, &a.Year, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AuthorityRepository) Count(ctx context.Context, f ports.AuthorityFilter) (int, error) {
	where, args := buildWhere(f)
	var n int
	err := r.pool.QueryRow(ctx, "SELECT count(*) FROM authorities "+where, args...).Scan(&n)
	return n, err
}

func buildWhere(f ports.AuthorityFilter) (string, []any) {
	var conds []string
	var args []any
	if q := strings.TrimSpace(f.Query); q != "" {
		args = append(args, "%"+q+"%")
		conds = append(conds, "name ILIKE $"+itoa(len(args)))
	}
	if f.UnderGovernment != nil {
		args = append(args, *f.UnderGovernment)
		conds = append(conds, "under_government = $"+itoa(len(args)))
	}
	if len(conds) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(conds, " AND "), args
}

func itoa(i int) string {
	return strings.TrimPrefix(strings.Repeat("", 0)+string(rune('0'+0)), "") + intToStr(i)
}

func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(b[pos:])
}
```
> Note: `itoa`/`intToStr` avoid importing `strconv` twice; if the engineer
> prefers, replace both with `strconv.Itoa` and add the import — behavior is
> identical. (Simpler: `import "strconv"` and use `strconv.Itoa`.)

- [ ] **Step 2: Simplify with strconv (cleanup)**

Replace the `itoa`/`intToStr` helpers with `strconv.Itoa`: add `"strconv"` to
imports, and change every `itoa(len(args))` → `strconv.Itoa(len(args))`. Delete
the two helper funcs.

- [ ] **Step 3: Verify it builds**

Run: `go build -C backend ./...`
Expected: success.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/riksdag/adapters/postgres/authority_repository.go
git commit -m "feat(riksdag): AuthorityRepository upsert/list/count"
```

---

### Task 6: Authorities ingestion worker (TDD on mapping + orchestration)

**Files:**
- Create: `backend/internal/ingestion/workers/authorities.go`
- Create: `backend/internal/ingestion/workers/authorities_test.go`

- [ ] **Step 1: Write the failing test**

`backend/internal/ingestion/workers/authorities_test.go`:
```go
package workers_test

import (
	"context"
	"errors"
	"testing"

	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
	"riksdagskollen/internal/ingestion/workers"
)

type mockRegisterClient struct {
	entries []ports.RegisterEntry
	err     error
}

func (m *mockRegisterClient) FetchRegister(_ context.Context) ([]ports.RegisterEntry, error) {
	return m.entries, m.err
}

type mockAuthorityRepo struct {
	upserted []domain.RegisteredAuthority
	upsertErr error
}

func (m *mockAuthorityRepo) UpsertAuthorities(_ context.Context, items []domain.RegisteredAuthority) (int, error) {
	if m.upsertErr != nil {
		return 0, m.upsertErr
	}
	m.upserted = append(m.upserted, items...)
	return len(items), nil
}
func (m *mockAuthorityRepo) List(_ context.Context, _ ports.AuthorityFilter) ([]domain.RegisteredAuthority, error) {
	return nil, nil
}
func (m *mockAuthorityRepo) Count(_ context.Context, _ ports.AuthorityFilter) (int, error) {
	return 0, nil
}

func TestAuthoritiesWorker_MapsAndUpserts(t *testing.T) {
	client := &mockRegisterClient{entries: []ports.RegisterEntry{
		{OrgNumber: "202100-2114", Slug: "arbetsformedlingen", Name: "Arbetsförmedlingen", Type: "Förvaltningsmyndighet", PrincipalBody: "Regeringen", UnderGovernment: true, Website: "www.arbetsformedlingen.se", SFS: "2007:1030"},
	}}
	repo := &mockAuthorityRepo{}
	w := workers.NewAuthoritiesWorker(client, repo)

	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.upserted) != 1 {
		t.Fatalf("expected 1 upserted, got %d", len(repo.upserted))
	}
	got := repo.upserted[0]
	if got.Name != "Arbetsförmedlingen" || got.OrgNumber != "202100-2114" || !got.UnderGovernment {
		t.Errorf("mapped wrong: %+v", got)
	}
	if got.ExpenditureMdkr != nil || got.HeadcountInt != nil {
		t.Errorf("Phase 1 must leave expenditure/headcount nil, got %+v", got)
	}
}

func TestAuthoritiesWorker_FetchErrorPropagates(t *testing.T) {
	client := &mockRegisterClient{err: errors.New("scrape down")}
	repo := &mockAuthorityRepo{}
	w := workers.NewAuthoritiesWorker(client, repo)

	if err := w.Run(context.Background()); err == nil {
		t.Error("expected error when fetch fails")
	}
	if len(repo.upserted) != 0 {
		t.Error("must not upsert on fetch failure (keep last good data)")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test -C backend ./internal/ingestion/workers/ -run TestAuthoritiesWorker -v`
Expected: FAIL — `undefined: workers.NewAuthoritiesWorker`.

- [ ] **Step 3: Write the worker**

`backend/internal/ingestion/workers/authorities.go`:
```go
package workers

import (
	"context"
	"log/slog"

	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
)

type AuthoritiesWorker struct {
	client ports.RegisterClient
	repo   ports.AuthorityRepository
}

func NewAuthoritiesWorker(client ports.RegisterClient, repo ports.AuthorityRepository) AuthoritiesWorker {
	return AuthoritiesWorker{client: client, repo: repo}
}

func (w *AuthoritiesWorker) Name() string { return "authorities" }

func (w *AuthoritiesWorker) Run(ctx context.Context) error {
	entries, err := w.client.FetchRegister(ctx)
	if err != nil {
		// Keep last good data — do not touch the table on a fetch failure.
		return err
	}

	items := make([]domain.RegisteredAuthority, 0, len(entries))
	for _, e := range entries {
		items = append(items, domain.RegisteredAuthority{
			OrgNumber:       e.OrgNumber,
			Slug:            e.Slug,
			Name:            e.Name,
			Type:            e.Type,
			PrincipalBody:   e.PrincipalBody,
			UnderGovernment: e.UnderGovernment,
			Website:         e.Website,
			SFS:             e.SFS,
			// ExpenditureMdkr / HeadcountInt intentionally nil in Phase 1.
		})
	}

	n, err := w.repo.UpsertAuthorities(ctx, items)
	if err != nil {
		return err
	}
	slog.Info("authorities: upserted register", "count", n)
	return nil
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test -C backend ./internal/ingestion/workers/ -run TestAuthoritiesWorker -v`
Expected: PASS (both cases).

- [ ] **Step 5: Commit**

```bash
git add backend/internal/ingestion/workers/authorities.go backend/internal/ingestion/workers/authorities_test.go
git commit -m "feat(ingestion): authorities worker scrapes register into DB"
```

---

### Task 7: Wire into main.go

**Files:**
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: Add the import**

In the import block of `backend/cmd/api/main.go`, alongside the other riksdag
adapter imports (near `riksdagSK`, `riksdagStatic`), add:
```go
	riksdagRegistret "riksdagskollen/internal/riksdag/adapters/registret"
```

- [ ] **Step 2: Construct repo + worker**

After the existing `agencyIntelRepo := riksdagPG.NewAgencyIntelRepository(db)`
line (~line 177), add:
```go
	authorityRepo := riksdagPG.NewAuthorityRepository(db)
```
After the existing `agencyIntelWorker := workers.NewAgencyIntelWorker(...)` line
(~line 225), add:
```go
	authoritiesWorker := workers.NewAuthoritiesWorker(riksdagRegistret.NewClient(), authorityRepo)
```

- [ ] **Step 3: Register on the scheduler**

After the existing `sched.RegisterSync("@weekly", &agencyIntelWorker)` block
(~line 236), add:
```go
	if err := sched.RegisterSync("@weekly", &authoritiesWorker); err != nil {
		slog.Error("failed to register authorities worker", "error", err)
	}
```

- [ ] **Step 4: Verify the full backend builds**

Run: `go build -C backend ./...`
Expected: success. (`authorityRepo` is used by the worker, so no unused-var
error. It will also be passed to the service in a later phase.)

- [ ] **Step 5: Commit**

```bash
git add backend/cmd/api/main.go
git commit -m "feat(riksdag): wire authorities worker on weekly schedule + initial sync"
```

---

### Task 8: End-to-end verification

**Files:** none (verification only).

- [ ] **Step 1: Bring up the stack with initial sync**

Run:
```bash
INITIAL_SYNC=true make run
```
(or set `INITIAL_SYNC=true` in `docker-compose.dev.yml` for the backend service).
Watch logs for: `ingestion worker starting worker=authorities` then
`authorities: upserted register count=<N>`.

- [ ] **Step 2: Verify the row count in the DB**

Run:
```bash
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U riksdagskollen -d riksdagskollen \
  -c "SELECT count(*) AS total,
             count(*) FILTER (WHERE under_government) AS under_gov
      FROM authorities;"
```
Expected: `total` in the ~400–460 range; `under_gov` ~240–380. (Exact numbers
vary with the register; the point is it is hundreds, not 10.)

- [ ] **Step 3: Spot-check a few rows**

Run:
```bash
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U riksdagskollen -d riksdagskollen \
  -c "SELECT name, org_number, type, under_government, slug
      FROM authorities ORDER BY name LIMIT 5;"
```
Expected: real agency names, valid `NNNNNN-NNNN` org numbers, group-derived
`type`, and slugs matching the existing detail-page slug format.

- [ ] **Step 4: Confirm no UI regression**

Open `http://localhost:5173/riksdag`. The MYNDIGHETER card and detail pages must
look exactly as before (they still use the live-fetch path; the new table is not
read yet). No errors in the console.

---

## Phase boundary

Phase 1 ends here: the register lives in the DB, refreshed weekly, with zero UI
change. Subsequent phases (separate plans, written after verifying each source's
current endpoints):

- **Phase 2 — Expenditure:** `esv` client → fill `expenditure_mdkr` +
  `expenditure_history` by org number.
- **Phase 3 — Headcount:** `arbetsgivarverket` client → fill `headcount_int` +
  `headcount_history`.
- **Phase 4 — Reads + UI:** service reads `AuthorityRepository`; card switches to
  DB top-10; new searchable `/riksdag/myndigheter` page; openapi + data-source
  discipline (mermaid, `docs/data-sources/*`, SourceMarker, `/data`).

---

## Self-Review

**Spec coverage (Phase 1 slice):** register source (scrape) ✓ Task 3–4;
`authorities` table ✓ Task 1; org-number identity + dedupe ✓ Task 4;
group→type/under_government ✓ Task 3; DB-backed + worker ✓ Task 5–7;
best-effort/keep-last-good (no wipe on failure) ✓ Task 6 + upsert-only Task 5;
"saknas" via nil pointers ✓ Task 2. Deferred to later phases (explicitly):
expenditure, headcount, list page, openapi, data-source docs, card switch.

**Placeholder scan:** Task 3 ships only the parser (file compiles on its own —
every import used by the test); Task 4 appends the HTTP client + its imports.
Task 5 shows a hand `itoa` then Step 2 replaces it with `strconv.Itoa` — both
shown in full, no TODO. No "implement later" anywhere.

**Type consistency:** `ports.RegisterEntry`, `ports.RegisterClient`,
`ports.AuthorityRepository`, `ports.AuthorityFilter`, and
`domain.RegisteredAuthority` are used identically across Tasks 2–7.
`NewAuthoritiesWorker(client, repo)` and `NewAuthorityRepository(pool)` match
their call sites in Task 7. `slugify` mirrors `riksdag.toSlug` exactly.
