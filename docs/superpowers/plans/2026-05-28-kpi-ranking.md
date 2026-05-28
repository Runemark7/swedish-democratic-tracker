# KPI Ranking Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a ranked bar chart page per KPI (`/kommun/:code/kpi/:kpiId`) showing all ~290 kommuner sorted by value, plus rank badges on the municipality detail page KPI cards showing position and mean diff.

**Architecture:** Two new backend endpoints hit Kolada's `municipality/all` API to get cross-municipality data. Frontend adds a `KommunKpiRankingPage` (mirroring `KommunBudgetAreaPage`) and modifies the KPI strip on `MunicipalityDetailPage` — remove Trend/GoalBadge, add clickable rank badge. Shared `kpiMeta.ts` provides label/unit for both pages.

**Tech Stack:** Go 1.26 (chi, pgx), React 19 + TypeScript 5, TanStack Query v5, inline styles (matching existing pattern)

---

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `backend/internal/regions/ports/external.go` | Add `KPIValueWithMun`, `KPIRankEntry`, `KPIRank` types; extend `KoladaClient` interface |
| Modify | `backend/internal/regions/adapters/kolada/client.go` | Implement `FetchKPIAllMunicipalities` |
| Modify | `backend/internal/regions/service.go` | Add `GetKPIRanking`, `GetMunicipalityKPIRanks` |
| Modify | `backend/internal/regions/adapters/http/handler.go` | Add two new route handlers |
| Modify | `api/openapi.yaml` | Add two new paths + two new schemas |
| Modify | `frontend/src/types/democracy.ts` | Add `kpiId` field to `Kpi` interface |
| Create | `frontend/src/features/municipalities/kpiMeta.ts` | Shared KPI strip metadata (label, unit, format) |
| Modify | `frontend/src/hooks/useDemocracy.ts` | Import from kpiMeta.ts; add `kpiId` to strip output; add two new hooks |
| Modify | `frontend/src/features/municipalities/api.ts` | Add two new API calls |
| Create | `frontend/src/features/municipalities/KommunKpiRankingPage.tsx` | Ranking page |
| Modify | `frontend/src/features/municipalities/MunicipalityDetailPage.tsx` | Remove Trend/GoalBadge, add rank badge |
| Modify | `frontend/src/App.tsx` | Register new route |

---

### Task 1: Backend ports — new types and extended interface

**Files:**
- Modify: `backend/internal/regions/ports/external.go`

- [ ] **Step 1: Add new types and extend KoladaClient interface**

Open `backend/internal/regions/ports/external.go`. After the existing `KPIValue` struct (currently around line 9), add:

```go
// KPIValueWithMun is a KPI value associated with a specific municipality,
// returned when querying Kolada's municipality/all endpoint.
type KPIValueWithMun struct {
	MunCode string
	KPI     string
	Year    int
	Value   float64
	Status  string
}

// KPIRankEntry is one row in a cross-municipality KPI ranking list.
type KPIRankEntry struct {
	MunCode string  `json:"mun_code"`
	Name    string  `json:"name"`
	Value   float64 `json:"value"`
	Year    int     `json:"year"`
	Rank    int     `json:"rank"`
	Total   int     `json:"total"`
}

// KPIRank holds the rank of a single KPI for one municipality.
type KPIRank struct {
	KPI   string  `json:"kpi"`
	Rank  int     `json:"rank"`
	Total int     `json:"total"`
	Mean  float64 `json:"mean"`
}
```

Then extend `KoladaClient` interface to add the new method:

```go
type KoladaClient interface {
	FetchKPIs(ctx context.Context, munCode string, kpiCodes []string, years []int) ([]KPIValue, error)
	FetchKPIAllMunicipalities(ctx context.Context, kpiCode string, years []int) ([]KPIValueWithMun, error)
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build -C backend ./...
```

Expected: compile error mentioning `FetchKPIAllMunicipalities` not implemented on `*Client` — this is correct, we implement it next.

---

### Task 2: Kolada client — FetchKPIAllMunicipalities

**Files:**
- Modify: `backend/internal/regions/adapters/kolada/client.go`

- [ ] **Step 1: Add FetchKPIAllMunicipalities method**

