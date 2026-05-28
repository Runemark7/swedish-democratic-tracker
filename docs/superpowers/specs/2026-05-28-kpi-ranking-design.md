# KPI Ranking — Design Spec

**Date:** 2026-05-28  
**Status:** Approved

---

## Goal

Add ranked comparison of all ~290 kommuner for each KPI shown on the municipality detail page. Two surfaces:

1. **Rank badge on KPI card** — `#47 AV 290` shown inside each KPI card on `MunicipalityDetailPage`, linking to the ranking page.
2. **Ranking page** `/kommun/:code/kpi/:kpiId` — full sorted bar list of all kommuner for one KPI, current highlighted, mean line, show-more at 20.

---

## Backend

### New Kolada method

`FetchKPIAllMunicipalities(ctx context.Context, kpiCode string, years []int) ([]KPIValueWithMun, error)`

- URL: `GET https://api.kolada.se/v3/data/kpi/{kpiCode}/municipality/all/year/{years}`
- Kolada returns the same payload shape as `fetchOneKPI` but with a `municipality` field per row.
- New struct (internal to kolada adapter):
  ```go
  type KPIValueWithMun struct {
      MunCode string
      KPI     string
      Year    int
      Value   float64
      Status  string
  }
  ```
- Filter: `gender == "T"`, `isdeleted == false`, take latest year per municipality.

### New ports types

Add to `backend/internal/regions/ports/external.go`:

```go
// KPIRankEntry is one row in a cross-municipality KPI ranking.
type KPIRankEntry struct {
    MunCode string  `json:"mun_code"`
    Name    string  `json:"name"`
    Value   float64 `json:"value"`
    Year    int     `json:"year"`
    Rank    int     `json:"rank"`
    Total   int     `json:"total"`
}

// KPIRank is the rank of a single KPI for one municipality.
type KPIRank struct {
    KPI   string `json:"kpi"`
    Rank  int    `json:"rank"`
    Total int    `json:"total"`
}
```

Add to `KoladaClient` interface:
```go
FetchKPIAllMunicipalities(ctx context.Context, kpiCode string, years []int) ([]KPIValueWithMun, error)
```

### New service methods

`GetKPIRanking(ctx, kpiCode string) ([]KPIRankEntry, error)`
1. Call `s.kolada.FetchKPIAllMunicipalities(ctx, kpiCode, rollingYears(2))`
2. For each municipality code, resolve name via `s.repo.ListMunicipalities(ctx)` (cached — already used by `listMunicipalities` handler)
3. Keep only the latest year per municipality
4. Sort by value descending
5. Assign `Rank` (1-based) and `Total = len(entries)`
6. Return

`GetMunicipalityKPIRanks(ctx, munCode string) ([]KPIRank, error)`
1. Call `GetKPIRanking` concurrently for each KPI in `defaultKPIs` (goroutines + WaitGroup, same pattern as existing `FetchKPIs`)
2. For each result, find the entry matching `munCode`, extract `Rank` and `Total`
3. Return `[]KPIRank` — one per KPI

**Caching:** Both methods are Kolada-live. Response time for `municipality/all` should be ~500ms per KPI. `GetMunicipalityKPIRanks` fires 12 concurrent calls — acceptable for a detail page load. No in-memory cache needed in v1; TanStack Query on frontend provides 10-min staleTime.

### New HTTP routes

Add to `handler.go` `Routes()`:

```
GET /municipalities/kpi/{kpiCode}/ranking   → getMunicipalityKPIRanking
GET /municipalities/{code}/kpi-ranks        → getMunicipalityKPIRanks
```

**`getMunicipalityKPIRanking`** — calls `svc.GetKPIRanking`, returns `[]KPIRankEntry` JSON array.

**`getMunicipalityKPIRanks`** — calls `svc.GetMunicipalityKPIRanks`, returns `[]KPIRank` JSON array.

Note: `/municipalities/kpi/{kpiCode}/ranking` must be registered **before** `/municipalities/{code}` so chi routes it correctly (static segment before param).

---

## Frontend

### OpenAPI spec (`api/openapi.yaml`)

Add two new paths before regenerating `api-contract.ts`.

### New hooks (`frontend/src/hooks/useDemocracy.ts`)

