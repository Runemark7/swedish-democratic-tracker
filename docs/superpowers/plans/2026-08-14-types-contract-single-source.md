# Frontend Type Single-Source-of-Truth Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Delete the hand-written `frontend/src/shared/types.ts` mirror so `api/openapi.yaml` + its generated `api-contract.ts` are the single source of truth for every API-described type.

**Architecture:** `api/openapi.yaml` is the single source of truth for HTTP types; `frontend/src/shared/api-contract.ts` is generated from it (never hand-edited). Currently `shared/types.ts` hand-declares 67 types, 61 of which duplicate a contract schema. This plan (1) adds the one undocumented endpoint (`GET /votes`) to the spec so its types become generatable, (2) creates `shared/ui-types.ts` for the genuinely frontend-only shapes the API cannot describe, (3) migrates every consumer from `@/shared/types` to `@/shared/api-contract` / `@/shared/ui-types`, and (4) deletes `types.ts`.

**Tech Stack:** TypeScript 5, openapi-typescript 7 (`npm run generate:api`), React 19, TanStack Query. No test framework in the frontend — `npx tsc --noEmit` is the discriminating check.

**Verification gate (CLAUDE.md rule 4):** `go build -C backend ./...` and `cd frontend && npx tsc --noEmit` must pass before any task is done. `npm run generate:api` regenerates `api-contract.ts`; never hand-edit it.

---

## Import convention used throughout

All contract types are imported as:
```ts
import type { components } from "@/shared/api-contract";
type X = components["schemas"]["X"];
```

Frontend-only shapes come from the new module:
```ts
import type { BudgetTier, BudgetSnapshot, MunicipalityAreaDataPoint } from "@/shared/ui-types";
```

No target type from `types.ts` is used as a runtime value (verified — all importers are `import type`, and `shared/design.ts` declares its own `PartyCode`). Deleting `types.ts` breaks nothing at runtime.

---

### Task 1: Document `GET /votes` in the spec

**Files:**
- Modify: `api/openapi.yaml`
- Modify: `frontend/src/shared/api-contract.ts` (regenerated, not hand-edited)

- [ ] **Step 1: Read the existing `VoteDetail` schema to match style**

Run: `grep -n "VoteDetail:" api/openapi.yaml`
Read the schema block to copy its indentation and property style for the new `VoteSummary` schema.

- [ ] **Step 2: Add the `VoteSummary` and `VoteSummaryListResponse` schemas**

In `api/openapi.yaml`, inside `components.schemas`, add after the `VoteDetail` schema (indent two spaces under `schemas:`):

```yaml
    VoteSummary:
      type: object
      required: [beteckning, forslagspunkt, documentTitle]
      properties:
        beteckning:
          type: string
        forslagspunkt:
          type: string
        documentTitle:
          type: string
        proposedByParty:
          type: string
        proposalType:
          type: string
    VoteSummaryListResponse:
      type: object
      required: [data, total, page, pageSize]
      properties:
        data:
          type: array
          items:
            $ref: "#/components/schemas/VoteSummary"
        total:
          type: integer
        page:
          type: integer
        pageSize:
          type: integer
```

- [ ] **Step 3: Add the `/votes` path**

In `api/openapi.yaml`, before the existing `/votes/recent` path, add:

```yaml
  /votes:
    get:
      operationId: listDistinctVotes
      tags: [votes]
      summary: Distinct vote points, newest first (matching projection)
      parameters:
        - name: page
          in: query
          required: false
          schema:
            type: integer
            minimum: 1
            default: 1
        - name: pageSize
          in: query
          required: false
          schema:
            type: integer
            minimum: 1
            default: 50
      responses:
        "200":
          description: Paginated list of distinct beteckning/förslagspunkt pairs
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/VoteSummaryListResponse"
```

- [ ] **Step 4: Regenerate the contract**

Run: `cd frontend && npm run generate:api`
Expected: openapi-typescript writes `src/shared/api-contract.ts`, now containing `VoteSummary` and `VoteSummaryListResponse`.

- [ ] **Step 5: Verify**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS (no consumers change yet — `types.ts` still exists).

- [ ] **Step 6: Commit**

```bash
git add api/openapi.yaml frontend/src/shared/api-contract.ts
git commit -m "feat(votes): document GET /votes in the OpenAPI spec"
```

---

### Task 2: Create `frontend/src/shared/ui-types.ts`

**Files:**
- Create: `frontend/src/shared/ui-types.ts`

- [ ] **Step 1: Create the file**

