# Source Verification — Design Spec

**Date**: 2026-06-06
**Status**: design / pre-implementation
**Branch**: `feat/source-verification`

## Why

Riksdagskollen's credibility rests on every value on the site being traceable to a verifiable primary source. A prior audit (commits `3d37401` "scrub repo paths + fix broken upstream URLs" and `f71b11a` "replace remaining hand-rolled Källa: lines with SectionSource") raised that bar significantly. Since then we have added new data sources (Statskontoret årsutfall, Kolada spending, several SCB tables, seed-budget-data, seed-party-goals, derived-agenda, wikimedia-svg) and new UI surfaces (KPI ranking, mandate composition, Regering/ministers, agency expenditure) that consume them.

We need a second-pass verification that:

1. Confirms every registered source's upstream URL still resolves and returns the expected payload shape.
2. Confirms every source MD contains a **narrative overview** of how the data was obtained — written so a third party could repeat the steps and arrive at the same data. The MD is an overview, not a runbook: no executable code, no curl/wget commands, no API keys, no internal endpoints, no scraper logic. Anything that would help an attacker reproduce our ingestion in an automated way belongs in the backend client code, not in public docs.
3. Confirms every UI value derived from a fetch/CSV/ZIP/document carries a `<SourceMarker>` pointing at a registered source.
4. Confirms every external HTTP client in `backend/internal/*/adapters/` maps to a registered source.
5. Tags any source that fails the above as **unsure** in the UI so the failure is visible to the user, not hidden in our docs.

## Scope

### In scope

- All 13 currently registered sources in `docs/data-sources/`.
- All `<SourceMarker sourceId=...>` and `<SectionSource sourceIds={[...]}>` occurrences in `frontend/src/`.
- All external HTTP clients in `backend/internal/*/adapters/` (riksdagen, kolada, scb, ted, statskontoret).
- Seed/derived sources (seed-budget-data, seed-party-goals, derived-agenda, wikimedia-svg) — verified via "frozen" path.
- Any UI value that quotes a number, date, or named entity sourced from external data: KPI values, mandate counts, populations, tax rates, agency expenditure, budget figures, vote outcomes, speech text — all in scope.

### Out of scope

- Real AI matching (`internal/matching/ports/ai.go` stays stubbed).
- Refactoring the source-registry build pipeline beyond adding the new fields.
- Migrating to a different source-tracking system; we extend the existing `docs/data-sources/` + `SourceRegistry.generated.ts` flow.

## Documentation rules for source MDs

Each `docs/data-sources/*.md` is a **public, citizen-facing overview**, not an operator runbook. The rule applies to all sources, including CSV/ZIP sources (which CLAUDE.md flags as needing "reproducible download steps" — that requirement is satisfied by the narrative, not by code).

Required prose answers (no specific format):
- Who publishes the data (organisation, authority, jurisdiction).
- What the dataset is and at what granularity (year, kommun, KPI, etc.).
- Where a user goes to find it (the public landing page, dataset id, table name, file name).
- Which filter/selection a user would apply to reach the same slice we use.
- License + freshness expectation.

Forbidden in source MDs (move to backend client code or internal docs):
- Shell commands (`curl`, `wget`, `psql`, `unzip`, etc.).
- Code blocks containing API request bodies, scraping logic, or transformation scripts.
- API keys, tokens, internal hostnames, staging URLs.
- Exact request headers, undocumented query parameters, or bypasses for rate limiting.
- Any step whose only purpose is to automate ingestion at scale.

Phase 1 audit treats violations as a `Gaps` entry; Phase 2 rewrites the MD as narrative before the source can be marked `verified` or `frozen`.

## Architecture

Two phases, single spec.

### Phase 1 — Audit (investigation, no production code)

Produce a single working document `docs/superpowers/audits/2026-06-06-source-verification.md` with:

**Main table** (one row per registered source, 13 rows):

| Field | Meaning |
|---|---|
| `id` | Registry id |
| `URL liveness` | HTTP status + content-type from upstream, or "n/a (seed)" |
| `Schema match` | Yes/No — sampled call still matches parser shape |
| `Repro steps OK?` | Yes/No — does the MD narrative cover origin + which dataset/endpoint + how to navigate to it, without code/curl/keys? Gap notes if no |
| `UI consumers` | File paths + line numbers grepped from `SourceMarker sourceId="<id>"` and `SectionSource sourceIds={[..., "<id>", ...]}` |
| `Backend client` | Package path under `backend/internal/`, or `seed` / `derived` |
| `Proposed status` | `verified` \| `frozen` \| `unsure` |
| `Gaps` | Concrete fixes required before sign-off |

