---
id: riksdagen
name: Riksdagen Open Data
kind: api
upstream: https://data.riksdagen.se
license: Public domain — Sveriges riksdag
freshness: "dagligen (cron: politicians @daily, speeches @daily, votes @daily)"
last_verified: 2026-06-08
verification_status: verified
verification_notes: ""
used_by:
  - / (Beslut idag, Aktuella debatter, Veckans omröstningar, Vad partierna säger)
  - /regeringen#riksdag (live votes, mandat, government composition)
  - /regeringen (regeringstyp)
  - /votes (alla)
  - /votes/:beteckning/:punkt (vote details)
  - /politicians (alla)
  - /politicians/:id (profile, votes, speeches)
  - /anforanden/:id (speech detail)
  - /beslut/:riksmöte/:beteckning och /motioner/:riksmöte/:beteckning (debatten: anföranden kopplade till betänkandet)
  - /parties/:party/goals/:goalId/votes (vote alignment)
---

## Vad det är
Riksdagens öppna data är portalen som Sveriges riksdag publicerar för
all kammaraktivitet: ledamöter, anföranden, omröstningar, betänkanden,
motioner och utskottsärenden. Det är primärkällan för all "vad har
politikerna gjort"-data på sidan.

## Hur du själv kommer åt datan
Alla endpoints är publika och kräver ingen API-nyckel. Startpunkten är
<https://data.riksdagen.se> där dokumentationen finns samlad på
<https://data.riksdagen.se/dokumentation/>.

Tre huvudtyper av data används:

**Ledamöter** — via `/personlista/` med filter på parti, valkrets eller
statusfältet `tjanstgorande`. Svaret kan begäras som JSON genom att
lägga till parametern `utformat=json`.

**Betänkanden och dokument** — via `/dokumentlista/` med filter på
dokumenttyp (`doktyp=bet` för betänkanden). Ger en lista med
`beteckning`-fält och `dok_id` som identifierar enskilda ärenden.

**Omröstningar och anföranden** — via `/voteringlista/` respektive
`/anforandelista/` med filter på ledamotens `iid`-kod. Returnerar
per-ledamots röst (Ja/Nej/Avstår/Frånvarande) och anförandetext.

## Schema/fält vi använder
- `personlista.person.intressent_id` — primärnyckel för politiker.
- `personlista.person.tilltalsnamn` / `efternamn` — namn.
- `personlista.person.parti` — partikod (S, M, ...).
- `dokumentlista.dokument.beteckning` / `dok_id` — för omröstningar.
- `voteringlista.votering.rost` — Ja/Nej/Avstår/Frånvarande per ledamot.
- `anforandelista.anforande.anforande_text` — anförandetext för debatter.

## Begränsningar och kända problem
- API:n returnerar XML som default. Lägg till `&utformat=json` för JSON.
- Stora datasetuppdateringar är inte realtid — workers kör @daily, så
  färska beslut syns nästa dag.
- `dok_datum` är ofta endast datum-precision, inte tid.
- **`beteckning` byter format 2014/15.** För riksmötena 2002/03 till
  2013/14 returnerar `/dokumentlista/?doktyp=votering` beteckningen med
  förslagspunkten inbakad — `FIU20p1` i stället för `FiU20`. Från 2014/15
  och framåt är formatet det senare. Brottet är rent: kontrollerat mot
  samtliga riksmöten 2002/03–2025/26 den 2026-08-22, tolv riksmöten på
  var sida, inget blandat.

  Det spelar roll eftersom `/voteringlista/?bet=FIU20p1` inte matchar
  någonting alls — rösterna ligger under `bet=FIU20`. Vi normaliserar
  därför bort suffixet vid inläsning. Förslagspunkten går inte förlorad:
  varje röstrad bär sin egen `forslagspunkt` från `/voteringlista/`.

## Hur vi bearbetar
Tre parallella jobb hämtar data från Riksdagen varje dygn — ett för
politikerprofiler (`personlista`), ett för anföranden
(`anforandelista`), och ett för voteringsresultat (`voteringlista`).
Allt skrivs in i vår databas och blir tillgängligt via vårt API som
sedan driver UI:t.

## Anföranden kopplade till betänkandet
Varje anförande kopplas till det betänkande det debatterar (jobbet `link-speeches`, dagligen; kan köras om utan risk):

1. `rel_dok_id` i `/anforandelista/` anger ett betänkande (till exempel `H901MJU33` för debatten 2022-09-21). Interpellationer, motioner och propositioner har id som bara slutar på siffror och räknas inte.
2. Annars: anförandets `avsnittsrubrik` är lika med betänkandets titel och protokollets datum (`dok_datum`) är lika med betänkandets debattdag (raden "Debatt" i `sokdata.statusrad` i betänkandelistan). Matchar fler än ett betänkande kopplas inget. Så ser 2002/03 ut: `rel_dok_id` saknas där helt.
3. Annars förblir anförandet okopplat. Det syns kvar på ledamotens sida. Sajten gissar inte.

Regeln som användes sparas med kopplingen (`rel_dok_id` eller `avsnittsrubrik_datum`) och visas på sidan.

Reproducera (2022-09-21, betänkande 2021/22:MJU33):

```bash
curl -s --compressed --http1.1 'https://data.riksdagen.se/anforandelista/?rm=2021/22&parti=MP&iid=&sz=10000&anftyp=Akt&utformat=json' \
  | python3 -c "import json,sys; print([a['anforande_nummer'] for a in json.load(sys.stdin)['anforandelista']['anforande'] if a['rel_dok_id']=='H901MJU33'])"
```
