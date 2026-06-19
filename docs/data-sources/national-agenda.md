---
id: national-agenda
name: Nationell politisk agenda (seed)
kind: seed
upstream: https://www.regeringen.se/rattsliga-dokument/
license: Public domain — Regeringskansliet
freshness: manuell; uppdateras när regeringen presenterar nya beslut
last_verified: 2026-06-19
verification_status: frozen
verification_notes: ~
used_by:
  - / (Kommande beslut-panel)
  - /riksdag (Agenda-kortet)
  - /agenda/:id (agendadetalj)
---

## Vad det är
En manuellt kurerad lista över den sittande regeringens uttalade
politiska prioriteringar på nationell nivå. Varje punkt är hämtad ur ett
namngivet regeringsdokument — Tidöavtalet, budgetpropositionen,
försvarspropositionen eller motsvarande — och bär sin egen primärkälla i
`source`-fältet. Det här är inte en härledning eller gissning: det är en
sammanställning av vad regeringen själv har skrivit, med källa.

## Hur du själv kommer åt datan
Källdokumenten är offentliga och publiceras av Regeringskansliet på
regeringen.se. Tidöavtalet finns under "Rättsliga dokument →
Överenskommelse"; budget- och försvarspropositioner under "Rättsliga
dokument → Proposition". Sök på dokumentets titel (t.ex. "Tidöavtalet"
eller "Budgetpropositionen 2024") för att läsa originaltexten och
verifiera varje punkt själv.

## Schema/fält vi använder
- `title` — kort rubrik för prioriteringen.
- `description` — beskrivning i klartext.
- `source` — det namngivna regeringsdokument punkten bygger på
  (t.ex. "Tidöavtalet 2022").
- `status` — `active`, `in_progress`, `completed` eller `cancelled`.

## Begränsningar och kända problem
- **Endast nationell nivå.** Det finns ingen motsvarande strukturerad
  källa för regioners eller kommuners årsplaner, så agendan visas inte på
  region- och kommunsidor.
- Listan uppdateras inte automatiskt — den speglar dokumenten vid senaste
  manuella genomgång (se `last_verified`), inte realtidsbeslut.
- Formuleringen i `title`/`description` är en förkortning av regeringens
  text; den fullständiga lydelsen finns alltid i den länkade primärkällan.

## Hur vi bearbetar
Vi läser regeringens egna policydokument, plockar ut konkreta åtgärder
och sammanfattar varje åtgärd i en rad med titel, beskrivning, status och
namngiven källa. Inga punkter härleds eller genereras — om en åtgärd inte
står i ett offentligt regeringsdokument tas den inte med.
