# Party-Goal Data Expansion (Promise Tracking A) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Roughly double curated `party_goals` coverage (~12/party, ~96 total), add `concrete` measurable promises, and back every goal with a verifiable source URL + verbatim quote — no AI.

**Architecture:** Reuse the existing working party-goal pipeline (curated `party_goals` → `keyword-matcher` worker → `party_scorecards` view → scorecards). Add two nullable source columns, expose them through domain→repo→handler→OpenAPI (additive), then add curated rows via an idempotent seed migration. No UI in this plan (that is Sub-project B).

**Tech Stack:** Go 1.26 (chi, pgx/v5), PostgreSQL 17, golang-migrate, OpenAPI → openapi-typescript, React/Vite frontend (typecheck only here).

## Global Constraints

- Go 1.26.1, Node 25.9.0 — exact versions.
- Raw parameterized SQL only — no ORM, never string-interpolate values.
- All migrations numbered; next free numbers are `000026` and `000027`.
- Seed migration MUST be idempotent: `ON CONFLICT (party, goal_text) DO NOTHING` (unique constraint from migration `000002`).
- `frontend/src/shared/api-contract.ts` is generated — never hand-edit; regenerate via `npm run generate:api`.
- FACT/INTERPRETATION separation: no "kept"/"broken" verdict copy or fields. Only neutral source + alignment data.
- `source_quote` MUST be verbatim text from the party's real published manifesto/platform; `source_url` MUST be a live URL to that document. **Never fabricate or paraphrase a quote.** If a verbatim quote for a goal cannot be located, do not invent the goal.
- No test harness exists in this repo. Verification gates are: `go build -C backend ./...`, migrations applying cleanly, DB row queries, HTTP responses, `npm run typecheck`, and `/data` page render — not unit tests.
- Verify in Docker per project convention where a running stack is needed (`make run` / `make migrate`).

---

## File Structure

- `backend/migrations/000026_party_goals_source_fields.{up,down}.sql` — schema: add `source_url`, `source_quote`.
- `backend/internal/goals/domain/goal.go` — add `SourceURL`, `SourceQuote` fields.
- `backend/internal/goals/adapters/postgres/repository.go` — add columns to `selectCols`, `Create`, `scanGoal`.
- `backend/internal/goals/adapters/http/handler.go` — add fields to `goalWithAlignment` + mapping.
- `api/openapi.yaml` — add `sourceUrl`, `sourceQuote` to `Goal` schema.
- `frontend/src/shared/api-contract.ts` — regenerated (do not hand-edit).
- `backend/migrations/000027_seed_party_goals_expansion.{up,down}.sql` — curated rows.
- `docs/data-sources/seed-party-goals.md` — corrected + updated source doc.
- `frontend/src/components/sources/SourceRegistry.generated.ts` — regenerated.

---

## Task 1: Schema — add source columns

**Files:**
- Create: `backend/migrations/000026_party_goals_source_fields.up.sql`
- Create: `backend/migrations/000026_party_goals_source_fields.down.sql`

**Interfaces:**
- Produces: nullable columns `party_goals.source_url TEXT`, `party_goals.source_quote TEXT`.

- [ ] **Step 1: Write the up migration**

Create `backend/migrations/000026_party_goals_source_fields.up.sql`:

```sql
-- Add verifiable source provenance to party goals.
-- Both nullable so existing rows are unaffected.
ALTER TABLE party_goals
  ADD COLUMN source_url   TEXT,
  ADD COLUMN source_quote TEXT;
```

- [ ] **Step 2: Write the down migration**

Create `backend/migrations/000026_party_goals_source_fields.down.sql`:

```sql
ALTER TABLE party_goals
  DROP COLUMN IF EXISTS source_quote,
  DROP COLUMN IF EXISTS source_url;
```

- [ ] **Step 3: Apply migration against local DB**

Run: `make migrate`
Expected: migrate reports version `26` applied, no error.

- [ ] **Step 4: Verify columns exist**