Add after the existing `fetchOneKPI` function:

```go
func (c *Client) FetchKPIAllMunicipalities(ctx context.Context, kpiCode string, years []int) ([]ports.KPIValueWithMun, error) {
	yearStrs := make([]string, len(years))
	for i, y := range years {
		yearStrs[i] = strconv.Itoa(y)
	}
	yearsStr := strings.Join(yearStrs, ",")

	url := fmt.Sprintf("%s/data/kpi/%s/municipality/all/year/%s", baseURL, kpiCode, yearsStr)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("kolada API returned %d for KPI %s/all", resp.StatusCode, kpiCode)
	}

	var payload struct {
		Values []struct {
			KPI          string `json:"kpi"`
			Municipality string `json:"municipality"`
			Period       int    `json:"period"`
			Values       []struct {
				Gender    string  `json:"gender"`
				Value     float64 `json:"value"`
				Status    string  `json:"status"`
				IsDeleted bool    `json:"isdeleted"`
			} `json:"values"`
		} `json:"values"`
	}
	err = json.NewDecoder(resp.Body).Decode(&payload)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("decode kolada all-municipalities response for %s: %w", kpiCode, err)
	}

	var result []ports.KPIValueWithMun
	for _, v := range payload.Values {
		for _, val := range v.Values {
			if val.Gender == "T" && !val.IsDeleted {
				result = append(result, ports.KPIValueWithMun{
					MunCode: v.Municipality,
					KPI:     v.KPI,
					Year:    v.Period,
					Value:   val.Value,
					Status:  val.Status,
				})
				break
			}
		}
	}
	return result, nil
}
```

- [ ] **Step 2: Verify compiles**

```bash
go build -C backend ./...
```

Expected: no errors (interface now satisfied).

- [ ] **Step 3: Commit**

```bash
git checkout -b feat/kpi-ranking
git add backend/internal/regions/ports/external.go \
        backend/internal/regions/adapters/kolada/client.go
git commit -m "feat(kpi): add KPIRankEntry/KPIRank types and FetchKPIAllMunicipalities Kolada method"
```

---

### Task 3: Service — GetKPIRanking and GetMunicipalityKPIRanks

**Files:**
- Modify: `backend/internal/regions/service.go`

- [ ] **Step 1: Add GetKPIRanking**

Add after the existing `GetMunicipalityKPIs` function (around line 146):

```go
// GetKPIRanking fetches the latest value for kpiCode across all municipalities,
// returns them sorted by value descending with 1-based rank assigned.
func (s *Service) GetKPIRanking(ctx context.Context, kpiCode string) ([]ports.KPIRankEntry, error) {
	allMuns, err := s.repo.ListMunicipalities(ctx)
	if err != nil {
		return nil, fmt.Errorf("list municipalities for ranking: %w", err)
	}
	nameByCode := make(map[string]string, len(allMuns))
	for _, m := range allMuns {
		nameByCode[m.Code] = m.Name
	}

	raw, err := s.kolada.FetchKPIAllMunicipalities(ctx, kpiCode, rollingYears(2))
	if err != nil {
		return nil, fmt.Errorf("fetch all municipalities KPI %s: %w", kpiCode, err)
	}

	// Keep only the latest year per municipality, skip missing values.
	latest := make(map[string]ports.KPIValueWithMun)
	for _, v := range raw {
		if v.Status == "M" {
			continue
		}
		if existing, ok := latest[v.MunCode]; !ok || v.Year > existing.Year {
			latest[v.MunCode] = v
		}
	}

	entries := make([]ports.KPIRankEntry, 0, len(latest))
	for munCode, v := range latest {
		entries = append(entries, ports.KPIRankEntry{
			MunCode: munCode,
			Name:    nameByCode[munCode],
			Value:   v.Value,
			Year:    v.Year,
		})
	}

	// Sort descending by value.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Value > entries[j].Value
	})

	total := len(entries)
	for i := range entries {
		entries[i].Rank = i + 1
		entries[i].Total = total
	}

	return entries, nil
}

// GetMunicipalityKPIRanks returns the rank (and mean) for each default KPI for munCode.
func (s *Service) GetMunicipalityKPIRanks(ctx context.Context, munCode string) ([]ports.KPIRank, error) {
	type result struct {
		rank ports.KPIRank
		err  error
	}

	results := make([]result, len(defaultKPIs))
	var wg sync.WaitGroup

	for i, kpi := range defaultKPIs {
		wg.Add(1)
		go func(idx int, kpiCode string) {
			defer wg.Done()
			entries, err := s.GetKPIRanking(ctx, kpiCode)
			if err != nil {
				results[idx] = result{err: err}
				return
			}
			var sum float64
			var rank ports.KPIRank
			rank.KPI = kpiCode
			rank.Total = len(entries)
			for _, e := range entries {
				sum += e.Value
				if e.MunCode == munCode {
					rank.Rank = e.Rank
				}
			}
			if len(entries) > 0 {
				rank.Mean = sum / float64(len(entries))
			}
			results[idx] = result{rank: rank}
		}(i, kpi)
	}
	wg.Wait()

	ranks := make([]ports.KPIRank, 0, len(defaultKPIs))
	for _, r := range results {
		if r.err != nil {
			continue // skip KPIs that fail, don't fail the whole response
		}
		if r.rank.Rank > 0 {
			ranks = append(ranks, r.rank)
		}
	}
	return ranks, nil
}
```

