# Myndighetsregister-pipeline — Design

**Datum:** 2026-05-25
**Status:** Godkänd design, redo för implementeringsplan
**Bygger på:** PR #35 (`feat/riksdag-myndigheter-table`) — MYNDIGHETER-kortet med tabellstil. Den här pipelinen fyller kortet + ny listsida med riktig, komplett data.

---

## Mål

Ersätt den hårdkodade listan på 10 myndigheter med Sveriges **hela** statliga
myndighetsregister (~449), hämtat från auktoritativa öppna källor, så att
Riksdag-sidan visar officiellt korrekt och komplett data — namn för alla,
utgift och anställda där källa finns.

## Problem (nuläge)

- `backend/internal/riksdag/adapters/statskontoret/client.go` laddar ner
  Statskontorets årsutfall-CSV men filtrerar till **10 hårdkodade anslagskoder**
  (`targetAnslag`). Allt annat slängs.
- `backend/internal/riksdag/adapters/static/client.go` = 10 hårdkodade
  myndigheter (fallback).
- `backend/internal/riksdag/adapters/scb/client.go` (anställda) mappar bara
  **8 KLS-koder**.
- Ingen DB-tabell för myndigheter — allt live-hämtas med 24h minnescache.
- Resultat: kortet visar 10 myndigheter. Sverige har ~449 i registret
  (~371 förvaltningsmyndigheter under regeringen).

## Scope

- **Population:** alla **449** myndigheter från SCB Myndighetsregistret,
  taggade med `huvudman`/`typ` och en flagga `under_regeringen` (~371) så att
  UI kan visa alla eller filtrera.
- **Per myndighet:**
  - Grunddata (namn, org-nr, typ, huvudman, departement) — **alltid**.
  - Utgift (ESV) — **best-effort**, `null` när källa saknas → "saknas" i UI.
  - Anställda (Arbetsgivarverket) — **best-effort**, `null` → "saknas".
- **Fakta-lager-princip:** aldrig hitta på värden. Saknad data visas som
  "saknas", inte 0 eller gissning. Varje värde har källmarkör.
- **Ej i scope:** kommunala/regionala nämnder (ingår inte i SCB-registret).

## Datakällor

| Källa | Ger | Åtkomst | Join-nyckel |
|---|---|---|---|
| SCB Myndighetsregistret | Lista 449: org-nr, namn, typ, huvudman | Företagsregister-API (gratis, kräver certifikat via mejl) **eller** webbregister-export | org-nr |
| ESV/Hermes öppna data (`esv.se/psidata`) | Utfall per anslag **och myndighet** | Öppna data-filer/endpoints | org-nr (fallback namn) |
| Arbetsgivarverket "Anställda i staten" | Anställda per myndighet, 1991→ | Öppna data/export, uppdateras dec/juni | org-nr (fallback namn) |

**Notera:** ESV bytte namn till Statskontoret 2026-01-01 (sammanslagning).
Endpoints/branding är i rörelse — exakta URL:er verifieras i planeringssteget.

**Identitet:** `org_nr` är kanonisk nyckel. Källor som saknar org-nr matchas
via normaliserat namn (gemener, trimmat, utan bolagsform-suffix). Omatchade
rader loggas — inga gissningsjoins.

## Arkitektur

Hexagonalt, i befintlig feature `backend/internal/riksdag/`.

### Domän
`domain/myndighet.go`:
```
type Myndighet struct {
    OrgNr            string
    Slug             string
    Name             string
    Typ              string   // t.ex. "Förvaltningsmyndighet", "Domstol"
    Huvudman         string   // "Regeringen", "Riksdagen", ...
    Departement      string
    UnderRegeringen  bool
    ExpenditureMdkr  *float64 // nil = saknas
    BudgetMdkr       *float64
    HeadcountInt     *int     // nil = saknas
    Year             int
    ExpenditureHistory []YearlyExpenditure
    HeadcountHistory   []YearlyHeadcount
    UpdatedAt        time.Time
}
```
Pekare för utgift/anställda gör "saknas" explicit i hela stacken.

