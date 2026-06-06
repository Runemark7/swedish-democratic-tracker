---
id: ted
name: TED — Tenders Electronic Daily
kind: api
upstream: https://api.ted.europa.eu/v3/notices/search
license: EU public data
freshness: löpande
last_verified: 2026-06-06
verification_status: unsure
verification_notes: "Upstream-URL i MD pekade på ted.europa.eu (webbgränssnitt) men backenden anropar api.ted.europa.eu/v3/notices/search; MD uppdaterat till korrekt API-host. Tidigare bash-block med curl mot fel host borttaget. Inga SourceMarker-konsumenter finns i UI idag."
used_by:
  - /kommun/:code (kommunal upphandling — om endpointen är wired)
---

## Vad det är
EU:s officiella databas över alla offentliga upphandlingar över
tröskelvärdet. Vi använder den för att visa kommunala kontrakt på
kommunsidor.

## Hur du själv kommer åt datan
TED:s publika webbgränssnitt finns på <https://ted.europa.eu/> och
kräver ingen inloggning. Du kan söka på köparens namn, CPV-kod
(produktkategori) eller land direkt i sökformuläret.

För programmatisk åtkomst tillhandahåller EU ett öppet API på
<https://api.ted.europa.eu/v3/notices/search>. API-dokumentationen
finns på <https://docs.ted.europa.eu/api/index.html>. Endpointen tar
en POST med ett JSON-sökuttryck och returnerar en lista med
upphandlingsbeslut. Ingen API-nyckel krävs för grundläggande sökning.

## Schema/fält vi använder
- `notices[].publication-date` — publiceringsdatum.
- `notices[].classification-cpv` — CPV-kod (produktkategori).
- `notices[].total-value` — kontraktsvärde.
- `notices[].total-value-cur` — valutakod.

## Begränsningar och kända problem
- TED är skrivet för EU-omfattande upphandling, vilket innebär att
  endast kontrakt över tröskelvärdet (~1,5 mkr) syns. Mindre kommunala
  kontrakt fångas inte.
- Sökningen är text-baserad på köparens namn; namnvarianter (t.ex.
  "Stockholms kommun" vs "Stockholms stad") kan missas.

## Hur vi bearbetar
Vår backend hämtar resultaten från TED API via en proxy-endpoint som
filtrerar på den valda kommunens namn, och returnerar listan till
frontend för rendering.
