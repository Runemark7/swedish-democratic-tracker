---
id: loften-valmanifest
name: Löften, kopplingar och Löftesstatus
kind: synthesized
upstream: ~
license: Citat ur partiernas egna valmanifest, med källänk; röstdata ur Riksdagens öppna data
freshness: löpande — Löften och kopplingar bekräftas av redaktören; Löftesstatus räknas om vid varje visning
last_verified: 2026-11-11
verification_status: verified
verification_notes: ~
used_by:
  - /loften (Löften per parti, med Löftesstatus och filter)
  - /loften/:id (ett Löfte, omröstningarna bakom statusen)
  - /kopplingar/:id (en bekräftad koppling)
  - /metod (så räknas Löftesstatus)
---

## Vad det är
Ett Löfte är ett konkret, prövbart uttalande i ett partis valmanifest. Varje
Löfte har ett citat och en länk till originaldokumentet. En koppling knyter
ett Löfte till en omröstning i riksdagen. Redaktören bekräftar varje koppling
och anger vilken röst, Ja eller Nej, som går i Löftets riktning. Först då
räknas kopplingen.

Löftesstatus är det enda som härleds: Röstat för, Röstat emot, Blandat eller
Inte prövat än. Den räknas ur partiets Partilinje i de bekräftade
omröstningarna. Webbplatsen säger aldrig att ett löfte är "uppfyllt" eller
något annat omdöme om utfallet. Se [Metod](/metod).

## Hur du själv kommer åt datan
Allt vi visar kan kontrolleras mot två öppna källor:

1. Löftets citat och sida i partiets valmanifest, via länken på Löftets sida.
   Katalogen över manifest finns i [party-manifestos](./party-manifestos.md).
2. Omröstningen i Riksdagens öppna data, se [voteringar](./voteringar.md):

```bash
curl -s 'https://data.riksdagen.se/votering/<votering_id>/json'
```

Samma data finns i vårt öppna API:

```bash
curl -s 'https://sdt.runevibe.se/api/loften?status=for'
curl -s 'https://sdt.runevibe.se/api/loften/<id>'
curl -s 'https://sdt.runevibe.se/api/kopplingar/<id>'
```

## Schema/fält vi använder
- `status` — Löftesstatus; saknas för arkiverade Löften.
- `kopplingar[].direction` — den röst (Ja eller Nej) redaktören bekräftat går i Löftets riktning.
- `kopplingar[].partilinje` — partiets Partilinje i omröstningen.
- `kopplingar[].confirmedAt` — när kopplingen bekräftades.
- `statusChangedAt`, `statusReason` — när och varför statusen senast ändrades.

## Begränsningar och kända problem
- Arkiverade löften från valet 2022 har ingen Löftesstatus: vi har inte kopplat dem till omröstningar, och en status skulle visa vår egen lucka som ett påstående om partiet.
- Bara omröstningar inom Löftets mandatperiod räknas.
- En omröstning avgör ofta mer än ett förslag. Därför bekräftar en människa varje koppling i stället för att vi gissar riktningen.
- Avvikare, ledamöter som röstar mot sin Partilinje, visas men ändrar aldrig partiets status.

## Hur vi bearbetar
Vid varje visning tar vi de bekräftade kopplingarna för Löftet och räknar
partiets Partilinje i varje omröstning: den röst som flest av partiets
ledamöter avgav, där Frånvarande inte räknas. Är Partilinjen densamma som
kopplingens riktning räknas omröstningen som "för", är den motsatt räknas den
som "emot". Är Partilinjen Avstår eller Delad anger den ingen riktning och
omröstningen räknas varken för eller emot, men visas ändå. Finns både "för"
och "emot" blir statusen Blandat; finns ingendera, Inte prövat än. Ångras en
bekräftelse försvinner kopplingen och statusen räknas om.

## Löften on decisions and the front page (#162)

The public read also answers "which Löften does this decision touch?" and
fills the front-page party tiles. Both count **Bekräftade kopplingar on public
Löften only**: a Löfte touches a betänkande when a Bekräftad koppling points at
a votering named by one of its förslagspunkter. Föreslagna, avvisade and
bortsorterade kopplingar, and Löften that are not public, never count. No new
table and no new external source; `GET /api/loften?betankande=2026/27:SkU2`
returns exactly the Löften behind a "berör N löften" link.