- [ ] **Step 2: Add missing imports**

Ensure `"sort"` and `"sync"` are imported at the top of `service.go`. The file already imports `"sync"` if it uses `FetchKPIs` — check with:

```bash
head -20 backend/internal/regions/service.go
```

Add any missing imports to the `import` block.

- [ ] **Step 3: Verify compiles**

```bash
go build -C backend ./...
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/regions/service.go
git commit -m "feat(kpi): add GetKPIRanking and GetMunicipalityKPIRanks service methods"
```

---

### Task 4: HTTP handler — two new routes

**Files:**
- Modify: `backend/internal/regions/adapters/http/handler.go`

- [ ] **Step 1: Register routes**

In the `Routes` method, add the two new routes **before** `r.Get("/municipalities/{code}", ...)` so chi's static-segment precedence is explicit:

```go
r.Get("/municipalities/kpi/{kpiCode}/ranking", h.getMunicipalityKPIRanking)
r.Get("/municipalities/{code}/kpi-ranks", h.getMunicipalityKPIRanks)
```

The full `Routes` block should look like:

```go
func (h *Handler) Routes(r chi.Router) {
	r.Get("/regions", h.listRegions)
	r.Get("/regions/budget/area/{areaName}", h.getAreaAcrossRegions)
	r.Get("/regions/{code}", h.getRegion)
	r.Get("/regions/{code}/budget", h.getRegionBudget)
	r.Get("/regions/{code}/budget/history", h.getRegionBudgetHistory)
	r.Get("/regions/{code}/kpi", h.getRegionKPI)
	r.Get("/municipalities", h.listMunicipalities)
	r.Get("/municipalities/budget/area/{areaName}", h.getAreaAcrossMunicipalities)
	r.Get("/municipalities/kpi/{kpiCode}/ranking", h.getMunicipalityKPIRanking)
	r.Get("/municipalities/{code}", h.getMunicipality)
	r.Get("/municipalities/{code}/kpi", h.getMunicipalityKPI)
	r.Get("/municipalities/{code}/kpi-ranks", h.getMunicipalityKPIRanks)
	r.Get("/municipalities/{code}/spending", h.getMunicipalitySpending)
	r.Get("/municipalities/{code}/budget/history", h.getMunicipalityBudgetHistory)
	r.Get("/municipalities/{code}/population-trend", h.getPopulationTrend)
	r.Get("/municipalities/{code}/procurement", h.getMunicipalityProcurement)
}
```

- [ ] **Step 2: Add handler methods**

Add these two handler functions at the end of the file:

```go
func (h *Handler) getMunicipalityKPIRanking(w http.ResponseWriter, r *http.Request) {
	kpiCode := chi.URLParam(r, "kpiCode")
	entries, err := h.svc.GetKPIRanking(r.Context(), kpiCode)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	if entries == nil {
		entries = []ports.KPIRankEntry{}
	}
	jsonOK(w, entries)
}

func (h *Handler) getMunicipalityKPIRanks(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	ranks, err := h.svc.GetMunicipalityKPIRanks(r.Context(), code)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	if ranks == nil {
		ranks = []ports.KPIRank{}
	}
	jsonOK(w, ranks)
}
```