Create `frontend/src/shared/ui-types.ts` with exactly:

```ts
// Frontend-only types — shapes the API does not describe and that are
// computed client-side. Keep nothing here that has a counterpart in
// api/openapi.yaml; those belong in @/shared/api-contract (generated).
// If a type here ever gains an API counterpart, move it there and delete
// this one.

/** National / regional / municipal budget tier, rendered by TierNav. */
export type BudgetTier = "national" | "regional" | "municipality";

/**
 * A single area's budget snapshot. The API exposes the region- and
 * municipality-typed variants (RegionBudgetSnapshot, MunicipalityBudgetSnapshot);
 * this is the untyped shape assembled client-side in useDemocracy.
 */
export interface BudgetSnapshot {
  area_name: string;
  value_mnkr: number;
  total_mnkr: number;
  pct: number;
}

/**
 * A municipality area data point. The API defines RegionAreaDataPoint but
 * has no municipal twin; the municipal query returns this shape.
 */
export interface MunicipalityAreaDataPoint {
  mun_code: string;
  value_mnkr: number;
  total_mnkr: number;
  pct: number;
}
```

- [ ] **Step 2: Verify**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS (module exists, not yet imported).

- [ ] **Step 3: Commit**

```bash
git add frontend/src/shared/ui-types.ts
git commit -m "feat(types): add ui-types for frontend-only shapes"
```

---

### Task 3: Migrate `features/budget/api.ts`

**Files:**
- Modify: `frontend/src/features/budget/api.ts`

- [ ] **Step 1: Read the current file**

Run: `cat frontend/src/features/budget/api.ts`
It currently imports `AreaTimeSeries, BudgetComparison, BudgetYear, BudgetYearDetail, ExpenditureArea` from `@/shared/types`, all used as `api.get<T>(...)` generics.

- [ ] **Step 2: Replace the import**

Replace:
```ts
import type { AreaTimeSeries, BudgetComparison, BudgetYear, BudgetYearDetail, ExpenditureArea } from "@/shared/types";
```
with:
```ts
import type { components } from "@/shared/api-contract";

type AreaTimeSeries = components["schemas"]["AreaTimeSeries"];
type BudgetComparison = components["schemas"]["BudgetComparison"];
type BudgetYear = components["schemas"]["BudgetYear"];
type BudgetYearDetail = components["schemas"]["BudgetYearDetail"];
type ExpenditureArea = components["schemas"]["ExpenditureArea"];
```

- [ ] **Step 3: Verify**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/budget/api.ts
git commit -m "refactor(budget): type from generated contract, not shared/types"
```

---

### Task 4: Migrate `features/context/api.ts`

**Files:**
- Modify: `frontend/src/features/context/api.ts`

- [ ] **Step 1: Read the current file**

Run: `cat frontend/src/features/context/api.ts`
It imports `TopicContext` from `@/shared/types`, used as an `api.get<T>` generic.

- [ ] **Step 2: Replace the import**

Replace:
```ts
import type { TopicContext } from "@/shared/types";
```
with:
```ts
import type { components } from "@/shared/api-contract";

type TopicContext = components["schemas"]["TopicContext"];
```

- [ ] **Step 3: Verify**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/context/api.ts
git commit -m "refactor(context): type from generated contract"
```

---

### Task 5: Migrate `features/committees/api.ts`

**Files:**
- Modify: `frontend/src/features/committees/api.ts`

- [ ] **Step 1: Read the current file**

Run: `cat frontend/src/features/committees/api.ts`
It imports `Committee, CommitteeVoteringPage, GoalWithAlignment` from `@/shared/types`, all `api.get<T>` generics.

- [ ] **Step 2: Replace the import**

Replace:
```ts
import type {
  Committee,
  CommitteeVoteringPage,
  GoalWithAlignment,
} from "@/shared/types";
```
with:
```ts
import type { components } from "@/shared/api-contract";

type Committee = components["schemas"]["Committee"];
type CommitteeVoteringPage = components["schemas"]["CommitteeVoteringPage"];
type GoalWithAlignment = components["schemas"]["GoalWithAlignment"];
```

- [ ] **Step 3: Verify**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/committees/api.ts
git commit -m "refactor(committees): type from generated contract"
```

---

### Task 6: Migrate `features/politicians/api.ts`

**Files:**
- Modify: `frontend/src/features/politicians/api.ts`

- [ ] **Step 1: Read the current file**

Run: `cat frontend/src/features/politicians/api.ts`
It imports `PoliticianListResponse, PoliticianSummary, PromiseWithMatches, Topic, VoteListResponse` from `@/shared/types`. `Topic` is used as a param type `{ topic?: Topic }`; the rest are `api.get<T>` generics.

- [ ] **Step 2: Replace the import**

Replace:
```ts
import type { PoliticianListResponse, PoliticianSummary, PromiseWithMatches, Topic, VoteListResponse } from "@/shared/types";
```
with:
```ts
import type { components } from "@/shared/api-contract";

