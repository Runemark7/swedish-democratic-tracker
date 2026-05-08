# Data Provenance System

**Date:** 2026-05-08
**Status:** Design (approved by user, pending implementation plan)

## Context

The whole reason this site exists is to aggregate Swedish government data
from many sources (Riksdagen, SCB, Kolada, Statskontoret, TED, etc.) and
present it in one place. Without rigorous source attribution the project
loses its purpose. Right now coverage is uneven: a few cards (the
myndigheter "receipt" on the Riksdag page, KPI items via `GoalBadge`'s
`sourceUrl`) carry attribution, but most values render bare. CSV-based
sources (Statskontoret årsutfall is a ZIP with CSVs inside) are the
worst case — a user who wants to verify a number has no way to do so
beyond the upstream homepage.

Goal: every data point on the site is traceable to a documented source,
and adding a new source becomes a disciplined process so the system
stays trustworthy as it grows. Existing sources get backfilled. Future
sources get a forced template + checklist + skill.

## Out of scope

- Last-fetched-at timestamps wired from the backend (defer — v1 uses
  static `last_verified` in frontmatter, updated by hand).
- Source-schema diffing over time (e.g. SCB renamed a column in 2024).
- A public `GET /api/data-sources` endpoint (frontend bundles work for v1).
- Caching / incremental rebuild optimisations for the markdown sync step.
- Changes to the upstream services themselves.

## Architecture

Three layers stacked on a single source of truth:

1. **`docs/data-sources/` (canonical)** — one Markdown file per source with
   YAML frontmatter and prose sections. `INDEX.md` is a short table of
   contents. `_TEMPLATE.md` is the skeleton.
2. **Registry + UI components** — `frontend/src/components/sources/`
   exposes a typed `SourceRegistry` mirroring the frontmatter, plus
   `<SourceMarker>` (per-value glyph + popover) and `<SectionSource>`
   (bottom-of-card line listing sources for the section).
3. **Methodology pages** — `/data` lists every entry, `/data/:id` renders
   the full Markdown for that source. Internal popover links route here.

Governance lives outside the runtime layers:

- `CLAUDE.md` gets a "Data Source Discipline" rule.
- `.claude/skills/add-data-source/SKILL.md` is a project-local skill
  loaded on demand to walk a developer through registering a new source.
- `docs/data-sources/_TEMPLATE.md` is the copy-this skeleton.

## Components

### `docs/data-sources/<id>.md`

YAML frontmatter is mandatory. Frontmatter shape:

```yaml
---
id: scb-kfmandat
name: SCB Kfmandat (kommunal mandatfördelning)
kind: api          # api | csv | seed | synthesized
upstream: https://api.scb.se/OV0104/v1/doris/sv/ssd/ME/ME0104/ME0104A/Kfmandat
license: PxWeb open data — fri användning
freshness: 4-årscykel (val)
last_verified: 2026-05-08
used_by:
  - /kommun/:code (mandat)
  - /region/:code (mandat — via Ltmandat sibling)
---
```

Mandatory prose sections (Swedish headings):

1. `## Vad det är` — one paragraph.
2. `## Hur du själv kommer åt datan` — concrete reproduction steps.
   APIs: a `curl` example. CSV/ZIP: download URL + extraction commands.
   Seeds: which migration file + where the citations live.
3. `## Schema/fält vi använder` — the fields we read.
4. `## Begränsningar och kända problem` — gotchas (e.g. Gotland → 0980L,
   late-publishing years return ".." in SCB).
5. `## Hur vi bearbetar` — transformations applied between upstream and
   what the user sees.

### Initial source list (13 entries)

| ID | Kind | Backs |
|---|---|---|
| `riksdagen` | api | politicians, speeches, votes, agenda, beslut feed |
| `kolada` | api | region/kommun KPIs |
| `kolada-spending` | api | kommun spending breakdown (subset of kolada) |
| `scb-kfmandat` | api | kommun election mandates |
| `scb-ltmandat` | api | region election mandates |
| `scb-kostndrlt` | api | region budget (KostnDRLT) |
| `scb-befolkning` | api | population |
| `ted` | api | EU procurement |
| `statskontoret-arsutfall` | csv | agency expenditure (ZIP/CSV) |
| `seed-party-goals` | seed | migration `000003_seed_party_goals.sql` |
| `seed-budget-data` | seed | migrations `000005`–`000009` |
| `derived-agenda` | synthesized | `generateAgenda()` in `useDemocracy.ts` |
| `wikimedia-svg` | seed | Sweden region/kommun SVG paths bundled from Wikimedia Commons (CC BY-SA 2.5) |

If `kolada` and `kolada-spending` end up with identical metadata they
collapse into one file with multiple `used_by` entries. Decision deferred
to implementation — the audit during MD-writing settles it.

