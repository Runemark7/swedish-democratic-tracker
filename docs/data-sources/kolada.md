---
id: kolada
name: Kolada API v3
kind: api
upstream: https://api.kolada.se/v3
license: Kommun- och regionkollade — fri användning
freshness: årlig
last_verified: 2026-05-08
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
Publik API utan nyckel:

```bash
# Lista en KPI över tid för en kommun
curl -s 'https://api.kolada.se/v3/data/kpi/N00945/municipality/0114'

# Hämta metadata om en KPI (definition + enhet)
curl -s 'https://api.kolada.se/v3/kpi/N00945'
```

API-referens: <https://github.com/Hypergene/kolada>.

## Schema/fält vi använder
- `data[].values[].value` — siffran (typiskt %).
- `data[].period` — år.
- `data[].kpi` — KPI-id (N#####).
- `data[].municipality` — 4-siffrig kommunkod.

KPI-id:n vi läser idag finns i
`backend/internal/regions/adapters/kolada/` (region) och `municipalities`
sub-paketet (kommun). Se exempelvis `N60008` (resultat per skattekrona).

## Begränsningar och kända problem
- Kolada uppdateras typiskt en gång per år. Färska siffror kan saknas
  för innevarande år.
- Inte alla KPI:n finns för alla kommuner — t.ex. "Bara" (1229) saknar
  budgetdata helt.

## Hur vi bearbetar
- Backend-handler i
  `backend/internal/regions/adapters/http/handler.go` proxar Kolada-svar
  via `KoladaClient` (backend/internal/regions/adapters/kolada/).
- Frontend-hook
  `useRegion(code)` / `useKommun(code)` i
  `frontend/src/hooks/useDemocracy.ts` kallar
  `regionsApi.getKPIs` / `municipalitiesApi.getMunicipalityKPIs` och
  mappar via `kpiItemsToStrip`.
