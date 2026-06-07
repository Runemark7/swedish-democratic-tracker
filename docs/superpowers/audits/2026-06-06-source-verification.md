# Source Verification Audit — 2026-06-06

Pass two. Tracks every registered source against the verification criteria
defined in `docs/superpowers/specs/2026-06-06-source-verification-design.md`.

## Main table

| id | URL liveness | Schema match | Repro narrative OK? | UI consumers | Backend client | Proposed status | Gaps |
|---|---|---|---|---|---|---|---|
| derived-agenda | n/a (synthesized) | n/a | Yes | AgendaDetailPage, RiksdagPage, region/municipality detail pages | derived | frozen | — |
| kolada | HTTP 200 (text/html) | Yes — `values[].kpi, values[].period, values[].values[].gender/value/status/isdeleted` matches struct | No | RegionDetailPage, MunicipalityDetailPage, RegionLandingPage, MunicipalityLandingPage | `internal/regions/adapters/kolada` | unsure | MD has `bash` code blocks with curl commands (forbidden) |
| kolada-spending | HTTP 200 (text/html) | Yes — same struct as kolada | No | MunicipalityDetailPage (sourceId="kolada-spending"), KommunBudgetAreaPage | `internal/regions/adapters/kolada` | unsure | MD has `bash` code block with curl (forbidden) |
| riksdagen | HTTP 200 (text/html) | Yes — politicians/speeches/votes clients all parse JSON from data.riksdagen.se | No | HomePage, VotesPage, VoteDetailPage, PartiesPage, PartyGoalsPage, GoalVotesPage, PoliticianPage, RiksdagPage, RegionDetailPage, MunicipalityDetailPage, ManifestosPage | `internal/politicians/adapters/riksdagen`, `internal/speeches/adapters/riksdagen`, `internal/votes/adapters/riksdagen`, `internal/ministers/adapters/riksdagen` | unsure | MD has `bash` code blocks with 3 curl examples (forbidden) |
| scb-befolkning | HTTP 200 (application/json) | Yes — `data[].key[0]=region, data[].values[0]=population` matches client | No | RegionLandingPage, MunicipalityLandingPage (tagged); MunicipalityDetailPage sub-header population untagged | `internal/regions/adapters/scb` (FetchPopulationTrend), `internal/regions/seeder` | unsure | MD has `bash` code block with curl + POST body (forbidden); MunicipalityDetailPage:194 renders `data.population` without SourceMarker |
| scb-kfmandat | HTTP 200 (application/json) | Yes — seeder uses `data[].key[1]=party, data[].values[0]=seats` | No | MunicipalityLandingPage (tagged), MunicipalityDetailPage (MandateComposition with sourceId), RegionDetailPage (region map) | `internal/regions/seeder` | unsure | MD has 2 `bash` code blocks with curl + POST body (forbidden) |
| scb-kostndrlt | HTTP 200 (application/json) | Yes — `data[].key[0]=region, data[].key[1]=area, data[].values[0]=mnkr` matches client | No | RegionDetailPage (BudgetHistorySection sourceId="scb-kostndrlt"), RegionBudgetAreaPage (no SourceMarker on chart values) | `internal/regions/adapters/scb` (FetchRegionBudget) | unsure | MD has `bash` code block with curl + POST body (forbidden); RegionBudgetAreaPage renders mnkr values without SourceMarker |
| scb-ltmandat | HTTP 200 (application/json) | Yes — same seeder struct as scb-kfmandat | No | RegionLandingPage (tagged), RegionDetailPage (MandateComposition defaults to scb-ltmandat) | `internal/regions/seeder` | unsure | MD has `bash` code block with curl + POST body (forbidden) |
| seed-budget-data | n/a (seed) | n/a | Yes | BudgetPage, AreaHistoryPage, BudgetHistorySection, RiksdagPage | seed | frozen | — |
| seed-party-goals | n/a (seed) | n/a | Yes | PartiesPage, PartyGoalsPage, GoalVotesPage, ManifestosPage | seed | frozen | — |
| statskontoret-arsutfall | HTTP 200 (landing page OK); dynamic ZIP URL `OpenDataArsUtfallPage/GetFile?…&Year=2025` → HTTP 200 (795 kB) | Yes — `;`-delimited CSV cols 2=Anslag, 4=År, 6=Statens budget, 7=Ändringsbudgetar, 10=Utfall (tkr); matches `parseCSV` and `parseAllAnslagCSV` | No | RiksdagPage (statskontoret-arsutfall SourceMarker), AuthorityDetailPage (expenditure + history — NO SourceMarker), MyndigheterListPage (stat cards — NO SourceMarker) | `internal/riksdag/adapters/statskontoret` | unsure | MD upstream `https://www.statskontoret.se/psidata/arsutfall` serves 200 but actual download path changed to `/OpenDataArsUtfallPage/GetFile?…`; MD has `bash` block with `curl -O` + `unzip` (forbidden); AuthorityDetailPage and MyndigheterListPage show expenditure/headcount numbers without SourceMarker |
| ted | HTTP 200 (ted.europa.eu/); backend endpoint `api.ted.europa.eu/v3/notices/search` → HTTP 405 (POST-only, correct) | Yes — `notices[].publication-date, notices[].classification-cpv, notices[].total-value, notices[].total-value-cur` matches struct | No | ted sourced from SourceRegistry but no feature page renders a SourceMarker with sourceId="ted" | `internal/regions/adapters/ted` | unsure | MD upstream is `https://ted.europa.eu/` but backend uses `https://api.ted.europa.eu/v3/notices/search` — different host + path; MD curl example shows outdated `ted.europa.eu/api/v3.0/notices/search` (wrong host); MD has `bash` code block with curl (forbidden); zero UI SourceMarker consumers |
| wikimedia-svg | n/a (seed) | n/a | Yes | SwedenRegionMap, SwedenKommunMap (static SVG paths) | seed | unsure | SVG is bundled statically — no SourceMarker anywhere on region/municipality map views; CC BY-SA 2.5 attribution obligation unfulfilled in UI |

