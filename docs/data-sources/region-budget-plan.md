---
id: region-budget-plan
name: Regioners mål och budget (seed)
kind: seed
upstream: ~
license: Public domain — respektive regions offentliga handlingar
freshness: manuell; länkarna pekar på regionens egen budgetsida
last_verified: 2026-06-19
verification_status: frozen
verification_notes: ~
used_by:
  - /region/:code (Regionens plan-kortet)
---

## Vad det är
Varje region är enligt kommunallagen (11 kap.) skyldig att varje år anta en
budget som innehåller mål och riktlinjer för verksamheten — regionens
faktiska årsplan. Dokumentet heter olika i olika regioner ("Mål och
budget", "Regionplan och budget", "Verksamhetsplan med budget" osv.) och
beslutas av regionfullmäktige. Vi länkar regionens egen officiella
budgetsida i stället för att generera eller sammanfatta innehållet — det
finns ingen strukturerad, maskinläsbar nationell källa för dessa planer.

## Hur du själv kommer åt datan
Varje region publicerar sitt budgetdokument på sin egen webbplats. Gå till
regionens sida (t.ex. regionstockholm.se, vgregion.se, skane.se) och sök på
"budget", "regionplan" eller "styrande dokument". Varje länk i tjänsten går
direkt till den sidan eller dokumentet hos regionen. Länkarna verifierades
manuellt 2026-06-19 så att de pekar på regionens egen budget-/styrdokumentsida.

## Schema/fält vi använder
- `label` — det namn regionen själv ger dokumentet (t.ex. "Regionplan och
  budget").
- `url` — officiell sida eller PDF hos regionens egen domän.

## Begränsningar och kända problem
- **Endast länk, inte extraherad data.** Vi visar inte enskilda punkter ur
  planen — bara en länk till primärkällan. Läsaren bedömer själv.
- Vissa regioner saknar en stabil HTML-landningssida och länken pekar då på
  ett dokument (PDF) eller en sida för styrande dokument där den senaste
  budgeten finns.
- Länkarna kan flyttas när regionen byter webbplatsstruktur eller publicerar
  ett nytt budgetår; de behöver ses över med jämna mellanrum.
- Endast regioner (21 st) ingår. Kommuner (290 st) saknar i nuläget
  motsvarande länkar.

## Hur vi bearbetar
Ingen bearbetning. Vi underhåller en handkurerad lista där varje regions
SCB-kod kopplas till namnet på dess budgetdokument och en länk till
regionens egen budgetsida. Inget innehåll härleds eller tolkas.
