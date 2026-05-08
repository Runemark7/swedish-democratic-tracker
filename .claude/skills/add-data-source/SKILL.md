---
name: add-data-source
description: Use when adding a new external HTTP client, CSV downloader, or seed migration. Walks through registering a data source so every UI value stays traceable. Triggers on "add a data source", "new external client", "fetching from X", "ingesting CSV", "new seed migration".
---

# Add a Data Source

Use this skill when the next task involves wiring a new data source into the
project — a new external HTTP client, a CSV/ZIP download script, or a static
seed migration. The goal is that every UI value remains traceable to a
documented source. See
`docs/superpowers/specs/2026-05-08-data-provenance-system-design.md` for the
underlying design.

## Checklist

You MUST complete these steps in order. Do not close the task until all are
done.

1. **Pick an `id`.** Use a stable, kebab-case identifier (e.g. `scb-kfmandat`,
   `statskontoret-arsutfall`). Match the filename and the frontmatter `id`
   field exactly.

2. **Update the mermaid diagram in `CLAUDE.md`.** Find the `ext` subgraph and
   add the new external service. Add the service→ext edge if it's a runtime
   data fetch.

3. **Copy the template.**

   ```bash
   cp docs/data-sources/_TEMPLATE.md docs/data-sources/<id>.md
   ```

   Fill in every frontmatter field. CSV/ZIP sources MUST include the download
   URL and extraction commands in the "Hur du själv kommer åt datan" section.

4. **Add an entry to `docs/data-sources/INDEX.md`.** One row in the table.

5. **Sync the registry.**

   ```bash
   cd frontend && npm run sync:data-sources
   ```

   This regenerates `frontend/src/components/sources/SourceRegistry.generated.ts`.
   The script will fail loudly if frontmatter is malformed.

6. **Tag every UI value.** For each component that renders a value sourced
   from this data:

   - Add `<SourceMarker sourceId="<id>" />` next to the value (inside the same
     containing element so it stays inline).
   - Add or update a `<SectionSource sourceIds={[...]}>` at the bottom of the
     containing card.

7. **Verify.**

   - `cd frontend && npx tsc --noEmit` clean.
   - `cd frontend && npm run sync:data-sources` idempotent (no diff on second
     run).
   - Visit `/data/<id>` locally — frontmatter card and prose render.
   - Click a `<SourceMarker>` for the new source — popover opens with the
     correct name and links to `/data/<id>`.

8. **Commit message** — use a `feat(sources):` prefix and reference the source
   ID:

   ```
   feat(sources): wire <source-name> (<id>)

   - Add docs/data-sources/<id>.md
   - Update mermaid in CLAUDE.md
   - Tag UI values on <pages>
   ```

## Common pitfalls

- **Frontmatter `id` doesn't match filename.** The sync script enforces this;
  fix the frontmatter, not the filename, since other code references the
  filename.
- **Forgetting `sync:data-sources`.** The MD file alone is invisible to the
  UI — the registry must be regenerated.
- **CSV/ZIP source without reproduction steps.** A user reading
  `/data/<id>` must be able to download and inspect the data themselves.
  An external link to the publisher's homepage is not enough.
- **Tagging only the section but not the values.** Per-value markers and
  per-section receipts serve different purposes — both are required.