Run:
```bash
docker compose -f docker-compose.dev.yml exec -T postgres \
  psql -U riksdagskollen -d riksdagskollen \
  -c "\d party_goals" 2>/dev/null | grep -E "source_url|source_quote"
```
Expected: two lines showing `source_url | text` and `source_quote | text`.

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/000026_party_goals_source_fields.up.sql backend/migrations/000026_party_goals_source_fields.down.sql
git commit -m "feat(goals): add source_url/source_quote columns to party_goals"
```

---

## Task 2: Backend — expose source fields end-to-end

**Files:**
- Modify: `backend/internal/goals/domain/goal.go`
- Modify: `backend/internal/goals/adapters/postgres/repository.go`
- Modify: `backend/internal/goals/adapters/http/handler.go`
- Modify: `api/openapi.yaml`
- Regenerate: `frontend/src/shared/api-contract.ts`

**Interfaces:**
- Consumes: columns from Task 1.
- Produces: `domain.Goal.SourceURL string` (`json:"sourceUrl,omitempty"`), `domain.Goal.SourceQuote string` (`json:"sourceQuote,omitempty"`); `goalWithAlignment` response carries `sourceUrl`/`sourceQuote`; OpenAPI `Goal` schema gains optional `sourceUrl`/`sourceQuote`.

- [ ] **Step 1: Add fields to the domain entity**

In `backend/internal/goals/domain/goal.go`, inside the `Goal` struct, add after `SourceDocument`:

```go
	SourceURL          string      `json:"sourceUrl,omitempty"`
	SourceQuote        string      `json:"sourceQuote,omitempty"`
```

- [ ] **Step 2: Add columns to the repository read path**

In `backend/internal/goals/adapters/postgres/repository.go`, replace the `selectCols` const:

```go
const selectCols = `id, party, goal_text, topic, specificity, source_document, source_url, source_quote, keywords, relevant_committees, created_at`
```

- [ ] **Step 3: Update the row scan**

In the same file, replace the `s.Scan(...)` call in `scanGoal`:

```go
	err := s.Scan(&g.ID, &g.Party, &g.GoalText, &g.Topic, &specificity,
		&g.SourceDocument, &g.SourceURL, &g.SourceQuote, &g.Keywords, &g.RelevantCommittees, &g.CreatedAt)
```

(`SourceURL`/`SourceQuote` scan into `*string`; NULL rows arrive as empty string only if columns are non-null — to tolerate NULLs use the field directly since pgx maps SQL NULL TEXT to "" when scanning into a Go string is NOT safe. Use `sql.NullString`? No — pgx returns an error on NULL→string.) Because existing rows have NULL `source_url`/`source_quote`, scan into `*string` is required. Replace with NULL-safe scan:

```go
	var srcURL, srcQuote *string
	err := s.Scan(&g.ID, &g.Party, &g.GoalText, &g.Topic, &specificity,
		&g.SourceDocument, &srcURL, &srcQuote, &g.Keywords, &g.RelevantCommittees, &g.CreatedAt)
	if srcURL != nil {
		g.SourceURL = *srcURL
	}
	if srcQuote != nil {
		g.SourceQuote = *srcQuote
	}
```

- [ ] **Step 4: Persist fields on Create**

In the same file, replace the `Create` query + args so new inserts store the fields:

```go
	const q = `INSERT INTO party_goals (party, goal_text, topic, specificity, source_document, source_url, source_quote, keywords, relevant_committees)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at`
	return r.db.QueryRow(ctx, q,
		g.Party, g.GoalText, g.Topic, string(g.Specificity), g.SourceDocument,
		nullIfEmpty(g.SourceURL), nullIfEmpty(g.SourceQuote),
		g.Keywords, g.RelevantCommittees,
	).Scan(&g.ID, &g.CreatedAt)
```

Add this helper at the bottom of the file:

```go
func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
```

- [ ] **Step 5: Expose fields in the HTTP response struct**

In `backend/internal/goals/adapters/http/handler.go`, add to the `goalWithAlignment` struct after `SourceDocument`:

```go
	SourceURL          string      `json:"sourceUrl,omitempty"`
	SourceQuote        string      `json:"sourceQuote,omitempty"`
```

And in `listGoals`, add to the `goalWithAlignment{...}` literal after `SourceDocument: g.SourceDocument,`:

```go
			SourceURL:          g.SourceURL,
			SourceQuote:        g.SourceQuote,
```

(The `goalVotes` endpoint serializes `domain.Goal` directly, so it picks up the new JSON tags automatically — no change needed there.)

- [ ] **Step 6: Build the backend**

Run: `go build -C backend ./...`
Expected: exit 0, no output.

- [ ] **Step 7: Update the OpenAPI Goal schema**

In `api/openapi.yaml`, under `Goal:` → `properties:`, add after the `sourceDocument:` block (before `keywords:`):

```yaml
        sourceUrl:
          type: string
          description: Live URL to the party's published source document for this goal.
          example: "https://moderaterna.se/.../valmanifest_2022.pdf"
        sourceQuote:
          type: string
          description: Verbatim excerpt from the source document the goal is derived from.
