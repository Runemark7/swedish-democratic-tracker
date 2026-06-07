---
id: scb-kls-headcount
name: SCB KLS — Månadsanställda i statlig sektor (AM0102)
kind: api
upstream: https://www.scb.se/hitta-statistik/statistik-efter-amne/arbetsmarknad/loner-och-arbetskostnader/konjunkturstatistik-loner-for-statlig-sektor-kls/
license: SCB öppna data — fri användning (CC0)
freshness: månadsvis (december-snapshot används som årsavstämning)
last_verified: 2026-06-06
verification_status: verified
verification_notes: ~
used_by:
  - /riksdag/myndigheter (anställda per myndighet — lista och headline)
  - /riksdag/myndigheter/:slug (anställdahistorik — stapeldiagram per år)
---

## Vad det är

SCB:s Konjunkturstatistik löner, sysselsättning och arbetstid (KLS) är en
månatlig enkät till statliga arbetsgivare om antal anställda. Vi använder
Tabell 14 — *Månadsanställda i statlig sektor per myndighet* — ur databasdelen
AM0102A. Varje post avser en specifik myndighet (kodad med KLS-myndighetskod)
och ett specifikt månadsslut. Vi tar december-snapshoten för varje år och
använder dem som årsavstämningar för att bygga personalhistorik.

KLS täcker ett begränsat antal stora myndigheter (polisen, domstolsväsendet,
Försäkringskassan, Skatteverket m.fl.) — inte samtliga ~449 i myndighetsregistret.

## Hur du själv kommer åt datan

SCB:s PxWeb-databas är öppen och kräver ingen registrering. Startpunkten för
statistikområdet är landningssidan ovan. Därifrån navigerar du:

*Arbetsmarknad → Konjunkturstatistik löner, sysselsättning och arbetstid
(KLS) → AM0102A → KLStabell14LpMan*

Via PxWeb kan du välja myndigheter (KLS-koder), tidsperioder (format ÅÅÅÅMmm)
och ladda ned som CSV, JSON eller via API. API-endpointen accepterar POST med
ett urval av variabler och returnerar JSON.

## Schema/fält vi använder

- `KLSkod` (variabel: `MyNamn`) — KLS-myndighetskod, t.ex. `C021` för
  Polismyndigheten. Koden är inte densamma som org-numret.
- `Tid` — period i formatet `ÅÅÅÅMmm`, t.ex. `2024M12`.
- `value` — antal månadsanställda (heltids- och deltidsanställda, ej
  heltidsekvivalenter).

## Begränsningar och kända problem

- KLS täcker bara ett urval myndigheter med egna KLS-koder. Myndigheter som
  saknar kod visas med "–" i personalkolumnen.
- Koden `C021` (polisväsendet) inkluderar både Polismyndigheten och
  Säkerhetspolisen — dessa kan inte särskiljas i KLS-data.
- KLS mäter *antal anställda*, inte årsarbetskrafter. Värdet är därmed inte
  direkt jämförbart med Statskontorets årsarbetskrafter, som är
  heltidsekvivalenter.
- Vi tar december som årsavstämning. Om en myndighet hade atypisk bemanning
  just i december syns det i vår tidsserie.

## Hur vi bearbetar

Datan hämtas via SCB:s PxWeb-API med ett POST-anrop per körning. Vi begär
december-värden för ett intervall av år och matchar KLS-koder mot
myndigheternas kanoniska namn via en intern kodtabell. Resultatet cachelagras
i 24 timmar och används för att berika personnelsiffran och stapeldiagrammet
på varje myndighetsprofil i realtid vid varje API-anrop från frontend.
