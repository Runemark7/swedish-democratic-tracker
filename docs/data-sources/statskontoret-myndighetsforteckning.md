---
id: statskontoret-myndighetsforteckning
name: Statskontoret — Myndighetsförteckning (årsarbetskrafter)
kind: csv
upstream: https://www.statskontoret.se/analys-och-statistik/oppna-data/myndighetsforteckning/
license: Statskontoret open data — fri användning
freshness: årlig (ny version publiceras vanligtvis under Q1 nästkommande år)
last_verified: 2026-06-06
verification_status: verified
verification_notes: >
  Den exakta XLSX-URL:en innehåller ett hårdkodat filnamn med årstal
  (statskontorets-myndighetsforteckning-2025.xlsx). URL:en måste uppdateras
  när Statskontoret publicerar nästa utgåva.
used_by:
  - /riksdag/myndigheter (Totalt anställda — övergripande stat)
  - /riksdag/myndigheter/:slug (anställda — headline-siffra och historik per år)
---

## Vad det är

Statskontoret publicerar varje år en Myndighetsförteckning — ett XLSX-dokument
med longitudinella personaluppgifter för samtliga statliga myndigheter.
Centrala måttet är *årsarbetskrafter*: heltidsekvivalenter per myndighet och
år, med täckning från 2007 och framåt. Statskontoret beräknar detta mått från
arbetsgivardeklarationer till Skatteverket och det är det officiella måttet
för personalstyrka i statlig sektor.

Det är denna källa som driver kolumnen "Anställda" och historikstaplarna på
myndighetssidorna.

## Hur du själv kommer åt datan

Statskontorets öppna data finns samlade på
<https://www.statskontoret.se/om-statskontoret/om-webbplatsen/oppna-data/>.
Där hittar du Myndighetsförteckningen listad som en nedladdningsbar XLSX-fil.
Filnamnet inkluderar publiceringsåret (t.ex. `statskontorets-myndighetsforteckning-2025.xlsx`).

Filen innehåller ett flertal flikar. Den relevanta är
*Förteckning 2007–2025* (årtalsintervallet i fliknamnet uppdateras varje år)
med ett rad per myndighet och år i ett long-format.

## Schema/fält vi använder

- `orgnr` — organisationsnummer, används som nyckel för att matcha mot SCB:s
  myndighetsregister.
- `myndighet` — myndighetens namn.
- `departement` — ansvarigt departement.
- `år` — kalenderår.
- `årsarbetskrafter` — heltidsekvivalenter det aktuella året.

## Begränsningar och kända problem

- URL:en innehåller ett hårdkodat årstal och måste uppdateras manuellt varje
  år när Statskontoret publicerar ny version.
- Siffran är årsarbetskrafter (heltidsekvivalenter), inte faktiska
  anställningsavtal. Deltidsanställda räknas som bråkdelar.
- Myndigheter som slagits ihop eller delats kan ha diskontinuerliga
  tidsserier — historiken kan brytas av namnbyten eller omorganiseringar.
- Datan gäller statlig sektor i vid mening; bolag som ägs av staten
  ingår inte.

## Hur vi bearbetar

En ingestionsjobb hämtar XLSX-filen en gång per dygn, läser bladet med
longitudinell data och skriver in årsarbetskrafter per myndighet i vår
databas matchat på org-nummer. Det senaste årets värde visas som
headline-siffra; samtliga år visas som stapeldiagram i historik-vyn på
myndighetsprofilen.
