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
**genererar** därför en agenda från koalitionens partisammansättning
genom en hårdkodad lookup.

Detta är uppmärkt i UI med `≈`-glyfen så användaren inte tror att det
är ett officiellt dokument.

## Hur du själv kommer åt datan
Logiken finns i `frontend/src/hooks/useDemocracy.ts → generateAgenda()`.

```bash
# Inspect generation rules
grep -A50 "function generateAgenda" frontend/src/hooks/useDemocracy.ts
```

## Schema/fält vi använder
Returnerar `AgendaItem[]`:
- `title` — kort rubrik.
- `description` — beskrivning.
- `source` — fast string (t.ex. "Mandatprogrammet M+KD+L 2022–2026").
- `status` — `active` | `in_progress` | `completed`.

## Begränsningar och kända problem
- **Detta är inte officiell data.** Härledningen baseras på vänster-
  vs högerblockslogik och är best-effort.
- För regionkoalitioner som blandar block (t.ex. M+MP "blågrön" i
  Stockholm) kan agendan se inkonsekvent ut.
- Rotation av items mellan sidladdningar är deterministisk: samma kod
  ger samma agenda.

## Hur vi bearbetar
Funktionen `generateAgenda(level, governingParties, code)`:
1. Bestämmer "lean" (left/right) från första partiet i `governingParties`.
2. Plockar lista från `REGION_LEFT` / `REGION_RIGHT` /
   `KOMMUN_LEFT` / `KOMMUN_RIGHT`.
3. Returnerar 4 items (deterministiskt urval baserat på `code`-hash).
