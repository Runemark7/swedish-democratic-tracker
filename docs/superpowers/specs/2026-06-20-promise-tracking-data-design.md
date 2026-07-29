# Promise tracking — Sub-project A: Curated party-goal data expansion

**Date:** 2026-06-20
**Branch:** `worktree-feat+promise-tracking-data`
**Status:** Design — awaiting user review

## Context

Riksdagskollen tracks Swedish parties' stated promises against their actual
Riksdagen voting records. Two promise pipelines exist in the codebase:

- **Party goals** (`party_goals` table) — hand-curated election-manifesto
  promises with `keywords` + `relevant_committees`. The `keyword-matcher`
  ingestion worker links each goal to relevant votes and the scorecard
  refresh computes alignment. **This pipeline works and has data.**
- **Politician promises** (`promises` table) — meant to be AI-extracted from
  speeches. The `AIService` is a stub (`StubAIService`); the table is empty.
  CLAUDE.md forbids wiring real AI until explicitly asked.

This sub-project expands the **working, no-AI party-goal pipeline**. It is the
foundation (data) for a later Sub-project B (new feature surface / UI).

Current state: ~44 goals total, 5–6 per party, **all `directional`**, all
sourced from 2022 valmanifest / Tidöavtalet. The existing data-source doc
(`docs/data-sources/seed-party-goals.md`) promises users can "verifiera
ordagrant" via a `source_citation` with page numbers — but no such field
exists in the table (`source_document` is a generic string like
`'Valmanifest 2022'`). The verifiability the doc claims is not backed by data.

## Core-principle guardrail

CLAUDE.md mandates strict separation of FACT from INTERPRETATION, and warns
specifically against verdict framing ("broke their promise"). Therefore:

- **No** "kept" / "broken" verdict labels anywhere.
- Goals are presented as: promise text + source (URL + verbatim quote) +
  the relevant votes + neutral alignment count. The reader draws the
  conclusion. (Verdict reframing was agreed with the user.)

This sub-project only adds data + the source fields that make neutrality
*verifiable*. Display work lives in Sub-project B.

## Scope

In scope:
1. Schema: add `source_url` + `source_quote` to `party_goals` (nullable).
2. Expand curated goals to ~12/party (~96 total): mix of `concrete`
   (manifesto states a number/date) and `directional`; no pure `rhetorical`.
   Broaden committee coverage beyond the current SoU/JuU/FiU cluster.
3. Expose `source_url` / `source_quote` through the goals domain → repository
   → API (additive, optional) so the new data is reachable, not dark.
4. Data-source discipline (CLAUDE.md rule 11): fix + update
   `seed-party-goals.md`, regenerate SourceRegistry, verify `/data` pages.

Out of scope (→ Sub-project B):
- Any new UI page, comparison view, filter, or alignment-summary rendering.
- Wiring real AI / politician-speech promise extraction.
- 2026 goals for parties whose 2026 manifestos are not yet published (only
  S has published a 2026 valplattform as of 2026-06-20).

## Data model

Reuse `party_goals`. One additive migration:

- **`000012_party_goals_source_fields.up.sql`**
  - `ALTER TABLE party_goals ADD COLUMN source_url TEXT;`
  - `ALTER TABLE party_goals ADD COLUMN source_quote TEXT;`
  - Both nullable → existing rows unaffected.
  - `.down.sql` drops both columns.

No change to the matcher: `keyword-matcher` links goals to votes by
`relevant_committees` + `keywords` only. `specificity` remains a display hint.
`concrete` goals match mechanically the same way `directional` ones do — so
`concrete` goals still need precise keywords + correct committees.

## Content plan

- **`000013_seed_party_goals_expansion.up.sql`** — additive seed, idempotent
  via existing `ON CONFLICT (party, goal_text) DO NOTHING` (migration 000002).
- ~12 goals/party (up from 5–6); ~96 total.
- Each new row sets: `party`, `goal_text`, `topic`, `specificity`,
  `source_document`, `keywords`, `relevant_committees`, `source_url`,
  `source_quote` (verbatim manifesto excerpt).
- `specificity`: `concrete` when the manifesto states a measurable target
  (number, date, kr); otherwise `directional`. No pure `rhetorical`.
- Committee coverage broadened where parties have stated positions, e.g.
  migration (SfU), EU/foreign (UU), transport (TU), culture (KrU),
  environment (MJU), education (UbU), finance/tax (FiU/SkU), defence (FöU),
  justice (JuU), labour (AU), health (SoU), housing (CU), business (NU).
- Sources: 2022 valmanifest + Tidöavtalet (M/KD/L/SD govt cooperation) for
  existing-era goals; S 2026 valplattform where applicable. Every row's
  `source_url` points at the party's own published document; `source_quote`
  is the exact wording the goal is derived from.
- **Quality bar:** keywords must be specific enough to match real Riksdagen
  vote subjects. Vague keywords create false matches and pollute scorecards.

## API exposure (additive)

So the new fields are reachable for Sub-project B and verifiable now:

- `goals/domain.Goal`: add `SourceURL string` (`sourceUrl,omitempty`) and
  `SourceQuote string` (`sourceQuote,omitempty`).
- `goals` postgres repository: add both columns to the SELECT(s) that load
  goals.
- `api/openapi.yaml`: add `sourceUrl` + `sourceQuote` (optional) to the goal
  schema; regenerate `frontend/src/shared/api-contract.ts` via
  `npm run generate:api`. Do not hand-edit the generated file.
- No frontend rendering of these fields in this sub-project (that's B).

## Data-source discipline (rule 11)

- Update `docs/data-sources/seed-party-goals.md`:
  - Correct the "Schema/fält" section to the actual columns
    (`goal_text`, not `title`/`description`; `source_document`), and document
    the new `source_url` / `source_quote`.
  - Note the now-backed verifiability (URL + verbatim quote per goal).
  - Bump `last_verified` to 2026-06-20; note 2026 partial coverage.
- Run `cd frontend && npm run sync:data-sources` to regenerate
  `SourceRegistry.generated.ts`.
- Verify `/data` lists the source and `/data/seed-party-goals` renders.

## Verification

1. `go build -C backend ./...` — clean.
2. Run migrations (`make migrate` or backend startup) — 000012 + 000013 apply.
3. Query DB: confirm ~96 goals, presence of `concrete` rows, non-null
   `source_url`/`source_quote` on new rows, broadened committee set.
4. Run / confirm `keyword-matcher` creates matches for new goals and
   scorecard refresh succeeds (no errors).
5. `GET /api/parties/:party/goals` returns new goals incl. source fields.
6. `cd frontend && npm run generate:api && npm run typecheck` — clean.
7. `npm run sync:data-sources`; `/data/seed-party-goals` renders updated body.

## Risks

- **Keyword quality** is the dominant risk: bad keywords → false vote matches
  → misleading (though still source-linked) scorecards. Mitigate with a
  precise keyword bar and committee correctness during curation.
- **Source accuracy**: `source_quote` must be verbatim and `source_url` live.
  Curation must cite real published manifesto text, never paraphrase as quote.
- **2026 coverage is thin** (only S published) — accepted; structured for
  incremental future adds.