- [ ] **Step 3: Verify compiles**

```bash
go build -C backend ./...
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/regions/adapters/http/handler.go
git commit -m "feat(kpi): add /municipalities/kpi/:kpiCode/ranking and /:code/kpi-ranks HTTP routes"
```

---

### Task 5: OpenAPI spec + regenerate frontend types

**Files:**
- Modify: `api/openapi.yaml`

- [ ] **Step 1: Add new schemas**

Find the `MunicipalityKPIItem:` schema block (around line 1845) and add two new schemas after it:

```yaml
    KPIRankEntry:
      type: object
      required: [mun_code, name, value, year, rank, total]
      properties:
        mun_code:
          type: string
          example: "0180"
        name:
          type: string
          example: "Stockholm"
        value:
          type: number
          format: float
          example: 22.85
        year:
          type: integer
          example: 2023
        rank:
          type: integer
          example: 47
        total:
          type: integer
          example: 290

    KPIRank:
      type: object
      required: [kpi, rank, total, mean]
      properties:
        kpi:
          type: string
          example: "N00900"
        rank:
          type: integer
          example: 47
        total:
          type: integer
          example: 290
        mean:
          type: number
          format: float
          example: 22.14
```

- [ ] **Step 2: Add new paths**

Find the `/municipalities/{code}/kpi:` path block (around line 693) and add two new paths **before** it (so static paths appear before parameterised ones in the spec):

```yaml
  /municipalities/kpi/{kpiCode}/ranking:
    get:
      operationId: getMunicipalityKPIRanking
      tags: [municipalities]
      summary: All municipalities ranked by a single KPI (highest first)
      parameters:
        - name: kpiCode
          in: path
          required: true
          schema:
            type: string
            example: "N00900"
      responses:
        "200":
          description: Array of rank entries sorted by value descending
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/KPIRankEntry"
        "502":
          description: Upstream Kolada API error

  /municipalities/{code}/kpi-ranks:
    get:
      operationId: getMunicipalityKPIRanks
      tags: [municipalities]
      summary: Rank of a municipality for each default KPI
      parameters:
        - name: code
          in: path
          required: true
          schema:
            type: string
            example: "0180"
      responses:
        "200":
          description: Array of KPI rank entries (one per default KPI)
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/KPIRank"
        "502":
          description: Upstream Kolada API error
```

- [ ] **Step 3: Regenerate frontend API types**

```bash
cd frontend && npm run generate:api
```

Expected: `src/shared/api-contract.ts` updated with `KPIRankEntry` and `KPIRank` types.

- [ ] **Step 4: Commit**

```bash
git add api/openapi.yaml frontend/src/shared/api-contract.ts
git commit -m "feat(kpi): add KPIRankEntry and KPIRank to OpenAPI spec"
```

---

### Task 6: Frontend — kpiMeta.ts shared constant

**Files:**
- Create: `frontend/src/features/municipalities/kpiMeta.ts`
- Modify: `frontend/src/hooks/useDemocracy.ts`

- [ ] **Step 1: Create kpiMeta.ts**