type PoliticianListResponse = components["schemas"]["PoliticianListResponse"];
type PoliticianSummary = components["schemas"]["PoliticianSummary"];
type PromiseWithMatches = components["schemas"]["PromiseWithMatches"];
type Topic = components["schemas"]["Topic"];
type VoteListResponse = components["schemas"]["VoteListResponse"];
```

- [ ] **Step 3: Verify**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/politicians/api.ts
git commit -m "refactor(politicians): type from generated contract"
```

---

### Task 7: Migrate `features/parties/api.ts`

**Files:**
- Modify: `frontend/src/features/parties/api.ts`

- [ ] **Step 1: Read the current file**

Run: `cat frontend/src/features/parties/api.ts`
It currently imports `GoalVoteBreakdown, GoalWithAlignment, PartySummary` from `@/shared/types`, AND already imports `components` from `@/shared/api-contract` for `PartyMeta`.

- [ ] **Step 2: Extend the existing contract import**

Add these type aliases after the existing `PartyMeta` type alias:
```ts
type GoalVoteBreakdown = components["schemas"]["GoalVoteBreakdown"];
type GoalWithAlignment = components["schemas"]["GoalWithAlignment"];
type PartySummary = components["schemas"]["PartySummary"];
```

- [ ] **Step 3: Remove the `@/shared/types` import**

Delete the line:
```ts
import type { GoalVoteBreakdown, GoalWithAlignment, PartySummary } from "@/shared/types";
```
(and add any alias types you still reference but did not previously define — ensure `GoalVoteBreakdown`, `GoalWithAlignment`, `PartySummary` are all now aliased as above.)

- [ ] **Step 4: Verify**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/features/parties/api.ts
git commit -m "refactor(parties): type from generated contract"
```

---

### Task 8: Migrate `features/votes/api.ts`

**Files:**
- Modify: `frontend/src/features/votes/api.ts`

- [ ] **Step 1: Read the current file**

Run: `cat frontend/src/features/votes/api.ts`
It imports `VoteDetail, VoteSummaryListResponse, RiksdagDocument, RiksdagDocumentFull` from `@/shared/types`, and already imports `components` from `@/shared/api-contract` for `RecentBetankande`.

- [ ] **Step 2: Extend the existing contract import**

Add these type aliases after the existing `RecentBetankande` alias:
```ts
type VoteDetail = components["schemas"]["VoteDetail"];
type VoteSummaryListResponse = components["schemas"]["VoteSummaryListResponse"];
type RiksdagDocument = components["schemas"]["RiksdagDocument"];
type RiksdagDocumentFull = components["schemas"]["RiksdagDocumentFull"];
```

- [ ] **Step 3: Remove the `@/shared/types` import**

Delete the line:
```ts
import type { VoteDetail, VoteSummaryListResponse, RiksdagDocument, RiksdagDocumentFull } from "@/shared/types";
```

- [ ] **Step 4: Verify**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/features/votes/api.ts
git commit -m "refactor(votes): type from generated contract"
```

---

### Task 9: Migrate `features/regions/api.ts`

**Files:**
- Modify: `frontend/src/features/regions/api.ts`

- [ ] **Step 1: Read the current file**

Run: `cat frontend/src/features/regions/api.ts`
It imports `KPIRank, MunicipalityKPIItem, RegionAreaDataPoint, RegionBudgetArea, RegionBudgetSnapshot, RegionDetail, RegionKPIRankEntry, RegionPlan, RegionSummary` from `@/shared/types`, all `api.get<T>` generics.

- [ ] **Step 2: Replace the import**

Replace:
```ts
import type { KPIRank, MunicipalityKPIItem, RegionAreaDataPoint, RegionBudgetArea, RegionBudgetSnapshot, RegionDetail, RegionKPIRankEntry, RegionPlan, RegionSummary } from "@/shared/types";
```
with:
```ts
import type { components } from "@/shared/api-contract";

type KPIRank = components["schemas"]["KPIRank"];
type MunicipalityKPIItem = components["schemas"]["MunicipalityKPIItem"];
type RegionAreaDataPoint = components["schemas"]["RegionAreaDataPoint"];
type RegionBudgetArea = components["schemas"]["RegionBudgetArea"];
type RegionBudgetSnapshot = components["schemas"]["RegionBudgetSnapshot"];
type RegionDetail = components["schemas"]["RegionDetail"];
type RegionKPIRankEntry = components["schemas"]["RegionKPIRankEntry"];
type RegionPlan = components["schemas"]["RegionPlan"];
type RegionSummary = components["schemas"]["RegionSummary"];
```

