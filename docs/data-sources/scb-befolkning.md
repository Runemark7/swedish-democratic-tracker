---
id: scb-befolkning
name: SCB BefolkningNy
kind: api
upstream: https://statistikdatabasen.scb.se/api/v2/tables/TAB638/metadata?lang=sv
license: PxWeb open data — fri användning
freshness: årlig (kvartal för kommunnivå)
last_verified: 2026-06-08
verification_status: verified
verification_notes: ""
used_by:
  - /region landing (population per region)
  - /kommun landing (population per kommun)
  - /kommun/:code (population sub-header)
  - /region/:code (population sub-header)
---

## Vad det är
SCB:s folkmängdsdatabas — total befolkning per kommun.

## Hur du själv kommer åt datan
Tabellen heter **BefolkningNy** och finns i SCB:s statistikdatabas under
ämnesområdet Befolkning (BE). Du når den via SCBs webbgränssnitt på
<https://www.statistikdatabasen.scb.se/> — sök på "BefolkningNy" eller
navigera via Befolkning → Befolkningsstatistik → Folkmängd → BefolkningNy.

Sedan november 2026 läser vi tabellen via SCB:s PxWebApi v2 som tabell
`TAB638` (`https://statistikdatabasen.scb.se/api/v2/tables/TAB638`; metadata
via `/metadata`, data via POST till `/data?outputFormat=json-px`). Det
gamla v1-API:t nedan används bara som reserv tills SCB stänger det vid
årsskiftet 2026/2027. Båda versionerna svarar ibland med fel (v2 500, v1
400) på giltiga frågor, så klienten (`backend/internal/scbpx`) försöker
igen innan den ger upp.

Teknisk åtkomst sker via SCB:s PxWeb API. Tabellens sökväg i API:et är
`/OV0104/v1/doris/sv/ssd/BE/BE0101/BE0101A/BefolkningNy`. En GET mot
denna adress returnerar tabellens metadata (vilka dimensioner och koder
som är tillgängliga). Data hämtas via en POST med ett JSON-urval där du
anger region (kommunkod), innehållskod (`BE0101N1` för folkmängd) och
tidsperiod (år). Svar ges i PxWeb JSON-format.

## Schema/fält vi använder
- `data[].key[0]` — kommunkod.
- `data[].values[0]` — antal invånare.

## Begränsningar och kända problem
- Endast kommunnivå publiceras direkt — regionsumma räknas fram av
  oss genom att summera ingående kommuner.
- SCB tillåter ~50 koder per query, så hämtningen sker i flera
  omgångar.

## Hur vi bearbetar
Vår backend hämtar folkmängd för alla kommuner i batch en gång per
driftsättning, och lagrar siffran per kommun. Frontend visar den i
sub-headern på region- och kommunsidan samt på respektive landing-sida.
