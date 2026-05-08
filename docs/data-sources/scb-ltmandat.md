---
id: scb-ltmandat
name: SCB Ltmandat (regionala mandat)
kind: api
upstream: https://api.scb.se/OV0104/v1/doris/sv/ssd/ME/ME0104/ME0104B/Ltmandat
license: PxWeb open data — fri användning
freshness: 4-årscykel (val)
last_verified: 2026-05-08
used_by:
  - /region/:code (mandat, hemicycle, koalitionstyp)
---

## Vad det är
SCB:s tabell över mandatfördelning i regionfullmäktige (landsting).
Syster till Kfmandat men för regionnivå.

## Hur du själv kommer åt datan
```bash
curl -s -X POST \
  'https://api.scb.se/OV0104/v1/doris/sv/ssd/ME/ME0104/ME0104B/Ltmandat' \
  -H 'Content-Type: application/json' \
  -d '{"query":[{"code":"Region","selection":{"filter":"item","values":["01L"]}},{"code":"Parti","selection":{"filter":"item","values":["M","C","FP","KD","MP","S","V","SD","ÖVRIGA"]}},{"code":"ContentsCode","selection":{"filter":"item","values":["ME0104C2"]}},{"code":"Tid","selection":{"filter":"item","values":["2022"]}}],"response":{"format":"json"}}'
```

Regionkod-format: tvåsiffrig kommunkod-prefix + `L` (t.ex. `01L`,
`14L`). Endast koder som slutar på `L` används.

## Schema/fält vi använder
- `data[].key[0]` — regionkod (`XXL`).
- `data[].key[1]` — partikod.
- `data[].values[0]` — antal mandat.

## Begränsningar och kända problem
- Gotland (kod `09`) finns inte i Ltmandat eftersom Gotland är både
  region och kommun — använd Kfmandat-koden `0980` istället när det
  gäller mandat. (Gäller även [scb-kostndrlt](./scb-kostndrlt.md).)

## Hur vi bearbetar
- `backend/internal/regions/seeder/seeder.go:88` (`fetchMandates` samma
  funktion, olika URL).
- Resultat skrivs till `regional_election_results`.
- Koalitionstyp ("Borgerlig majoritet" / "Rödgrön minoritet" / etc.) härleds
  från mandatfördelningen i `deriveGoverning()` (`seeder.go:365`). Den
  härledningen hör hemma här, inte i en egen synthesized-källa, eftersom
  den är trivial och deterministisk.
