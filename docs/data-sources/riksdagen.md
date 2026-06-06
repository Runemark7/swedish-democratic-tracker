---
id: riksdagen
name: Riksdagen Open Data
kind: api
upstream: https://data.riksdagen.se
license: Public domain — Sveriges riksdag
freshness: "dagligen (cron: politicians @daily, speeches @daily, votes @daily)"
last_verified: 2026-06-06
verification_status: unsure
verification_notes: "Reproduktionsnarrativet innehöll tre curl-kommandon (förbjudna i käll-MD); ersatta med prosaöversikt. Repro-narrativet är nu godkänt."
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

## Hur vi bearbetar
Tre parallella jobb hämtar data från Riksdagen varje dygn — ett för
politikerprofiler (`personlista`), ett för anföranden
(`anforandelista`), och ett för voteringsresultat (`voteringlista`).
Allt skrivs in i vår databas och blir tillgängligt via vårt API som
sedan driver UI:t.
