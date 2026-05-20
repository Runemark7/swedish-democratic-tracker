# Budget Layout Unification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Unify the budget section on RegionDetailPage and MunicipalityDetailPage into one shared `BudgetHistorySection` component (year tabs + delta table, full-width below the top grid), backed by a new multi-year municipality budget history endpoint.

**Architecture:** The `regions` feature (Go) already owns municipalities. We add a `municipality_budget_snapshots` table, a new ingestion worker that converts Kolada `kr/inv` × population into mnkr snapshots, and a `GET /api/municipalities/{code}/budget/history` handler. On the frontend a shared `BudgetHistorySection` component replaces the inline JSX in both detail pages.

**Tech Stack:** Go 1.26 · pgx/v5 · golang-migrate · chi · React 19 · TypeScript 5 · TanStack Query v5 · Tailwind CSS v4

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `backend/migrations/000019_municipality_budget_snapshots.up.sql` | create | Table DDL |
| `backend/migrations/000019_municipality_budget_snapshots.down.sql` | create | Drop table |
| `backend/internal/regions/ports/external.go` | modify | Add `MunicipalityBudgetSnapshot` type + repo port methods |
| `backend/internal/regions/adapters/postgres/repository.go` | modify | SQL impl for upsert + history |
| `backend/internal/regions/adapters/postgres/budget_repo_test.go` | modify | Integration tests for municipality table |
| `backend/internal/regions/service.go` | modify | Service methods for municipality budget |
| `backend/internal/regions/adapters/http/handler.go` | modify | New route + handler |
| `backend/internal/ingestion/workers/municipality_budget.go` | create | Ingestion worker |
| `backend/cmd/api/main.go` | modify | Wire worker + (no new route reg needed — handler.go owns its routes) |
| `api/openapi.yaml` | modify | New path + schema |
| `frontend/src/shared/types.ts` | modify | Add `BudgetSnapshot` interface |
| `frontend/src/features/municipalities/api.ts` | modify | Add `getMunicipalityBudgetHistory` |
| `frontend/src/hooks/useDemocracy.ts` | modify | Add `useKommunBudgetHistory` |
| `frontend/src/features/budget/components/BudgetHistorySection.tsx` | create | Shared component |
| `frontend/src/features/regions/RegionDetailPage.tsx` | modify | Use BudgetHistorySection, remove right budget slot |
| `frontend/src/features/municipalities/MunicipalityDetailPage.tsx` | modify | Use BudgetHistorySection, remove donut+hbars budget |

---

## Task 1: DB migration

**Files:**
- Create: `backend/migrations/000019_municipality_budget_snapshots.up.sql`
- Create: `backend/migrations/000019_municipality_budget_snapshots.down.sql`

- [ ] **Step 1: Write up migration**

```sql
-- backend/migrations/000019_municipality_budget_snapshots.up.sql
CREATE TABLE municipality_budget_snapshots (
    mun_code    TEXT        NOT NULL,
    area_name   TEXT        NOT NULL,
    year        INT         NOT NULL,
    value_mnkr  NUMERIC     NOT NULL,
    total_mnkr  NUMERIC     NOT NULL,
    pct         NUMERIC     NOT NULL,
    fetched_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (mun_code, area_name, year)
);

CREATE INDEX ON municipality_budget_snapshots (mun_code);
CREATE INDEX ON municipality_budget_snapshots (area_name, year);
```

- [ ] **Step 2: Write down migration**

```sql
-- backend/migrations/000019_municipality_budget_snapshots.down.sql
DROP TABLE IF EXISTS municipality_budget_snapshots;
```

- [ ] **Step 3: Run migration**

```bash
make migrate
```

Expected: migration `000019` applied with no errors.

- [ ] **Step 4: Commit**

```bash
git add backend/migrations/
git commit -m "feat(db): add municipality_budget_snapshots table"
```

---

## Task 2: Backend ports — new type + repository interface

**Files:**
- Modify: `backend/internal/regions/ports/external.go`

- [ ] **Step 1: Add `MunicipalityBudgetSnapshot` type and two new methods to `RegionRepository`**

In `backend/internal/regions/ports/external.go`, add after the `RegionBudgetSnapshot` block:

```go
type MunicipalityBudgetSnapshot struct {
	MunCode   string  `json:"mun_code"`
	AreaName  string  `json:"area_name"`
	Year      int     `json:"year"`
	ValueMnkr float64 `json:"value_mnkr"`
	TotalMnkr float64 `json:"total_mnkr"`
	Pct       float64 `json:"pct"`
}
```

In `backend/internal/regions/ports/repository.go`, add two methods to `RegionRepository`:

```go
UpsertMunicipalityBudgetSnapshots(ctx context.Context, snapshots []MunicipalityBudgetSnapshot) (int, error)
GetMunicipalityBudgetHistory(ctx context.Context, munCode string, years []int) ([]MunicipalityBudgetSnapshot, error)
```

