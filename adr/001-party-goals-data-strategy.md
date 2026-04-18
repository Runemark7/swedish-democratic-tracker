# ADR-001: Party Goals Data Strategy

**Status:** Accepted
**Date:** 2026-04-05
**Decided:** 2026-04-05

## Context

Riksdagskollen tracks how Swedish parliamentary parties' voting records align with their stated goals. The data pipeline has two distinct categories:

### Live API data (politicians, votes, speeches)
- Fetched from `data.riksdagen.se` on startup (`INITIAL_SYNC=true`) and daily via cron
- ~2,200 politicians, ~4,000 votes, ~1,500 speeches per sync
- Stored in PostgreSQL, deduplicated via UNIQUE constraints and upserts
- No caching layer exists; every container restart re-fetches everything
- Hardcoded 100-200ms delays between API calls as rate limiting
- The `ingestion_cursors` table exists in the schema but is unused (designed for incremental sync)

### Editorial data (party goals)
- 44 manually curated goals extracted from valmanifest 2022, Tidöavtalet, and party programs
- Each goal has: text, party, topic, specificity, source document, keywords, relevant committee codes
- Currently stored in `backend/seeds/dev_seed.sql` (214 lines)
- Loaded by a dedicated `seed` service in `docker-compose.yml` that runs `psql -f dev_seed.sql`
- The keyword matcher worker depends on goals existing in DB before it can match votes
- Uses `ON CONFLICT (party, goal_text) DO NOTHING` for idempotency

### Problems with current approach
1. **Separate service for seed data** adds docker-compose complexity and a race condition with initial sync
2. **Full API re-fetch on every restart** wastes bandwidth and time (~10-25 min for full sync)
3. **No distinction between dev and prod data paths** — the seed is called "dev_seed" but is the only source of goals
4. **Goals have no update workflow** — editing requires changing SQL and redeploying

## Decision Required

Two separate concerns need decisions:

### Concern A: Where should party goals live?

#### Option A1: Bake into a database migration

Move the INSERT statements from `dev_seed.sql` into `backend/migrations/000003_seed_party_goals.up.sql`. Remove the `seed` service from docker-compose and delete `backend/seeds/`.

| Aspect | Detail |
|--------|--------|
| **Pros** | Single mechanism (migrations auto-run); deterministic; version-controlled; no separate service |
| **Cons** | Purist concern: migrations are for schema, not data; updating goals requires a new migration file |
| **Effort** | ~30 minutes |
| **Files changed** | Move SQL, remove `seed` service from docker-compose, delete `backend/seeds/` |

#### Option A2: Go `embed` in application binary

Use `//go:embed` to include goals as a SQL or JSON file compiled into the Go binary. On startup, after migrations but before ingestion, check if `party_goals` is empty and insert.

| Aspect | Detail |
|--------|--------|
| **Pros** | Self-contained single binary; no external files or services; version-controlled |
| **Cons** | Binary includes editorial content; requires rebuild to update goals; mixes concerns |
| **Effort** | ~1-2 hours |
| **Files changed** | New embed file, new startup function, remove seed service |

#### Option A3: Keep seed, rename to "fixtures"

Rename `backend/seeds/` to `backend/fixtures/`, rename docker-compose service. Semantic improvement, same mechanism.

| Aspect | Detail |
|--------|--------|
| **Pros** | Minimal change; clear separation of schema (migrations) from content (fixtures) |
| **Cons** | Still requires separate docker-compose service; still called at container startup |
| **Effort** | ~15 minutes |
| **Files changed** | Rename directory, update docker-compose volume mount |

#### Option A4: Admin API for goals CRUD

Migration seeds the initial 44 goals. Add REST endpoints for creating, updating, and deleting goals. Future admin UI for editorial workflow.

| Aspect | Detail |
|--------|--------|
| **Pros** | Non-developers can update goals without code changes; proper editorial workflow |
| **Cons** | Requires authentication; significantly more work; admin UI not yet planned |
| **Effort** | 1-2 days |
| **Files changed** | New handler, new routes, auth middleware, admin frontend |

### Concern B: How should Riksdagen API data be cached?

#### Option B1: PostgreSQL-level caching (current implicit approach)

The current UNIQUE constraints + upserts effectively cache API data in PostgreSQL. On restart, the sync re-fetches and upserts, but existing data is served immediately while sync runs in background.

| Aspect | Detail |
|--------|--------|
| **Pros** | Already works; no new infrastructure; data survives container restarts if volume persists |
| **Cons** | Full re-fetch every restart wastes API bandwidth; ~10-25 min sync time |