- [ ] **Step 3: Verify**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/regions/api.ts
git commit -m "refactor(regions): type from generated contract"
```

---

### Task 10: Migrate `features/municipalities/api.ts`

**Files:**
- Modify: `frontend/src/features/municipalities/api.ts`

- [ ] **Step 1: Read the current file**

Run: `cat frontend/src/features/municipalities/api.ts`
It imports `MunicipalityDetail, MunicipalitySummary, MunicipalityKPIItem, PopulationTrendEntry, ProcurementCategorySummary, MunicipalityBudgetSnapshot, MunicipalityAreaDataPoint` from `@/shared/types`, all `api.get<T>` generics.

- [ ] **Step 2: Replace the import — two sources**

`MunicipalityAreaDataPoint` has NO contract schema (frontend-only). The rest do. Replace:
```ts
import type { MunicipalityDetail, MunicipalitySummary, MunicipalityKPIItem, PopulationTrendEntry, ProcurementCategorySummary, MunicipalityBudgetSnapshot, MunicipalityAreaDataPoint } from "@/shared/types";
```
with:
```ts
import type { components } from "@/shared/api-contract";
import type { MunicipalityAreaDataPoint } from "@/shared/ui-types";

type MunicipalityDetail = components["schemas"]["MunicipalityDetail"];
type MunicipalitySummary = components["schemas"]["MunicipalitySummary"];
type MunicipalityKPIItem = components["schemas"]["MunicipalityKPIItem"];
type PopulationTrendEntry = components["schemas"]["PopulationTrendEntry"];
type ProcurementCategorySummary = components["schemas"]["ProcurementCategorySummary"];
type MunicipalityBudgetSnapshot = components["schemas"]["MunicipalityBudgetSnapshot"];
```

- [ ] **Step 3: Verify**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/municipalities/api.ts
git commit -m "refactor(municipalities): type from generated contract"
```

---

### Task 11: Migrate `hooks/useDemocracy.ts` (the 20-import hotspot)

**Files:**
- Modify: `frontend/src/hooks/useDemocracy.ts`

- [ ] **Step 1: Read the current import**

Run: `sed -n '1,12p' frontend/src/hooks/useDemocracy.ts`
It currently imports (types-only usage):
```ts
import type {
  RecordCoverage, ElectionResult, KPIRank, RegionSummary, MunicipalitySummary, MunicipalityKPIItem, RiksdagDocument, RiksdagDocumentFull, RegionBudgetSnapshot, RegionAreaDataPoint, RegionKPIRankEntry, RegionPlan, MunicipalityBudgetSnapshot, MunicipalityAreaDataPoint, BudgetSnapshot } from "@/shared/types";
```

- [ ] **Step 2: Split into two imports**

`BudgetSnapshot` and `MunicipalityAreaDataPoint` have no contract schema (frontend-only → `@/shared/ui-types`). Replace the single import with:

```ts
import type { components } from "@/shared/api-contract";
import type { BudgetSnapshot, MunicipalityAreaDataPoint } from "@/shared/ui-types";

type RecordCoverage = components["schemas"]["RecordCoverage"];
type ElectionResult = components["schemas"]["ElectionResult"];
type KPIRank = components["schemas"]["KPIRank"];
type RegionSummary = components["schemas"]["RegionSummary"];
type MunicipalitySummary = components["schemas"]["MunicipalitySummary"];
type MunicipalityKPIItem = components["schemas"]["MunicipalityKPIItem"];
type RiksdagDocument = components["schemas"]["RiksdagDocument"];
type RiksdagDocumentFull = components["schemas"]["RiksdagDocumentFull"];
type RegionBudgetSnapshot = components["schemas"]["RegionBudgetSnapshot"];
type RegionAreaDataPoint = components["schemas"]["RegionAreaDataPoint"];
type RegionKPIRankEntry = components["schemas"]["RegionKPIRankEntry"];
type RegionPlan = components["schemas"]["RegionPlan"];
type MunicipalityBudgetSnapshot = components["schemas"]["MunicipalityBudgetSnapshot"];
```

