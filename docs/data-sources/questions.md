---
id: questions
name: Skriftliga frågor, svar och interpellationer (Riksdagen)
kind: api
upstream: https://data.riksdagen.se/dokumentlista/?doktyp=fr
license: Public domain — Sveriges riksdag
freshness: "dagligen för pågående riksmöte (cron: fetch-questions @daily); äldre riksmöten via cmd/backfill-questions"
last_verified: 2026-11-11
verification_status: verified
verification_notes: ""
used_by:
  - /api/politicians/{id}/questions (frågor och interpellationer en ledamot ställt eller, som statsråd, besvarat)
  - /api/motions/{riksmote}/{beteckning}/questions (frågor kopplade till en motion)
  - /api/questions/{kind}/{riksmote}/{beteckning} (en fråga med svar, debatt och länkade motioner)
---

## Vad det är
En skriftlig fråga (`doktyp=fr`) är en ledamots fråga till ett statsråd som besvaras skriftligt; svaret
är ett eget dokument (`frs`). En interpellation (`ip`) besvaras och debatteras i kammaren. Riksdagen
publicerar båda med frågeställare, statsråd och datum.

## Hur du själv kommer åt datan
Alla endpoints är publika och kräver ingen API-nyckel.

```bash
# lista frågor respektive interpellationer i ett riksmöte (@traffar är totalen)
curl -s 'https://data.riksdagen.se/dokumentlista/?doktyp=fr&rm=2022/23&utformat=json&sz=200&p=1&sort=datum&sortorder=desc'
curl -s 'https://data.riksdagen.se/dokumentlista/?doktyp=ip&rm=2022/23&utformat=json&sz=200&p=1&sort=datum&sortorder=desc'
# fråga 2022/23:339 och dess svar
curl -s 'https://data.riksdagen.se/dokumentstatus/HA11339.json'
curl -s 'https://data.riksdagen.se/dokumentstatus/HA12339.json'
```

## Schema/fält vi använder
- `dokumentlista.@traffar`, `dok_id`, `systemdatum` — täckning och version; en fråga läses om bara när `systemdatum` ändrats.
- `dokumentstatus.dokument` — `rm`, `beteckning`, `titel`, `datum`, `status`, `html` (frågetexten).
- `dokintressent` — `undertecknare`/`fragestallare` (ledamoten), `stalldtill` (statsrådet som frågan ställs till), `besvaradav` (statsrådet som svarade).
- `dokreferens` med `referenstyp=frågesvar` — pekar på svarsdokumentet (`frs`); dess `datum` är svarsdatum.
- `dokaktivitet` `Besvarad` — dagen en interpellation besvarades i kammaren.
- `dokreferens` med `ref_dok_typ=mot` — en motion som Riksdagen själv pekar ut (regel `riksdagen-reference`).

## Motionslänkar
En fråga kopplas till en motion bara när Riksdagens referens pekar på motionen, eller frågetexten
anger motionens beteckning (t.ex. "motion 2017/18:3293", regel `cited-beteckning`). Regeln sparas
med länken. En beteckning som står efter "proposition", "skrivelse" eller "betänkande", eller som
pekar på en motion vi inte har, ger ingen länk och sparas inte. En motion som nämns bara i löptext
("en kommittémotion från Moderaterna redan 2017") kopplas inte, t.ex. fråga 2022/23:339.

## Interpellationsdebatten
Anföranden hittas bland lagrade anföranden på avsnittsrubriken ("Svar på interpellation
2022/23:397 om …") och dagen interpellationen besvarades. Inget sparas; det slås upp vid läsning.

## Begränsningar och kända problem
- Ledamöter och statsråd identifieras med Riksdagens `intressent_id`; statsråd saknas ofta i ledamotslistan, så namnet lagras med frågan.
- Ett statsråd listas som besvarare först när frågan är besvarad.
- Täckningen jämför det vi har med det Riksdagen anger. Äldre riksmöten läses in av en operatör med `backfill-questions`, aldrig vid driftsättning.
- När detta skrevs fanns ingen inspelad fråga eller interpellation som citerar en motionsbeteckning eller har en motionsreferens; reglerna är testade på ett inspelat dokument med en tillagd mening.

## Hur vi bearbetar
Ett nattligt jobb (`fetch-questions`) tar det pågående riksmötet, högst 500 dokument per slag och
körning, och räknar sedan om motionslänkarna från den lagrade texten. Att köra om ändrar inget.
