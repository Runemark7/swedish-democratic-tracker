---
id: seed-budget-data
name: Statsbudget (seed)
kind: seed
upstream: https://www.regeringen.se/rattsliga-dokument/proposition/
license: Public domain — Regeringskansliet
freshness: engångs per år; uppdateras vid ny budget
last_verified: 2026-05-08
used_by:
  - /budget (budgetöversikt)
  - /budget/areas/:code (UO-historik)
  - /riksdag (Budget-kortet)
---

## Vad det är
Hand-kuraterad data från statens budgetproposition (BP) över flera år.
Innehåller anslagna belopp per utgiftsområde (UO 1–27) samt vilken
status posten har — föreslagen av regeringen eller beslutad av
riksdagen.

## Hur du själv kommer åt datan
Budgetpropositionen publiceras årligen i september på Regeringskansliets
webbplats. Senaste versioner:

- 2026/27:1 — <https://www.regeringen.se/rattsliga-dokument/proposition/2025/09/>
- 2025/26:1 — <https://www.regeringen.se/rattsliga-dokument/proposition/2024/09/>

PDF:erna har en sammanfattningstabell i kapitel 1 där summorna per UO
listas. Vi har plockat ut dessa siffror manuellt och fört in dem i vår
databas, med citation till exakt sida i propositionen som källa.

## Schema/fält vi använder
- `year` — budgetår.
- `status` — `decided` (riksdagen har antagit) eller `proposed` (endast
  regeringens förslag, inte slutligt beslut).
- `area_code` — utgiftsområde (1–27).
- `amount_ksek` — anslag i tusen kronor.

## Begränsningar och kända problem
- Datan är en ögonblicksbild av propositionen vid antagandet. Senare
  ändringsbudgetar fångas inte in automatiskt.
- "Beslutad budget" syftar på riksdagens slutligt antagna ramar, inte
  regeringens första proposition. Skillnaden kan vara stor när
  riksdagen avviker från regeringens förslag.
- Upptäckta fel i historiska siffror rättas via en ny korrigeringspost,
  inte genom att skriva över den gamla — så historiken bevaras.

## Hur vi bearbetar
Manuell extrahering från PDF → införd i vår databas. Frontend hämtar
datan via det publika API:et `/api/budget/years` och tillhörande
endpoints, och visar den på Budget-sidan och Riksdag-sidans
budgetkort.