- [ ] **Step 3: Check for name collisions**

Run: `grep -n "^type \|^function \|^const \|^export " frontend/src/hooks/useDemocracy.ts | grep -iE "BudgetSnapshot|MunicipalityAreaDataPoint|RecordCoverage|ElectionResult|KPIRank|RegionSummary|MunicipalitySummary|MunicipalityKPIItem|RiksdagDocument|RegionBudgetSnapshot|RegionAreaDataPoint|RegionKPIRankEntry|RegionPlan|MunicipalityBudgetSnapshot"`
If any of these names is already declared locally in the file (a function/const with the same name), the `type X = ...` alias will collide — rename the alias instead. If no collision, proceed.

- [ ] **Step 4: Verify**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/hooks/useDemocracy.ts
git commit -m "refactor(useDemocracy): type from generated contract"
```

---

### Task 12: Delete `shared/types.ts` and fix remaining stragglers

**Files:**
- Delete: `frontend/src/shared/types.ts`
- Modify: any remaining file that imports from `@/shared/types` (found in Step 2)

- [ ] **Step 1: Find every remaining importer**

Run: `rg 'shared/types' frontend/src --glob '!shared/types.ts'`
Expected output: a list of every file still importing from `@/shared/types`.

- [ ] **Step 2: For each remaining importer, migrate it**

Each file's imports break the same way as Tasks 3–11: contract-backed types → `@/shared/api-contract` (`components["schemas"]["X"]`); the three frontend-only types (`BudgetTier`, `BudgetSnapshot`, `MunicipalityAreaDataPoint`) → `@/shared/ui-types`; and `PoliticianPromise` → the contract's `Promise` schema (rename the reference).

For each file:
1. `cat <file>` to read its exact `@/shared/types` import.
2. Replace with the `@/shared/api-contract` (`import type { components }`) + alias lines for contract types, or `@/shared/ui-types` for the three frontend-only types, or the contract `Promise` schema for `PoliticianPromise`.
3. Run `cd frontend && npx tsc --noEmit`; fix any error the migration surfaces (a missing type means it has no schema → decide if it goes in `ui-types.ts`).

**Known frontend-only types that must go to `@/shared/ui-types` (if referenced):** `BudgetTier` (used by `features/budget/components/TierNav.tsx`), `BudgetSnapshot` (used by `features/budget/components/BudgetHistorySection.tsx`), `MunicipalityAreaDataPoint` (used by `features/municipalities/api.ts`). The `BudgetTier`/`BudgetSnapshot`/`MunicipalityAreaDataPoint` definitions already live in `ui-types.ts` from Task 2.

**Known name-drift:** `PoliticianPromise` is the contract's `Promise`. Any file referencing `PoliticianPromise` must import the contract `Promise` and use that name.

- [ ] **Step 3: Delete the mirror**

After every importer is migrated (Step 2 returns no files), delete the file:

```bash
rm frontend/src/shared/types.ts
```

- [ ] **Step 4: Full verification**

Run:
```bash
cd frontend && npx tsc --noEmit && npm run lint && npm run build
rg 'shared/types' frontend/src
```
Expected: `tsc`, `lint`, `build` all pass; `rg 'shared/types'` returns NOTHING.

Also run the backend gate:
```bash
go build -C backend ./...
```
Expected: PASS (no backend change, but the gate must stay green).

- [ ] **Step 5: Commit**

```bash
git add -A frontend/src/shared
git add frontend/src/hooks frontend/src/features
git commit -m "refactor(types): delete shared/types.ts — contract is the single source of truth"
```

---

## Self-review notes

- **Spec coverage:** Task 1 = spec section A (document `GET /votes`); Task 2 = spec section B (create `ui-types.ts`); Tasks 3–11 + 12 = spec section C/D (migrate consumers, delete mirror). All four spec sections are covered.
- **Type consistency:** `VoteSummary`/`VoteSummaryListResponse` added in Task 1 are consumed in Task 8; `BudgetSnapshot`/`MunicipalityAreaDataPoint` created in Task 2 are consumed in Tasks 10–12; `BudgetTier` created in Task 2 is consumed in Task 12. The `PoliticianPromise` → `Promise` rename is handled in Task 12 with a specific note.
- **Placeholder scan:** No TBD/TODO. Every code step shows the exact import block to write. Task 12 is deliberately open-ended (depends on the unknown straggler set) but names the three known frontend-only types and the known name-drift, so it is not a blind "fix whatever" — it gives the discriminating rules.