## Appendix 1 — Untagged values

UI strings quoting a number/date sourced from external data but lacking `<SourceMarker>`.

| File:line | Value | Should reference source |
|---|---|---|
| `frontend/src/features/municipalities/MunicipalityDetailPage.tsx:194` | `data.population` (population sub-header, all municipality detail pages) | `scb-befolkning` |
| `frontend/src/features/riksdag/AuthorityDetailPage.tsx:145` | `authority.expenditureMdkr.toFixed(1) mdkr` (headline cost, historik header) | `statskontoret-arsutfall` |
| `frontend/src/features/riksdag/AuthorityDetailPage.tsx:154` | `authority.headcountInt.toLocaleString` (headline headcount, historik header) | `statskontoret-myndighetsforteckning` (T18: done) |
| `frontend/src/features/riksdag/AuthorityDetailPage.tsx:176` | `row.expenditureMdkr.toFixed(1) mdkr` (per-year rows in history table) | `statskontoret-arsutfall` |
| `frontend/src/features/riksdag/AuthorityDetailPage.tsx:187` | `row.headcountInt.toLocaleString` (per-year headcount in history table) | `scb-kls-headcount` (T18: done — history bars sourced from KLS) |
| `frontend/src/features/riksdag/MyndigheterListPage.tsx:82` | `stats.totalExpenditureMdkr.toFixed(0) mdkr` (overview stat card) | `statskontoret-arsutfall` |
| `frontend/src/features/riksdag/MyndigheterListPage.tsx:88` | `stats.totalHeadcount.toLocaleString` (overview stat card) | `statskontoret-myndighetsforteckning` (T18: done) |
| `frontend/src/features/riksdag/MyndigheterListPage.tsx:172` | `a.headcountInt.toLocaleString` (per-agency row) | `statskontoret-myndighetsforteckning` (T18: done) |
| `frontend/src/features/riksdag/MyndigheterListPage.tsx:175` | `a.expenditureMdkr.toFixed(1) mdkr` (per-agency row) | `statskontoret-arsutfall` |
| `frontend/src/features/regions/RegionBudgetAreaPage.tsx:140` | `Math.round(p.value).toLocaleString("sv-SE")` (chart dot labels — mnkr values) | `scb-kostndrlt` |
| `frontend/src/features/municipalities/KommunBudgetAreaPage.tsx:143` | `Math.round(p.value).toLocaleString("sv-SE")` (chart dot labels — kr/inv values) | `kolada-spending` |
| `frontend/src/features/regering/RegeringPage.tsx` | Minister names, titles, department (entire page) | `riksdagen` |

