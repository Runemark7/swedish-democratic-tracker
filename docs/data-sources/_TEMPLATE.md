---
id: SOURCE_ID
name: Display name (Swedish)
kind: api          # api | csv | seed | synthesized
upstream: https://example.com   # set to ~ if no public upstream URL
license: License terms (e.g. "PxWeb open data — fri användning")
freshness: How often data refreshes (e.g. "daglig", "4-årscykel", "engångs")
last_verified: 2026-05-08
used_by:
  - /route/path (what value this source backs)
---

## Vad det är
One paragraph describing what this source is and why it exists upstream.

## Hur du själv kommer åt datan
Concrete reproduction steps. Pick the matching pattern:

**For APIs:** include a `curl` example and the response shape:

```bash
curl -s 'https://api.example.com/foo?bar=1'
```

**For CSV/ZIP:** explicit download URL + extraction commands:

```bash
curl -O 'https://example.com/data.zip'
unzip data.zip
```

**For seed migrations:** name the migration file and where citations live:

```
backend/migrations/000003_seed_party_goals.sql — citations in commit
60ec51a and original PDFs at <upstream URL>
```

**For synthesized data:** describe the derivation rule and link the code.

## Schema/fält vi använder
- `field_one` — what it means and how we use it
- `field_two` — ...

## Begränsningar och kända problem
- Real gotchas users should know (e.g. "Gotland uses code 0980L, not 09L").
- Dates / years that are still preliminary upstream.

## Hur vi bearbetar
Transformations applied between upstream and what users see. Reference the
backend file or hook that does the transform (e.g.
`backend/internal/regions/adapters/scb/client.go:44`).
