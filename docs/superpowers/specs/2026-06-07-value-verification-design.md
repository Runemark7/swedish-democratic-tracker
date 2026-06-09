# Value Verification — Design Spec

**Date**: 2026-06-07
**Status**: design / pre-implementation
**Depends on**: `docs/superpowers/specs/2026-06-06-source-verification-design.md` (source-level verification, merged in #53)

## Why

The source-verification work (#53) gave every data source a
`verificationStatus` badge. But that badge measures **plumbing**, not
**truth**: `verify-sources.mjs` only checks that an upstream URL returns 200
with the expected content-type, and that the source MD has no forbidden
`curl` blocks. It never compares a number *rendered on the site* to the
number the upstream authority *actually publishes*.

That gap matters because the region, municipality, and riksdagen pages are
the project's core promise — a neutral FACT layer where every value is
traceable and correct. A green badge that does not prove the rendered
population, mandate count, KPI, or vote tally is right gives the citizen
false confidence. "If they are not correct, it gives no value."

This spec closes two things at once:

1. **Stream A — honest badges.** Most `unsure` sources were flagged during
   the audit for documentation gaps that are *already fixed* (the notes say
   "narrative now approved") but were never promoted. A few have genuine
   missing `<SourceMarker>`s. One (`statskontoret-arsutfall`) has a real
   upstream drift. Close these so the badge reflects reality.

2. **Stream B — value verification.** A new on-demand check that compares
   every value the site shows against a freshly-fetched upstream value,
   across all three tiers (regions, kommuner, riksdagen), and writes a dated
   review report. This is the actual trust foundation.

## Scope

### In scope

**Stream A**
- Promote stale `unsure → verified` for sources whose only gap was a removed
  `curl` block: `kolada`, `kolada-spending`, `riksdagen`, `scb-kfmandat`,
  `scb-ltmandat`.
- Add the four missing `<SourceMarker>`s:
  - `RegionBudgetAreaPage.tsx:140` (mnkr chart labels) → `scb-kostndrlt`
  - `MunicipalityDetailPage:194` (population sub-header) → `scb-befolkning`
  - authority cost/staff values (`AuthorityDetailPage`, `MyndigheterListPage`)
    → `statskontoret-arsutfall`
  - map attribution (`SwedenRegionMap`, `SwedenKommunMap`) → `wikimedia-svg`
    (CC BY-SA 2.5 *requires* visible attribution).
- After markers land, promote `scb-befolkning`, `scb-kostndrlt`,
  `wikimedia-svg` → `verified`.
- Re-investigate `statskontoret-arsutfall`'s dynamic GetFile ZIP link; fix
  the MD; promote only once the data path is confirmed. Until then it stays
  `unsure`.

**Stream B — value verification** (`backend/cmd/verifyvalues/`)

| Tier | Entities | Metrics compared (ours vs live upstream) | Source(s) |
|---|---|---|---|
| Regions | all 21 | population, mandate total, displayed KPIs, budget areas | `scb-befolkning`, `scb-ltmandat`, `kolada`, `scb-kostndrlt` |
| Kommuner | all 290 | population, mandate total, displayed KPIs, spending | `scb-befolkning`, `scb-kfmandat`, `kolada`, `kolada-spending` |
| Riksdagen | n/a | party seat totals, sample vote outcomes, member roster counts, speech counts | `riksdagen` |

KPI set = the codes the UI actually renders (the region/municipality KPI
strips already defined in `CLAUDE.md` and `service.go`). Kolada's
all-entities-per-KPI batch endpoint returns all 290 municipalities (or all
regions) in one call, so KPI/spending coverage is cheap.

### Out of scope

- Refactoring the source-registry pipeline (covered by #53).
- Real AI matching (`internal/matching/ports/ai.go` stays stubbed).
- Per-record provenance (per-vote source URLs). Entity+metric is the unit.
- Auto-fixing mismatches. The report flags; a human reacts.
- Blocking CI. The check is on-demand only (see Run mode).

## Architecture

### Stream A — gap closure

Pure edits: frontmatter `verificationStatus` flips in `docs/data-sources/*.md`,
four `<SourceMarker>` additions in `frontend/src/`, one MD URL fix for
`statskontoret-arsutfall`. Re-run `npm run sync:data-sources` to regenerate
`SourceRegistry.generated.ts`, then `npm run verify:sources` to confirm
liveness. No backend code.

### Stream B — `backend/cmd/verifyvalues/`

A single cross-feature Go command. **Why Go, not a Node script:** fetching
the upstream values means replaying SCB's PxWeb POST bodies, Kolada's batch
calls, and riksdagen's JSON endpoints — logic that already lives in the
existing adapter clients (`internal/regions/adapters/scb`,
`internal/regions/adapters/kolada`, `internal/*/adapters/riksdagen`). A Node
script would duplicate that logic and re-expose request bodies the
source-verification spec deliberately keeps out of public/JS surfaces. The Go
command reuses the clients in-process: no duplication, nothing leaked.

The command mirrors `cmd/api/main.go`'s wiring style — it constructs a
`pgxpool` from `DATABASE_URL` plus the adapter clients it needs, then runs a
list of **checks**.

```
cmd/verifyvalues/
├── main.go          # wiring: pool + clients, run checks, write report
├── check.go         # Check type + Result type + diff/classify logic
├── regions.go       # region + municipality population/mandate/KPI/budget checks
└── riksdagen.go     # seat totals, vote outcomes, roster, speech counts
```

Each check produces one or more `Result`:

```go
type Result struct {
    Tier     string // "region" | "kommun" | "riksdag"
    Entity   string // "Dalarna (20)" | "Stockholm (0180)" | "Riksdag"
    Metric   string // "population 2024" | "mandate total" | "KPI N60008 2023" | "vote H801..."
    SourceID string // registry id, e.g. "scb-befolkning"
    Ours     string // value from our service (same code the HTTP API serves)
    Upstream string // value from a live upstream fetch
    Status   Status // Match | Review | Info
}
```

**Reading "ours":** call the same feature service the HTTP handler calls
(e.g. `regions.Service.GetRegion`, `votes.Service...`). This verifies the
exact value path the user sees, not a separate query.

**Reading "upstream":** call the existing adapter client live
(`scb.Client.FetchPopulationTrend`, `kolada.Client.FetchKPIAllMunicipalities`,
the riksdagen clients).

**Classification:**
- population, mandate totals, party seat totals, roster counts, vote tallies,
  KPI values, spending, budget areas → exact match expected. Any diff →
  `Review`.
- speech counts → volatile (data changes constantly). Always `Info`: delta
  recorded, never flagged.

### Run mode

On-demand, committed report. No CI gate (live external calls are slow/flaky
and the decision is review-not-fail).

```
make verify-values
# docker compose run --rm backend go run ./cmd/verifyvalues
# requires: postgres up + seeded, outbound network for live upstream calls
```

## Data flow

```
make verify-values
  │
  ▼
cmd/verifyvalues
  ├─ ours:     feature services ── pgxpool ── PostgreSQL (seeded/ingested)
  └─ upstream: existing adapter clients ── live HTTP ── SCB / Kolada / riksdagen
  │
  ▼ diff + classify each (entity, metric)
  │
  ▼
docs/superpowers/audits/2026-06-07-value-verification.md   (committed)
  ├─ "⚠ Needs review" summary (all Status=Review rows)
  ├─ Regions table
  ├─ Kommuner table
  └─ Riksdagen table  (speech counts under "Informational")
```

## Report format

Markdown, dated, committed. Top section first so a reader sees problems
immediately:

```markdown
# Value Verification — 2026-06-07

Run: <UTC timestamp>   Entities: 21 regions, 290 kommuner, riksdag

## ⚠ Needs review (N)
| Tier | Entity | Metric | Ours | Upstream | Source |
|---|---|---|---|---|---|
| kommun | Malmö (1280) | population 2024 | 357 377 | 357 891 | scb-befolkning |

## Regions (21)
| Entity | Metric | Ours | Upstream | Status | Source |
...

## Kommuner (290)
...

## Riksdagen
| Metric | Ours | Upstream | Status | Source |
| party seats S | 107 | 107 | Match | riksdagen |
...

### Informational (volatile)
| anföranden total | 41 233 | 41 980 | Info | riksdagen |
```

Empty "Needs review" section ⇒ every checked value matches upstream right now.

## Error handling

- Upstream fetch fails (timeout, 5xx, schema mismatch) → row marked
  `Status: error` with the reason; recorded, run continues (one bad endpoint
  must not abort the whole report).
- DB read fails → fatal (the stack must be up to run this).
- Kolada/SCB throttling → checks already batch one call per KPI/year; add a
  small inter-call delay if rate limits appear.
- Report is always written, even on partial failure, so the audit trail
  captures what was reachable.

## Testing

- `cmd/verifyvalues` diff/classify logic unit-tested with table tests:
  - equal values → `Match`
  - differing → `Review`
  - speech-count metric → `Info` regardless of delta
  - upstream error → `error` status, run continues
- Upstream clients are reused as-is (already covered by their own adapter
  tests); the command's tests mock the client interfaces, no live calls.
- Manual smoke: `make verify-values` against the dev stack produces a report
  with non-empty Regions/Kommuner/Riksdagen tables.

## Acceptance criteria (definition of done)

1. Stream A: the five stale `unsure` sources are `verified`; the four missing
   `<SourceMarker>`s are added; `scb-befolkning`, `scb-kostndrlt`,
   `wikimedia-svg` promoted; `statskontoret-arsutfall` either fixed+promoted
   or left `unsure` with an updated note explaining the drift.
2. `npm run sync:data-sources` regenerates the registry; `npm run
   verify:sources` runs green.
3. `backend/cmd/verifyvalues/` exists and compiles (`go build -C backend ./...`).
4. `make verify-values` target added; runs in Docker against the stack.
5. Running it writes `docs/superpowers/audits/2026-06-07-value-verification.md`
   with regions, kommuner, riksdagen tables and a "Needs review" summary.
6. Diff/classify unit tests pass.
7. `go build -C backend ./...` and `cd frontend && npx tsc --noEmit` pass.
8. Any mismatch the first run surfaces is triaged: either our ingestion is
   fixed, or the diff is explained in the report (legitimate upstream
   revision).

## Non-goals

- Not a blocking gate; never fails a build.
- Not auto-fixing data; humans react to the report.
- Not a live on-page "verified now" widget; verification is an out-of-band
  audit producing a committed artifact.
- Not per-record provenance.

## Risks / open questions

- **Seeded mandate data is from the 2022 election.** Re-fetching `ltmandat`/
  `kfmandat` live should return the same 2022 result (static), so a diff
  there means our seed is wrong — exactly what we want to catch. Confirm the
  live SCB tables still expose the 2022 selection.
- **Population year alignment.** Our DB stores a population year per entity;
  the upstream fetch must request the *same* year, or every row falsely
  flags. The check reads our stored year and pins the upstream selection to
  it.
- **Region KPI/budget included for symmetry** with kommun KPI/spending (same
  batch calls). Drop from scope if the first pass should stay minimal.
- **Speech counts** are informational only; revisit if a stable per-session
  count proves checkable.
- Runtime: ~21 + 290 entities, but KPI/population/mandate use batch calls, so
  expect tens of HTTP calls, not thousands. Acceptable for on-demand.