### `frontend/src/components/sources/SourceRegistry.ts`

Typed mirror of the frontmatter so call sites get autocomplete on `sourceId`:

```ts
export type SourceKind = "api" | "csv" | "seed" | "synthesized";

export interface SourceEntry {
  id: string;
  name: string;
  kind: SourceKind;
  upstream?: string;
  license?: string;
  freshness?: string;
  lastVerified: string;
  blurb: string; // first paragraph of "Vad det är"
}

export const SOURCES: Record<string, SourceEntry> = { /* ... */ };
export type SourceId = keyof typeof SOURCES;
```

Populated at build time by a sync script that parses every
`docs/data-sources/*.md`. The script writes `SourceRegistry.generated.ts`
and is kept honest by a `// AUTOGENERATED — do not edit by hand` banner.

### `frontend/src/components/sources/SourceMarker.tsx`

Inline glyph next to a value:

```tsx
<span>
  68 mandat
  <SourceMarker sourceId="scb-kfmandat" />
</span>
```

Glyph by `kind`:

- `api` → `ⓘ`
- `csv` → `⤓`
- `seed` → `✎`
- `synthesized` → `≈`

Default styling: 11 px, `var(--color-fg-muted)`, 4 px left margin,
keyboard-focusable button. Tap or hover opens `<SourcePopover>`.

### `frontend/src/components/sources/SourcePopover.tsx`

Anchored under the marker, max-width 280 px. Closes on outside-click and
Escape. Content:

```
Name (serif 14)
Kind chip · "senast verifierad <date>" (mono 10 muted)
─────
Blurb (body 12)
─────
→ Läs hur vi använder datan          (internal Link to /data/:id)
↗ Öppna upstream-källan              (only if upstream != null; new tab)
```

### `frontend/src/components/sources/SectionSource.tsx`

Bottom-of-card line. Replaces hand-rolled `Källa: ...` strings:

```tsx
<SectionSource sourceIds={["scb-kfmandat", "scb-befolkning"]} />
```

Renders one line: `Källor: SCB Kfmandat · SCB Befolkning ↗`. Each name
links to the in-app methodology page.

### `frontend/src/features/data/DataIndexPage.tsx`

Route `/data`. Lists every entry from `SOURCES`, grouped by kind with
filter pills (Alla / API / CSV / Seed / Härledd). One row per source:
name (serif 18) · kind chip · blurb · "Använd på X sidor" with a linkable
list of routes from `used_by`.

### `frontend/src/features/data/DataSourcePage.tsx`

Route `/data/:id`. Renders the source's MD file. Header card shows the
frontmatter (name, kind chip, upstream link, license, freshness,
last_verified, used-by routes). Below: the parsed prose body via
`react-markdown` + `remark-gfm`.

Markdown delivery: build-time bundling. Vite imports the file via
`?raw`:

```ts
import md from "@/data-sources/scb-kfmandat.md?raw";
```

A small npm script `npm run sync:data-sources` copies (or symlinks)
`docs/data-sources/*.md` to `frontend/src/data-sources/` before build.
Runs as part of `npm run build` and `npm run dev`.

### Footer link

A new "Datakällor" link in the footer (not main nav) routes to `/data`.
Site footer doesn't exist yet at the level needed; add a minimal one in
`App.tsx` below `<main>`. Visible on all pages.

## Governance

### CLAUDE.md addition

New section under "Agent Directives: Mechanical Overrides":

```md
### Data Source Discipline

Whenever you add, change, or remove an external data source (HTTP API,
CSV/ZIP download, scraped feed) or a static seed migration:

1. Update `CLAUDE.md`'s mermaid architecture diagram (the `ext` subgraph
   and the service→ext edges).
2. Add or update a markdown file in `docs/data-sources/<id>.md` using
   `docs/data-sources/_TEMPLATE.md`. Frontmatter is mandatory; prose
   sections are mandatory. CSV/ZIP sources MUST include reproducible
   download steps.
3. Register the source in
   `frontend/src/components/sources/SourceRegistry.generated.ts` via
   `npm run sync:data-sources`.
4. Tag every UI value sourced from this data with
   `<SourceMarker sourceId="...">` and the surrounding card with
   `<SectionSource sourceIds={[...]}>`.
5. Run `npm run sync:data-sources` so the build picks up the new MD file.

Closing the task without these steps is incomplete.
```

### `docs/data-sources/_TEMPLATE.md`

Copy-this skeleton with placeholder text in every section.

### `.claude/skills/add-data-source/SKILL.md`

Project-local skill (lives in repo, travels with checkout). Trigger
description targets phrases like "add a data source", "new external
client", "fetching from X". Body mirrors the CLAUDE.md checklist with
exact paths and an example commit message.

## UI audit

