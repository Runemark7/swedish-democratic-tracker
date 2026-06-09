# Value Verification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove every value the site renders (regions, kommuner, riksdagen) matches a freshly-fetched upstream value, via an on-demand Go command that writes a dated review report — and make the source-verification badges honest.

**Architecture:** Two independent streams. **Stream A** closes documentation/markup gaps so `verification_status` badges reflect reality (frontend `<SourceMarker>` additions + frontmatter flips). **Stream B** adds `backend/cmd/verifyvalues/`, a cross-feature Go command that reads each value from the same service the HTTP API uses ("ours") and from the existing upstream adapter clients ("upstream"), diffs them, and writes `docs/superpowers/audits/2026-06-07-value-verification.md`. No CI gate; report-only.

**Tech Stack:** Go 1.26.1 (chi, pgx/v5, existing feature services + adapter clients), Node 25.9.0 (frontend sync/verify scripts), React 19 + TypeScript.

**Spec:** `docs/superpowers/specs/2026-06-07-value-verification-design.md`

---

## File Structure

**Stream A (frontend + docs):**
- Modify: `frontend/src/features/regions/RegionBudgetAreaPage.tsx` — add `<SourceMarker sourceId="scb-kostndrlt">` on mnkr chart labels (~line 140)
- Modify: `frontend/src/features/municipalities/MunicipalityDetailPage.tsx` — add `<SourceMarker sourceId="scb-befolkning">` on population sub-header (~line 194)
- Modify: `frontend/src/features/riksdag/AuthorityDetailPage.tsx` — add `<SourceMarker sourceId="statskontoret-arsutfall">` on cost/staff values
- Modify: `frontend/src/features/riksdag/MyndigheterListPage.tsx` — same source marker on cost/staff values
- Modify: `frontend/src/features/regions/components/SwedenRegionMap.tsx` — add `<SourceMarker sourceId="wikimedia-svg">` attribution
- Modify: `frontend/src/features/municipalities/components/SwedenKommunMap.tsx` — same attribution
- Modify: `docs/data-sources/{kolada,kolada-spending,riksdagen,scb-kfmandat,scb-ltmandat,scb-befolkning,scb-kostndrlt,wikimedia-svg,statskontoret-arsutfall}.md` — `verification_status` frontmatter

**Stream B (`backend/cmd/verifyvalues/`, package `main`):**
- Create: `backend/cmd/verifyvalues/check.go` — `Status`, `Result`, `classify` (pure)
- Create: `backend/cmd/verifyvalues/check_test.go`
- Create: `backend/cmd/verifyvalues/report.go` — `renderReport` (pure)
- Create: `backend/cmd/verifyvalues/report_test.go`
- Create: `backend/cmd/verifyvalues/regions.go` — region + kommun checks
- Create: `backend/cmd/verifyvalues/riksdagen.go` — seat/roster, vote, speech checks
- Create: `backend/cmd/verifyvalues/main.go` — wiring + report write
- Modify: `backend/internal/regions/seeder/seeder.go` — add exported `FetchMandateTotals`
- Create: `backend/internal/regions/seeder/mandates_export_test.go`
- Modify: `Makefile` — add `verify-values` target

---

## Stream A — Honest badges

### Task 1: Add SourceMarkers — region budget + kommun population

**Files:**
- Modify: `frontend/src/features/regions/RegionBudgetAreaPage.tsx`
- Modify: `frontend/src/features/municipalities/MunicipalityDetailPage.tsx`

- [ ] **Step 1: Locate the untagged region budget value**

Run: `grep -n 'toLocaleString\|SourceMarker' frontend/src/features/regions/RegionBudgetAreaPage.tsx`
Expected: a `Math.round(p.value).toLocaleString("sv-SE")` near line 140 with no adjacent `<SourceMarker>`.

- [ ] **Step 2: Ensure the import exists**

At the top of `RegionBudgetAreaPage.tsx`, confirm or add:

```tsx
import { SourceMarker } from "@/components/sources/SourceMarker";
```

- [ ] **Step 3: Tag the mnkr value**

Find the JSX rendering the chart-dot / area mnkr value (the `toLocaleString("sv-SE")` expression) and append the marker inline, matching the existing pattern used in `RegionLandingPage.tsx:106`:

```tsx
{Math.round(p.value).toLocaleString("sv-SE")}
<SourceMarker sourceId="scb-kostndrlt" />
```

Place it on the most prominent mnkr value of the page (the header/total figure), not inside the SVG `<text>` nodes where a React component cannot render. If the only values are SVG dot labels, add a single `<SourceMarker sourceId="scb-kostndrlt" />` next to the chart title/legend instead.

- [ ] **Step 4: Locate the untagged kommun population**

Run: `grep -n 'population\|SourceMarker\|import' frontend/src/features/municipalities/MunicipalityDetailPage.tsx`
Expected: a `data.population` / `.population` render around line 194 with no adjacent marker.

- [ ] **Step 5: Tag the kommun population**

Confirm/add the import, then append after the population value:

```tsx
{data.population.toLocaleString("sv-SE")}
<SourceMarker sourceId="scb-befolkning" />
```

- [ ] **Step 6: Typecheck**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS (no errors).

- [ ] **Step 7: Commit**

```bash
git add frontend/src/features/regions/RegionBudgetAreaPage.tsx frontend/src/features/municipalities/MunicipalityDetailPage.tsx
git commit -m "feat(sources): tag region budget mnkr and kommun population values"
```

---

### Task 2: Add SourceMarkers — authority pages + map attribution

**Files:**
- Modify: `frontend/src/features/riksdag/AuthorityDetailPage.tsx`
- Modify: `frontend/src/features/riksdag/MyndigheterListPage.tsx`
- Modify: `frontend/src/features/regions/components/SwedenRegionMap.tsx`
- Modify: `frontend/src/features/municipalities/components/SwedenKommunMap.tsx`

