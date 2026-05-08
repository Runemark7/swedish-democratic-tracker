---
id: ted
name: TED — Tenders Electronic Daily
kind: api
upstream: https://ted.europa.eu/
license: EU public data
freshness: löpande
last_verified: 2026-05-08
used_by:
  - /kommun/:code (kommunal upphandling — om endpointen är wired)
---

## Vad det är
EU:s officiella databas över alla offentliga upphandlingar över
tröskelvärdet. Vi använder den för att visa kommunala kontrakt på
kommunsidor.

## Hur du själv kommer åt datan
TED:s sökgränssnitt är publikt och kräver ingen inloggning.

```bash
# Sök upphandlingar för en organisation (kommun)
curl -s 'https://ted.europa.eu/api/v3.0/notices/search?q=BUYER-NAME%3D%22Stockholms%20kommun%22'
```

Full API-referens: <https://docs.ted.europa.eu/api/index.html>.
Webbgränssnitt: <https://ted.europa.eu/>.

## Schema/fält vi använder
- `notices[].id` — TED-ID.
- `notices[].title` — kontraktstitel.
- `notices[].buyer.name` — köpande organisation.
- `notices[].publication-date` — publiceringsdatum.

## Begränsningar och kända problem
- TED är skrivet för EU-omfattande upphandling, vilket innebär att
  endast kontrakt över tröskelvärdet (~1,5 mkr) syns. Mindre kommunala
  kontrakt fångas inte.
- Sökningen är text-baserad på köparens namn; namnvarianter (t.ex.
  "Stockholms kommun" vs "Stockholms stad") kan missas.

## Hur vi bearbetar
Vår backend hämtar resultaten från TED via en proxy-endpoint som
filtrerar på den valda kommunens namn, och returnerar listan till
frontend för rendering.
