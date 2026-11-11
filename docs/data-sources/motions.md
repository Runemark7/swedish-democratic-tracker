---
id: motions
name: Motioner (Riksdagen)
kind: api
upstream: https://data.riksdagen.se/dokumentlista/?doktyp=mot
license: Public domain — Sveriges riksdag
freshness: "dagligen för pågående riksmöte (cron: fetch-motions @daily); äldre riksmöten via backfill-motions"
last_verified: 2026-11-11
verification_status: verified
verification_notes: ""
used_by:
  - /motioner/:riksmöte/:beteckning (titel, typ, undertecknare, yrkanden och beslut som de anges)
---

## Vad det är
En motion är ett förslag till riksdagen från en eller flera ledamöter. Riksdagen
publicerar varje motion med undertecknare, ett numrerat yrkande per krav och,
när utskottet och kammaren har behandlat den, vad som beslutades. Samma uppgifter
finns i Riksdagens öppna data.

## Hur du själv kommer åt datan
Alla endpoints är publika och kräver ingen API-nyckel.

Lista motionerna i ett riksmöte (sidindelad, `@traffar` anger totalen):

```bash
curl -s 'https://data.riksdagen.se/dokumentlista/?doktyp=mot&rm=2017/18&utformat=json&sz=200&p=1&sort=datum&sortorder=desc'
```

Hämta en motion med undertecknare och yrkanden (här 2017/18:3293, `dok_id` H5023293):

```bash
curl -s 'https://data.riksdagen.se/dokumentstatus/H5023293.json'
```

Läs motionen på Riksdagens sida: <https://data.riksdagen.se/dokument/H5023293.html>.

## Schema/fält vi använder
- `dokumentlista.@traffar` — antal motioner Riksdagen anger för riksmötet (täckningens nämnare).
- `dokumentlista.dokument.dok_id` / `systemdatum` — identitet och version; en motion läses om bara när `systemdatum` ändrats.
- `dokumentstatus.dokument.rm` / `beteckning` — riksmöte (`2017/18`) och motionens nummer (`3293`).
- `dokumentstatus.dokument.subtyp` — Riksdagens egen typ: Kommittémotion, Partimotion eller Enskild motion. Sparas som den ges.
- `dokumentstatus.dokument.titel`, `datum`, `organ` — rubrik, inlämningsdag och utskottskod.
- `dokumentstatus.dokuppgift` (`kod=tilldelat`) — utskottets namn.
- `dokumentstatus.dokintressent` — undertecknare med `partibet`, partiet som Riksdagen angav på motionen.
- `dokumentstatus.dokforslag.forslag` — ett yrkande per post: `nummer`, `lydelse` (ordalydelse), `utskottet` (utskottets förslag), `kammaren` (kammarens beslut), `behandlas_i`.
- `dokumentstatus.dokreferens` (`referenstyp=behandlas_i`) — betänkandet motionen behandlas i, t.ex. `2017/18:SkU11`.

## Begränsningar och kända problem
- Riksdagen har ingen typ "Flerpartimotion". Sajten visar en enskild motion vars undertecknare kommer från mer än ett parti (t.ex. 2017/18:4202 av C, KD och M) som Flerpartimotion; originaltypen "Enskild motion" sparas oförändrad. Ledamöter utan parti räknas inte som ett parti.
- En typ som sajten inte känner igen visas exakt som Riksdagen anger den, utan förklaring.
- Partiet är det som stod på motionen när den lämnades in, inte ledamotens parti i dag.
- `behandlas_i` kan peka på ett senare riksmöte än motionens (2017/18:4202 behandlas i 2018/19:CU2).
- Fälten `dokforslag`, `dokintressent` m.fl. är ett objekt när det finns en post och en lista när det finns flera; en motion utan yrkanden har `null`.
- Riksmötet skrivs `2017/18` i Riksdagens data och på sajten, aldrig `2017-18`.
- Täckningen jämför det vi har med det Riksdagen anger. Äldre riksmöten läses in i omgångar över flera dygn.

## Hur vi bearbetar
Varje natt tar ett jobb det pågående riksmötet (högst 500 motioner per körning). Äldre riksmöten
läses in av en operatör med kommandot `backfill-motions`, ett riksmöte i taget, de nyaste först
tillbaka till 2002/03 (se `docs/runbooks/backfill-motions.md`). Kommandot kör aldrig av sig självt
och inte vid driftsättning; under Data visas vilka riksmöten som är inlästa. För varje riksmöte läser det Riksdagens lista, noterar hur många motioner
listan anger och läser in de motioner som är nya eller har ändrats. Att köra om jobbet
ändrar inget. Sajten visar uppgifterna som de anges, med länk till Riksdagens
dokument, och drar inga slutsatser av dem. Typen Flerpartimotion är det enda vi
härleder, och den följer av undertecknarnas partier.