- [ ] **Step 2: Verify build fails as expected** (interface not satisfied yet)

```bash
go build -C backend ./...
```

Expected: compile error — `postgres.Repository does not implement ports.RegionRepository`.

---

## Task 3: Postgres adapter — upsert + history

**Files:**
- Modify: `backend/internal/regions/adapters/postgres/repository.go`
- Modify: `backend/internal/regions/adapters/postgres/budget_repo_test.go`

- [ ] **Step 1: Add `UpsertMunicipalityBudgetSnapshots`**

Append to `backend/internal/regions/adapters/postgres/repository.go`:

```go
func (r *Repository) UpsertMunicipalityBudgetSnapshots(ctx context.Context, snapshots []ports.MunicipalityBudgetSnapshot) (int, error) {
	if len(snapshots) == 0 {
		return 0, nil
	}
	const q = `
		INSERT INTO municipality_budget_snapshots (mun_code, area_name, year, value_mnkr, total_mnkr, pct, fetched_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (mun_code, area_name, year) DO UPDATE SET
			value_mnkr = EXCLUDED.value_mnkr,
			total_mnkr = EXCLUDED.total_mnkr,
			pct        = EXCLUDED.pct,
			fetched_at = NOW()
	`
	var count int
	for _, s := range snapshots {
		tag, err := r.db.Exec(ctx, q, s.MunCode, s.AreaName, s.Year, s.ValueMnkr, s.TotalMnkr, s.Pct)
		if err != nil {
			return count, err
		}
		count += int(tag.RowsAffected())
	}
	return count, nil
}
```

- [ ] **Step 2: Add `GetMunicipalityBudgetHistory`**

Append to `backend/internal/regions/adapters/postgres/repository.go`:

```go
func (r *Repository) GetMunicipalityBudgetHistory(ctx context.Context, munCode string, years []int) ([]ports.MunicipalityBudgetSnapshot, error) {
	rows, err := r.db.Query(ctx, `
		SELECT mun_code, area_name, year, value_mnkr, total_mnkr, pct
		FROM municipality_budget_snapshots
		WHERE mun_code = $1 AND year = ANY($2)
		ORDER BY year, area_name
	`, munCode, years)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ports.MunicipalityBudgetSnapshot
	for rows.Next() {
		var s ports.MunicipalityBudgetSnapshot
		if err := rows.Scan(&s.MunCode, &s.AreaName, &s.Year, &s.ValueMnkr, &s.TotalMnkr, &s.Pct); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
```

- [ ] **Step 3: Write integration tests**

Add to `backend/internal/regions/adapters/postgres/budget_repo_test.go` (after the existing tests):

```go
func cleanMunBudgetSnapshots(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `DELETE FROM municipality_budget_snapshots WHERE mun_code LIKE 'ZZ%'`)
	if err != nil {
		t.Fatalf("clean mun snapshots: %v", err)
	}
}

func TestUpsertMunicipalityBudgetSnapshots_Idempotency(t *testing.T) {
	pool := connectTestDB(t)
	cleanMunBudgetSnapshots(t, pool)
	t.Cleanup(func() { cleanMunBudgetSnapshots(t, pool) })

	repo := regionsPG.NewRepository(pool)
	snaps := []ports.MunicipalityBudgetSnapshot{
		{MunCode: "ZZ00", AreaName: "Grundskola", Year: 2023, ValueMnkr: 500, TotalMnkr: 2000, Pct: 25.0},
		{MunCode: "ZZ00", AreaName: "Förskola",   Year: 2023, ValueMnkr: 300, TotalMnkr: 2000, Pct: 15.0},
	}

	n, err := repo.UpsertMunicipalityBudgetSnapshots(context.Background(), snaps)
	if err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 rows, got %d", n)
	}

	// Second upsert with changed value — should update, not duplicate.
	snaps[0].ValueMnkr = 510
	n2, err := repo.UpsertMunicipalityBudgetSnapshots(context.Background(), snaps)
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	if n2 != 2 {
		t.Fatalf("expected 2 rows on re-upsert, got %d", n2)
	}

	history, err := repo.GetMunicipalityBudgetHistory(context.Background(), "ZZ00", []int{2023})
	if err != nil {
		t.Fatalf("get history: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 history rows, got %d", len(history))
	}
	grundskola := history[0] // ordered by area_name
	if grundskola.ValueMnkr != 510 {
		t.Errorf("expected updated ValueMnkr=510, got %v", grundskola.ValueMnkr)
	}
}
```

- [ ] **Step 4: Run tests (requires TEST_DATABASE_URL)**

```bash
TEST_DATABASE_URL=postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable go test ./internal/regions/adapters/postgres/... -v -run TestUpsertMunicipalityBudgetSnapshots
```

Expected: PASS (or SKIP if TEST_DATABASE_URL not set).

- [ ] **Step 5: Build check**

```bash
go build -C backend ./...
```

Expected: success.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/regions/
git commit -m "feat(regions): municipality budget snapshot repo + integration tests"
```

---

## Task 4: Service methods

**Files:**
- Modify: `backend/internal/regions/service.go`

- [ ] **Step 1: Add spending KPI name map and service methods**

In `backend/internal/regions/service.go`, after the `spendingKPIs` var, add:

```go
// spendingKPINames maps Kolada KPI codes to human-readable Swedish area names.
var spendingKPINames = map[string]string{
	"N11004": "Förskola",
	"N15028": "Grundskola",
	"N17014": "Gymnasieskola",
	"N20014": "Äldreomsorg",
	"N30005": "Individ & familj",
	"N07037": "Gata, park, plan",
	"N09022": "Fritid & kultur",
	"N05011": "Politisk verksamhet",
	"N45014": "Vatten & avlopp",
}
```

Then add these three methods to `Service` (after `GetMunicipalitySpending`):

```go
// GetMunicipalityBudgetMultiYear fetches Kolada spending KPIs for multiple years,
// multiplies kr/inv by population to derive mnkr totals, and returns snapshots.
func (s *Service) GetMunicipalityBudgetMultiYear(ctx context.Context, munCode string, population int, years []int) ([]ports.MunicipalityBudgetSnapshot, error) {
	kpis, err := s.kolada.FetchKPIs(ctx, munCode, spendingKPIs, years)
	if err != nil {
		return nil, err
	}

	// Group by year → kpiCode → value.
	byYear := map[int]map[string]float64{}
	for _, k := range kpis {
		if byYear[k.Year] == nil {
			byYear[k.Year] = map[string]float64{}
		}
		byYear[k.Year][k.KPI] = k.Value
	}

	var snapshots []ports.MunicipalityBudgetSnapshot
	for _, year := range years {
		kpiMap, ok := byYear[year]
		if !ok {
			continue
		}
		var totalMnkr float64
		type pair struct{ code, name string }
		var areas []pair
		for _, code := range spendingKPIs {
			if v, ok := kpiMap[code]; ok && v > 0 {
				areas = append(areas, pair{code, spendingKPINames[code]})
				totalMnkr += (v * float64(population)) / 1_000_000
			}
		}
		if totalMnkr == 0 {
			continue
		}
		for _, a := range areas {
			v := kpiMap[a.code]
			valueMnkr := (v * float64(population)) / 1_000_000
			pct := math.Round((valueMnkr/totalMnkr)*1000) / 10
			snapshots = append(snapshots, ports.MunicipalityBudgetSnapshot{
				MunCode:   munCode,
				AreaName:  a.name,
				Year:      year,
				ValueMnkr: math.Round(valueMnkr*10) / 10,
				TotalMnkr: math.Round(totalMnkr*10) / 10,
				Pct:       pct,
			})
		}
	}
	return snapshots, nil
}

func (s *Service) UpsertMunicipalityBudgetSnapshots(ctx context.Context, snapshots []ports.MunicipalityBudgetSnapshot) (int, error) {
	return s.repo.UpsertMunicipalityBudgetSnapshots(ctx, snapshots)
}

func (s *Service) GetMunicipalityBudgetHistory(ctx context.Context, munCode string, years []int) ([]ports.MunicipalityBudgetSnapshot, error) {
	return s.repo.GetMunicipalityBudgetHistory(ctx, munCode, years)
}
```

Add `"math"` to the imports of service.go.

- [ ] **Step 2: Build check**

```bash
go build -C backend ./...
```

Expected: success.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/regions/service.go
git commit -m "feat(regions): municipality budget multi-year service methods"
```

---

## Task 5: HTTP handler — new route

**Files:**
- Modify: `backend/internal/regions/adapters/http/handler.go`

- [ ] **Step 1: Register route**

In `handler.go`, add to the `Routes` method after `r.Get("/municipalities/{code}/spending", h.getMunicipalitySpending)`:

```go
r.Get("/municipalities/{code}/budget/history", h.getMunicipalityBudgetHistory)
```

- [ ] **Step 2: Add handler function**

Append to `handler.go`:

```go
func (h *Handler) getMunicipalityBudgetHistory(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	yearsParam := r.URL.Query().Get("years")
	n := 4
	if yearsParam != "" {
		if parsed, err := strconv.Atoi(yearsParam); err == nil && parsed > 0 {
			n = parsed
		}
	}
	current := time.Now().Year()
	years := make([]int, n)
	for i := range years {
		years[i] = current - 1 - i
	}
	snapshots, err := h.svc.GetMunicipalityBudgetHistory(r.Context(), code, years)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if snapshots == nil {
		snapshots = []ports.MunicipalityBudgetSnapshot{}
	}
	jsonOK(w, snapshots)
}
```

- [ ] **Step 3: Build check**

```bash
go build -C backend ./...
```

Expected: success.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/regions/adapters/http/handler.go
git commit -m "feat(regions): GET /municipalities/{code}/budget/history endpoint"
```

---

## Task 6: Ingestion worker + wiring

**Files:**
- Create: `backend/internal/ingestion/workers/municipality_budget.go`
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: Write the worker**

```go
// backend/internal/ingestion/workers/municipality_budget.go
package workers

import (
	"context"
	"log/slog"

	ingPorts "riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/regions"
)

// MunicipalityBudgetWorker fetches Kolada spending KPIs for all tracked municipalities
// across the last 4 completed years, converts kr/inv × population to mnkr snapshots,
// and upserts them into municipality_budget_snapshots.
type MunicipalityBudgetWorker struct {
	regionSvc *regions.Service
	runRepo   ingPorts.IngestionRunRepository
}

func NewMunicipalityBudgetWorker(regionSvc *regions.Service, runRepo ingPorts.IngestionRunRepository) MunicipalityBudgetWorker {
	return MunicipalityBudgetWorker{regionSvc: regionSvc, runRepo: runRepo}
}

func (w *MunicipalityBudgetWorker) Name() string { return "municipality-budget" }

func (w *MunicipalityBudgetWorker) Run(ctx context.Context) error {
	runID, err := w.runRepo.StartRun(ctx, w.Name())
	if err != nil {
		slog.Error("municipality-budget: failed to start run record", "error", err)
	}

	totalRows, runErr := w.doRun(ctx)

	if runID != 0 {
		if ferr := w.runRepo.FinishRun(ctx, runID, totalRows, runErr); ferr != nil {
			slog.Warn("municipality-budget: failed to finish run record", "error", ferr)
		}
	}
	return runErr
}

func (w *MunicipalityBudgetWorker) doRun(ctx context.Context) (int, error) {
	// List all municipalities (empty regionCode = all).
	muns, err := w.regionSvc.ListMunicipalities(ctx, "")
	if err != nil {
		return 0, err
	}

	years := rollingYearsWorker(4)
	var totalRows int

	for _, mun := range muns {
		if ctx.Err() != nil {
			return totalRows, ctx.Err()
		}

		snapshots, err := w.regionSvc.GetMunicipalityBudgetMultiYear(ctx, mun.Code, mun.Population, years)
		if err != nil {
			slog.Warn("municipality-budget: fetch failed", "mun", mun.Code, "error", err)
			continue
		}
		if len(snapshots) == 0 {
			continue
		}

		n, err := w.regionSvc.UpsertMunicipalityBudgetSnapshots(ctx, snapshots)
		if err != nil {
			return totalRows, err
		}
		totalRows += n
		slog.Info("municipality-budget: upserted snapshots", "mun", mun.Code, "rows", n)
	}

	return totalRows, nil
}

var _ interface {
	Name() string
	Run(ctx context.Context) error
} = (*MunicipalityBudgetWorker)(nil)
```

- [ ] **Step 2: Wire the worker in `main.go`**

In `backend/cmd/api/main.go`, after the `regionBudgetWorker` registration block (around line 244), add:

```go
munBudgetWorker := workers.NewMunicipalityBudgetWorker(regionsSvc, ingestionRunsRepo)
if err := sched.RegisterSync("@weekly", &munBudgetWorker); err != nil {
    slog.Error("failed to register municipality-budget worker", "error", err)
    os.Exit(1)
}
```

- [ ] **Step 3: Build check**

```bash
go build -C backend ./...
```

Expected: success.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/ingestion/workers/municipality_budget.go backend/cmd/api/main.go
git commit -m "feat(ingestion): municipality-budget worker @weekly"
```

---

## Task 7: OpenAPI spec + API client function

**Files:**
- Modify: `api/openapi.yaml`
- Modify: `frontend/src/features/municipalities/api.ts`
- Modify: `frontend/src/shared/api-contract.ts` (regenerated)

- [ ] **Step 1: Add schema to openapi.yaml**

In `api/openapi.yaml`, add under `components.schemas` (after `RegionBudgetSnapshot`):

```yaml
    MunicipalityBudgetSnapshot:
      type: object
      required: [mun_code, area_name, year, value_mnkr, total_mnkr, pct]
      properties:
        mun_code:
          type: string
        area_name:
          type: string
        year:
          type: integer
        value_mnkr:
          type: number
        total_mnkr:
          type: number
        pct:
          type: number
```

- [ ] **Step 2: Add path to openapi.yaml**

In `api/openapi.yaml`, add after the `/municipalities/{code}/spending` path:

```yaml
  /municipalities/{code}/budget/history:
    get:
      operationId: getMunicipalityBudgetHistory
      tags: [municipalities]
      summary: Multi-year budget history for a municipality
      parameters:
        - name: code
          in: path
          required: true
          schema:
            type: string
        - name: years
          in: query
          schema:
            type: integer
            default: 4
      responses:
        "200":
          description: Array of budget snapshots across years
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/MunicipalityBudgetSnapshot"
```

- [ ] **Step 3: Regenerate api-contract.ts**

```bash
cd frontend && npm run generate:api
```

Expected: `src/shared/api-contract.ts` updated with `MunicipalityBudgetSnapshot` type.

- [ ] **Step 4: Add API function to municipalities/api.ts**

In `frontend/src/features/municipalities/api.ts`, add to the imports:

```typescript
import type { MunicipalityBudgetSnapshot, MunicipalityDetail, MunicipalitySummary, MunicipalityKPIItem, PopulationTrendEntry, ProcurementCategorySummary } from "@/shared/types";
```

Add to `municipalitiesApi`:

```typescript
  getMunicipalityBudgetHistory: (code: string, years = 4) =>
    api.get<MunicipalityBudgetSnapshot[]>(`/municipalities/${code}/budget/history?years=${years}`),
```

- [ ] **Step 5: Add `MunicipalityBudgetSnapshot` to shared/types.ts**

In `frontend/src/shared/types.ts`, after `RegionBudgetSnapshot`, add:

```typescript
export interface MunicipalityBudgetSnapshot {
  mun_code: string;
  area_name: string;
  year: number;
  value_mnkr: number;
  total_mnkr: number;
  pct: number;
}
```

Also add `BudgetSnapshot` (the minimal interface the shared component will use):

```typescript
export interface BudgetSnapshot {
  area_name: string;
  year: number;
  value_mnkr: number;
  total_mnkr: number;
  pct: number;
}
```

- [ ] **Step 6: Type check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 7: Commit**

```bash
git add api/openapi.yaml frontend/src/shared/api-contract.ts frontend/src/features/municipalities/api.ts frontend/src/shared/types.ts
git commit -m "feat(api): municipality budget history endpoint + types"
```

---

## Task 8: Frontend hook

**Files:**
- Modify: `frontend/src/hooks/useDemocracy.ts`

- [ ] **Step 1: Add `useKommunBudgetHistory`**

In `frontend/src/hooks/useDemocracy.ts`, add the import for `MunicipalityBudgetSnapshot` to the types import block, then add after `useRegionBudgetHistory`:

```typescript
// ── Municipality budget history ───────────────────────────────────────────────
export function useKommunBudgetHistory(code: string) {
  return useQuery<MunicipalityBudgetSnapshot[]>({
    queryKey: ["kommun-budget-history", code],
    queryFn: () => municipalitiesApi.getMunicipalityBudgetHistory(code),
    staleTime: 300_000,
    enabled: !!code,
  });
}
```

- [ ] **Step 2: Type check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/hooks/useDemocracy.ts
git commit -m "feat(hooks): useKommunBudgetHistory"
```

---

## Task 9: BudgetHistorySection component

**Files:**
- Create: `frontend/src/features/budget/components/BudgetHistorySection.tsx`

- [ ] **Step 1: Create the component**

```typescript
// frontend/src/features/budget/components/BudgetHistorySection.tsx
import { useState } from "react";
import { Link } from "react-router-dom";
import { DeltaIndicator } from "./DeltaIndicator";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import type { BudgetSnapshot } from "@/shared/types";

const BUDGET_COLORS = [
  "#0b3d7a",
  "#2d6fa8",
  "#5a9fd0",
  "#8bc0e0",
  "#c8a13b",
  "#d0533f",
  "#7a8390",
  "#b3bcc5",
];

interface BudgetHistorySectionProps {
  snapshots: BudgetSnapshot[];
  makeAreaLink?: (areaName: string) => string;
  emptyMessage?: string;
  sourceId: string;
}

export function BudgetHistorySection({
  snapshots,
  makeAreaLink,
  emptyMessage = "Budgetdata saknas.",
  sourceId,
}: BudgetHistorySectionProps) {
  const isMobile = useMediaQuery("(max-width: 640px)");

  // Group snapshots by year.
  const historyByYear = new Map<number, BudgetSnapshot[]>();
  for (const snap of snapshots) {
    const list = historyByYear.get(snap.year) ?? [];
    list.push(snap);
    historyByYear.set(snap.year, list);
  }
  const sortedYears = [...historyByYear.keys()].sort((a, b) => b - a);

  const [activeTab, setActiveTab] = useState<number | null>(null);
  const activeYear = activeTab ?? sortedYears[0] ?? null;

  return (
    <div style={{ background: "var(--color-sdt-surface)", padding: isMobile ? 16 : 24 }}>
      {/* Section label */}
      <div
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 10,
          letterSpacing: "0.15em",
          color: "var(--color-fg-muted)",
          marginBottom: 12,
        }}
      >
        BUDGET · HISTORIK
        <SourceMarker sourceId={sourceId} />
      </div>

      {sortedYears.length === 0 ? (
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-fg-muted)",
            padding: "24px 0",
            textAlign: "center",
            lineHeight: 1.5,
          }}
        >
          {emptyMessage}
        </div>
      ) : (
        <>
          {/* Year tabs */}
          <div style={{ display: "flex", gap: 4, flexWrap: "wrap", marginBottom: 16 }}>
            {sortedYears.map((yr, idx) => {
              const prevYear = sortedYears[idx + 1] ?? null;
              const prevSnaps = prevYear != null ? (historyByYear.get(prevYear) ?? []) : null;
              const curTotal = (historyByYear.get(yr) ?? []).reduce((s, a) => s + a.value_mnkr, 0);
              const prevTotal = prevSnaps ? prevSnaps.reduce((s, a) => s + a.value_mnkr, 0) : null;
              const delta =
                prevTotal != null && prevTotal > 0
                  ? ((curTotal - prevTotal) / prevTotal) * 100
                  : null;
              const isActive = yr === activeYear;
              const tabDeltaColor =
                delta == null
                  ? "var(--color-fg-muted)"
                  : delta >= 0
                  ? "#4caf7d"
                  : "#e05c5c";

              return (
                <button
                  key={yr}
                  onClick={() => setActiveTab(yr)}
                  className="px-3 py-1.5 text-xs font-mono font-bold rounded-md transition-all"
                  style={{
                    background: isActive
                      ? "var(--color-primary)"
                      : "var(--color-surface-low)",
                    color: isActive
                      ? "var(--color-on-primary)"
                      : "var(--color-on-surface)",
                  }}
                >
                  {yr}
                  {delta != null && (
                    <span
                      style={{
                        marginLeft: 4,
                        fontSize: 10,
                        color: isActive ? "var(--color-on-primary)" : tabDeltaColor,
                      }}
                    >
                      {delta >= 0 ? "+" : ""}
                      {delta.toFixed(1)}%
                    </span>
                  )}
                </button>
              );
            })}
          </div>

          {/* Active year table */}
          {activeYear != null &&
            (() => {
              const areas = historyByYear.get(activeYear) ?? [];
              const prevYear =
                sortedYears[sortedYears.indexOf(activeYear) + 1] ?? null;
              const prevAreas =
                prevYear != null ? (historyByYear.get(prevYear) ?? []) : [];
              const prevByName = new Map(prevAreas.map((a) => [a.area_name, a]));

              return (
                <div
                  style={{
                    borderRadius: 8,
                    overflow: "hidden",
                    border: "1px solid var(--color-surface-high)",
                  }}
                >
                  {/* Header row */}
                  <div
                    style={{
                      display: "grid",
                      gridTemplateColumns: "1fr auto auto",
                      gap: 8,
                      padding: "6px 12px",
                      background: "var(--color-surface-low)",
                      fontFamily: "var(--font-mono)",
                      fontSize: 9,
                      letterSpacing: "0.12em",
                      color: "var(--color-on-surface-variant)",
                      textTransform: "uppercase",
                    }}
                  >
                    <span>Område</span>
                    <span style={{ textAlign: "right", minWidth: 60 }}>
                      {prevYear ?? "—"}
                    </span>
                    <span style={{ textAlign: "right", minWidth: 80 }}>
                      Förändring
                    </span>
                  </div>

                  {/* Area rows */}
                  {areas.map((a, i) => {
                    const prev = prevByName.get(a.area_name);
                    const deltaPct =
                      prev && prev.value_mnkr > 0
                        ? ((a.value_mnkr - prev.value_mnkr) / prev.value_mnkr) * 100
                        : null;

                    const rowStyle = {
                      display: "grid",
                      gridTemplateColumns: "1fr auto auto",
                      gap: 8,
                      alignItems: "center",
                      padding: "8px 12px",
                      textDecoration: "none" as const,
                      color: "inherit",
                      borderTop: "1px solid var(--color-surface-high)",
                      background: "var(--color-sdt-surface)",
                    };

                    const inner = (
                      <>
                        <div
                          style={{
                            display: "flex",
                            alignItems: "center",
                            gap: 8,
                            minWidth: 0,
                          }}
                        >
                          <div
                            style={{
                              width: 10,
                              height: 10,
                              borderRadius: 2,
                              flexShrink: 0,
                              background: BUDGET_COLORS[i % BUDGET_COLORS.length],
                            }}
                          />
                          <span
                            style={{
                              fontSize: 12,
                              fontWeight: 500,
                              whiteSpace: "nowrap",
                              overflow: "hidden",
                              textOverflow: "ellipsis",
                            }}
                          >
                            {a.area_name}
                          </span>
                        </div>
                        <span
                          style={{
                            fontFamily: "var(--font-mono)",
                            fontSize: 11,
                            color: "var(--color-on-surface-variant)",
                            textAlign: "right",
                            minWidth: 60,
                          }}
                        >
                          {prev ? `${Math.round(prev.value_mnkr)} mnkr` : "—"}
                        </span>
                        <div
                          style={{
                            minWidth: 80,
                            display: "flex",
                            justifyContent: "flex-end",
                          }}
                        >
                          {deltaPct != null ? (
                            <DeltaIndicator
                              pct={Math.round(deltaPct * 10) / 10}
                              showBar={false}
                            />
                          ) : (
                            <span
                              style={{
                                fontFamily: "var(--font-mono)",
                                fontSize: 11,
                                color: "var(--color-on-surface-variant)",
                              }}
                            >
                              —
                            </span>
                          )}
                        </div>
                      </>
                    );

                    return makeAreaLink ? (
                      <Link
                        key={a.area_name}
                        to={makeAreaLink(a.area_name)}
                        style={rowStyle}
                      >
                        {inner}
                      </Link>
                    ) : (
                      <div key={a.area_name} style={rowStyle}>
                        {inner}
                      </div>
                    );
                  })}

                  {/* Total row */}
                  <div
                    style={{
                      display: "grid",
                      gridTemplateColumns: "1fr auto auto",
                      gap: 8,
                      alignItems: "center",
                      padding: "8px 12px",
                      borderTop: "2px solid var(--color-surface-highest)",
                      background: "var(--color-surface-low)",
                      fontWeight: 700,
                    }}
                  >
                    <span style={{ fontFamily: "var(--font-mono)", fontSize: 11 }}>
                      TOTALT
                    </span>
                    <span
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 11,
                        color: "var(--color-on-surface-variant)",
                        textAlign: "right",
                        minWidth: 60,
                      }}
                    >
                      {prevAreas.length > 0
                        ? `${Math.round(
                            prevAreas.reduce((s, a) => s + a.value_mnkr, 0)
                          )} mnkr`
                        : "—"}
                    </span>
                    <div
                      style={{
                        minWidth: 80,
                        display: "flex",
                        justifyContent: "flex-end",
                      }}
                    >
                      {(() => {
                        const curTot = areas.reduce((s, a) => s + a.value_mnkr, 0);
                        const prevTot = prevAreas.reduce(
                          (s, a) => s + a.value_mnkr,
                          0
                        );
                        return prevTot > 0 ? (
                          <DeltaIndicator
                            pct={
                              Math.round(
                                ((curTot - prevTot) / prevTot) * 1000
                              ) / 10
                            }
                            showBar={false}
                          />
                        ) : (
                          <span
                            style={{
                              fontFamily: "var(--font-mono)",
                              fontSize: 11,
                              color: "var(--color-on-surface-variant)",
                            }}
                          >
                            —
                          </span>
                        );
                      })()}
                    </div>
                  </div>
                </div>
              );
            })()}
        </>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Type check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/features/budget/components/BudgetHistorySection.tsx
