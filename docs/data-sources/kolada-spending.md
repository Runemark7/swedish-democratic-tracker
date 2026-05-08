---
id: kolada-spending
name: Kolada — kommunal verksamhetsutgift
kind: api
upstream: https://api.kolada.se/v3
license: Kommun- och regionkollade — fri användning
freshness: årlig
last_verified: 2026-05-08
used_by:
  - /kommun/:code (Budget-kortet)
---

## Vad det är
Samma upstream som [kolada](./kolada.md), men begränsad till de
KPI-id:n som beskriver kommunens verksamhetsutgifter per
ansvarsområde — skola, omsorg, plan, gata, etc. Separata MD eftersom
endpointen, mappningen, och visualiseringen skiljer sig från
KPI-stripen.

## Hur du själv kommer åt datan
```bash
# Hämta verksamhetsutgift per kommun
curl -s 'https://api.kolada.se/v3/data/kpi/N11004/municipality/0114'
```

KPI:n vi läser:
- `N11004` — Verksamhetens kostnad, kr/inv
- `N15028` — Pedagogisk verksamhet
- `N17014` — Vård och omsorg
- `N20014` — Infrastruktur
- `N30005` — Kultur och fritid
- `N07037` — Politisk verksamhet
- `N09022` — Affärsverksamhet
- `N05011` — Räddningstjänst
- `N45014` — Övrig

(Lista underhålls i CLAUDE.md → Architecture Diagram.)

## Schema/fält vi använder
- `data[].values[].value` — kr/invånare.
- `data[].period` — år.

## Begränsningar och kända problem
- Endast utvalda KPI:n stöds av Kolada för kommunal verksamhetsutgift.
  Övriga områden visar "saknas" i UI.
- Kommuner som ej rapporterar (t.ex. Bara) får helt tom budget — UI
  visar "Budgetdata saknas..."-meddelande.

## Hur vi bearbetar
Vår backend hämtar de utvalda KPI:erna för varje kommun från Kolada
och returnerar en lista. Frontend räknar om kr/invånare till totalt
belopp (kr × kommunens befolkning) och fördelar i en donut +
horisontella staplar på kommunens budgetkort.
