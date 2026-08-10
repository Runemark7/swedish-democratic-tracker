---
id: utskott-utgiftsomrade
name: Utskottens utgiftsområden (riksdagsordningen)
kind: seed
upstream: https://www.riksdagen.se/sv/dokument-och-lagar/dokument/svensk-forfattningssamling/riksdagsordning-2014801_sfs-2014-801/
license: Svensk författningssamling — fri användning
freshness: ändras genom lagändring; kontrolleras vid ny lydelse
last_verified: 2026-08-05
verification_status: verified
verification_notes: ~
used_by:
  - /committees (utgiftsområden per utskott)
  - /committees/:code (KOSTAR-raden)
---

## Vad det är
Vilket utskott som bereder vilket utgiftsområde (UO) är inte vår tolkning —
det är lag. Fördelningen står i Bilagan (tilläggsbestämmelse 7.5.1) till
riksdagsordningen (2014:801), den författning som reglerar riksdagens
arbetsformer. Det är **inte** hämtat från riksdagen.se:s sidor om de
enskilda utskotten, som beskriver ansvarsområden i löptext utan att lista
en maskinläsbar koppling till UO-koder.

Fördelningen är en strikt partition: alla 27 utgiftsområden har exakt ett
utskott, och varje utskott som förekommer i bilagan har minst ett
utgiftsområde. Inget UO saknas och inget UO har mer än en huvudman.

## Hur du själv kommer åt datan
Riksdagsordningen med samtliga bilagor och tilläggsbestämmelser publiceras
i Svensk författningssamling och finns i fulltext på riksdagen.se:

<https://www.riksdagen.se/sv/dokument-och-lagar/dokument/svensk-forfattningssamling/riksdagsordning-2014801_sfs-2014-801/>

Sök efter tilläggsbestämmelse 7.5.1 — det är tabellen som listar varje
utskott och de utgiftsområden det bereder. Vid en lagändring publiceras en
ny SFS-nummer-version; senaste kända ändring av bilagan är SFS 2026:1349,
som träder i kraft 2026-09-01. Den ändringen påverkar inte den fördelning
som används här — se "Begränsningar" nedan.

## Schema/fält vi använder
- `utskott_code` — utskottets kod (t.ex. `FiU`, `SfU`), samma kodrymd som
  övriga utskottstabeller i databasen.
- `uo_code` — utgiftsområdets kod i formen `UO1`–`UO27`, aldrig namnet.
- `in_force_from` — datum då raden blev gällande. Fördelningen versioneras
  per ikraftträdandedatum, inte per riksmöte: attribueringen av ett
  historiskt beslut är permanent (beteckningen kodar redan vilket utskott
  som avgav betänkandet), så en gammal omröstning behöver aldrig
  omkodas. Bara beskrivningen av vad ett utskott bereder *idag* kan ändras
  över tid, och det är det denna kolumn fångar.

## Begränsningar och kända problem
- **Kodad på UO-kod, aldrig på namn.** UO13 ("Jämställdhet och nyanlända
  invandrares etablering") och UO20 ("Allmän miljö- och naturvård") har
  båda bytt namn i förhållande till äldre riksdagstryck utan att
  huvudmannaskapet ändrats. En koppling på namn skulle tyst tappa dessa
  rader vid en namnändring; kopplingen görs därför alltid på `uo_code` mot
  `expenditure_areas.code`.
- SFS 2026:1349 ändrar bilagan med verkan från 2026-09-01. Den ändringen
  rör inte någon av de 27 UO-utskott-kopplingarna som redan finns i denna
  tabell — den påverkar alltså ingenting som sajten visar idag. Om en
  framtida lagändring faktiskt omfördelar ett UO läggs en ny rad till med
  ett senare `in_force_from`, den gamla raden behålls oförändrad.
- Detta är en engångs-seed, inte ett API. Den uppdateras manuellt när
  bilagan ändras i sak, inte på ett fast schema.

## Hur vi bearbetar
Fördelningen är avskriven direkt ur bilagans tabell och lagd i en egen
databastabell, en rad per utgiftsområde. Ingen bearbetning eller
härledning sker — varje rad speglar exakt vad lagtexten säger om vilket
utskott som bereder vilket UO.
