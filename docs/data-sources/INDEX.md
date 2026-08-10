# Datakällor — innehållsförteckning

Alla källor som används någonstans på sidan. Frontmatter i varje fil är
maskinläsbar; `frontend/scripts/sync-data-sources.mjs` parsar dem och
skapar `SourceRegistry.generated.ts`.

| ID | Namn | Kind | Backas i UI av |
|---|---|---|---|
| [riksdagen](./riksdagen.md) | Riksdagen Open Data | api | beslut, omröstningar, anföranden |
| [kolada](./kolada.md) | Kolada API v3 | api | region/kommun KPIs |
| [kolada-spending](./kolada-spending.md) | Kolada — kommunal verksamhetsutgift | api | kommun budget |
| [scb-kfmandat](./scb-kfmandat.md) | SCB Kfmandat | api | kommunala mandat |
| [scb-ltmandat](./scb-ltmandat.md) | SCB Ltmandat | api | regionala mandat |
| [scb-kostndrlt](./scb-kostndrlt.md) | SCB KostnDRLT (regionkostnader) | api | region budget |
| [scb-befolkning](./scb-befolkning.md) | SCB BefolkningNy | api | befolkning per kommun |
| [ted](./ted.md) | TED Tenders Electronic Daily | api | upphandling |
| [statskontoret-arsutfall](./statskontoret-arsutfall.md) | Statskontoret årsutfall | csv | myndigheters driftkostnader |
| [seed-party-goals](./seed-party-goals.md) | Partimål (seed) | seed | partimål och scorecards |
| [party-manifestos](./party-manifestos.md) | Partiernas valmanifest (katalog) | seed | valmanifestkatalogen (/manifestos) |
| [seed-budget-data](./seed-budget-data.md) | Statsbudget (seed) | seed | budget per UO |
| [national-agenda](./national-agenda.md) | Nationell politisk agenda (seed) | seed | nationell agenda (/, /riksdag, /agenda/:id) |
| [region-budget-plan](./region-budget-plan.md) | Regioners mål och budget (seed) | seed | region: länk till officiell budget (/region/:code) |
| [wikimedia-svg](./wikimedia-svg.md) | Sverige-karta SVG | seed | region/kommun-kartor |
| [scb-myndighetsregistret](./scb-myndighetsregistret.md) | SCB Myndighetsregistret | api | myndighetsnamn, org-nummer, typ |
| [statskontoret-myndighetsforteckning](./statskontoret-myndighetsforteckning.md) | Statskontoret Myndighetsförteckning | csv | årsarbetskrafter per myndighet |
| [scb-kls-headcount](./scb-kls-headcount.md) | SCB KLS AM0102 — månadsanställda | api | anställda per myndighet (KLS-urval) |
| [scb-aku-arbetsloshet](./scb-aku-arbetsloshet.md) | SCB AKU AM0401 — arbetslöshetstal | api | nationell KPI-strip (/riksdag) |
| [scb-kpi-inflation](./scb-kpi-inflation.md) | SCB KPI PR0101 — inflationstakt | api | nationell KPI-strip (/riksdag) |
| [utskott-utgiftsomrade](./utskott-utgiftsomrade.md) | Utskottens utgiftsområden (riksdagsordningen) | seed | utgiftsområden per utskott (/committees) |

För att lägga till en ny källa: kopiera `_TEMPLATE.md`, fyll i, och kör
`npm run sync:data-sources` i `frontend/`. Se även CLAUDE.md → "Data
Source Discipline".
