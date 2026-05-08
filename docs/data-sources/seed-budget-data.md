---
id: seed-budget-data
name: Statsbudget (seed)
kind: seed
upstream: https://www.regeringen.se (budgetpropositionerna)
license: Public domain — Regeringskansliet
freshness: engångs per år; uppdateras vid ny budget
last_verified: 2026-05-08
used_by:
  - /budget (budgetöversikt)
  - /budget/areas/:code (UO-historik)
  - /riksdag (Budget-kortet)
---

## Vad det är
Hand-kuraterad data från budgetpropositionen för flera år, sparad i
`backend/migrations/000005`–`000009`. Innehåller belopp per
utgiftsområde (UO 1–27).

## Hur du själv kommer åt datan
Budgetpropositionen publiceras årligen i september på regeringens
webbplats. Senaste versioner:

- 2026/27:1 — `https://www.regeringen.se/rattsliga-dokument/proposition/2025/09/`
- 2025/26:1 — `https://www.regeringen.se/rattsliga-dokument/proposition/2024/09/`

PDF:erna har en sammanfattningstabell i kapitel 1 där summorna per UO
listas. De citaten finns som SQL-kommentarer i varje migration.

## Schema/fält vi använder
Tabellerna `budget_years` och `budget_allocations`:
- `budget_years.year` — budgetår.
- `budget_years.status` — `decided` (passerat riksdagen) eller `proposed`.
- `budget_allocations.area_code` — UO-kod (1–27).
- `budget_allocations.amount_ksek` — belopp i tusen kronor.

## Begränsningar och kända problem
- En migration per år innebär att upptäckta fel i historiska siffror
  kräver en korrigerande migration, inte en redigering av den gamla.
- "Beslutad budget" syftar på riksdagens slutligt antagna ramar, inte
  regeringens första proposition.

## Hur vi bearbetar
- Migrationer 000005–000009 inserter rader.
- Frontend läser via
  `frontend/src/features/riksdag/api.ts → budgetApi.listYears`
  / `getYear(year)`.
