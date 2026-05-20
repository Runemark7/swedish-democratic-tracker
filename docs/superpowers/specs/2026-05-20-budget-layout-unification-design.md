# Budget Layout Unification

**Date:** 2026-05-20
**Status:** Approved

## Goal

Unify the budget section layout across region and municipality detail pages. The standalone `/budget` national page stays unique. All other budget sections share one component and one layout.

## Decisions

- Layout direction: year tabs + delta table (region's existing pattern)
- Position: full-width section below the existing top grid (not in the 2-col right slot)
- Municipality gets multi-year backend data (not stubbed to single year)

---

## Component API

**File:** `frontend/src/features/budget/components/BudgetHistorySection.tsx`

```typescript
interface BudgetHistorySectionProps {
  snapshots: RegionBudgetSnapshot[];     // flat list — grouped by year internally
  makeAreaLink?: (areaName: string) => string;  // undefined = rows not clickable
  emptyMessage?: string;
  sourceId: string;
}
```

- Owns: year tab state, grouping by year, delta calculations, total row, empty state
- Reuses: existing `DeltaIndicator`, `BUDGET_COLORS` (moved here from both pages)
- Region passes `makeAreaLink` → `/region/:code/budget/:area`
- Municipality passes `makeAreaLink` → `/municipality/:code/budget/:area` (or omits if no drill-down page exists yet)

---

## Page Layout

Both `RegionDetailPage` and `MunicipalityDetailPage`:

1. Top 2-col grid: right "budget" slot removed. Political composition / mandate section expands to full width.
2. New full-width surface card appended below the top grid, containing `<BudgetHistorySection>`.
3. Mobile: already single-col — no breakpoint changes needed.

---

## Backend

### Migration 000012

New table `municipality_budget_snapshots` — same shape as `region_budget_snapshots`:

```sql
code       TEXT NOT NULL,
year       INT  NOT NULL,
area_name  TEXT NOT NULL,
value_mnkr FLOAT NOT NULL,
total_mnkr FLOAT NOT NULL
```

Verify exact column names against `region_budget_snapshots` before writing migration.

### Ingestion worker

**File:** `backend/internal/ingestion/workers/municipality_budget.go`

- Fetches Kolada KPIs already used for municipality spending:
  `N11004, N15028, N17014, N20014, N30005, N07037, N09022, N05011, N45014`
- Years: 2020 to current year
- Upserts into `municipality_budget_snapshots`
- Registered in ingestion scheduler `@weekly`

### HTTP handler

**Path:** `GET /api/municipalities/{code}/budget/history`
**Feature:** `internal/municipalities/adapters/http/`
**Response:** `[]MunicipalityBudgetSnapshot` — same JSON shape as region budget history

### Spec / types

- Add path + response schema to `api/openapi.yaml`
- Run `cd frontend && npm run generate:api` to regenerate `api-contract.ts`

---

## Frontend Data Flow

### New hook

```typescript
// frontend/src/hooks/useDemocracy.ts
useKommunBudgetHistory(code: string) // → RegionBudgetSnapshot[]
```

Calls `GET /api/municipalities/:code/budget/history`. Same query shape as `useRegionBudgetHistory`.

### RegionDetailPage changes

- Remove: `budgetTab` state, `historyByYear` map, `activeBudgetYear` calc, inline budget tab/table JSX (~150 lines)
- Remove: `BUDGET_COLORS` constant
- Add: `<BudgetHistorySection snapshots={budgetHistory} makeAreaLink={a => `/region/${code}/budget/${encodeURIComponent(a)}`} sourceId="scb-kostndrlt" />` below top grid

### MunicipalityDetailPage changes

- Remove: `donutSegments`, `hbarsItems`, `Donut`, `HBars` budget vars, `BUDGET_COLORS` constant
- Remove: budget rendering from top grid right slot
- Add: `useKommunBudgetHistory(code)` call
- Add: `<BudgetHistorySection snapshots={kommunBudgetHistory} sourceId="kolada-spending" />` below top grid

---

## File Checklist

| File | Change |
|------|--------|
| `backend/migrations/000012_municipality_budget_snapshots.up.sql` | new |
| `backend/internal/municipalities/domain/` | add `MunicipalityBudgetSnapshot` entity |
| `backend/internal/municipalities/ports/` | add repository port |
| `backend/internal/municipalities/adapters/postgres/` | SQL impl |
| `backend/internal/municipalities/adapters/http/` | GET handler |
| `backend/internal/ingestion/workers/municipality_budget.go` | new worker |
| `backend/cmd/api/main.go` | wire worker + route |
| `api/openapi.yaml` | new path + schema |
| `frontend/src/shared/api-contract.ts` | regenerated |
| `frontend/src/hooks/useDemocracy.ts` | add `useKommunBudgetHistory` |
| `frontend/src/features/budget/components/BudgetHistorySection.tsx` | new |
| `frontend/src/features/regions/RegionDetailPage.tsx` | replace budget section |
| `frontend/src/features/municipalities/MunicipalityDetailPage.tsx` | replace budget section |
