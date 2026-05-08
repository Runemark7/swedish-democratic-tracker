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
Concrete reproduction steps users can follow. Pick the matching pattern:

**For APIs:** include a `curl` example and a brief shape sketch:

```bash
curl -s 'https://api.example.com/foo?bar=1'
```

**For CSV/ZIP:** explicit download URL + extraction commands:

```bash
curl -O 'https://example.com/data.zip'
unzip data.zip
```

**For seed data:** describe what the source documents are (PDF, web
page, etc.) and where the upstream lives. Do not reference internal
repo paths — write so a non-developer can follow.

**For synthesized data:** describe the derivation rule in plain
language. Do not reference function names or file paths.

## Schema/fält vi använder
- `field_one` — what it means and how we use it
- `field_two` — ...

## Begränsningar och kända problem
- Real gotchas users should know (e.g. "Gotland uses code 0980L, not 09L").
- Dates / years that are still preliminary upstream.

## Hur vi bearbetar
Transformations between upstream and what users see, written for an
end-user audience. Describe the data flow without referencing internal
file paths, function names, or commit hashes — those belong in the
codebase, not on a public methodology page.
