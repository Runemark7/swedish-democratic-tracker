---
id: betankanden
name: Betänkanden och förslagspunkter (Riksdagen)
kind: api
upstream: https://data.riksdagen.se/dokumentlista/?doktyp=bet
license: Public domain — Sveriges riksdag
freshness: "dagligen (cron: fetch-betankanden @daily, högst 25 betänkanden per körning; sync-beslut-list @daily för listan)"
last_verified: 2026-11-10
verification_status: verified
verification_notes: ""
used_by:
  - /motioner/:riksmöte/:beteckning (betänkande, förslagspunkt, beslut och Partilinje per yrkande)
  - /beslut (listan över betänkanden: kommande och fattade beslut, med filter)
---

## Vad det är
Ett betänkande är utskottets förslag till riksdagen. Det delas in i förslagspunkter, och
varje motions yrkanden avgörs i en förslagspunkt. Kammaren beslutar antingen genom
acklamation (ingen omröstning) eller genom votering. Riksdagen publicerar betänkandet,
varje punkts förslagstext, beslutstyp och, vid votering, votering-id.

## Hur du själv kommer åt datan
Alla endpoints är publika och kräver ingen API-nyckel.

Hitta betänkandet (här 2017/18:SkU11, `dok_id` H501SkU11):

```bash
curl -s 'https://data.riksdagen.se/dokumentlista/?doktyp=bet&rm=2017/18&bet=SkU11&sz=5&utformat=json'
```

Hämta betänkandet med förslagspunkterna (`dokutskottsforslag`):

```bash
curl -s 'https://data.riksdagen.se/dokumentstatus/H501SkU11.json'
```

Läs betänkandet på Riksdagens sida: <https://data.riksdagen.se/dokument/H501SkU11.html>.

Reservationer, särskilda yttranden och "Utskottets ställningstagande" finns bara i HTML-dokumentet ovan. Sajten läser dem per avsnittsrubrik ("Utskottets överväganden", "Reservationer", "Särskilda yttranden"). Saknas eller går ett avsnitt inte att tolka visas det som "inte tillgänglig från Riksdagen" med länk till källan; inget fylls i.

## Schema/fält vi använder
- `dokumentlista.dokument.dok_id` — betänkandets identitet, slås upp via riksmöte och beteckning.
- `dokumentstatus.dokument` (`rm`, `beteckning`, `titel`, `organ`, `systemdatum`).
- `dokutskottsforslag.utskottsforslag` — en post per förslagspunkt:
  - `punkt`, `rubrik` — numret och rubriken.
  - `forslag` — utskottets förslagstext, där motionerna räknas upp ("Riksdagen avslår motionerna 2017/18:3293 av … och 2017/18:3748 av … yrkande 18").
  - `beslutstyp` — `acklamation` eller `röstning`.
  - `votering_id` — kopplar till omröstningen vid `röstning`.
  - `votering_sammanfattning_html` — raden som anger vad omröstningen gällde, t.ex. "Utskottets förslag mot reservation 12 (M)".
  - `vinnare` — vilket förslag som vann.
- `dokumentlista.dokument` (`status`, `beslutsdag`, `beslutad`) — betänkandets beslutsstatus och beslutsdag. Läses vid varje visning av betänkandesidan, aldrig sparat, så att sidan inte hamnar efter när kammaren beslutar.
- `dokumentlista` per riksmöte (`doktyp=bet&rm=2026/27&sz=200&sort=datum&sortorder=desc`) — en rad per betänkande med `organ`, `titel`, `status`, `beslutsdag`, `beslutad` och `datum`. Listan över beslut (`/beslut`) bygger på de här raderna. Till skillnad från betänkandesidan sparas de, i tabellen `beslut_listing`, och skrivs över vid varje körning; "avgjort" räknas ut från de sparade fälten med samma regel som betänkandesidan (`beslutad`, annars `status` och `beslutsdag`) och sparas aldrig som ett eget värde.
- `dokumentlista.dokument.sokdata.statusrad` — raden `<dt>Debatt</dt> <dd>2018-03-22</dd>` ger debattdagen. Den sparas och används för att koppla anföranden som saknar `rel_dok_id` (se källan Riksdagen Open Data).
- Motionens `dokforslag.forslag.behandlas_i` och `behandlas_i_punkt` (se källan Motioner).

