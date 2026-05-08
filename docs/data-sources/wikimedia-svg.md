---
id: wikimedia-svg
name: Sverige-karta SVG (Wikimedia Commons)
kind: seed
upstream: https://commons.wikimedia.org/wiki/Category:Maps_of_municipalities_of_Sweden
license: CC BY-SA 2.5
freshness: engångs (statiska SVG-banor inkluderade i bundle)
last_verified: 2026-05-08
used_by:
  - /region (kartan)
  - /region/:code (kartan över kommunerna i regionen)
  - /kommun (kartan)
---

## Vad det är
SVG-paths för Sveriges 21 regioner och 290 kommuner, bundlade direkt i
frontend-koden. Ursprungligen från Wikimedia Commons, härlednings-
data ursprungligen SCB:s administrativa indelning.

## Hur du själv kommer åt datan
Källfilerna ligger i:
- `frontend/src/features/regions/data/sweden-regions-svg.ts`
- `frontend/src/features/municipalities/data/sweden-kommuner-svg.ts`

Original-SVG på Wikimedia: <https://commons.wikimedia.org/wiki/File:Sverigekarta-Landskap.svg>
(och tillhörande filer för kommun- och regionnivå).

För att uppdatera (t.ex. vid kommunsammanslagning):
1. Ladda hem ny SVG från Wikimedia.
2. Extrahera path-data per kommun/region (kod = `id`-attribut).
3. Generera `SWEDEN_KOMMUN_PATHS[]` / `SWEDEN_REGION_PATHS[]`.

## Schema/fält vi använder
- `code` — SCB-kod (4 siffror för kommun, 2 för region).
- `regionCode` — för kommuner: 2-prefix av `code`.
- `d` — SVG path-data.

## Begränsningar och kända problem
- CC BY-SA innebär att vi måste behålla attributionen någonstans synlig
  (sker via `<SectionSource>` på kart-kortet).
- Wikimedia-SVG:erna är inte alltid pixelperfekta; små glapp kan synas
  vid extrem zoom.

## Hur vi bearbetar
- Statiska data laddas direkt; ingen runtime-fetch.
- Komponenter `SwedenRegionMap` och `SwedenKommunMap` renderar SVG
  inline.
