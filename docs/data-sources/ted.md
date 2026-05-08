---
id: ted
name: TED — Tenders Electronic Daily
kind: api
upstream: https://ted.europa.eu/api/v3.0
license: EU public data
freshness: löpande
last_verified: 2026-05-08
used_by:
  - /kommun/:code (kommunal upphandling — om endpointen är wired)
---

## Vad det är
EU:s officiella databas över alla offentliga upphandlingar över
tröskelvärdet. Vi använder den för kommunala kontrakt.

## Hur du själv kommer åt datan
```bash
# Sök upphandlingar för en organisation (kommun)
curl -s 'https://ted.europa.eu/api/v3.0/notices/search?q=BUYER-NAME%3D%22Stockholms%20kommun%22'
```

Full referens: <https://ted.europa.eu/TED/misc/aboutTedApi.do>.

## Schema/fält vi använder
- `notices[].id` — TED-ID.
- `notices[].title` — kontraktstitel.
- `notices[].buyer.name` — köpande organisation.
- `notices[].publication-date` — publiceringsdatum.

## Begränsningar och kända problem
- TED är skrivet för EU-omfattande upphandling, vilket innebär att
  endast kontrakt över tröskelvärdet (~1,5 mkr) syns. Mindre kommunala
  kontrakt fångas inte.

## Hur vi bearbetar
- `backend/internal/regions/adapters/http/handler.go:164`
  (`getMunicipalityProcurement`) proxar via en TED-klient.
