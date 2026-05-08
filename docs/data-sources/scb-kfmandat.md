---
id: scb-kfmandat
name: SCB Kfmandat (kommunala mandat)
kind: api
upstream: https://api.scb.se/OV0104/v1/doris/sv/ssd/ME/ME0104/ME0104A/Kfmandat
license: PxWeb open data — fri användning
freshness: 4-årscykel (val)
last_verified: 2026-05-08
used_by:
  - /kommun/:code (mandat, hemicycle)
  - /region/:code (kommun-karta — vilka koder är klickbara)
---

## Vad det är
SCB:s tabell över mandatfördelning i kommunfullmäktige efter senaste
val. Källa till alla mandatkort på kommunsidor.

## Hur du själv kommer åt datan
Tabellen är POST-only:

```bash
curl -s -X POST \
  'https://api.scb.se/OV0104/v1/doris/sv/ssd/ME/ME0104/ME0104A/Kfmandat' \
  -H 'Content-Type: application/json' \
  -d '{"query":[{"code":"Region","selection":{"filter":"item","values":["1490"]}},{"code":"Parti","selection":{"filter":"item","values":["M","C","FP","KD","MP","S","V","SD","ÖVRIGA"]}},{"code":"ContentsCode","selection":{"filter":"item","values":["ME0104C1"]}},{"code":"Tid","selection":{"filter":"item","values":["2022"]}}],"response":{"format":"json"}}'
```

Metadata (vilka kommuner finns, vilka år, vilka partier):

```bash
curl -s 'https://api.scb.se/OV0104/v1/doris/sv/ssd/ME/ME0104/ME0104A/Kfmandat'
```

## Schema/fält vi använder
- `data[].key[0]` — kommunkod (4 siffror).
- `data[].key[1]` — partikod (M, C, FP, KD, MP, S, V, SD, ÖVRIGA).
- `data[].key[2]` — år ("2022").
- `data[].values[0]` — antal mandat.

## Begränsningar och kända problem
- SCB använder fortfarande den gamla partikoden `FP` i datan; vi
  visar den som `L` (Liberalerna) på sidan.
- Värden `..` betyder att SCB inte publicerat siffran än och hoppas
  över i vårt UI.

## Hur vi bearbetar
Vår backend hämtar mandaten från SCB och lagrar dem per kommun.
Frontend läser därefter dessa och renderar en hemicycle (halvrunds-
diagram), partilegend, samt total mandatsumma på varje kommunsida.
Hämtningen sker idempotent vid driftsättning så data fylls på vid
behov utan att skriva över korrekta rader.
