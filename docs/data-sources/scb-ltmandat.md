---
id: scb-ltmandat
name: SCB Ltmandat (regionala mandat)
kind: api
upstream: https://api.scb.se/OV0104/v1/doris/sv/ssd/ME/ME0104/ME0104B/Ltmandat
license: PxWeb open data — fri användning
freshness: 4-årscykel (val)
last_verified: 2026-06-08
verification_status: verified
verification_notes: ""
used_by:
  - /region/:code (mandat, hemicycle, koalitionstyp)
---

## Vad det är
SCB:s tabell över mandatfördelning i regionfullmäktige (landsting).
Syster till Kfmandat men för regionnivå.

## Hur du själv kommer åt datan
Tabellen heter **Ltmandat** och finns i SCB:s statistikdatabas under
val och demokrati. Du når den via SCBs webbgränssnitt på
<https://www.statistikdatabasen.scb.se/> — sök på "Ltmandat" eller
navigera via Val → Allmänna val → Landstingsfullmäktige → Ltmandat.

Teknisk åtkomst sker via SCB:s PxWeb API med sökvägen
`/OV0104/v1/doris/sv/ssd/ME/ME0104/ME0104B/Ltmandat`. En GET mot
rotadressen returnerar metadata med dimensionerna Region (regionkod),
Parti och Tid (valår). Data hämtas via en POST med ett JSON-urval där
du väljer önskade regioner (koder i formatet `XXL`, t.ex. `01L` för
Stockholm), partier och år. Innehållskoden för mandat är `ME0104C2`.
Svar ges i PxWeb JSON-format.

Regionkod-format: tvåsiffrig länsbokstavsprefix + `L` (t.ex. `01L`,
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
Vår backend hämtar mandaten från SCB en gång per driftsättning och
lagrar dem per region. Frontend ritar därefter en hemicycle med
mandatfördelning på varje regionsida.

Koalitionstypen ("Borgerlig majoritet", "Rödgrön minoritet" osv.)
härleds direkt från mandatfördelningen: räkna mandat per block och
jämför mot 50%-tröskeln. Härledningen är trivial och deterministisk
så den dokumenteras här i stället för som en egen källa.