- [ ] **Step 1: Locate authority cost/staff values**

Run: `grep -n 'kostnad\|kr\|årsarbets\|personal\|toLocaleString\|SourceMarker\|import' frontend/src/features/riksdag/AuthorityDetailPage.tsx frontend/src/features/riksdag/MyndigheterListPage.tsx`
Expected: rendered expenditure (kr) and headcount values with no adjacent marker.

- [ ] **Step 2: Tag authority values**

In each file confirm/add `import { SourceMarker } from "@/components/sources/SourceMarker";`, then append inline after the expenditure and headcount values:

```tsx
<SourceMarker sourceId="statskontoret-arsutfall" />
```

For `MyndigheterListPage.tsx`, if values are in a repeated table row, add ONE marker in the column header (e.g. the "Kostnad" / "Årsarbetskrafter" `<th>`), not per row.

- [ ] **Step 3: Add map attribution markers**

In `SwedenRegionMap.tsx` and `SwedenKommunMap.tsx`, add the import and render a single attribution marker near the map caption/title (the CC BY-SA 2.5 license requires visible attribution):

```tsx
<SourceMarker sourceId="wikimedia-svg" />
```

Place it where it is visible in the layout (a caption row beneath the map), not on an SVG `<path>`.

- [ ] **Step 4: Typecheck**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/features/riksdag/AuthorityDetailPage.tsx frontend/src/features/riksdag/MyndigheterListPage.tsx frontend/src/features/regions/components/SwedenRegionMap.tsx frontend/src/features/municipalities/components/SwedenKommunMap.tsx
git commit -m "feat(sources): tag authority expenditure/headcount and add map attribution"
```

---

### Task 3: Promote verification_status flags + statskontoret drift check

**Files:**
- Modify: `docs/data-sources/kolada.md`, `kolada-spending.md`, `riksdagen.md`, `scb-kfmandat.md`, `scb-ltmandat.md`, `scb-befolkning.md`, `scb-kostndrlt.md`, `wikimedia-svg.md`, `statskontoret-arsutfall.md`

- [ ] **Step 1: Promote the five already-fixed sources**

In each of `docs/data-sources/{kolada,kolada-spending,riksdagen,scb-kfmandat,scb-ltmandat}.md`, change the frontmatter:

```yaml
verification_status: verified
verification_notes: ""
last_verified: 2026-06-07
```

(Their notes already state "Repro-narrativet är nu godkänt"; the only remaining reason for `unsure` was the now-removed curl block.)

- [ ] **Step 2: Promote the two now-tagged sources**

In `docs/data-sources/scb-befolkning.md` and `scb-kostndrlt.md` (markers added in Tasks 1–2):

```yaml
verification_status: verified
verification_notes: ""
last_verified: 2026-06-07
```

- [ ] **Step 3: Promote wikimedia-svg (attribution now present)**

In `docs/data-sources/wikimedia-svg.md`:

```yaml
verification_status: frozen
verification_notes: "Statisk SVG; CC BY-SA 2.5-attribution renderas nu i kartvyerna."
last_verified: 2026-06-07
```

- [ ] **Step 4: Investigate statskontoret-arsutfall drift**

Run: `grep -n 'upstream\|GetFile\|psidata\|verification' docs/data-sources/statskontoret-arsutfall.md`
Open the documented landing page `https://statskontoret.se/psidata/arsutfall` in a browser (or `curl -sI`). Find the current ZIP/GetFile link. Compare against the URL the backend client uses:

Run: `grep -rn 'psidata\|arsutfall\|GetFile\|http' backend/internal/riksdag/adapters/ 2>/dev/null`

- [ ] **Step 5: Resolve statskontoret-arsutfall status**

- If the backend still fetches the correct current file → update the MD prose to describe the current navigation path (no curl), set `verification_status: verified`, `last_verified: 2026-06-07`, empty notes.
- If the link genuinely drifted and the backend would now fetch stale/wrong data → keep `verification_status: unsure`, and rewrite `verification_notes` to: `"Statskontorets GetFile-länk för årsutfall-ZIP har ändrats; backend-klienten måste uppdateras innan källan kan verifieras."` Leave the authority `<SourceMarker>`s (added in Task 2) in place — the badge will correctly show amber.

- [ ] **Step 6: Regenerate the registry**

Run: `cd frontend && npm run sync:data-sources`
Expected: `sync-data-sources: N sources → src/components/sources/SourceRegistry.generated.ts`, no error. (The sync fails if any `unsure` source has empty `verification_notes` — statskontoret-arsutfall, if left `unsure`, must keep a non-empty note from Step 5.)

- [ ] **Step 7: Verify source liveness**

Run: `cd frontend && npm run verify:sources`
Expected: exit 0; report lists the promoted sources as reachable. If `statskontoret-arsutfall` fails liveness, that is expected when left `unsure`.

