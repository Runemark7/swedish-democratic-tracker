# Agency Register Pipeline — Design

**Date:** 2026-05-25
**Status:** Design approved, ready for implementation plan
**Builds on:** PR #35 (`feat/riksdag-myndigheter-table`) — the MYNDIGHETER card in table style. This pipeline fills that card + a new list page with real, complete data.

> Note: the product UI is Swedish, so user-facing strings (e.g. "MYNDIGHETER", "Läs mer", "saknas") and Swedish proper nouns (SCB Myndighetsregistret, Statskontoret, ESV, Arbetsgivarverket) stay in Swedish. Code, identifiers, and this document are English.

---

## Goal

Replace the hardcoded list of 10 agencies with Sweden's **entire** state agency
register (~449), pulled from authoritative open sources, so the Riksdag page
shows officially correct and complete data — names for all, expenditure and
headcount wherever a source exists.

## Problem (current state)

- `backend/internal/riksdag/adapters/statskontoret/client.go` downloads the
  Statskontoret year-outcome CSV but filters to **10 hardcoded appropriation
  codes** (`targetAnslag`). Everything else is dropped.
- `backend/internal/riksdag/adapters/static/client.go` = 10 hardcoded agencies
  (fallback).
- `backend/internal/riksdag/adapters/scb/client.go` (headcount) maps only
  **8 KLS codes**.
- No DB table for agencies — everything is fetched live with a 24h in-memory
  cache.
- Result: the card shows 10 agencies. Sweden has ~449 in the register
  (~371 administrative agencies under the Government).

## Scope

- **Population:** all **449** agencies from SCB Myndighetsregistret, tagged with
  `type`/`principal_body` and an `under_government` flag (~371) so the UI can
  show all or filter.
- **Per agency:**
  - Base data (name, org number, type, principal body, department) — **always**.
  - Expenditure (ESV) — **best-effort**, `null` when no source → "saknas" in UI.
  - Headcount (Arbetsgivarverket) — **best-effort**, `null` → "saknas".
- **Fact-layer principle:** never invent values. Missing data shows as "saknas",
  not 0 or a guess. Every value carries a source marker.
- **Out of scope:** municipal/regional boards (not in the SCB register).

## Data sources

| Source | Provides | Access | Join key |
|---|---|---|---|
| SCB Myndighetsregistret | List (~449): name, org number, SFS, website | **Web-register scrape (chosen, no cert)** — see below | org number |
| ESV/Hermes open data (`esv.se/psidata`) | Outcome per appropriation **and agency** | Open-data files/endpoints | org number (fallback: name) |
| Arbetsgivarverket "Anställda i staten" | Headcount per agency, 1991→ | Open data/export, updated Dec/June | org number (fallback: name) |

### Register scrape (verified 2026-05-25)

`POST https://myndighetsregistret.scb.se/Myndighet/HamtaMynd`
- Header: `Content-Type: application/json; charset=utf-8`
- Body: `{"mynd":"<group name>"}` (the exact group label, see below)
- Response: an **HTML table fragment** with columns `Namn`, `Organisationsnr`,
  `SFS`, `WebbAdress`. (Not JSON.)

The `#MyId` group selector enumerates six groups; iterate all to get the full
register, and map each group to `type` / `under_government`:

| Group label (`mynd`) | type | under_government |
|---|---|---|
| `Statliga förvaltningsmyndigheter` | Förvaltningsmyndighet | true |
| `Myndigheter under riksdagen` | Riksdagsmyndighet | false |
| `Statliga affärsverk` | Affärsverk | true |
| `AP-fonder` | AP-fond | true |
| `Sveriges domstolar samt Domstolsverket` | Domstol | false |
| `Svenska utlandsmyndigheter` | Utlandsmyndighet | true |

`Statliga förvaltningsmyndigheter` returned 244 agencies on 2026-05-25.
`principal_body` = "Regeringen" when `under_government`, else "Riksdagen".
**`department` is NOT in this register** → stays blank (default ''); sourced
later if at all.

**Two groups return different table schemas (discovered during Phase 1 e2e):**
- *Sveriges domstolar samt Domstolsverket* (~85 rows): columns
  `Namn, CfarNr, Postadress, Postnr, Postort, WebbAdress`. Courts share
  Domstolsverket's organisationsnummer, so `org_number` is synthesized as
  `cfar:<NNNNNNNN>` from the CfarNr (workplace identifier).