```

(Leave `required:` unchanged — both fields are optional.)

- [ ] **Step 8: Regenerate the frontend contract**

Run: `cd frontend && npm run generate:api`
Expected: `src/shared/api-contract.ts` regenerated; `git diff --stat` shows it changed and now contains `sourceUrl`/`sourceQuote`.

- [ ] **Step 9: Typecheck the frontend**

Run: `cd frontend && npm run typecheck`
Expected: exit 0, no errors.

- [ ] **Step 10: Commit**

```bash
git add backend/internal/goals/domain/goal.go backend/internal/goals/adapters/postgres/repository.go backend/internal/goals/adapters/http/handler.go api/openapi.yaml frontend/src/shared/api-contract.ts
git commit -m "feat(goals): expose sourceUrl/sourceQuote through goals API"
```

---

## Task 3: Curated seed migration (~96 goals)

**Files:**
- Create: `backend/migrations/000027_seed_party_goals_expansion.up.sql`
- Create: `backend/migrations/000027_seed_party_goals_expansion.down.sql`

**Interfaces:**
- Consumes: source columns (Task 1) + Create/read path (Task 2).
- Produces: ~96 total `party_goals` rows (~12/party), incl. `concrete` rows and non-null `source_url`/`source_quote`.

**Curation rules (apply to every row):**
- Target ~12 goals per party (S, M, SD, C, V, KD, L, MP). Keep all existing rows (idempotent insert adds only new `goal_text`).
- `specificity = 'concrete'` only when the source states a measurable target (a number, a date, an amount in kr); otherwise `'directional'`. Do not add `'rhetorical'` rows.
- Broaden committee coverage beyond the current SoU/JuU/FiU cluster. Valid committee codes to draw from: `FiU` (finans), `SkU` (skatt), `JuU` (justitie), `SfU` (socialförsäkring/migration), `AU` (arbetsmarknad), `SoU` (socialutskottet/vård), `UbU` (utbildning), `MJU` (miljö/jordbruk), `NU` (näring/energi), `FöU` (försvar), `UU` (utrikes), `TU` (trafik), `KrU` (kultur), `CU` (civil/bostad), `KU` (konstitution).
- `keywords`: specific enough to match real Riksdagen vote subjects (the matcher is keyword-driven; vague keywords create false matches). Use lowercase Swedish nouns/stems.
- `source_document`: short label (`'Valmanifest 2022'`, `'Tidöavtalet 2022'`, `'Valplattform 2026'`, `'Partiprogram'`).
- `source_url`: live URL to the actual published document (see per-party primary sources below).
- `source_quote`: **verbatim** text copied from that document. If you cannot find verbatim text, drop the goal.
- 2026 goals only where the party has published a 2026 document (as of 2026-06-20: only S). Everything else uses 2022 manifest / Tidöavtalet / partiprogram.

**Per-party primary sources (start here; find the actual PDF/page):**
- S: <https://www.socialdemokraterna.se/var-politik> (+ 2026 valplattform: "Plan för Sverige")
- M: <https://moderaterna.se/var-politik> (valmanifest 2022 PDF; Tidöavtalet)
- SD: <https://sd.se/var-politik/> (valmanifest 2022 PDF; Tidöavtalet)
- C: <https://www.centerpartiet.se/var-politik>
- V: <https://www.vansterpartiet.se/politik/>
- KD: <https://kristdemokraterna.se/politik/> (Tidöavtalet)
- L: <https://www.liberalerna.se/politik/> (Tidöavtalet)
- MP: <https://www.mp.se/politik>

- [ ] **Step 1: Write the seed file header + worked example block**

Create `backend/migrations/000027_seed_party_goals_expansion.up.sql`. Start with this header and the first party block as the format template (every other block follows this exact column shape). Replace the example quote/url with the real verbatim values you locate:

```sql
-- Expansion of curated party goals (Promise Tracking A).
-- Idempotent: ON CONFLICT (party, goal_text) DO NOTHING (constraint from 000002).
-- Every row carries source_url + verbatim source_quote. Quotes MUST be
-- verbatim from the party's published document — never fabricated.