```ts
export interface KpiMeta {
  label: string;
  description: string;
  unit: string;
  target: number;
  worseHigher: boolean;
  format: (v: number) => string;
}

export const STRIP_KPI_META: Record<string, KpiMeta> = {
  N00900: {
    label: "Kommunalskatt",
    description: "Din inkomstskatt till kommunen. Lägre skatt ger mer kvar i plånboken — men kan också innebära sämre service.",
    unit: "%", target: 31.0, worseHigher: true, format: v => `${v.toFixed(2)} %`,
  },
  N03102: {
    label: "Resultat/skatt",
    description: "Kommunens överskott i förhållande till skatteintäkterna. Under 2 % riskerar kommunen att tvingas skära i välfärden.",
    unit: "%", target: 2.0, worseHigher: false, format: v => `${v.toFixed(1)} %`,
  },
  N03106: {
    label: "Soliditet",
    description: "Hur stor del av kommunens tillgångar som är skuldfria. Låg soliditet ökar sårbarheten vid ekonomiska kriser.",
    unit: "%", target: 25.0, worseHigher: false, format: v => `${v.toFixed(0)} %`,
  },
  N15428: {
    label: "Gymnasiebehörighet",
    description: "Andel elever i åk 9 som är behöriga till gymnasiet. Viktig signal om skolkvaliteten i kommunen.",
    unit: "%", target: 85.0, worseHigher: false, format: v => `${v.toFixed(1)} %`,
  },
  N00708: {
    label: "Arbetslöshet",
    description: "Andel av befolkningen 20–64 år som var arbetslösa någon gång under året. Låg arbetslöshet stärker kommunens skatteunderlag.",
    unit: "%", target: 5.0, worseHigher: true, format: v => `${v.toFixed(1)} %`,
  },
};

export const STRIP_ORDER = ["N00900", "N03102", "N03106", "N15428", "N00708"];
```

- [ ] **Step 2: Remove duplicate from useDemocracy.ts**

In `frontend/src/hooks/useDemocracy.ts`, replace the existing inline `STRIP_KPI_META` and `STRIP_ORDER` definitions (around lines 209–236) with an import:

```ts
import { STRIP_KPI_META, STRIP_ORDER, type KpiMeta } from "@/features/municipalities/kpiMeta";
```

Remove the now-redundant `type KpiMeta` inline definition (around line 200) and the `const STRIP_KPI_META = {...}` and `const STRIP_ORDER = [...]` blocks.

- [ ] **Step 3: Type check**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -20
```

Expected: zero errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/municipalities/kpiMeta.ts \
        frontend/src/hooks/useDemocracy.ts
git commit -m "refactor(kpi): extract STRIP_KPI_META to shared kpiMeta.ts"
```

---

### Task 7: Frontend — add kpiId to Kpi type and strip output

**Files:**
- Modify: `frontend/src/types/democracy.ts`
- Modify: `frontend/src/hooks/useDemocracy.ts`

- [ ] **Step 1: Add kpiId to Kpi interface**

In `frontend/src/types/democracy.ts`, update the `Kpi` interface to add `kpiId`:

```ts
export interface Kpi {
  kpiId: string;      // Kolada KPI code, e.g. "N00900"
  label: string;
  description: string;
  value: string;
  raw: number;
  target: number;
  worseHigher: boolean;
  unit: string;
  trend: "up" | "down" | "flat";
  delta: string;
  note: string;
  sourceUrl?: string;
}
```

- [ ] **Step 2: Set kpiId in kpiItemsToStrip**

In `frontend/src/hooks/useDemocracy.ts`, in the `kpiItemsToStrip` function (around line 263), add `kpiId: code` to the returned object:

```ts
    return [{
      kpiId: code,
      label: m.label,
      description: m.description,
      value: m.format(latest.value),
      raw: latest.value,
      target: m.target,
      worseHigher: m.worseHigher,
      unit: m.unit,
      trend,
      delta: prev ? `${sign}${absDelta} ${m.unit}` : "–",
      note: `Källa: Kolada ${latest.year}`,
    } satisfies Kpi];
```

- [ ] **Step 3: Type check**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -20
```

Expected: zero errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/types/democracy.ts \
        frontend/src/hooks/useDemocracy.ts
git commit -m "feat(kpi): add kpiId field to Kpi interface and strip output"
```

---

### Task 8: Frontend — new API calls and hooks

**Files:**
- Modify: `frontend/src/features/municipalities/api.ts`
- Modify: `frontend/src/hooks/useDemocracy.ts`

- [ ] **Step 1: Add API methods to municipalitiesApi**

In `frontend/src/features/municipalities/api.ts`, add to the `municipalitiesApi` object:

```ts
  getKPIRanking: (kpiCode: string) =>
    api.get<{ mun_code: string; name: string; value: number; year: number; rank: number; total: number }[]>(
      `/municipalities/kpi/${kpiCode}/ranking`
    ),
  getMunicipalityKPIRanks: (code: string) =>
    api.get<{ kpi: string; rank: number; total: number; mean: number }[]>(
      `/municipalities/${code}/kpi-ranks`
    ),
```

