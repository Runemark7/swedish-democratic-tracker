---
id: derived-agenda
name: Härledd agenda (per koalition)
kind: synthesized
upstream: ~
license: ~
freshness: deterministisk (rendrar samma resultat för samma input)
last_verified: 2026-05-08
used_by:
  - / (Kommande beslut-panel)
  - /riksdag (Agenda-kortet)
  - /region/:code (Agenda-kortet)
  - /kommun/:code (Agenda-kortet)
---

## Vad det är
Vi har ingen primärkälla i Sverige som publicerar regeringens /
regionstyrets / kommunstyrets faktiska årsplan i strukturerad form. Vi
**genererar** därför en agenda från koalitionens partisammansättning.

Detta är uppmärkt i UI med `≈`-glyfen så det går att se på en gång att
det inte är ett officiellt dokument utan en härledning.

## Hur du själv kommer åt datan
Det finns ingen att hämta. Datan skapas i webbläsaren från
mandatfördelningen, alltid med samma resultat för samma kommun eller
region. Reglerna är dokumenterade nedan i avsnittet "Hur vi bearbetar".

## Schema/fält vi använder
- `title` — kort rubrik.
- `description` — beskrivning.
- `source` — fast etikett (t.ex. "Mandatprogrammet M+KD+L 2022–2026").
- `status` — `active`, `in_progress`, eller `completed`.

## Begränsningar och kända problem
- **Detta är inte officiell data.** Härledningen baseras på vänster-
  vs högerblockslogik och är best-effort.
- För regionkoalitioner som blandar block (t.ex. M+MP "blågrön" i
  Stockholm) kan agendan se inkonsekvent ut.
- Innehållet ändras inte över tid — det är en fast lista per block, inte
  en spegling av faktiska politiska beslut.

## Hur vi bearbetar
1. Vi tittar på koalitionens första parti och bedömer om koalitionen
   är vänster- eller högerlutande.
2. Beroende på lutning + nivå (region eller kommun) plockas en av fyra
   förskrivna listor.
3. Fyra punkter från listan väljs deterministiskt baserat på regionens
   eller kommunens kod, så samma plats alltid får samma agenda.
