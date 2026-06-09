---
id: kolada
name: Kolada API v3
kind: api
upstream: https://api.kolada.se/v3
license: Kommun- och regionkollade — fri användning
freshness: årlig
last_verified: 2026-06-08
verification_status: verified
verification_notes: ""
used_by:
  - /region/:code (KPI strip)
  - /kommun/:code (KPI strip)
---

## Vad det är
Kolada är Rådet för främjande av kommunala analyser (RKA) sin databas
med jämförelsetal för svenska kommuner och regioner — ekonomi, vård,
skola, befolkning. Vi använder underlag för KPI-stripen på region- och
kommunsidor.

## Hur du själv kommer åt datan
API:et är publikt och kräver ingen API-nyckel. Du hittar det på
<https://api.kolada.se/v3> och den fullständiga API-referensen finns
på <https://github.com/Hypergene/kolada>.

För att hitta data om en specifik KPI besöker du API:ets rotdokument
och väljer endpoint `/data/kpi/{kpi_id}/municipality/{kommunkod}` — där
`{kpi_id}` är ett N-nummer (t.ex. `N60008`) och `{kommunkod}` är den
fyrsiffriga SCB-kommunkoden (t.ex. `0114` för Upplands Väsby). Metadata
om enskilda KPI:n (definition, enhet) hittar du via `/kpi/{kpi_id}`.

## Schema/fält vi använder
- `data[].values[].value` — siffran (typiskt %).
- `data[].period` — år.
- `data[].kpi` — KPI-id (N#####).
- `data[].municipality` — 4-siffrig kommunkod.

Vi läser ett urval av KPI:n per region och per kommun — t.ex. `N60008`
(resultat per skattekrona). Listan finns dokumenterad i CLAUDE.md i
sidans repo.

## Begränsningar och kända problem
- Kolada uppdateras typiskt en gång per år. Färska siffror kan saknas
  för innevarande år.
- Inte alla KPI:n finns för alla kommuner — t.ex. "Bara" (1229) saknar
  budgetdata helt.

## Hur vi bearbetar
Vår backend agerar proxy mot Kolada — vi cacher resultaten kortvarigt
för att inte överbelasta deras API, och mappar varje KPI till en
visningsklar form (siffra, enhet, mål, trendpil). Frontend hämtar
resultatet via vårt API och visar det på region- och kommunsidans
KPI-strip.