### Portar (`ports/`)
```
type RegisterClient interface {
    FetchRegister(ctx) ([]RegisterEntry, error)   // 449 grunddata
}
type ExpenditureClient interface {
    FetchExpenditure(ctx) ([]AgencyExpenditure, error) // ESV, per org-nr
}
type HeadcountClient interface {
    FetchHeadcounts(ctx) ([]AgencyHeadcount, error)    // Arbetsgivarverket
}
type MyndighetRepository interface {
    UpsertMyndigheter(ctx, []domain.Myndighet) error
    List(ctx, MyndighetFilter) ([]domain.Myndighet, error) // sök/filter/paginering
    ListTop(ctx, n int) ([]domain.Myndighet, error)        // topp-N efter utgift
    GetBySlug(ctx, slug string) (*domain.Myndighet, error)
}
```

### Adapters (`adapters/`)
- `registret/` — SCB Myndighetsregistret (ny).
- `esv/` — ESV/Hermes utfall (ny; ersätter anslag-filtret i `statskontoret/`).
- `arbetsgivarverket/` — anställda (ny; ersätter den begränsade `scb/`-klienten).
- `postgres/myndighet_repository.go` — ny.
- Behåll `agency_intel_repository` (regleringsbrev/beslut) och statiska
  10-listan som **nöd-fallback** om DB-tabellen är tom (första boot).

### Ingestion-worker
`internal/ingestion/workers/myndigheter.go`, cron **@weekly**:
1. `FetchRegister` → upsert grundrader (org-nr nyckel, beräkna slug + `under_regeringen`).
2. `FetchExpenditure` → uppdatera utgift + historik per org-nr.
3. `FetchHeadcounts` → uppdatera anställda + historik per org-nr (namn-fallback).

**Robusthet:** varje steg har egen try/catch. Misslyckas en källa: logga,
behåll senaste lyckade data — **wipe:a aldrig** vid hämtningsfel. Partiellt
resultat sparas.

### Service
`service.go` läser från `MyndighetRepository` (ej live):
- `GetAuthorities(ctx)` / `ListTop(ctx, 10)` → kortet.
- `ListMyndigheter(ctx, filter)` → listsidan (sök/filter/paginering).
- `GetAuthority(ctx, slug)` → detaljsidan (befintlig, läser nu DB + intel).

## Datamodell (migration)

Ny migration `000020_myndigheter.up.sql`:
```sql
CREATE TABLE myndigheter (
    org_nr            TEXT PRIMARY KEY,
    slug              TEXT NOT NULL UNIQUE,
    name              TEXT NOT NULL,
    typ               TEXT NOT NULL DEFAULT '',
    huvudman          TEXT NOT NULL DEFAULT '',
    departement       TEXT NOT NULL DEFAULT '',
    under_regeringen  BOOLEAN NOT NULL DEFAULT FALSE,
    expenditure_mdkr  DOUBLE PRECISION,            -- NULL = saknas
    budget_mdkr       DOUBLE PRECISION,
    headcount_int     INTEGER,                     -- NULL = saknas
    year              INTEGER NOT NULL DEFAULT 0,
    expenditure_history JSONB NOT NULL DEFAULT '[]',
    headcount_history   JSONB NOT NULL DEFAULT '[]',
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_myndigheter_under_regeringen ON myndigheter (under_regeringen);
CREATE INDEX idx_myndigheter_expenditure ON myndigheter (expenditure_mdkr DESC NULLS LAST);
CREATE INDEX idx_myndigheter_name_trgm ON myndigheter USING gin (name gin_trgm_ops);
```
Historik som JSONB (visas, frågas inte) → en tabell, inga extra join-tabeller.
Sök via `pg_trgm` på namn (kräver `CREATE EXTENSION IF NOT EXISTS pg_trgm`).

## API (openapi.yaml först → regenerera api-contract.ts)

- `GET /api/riksdag/authorities?limit=10` — topp-N efter utgift (befintlig,
  läser nu DB). Driver kortet.
- `GET /api/riksdag/myndigheter?q=&huvudman=&underRegeringen=&page=&pageSize=`
  — sökbar/filtrerbar lista (ny). Driver listsidan.
- `GET /api/riksdag/myndigheter/{slug}` — detalj (befintlig).

