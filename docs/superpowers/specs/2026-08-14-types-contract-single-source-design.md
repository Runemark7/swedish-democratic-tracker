# Single source of truth for frontend types

**Date:** 2026-08-14. **Branch:** `feat/types-contract-single-source`.

## The problem

`frontend/src/shared/types.ts` hand-mirrors the generated OpenAPI contract.
It declares 67 exported types; 61 have an identical schema in the generated
`frontend/src/shared/api-contract.ts`. Every change to an API shape must
therefore be made in two places by hand, and `CLAUDE.md`'s BFF contract rule
names `api/openapi.yaml` as the single source of truth with `api-contract.ts`
as its generated output.

The mirror is a drift surface by construction. It has already bitten:

- `RecentBetankande` was declared in both files with the same three new
  fields; keeping them in sync was a manual, silently-failing step (Phase 3
  re-review).
- `PoliticianPromise` in `types.ts` is the contract's `Promise` under a
  different name — a rename drift with no mechanical check.
- 20 of the 61 schema-backed types in `types.ts` are never imported from it;
  their consumers already use the contract, so the hand copies are dead
  mirrors.
- `GET /votes` (the list-all endpoint backing `VotesPage.tsx`) is not
  described in `openapi.yaml` at all, so `VoteSummary` /
  `VoteSummaryListResponse` have no generated source.

This site's failure mode is publishing something that looks authoritative and
is wrong. A type mirror that can drift is that failure mode in its quietest
form: no crash, no red test — just two declarations walking apart.

## Goal

`api/openapi.yaml` becomes the single source of truth for every API-described
type, with no hand-maintained duplicate. `shared/types.ts` is deleted.

- Every API type is imported as `components["schemas"]["X"]` from
  `@/shared/api-contract` (generated, never hand-edited).
- Frontend-only shapes that the API genuinely cannot describe — computed
  client-side composites — move to `frontend/src/shared/ui-types.ts`, clearly
  documented as non-API types.
- `GET /votes` is added to `openapi.yaml` so its response types become
  generatable.

**Outcome test:** `rg 'from "@/shared/types"'` returns zero;
`npm run generate:api && npx tsc --noEmit` pass with no hand-edit to
`api-contract.ts`.

## Approach

### A. Document the missing endpoint

Add `GET /votes` to `api/openapi.yaml`:

- Path `/votes` with `page` / `pageSize` query params.
- Schemas `VoteSummary` and `VoteSummaryListResponse`
  (`{ data: VoteSummary[]; total: number; page: number; pageSize: number }`),
  matching the `listAll` handler's actual response shape.

Run `npm run generate:api`.

### B. Create `frontend/src/shared/ui-types.ts`

Move the types the API does not describe, with a header comment stating they
are frontend-only computed shapes, not API types:

- `BudgetTier` — local union ("national" | "regional" | "municipality"),
  used by `TierNav`.
- `BudgetSnapshot` — generic shape carrying `total_mnkr`; the contract only
  has `RegionBudgetSnapshot` / `MunicipalityBudgetSnapshot`.
- `MunicipalityAreaDataPoint` — the contract has `RegionAreaDataPoint` but no
  municipal twin.
- `PromiseWithMatches` — frontend composite: a `Promise` plus `voteMatches`.
- `PoliticianPromise` — the name-drift case. The contract names this type
  `Promise`; migrate to the contract name and delete the hand alias.

### C. Migrate consumers

Feature-by-feature, keeping `npx tsc --noEmit` green after each feature:

- `features/*/api.ts` wrappers switch response types from `@/shared/types`
  to `components["schemas"]["X"]`.
- Hooks and components importing from `@/shared/types` switch to
  `@/shared/api-contract` (or `@/shared/ui-types` for the frontend-only
  shapes).
- `useDemocracy.ts` is the largest single consumer (~20 imports) and is
  migrated by the controller.
- Feature files that already import from the contract or define their own
  locals (`regering`, `riksdag`, `speeches`) are untouched except where they
  import from `@/shared/types`.

`extends`-based interfaces in `types.ts` (`RegionDetail extends
RegionSummary`, etc.) become flat contract schemas — identical fields, no
structural change to consumers; sub-agents verify any `extends` usage
explicitly.

### D. Delete the mirror

Delete `shared/types.ts` once the last `@/shared/types` import is gone.
The ~20 schema-backed types never imported from `types.ts` are deleted
outright; anything that needs them uses the contract type.

## Execution

Disjoint sets of feature files are migrated by parallel sub-agents, each told
to: swap imports to `@/shared/api-contract`, leave field usage untouched, and
run `npx tsc --noEmit`. The controller does the `openapi.yaml` change,
`ui-types.ts` creation, `useDemocracy.ts`, and the final deletion, then runs
full verification.

## Verification

- After each feature migration: `cd frontend && npx tsc --noEmit`.
- Final: `go build -C backend ./...`, `cd frontend && npx tsc --noEmit`,
  `npm run lint`, `npm run build`.
- `rg 'from "@/shared/types"'` returns zero.
- `shared/types.ts` deleted; `ui-types.ts` holds only frontend-only shapes;
  `GET /votes` documented in `openapi.yaml`.

`tsc` is the discriminating check: if a contract type's fields diverge from
what the frontend reads, typecheck fails loudly. The mirror made that drift
silent; the contract makes it loud. No new test framework (the frontend has
none).

## Scope guard

`shared/nav.ts`, `shared/design.ts`, `shared/dates.ts`, `shared/components.tsx`,
and the `api-contract.ts` file itself are untouched. The backend's JSON output
is unchanged — this only relocates where the frontend declares its types.