git commit -m "feat(budget): shared BudgetHistorySection component"
```

---

## Task 10: RegionDetailPage — replace budget section

**Files:**
- Modify: `frontend/src/features/regions/RegionDetailPage.tsx`

- [ ] **Step 1: Read file before editing** (CLAUDE.md rule — re-read before every edit)

Read `frontend/src/features/regions/RegionDetailPage.tsx` in full.

- [ ] **Step 2: Add import for BudgetHistorySection, remove unused imports**

At the top of `RegionDetailPage.tsx`:
1. Add: `import { BudgetHistorySection } from "@/features/budget/components/BudgetHistorySection";`
2. Remove: `import { DeltaIndicator } from "@/features/budget/components/DeltaIndicator";` (no longer used directly)
3. Remove `Link` from react-router-dom imports if it's no longer used elsewhere in the file (check first).

- [ ] **Step 3: Remove `BUDGET_COLORS` constant and budget state**

Remove the `BUDGET_COLORS` array (it now lives in `BudgetHistorySection`).

Remove the budget-related state and derived values:
```typescript
const [budgetTab, setBudgetTab] = useState<number | null>(null);
// ...
const historyByYear = new Map<number, RegionBudgetSnapshot[]>();
// ...
const activeBudgetYear = budgetTab ?? sortedYears[0] ?? null;
```

- [ ] **Step 4: Change the top 2-col grid to single column**

Find the top grid wrapper (around line 340):
```typescript
gridTemplateColumns: isMobile ? "1fr" : "1fr 1fr",
```

Change to:
```typescript
gridTemplateColumns: "1fr",
```

Delete the entire right card div (`{/* Right card — BUDGET history tab strip */}` through its closing `</div>`) — approximately lines 467–634.

- [ ] **Step 5: Add BudgetHistorySection below the top grid**

After the top grid's closing `</div>` and before the bottom grid div (`{/* ── Bottom grid`), add:

```tsx
{/* ── Budget history ──────────────────────────────────────────── */}
<div
  style={{
    border: "1px solid var(--color-border)",
    borderTop: "none",
    margin: isMobile ? "1px 14px 0" : "1px 32px 0",
  }}