## Begränsningar och kända problem
- `behandlas_i_punkt` är oftast tom. Då läser sajten förslagstexten: en motion som anges utan yrkande avser alla dess yrkanden, och "yrkandena 1–3 och 5", "63-66" och "yrkande 4 i denna del" tolkas. En motion som ingen punkt nämner får ingen koppling och visas som "Ej avgjord än" – sajten gissar inte.
- Varje koppling mellan ett yrkande och en punkt sparas med hur den gjordes (`behandlas_i_punkt` eller `forslag_text`).
- Kammarens beslut för ett yrkande är ibland tomt (t.ex. 2018/19:2900 yrkande 26). Då används punktens beslut enligt förslagstexten ("avslår", "bifaller", "bifaller delvis").
- Ett beslutsord som sajten inte känner igen visas som Riksdagen anger det.
- Votering-id skrivs med små bokstäver här och med stora i voteringsdatan; sajten jämför utan hänsyn till skiftläge.
- Beslutsdatum och "avgjort" kommer från Riksdagens `beslutad`, `status` och `beslutsdag` (inte från vår egen tidsstämpel). Ett planerat betänkande visas som "Ej avgjort än" med planerad dag. Kan listan inte läsas visas ingen status.
- Riksmötet skrivs `2017/18`, aldrig `2017-18`.

## Hur vi bearbetar
Samma jobb läser även betänkandets HTML (en begäran till). Misslyckas den läsningen försöker jobbet igen efter ett dygn; redan inlästa betänkanden utan HTML läses om i jobbets vanliga takt (högst 25 per körning).

Varje natt läser ett jobb högst 25 betänkanden som sparade motioner från det pågående riksmötet pekar på och som inte är
inlästa (eller är äldre än motionen som pekar på dem). Det går vidare nästa natt där det slutade. Betänkanden för äldre riksmöten läses in av operatörens kommando `backfill-motions` (se `docs/runbooks/backfill-motions.md`).
För varje betänkande sparas förslagspunkterna och varje yrkande kopplas till den eller de
punkter som avgjorde det. Att köra om jobbet ändrar inget. Vid votering visas Partilinje per
parti enligt sajtens gemensamma definition; vid acklamation visas "Avgjord utan omröstning".

## Listan över beslut
Jobbet `sync-beslut-list` (dagligen) läser betänkandelistan för det pågående och det föregående riksmötet vid varje körning. Äldre riksmöten från mandatperioderna läses en gång, högst tre nya per körning, och hoppas sedan över. Att köra om jobbet ändrar inget; ett riksmöte skrivs först när alla dess sidor lästs. Listan visar kommande beslut först (närmast planerad beslutsdag först, utan dag sist) och därefter fattade beslut, senaste först. Partifiltret visar betänkanden som avgör en motion från partiet eller där partiet har en reservation; ett parti röstar i varje votering, så "har röstat" skulle inte avgränsa något. Partiets linje i listan är Partilinje per votering enligt den gemensamma definitionen; skiljer den sig mellan betänkandets voteringar visas "Varierar".

## Länkregler
- **Riksdagens referens**: motionens `behandlas_i` (och `behandlas_i_punkt` när den finns) pekar ut betänkandet och punkten.
- **Citerad beteckning**: när `behandlas_i_punkt` saknas läser sajten förslagstexten, där motionen och dess yrkanden anges med beteckning (regel `forslag_text`).
- **Identisk ordalydelse** används inte för att koppla ett yrkande till en punkt, bara för att gruppera återkommande krav (se Motioner).

## Täckning
Betänkanden räknas som inlästa för ett riksmöte när inget betänkande som riksmötets motioner pekar på återstår att läsa. Status och senaste uppdatering per riksmöte står under Data (`/data`, API: `/api/data/coverage`).