`<SourceMarker>` placement (per-value), by route:

| Route | Values + sources |
|---|---|
| `/` | Beslut idag rows → `riksdagen`. Veckans omröstningar → `riksdagen`. Aktuella debatter → `riksdagen`. Kommande beslut → `derived-agenda`. Vad partierna säger cards → `riksdagen`. |
| `/riksdag` | KPI strip → per-KPI source (mostly seeds). Hemicycle total + seats → `riksdagen-government` (or `riksdagen` if it's API-derived; resolved during audit). Coalition type → `riksdagen-government`. Budget total + areas → `seed-budget-data`. Live votes → `riksdagen`. Agenda → `derived-agenda`. Myndigheter rows + donut + dual-line → `statskontoret-arsutfall`. |
| `/region/:code` | KPI strip → `kolada`. Hemicycle + seats → `scb-ltmandat`. Coalition type → `scb-ltmandat` (derivation rule documented in that file's "Hur vi bearbetar" section). Budget → `scb-kostndrlt`. Riksdag relevanta beslut → `riksdagen`. Agenda → `derived-agenda`. Kommun map → `wikimedia-svg` for paths, `scb-kfmandat` for clickable kommun list. |
| `/kommun/:code` | KPI strip → `kolada`. Hemicycle + seats → `scb-kfmandat`. Coalition type → `scb-kfmandat` (derivation rule documented in that file). Budget → `kolada-spending`. Riksdag relevanta beslut → `riksdagen`. Agenda → `derived-agenda`. |
| `/budget`, `/budget/areas/:code` | Year list, allocations → `seed-budget-data`. |
| `/parties`, `/parties/:party/goals`, `/parties/:party/goals/:goalId/votes` | Scorecard, goals → `seed-party-goals`. Vote alignment → `riksdagen`. |
| `/politicians`, `/politicians/:id` | Profile, votes, speeches → `riksdagen`. |
| `/votes/:beteckning/:punkt` | Tally + party split → `riksdagen`. |
| `/region`, `/kommun` (landing maps) | Map paths → `wikimedia-svg`. Card stats → respective sources. |
| `/anforanden/:id` | Speech → `riksdagen`. |

`<SectionSource>` placement (per-card): every `.sdt-card`-style card on
the routes above gets a bottom-of-card line summarising its sources.
Existing hand-rolled "Källa: ..." strings are removed in favour of the
new component.

## Data flow

1. Developer writes `docs/data-sources/<id>.md` with frontmatter.
2. `npm run sync:data-sources` parses every file, copies them to
   `frontend/src/data-sources/`, and writes `SourceRegistry.generated.ts`.
3. Vite bundles MD files via `?raw` imports.
4. UI components (`SourceMarker`, `SectionSource`) read from
   `SOURCES[id]` for popover content and link targets.
5. `/data` and `/data/:id` use `SOURCES` for the index and parse the raw
   MD body via `react-markdown` for the detail page.

## Error handling

- Unknown `sourceId` passed to `<SourceMarker>` or `<SectionSource>`:
  TypeScript catches at compile time (`SourceId` is a literal union from
  the generated registry). At runtime, defensively render a muted "—"
  glyph instead of crashing.
- Missing MD file at `/data/:id`: render a 404-style empty state with a
  link back to `/data`.
- `sync:data-sources` failure (malformed frontmatter): script exits
  non-zero, build fails, with a clear pointer to the offending file.

## Testing

- `npm run sync:data-sources` is idempotent. Run twice, no diff.
- Build succeeds: `cd frontend && npm run build`.
- Type-check clean: `cd frontend && npx tsc --noEmit`.
- Visual smoke: visit `/data`, `/data/scb-kfmandat`, click an `<SourceMarker>`
  on `/`, confirm popover renders + internal link works.
- Mobile (375 × 812): popover doesn't overflow viewport, marker glyph is
  tap-friendly (≥ 24 × 24 hit area).
- A grep across the repo for `Källa: ` (hand-rolled strings) returns
  zero matches outside the `docs/data-sources/` MD files after the
  refactor.
- Every entry in the UI audit table above has at least one
  `<SourceMarker>` or `<SectionSource>` reference (verified by a small
  audit script that greps for sourceIds and cross-references the audit).

## Open questions for implementation phase

- `riksdagen-government` vs `riksdagen` — same source or distinct?
  Likely the government composition currently comes from a hand-curated
  helper, not the open API. Resolve during MD writing for `/riksdag`.
- `kolada` vs `kolada-spending` — single MD file with multi-`used_by` or
  two files? The endpoints differ. Lean toward one file with both
  endpoints documented in the prose.
- Footer placement — the project does not have a global footer today.
  Adding one is in scope (small, single Datakällor link). If the design
  team objects, the link can move into the main nav as a low-emphasis
  trailing item.
