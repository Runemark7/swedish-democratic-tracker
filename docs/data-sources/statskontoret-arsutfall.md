---
id: statskontoret-arsutfall
name: Statskontoret — Årsutfall för myndigheter
kind: csv
upstream: https://www.statskontoret.se/analys-och-statistik/oppna-data/arsutfall/
license: Statskontoret open data — fri användning
freshness: årlig
last_verified: 2026-06-06
verification_status: unsure
verification_notes: "Det dokumenterade upstream-URL:et /psidata/arsutfall returnerar 200 men den faktiska ZIP-filen hämtas via en dynamisk GetFile-länk som ändrats sedan MD skrevs; tidigare bash-block med curl/unzip borttaget. Dessutom saknar AuthorityDetailPage och MyndigheterListPage SourceMarker på kostnads- och personalvärden (se audit Appendix 1)."
used_by:
  - /riksdag (Myndigheter-kortet — donut, lista, dual-line)
  - /riksdag/myndigheter/:slug (myndighetsdetalj)
---

## Vad det är
Statskontoret publicerar årligen statliga myndigheters faktiska
driftkostnader (löner, lokaler, IT — exkl. transfereringar) som en
ZIP med CSV-filer. Detta är "kvittot" för Myndigheter-kortet på
Riksdag-sidan.

## Hur du själv kommer åt datan
Datan publiceras som öppen data på Statskontorets webbplats. Landningssidan
är <https://www.statskontoret.se/psidata/arsutfall> — där finns en länk
till den senaste ZIP-filen samt arkiv för tidigare år. Filnamn och
exakta URL:er till ZIP:en varierar per år och uppdateras på landningssidan
när ny data publiceras.

Inne i ZIP:en finns semikolonseparerade CSV-filer (svensk konvention).
Den primära filen innehåller ett anslag per rad med kolumner för
myndighetens anslag, år, statsbudgetramen och det faktiska utfallet.

## Schema/fält vi använder
Vi läser följande kolumner ur CSV:n (kolumnpositioner indexerade från 1):
- Kolumn 2 — Anslag (myndighets-/anslagsnamn).
- Kolumn 4 — År.
- Kolumn 6 — Statens budget (mkr).
- Kolumn 7 — Ändringsbudgetar (mkr).
- Kolumn 10 — Utfall (tkr).

## Begränsningar och kända problem
- Filformatet ändras ibland mellan år (kolumnnamn, separator). Vi
  verifierar schemat efter varje årlig uppdatering.
- Den dynamiska URL:en till ZIP-filen ändrar sig per år — kontrollera
  alltid landningssidan för aktuell nedladdningslänk.
- Vissa myndigheter byter namn eller slås ihop över tid — vi
  konsoliderar dem manuellt så historiken hålls ihop.

## Hur vi bearbetar
ZIP:en laddas hem en gång per år, CSV:erna parseras och skrivs in i
vår databas. Frontend hämtar listan via vårt API och visar
Myndigheter-kortet på Riksdag-sidan med donut, lista och en linjegraf
för kostnad + anställda över tid.
