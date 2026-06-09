---
id: scb-kostndrlt
name: SCB KostnDRLT (regionernas driftkostnader)
kind: api
upstream: https://api.scb.se/OV0104/v1/doris/sv/ssd/OE/OE0107/OE0107D/KostnDRLT
license: PxWeb open data — fri användning
freshness: årlig
last_verified: 2026-06-08
verification_status: verified
verification_notes: ""
used_by:
  - /region/:code (Budget-kortet)
---

## Vad det är
SCB:s tabell över regioners nettokostnader per verksamhetsområde
(hälso- och sjukvård, kollektivtrafik, regional utveckling).

## Hur du själv kommer åt datan
Tabellen heter **KostnDRLT** och finns i SCB:s statistikdatabas under
offentlig ekonomi. Du når den via SCBs webbgränssnitt på
<https://www.statistikdatabasen.scb.se/> — sök på "KostnDRLT" eller
navigera via Offentlig ekonomi → Landstingsekonomi → Driftkostnader.

Teknisk åtkomst sker via SCB:s PxWeb API med sökvägen
`/OV0104/v1/doris/sv/ssd/OE/OE0107/OE0107D/KostnDRLT`. En GET mot
rotadressen returnerar metadata med dimensionerna Region (regionkod i
formatet `XXL`), Verksomrkom (verksamhetsområde) och Tid (år). Data
hämtas via en POST där du väljer region, verksamhetsområden och år;
innehållskoden för nettokostnad är `000000A7`. Svar ges i PxWeb
JSON-format med värden i miljoner kronor (mnkr).

## Schema/fält vi använder
- `data[].key[0]` — regionkod (`XXL`).
- `data[].key[1]` — verksamhetsområde (`1` Hälso-/sjukvård, `2`
  Tandvård, `3` Övrig vård, ..., `0-9` total).
- `data[].values[0]` — mnkr (miljoner kronor).

## Begränsningar och kända problem
- **Gotland (region 09)**: SCB använder `0980L` här (kommun-style) —
  inte `09L`, eftersom Gotland är både region och kommun. Vår backend
  hanterar specialfallet automatiskt.
- Aktuella års data är ofta `..` (preliminär). Vår backend faller
  tillbaka från år-1 till år-5 för att hitta publicerade värden.

## Hur vi bearbetar
Vår backend hämtar nettokostnaderna från SCB med automatisk fallback
över flera år, mappar verksamhetsområdes-koderna till läsbara namn
(t.ex. `1` → "Hälso- och sjukvård"), och returnerar listan till
frontend för rendering på regionens budgetkort.
