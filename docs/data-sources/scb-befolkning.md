---
id: scb-befolkning
name: SCB BefolkningNy
kind: api
upstream: https://api.scb.se/OV0104/v1/doris/sv/ssd/BE/BE0101/BE0101A/BefolkningNy
license: PxWeb open data — fri användning
freshness: årlig (kvartal för kommunnivå)
last_verified: 2026-05-08
used_by:
  - /region landing (population per region)
  - /kommun landing (population per kommun)
  - /kommun/:code (population sub-header)
  - /region/:code (population sub-header)
---

## Vad det är
SCB:s folkmängdsdatabas — total befolkning per kommun.

## Hur du själv kommer åt datan
```bash
curl -s -X POST \
  'https://api.scb.se/OV0104/v1/doris/sv/ssd/BE/BE0101/BE0101A/BefolkningNy' \
  -H 'Content-Type: application/json' \
  -d '{"query":[{"code":"Region","selection":{"filter":"item","values":["0114"]}},{"code":"ContentsCode","selection":{"filter":"item","values":["BE0101N1"]}},{"code":"Tid","selection":{"filter":"item","values":["2023"]}}],"response":{"format":"json"}}'
```

## Schema/fält vi använder
- `data[].key[0]` — kommunkod.
- `data[].values[0]` — antal invånare.

## Begränsningar och kända problem
- Endast kommunnivå publiceras direkt; regionsumma räknas i seeder
  genom att summera ingående kommuner.
- SCB tillåter ~50 koder per query. Seeder batcher i grupper om 50.

## Hur vi bearbetar
- `backend/internal/regions/seeder/seeder.go:288` (`fetchPopulations`)
  batchar och hämtar.
- Lagras i `municipalities.population`.