- [ ] **Step 8: Typecheck (registry type changes flow through)**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add docs/data-sources frontend/src/data-sources frontend/src/components/sources/SourceRegistry.generated.ts
git commit -m "feat(sources): promote verified badges; document statskontoret-arsutfall status"
```

---

## Stream B — Value verification command

### Task 4: Check primitives — Status, Result, classify

**Files:**
- Create: `backend/cmd/verifyvalues/check.go`
- Test: `backend/cmd/verifyvalues/check_test.go`

- [ ] **Step 1: Write the failing test**

```go
package main

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		name     string
		ours     string
		upstream string
		volatile bool
		want     Status
	}{
		{"equal non-volatile", "283649", "283649", false, StatusMatch},
		{"differ non-volatile", "283649", "283700", false, StatusReview},
		{"equal volatile", "41233", "41233", true, StatusInfo},
		{"differ volatile", "41233", "41980", true, StatusInfo},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := classify(c.ours, c.upstream, c.volatile)
			if got != c.want {
				t.Fatalf("classify(%q,%q,%v) = %q, want %q", c.ours, c.upstream, c.volatile, got, c.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./cmd/verifyvalues/ -run TestClassify -v`
Expected: FAIL — `undefined: classify` / `undefined: Status`.

- [ ] **Step 3: Write minimal implementation**

```go
package main

// Status is the verdict for one compared value.
type Status string

const (
	StatusMatch  Status = "match"  // ours == upstream
	StatusReview Status = "review" // ours != upstream, needs human eyes
	StatusInfo   Status = "info"   // volatile metric, recorded but never flagged
	StatusError  Status = "error"  // upstream fetch failed
)

// Result is one (entity, metric) comparison row in the report.
type Result struct {
	Tier     string // "region" | "kommun" | "riksdag"
	Entity   string // "Dalarna (20)" | "Stockholm (0180)" | "Riksdag"
	Metric   string // "population 2024" | "mandate total" | "KPI N60008 2023"
	SourceID string // registry id, e.g. "scb-befolkning"
	Ours     string
	Upstream string
	Status   Status
}

// classify compares two canonical string values. Volatile metrics (e.g.
// speech counts) are always Info: their delta is recorded but never flagged.
func classify(ours, upstream string, volatile bool) Status {
	if volatile {
		return StatusInfo
	}
	if ours == upstream {
		return StatusMatch
	}
	return StatusReview
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./cmd/verifyvalues/ -run TestClassify -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/cmd/verifyvalues/check.go backend/cmd/verifyvalues/check_test.go
git commit -m "feat(verifyvalues): add Result type and classify"
```

---

### Task 5: Report rendering

**Files:**
- Create: `backend/cmd/verifyvalues/report.go`
- Test: `backend/cmd/verifyvalues/report_test.go`

- [ ] **Step 1: Write the failing test**

```go
package main

import (
	"strings"
	"testing"
	"time"
)

func TestRenderReport(t *testing.T) {
	results := []Result{
		{Tier: "region", Entity: "Dalarna (20)", Metric: "population 2024", SourceID: "scb-befolkning", Ours: "287966", Upstream: "287966", Status: StatusMatch},
		{Tier: "kommun", Entity: "Malmö (1280)", Metric: "population 2024", SourceID: "scb-befolkning", Ours: "357377", Upstream: "357891", Status: StatusReview},
		{Tier: "riksdag", Entity: "Riksdag", Metric: "anföranden total", SourceID: "riksdagen", Ours: "41233", Upstream: "41980", Status: StatusInfo},
	}
	out := renderReport(results, time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC))

	if !strings.Contains(out, "# Value Verification — 2026-06-07") {
		t.Errorf("missing dated title:\n%s", out)
	}
	// Needs-review summary must contain the Malmö review row...
	reviewSection := section(out, "## ⚠ Needs review")
	if !strings.Contains(reviewSection, "Malmö (1280)") {
		t.Errorf("review summary missing Malmö row:\n%s", reviewSection)
	}
	// ...and must NOT contain the matching Dalarna row.
	if strings.Contains(reviewSection, "Dalarna") {
		t.Errorf("review summary wrongly includes a Match row:\n%s", reviewSection)
	}
	// Info row goes under Informational, not Needs review.
	if strings.Contains(reviewSection, "anföranden") {
		t.Errorf("review summary wrongly includes an Info row:\n%s", reviewSection)
	}
	if !strings.Contains(out, "### Informational") || !strings.Contains(out, "anföranden total") {
		t.Errorf("missing Informational section:\n%s", out)
	}
}

// section returns the substring from the given heading up to the next "## "
// heading (test helper).
func section(doc, heading string) string {
	i := strings.Index(doc, heading)
	if i < 0 {
		return ""
	}
	rest := doc[i+len(heading):]
	if j := strings.Index(rest, "\n## "); j >= 0 {
		return rest[:j]
	}
	return rest
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./cmd/verifyvalues/ -run TestRenderReport -v`
Expected: FAIL — `undefined: renderReport`.

- [ ] **Step 3: Write minimal implementation**

```go
package main

import (
	"fmt"
	"strings"
	"time"
)

// renderReport produces the dated markdown audit document. The "Needs review"
// summary (Status==Review only) goes on top so problems are seen first;
// per-tier tables follow; volatile Info rows are grouped at the end.
func renderReport(results []Result, runAt time.Time) string {
	var b strings.Builder
	day := runAt.Format("2006-01-02")
	fmt.Fprintf(&b, "# Value Verification — %s\n\n", day)
	fmt.Fprintf(&b, "Run: %s UTC\n\n", runAt.Format("2006-01-02 15:04:05"))

	// Needs review summary.
	b.WriteString("## ⚠ Needs review")
	var review []Result
	for _, r := range results {
		if r.Status == StatusReview {
			review = append(review, r)
		}
	}
	fmt.Fprintf(&b, " (%d)\n\n", len(review))
	if len(review) == 0 {
		b.WriteString("None — every checked value matches upstream.\n\n")
	} else {
		writeTable(&b, []string{"Tier", "Entity", "Metric", "Ours", "Upstream", "Source"}, review,
			func(r Result) []string { return []string{r.Tier, r.Entity, r.Metric, r.Ours, r.Upstream, r.SourceID} })
	}

	// Per-tier tables (non-Info).
	for _, tier := range []string{"region", "kommun", "riksdag"} {
		var rows []Result
		for _, r := range results {
			if r.Tier == tier && r.Status != StatusInfo {
				rows = append(rows, r)
			}
		}
		if len(rows) == 0 {
			continue
		}
		fmt.Fprintf(&b, "## %s (%d)\n\n", strings.Title(tier), len(rows))
		writeTable(&b, []string{"Entity", "Metric", "Ours", "Upstream", "Status", "Source"}, rows,
			func(r Result) []string { return []string{r.Entity, r.Metric, r.Ours, r.Upstream, string(r.Status), r.SourceID} })
	}

	// Informational (volatile).
	var info []Result
	for _, r := range results {
		if r.Status == StatusInfo {
			info = append(info, r)
		}
	}
	if len(info) > 0 {
		b.WriteString("### Informational (volatile)\n\n")
		writeTable(&b, []string{"Entity", "Metric", "Ours", "Upstream", "Source"}, info,
			func(r Result) []string { return []string{r.Entity, r.Metric, r.Ours, r.Upstream, r.SourceID} })
	}

	return b.String()
}

func writeTable(b *strings.Builder, headers []string, rows []Result, cols func(Result) []string) {
	b.WriteString("| " + strings.Join(headers, " | ") + " |\n")
	b.WriteString("|" + strings.Repeat("---|", len(headers)) + "\n")
	for _, r := range rows {
		b.WriteString("| " + strings.Join(cols(r), " | ") + " |\n")
	}
	b.WriteString("\n")
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./cmd/verifyvalues/ -run TestRenderReport -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/cmd/verifyvalues/report.go backend/cmd/verifyvalues/report_test.go
git commit -m "feat(verifyvalues): add markdown report renderer"
```

---

### Task 6: Expose mandate totals from the seeder

The seeder fetches council-seat mandates from SCB on startup via unexported
helpers (`getMeta`, `codesAndNames`, `fetchMandates`, `sumMandates`). The
verify command needs the same live data, keyed by entity **name** (avoids
reproducing the region code→SCB code mapping). Add one exported wrapper.

**Files:**
- Modify: `backend/internal/regions/seeder/seeder.go`
- Test: `backend/internal/regions/seeder/mandates_export_test.go`

- [ ] **Step 1: Confirm the helpers to reuse**

Run: `grep -n 'func getMeta\|func codesAndNames\|func fetchMandates\|func sumMandates\|kfmandatURL\|ltmandatURL\|ME0104C1\|ME0104C2' backend/internal/regions/seeder/seeder.go`
Expected: shows `getMeta(ctx, c, url)`, `codesAndNames(meta) (codes, names []string)`, `fetchMandates(ctx, c, url, codes, contentsCode) (map[string]map[string]int, error)`, `sumMandates(map[string]int) int`, and the two URLs + contents codes (`ME0104C1` mun, `ME0104C2` region).

- [ ] **Step 2: Write the failing test**

```go
package seeder

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestFetchMandateTotals_LiveSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("hits live SCB; run without -short")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	mun, reg, err := FetchMandateTotals(ctx, &http.Client{Timeout: 20 * time.Second})
	if err != nil {
		t.Fatalf("FetchMandateTotals: %v", err)
	}
	if len(reg) < 20 || len(reg) > 25 {
		t.Errorf("region count = %d, want ~21", len(reg))
	}
	if len(mun) < 250 {
		t.Errorf("municipality count = %d, want ~290", len(mun))
	}
	// Every total must be positive.
	for name, total := range reg {
		if total <= 0 {
			t.Errorf("region %q total = %d", name, total)
		}
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd backend && go test ./internal/regions/seeder/ -run TestFetchMandateTotals -v`
Expected: FAIL — `undefined: FetchMandateTotals`.

- [ ] **Step 4: Add the exported wrapper**

Append to `backend/internal/regions/seeder/seeder.go`:

```go
// FetchMandateTotals returns total elected council seats per entity NAME,
// fetched live from SCB's Kfmandat (municipalities) and Ltmandat (regions)
// tables. It reuses the same PxWeb calls the startup seeder uses. Keyed by
// name so callers need not reproduce the region code→SCB code mapping.
func FetchMandateTotals(ctx context.Context, c *http.Client) (mun, reg map[string]int, err error) {
	munMeta, err := getMeta(ctx, c, kfmandatURL)
	if err != nil {
		return nil, nil, fmt.Errorf("kfmandat meta: %w", err)
	}
	munCodes, munNames := codesAndNames(munMeta)
	munByCode, err := fetchMandates(ctx, c, kfmandatURL, munCodes, "ME0104C1")
	if err != nil {
		return nil, nil, fmt.Errorf("kfmandat: %w", err)
	}
	regMeta, err := getMeta(ctx, c, ltmandatURL)
	if err != nil {
		return nil, nil, fmt.Errorf("ltmandat meta: %w", err)
	}
	regCodes, regNames := codesAndNames(regMeta)
	regByCode, err := fetchMandates(ctx, c, ltmandatURL, regCodes, "ME0104C2")
	if err != nil {
		return nil, nil, fmt.Errorf("ltmandat: %w", err)
	}

	mun = totalsByName(munCodes, munNames, munByCode)
	reg = totalsByName(regCodes, regNames, regByCode)
	return mun, reg, nil
}

// totalsByName maps each code's per-party mandate map to a name→total entry.
func totalsByName(codes, names []string, byCode map[string]map[string]int) map[string]int {
	out := make(map[string]int, len(codes))
	for i, code := range codes {
		name := code
		if i < len(names) {
			name = names[i]
		}
		out[name] = sumMandates(byCode[code])
	}
	return out
}
```

If `net/http`, `context`, or `fmt` are not already imported in `seeder.go`, add them (run `cd backend && goimports -w internal/regions/seeder/seeder.go` or add manually).

- [ ] **Step 5: Run the test (live)**

Run: `cd backend && go test ./internal/regions/seeder/ -run TestFetchMandateTotals -v`
Expected: PASS (~21 regions, ~290 municipalities). If SCB is unreachable in the sandbox, run later against the dev stack; the build must still compile:
Run: `go build -C backend ./...` → Expected: success.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/regions/seeder/seeder.go backend/internal/regions/seeder/mandates_export_test.go
git commit -m "feat(seeder): export FetchMandateTotals for value verification"
```

---

### Task 7: Region + kommun checks

Builds `[]Result` for population (vs `scb.FetchPopulationTrend`), mandate total
(vs `seeder.FetchMandateTotals`), and KPI/spending (vs
`kolada.FetchKPIAllMunicipalities`). Upstream-fetch failures become a
`StatusError` row, never a panic.

**Files:**
- Create: `backend/cmd/verifyvalues/regions.go`

- [ ] **Step 1: Confirm the service + client signatures**

Run: `grep -n 'func (s \*Service) ListRegions\|func (s \*Service) ListMunicipalities\|func (s \*Service) GetMunicipalityKPIs\|func (s \*Service) GetRegionKPIs' backend/internal/regions/service.go`
Run: `grep -n 'regionStripKPIs\|spendingKPIs\|stripKPIs' backend/internal/regions/service.go`
Expected: confirms `ListRegions(ctx) ([]*domain.Region, error)`, `ListMunicipalities(ctx, regionCode string) ([]*domain.Municipality, error)`, and the KPI-code slice names.

- [ ] **Step 2: Write the checks file**

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"riksdagskollen/internal/regions"
	"riksdagskollen/internal/regions/adapters/kolada"
	"riksdagskollen/internal/regions/adapters/scb"
	"riksdagskollen/internal/regions/seeder"
)

// regionStripKPIs and spendingKPIs mirror the codes the UI renders.
// Kept in sync with backend/internal/regions/service.go.
var verifyRegionKPIs = []string{"N60008", "N63016", "N63007", "N79173", "N79179", "N60404", "N85012"}
var verifyKommunKPIs = []string{"N00900", "N03102", "N03106", "N15428", "N00708"}
var verifySpendingKPIs = []string{"N11004", "N15028", "N17014", "N20014", "N30005", "N07037", "N09022", "N05011", "N45014"}

// regionAndKommunChecks compares DB values to live SCB/Kolada values.
func regionAndKommunChecks(ctx context.Context, svc *regions.Service, scbCl *scb.Client, koladaCl *kolada.Client) []Result {
	var out []Result
	year := time.Now().Year() - 1 // KPI/population publish with a lag

	// Live mandate totals (name-keyed), fetched once.
	munMandates, regMandates, mErr := seeder.FetchMandateTotals(ctx, &http.Client{Timeout: 30 * time.Second})

	// --- Regions: population + mandate ---
	regionsList, err := svc.ListRegions(ctx)
	if err == nil {
		for _, r := range regionsList {
			entity := fmt.Sprintf("%s (%s)", r.Name, r.Code)
			out = append(out, populationCheck(ctx, "region", entity, r.Code, r.Population, year, scbCl))
			out = append(out, mandateCheck("region", entity, r.Name, r.TotalMandates, regMandates, mErr))
		}
	} else {
		out = append(out, Result{Tier: "region", Entity: "(all)", Metric: "list", SourceID: "scb-befolkning", Status: StatusError, Upstream: err.Error()})
	}

	// --- Kommuner: population + mandate ---
	munList, err := svc.ListMunicipalities(ctx, "")
	if err == nil {
		for _, m := range munList {
			entity := fmt.Sprintf("%s (%s)", m.Name, m.Code)
			out = append(out, populationCheck(ctx, "kommun", entity, m.Code, m.Population, year, scbCl))
			out = append(out, mandateCheck("kommun", entity, m.Name, m.TotalMandates, munMandates, mErr))
		}
	} else {
		out = append(out, Result{Tier: "kommun", Entity: "(all)", Metric: "list", SourceID: "scb-kfmandat", Status: StatusError, Upstream: err.Error()})
	}

	// --- KPI / spending (Kolada batch: one call per KPI returns all entities) ---
	out = append(out, kpiChecks(ctx, "kommun", append(verifyKommunKPIs, verifySpendingKPIs...), year, svc, koladaCl, false)...)
	out = append(out, kpiChecks(ctx, "region", verifyRegionKPIs, year, svc, koladaCl, true)...)

	return out
}

func populationCheck(ctx context.Context, tier, entity, code string, ours, year int, scbCl *scb.Client) Result {
	r := Result{Tier: tier, Entity: entity, Metric: fmt.Sprintf("population %d", year), SourceID: "scb-befolkning", Ours: strconv.Itoa(ours)}
	entries, err := scbCl.FetchPopulationTrend(ctx, code, []int{year})
	if err != nil || len(entries) == 0 {
		r.Status = StatusError
		if err != nil {
			r.Upstream = err.Error()
		} else {
			r.Upstream = "no data"
		}
		return r
	}
	r.Upstream = strconv.Itoa(entries[len(entries)-1].Population)
	r.Status = classify(r.Ours, r.Upstream, false)
	return r
}

func mandateCheck(tier, entity, name string, ours int, byName map[string]int, fetchErr error) Result {
	src := "scb-ltmandat"
	if tier == "kommun" {
		src = "scb-kfmandat"
	}
	r := Result{Tier: tier, Entity: entity, Metric: "mandate total", SourceID: src, Ours: strconv.Itoa(ours)}
	if fetchErr != nil {
		r.Status = StatusError
		r.Upstream = fetchErr.Error()
		return r
	}
	up, ok := byName[name]
	if !ok {
		r.Status = StatusError
		r.Upstream = "name not found in SCB mandate set"
		return r
	}
	r.Upstream = strconv.Itoa(up)
	r.Status = classify(r.Ours, r.Upstream, false)
	return r
}

// kpiChecks compares each entity's DB KPI value against Kolada's batch result.
// isRegion selects GetRegionKPIs vs GetMunicipalityKPIs for the "ours" side.
func kpiChecks(ctx context.Context, tier string, kpiCodes []string, year int, svc *regions.Service, koladaCl *kolada.Client, isRegion bool) []Result {
	var out []Result
	for _, kpi := range kpiCodes {
		upstream, err := koladaCl.FetchKPIAllMunicipalities(ctx, kpi, []int{year})
		if err != nil {
			out = append(out, Result{Tier: tier, Entity: "(all)", Metric: "KPI " + kpi, SourceID: "kolada", Status: StatusError, Upstream: err.Error()})
			continue
		}
		upByCode := make(map[string]float64, len(upstream))
		for _, u := range upstream {
			upByCode[u.MunCode] = u.Value
		}
		// "Ours" side: pull each entity's stored KPI value for this code.
		// To bound cost, we compare only entities present in the upstream batch.
		for code, upVal := range upByCode {
			var ours float64
			var found bool
			var oursList, oerr = svc.GetMunicipalityKPIs(ctx, code)
			if isRegion {
				oursList, oerr = svc.GetRegionKPIs(ctx, code)
			}
			if oerr != nil {
				continue
			}
			for _, kv := range oursList {
				if kv.KPI == kpi {
					ours, found = kv.Value, true
					break
				}
			}
			if !found {
				continue
			}
			oursStr := strconv.FormatFloat(ours, 'f', -1, 64)
			upStr := strconv.FormatFloat(upVal, 'f', -1, 64)
			out = append(out, Result{
				Tier:     tier,
				Entity:   code,
				Metric:   fmt.Sprintf("KPI %s %d", kpi, year),
				SourceID: "kolada",
				Ours:     oursStr,
				Upstream: upStr,
				Status:   classify(oursStr, upStr, false),
			})
		}
	}
	return out
}
```

- [ ] **Step 3: Verify it compiles**

Run: `go build -C backend ./...`
Expected: success. Fix any signature mismatches surfaced (e.g. if `GetRegionKPIs` expects a region code format — adjust the loop key accordingly; the compiler/smoke run will reveal it).

- [ ] **Step 4: Commit**

```bash
git add backend/cmd/verifyvalues/regions.go
git commit -m "feat(verifyvalues): region + kommun population/mandate/KPI checks"
```

---

### Task 8: Riksdagen checks

Compares party seat totals + roster counts (vs `politicians` client), sample
vote outcomes (vs `votes` client), and speech counts (informational, vs
`speeches` client).

**Files:**
- Create: `backend/cmd/verifyvalues/riksdagen.go`

- [ ] **Step 1: Confirm signatures**

Run: `grep -n 'func (c \*Client) FetchMembers\|func (c \*Client) FetchVotes\|func (c \*Client) FetchSpeeches' backend/internal/politicians/adapters/riksdagen/client.go backend/internal/votes/adapters/riksdagen/client.go backend/internal/speeches/adapters/riksdagen/client.go`
Run: `grep -n 'func (s \*Service) List\b\|func (s \*Service) ListByBeteckning' backend/internal/politicians/service.go backend/internal/votes/service.go`
Run: `grep -n 'type FetchVotesFilter\|type FetchSpeechesFilter' backend/internal/votes/ports/riksdagen.go backend/internal/speeches/ports/*.go`
Expected: `politicians.Client.FetchMembers(ctx, party, status string) ([]*domain.Politician, error)`, `votes.Client.FetchVotes(ctx, FetchVotesFilter) ([]*domain.Vote, error)` with `Beteckning` field, `votes.Service.ListByBeteckning(ctx, bet, punkt string) ([]*domain.Vote, error)`, `politicians.Service.List(ctx, ports.ListFilter) (ports.ListResult, error)`.

- [ ] **Step 2: Write the checks file**

```go
package main

import (
	"context"
	"fmt"
	"strconv"

	polrd "riksdagskollen/internal/politicians/adapters/riksdagen"
	politicians "riksdagskollen/internal/politicians"
	polports "riksdagskollen/internal/politicians/ports"
	speechrd "riksdagskollen/internal/speeches/adapters/riksdagen"
	speeches "riksdagskollen/internal/speeches"
	voterd "riksdagskollen/internal/votes/adapters/riksdagen"
	votes "riksdagskollen/internal/votes"
	voteports "riksdagskollen/internal/votes/ports"
	votedomain "riksdagskollen/internal/votes/domain"
)

// riksdagParties are the eight Riksdag parties whose seat counts we verify.
var riksdagParties = []string{"S", "M", "SD", "C", "V", "KD", "L", "MP"}

// sampleVotes are fixed (beteckning, punkt) pairs whose tallies we spot-check.
// Pick ids known to exist in the seeded DB; adjust after the first smoke run.
var sampleVotes = []struct{ Beteckning, Punkt string }{
	{"FiU1", "1"},
}

func riksdagenChecks(
	ctx context.Context,
	polSvc *politicians.Service, polCl *polrd.Client,
	voteSvc *votes.Service, voteCl *voterd.Client,
	speechCl *speechrd.Client,
) []Result {
	var out []Result

	// --- Party seat totals + roster counts (per party) ---
	for _, party := range riksdagParties {
		ours, oerr := polSvc.List(ctx, polports.ListFilter{Party: party, ActiveOnly: true, Page: 1, PageSize: 1})
		upstream, uerr := polCl.FetchMembers(ctx, party, "tjanstgorande")
		r := Result{Tier: "riksdag", Entity: "Party " + party, Metric: "seat total", SourceID: "riksdagen"}
		switch {
		case oerr != nil:
			r.Status, r.Upstream = StatusError, oerr.Error()
		case uerr != nil:
			r.Status, r.Ours, r.Upstream = StatusError, strconv.Itoa(ours.Total), uerr.Error()
		default:
			r.Ours = strconv.Itoa(ours.Total)
			r.Upstream = strconv.Itoa(len(upstream))
			r.Status = classify(r.Ours, r.Upstream, false)
		}
		out = append(out, r)
	}

	// --- Roster grand total ---
	allOurs, oerr := polSvc.List(ctx, polports.ListFilter{ActiveOnly: true, Page: 1, PageSize: 1})
	r := Result{Tier: "riksdag", Entity: "Riksdag", Metric: "active member roster", SourceID: "riksdagen"}
	if oerr != nil {
		r.Status, r.Upstream = StatusError, oerr.Error()
	} else {
		all, uerr := polCl.FetchMembers(ctx, "", "tjanstgorande")
		r.Ours = strconv.Itoa(allOurs.Total)
		if uerr != nil {
			r.Status, r.Upstream = StatusError, uerr.Error()
		} else {
			r.Upstream = strconv.Itoa(len(all))
			r.Status = classify(r.Ours, r.Upstream, false)
		}
	}
	out = append(out, r)

	// --- Sample vote outcomes ---
	for _, sv := range sampleVotes {
		oursVotes, oerr := voteSvc.ListByBeteckning(ctx, sv.Beteckning, sv.Punkt)
		upVotes, uerr := voteCl.FetchVotes(ctx, voteports.FetchVotesFilter{Beteckning: sv.Beteckning})
		entity := fmt.Sprintf("Vote %s p%s", sv.Beteckning, sv.Punkt)
		if oerr != nil || uerr != nil {
			msg := ""
			if oerr != nil {
				msg = oerr.Error()
			} else {
				msg = uerr.Error()
			}
			out = append(out, Result{Tier: "riksdag", Entity: entity, Metric: "vote tally", SourceID: "riksdagen", Status: StatusError, Upstream: msg})
			continue
		}
		out = append(out, Result{
			Tier: "riksdag", Entity: entity, Metric: "vote tally (Ja/Nej/Avstår/Frånv)", SourceID: "riksdagen",
			Ours:     tally(oursVotes),
			Upstream: tally(upVotes),
			Status:   classify(tally(oursVotes), tally(upVotes), false),
		})
	}

	// --- Speech counts (informational/volatile) ---
	recent, err := speechCl.FetchSpeeches(ctx, speeches.RecentFilter())
	sc := Result{Tier: "riksdag", Entity: "Riksdag", Metric: "anföranden (senaste hämtning)", SourceID: "riksdagen"}
	if err != nil {
		sc.Status, sc.Upstream = StatusError, err.Error()
	} else {
		sc.Upstream = strconv.Itoa(len(recent))
		sc.Ours = "n/a (volatile)"
		sc.Status = classify(sc.Ours, sc.Upstream, true) // always Info
	}
	out = append(out, sc)

	return out
}

// tally counts a slice of votes into a canonical "Ja/Nej/Avstår/Frånvarande"
// string for comparison.
func tally(vs []*votedomain.Vote) string {
	var ja, nej, av, fr int
	for _, v := range vs {
		switch v.VoteResult {
		case votedomain.VoteJa:
			ja++
		case votedomain.VoteNej:
			nej++
		case votedomain.VoteAvstar:
			av++
		case votedomain.VoteFranvarande:
			fr++
		}
	}
	return fmt.Sprintf("%d/%d/%d/%d", ja, nej, av, fr)
}
```

- [ ] **Step 3: Resolve the speech-filter helper**

The speech client needs a `FetchSpeechesFilter`. Inspect the existing filter:
Run: `grep -n 'type FetchSpeechesFilter\|RecentFilter\|func.*Filter' backend/internal/speeches/ports/*.go backend/internal/speeches/*.go`
If no `RecentFilter()` helper exists, replace `speeches.RecentFilter()` with a literal `speechports.FetchSpeechesFilter{...}` populated from the struct's actual fields (add `speechports "riksdagskollen/internal/speeches/ports"` to imports). Use the smallest filter that returns a recent batch (e.g. current session, limit 200).

- [ ] **Step 4: Verify it compiles**

Run: `go build -C backend ./...`
Expected: success. Adjust import aliases / filter fields until it builds.

- [ ] **Step 5: Commit**

```bash
git add backend/cmd/verifyvalues/riksdagen.go
git commit -m "feat(verifyvalues): riksdagen seat/roster/vote/speech checks"
```

---

### Task 9: Wiring, report write, Makefile target, smoke

**Files:**
- Create: `backend/cmd/verifyvalues/main.go`
- Modify: `Makefile`

- [ ] **Step 1: Confirm constructors used in cmd/api**

Run: `grep -n 'NewService\|NewClient\|pgxpool.New\|postgres.New\|Repository' backend/cmd/api/main.go | head -40`
Expected: shows how each feature's repo + client + service is constructed (e.g. `regions.NewService(regionsRepo, koladaCl, scbCl, tedCl)`, `politicians.NewService(polRepo, polRD)`, etc.). Mirror exactly the same construction here.

- [ ] **Step 2: Write main.go**

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	politicians "riksdagskollen/internal/politicians"
	polpg "riksdagskollen/internal/politicians/adapters/postgres"
	polrd "riksdagskollen/internal/politicians/adapters/riksdagen"
	regions "riksdagskollen/internal/regions"
	"riksdagskollen/internal/regions/adapters/kolada"
	regpg "riksdagskollen/internal/regions/adapters/postgres"
	"riksdagskollen/internal/regions/adapters/scb"
	"riksdagskollen/internal/regions/adapters/ted"
	speeches "riksdagskollen/internal/speeches"
	speechpg "riksdagskollen/internal/speeches/adapters/postgres"
	speechrd "riksdagskollen/internal/speeches/adapters/riksdagen"
	votes "riksdagskollen/internal/votes"
	votepg "riksdagskollen/internal/votes/adapters/postgres"
	voterd "riksdagskollen/internal/votes/adapters/riksdagen"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(2)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db connect: %v\n", err)
		os.Exit(2)
	}
	defer pool.Close()

	// --- Construct services + upstream clients (mirror cmd/api/main.go).
	// NOTE: adjust constructor names/args to match Step 1 output exactly.
	scbCl := scb.NewClient("")
	koladaCl := kolada.NewClient()
	tedCl := ted.NewClient()
	regionsSvc := regions.NewService(regpg.NewRepository(pool), koladaCl, scbCl, tedCl)

	polCl := polrd.NewClient()
	polSvc := politicians.NewService(polpg.NewRepository(pool), polCl)

	voteCl := voterd.NewClient()
	voteSvc := votes.NewService(votepg.NewRepository(pool), voteCl)

	speechCl := speechrd.NewClient()
	_ = speeches.NewService(speechpg.NewRepository(pool), speechCl) // service not needed; client used directly

	// --- Run all checks.
	var results []Result
	results = append(results, regionAndKommunChecks(ctx, regionsSvc, scbCl, koladaCl)...)
	results = append(results, riksdagenChecks(ctx, polSvc, polCl, voteSvc, voteCl, speechCl)...)

	// --- Write the dated report.
	now := time.Now().UTC()
	doc := renderReport(results, now)
	outPath := filepath.Join("docs", "superpowers", "audits",
		fmt.Sprintf("%s-value-verification.md", now.Format("2006-01-02")))
	if err := os.WriteFile(outPath, []byte(doc), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write report: %v\n", err)
		os.Exit(1)
	}

	// --- Summary to stdout.
	var review int
	for _, r := range results {
		if r.Status == StatusReview {
			review++
		}
	}
	fmt.Printf("verifyvalues: %d checks, %d need review → %s\n", len(results), review, outPath)
	_ = http.DefaultClient // keep net/http imported if unused after edits
}
```

- [ ] **Step 3: Fix constructor names**

Run: `go build -C backend ./...`
Expected: compiler errors naming the real constructor signatures (e.g. `regpg.NewRepository` may actually be `postgres.New`). Replace each with the exact names from Step 1 until it builds. Remove the `_ = http.DefaultClient` line and the `net/http` import if unused.
Final: `go build -C backend ./...` → success.

- [ ] **Step 4: Add the Makefile target**

Append to `Makefile` (use a tab for the recipe line, matching the existing `run`/`migrate` targets):

```makefile
verify-values:
	docker compose -f docker-compose.dev.yml run --rm \
		-v $(PWD):/repo -w /repo/backend \
		backend go run ./cmd/verifyvalues
```

Confirm the dev compose service name is `backend`:
Run: `grep -n 'backend:\|services:' docker-compose.dev.yml | head`
Adjust the service name and mount if they differ. The mount makes the report land in the repo's `docs/superpowers/audits/`.

- [ ] **Step 5: Run the unit tests**

Run: `cd backend && go test ./cmd/verifyvalues/ -v`
Expected: PASS (classify + renderReport).

- [ ] **Step 6: Smoke run against the dev stack**

Run: `make run` in one shell (postgres + backend up, DB seeded), then in another:
Run: `make verify-values`
Expected: `verifyvalues: N checks, M need review → docs/superpowers/audits/2026-06-07-value-verification.md`. Open the report; confirm Regions, Kommuner, Riksdagen tables are populated.

- [ ] **Step 7: Triage the first report**

For each row in "⚠ Needs review", determine: is our ingestion wrong (fix it) or is it a legitimate upstream revision / year-alignment artefact (note it)? Common first-run issues:
  - **All population rows flagged** → year mismatch. Confirm the DB stores which year; align the `year` variable in `regions.go` to it.
  - **All mandate rows flagged for one tier** → name-key mismatch (e.g. "Region Stockholm" vs "Stockholm"). Add a `normalizeName` helper in `regions.go` applied to both sides.
  - **`StatusError` on a sample vote** → the `sampleVotes` id is not in the DB; replace with a real `(beteckning, punkt)` from `SELECT DISTINCT beteckning, forslagspunkt FROM votes LIMIT 5`.

- [ ] **Step 8: Commit**

```bash
git add backend/cmd/verifyvalues/main.go Makefile docs/superpowers/audits/2026-06-07-value-verification.md
git commit -m "feat(verifyvalues): wire command, add make target, commit first report"
```

---

## Self-Review notes (addressed)

- **Spec coverage:** Stream A → Tasks 1–3 (markers + promotions + statskontoret). Stream B tiers → Task 7 (region/kommun population, mandate, KPI/spending), Task 8 (riksdagen seats/roster/votes/speeches). Report format → Task 5. On-demand run + committed report → Task 9. Mandate-without-client gap → Task 6.
- **Volatile speech counts:** `classify(..., true)` always returns `StatusInfo` (Task 4 test asserts this) and renders under "Informational" (Task 5 test asserts this).
- **No hard fail:** `main.go` exits 0 on completion regardless of review count; only I/O/DB errors exit non-zero. Matches the review-not-fail decision.
- **Type consistency:** `Result`/`Status`/`classify` defined in Task 4 are used unchanged in Tasks 5,7,8,9. `FetchMandateTotals` (Task 6) signature matches its call in Task 7.
- **Known iteration points** (flagged inline, resolved at smoke time, not placeholders): exact repo constructor names (Task 9 Step 3), population year alignment + mandate name normalization + real sample-vote ids (Task 9 Step 7). These require the live DB to resolve and are explicit verification steps, not deferred design.
```
