---
id: speeches-backfill
name: Anföranden, äldre riksmöten (Riksdagen)
kind: api
upstream: https://data.riksdagen.se/anforandelista/
license: Public domain — Sveriges riksdag
freshness: "engångskörning av en operatör, ett riksmöte i taget (kommandot backfill-speeches); löpande riksmöte via fetch-speeches @daily"
last_verified: 2026-11-11
verification_status: verified
verification_notes: ""
used_by:
  - /data (täckning per riksmöte: anföranden lästa, kopplade och okopplade)
  - /beslut/:riksmöte/:beteckning och /motioner/:riksmöte/:beteckning (debatten: anföranden kopplade till betänkandet)
  - /politicians/:id (ledamotens anföranden, även äldre riksmöten)
---

## Vad det är
Riksdagen publicerar varje anförande i kammaren, med talare, parti, protokolldatum,
rubrik och, för de flesta debatter, vilket betänkande det gäller. Sajten läste länge
bara in de senaste riksmötena. Den här källan är inläsningen av äldre riksmöten, ett
i taget och de nyaste först, så att man kan följa en fråga genom debatten i kammaren
tillbaka i tiden. Första omgången är de två senaste mandatperioderna och 2014–2018.

## Hur du själv kommer åt datan
Alla endpoints är publika och kräver ingen API-nyckel. Listan över ett riksmöte
hämtas ett parti i taget (den kan inte bläddras, och svaret tar slut vid 10 000 rader):

```bash
curl -s --compressed --http1.1 \
  'https://data.riksdagen.se/anforandelista/?rm=2017/18&parti=M&iid=&sz=10000&anftyp=Akt&utformat=json'
```

Själva texten ligger inte i listan. Den hämtas per anförande (här anförande 158 i
protokoll H50989):

```bash
curl -s 'https://data.riksdagen.se/anforande/H50989-158.json'
```

Talaren hämtas ur personregistret, även när hen inte längre är ledamot
(Riksdagen svarar med ett objekt, inte en lista, när det finns exakt en träff):

```bash
curl -s 'https://data.riksdagen.se/personlista/?iid=0117203219614&utformat=json'
```

## Schema/fält vi använder
- `anforandelista.anforande.dok_id` + `anforande_nummer` — anförandets identitet.
- `dok_rm` — riksmötet protokollet tillhör.
- `intressent_id`, `parti` — talare och partikod vid tillfället.
- `dok_datum`, `avsnittsrubrik`, `rel_dok_id` — datum, rubrik och betänkandet anförandet gäller.
- `anforande.anforandetext` — texten, hämtad per anförande.
- `personlista.person.*` — namn, parti, valkrets och status för talare som saknas hos oss.

## Begränsningar och kända problem
- Listan hämtas för S, M, SD, C, V, KD, L, MP och FP (Folkpartiet hette FP till 2015).
  Talare utan partikod, som talmannen och statsråd utanför riksdagen, ingår inte i någon
  partis lista och läses därför inte in.
- Äldre riksmöten har partikoder med små bokstäver (`s`, `fp`); Riksdagen matchar
  parametern utan hänsyn till versaler.
- I 2002/03 saknas `rel_dok_id` helt. Anförandena kopplas då bara om rubriken
  och protokollets datum entydigt pekar ut ett betänkande; annars förblir de okopplade.
- Talare som Riksdagen inte har någon person för skapas inte. Deras anföranden läses inte
  in, och riksmötet markeras inte som klart. Det syns under Data.

## Hur vi bearbetar
En operatör kör kommandot ett riksmöte i taget, de nyaste först. Det läser listan,
skapar de talare som saknas ur Riksdagens personregister (före anförandet sparas),
sparar anförandena, läser texten med en liten paus mellan varje förfrågan och kopplar
varje anförande till betänkandet det debatterar enligt samma regler som övriga
anföranden (se Riksdagen Open Data). Ett riksmöte räknas som klart först när alla
anföranden som listan anger är sparade med text. Kommandot kan köras om utan risk:
klara riksmöten hoppas över och inget sparat ändras. För varje riksmöte noteras hur
mycket lagringsutrymme det tog (rader, text, tabell och index). Under Data visas, per
riksmöte, hur många anföranden som lästs, kopplats och inte kopplats. Sajten drar inga
slutsatser av dem.

## Länkregler
Ett anförande kopplas till ett betänkande bara via Riksdagens referens (`rel_dok_id`) eller när avsnittsrubriken och protokollets datum entydigt pekar ut ett betänkande. Annars förblir det okopplat.

## Täckning
Status och senaste uppdatering per riksmöte står under Data (`/data`, API: `/api/data/coverage`).