- *Svenska utlandsmyndigheter* (~110 rows): columns
  `Land, LopNr, Namn, Ambassadör/Generalkonsul, WebbAdress`. Embassies share
  Utrikesdepartementet's organisationsnummer; `org_number` is synthesized as
  `utland:<LopNr>` from the register's sequence number.

Total observed register (all 6 groups): **448** entries (244 förvaltning, 107
utland, 83 domstol, 6 AP-fond, 5 under riksdagen, 3 affärsverk) — matches the
"~449" figure from SCB. The synthetic-key prefix (`cfar:` / `utland:`) keeps the
PK stable across runs.

**Note:** ESV was renamed to Statskontoret on 2026-01-01 (merger). Endpoints and
branding for the expenditure source (Phase 2) are in flux — exact URLs are
verified when planning that phase.

**Identity:** `org_number` is the canonical key. Sources lacking org number are
matched via a normalized name (lowercased, trimmed, company-suffix stripped).
Unmatched rows are logged — no guessed joins.

## Architecture

Hexagonal, inside the existing `backend/internal/riksdag/` feature. Reuse the
existing English domain term `Authority` (already used for agencies across the
codebase: `domain.Authority`, `GetAuthorities`, the `/authorities` endpoint, the
frontend `Authority` type).

### Domain
Extend `domain/authority.go`:
```go
type Authority struct {
    OrgNumber          string
    Slug               string
    Name               string
    Type               string   // e.g. "Förvaltningsmyndighet", "Domstol"
    PrincipalBody      string   // "Regeringen", "Riksdagen", ...
    Department         string
    UnderGovernment    bool
    ExpenditureMdkr    *float64 // nil = missing
    BudgetMdkr         *float64
    HeadcountInt       *int     // nil = missing
    Year               int
    ExpenditureHistory []YearlyExpenditure
    HeadcountHistory   []YearlyHeadcount
    UpdatedAt          time.Time
}
```
Pointers for expenditure/headcount make "missing" explicit through the stack.

### Ports (`ports/`)
```go
type RegisterClient interface {
    FetchRegister(ctx) ([]RegisterEntry, error)        // 449 base entries
}
type ExpenditureClient interface {
    FetchExpenditure(ctx) ([]AgencyExpenditure, error) // ESV, per org number
}
type HeadcountClient interface {
    FetchHeadcounts(ctx) ([]AgencyHeadcount, error)    // Arbetsgivarverket
}
type AuthorityRepository interface {
    UpsertAuthorities(ctx, []domain.Authority) error
    List(ctx, AuthorityFilter) ([]domain.Authority, error) // search/filter/paginate
    ListTop(ctx, n int) ([]domain.Authority, error)        // top-N by expenditure
    GetBySlug(ctx, slug string) (*domain.Authority, error)
}
```

### Adapters (`adapters/`)
- `registret/` — SCB Myndighetsregistret (new).
- `esv/` — ESV/Hermes outcome (new; replaces the appropriation filter in
  `statskontoret/`).
- `arbetsgivarverket/` — headcount (new; replaces the limited `scb/` client).
- `postgres/authority_repository.go` — new.
- Keep `agency_intel_repository` (regleringsbrev/decisions) and the static
  10-agency list as an **emergency fallback** if the DB table is empty (first
  boot).

### Ingestion worker
`internal/ingestion/workers/authorities.go`, cron **@weekly**:
1. `FetchRegister` → upsert base rows (org-number key, compute slug +
   `under_government`).
2. `FetchExpenditure` → update expenditure + history per org number.
3. `FetchHeadcounts` → update headcount + history per org number (name fallback).

**Resilience:** each step has its own try/catch. If a source fails: log, keep
the last good data — **never wipe** on a fetch error. Partial results are saved.

### Service
`service.go` reads from `AuthorityRepository` (not live):
- `GetAuthorities(ctx)` / `ListTop(ctx, 10)` → the card.
- `ListAuthorities(ctx, filter)` → the list page (search/filter/paginate).
- `GetAuthority(ctx, slug)` → the detail page (existing, now reads DB + intel).

## Data model (migration)