- [ ] **Step 2: Add hooks to useDemocracy.ts**

At the end of `frontend/src/hooks/useDemocracy.ts`, add:

```ts
// ── KPI ranking — all municipalities for one KPI ──────────────────────────────
export interface KPIRankEntry {
  mun_code: string;
  name: string;
  value: number;
  year: number;
  rank: number;
  total: number;
}

export function useKpiRanking(kpiCode: string) {
  return useQuery<KPIRankEntry[]>({
    queryKey: ["kpi-ranking", kpiCode],
    queryFn: () => municipalitiesApi.getKPIRanking(kpiCode),
    staleTime: 10 * 60 * 1000,
    enabled: !!kpiCode,
  });
}

// ── KPI ranks for one municipality ────────────────────────────────────────────
export interface KPIRank {
  kpi: string;
  rank: number;
  total: number;
  mean: number;
}

export function useKommunKpiRanks(munCode: string) {
  return useQuery<KPIRank[]>({
    queryKey: ["kommun-kpi-ranks", munCode],
    queryFn: () => municipalitiesApi.getMunicipalityKPIRanks(munCode),
    staleTime: 10 * 60 * 1000,
    enabled: !!munCode,
  });
}
```

- [ ] **Step 3: Type check**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -20
```

Expected: zero errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/municipalities/api.ts \
        frontend/src/hooks/useDemocracy.ts
git commit -m "feat(kpi): add getKPIRanking and getMunicipalityKPIRanks API calls + hooks"
```

---

### Task 9: Frontend — KommunKpiRankingPage

**Files:**
- Create: `frontend/src/features/municipalities/KommunKpiRankingPage.tsx`

- [ ] **Step 1: Create the page**

