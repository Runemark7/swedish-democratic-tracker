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
- `myndighet_namn` → `Authority.name`.
- `år` + `driftkostnad_mkr` → `Authority.history[].expenditureMdkr`.
- `år` + `antal_anstallda` → `Authority.headcountHistory[].headcountInt`.

## Begränsningar och kända problem
- Filformatet ändras ibland mellan år (kolumnnamn, separator). Verifiera
  schemat efter varje årlig uppdatering.
- Vissa myndigheter byter namn eller fusioneras — det måste hanteras
  manuellt i ingestion.

## Hur vi bearbetar
- ZIP:en laddas hem och läses i en
  `backend/internal/riksdag/seeder/` migration (en gång per år).
- Resultat skrivs till `authorities` och `authority_history` -tabellerna.
- Frontend läser via `useRiksdag().authorities`.
