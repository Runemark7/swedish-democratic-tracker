---
id: scb-aku-arbetsloshet
name: SCB AKU — Arbetslöshetstal (AM0401)
kind: api
upstream: https://www.statistikdatabasen.scb.se/pxweb/sv/ssd/START__AM__AM0401__AM0401A/AKURLBefK/
license: SCB öppna data — fri användning (CC0)
freshness: kvartalsvis
last_verified: 2026-06-10
verification_status: verified
verification_notes: ~
used_by:
  - /riksdag (nationell KPI-strip — Arbetslöshet)
---

## Vad det är

Arbetskraftsundersökningarna (AKU) är SCB:s officiella mätning av
arbetsmarknaden. Vi använder det säsongrensade relativa arbetslöshetstalet
för hela befolkningen 15–74 år — andelen av arbetskraften som är arbetslös —
som nationell indikator på riksdagssidan.

## Hur du själv kommer åt datan

Datan ligger i SCB:s statistikdatabas. Gå till statistikdatabasen.scb.se och
navigera till Arbetsmarknad → Arbetskraftsundersökningarna (AKU) →
Befolkningen efter arbetskraftstillhörighet → tabellen "Befolkningen 15-74 år
(AKU) efter arbetskraftstillhörighet, typ av data, kön och ålder. Kvartal".
Välj arbetskraftstillhörighet "arbetslöshetstal, procent", typ av data
"säsongrensad", kön "totalt", ålder "totalt 15-74 år" och önskade kvartal.
Samma tabell är åtkomlig maskinellt via SCB:s öppna PxWeb-API.

## Schema/fält vi använder

- `Arbetskraftstillh = ALÖSP` — relativt arbetslöshetstal i procent
- `TypData = SR_DATA` — säsongrensad serie
- `Kon = 1+2`, `Alder = tot15-74` — hela befolkningen 15–74 år
- `Tid` — kvartal (de två senaste hämtas; det äldre används för delta)

## Hur ofta vi hämtar

Dagligen via ingestionsschemaläggaren (national-kpis-arbetaren). SCB
publicerar nya kvartalsvärden cirka en månad efter kvartalets slut; vårt
värde uppdateras automatiskt när ett nytt kvartal publiceras.

## Kända begränsningar

- Säsongrensade serier kan revideras retroaktivt av SCB.
- Kvartalsupplösning — månadsvärden finns men har större osäkerhet.