```tsx
import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useKpiRanking, useKommunList } from "@/hooks/useDemocracy";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { BarsWithMean } from "@/features/budget/components/BarsWithMean";
import { STRIP_KPI_META } from "./kpiMeta";

const INITIAL_LIMIT = 20;

function Skeleton() {
  return (
    <div className="sdt-page" style={{ padding: "40px 32px" }}>
      {[80, 200, 320].map((h, i) => (
        <div key={i} style={{ height: h, background: "var(--color-track)", borderRadius: 4, marginBottom: 16 }} />
      ))}
    </div>
  );
}

export function KommunKpiRankingPage() {
  const { code = "", kpiId = "" } = useParams<{ code: string; kpiId: string }>();
  const isMobile = useMediaQuery("(max-width: 640px)");
  const [showAll, setShowAll] = useState(false);

  const { data: kommunList } = useKommunList();
  const { data: rankEntries = [], isLoading } = useKpiRanking(kpiId);

  const kommunName = kommunList?.find((k) => k.code === code)?.name ?? code;
  const meta = STRIP_KPI_META[kpiId];
  const kpiLabel = meta?.label ?? kpiId;
  const kpiUnit = meta?.unit ?? "";

  // Sort descending (already sorted by backend, but ensure)
  const sorted = [...rankEntries].sort((a, b) => b.value - a.value);

  const mean = sorted.length > 0
    ? sorted.reduce((s, e) => s + e.value, 0) / sorted.length
    : 0;
  const maxValue = sorted[0]?.value ?? 1;

  const visible = showAll ? sorted : sorted.slice(0, INITIAL_LIMIT);

  if (isLoading) return <Skeleton />;

  const pad = isMobile ? "14px" : "32px";

  return (
    <div className="sdt-page">
      {/* Breadcrumb */}
      <div style={{ padding: isMobile ? "14px 14px 0" : "28px 32px 0", display: "flex", alignItems: "center", gap: 8 }}>
        <Link
          to={`/kommun/${code}`}
          style={{ fontFamily: "var(--font-mono)", fontSize: 11, letterSpacing: "0.1em", color: "var(--color-accent)", textDecoration: "none", textTransform: "uppercase" }}
        >
          ← {kommunName}
        </Link>
        <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-fg-muted)", letterSpacing: "0.08em" }}>
          / {kpiLabel}
        </span>
      </div>

      {/* Comparison section */}
      <div style={{
        border: "1px solid var(--color-border)",
        margin: isMobile ? "14px 14px 28px" : "20px 32px 28px",
        background: "var(--color-sdt-surface)",
        padding: isMobile ? "14px" : "20px 24px",
      }}>
        <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, letterSpacing: "0.15em", color: "var(--color-fg-muted)", marginBottom: 4, textTransform: "uppercase" }}>
          Jämförelse med alla kommuner
        </div>
        <div style={{ fontFamily: "var(--font-mono)", fontSize: 9, letterSpacing: "0.08em", color: "var(--color-fg-muted)", marginBottom: 20, opacity: 0.7 }}>
          Råvärde ({kpiUnit}), senaste år · {kommunName} markerad
        </div>

        {sorted.length === 0 ? (
          <div style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-fg-muted)", padding: "20px 0", textAlign: "center" }}>
            Data saknas
          </div>
        ) : (
          <>
            <BarsWithMean
              points={visible.map((e) => ({ region_code: e.mun_code, pct: e.value }))}
              mean={mean}
              maxValue={maxValue}
              highlightCode={code}
              codeToName={new Map(rankEntries.map((e) => [e.mun_code, e.name]))}
              isMobile={isMobile}
            />
            {!showAll && sorted.length > INITIAL_LIMIT && (
              <button
                onClick={() => setShowAll(true)}
                style={{
                  marginTop: 16,
                  padding: "7px 14px",
                  fontFamily: "var(--font-mono)",
                  fontSize: 10,
                  letterSpacing: "0.1em",
                  color: "var(--color-accent)",
                  background: "transparent",
                  border: "1px solid var(--color-border)",
                  cursor: "pointer",
                  display: "block",
                }}
              >
                VISA FLER KOMMUNER ({sorted.length - INITIAL_LIMIT} ST)
              </button>
            )}
          </>
        )}
      </div>

      {/* Legend */}
      <div style={{ padding: `0 ${pad} 16px`, display: "flex", gap: 16, flexWrap: "wrap" }}>
        <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
          <span style={{ display: "inline-block", width: 12, height: 4, borderRadius: 2, background: "var(--color-accent, #0b3d7a)" }} />
          <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", letterSpacing: "0.08em" }}>
            {kommunName.toUpperCase()}
          </span>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
          <span style={{ display: "inline-block", width: 12, height: 4, borderRadius: 2, background: "var(--color-fg-muted, #7a8390)" }} />
          <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", letterSpacing: "0.08em" }}>
            ÖVRIGA KOMMUNER
          </span>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
          <span style={{ display: "inline-block", width: 1.5, height: 12, borderLeft: "1.5px dashed #d97706" }} />
          <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", letterSpacing: "0.08em" }}>
            MEDELVÄRDE
          </span>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Type check**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -20
```

Expected: zero errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/features/municipalities/KommunKpiRankingPage.tsx
git commit -m "feat(kpi): add KommunKpiRankingPage"
```

---

### Task 10: Frontend — update MunicipalityDetailPage KPI cards

**Files:**
- Modify: `frontend/src/features/municipalities/MunicipalityDetailPage.tsx`

- [ ] **Step 1: Add useKommunKpiRanks import and hook call**

At the top of the file, add `useKommunKpiRanks` to the import from `@/hooks/useDemocracy`:

```ts
import { useKommun, useKommunList, useKommunBudgetHistory, useKommunKpiRanks } from "@/hooks/useDemocracy";
```

Inside `MunicipalityDetailPage`, after the existing hook calls, add:

```ts
const { data: kpiRanks = [] } = useKommunKpiRanks(code ?? "");
const rankMap = new Map(kpiRanks.map((r) => [r.kpi, r]));
```

- [ ] **Step 2: Replace KPI card body — remove Trend + GoalBadge, add rank badge**

Find the KPI card render block (around lines 269–333). Replace the inner content of each card from after the value `<div>` block through the description. The current code:

```tsx
                <div style={{ marginBottom: 12 }}>
                  <Trend trend={kpi.trend} delta={kpi.delta} worseHigher={kpi.worseHigher} />
                </div>
                <GoalBadge
                  raw={kpi.raw}
                  target={kpi.target}
                  worseHigher={kpi.worseHigher}
                  unit={kpi.unit}
                  note={kpi.note}
                  sourceUrl={kpi.sourceUrl}
                />