#### Option B2: Incremental sync via `ingestion_cursors`

The `ingestion_cursors` table already exists in the schema. Implement cursor tracking: after each sync, record the latest date/ID. Next sync only fetches data newer than the cursor.

| Aspect | Detail |
|--------|--------|
| **Pros** | Dramatically reduces API calls after first sync; faster daily syncs; already has schema support |
| **Cons** | Need to implement cursor read/write logic; some API endpoints may not support date filtering well |
| **Effort** | ~3-4 hours |
| **Files changed** | New `ingestion_cursors` repository, update each worker to read/write cursors |

#### Option B3: HTTP response caching (Redis or in-memory)

Add a cache layer (Redis or `sync.Map` with TTL) that intercepts Riksdagen API responses. Cache key = URL, TTL = 24h. On cache hit, skip the API call.

| Aspect | Detail |
|--------|--------|
| **Pros** | Transparent to workers; reduces API load; enables offline development |
| **Cons** | New dependency (Redis) or memory overhead; cache invalidation complexity; doesn't survive container restarts (in-memory) |
| **Effort** | ~4-6 hours |
| **Files changed** | New cache adapter, wrap each Riksdagen client |

#### Option B4: Pre-built database dump

Create a `pg_dump` snapshot of a fully-synced database. On first startup, restore from dump instead of syncing from API. Periodically regenerate the dump in CI.

| Aspect | Detail |
|--------|--------|
| **Pros** | Instant startup; no API dependency in dev; deterministic |
| **Cons** | Large file in repo (or artifact storage); stale between dump regenerations; CI pipeline needed |
| **Effort** | ~2-3 hours |
| **Files changed** | CI job, docker-compose init container, dump storage |

## Recommendation

**Concern A (goals):** A1 (migration) is simplest and most pragmatic. Goals change infrequently (once per election cycle). A new migration per update is a feature, not a bug — it gives a clear audit trail of when goals changed.

**Concern B (caching):** B2 (incremental cursors) gives the best ROI. The schema already exists, and it eliminates the ~10-25 minute full re-fetch on every daily sync. B4 (db dump) is worth considering for dev experience but adds CI complexity.

## Decision

**Concern A: A1 — Bake party goals into a database migration.**

Move the INSERT statements from `backend/seeds/dev_seed.sql` into `backend/migrations/000003_seed_party_goals.up.sql`. Remove the `seed` service from `docker-compose.yml` and delete `backend/seeds/`. Goals now run automatically as part of the migration pipeline — no separate seeding step needed. Future goal updates are new migrations (e.g., `000004_update_party_goals_2026.up.sql`), providing a clear audit trail.

**Concern B: B2 — Incremental sync via `ingestion_cursors` table.**

Implement cursor-based tracking using the existing `ingestion_cursors` table (already in schema since migration 000001). After each worker completes, record the latest timestamp or ID as the cursor. On the next run, only fetch records newer than the stored cursor. This eliminates the ~10-25 minute full re-fetch on daily syncs and dramatically reduces Riksdagen API load.

## Consequences

### Concern A consequences
- `docker-compose.yml` loses the `seed` service (simpler orchestration, no race condition)
- `backend/seeds/` directory is deleted
- Party goals are guaranteed to exist before any worker runs (migrations execute before application start)
- Updating goals requires a new migration file and a deploy — acceptable since goals change once per election cycle
- The `ON CONFLICT (party, goal_text) DO NOTHING` clause makes the migration idempotent and safe to re-run

### Concern B consequences
- Each worker (politicians, speeches, votes) gains cursor read/write logic
- First startup still does a full sync (no cursor exists yet); subsequent runs are incremental
- The `ingestion_cursors` table stores `(data_type, last_date, last_id, updated_at)` per worker
- Riksdagen API endpoints that don't support date filtering will still do full fetches but upsert idempotently
- `EnrichOrigins` worker naturally tracks progress via the `origin_enriched` boolean — no cursor needed
- Daily cron syncs drop from ~10-25 minutes to seconds (only new data since last cursor)
- `INITIAL_SYNC` env var remains for first-boot full sync; subsequent container restarts with a persisted volume skip the full fetch

### Implementation order
1. Move seed SQL into migration 000003 and remove seed service (Concern A)
2. Add cursor repository (port + postgres adapter) for `ingestion_cursors`
3. Update politicians, speeches, and votes workers to read/write cursors
4. Verify incremental sync: restart containers with persisted volume, confirm no full re-fetch
