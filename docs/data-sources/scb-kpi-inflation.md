---
id: scb-kpi-inflation
name: SCB KPI — Inflationstakt, årsförändring (PR0101)
kind: api
upstream: https://www.statistikdatabasen.scb.se/pxweb/sv/ssd/START__PR__PR0101__PR0101A/KPI2020M/
license: SCB öppna data — fri användning (CC0)
freshness: månadsvis
last_verified: 2026-06-10
verification_status: verified
verification_notes: ~
used_by:
  - /riksdag (nationell KPI-strip — Inflation (KPI))
---

## Vad det är

Konsumentprisindex (KPI) är SCB:s officiella mått på prisutvecklingen för
hushållens konsumtion. Vi använder seriens publicerade *årsförändring* — hur
mycket priserna ändrats de senaste tolv månaderna, dvs. inflationstakten —
som nationell indikator på riksdagssidan. Riksbankens inflationsmål är 2 %.

## Hur du själv kommer åt datan

Datan ligger i SCB:s statistikdatabas. Gå till statistikdatabasen.scb.se och
navigera till Priser och konsumtion → Konsumentprisindex (KPI) → tabellen
"Konsumentprisindex (KPI), totalt, 2020=100. Månad". Välj tabellinnehåll
"Årsförändring" och önskade månader. Samma tabell är åtkomlig maskinellt via
SCB:s öppna PxWeb-API.

## Schema/fält vi använder

- `ContentsCode = 00000804` — Årsförändring, procent (12-månaders
  inflationstakt, publicerad av SCB — vi räknar inte själva)
- `Tid` — månad (de två senaste hämtas; den äldre används för delta)

## Hur ofta vi hämtar

Dagligen via ingestionsschemaläggaren (national-kpis-arbetaren). SCB
publicerar KPI för en månad cirka två veckor in i nästföljande månad; vårt
värde uppdateras automatiskt vid nästa körning.

## Kända begränsningar

- Senaste månadens tal kan vara preliminärt tills det fastställs.
- Årsförändringen är KPI (inte KPIF); Riksbanken styr formellt mot KPIF.