```

Replace with rank badge:

```tsx
                {(() => {
                  const rank = rankMap.get(kpi.kpiId);
                  if (!rank) return null;
                  const diff = kpi.raw - rank.mean;
                  return (
                    <a
                      href={`/kommun/${code}/kpi/${kpi.kpiId}`}
                      style={{ textDecoration: "none", display: "inline-flex", alignItems: "baseline", gap: 7, marginBottom: 10, flexWrap: "wrap" }}
                    >
                      <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "#b91c1c", fontWeight: 700 }}>
                        #{rank.rank} AV {rank.total}
                      </span>
                      <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)" }}>·</span>
                      <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)" }}>
                        <span style={{ fontWeight: 700, color: "var(--color-fg)" }}>
                          {diff >= 0 ? "+" : ""}{diff.toFixed(2)} {kpi.unit}
                        </span>
                        {" "}{diff >= 0 ? "ÖVER" : "UNDER"} MEDEL →
                      </span>
                    </a>
                  );
                })()}
```

- [ ] **Step 3: Remove unused Trend and GoalBadge imports if no longer used**

```bash
grep -n "Trend\|GoalBadge" frontend/src/features/municipalities/MunicipalityDetailPage.tsx
```

If `Trend` and `GoalBadge` only appear in imports (no other usages), remove them from the import line:

```ts
// Before:
import { MandateComposition, Pill, Trend, GoalBadge } from "@/components/charts";
// After (if both unused):
import { MandateComposition, Pill } from "@/components/charts";
```

- [ ] **Step 4: Type check**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -20
```

Expected: zero errors.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/features/municipalities/MunicipalityDetailPage.tsx
git commit -m "feat(kpi): replace Trend/GoalBadge with rank badge + mean diff on KPI cards"
```

---

### Task 11: Frontend — register route in App.tsx

**Files:**
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: Add import and route**

Add import near the other municipality page imports:

```ts
import { KommunKpiRankingPage } from "./features/municipalities/KommunKpiRankingPage";
```

Find the existing route:
```tsx
<Route path="/kommun/:code/budget/:areaName" element={<KommunBudgetAreaPage />} />
```

Add the new route directly after it:
```tsx
<Route path="/kommun/:code/kpi/:kpiId" element={<KommunKpiRankingPage />} />
```

- [ ] **Step 2: Type check**

```bash
cd frontend && npx tsc --noEmit 2>&1
```

Expected: zero errors.

- [ ] **Step 3: Build check**

```bash
go build -C backend ./...
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/App.tsx
git commit -m "feat(kpi): register /kommun/:code/kpi/:kpiId route"
```

---

### Task 12: Push and open PR

- [ ] **Step 1: Push branch**

```bash
git push -u origin feat/kpi-ranking
```

- [ ] **Step 2: Open PR**

```bash
gh pr create \
  --title "feat(kpi): KPI ranking page + rank badges on municipality detail" \
  --body "$(cat <<'EOF'
## Summary
- New page \`/kommun/:code/kpi/:kpiId\` shows all ~290 kommuner ranked by a single KPI (highest first), mean line, show-more at 20, current highlighted red
- KPI cards on municipality detail page now show rank badge (#N AV 290) and mean diff (+X % ÖVER/UNDER MEDEL) instead of Trend arrows and GoalBadge pills
- Backend: two new endpoints using Kolada's \`municipality/all\` API — one for full ranking list, one for per-KPI ranks of a single municipality
- Shared \`kpiMeta.ts\` extracts KPI label/unit metadata

## Test plan
- [ ] Navigate to any kommun → KPI cards show rank + mean diff
- [ ] Click rank badge → /kommun/:code/kpi/N00900 ranking page opens
- [ ] Ranking page shows bars sorted high→low, current kommun highlighted red
- [ ] Show-more button reveals all communes
- [ ] Mean dashed line correct
- [ ] Data saknas shown for empty results

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```