New migration `000020_authorities.up.sql`:
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
    expenditure_mdkr    DOUBLE PRECISION,             -- NULL = missing
    budget_mdkr         DOUBLE PRECISION,
    headcount_int       INTEGER,                      -- NULL = missing
    year                INTEGER NOT NULL DEFAULT 0,
    expenditure_history JSONB NOT NULL DEFAULT '[]',
    headcount_history   JSONB NOT NULL DEFAULT '[]',
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_authorities_under_government ON authorities (under_government);
CREATE INDEX idx_authorities_expenditure ON authorities (expenditure_mdkr DESC NULLS LAST);
CREATE INDEX idx_authorities_name_trgm ON authorities USING gin (name gin_trgm_ops);
```
History as JSONB (displayed, not queried) → one table, no extra join tables.
Search via `pg_trgm` on `name`.

## API (openapi.yaml first → regenerate api-contract.ts)

- `GET /api/riksdag/authorities?limit=10` — top-N by expenditure (existing, now
  reads DB). Drives the card.
- `GET /api/riksdag/authorities/list?q=&principalBody=&underGovernment=&page=&pageSize=`
  — searchable/filterable list (new). Drives the list page.
- `GET /api/riksdag/myndigheter/{slug}` — detail (existing UI route slug stays
  Swedish to match the current detail page).

Update `api/openapi.yaml` first, run `npm run generate:api`. Never hand-edit
`api-contract.ts`.

## Frontend

- **Card** (`RiksdagPage`, MYNDIGHETER, from PR #35): top-10 by expenditure +
  a "Läs om fler myndigheter →" button to `/riksdag/myndigheter`.
- **New page** `/riksdag/myndigheter` (`AuthorityListPage`): search box (name),
  filters (principal body, an "under regeringen" toggle), same table style
  (Myndighet / Senaste / Andel / Förändring + chevron), pagination/virtualization
  for hundreds of rows. Row expand = current year + headcount + cost +
  "Läs mer →" (detail page).
- Detail page `/riksdag/myndigheter/:slug` exists — now reads DB data.

## Data-source discipline (CLAUDE.md requirement)

1. Update the mermaid diagram in `CLAUDE.md`: add SCB Myndighetsregistret,
   ESV/Hermes, Arbetsgivarverket; adjust Statskontoret year-outcome.
2. `docs/data-sources/`: new MD files (`scb-myndighetsregistret.md`,
   `esv-utfall.md`, `arbetsgivarverket-anstallda.md`) from `_TEMPLATE.md` with
   frontmatter + reproducible download steps.
3. `cd frontend && npm run sync:data-sources` → regenerate SourceRegistry.
4. `<SourceMarker sourceId="...">` on all new UI values + `<SectionSource>`.
5. Verify the sources appear at `/data` and `/data/<id>`.

## Error handling

- Worker: per-source isolation, partial results saved, no wipe on failure.
- Missing expenditure/headcount → `NULL` → UI "saknas".
- Org number canonical; normalized name fallback; unmatched rows logged, never
  guessed.
- DB empty (first boot before worker) → static 10 as emergency fallback.
- `INITIAL_SYNC=true` runs the worker immediately on boot (existing pattern).

## Testing

The repo has no tests yet, but the new high-risk logic is tested (TDD):
- Parsers: register, ESV outcome, headcount (fixtures from real data).
- Merge/join: org-number match + name fallback + null handling.
- Repo: upsert (idempotent), List filter/search/paginate (integration against a
  test DB).
Pure functions (parse/merge) are prioritized — highest value, easiest to test.

## Phased rollout

Each phase is independently shippable and testable:

1. **Register foundation:** migration + `registret` client + repo + worker
   (register only) + service reads DB + card shows 449 names (expenditure/
   headcount may be "saknas" initially). Replaces the hardcoded 10.
2. **Expenditure:** `esv` client + worker step + expenditure in card/detail.
3. **Headcount:** `arbetsgivarverket` client + worker step + headcount.
4. **List page + search:** `/riksdag/myndigheter` + "fler" button + list API
   endpoint.

Max 5 files per phase (CLAUDE.md), verify (`go build`, `tsc`) between phases.

## Risks / open questions (verified during planning)

- **Register scrape is HTML-structure-dependent** (chosen over the cert-gated
  API). If SCB changes the `HamtaMynd` markup, the parser breaks — worker keeps
  last good data and logs. Fallback: SCB business-register API (needs cert).
- **ESV↔Statskontoret rename** (Jan 2026): confirm current open-data
  endpoints/formats.
- **Org number** may be absent in ESV/Arbetsgivarverket exports → name-matching
  fallback; measure hit rate.
- **Expenditure coverage:** agencies without their own appropriation get
  "saknas" — expected.

## Success criteria

- `/riksdag/myndigheter` lists all ~449 with working search + filter.
- The card shows top-10 by expenditure + a "fler" button.
- Expenditure/headcount shown where a source exists, "saknas" otherwise — no
  guessed values.
- Every value has a source marker; sources appear at `/data`.
- Data persists in the DB; one source failing does not break the page.
