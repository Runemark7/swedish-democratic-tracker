---
id: region-budget-plan
name: Regioners mål och budget (seed)
kind: seed
upstream: ~
license: Public domain — respektive regions offentliga handlingar
freshness: manuell; länkarna pekar på regionens egen budgetsida
last_verified: 2026-06-20
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
beslutas av regionfullmäktige. Det finns ingen strukturerad, maskinläsbar
nationell källa för dessa planer. Vi gör två saker: vi återger regionens
egna övergripande mål ordagrant ur det beslutade dokumentet, och vi länkar
alltid till hela originalplanen. Vi sammanfattar eller tolkar inte —
formuleringarna är regionens egna och hela uppsättningen mål visas, aldrig
ett urval.

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
- `goalsLabel` — vad regionen själv kallar målen (t.ex. "Effektmål",
  "Fokusområden", "Strategiska mål").
- `period` — den planperiod målen är hämtade ur (t.ex. "2025–2027").
- `goals` — regionens egna övergripande mål, ordagrant och i sin helhet.

## Begränsningar och kända problem
- **Mål återges ordagrant, inte tolkat.** Vi visar regionens egna
  övergripande mål i sin helhet med länk till originalet — vi väljer inte ut
  eller formulerar om. En region utan reproducerbara mål visas med enbart länk.
- I nuläget har 13 av 21 regioner ordagranna mål inlagda. Övriga visas med
  enbart länk eftersom deras mål är nästlade per nämnd, uppdelade i en matris,
  eller bara finns i dokument som inte gick att läsa maskinellt.
- Vissa regioner saknar en stabil HTML-landningssida och länken pekar då på
  ett dokument (PDF) eller en sida för styrande dokument där den senaste
  budgeten finns.
- Länkarna kan flyttas när regionen byter webbplatsstruktur eller publicerar
  ett nytt budgetår; de behöver ses över med jämna mellanrum.
- Endast regioner (21 st) ingår. Kommuner (290 st) saknar i nuläget
  motsvarande länkar.

## Hur vi bearbetar
Ingen tolkning. Vi underhåller en handkurerad lista där varje regions
SCB-kod kopplas till namnet på dess budgetdokument, en länk till regionens
egen budgetsida, och regionens egna övergripande mål avskrivna ordagrant ur
det beslutade dokumentet. Vi sammanfattar inte och plockar inte ut delar —
hela måluppsättningen återges med regionens egna formuleringar, alltid med
länk till originalet.