**Appendices** — exhaustive, no skipping:

- **Appendix 1: Untagged values** — every UI string that quotes a number/date sourced from external data but lacks a `<SourceMarker>`. Format: file:line + value + which source it _should_ link to.
- **Appendix 2: Untracked backend clients** — every `http.Get`/`http.Post`/CSV downloader in `backend/internal/` whose target URL does not map to a registered source.
- **Appendix 3: Sources not yet in registry** — anything discovered during the audit that is neither in `docs/data-sources/` nor seed.

The audit doc itself is the Phase 1 deliverable. It is committed to the repo so the trail is auditable.

### Phase 2 — Codification (production code)

Based on Phase 1 findings:

1. Close gaps in `docs/data-sources/*.md` — fix repro steps, update URLs, add missing fields.
2. Tag every Appendix 1 UI value with `<SourceMarker>` or remove the value.
3. Map every Appendix 2 backend client to a registered source (creating new MDs if needed).
4. Create MDs for every Appendix 3 source.
5. Land schema + UI + tooling (next sections).

## Data flow

```
docs/data-sources/<id>.md ──┐
  frontmatter:              │  sync-data-sources.mjs
    verificationStatus      ├──► SourceRegistry.generated.ts
    verificationNotes       │     (typed; new fields)
    lastVerified            │
    reproSteps (existing)   │
                            ▼
            <SourceMarker sourceId> ──► popover badge if status≠verified/frozen
            <SectionSource sourceIds> ─► popover lists per-source status
                            ▼
                  /data (index, +status col)
                  /data/<id> (detail, +verification block)

verify-sources.mjs (Node, runs in Docker):
  reads each MD frontmatter.upstream + kind
  api  → fetch HEAD (fallback GET); assert 2xx + content-type match
  csv  → GET; assert content-length>0 + content-type contains csv|zip
  seed/synthesized → skip (status human-managed)
  emits verify-report.json
  exits non-zero on any regression
  GH Action: weekly cron + on-PR-touching docs/data-sources/**
  failure opens issue labelled data-source-drift
```

## Components

### Frontmatter schema additions

Each `docs/data-sources/*.md` gains:

```yaml
verificationStatus: verified | frozen | unsure
verificationNotes: "string — why this status; what's missing"
lastVerified: 2026-06-06
```

Existing fields (`id`, `name`, `kind`, `upstream`, `license`, `freshness`, `blurb`, repro instructions) stay as-is.

### Status enum semantics

- `verified` — live API or CSV, upstream URL resolves with expected content-type, payload shape still matches parser, repro steps complete, `lastVerified` within 90 days.
- `frozen` — seed or derived (e.g. `seed-budget-data`, `derived-agenda`, `wikimedia-svg`), with extraction recipe documented and source artefact committed to the repo. Not expected to drift.
- `unsure` — any of: URL dead, schema drift, repro steps missing/incomplete, `lastVerified` > 90 days, or human flagged a concern in `verificationNotes`.

### Registry sync (`frontend/scripts/sync-data-sources.mjs`)

- Parse `verificationStatus`, `verificationNotes`, `lastVerified` from frontmatter.
- Pass through to `SourceRegistry.generated.ts`.
- Update `SourceEntry` TypeScript type in `frontend/src/components/sources/SourceRegistry.ts` to include the new fields (`verificationStatus` is required; `verificationNotes` optional).
- Fail the sync if any MD declares `verificationStatus: unsure` without a non-empty `verificationNotes`.

### UI

- `SourceMarker.tsx` — when `entry.verificationStatus === "unsure"`: dot renders in amber, popover appends a row `"⚠ Ej verifierad"` followed by `verificationNotes`. When `"frozen"`: popover shows neutral row `"Statisk källa — verifierad en gång ({lastVerified})"`. When `"verified"`: no change.
- `SectionSource.tsx` — no visual change to the trigger, but popover lists per-source status per registered id.
- `/data` index page — add a Status column with chip: green "verifierad" / blue "statisk" / amber "ej verifierad". Sortable.
- `/data/<id>` detail page — verification block above the prose body: status chip, `lastVerified` formatted date, `verificationNotes` if present.