Uppdatera `api/openapi.yaml` först, kör `npm run generate:api`. Hand-editera
aldrig `api-contract.ts`.

## Frontend

- **Kortet** (`RiksdagPage`, MYNDIGHETER, från PR #35): topp-10 efter utgift +
  knapp "Läs om fler myndigheter →" till `/riksdag/myndigheter`.
- **Ny sida** `/riksdag/myndigheter` (`MyndigheterListPage`):
  sökfält (namn), filter (huvudman, toggle "under regeringen"), samma
  tabellstil (Myndighet / Senaste / Andel / Förändring + chevron),
  paginering/virtualisering för hundratals rader. Rad-expand = aktuellt år +
  anställda + kostnad + "Läs mer →" (detaljsida).
- Detaljsida `/riksdag/myndigheter/:slug` finns — läser nu DB-data.

## Data-source-disciplin (CLAUDE.md-krav)

1. Uppdatera mermaid-diagrammet i `CLAUDE.md`: lägg till SCB Myndighetsregistret,
   ESV/Hermes, Arbetsgivarverket; justera Statskontoret årsutfall.
2. `docs/data-sources/`: nya MD-filer (`scb-myndighetsregistret.md`,
   `esv-utfall.md`, `arbetsgivarverket-anstallda.md`) från `_TEMPLATE.md` med
   frontmatter + reproducerbara nedladdningssteg.
3. `cd frontend && npm run sync:data-sources` → regenerera SourceRegistry.
4. `<SourceMarker sourceId="...">` på alla nya UI-värden + `<SectionSource>`.
5. Verifiera att källorna syns på `/data` och `/data/<id>`.

## Felhantering

- Worker: per-källa isolering, partiellt resultat sparas, ingen wipe vid fel.
- Saknad utgift/anställda → `NULL` → UI "saknas".
- Org-nr kanonisk; namn-fallback normaliserat; omatchade rader loggas, ej
  gissade.
- DB tom (första boot före worker) → statiska 10 som nöd-fallback.
- `INITIAL_SYNC=true` kör worker direkt vid boot (befintligt mönster).

## Testning

Repo har inga tester än, men ny logik med hög buggrisk testas (TDD):
- Parsers: register, ESV-utfall, anställda (fixtures från riktig data).
- Merge/join: org-nr-match + namn-fallback + null-hantering.
- Repo: upsert (idempotent), List-filter/sök/paginering (integration mot test-DB).
Rena funktioner (parse/merge) prioriteras — störst värde, lätta att testa.

## Fasad utrullning

Varje fas är självständigt levererbar och testbar:

1. **Register-grund:** migration + `registret`-klient + repo + worker (bara
   register) + service läser DB + kortet visar 449 namn (utgift/anställda kan
   vara "saknas" initialt). Ersätter de hårdkodade 10.
2. **Utgift:** `esv`-klient + worker-steg + utgift i kort/detalj.
3. **Anställda:** `arbetsgivarverket`-klient + worker-steg + anställda.
4. **Listsida + sök:** `/riksdag/myndigheter` + "fler"-knapp + list-API-endpoint.

Max 5 filer per fas (CLAUDE.md), verifiera (`go build`, `tsc`) mellan faser.

## Risker / öppna frågor (verifieras i plan)

- **SCB register-API kräver certifikat** (mejl-godkännande, kan dröja).
  Fallback: webbregister-export/scrape, eller periodisk manuell CSV.
- **ESV↔Statskontoret-namnbyte** (jan 2026): bekräfta aktuella öppna
  data-endpoints/format.
- **org-nr** kanske saknas i ESV/Arbetsgivarverket-export → namn-matchning
  fallback; mät träffgrad.
- **Utgift-täckning**: myndigheter utan eget anslag får "saknas" — förväntat.

## Framgångskriterier

- `/riksdag/myndigheter` listar alla ~449 med fungerande sök + filter.
- Kortet visar topp-10 efter utgift + "fler"-knapp.
- Utgift/anställda visas där källa finns, "saknas" annars — inga gissade värden.
- Varje värde har källmarkör; källorna syns på `/data`.
- Data persisterar i DB; en källas utfall kraschar inte sidan.