```ts
useKpiRanking(kpiCode: string)        // GET /municipalities/kpi/:kpiCode/ranking
useKommunKpiRanks(munCode: string)    // GET /municipalities/:code/kpi-ranks
```

Both use TanStack Query with `staleTime: 10 * 60 * 1000`.

### New page: `KommunKpiRankingPage`

Path: `frontend/src/features/municipalities/KommunKpiRankingPage.tsx`  
Route: `/kommun/:code/kpi/:kpiId`

Layout (mirrors `KommunBudgetAreaPage` without the trend section):

```
← Ale / Kommunalskatt              ← breadcrumb

┌──────────────────────────────────┐
│ JÄMFÖRELSE MED ALLA KOMMUNER     │
│ Råvärde, senaste år · Ale markerad│
│                                   │
│  [BarsWithMean — all kommuner]    │
│  [VISA FLER KOMMUNER (N ST)]      │
└──────────────────────────────────┘

● ALE   ▬ ÖVRIGA KOMMUNER   ╌ MEDELVÄRDE   ← legend
```

- Reuses `BarsWithMean` directly — `points` shape `{region_code, pct}` → map `mun_code → region_code`, `value → pct` (BarsWithMean is value-agnostic; label it via subtitle)
- `maxValue` = highest value in list; bars show raw value not pct — subtitle reflects unit (e.g. `%` or `kr/inv`)
- Initial 20, show-more button for remainder
- Mean = arithmetic mean of all values

**KPI label map** (same 12 KPIs as `MunicipalityComparePage`'s `ALL_KPI_LABELS`): extract to a shared `frontend/src/features/municipalities/kpiMeta.ts` constant so both pages and the new ranking page share it without duplication.

### Modified: `MunicipalityDetailPage`

**Remove from each KPI card:**
- Delta/trend row (year-over-year change arrow + %)
- Target pill ("X pp över/under mål")

**Add to each KPI card** (directly below the value, above the description):

```tsx
<a href={`/kommun/${code}/kpi/${kpi.kpiId}`} style={{ textDecoration: "none", display: "inline-flex", alignItems: "baseline", gap: 7 }}>
  <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-rank, #b91c1c)", fontWeight: 700 }}>
    #{rank.rank} AV {rank.total}
  </span>
  <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)" }}>·</span>
  <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)" }}>
    <span style={{ fontWeight: 700, color: "var(--color-fg)" }}>
      {diff >= 0 ? "+" : ""}{diff.toFixed(2)} {unit}
    </span>
    {" "}{diff >= 0 ? "ÖVER" : "UNDER"} MEDEL →
  </span>
</a>
```

Where `diff = kpi.value - rank.mean` and `unit` comes from the KPI metadata (e.g. `%`, `kr/inv`).

The badge is clickable — navigates to `/kommun/:code/kpi/:kpiId`.

`rank.mean` must be added to the `KPIRank` port type:
```go
type KPIRank struct {
    KPI   string  `json:"kpi"`
    Rank  int     `json:"rank"`
    Total int     `json:"total"`
    Mean  float64 `json:"mean"`
}
```

### New route in `App.tsx`

```tsx
<Route path="/kommun/:code/kpi/:kpiId" element={<KommunKpiRankingPage />} />
```

Register alongside the existing `/kommun/:code/budget/:areaName` route.

---

## Files touched

| Action | Path |
|--------|------|
| Modify | `backend/internal/regions/ports/external.go` |
| Modify | `backend/internal/regions/adapters/kolada/client.go` |
| Modify | `backend/internal/regions/service.go` |
| Modify | `backend/internal/regions/adapters/http/handler.go` |
| Modify | `api/openapi.yaml` |
| Create | `frontend/src/features/municipalities/kpiMeta.ts` |
| Create | `frontend/src/features/municipalities/KommunKpiRankingPage.tsx` |
| Modify | `frontend/src/features/municipalities/MunicipalityDetailPage.tsx` |
| Modify | `frontend/src/hooks/useDemocracy.ts` |
| Modify | `frontend/src/App.tsx` |

---

## Out of scope

- Region KPI ranking (same pattern, addable later)
- Persistent DB cache for KPI rankings
- "Best performer" direction-adjusted ranking
