---
id: wikimedia-svg
name: Sverige-karta SVG (Wikimedia Commons)
kind: seed
upstream: https://commons.wikimedia.org/wiki/Category:Maps_of_municipalities_of_Sweden
license: CC BY-SA 2.5
freshness: engångs (statiska SVG-banor inkluderade i bundle)
last_verified: 2026-06-08
verification_status: frozen
verification_notes: "Statisk SVG; CC BY-SA 2.5-attribution renderas nu i kartvyerna."
used_by:
  - /region (kartan)
  - /region/:code (kartan över kommunerna i regionen)
  - /kommun (kartan)
---

## Vad det är
SVG-paths för Sveriges 21 regioner och 290 kommuner. De är inbyggda
direkt i sidan så kartan kan ritas omedelbart utan extra hämtning.
Ursprunget är Wikimedia Commons, baserat på SCB:s administrativa
indelning.

## Hur du själv kommer åt datan
Originalfilerna ligger på Wikimedia Commons och är fria att ladda hem
och använda under CC BY-SA 2.5:

- Region/landskapskarta: <https://commons.wikimedia.org/wiki/File:Sverigekarta-Landskap.svg>
- Kommun- och länskartor (kategori): <https://commons.wikimedia.org/wiki/Category:Maps_of_municipalities_of_Sweden>

Varje kommun/region är en separat `<path>` i SVG-filen, identifierad
med kommunens eller regionens SCB-kod som `id`-attribut.

## Schema/fält vi använder
- `code` — SCB-kod (4 siffror för kommun, 2 för region).
- `regionCode` — för kommuner: de två första siffrorna av kommunkoden.
- `d` — SVG path-data (geometrin).

## Begränsningar och kända problem
- CC BY-SA innebär att attributionen måste synas — den finns i
  käll-listan ("Källor: ...") längst ner på kart-kortet.
- SVG-banorna är inte pixelperfekta; små glapp kan synas vid extrem
  zoom.
- När en kommun byter form (sammanslagning, gränsjustering) måste
  banorna uppdateras manuellt från Wikimedia.

## Hur vi bearbetar
Banorna packas in i sidan vid bygge och ritas direkt i webbläsaren —
ingen runtime-hämtning. Klick på en kommun eller region navigerar
till motsvarande detaljsida.