### Tooling

`frontend/scripts/verify-sources.mjs` — Node script, runs inside Docker via `docker compose run --rm frontend node scripts/verify-sources.mjs`.

- Reads every `docs/data-sources/*.md` frontmatter.
- For `kind: api` — `fetch(upstream, { method: 'HEAD', signal: AbortSignal.timeout(10_000) })`, fall back to GET on 405. Assert 2xx; record content-type.
- For `kind: csv` — GET with same timeout; assert `content-length > 0`; assert content-type contains `csv` or `zip` or `octet-stream`.
- For `kind: seed | synthesized` — skip; status is human-managed.
- Emits `verify-report.json` to working dir.
- Exits 0 if all in-scope sources reachable; exit 1 with summary if any regress.
- GH Action `.github/workflows/verify-sources.yml`:
  - Weekly cron (Monday 06:00 UTC).
  - Trigger on PR touching `docs/data-sources/**` or the script itself.
  - On failure: `gh issue create --label data-source-drift` with the report attached.

## Error handling

- `verify-sources.mjs` HTTP failures — record code + body snippet (first 200 chars) in report; exit 1.
- `verify-sources.mjs` timeout — record `timeout` status; exit 1.
- `sync-data-sources.mjs` — fail if `verificationStatus` value not in enum, or if `unsure` with empty `verificationNotes`.
- UI — `SourceMarker` already throws when its `sourceId` is not in the registry; keep that behaviour. Add a frontend typecheck-time test that every id referenced in source code exists in `SOURCES`.
- Missing `<SourceMarker>` — caught by audit, not runtime; no automatic guard since we cannot tell from code alone which strings are "values".

## Testing

- `frontend/scripts/__tests__/verify-sources.test.mjs` — mocked fetch; cases:
  - 200 + matching content-type → exit 0
  - 404 → exit 1, report row marks `unreachable`
  - 5xx → retry once, then `unreachable`
  - Timeout → `timeout`
  - `kind: seed` → skipped from report
- Manual smoke after Phase 2 lands:
  - `/data` index renders Status column with at least one of each chip variant.
  - `/data/<id>` for an `unsure` source renders amber chip + notes.
  - `SourceMarker` popover for an `unsure` source renders amber dot + warning row.

## Acceptance criteria (definition of done)

1. Audit doc `docs/superpowers/audits/2026-06-06-source-verification.md` committed; all 13 registered sources have a row; appendices 1–3 listed exhaustively.
2. Every untagged value from Appendix 1 either gets a `<SourceMarker>` or is removed from the UI.
3. Every untracked backend client from Appendix 2 maps to a registered source; new MD created if needed.
4. Every Appendix 3 source has its own MD in `docs/data-sources/`.
5. Every `docs/data-sources/*.md` has `verificationStatus`, `verificationNotes` (where applicable), and fresh `lastVerified`.
6. `cd frontend && npm run sync:data-sources` regenerates `SourceRegistry.generated.ts` with the new fields.
7. `cd frontend && npm run verify:sources` runs green locally and in CI (new script wired in `frontend/package.json`).
8. `/data` renders status column; `/data/<id>` renders verification block.
9. `cd frontend && npx tsc --noEmit` passes.
10. `go build -C backend ./...` passes (no backend code changes expected, but verify nothing broke).

## Non-goals

- We are not building an automated semantic check on UI values. Detection of "this number lacks a marker" remains a human task during the audit. The runtime guard is limited to "every marker references a real registry id".
- We are not auto-fixing drift. `verify-sources.mjs` detects and reports; humans react.
- We do not add per-record provenance for each ingested row (e.g. per-vote source URL). Source-level provenance is the unit of granularity.

## Risks / open questions

- Some SCB PxWeb endpoints are POST-only and large; `verify-sources.mjs` may need a per-source `verifyMethod` override in frontmatter for those. Decide during Phase 1 once we see actual failures.
- 90-day `lastVerified` window for `verified` status is a guess. Re-evaluate after first weekly run.
- TED API rate limits are not documented in our current source MD; verification calls must be throttled if we automate them at scale (mitigation: the script hits each source exactly once per run).
