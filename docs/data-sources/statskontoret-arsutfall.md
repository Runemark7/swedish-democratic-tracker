---
id: statskontoret-arsutfall
name: Statskontoret — Årsutfall för myndigheter
kind: csv
upstream: https://www.statskontoret.se/psidata/arsutfall
license: Statskontoret open data — fri användning
freshness: årlig
last_verified: 2026-05-08
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
Datan är inte en API utan en zippad CSV-katalog. Direktlänk:

```bash
# Ladda hem zip:en (URL kan ändras per år; kolla på upstream-sidan)
curl -O 'https://www.statskontoret.se/psidata/arsutfall/arsutfall-2024.zip'
unzip arsutfall-2024.zip -d arsutfall-2024/
ls arsutfall-2024/
```

CSV-filerna inom zip:en är semikolonseparerade (svensk konvention).
Den primära filen är `myndigheter.csv` med kolumnerna:

| Kolumn | Beskrivning |
|---|---|
| `myndighet_namn` | T.ex. "Polismyndigheten" |
| `org_nummer` | Organisationsnummer |
| `år` | Aktuellt år |
| `driftkostnad_mkr` | Driftkostnader exkl. transfereringar, miljoner kr |
| `antal_anstallda` | Helårsekvivalenter |

## Schema/fält vi använder
- `myndighet_namn` — myndighetens namn.
- `år` + `driftkostnad_mkr` — driftkostnad per år (i miljarder kr i UI).
- `år` + `antal_anstallda` — antal anställda per år.

## Begränsningar och kända problem
- Filformatet ändras ibland mellan år (kolumnnamn, separator). Vi
  verifierar schemat efter varje årlig uppdatering.
- Vissa myndigheter byter namn eller slås ihop över tid — vi
  konsoliderar dem manuellt så historiken hålls ihop.

## Hur vi bearbetar
ZIP:en laddas hem en gång per år, CSV:erna parseras och skrivs in i
vår databas. Frontend hämtar listan via vårt API och visar
Myndigheter-kortet på Riksdag-sidan med donut, lista och en linjegraf
för kostnad + anställda över tid.
