---
id: scb-kostndrlt
name: SCB KostnDRLT (regionernas driftkostnader)
kind: api
upstream: https://api.scb.se/OV0104/v1/doris/sv/ssd/OE/OE0107/OE0107D/KostnDRLT
license: PxWeb open data — fri användning
freshness: årlig
last_verified: 2026-05-08
used_by:
  - /region/:code (Budget-kortet)
---

## Vad det är
SCB:s tabell över regioners nettokostnader per verksamhetsområde
(hälso- och sjukvård, kollektivtrafik, regional utveckling).

## Hur du själv kommer åt datan
```bash
curl -s -X POST \
  'https://api.scb.se/OV0104/v1/doris/sv/ssd/OE/OE0107/OE0107D/KostnDRLT' \
  -H 'Content-Type: application/json' \
  -d '{"query":[{"code":"Region","selection":{"filter":"item","values":["22L"]}},{"code":"Verksomrkom","selection":{"filter":"item","values":["1","2","3","5","6","7","8","4","0-9"]}},{"code":"ContentsCode","selection":{"filter":"item","values":["000000A7"]}},{"code":"Tid","selection":{"filter":"item","values":["2024"]}}],"response":{"format":"json"}}'
```

## Schema/fält vi använder
- `data[].key[0]` — regionkod (`XXL`).
- `data[].key[1]` — verksamhetsområde (`1` Hälso-/sjukvård, `2`
  Tandvård, `3` Övrig vård, ..., `0-9` total).
- `data[].values[0]` — mnkr (miljoner kronor).

## Begränsningar och kända problem
- **Gotland (region 09)**: SCB använder `0980L` här (kommun-style) —
  inte `09L`. `client.go:45` har en hårdkodad specialregel för det.
- Aktuella års data är ofta `..` (preliminär). Backend-handlern
  `getRegionBudget` faller tillbaka från år-1 till år-5 för att hitta
  publicerade värden.

## Hur vi bearbetar
- `backend/internal/regions/adapters/scb/client.go:44`
  (`FetchRegionBudget`) hämtar och mappar verksamhetsområden till
  regionsbudgetnamn.
- Handler-fallback i
  `backend/internal/regions/adapters/http/handler.go:77`.