>
  <BudgetHistorySection
    snapshots={budgetHistory}
    makeAreaLink={(a) => `/region/${code}/budget/${encodeURIComponent(a)}`}
    emptyMessage="Budgetdata saknas för denna region."
    sourceId="scb-kostndrlt"
  />
</div>
```

- [ ] **Step 6: Type check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/features/regions/RegionDetailPage.tsx
git commit -m "refactor(regions): use shared BudgetHistorySection, full-width below grid"
```

---

## Task 11: MunicipalityDetailPage — replace budget section

**Files:**
- Modify: `frontend/src/features/municipalities/MunicipalityDetailPage.tsx`

- [ ] **Step 1: Read file before editing**

Read `frontend/src/features/municipalities/MunicipalityDetailPage.tsx` in full.

- [ ] **Step 2: Add imports, remove unused ones**

Add:
```typescript
import { BudgetHistorySection } from "@/features/budget/components/BudgetHistorySection";
import { useKommunBudgetHistory } from "@/hooks/useDemocracy";
```

Remove (no longer used in this file):
- `Donut, HBars` from chart imports
- `BUDGET_COLORS` constant

- [ ] **Step 3: Add hook call**

After `const { data, isLoading } = useKommun(code ?? "");`, add:

```typescript
const { data: budgetHistory = [] } = useKommunBudgetHistory(code ?? "");
```

