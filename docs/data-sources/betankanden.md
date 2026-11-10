---
id: betankanden
name: Betänkanden och förslagspunkter (Riksdagen)
kind: api
upstream: https://data.riksdagen.se/dokumentlista/?doktyp=bet
license: Public domain — Sveriges riksdag
freshness: "dagligen (cron: fetch-betankanden @daily, högst 25 betänkanden per körning)"
last_verified: 2026-11-10
verification_status: verified
verification_notes: ""
used_by:
  - /motioner/:riksmöte/:beteckning (betänkande, förslagspunkt, beslut och Partilinje per yrkande)
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
- Motionens `dokforslag.forslag.behandlas_i` och `behandlas_i_punkt` (se källan Motioner).

## Begränsningar och kända problem
- `behandlas_i_punkt` är oftast tom. Då läser sajten förslagstexten: en motion som anges utan yrkande avser alla dess yrkanden, och "yrkandena 1–3 och 5", "63-66" och "yrkande 4 i denna del" tolkas. En motion som ingen punkt nämner får ingen koppling och visas som "Ej avgjord än" – sajten gissar inte.
- Varje koppling mellan ett yrkande och en punkt sparas med hur den gjordes (`behandlas_i_punkt` eller `forslag_text`).
- Kammarens beslut för ett yrkande är ibland tomt (t.ex. 2018/19:2900 yrkande 26). Då används punktens beslut enligt förslagstexten ("avslår", "bifaller", "bifaller delvis").
- Ett beslutsord som sajten inte känner igen visas som Riksdagen anger det.
- Votering-id skrivs med små bokstäver här och med stora i voteringsdatan; sajten jämför utan hänsyn till skiftläge.
- Riksmötet skrivs `2017/18`, aldrig `2017-18`.

## Hur vi bearbetar
Varje natt läser ett jobb högst 25 betänkanden som sparade motioner pekar på och som inte är
inlästa (eller är äldre än motionen som pekar på dem). Det går vidare nästa natt där det slutade.
För varje betänkande sparas förslagspunkterna och varje yrkande kopplas till den eller de
punkter som avgjorde det. Att köra om jobbet ändrar inget. Vid votering visas Partilinje per
parti enligt sajtens gemensamma definition; vid acklamation visas "Avgjord utan omröstning".
