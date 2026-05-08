---
id: riksdagen
name: Riksdagen Open Data
kind: api
upstream: https://data.riksdagen.se
license: Public domain — Sveriges riksdag
freshness: "dagligen (cron: politicians @daily, speeches @daily, votes @daily)"
last_verified: 2026-05-08
used_by:
  - / (Beslut idag, Aktuella debatter, Veckans omröstningar, Vad partierna säger, Kommande beslut)
  - /riksdag (live votes, agenda, mandat, government composition)
  - /votes (alla)
  - /votes/:beteckning/:punkt (vote details)
  - /politicians (alla)
  - /politicians/:id (profile, votes, speeches)
  - /anforanden/:id (speech detail)
  - /parties/:party/goals/:goalId/votes (vote alignment)
---

## Vad det är
Riksdagens öppna data är portalen som sveriges riksdag publicerar för
all kammaraktivitet: ledamöter, anföranden, omröstningar, betänkanden,
motioner och utskottsärenden. Det är primärkällan för all "vad har
politikerna gjort" -data på sidan.

## Hur du själv kommer åt datan
Alla endpoints är publika utan API-nyckel. Exempel:

```bash
# Lista nuvarande ledamöter i Moderaterna
curl -s 'https://data.riksdagen.se/personlista/?utformat=json&parti=M&iid=&fnamn=&enamn=&f_ar=&kn=&valkrets=&rdlstatus=tjanstgorande'

# Senaste betänkanden
curl -s 'https://data.riksdagen.se/dokumentlista/?doktyp=bet&utformat=json&sz=20'

# Anföranden för en ledamot
curl -s 'https://data.riksdagen.se/anforandelista/?iid=0123456789&utformat=json'
```

Full referens: <https://data.riksdagen.se/dokumentation/>.

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

## Hur vi bearbetar
- `backend/internal/politicians/adapters/riksdagen/client.go` hämtar
  `personlista`.
- `backend/internal/speeches/adapters/riksdagen/client.go` hämtar
  `anforandelista`.
- `backend/internal/votes/adapters/riksdagen/client.go` hämtar
  `voteringlista`.
- Alla resultat upsertas via `UpsertMany` i sina respektive Postgres
  adapters. Se schedulern i
  `backend/internal/ingestion/workers/` för cronen.