## Appendix 2 — Untracked backend clients

External HTTP calls / CSV downloads in `backend/internal/` whose target URL
does not map to a registered source.

| Backend file:line | Target URL | Should reference source (or "new") |
|---|---|---|
| `backend/internal/riksdag/adapters/registret/client.go:188` | `https://myndighetsregistret.scb.se/Myndighet/HamtaMynd` | new: `scb-myndighetsregistret` |
| `backend/internal/riksdag/adapters/myndighetsforteckning/client.go:20` | `https://www.statskontoret.se/contentassets/bbd19bc969054c86bbc1aa757e7d5c85/statskontorets-myndighetsforteckning-2025.xlsx` | `statskontoret-myndighetsforteckning` (renamed from `scb-myndighetsforteckning` — Statskontoret is publisher, not SCB) |
| `backend/internal/riksdag/adapters/scb/client.go:18` | `https://api.scb.se/OV0104/v1/doris/sv/ssd/AM/AM0102/AM0102A/KLStabell14LpMan` | new: `scb-kls-headcount` (SCB KLS AM0102 — månadsanställda per myndighet) |
| `backend/internal/riksdag/service.go:330` | `https://www.statskontoret.se/statsliggaren/regleringsbrev` (base URL, used to generate per-agency regleringsbrev links) | informational link only — no HTTP call made to this URL; no new source needed |

## Appendix 3 — Sources not yet in registry

Discoveries from Appendix 1/2 that warrant a new MD in `docs/data-sources/`.

| Proposed id | Upstream | Notes |
|---|---|---|
| `scb-myndighetsregistret` | `https://myndighetsregistret.scb.se/Myndighet/HamtaMynd` | SCB Myndighetsregistret — full list of ~449 Swedish government agencies. Scraped HTML table. Used by `internal/riksdag/adapters/registret`. Powers agency count, name, org-number, type on MyndigheterListPage. HTTP 200 confirmed. |
| `statskontoret-myndighetsforteckning` | `https://www.statskontoret.se/contentassets/bbd19bc969054c86bbc1aa757e7d5c85/statskontorets-myndighetsforteckning-2025.xlsx` | Statskontorets Myndighetsförteckning XLSX — longitudinal headcount (årsarbetskrafter) per agency 2007–2025. Used by `internal/riksdag/adapters/myndighetsforteckning`. HTTP 200 confirmed. Note: hardcoded year in URL (2025) will need updating annually. Renamed from `scb-myndighetsforteckning` — publisher is Statskontoret. |
| `scb-kls-headcount` | `https://api.scb.se/OV0104/v1/doris/sv/ssd/AM/AM0102/AM0102A/KLStabell14LpMan` | SCB KLS tabell 14 — månadsanställda i statlig sektor per myndighet (december snapshot). Used by `internal/riksdag/adapters/scb`. Powers headcount history bars in AuthorityDetailPage. PxWeb POST endpoint; HTTP 200 confirmed. |