-- === S (Socialdemokraterna) — top up to ~12 ===
INSERT INTO party_goals
  (party, goal_text, topic, specificity, source_document, source_url, source_quote, keywords, relevant_committees) VALUES
  ('S', 'Återställa pensionerna och höja garantipensionen',
   'pension', 'directional', 'Valmanifest 2022',
   'https://www.socialdemokraterna.se/var-politik',
   '<VERBATIM QUOTE FROM SOURCE>',
   ARRAY['pension','garantipension','pensionär','pensionssystem'],
   ARRAY['SfU']),
  ('S', 'Anställa 10 000 fler i vården',
   'sjukvard', 'concrete', 'Valmanifest 2022',
   'https://www.socialdemokraterna.se/var-politik',
   '<VERBATIM QUOTE FROM SOURCE>',
   ARRAY['vård','anställa','vårdpersonal','sjuksköterska','bemanning'],
   ARRAY['SoU'])
  -- ...continue to ~12 total S rows...
ON CONFLICT (party, goal_text) DO NOTHING;
```

- [ ] **Step 2: Add the remaining seven party blocks**

Append one `INSERT ... ON CONFLICT (party, goal_text) DO NOTHING;` block per party (M, SD, C, V, KD, L, MP), each ~12 rows total (counting existing rows already in `000003`), following the identical column order and curation rules above. Every row must have a real `source_url` and verbatim `source_quote`. Include at least 2 `concrete` rows per party where the source provides a measurable target.

- [ ] **Step 3: Write the down migration**

Create `backend/migrations/000027_seed_party_goals_expansion.down.sql`. Delete only the rows this migration added, keyed by `source_document IN (...)` is unsafe (overlaps 000003). Instead delete by exact `(party, goal_text)` pairs added here:

```sql
-- Remove only rows added by 000027. List every (party, goal_text) added above.
DELETE FROM party_goals WHERE (party, goal_text) IN (
  ('S', 'Återställa pensionerna och höja garantipensionen'),
  ('S', 'Anställa 10 000 fler i vården')
  -- ...one tuple per row added in the .up.sql...
);
```

- [ ] **Step 4: Apply the migration**

Run: `make migrate`
Expected: migrate reports version `27` applied, no error.

- [ ] **Step 5: Verify counts and quality**

Run:
```bash
docker compose -f docker-compose.dev.yml exec -T postgres \
  psql -U riksdagskollen -d riksdagskollen -c \
  "SELECT party, count(*) AS goals, count(*) FILTER (WHERE specificity='concrete') AS concrete, count(*) FILTER (WHERE source_quote IS NOT NULL) AS sourced FROM party_goals GROUP BY party ORDER BY party;"
```
Expected: each party ~12 goals, ≥2 concrete, and `sourced` ≥ the number of new rows.

- [ ] **Step 6: Verify no empty/placeholder quotes leaked**

Run:
```bash
docker compose -f docker-compose.dev.yml exec -T postgres \
  psql -U riksdagskollen -d riksdagskollen -c \
  "SELECT party, goal_text FROM party_goals WHERE source_quote LIKE '%VERBATIM%' OR source_quote = '';"