- [ ] **Step 4: Remove budget derived vars**

Remove:
```typescript
const budgetAreas = data.budget.areas;
const donutSegments = budgetAreas.map(...);
const hbarsItems = budgetAreas.map(...);
```

- [ ] **Step 5: Make top grid single column, remove budget right slot**

Find the top grid wrapper containing `{/* Right — BUDGET */}`. Change `gridTemplateColumns` to `"1fr"` and delete the entire right budget card div (from `{/* Right — BUDGET */}` to its closing `</div>`).

- [ ] **Step 6: Add BudgetHistorySection below the top grid**

After the top grid's closing `</div>` and before the next section, add:

```tsx
{/* ── Budget history ──────────────────────────────────────────── */}
<div
  style={{
    border: "1px solid var(--color-border)",
    borderTop: "none",
    margin: isMobile ? "1px 14px 0" : "1px 32px 0",
  }}
>
  <BudgetHistorySection
    snapshots={budgetHistory}
    emptyMessage="Budgetdata saknas för denna kommun. Kolada — endast utvalda kommuner stöds f.n."
    sourceId="kolada-spending"
  />
</div>
```

- [ ] **Step 7: Type check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add frontend/src/features/municipalities/MunicipalityDetailPage.tsx
git commit -m "refactor(municipalities): use shared BudgetHistorySection, full-width below grid"
```

---

## Task 12: Final verification

- [ ] **Step 1: Full backend build**

```bash
go build -C backend ./...
```

Expected: success, no warnings.

- [ ] **Step 2: Full frontend type check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Start dev stack and verify visually**

```bash
make run
```

Open `http://localhost:5173/region/01` (or any region) — verify:
- Top grid is single column with mandate/hemicycle full-width
- Budget section appears below as full-width card
- Year tabs switch correctly
- Delta indicators show on tab badges and in table

Open `http://localhost:5173/municipality/0180` (Stockholm) — verify:
- Same layout as region detail
- Budget section shows (may be empty until worker runs — that's expected)

Open `http://localhost:5173/budget` — verify the national budget page is unchanged.

- [ ] **Step 4: Populate municipality budget data**

The `municipality-budget` worker runs `@weekly`. To populate data immediately in dev, restart the stack with `INITIAL_SYNC=true` (already the default in `docker-compose.dev.yml`):

```bash
make run
```

The worker will run on startup as part of the initial sync sequence. After it completes, `http://localhost:5173/municipality/0180` should show budget history rows.

If the municipality detail page still shows empty state after startup sync, check backend logs for `municipality-budget:` entries to confirm the worker ran successfully.

- [ ] **Step 5: Final commit if any fixups**

If any small issues were found and fixed:
```bash
git add -p
git commit -m "fix(budget): layout fixups after visual verification"
```
