---
id: voteringar
name: Voteringar (Riksdagen)
kind: api
upstream: https://data.riksdagen.se/voteringlista/
license: Public domain — Sveriges riksdag
freshness: "dagligen (cron: fetch-votes @daily); täckningens nämnare räknas om veckovis (söndagar)"
last_verified: 2026-11-11
verification_status: verified
verification_notes: ""
used_by:
  - /data (täckning per riksmöte: voteringar lästa mot Riksdagens egen siffra)
  - /beslut/:riksmöte/:beteckning (hur varje ledamot och parti röstade)
  - /politicians/:id (ledamotens röster)
---

## Vad det är
En votering är en omröstning i kammaren om en förslagspunkt i ett betänkande. Riksdagen
publicerar hur varje ledamot röstade (Ja, Nej, Avstår eller Frånvarande) och vilket datum
omröstningen hölls. Beslut som fattas utan omröstning (acklamation) har ingen votering.

## Hur du själv kommer åt datan
Alla endpoints är publika och kräver ingen API-nyckel.

```bash
# listan över voteringar i ett riksmöte (@traffar är Riksdagens siffra)
curl -s 'https://data.riksdagen.se/dokumentlista/?doktyp=votering&rm=2021/22&utformat=json&sz=200&p=1&sort=datum&sortorder=desc'
# rösterna i ett betänkande
curl -s 'https://data.riksdagen.se/voteringlista/?rm=2021/22&parti=&iid=&bet=SkU16&sz=10000&utformat=json'
```

## Schema/fält vi använder
- `dokumentlista.@traffar` — antal voteringar Riksdagen anger för riksmötet (täckningens nämnare).
- `voteringlista.votering.votering_id` — identiteten; två voteringar kan dela beteckning och förslagspunkt.
- `voteringlista.votering.forslagspunkt` / `rost` / `parti` — punkt, ledamotens röst och parti.
- `systemdatum` — när Riksdagen senast ändrade posten, inte beslutsdagen.

## Begränsningar och kända problem
- Voteringar utan beteckning (om en motion i stället för ett betänkande) kan inte hämtas. De räknas som "onåbara" så ett underskott är förklarat i stället för okänt.
- Acklamation lämnar ingen votering. Sådana beslut syns som "Avgjord utan omröstning" och ingår inte i täckningen.
- `beteckning` byter format 2014/15 (se Riksdagen Open Data).
- Täckningens nämnare räknas om separat från inläsningen och kan därför vara äldre än antalet lästa.

## Länkregler
En votering hör till ett betänkande och en förslagspunkt via Riksdagens egen `beteckning` och `forslagspunkt`, nycklade på riksmöte. Sajten kopplar inte röster till löften eller motioner på egen hand.

## Täckning
Status och senaste uppdatering per riksmöte står under Data (`/data`, API: `/api/data/coverage`). Ett riksmöte är klart när det vi har når Riksdagens egen siffra.

## Hur vi bearbetar
Ett nattligt jobb läser voteringarna för det pågående riksmötet. Ett veckojobb räknar om
täckningen mot Riksdagens lista. Att köra om ändrar inget.