```
Expected: 0 rows.

- [ ] **Step 7: Commit**

```bash
git add backend/migrations/000027_seed_party_goals_expansion.up.sql backend/migrations/000027_seed_party_goals_expansion.down.sql
git commit -m "feat(goals): expand curated party goals to ~12/party with sources"
```

---

## Task 4: Run matcher + verify API surface

**Files:** none (verification task).

**Interfaces:**
- Consumes: seeded rows (Task 3), exposed API (Task 2).

- [ ] **Step 1: Trigger the keyword matcher**

The `keyword-matcher` worker runs daily and on startup paths; force a run by restarting the backend (which also reruns ingestion wiring) or invoking the worker. Restart:

Run: `docker compose -f docker-compose.dev.yml restart backend`
Then watch logs:
```bash
docker compose -f docker-compose.dev.yml logs backend --since=2m | grep -i "keyword matcher done"
```
Expected: a log line `keyword matcher done goals=<N> matchesCreated=<M>` with `goals` reflecting the new count and no errors.

- [ ] **Step 2: Verify new goals appear via API with source fields**

Run:
```bash
curl -s localhost:8080/api/parties/S/goals | python3 -m json.tool | grep -E "sourceUrl|sourceQuote|goalText" | head
```
Expected: goal objects including `sourceUrl` and `sourceQuote` fields populated.

- [ ] **Step 3: Verify the goal-votes endpoint serializes source fields**

Run:
```bash
GID=$(curl -s localhost:8080/api/parties/S/goals | python3 -c "import sys,json;print(json.load(sys.stdin)[0]['id'])")
curl -s localhost:8080/api/parties/S/goals/$GID/votes | python3 -m json.tool | grep -E "sourceUrl|sourceQuote"
```
Expected: `sourceUrl`/`sourceQuote` present on the `goal` object.

- [ ] **Step 4: No commit** (verification only).

---

## Task 5: Data-source discipline (CLAUDE.md rule 11)

**Files:**
- Modify: `docs/data-sources/seed-party-goals.md`
- Regenerate: `frontend/src/components/sources/SourceRegistry.generated.ts`

**Interfaces:**
- Consumes: nothing new.
- Produces: corrected source doc + regenerated registry.

- [ ] **Step 1: Correct the schema section**

In `docs/data-sources/seed-party-goals.md`, replace the "Schema/fält vi använder" section so it lists the **actual** columns. Replace the bullet list with:

```markdown
## Schema/fält vi använder
- `party` — partikod (S, M, SD, C, V, KD, L, MP).
- `goal_text` — målets formulering som vi har skrivit.
- `topic` — kategori (vård, skola, skatt, etc).
- `specificity` — vår klassificering: `concrete` (mätbart),
  `directional` (riktning utan siffra), `rhetorical` (slagord).
- `source_document` — källetikett (t.ex. "Valmanifest 2022").
- `source_url` — länk till partiets publicerade källdokument.
- `source_quote` — ordagrant citat ur källdokumentet som målet bygger på.
- `keywords` — sökord för att matcha relevanta riksdagsomröstningar.
- `relevant_committees` — utskott vars omröstningar är relevanta.
```

- [ ] **Step 2: Update limitations + frontmatter**

In the same file, update the "Begränsningar" section to note that verifiability is now backed by `source_url` + verbatim `source_quote` per goal, and that 2026 coverage is partial (only S published as of 2026-06-20). Bump frontmatter `last_verified: 2026-06-20`.

- [ ] **Step 3: Regenerate the source registry**

Run: `cd frontend && npm run sync:data-sources`
Expected: `src/components/sources/SourceRegistry.generated.ts` regenerated; `git diff --stat` shows it changed (or unchanged if body hash identical — acceptable).

- [ ] **Step 4: Verify the data pages render**

Run: `cd frontend && npm run dev` (or use the running stack) and load `/data` then `/data/seed-party-goals`.
Expected: source `seed-party-goals` listed at `/data`; `/data/seed-party-goals` renders the updated markdown body including the new schema bullets.

- [ ] **Step 5: Commit**

```bash
git add docs/data-sources/seed-party-goals.md frontend/src/components/sources/SourceRegistry.generated.ts
git commit -m "docs(data-sources): update party-goals source doc for source fields"
```

---

## Self-Review

**Spec coverage:**
- Schema add `source_url`+`source_quote` → Task 1. ✓
- ~12/party, concrete+directional, broadened committees → Task 3. ✓
- Expose fields domain→repo→API + regenerate contract → Task 2. ✓
- Data-source discipline (doc fix + registry + /data render) → Task 5. ✓
- No verdict/kept-broken framing → enforced in Global Constraints + curation rules. ✓
- 2026 only where published → curation rules + Task 5 note. ✓
- Verification (build, migrate, matcher, API, typecheck) → Tasks 1/2/3/4/5. ✓

**Placeholder scan:** The `<VERBATIM QUOTE FROM SOURCE>` markers in Task 3 are intentional curation slots, not plan placeholders — the plan cannot pre-bake quotes without fabricating source text (forbidden by the spec). Task 3 Step 6 explicitly fails the build if any leak unfilled.

**Type consistency:** `SourceURL`/`SourceQuote` (Go) ↔ `source_url`/`source_quote` (SQL) ↔ `sourceUrl`/`sourceQuote` (JSON/OpenAPI) used consistently across Tasks 1, 2, 5. `selectCols` column order matches `scanGoal` scan order (Task 2 Steps 2–3). `nullIfEmpty` helper defined and used in Task 2 Step 4.
