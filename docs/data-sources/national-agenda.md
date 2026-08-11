---
id: national-agenda
name: Regeringens dokument (katalog)
kind: seed
upstream: https://www.regeringen.se/rattsliga-dokument/
license: Regeringens dokument är offentliga; Tidöavtalet publiceras av partierna själva
freshness: manuell; uppdateras när ett nytt program- eller budgetdokument registreras
last_verified: 2026-08-11
verification_status: verified
verification_notes: ~
used_by:
  - /regering (katalogen över regeringens program- och budgetdokument)
---

## Vad det är
En katalog över regeringens program- och budgetdokument för mandatperioden.
Katalogen pekar på dokumenten i sin helhet. Den innehåller inga
sammanfattningar, inga citat och ingen bedömning av hur långt regeringen har
kommit med något.

Varje post består av titel, dokumentbeteckning, avsändare, publiceringsdatum
och en länk till dokumentet.

## Hur du själv kommer åt datan
Dokumenten publiceras av avsändaren själv. Katalogen lagrar en `url` till den
publiceringen, och länken går till hela dokumentet — inte till vår läsning av
det.

Kontrollerade 2026-08-11:

- Tidöavtalet – Överenskommelse för Sverige, 2022-10-14:
  <https://www.liberalerna.se/wp-content/uploads/tidoavtalet-overenskommelse-for-sverige-slutlig.pdf>
- Budgetpropositionen för 2026 (Prop. 2025/26:1), Finansdepartementet,
  2025-09-22:
  <https://www.regeringen.se/rattsliga-dokument/proposition/2025/09/2025261/>

Tidöavtalet är en överenskommelse mellan fyra partier och inte en statlig
handling, så det finns ingen neutral utgivare att länka till. Vi länkar den
version en av de undertecknande partierna publicerar, och säger det här.

## Schema/fält vi använder
- `title` — dokumentets titel som avsändaren skrivit den.
- `source` — dokumentbeteckning, t.ex. "Prop. 2025/26:1".
- `issuer` — vem som publicerat dokumentet: ett departement, eller parterna
  bakom en överenskommelse.
- `url` — länk till hela dokumentet.
- `published` — publiceringsdatum.
- `sort_order` — visningsordning.

## Begränsningar och kända problem
- **Katalogen är vår, inte regeringens.** Att ett dokument saknas betyder att vi
  inte har registrerat det — inte att det inte finns. Sidan skriver det.
- **Vi sammanfattar inte och vi betygsätter inte.** Tidigare innehöll den här
  källan fem handplockade punkter ur dokument som Tidöavtalet, var och en med en
  beskrivning som vi hade formulerat och en `status` (`active`,
  `in_progress`) som vi hade satt. Statusen var ett påstående om nuläget — hur
  regeringen låg till — utan källa, och den hade inte rörts sedan raderna lades
  in 2026-05-19. Att välja fem punkter ur hundratals är dessutom ett redaktionellt
  val. Både beskrivningarna och statusen är borttagna; se
  [#102](https://github.com/Runemark7/swedish-democratic-tracker/issues/102).
- **Urvalet av dokument är ett val vi gör**, och det är synligt: läsaren kan
  invända mot vilka dokument vi har registrerat. Det är en invändning som går att
  bemöta, till skillnad från en omskrivning läsaren inte kan jämföra med
  originalet.
- **Uppdatering kräver driftsättning.** Katalogen ligger i en seed-migration, så
  ett nytt dokument kräver en ny version av sajten.

## Hur vi bearbetar
Ingen bearbetning: dokumenten läses inte in, sammanfattas inte och tolkas inte.
Katalogen består av titel, beteckning, avsändare, datum och länk, kontrollerade
mot avsändarens egen publicering.
