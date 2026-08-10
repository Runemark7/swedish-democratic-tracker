# Utskott Spine — Phase 1: Committee Data Spine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every committee in the voting record addressable from the backend — its code, its display name, the utgiftsområden it bereder, and the votes and party goals belonging to it — with no UI changes.

**Architecture:** A new hexagonal feature `backend/internal/committees/` (domain → ports → adapters → service), wired in `cmd/api/main.go`. The committee list is **derived from the record per mandate period**, never from a hardcoded set; a name lookup with a fallback supplies display names and must never hide an unknown code. The utskott→utgiftsområde allocation is a registered data source seeded by migration, keyed on UO code and versioned by in-force date.

**Tech Stack:** Go 1.26.1, chi v5, pgx/v5 (raw SQL, no ORM), golang-migrate, PostgreSQL 17. Types flow spec-first: `api/openapi.yaml` → `npm run generate:api`.

## Global Constraints

- **Go 1.26.1, Node 25.9.0** — exact versions in any tooling.
- **Parameterised SQL only.** Never string-interpolate SQL.
- **Never import an adapter package into domain.** Define ports before adapters.
- **`cmd/api/main.go` is the only wiring point.**
- **`api/openapi.yaml` is the single source of truth** for HTTP types. Update it *first*, then run `cd frontend && npm run generate:api`. Never hand-edit `frontend/src/shared/api-contract.ts`.
- **Data source discipline (`CLAUDE.md` rule 11):** any new external source or seed migration requires `docs/data-sources/<id>.md` from `_TEMPLATE.md`, the mermaid diagram in `CLAUDE.md` updated, and `cd frontend && npm run sync:data-sources`.
- **Never push to `main`.** Feature branch + PR, always.
- **Verification gate before claiming done:** `go build -C backend ./...` and `cd frontend && npx tsc --noEmit` and `npm run lint` must all pass. Lint is enforced in CI as of `6230755`.
- **DB-backed Go tests** use the established pattern: read `TEST_DATABASE_URL`, `t.Skip` when unset (see `backend/internal/matching/adapters/postgres/scorecard_view_test.go:12-22`).
- **No derived comparison across parties or committees.** No shares of a total, no rankings, no ordering that implies precedence. Alphabetical everywhere. This is house style, settled three times on map #82 — it constrains API design too: do not add a `share` or `rank` field even if a consumer would find it convenient.
- **Copy is Swedish; code, comments, commits and docs are English.**

## Decisions this plan implements

| Source | Decision |
|---|---|
| [#84](https://github.com/Runemark7/swedish-democratic-tracker/issues/84) | utskott→UO is **statutory and a strict partition**: every UO has exactly one utskott, every utskott ≥1 UO. Source is the Bilaga to riksdagsordningen 2014:801. **Version by in-force date, not per riksmöte.** |
| [#97](https://github.com/Runemark7/swedish-democratic-tracker/issues/97) | The committee list is **derived from the record, per period**. Joint committees (UFöU) and abolished ones (LU, BoU) are real. `COMMITTEES` becomes a display-name lookup with a fallback, **never a gate**. |
| [#91](https://github.com/Runemark7/swedish-democratic-tracker/issues/91) | Attribution is **as of the decision** — the beteckning encodes it permanently. The remit description is as of today. State the rule; build no per-riksmöte machinery. |
| [#90](https://github.com/Runemark7/swedish-democratic-tracker/issues/90) | KOSTAR is **amounts per utgiftsområde, never a share of the total and never a ranking.** |
| [#100](https://github.com/Runemark7/swedish-democratic-tracker/issues/100) | Committees are the **single subject taxonomy**. `party_goals.relevant_committees` is the join for LOVAT. |

## What is already done — do not redo

Verified in the codebase on 2026-08-05:

- **Cursor bug is fixed.** `backend/internal/ingestion/workers/votes.go:105-109` returns early when nothing was fetched and writes the newest fetched date, not `time.Now()`.
- **MJU regex is fixed.** `frontend/src/shared/design.ts:113` matches interior uppercase and Swedish vowels.
- **Backfill of 2022–2026 is done.** 2 558 voteringar across four riksmöten.
- **Coverage honesty is done** (#107, `be6212e`): numerator counted live, denominator refreshed `@daily`, age stated.
- **`COMMITTEES` holds all 15 standing committees** (`design.ts:85-100`).

## File Structure

| File | Responsibility |
|---|---|
| `backend/internal/committees/domain/committee.go` | Entities: `Committee`, `ExpenditureArea`. No external deps. |
| `backend/internal/committees/domain/beteckning.go` | Pure parsing: beteckning → committee code, with case normalisation. No deps. |
| `backend/internal/committees/domain/beteckning_test.go` | Unit tests for the parser. No DB. |
| `backend/internal/committees/domain/names.go` | Display-name lookup **with fallback**. |
| `backend/internal/committees/ports/repository.go` | `Repository` interface. |
| `backend/internal/committees/adapters/postgres/repository.go` | SQL implementation. |
| `backend/internal/committees/adapters/postgres/repository_test.go` | DB-backed tests, `TEST_DATABASE_URL`-gated. |
| `backend/internal/committees/service.go` | Application service. |
| `backend/internal/committees/adapters/http/handler.go` | chi routes. |
| `backend/migrations/000035_utskott_utgiftsomrade.{up,down}.sql` | The statutory allocation, seeded. |
| `docs/data-sources/utskott-utgiftsomrade.md` | Registered source doc. |
| `api/openapi.yaml` | `/committees`, `/committees/{code}` + schemas. |
| `backend/cmd/api/main.go` | Wiring only. |

---

### Task 1: Beteckning → committee code, with case normalisation

The record contains `UFöU` (11 voteringar in 2022–2026) which today's frontend lookup resolves to `undefined`, so those votes display no committee at all. 2002–2006 additionally contains ALL-CAPS prefixes (`UBU`, `JUU` and six more, ~1 070 voteringar) which would be treated as distinct committees when that period is ingested. This task makes the parser the single place that decides a code.

**Files:**
- Create: `backend/internal/committees/domain/beteckning.go`
- Test: `backend/internal/committees/domain/beteckning_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `func CommitteeCode(beteckning string) string` — canonical code from a full beteckning (`"UBU14"` → `"UbU"`), or `""` when it carries none.
  - `func Canonical(rawCode string) string` — canonical form of a bare code (`"UBU"` → `"UbU"`), or `""` when it is not a committee code. Task 4 needs this because its SQL yields a bare prefix; appending a fake digit to reuse `CommitteeCode` would be a hack.

- [ ] **Step 1: Write the failing test**

Create `backend/internal/committees/domain/beteckning_test.go`:

```go
package domain_test

import (
	"testing"

	"riksdagskollen/internal/committees/domain"
)

func TestCommitteeCode(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		// Ordinary codes.
		{"SoU12", "SoU"},
		{"AU9", "AU"},
		{"FiU1", "FiU"},
		// Interior uppercase and Swedish vowels.
		{"MJU5", "MJU"},
		{"FöU3", "FöU"},
		// Joint committee: real, 11 voteringar in 2022-2026.
		{"UFöU2", "UFöU"},
		// Abolished 2006, present in 2002-2006.
		{"LU21", "LU"},
		{"BoU8", "BoU"},
		// ALL-CAPS vintage from 2002-2006 normalises to the canonical form,
		// otherwise the same committee appears twice.
		{"UBU14", "UbU"},
		{"JUU9", "JuU"},
		{"SOU1", "SoU"},
		{"MJU1", "MJU"},
		// Suffixed beteckning (yttrande) still resolves.
		{"AU1y", "AU"},
		// No committee in the beteckning.
		{"", ""},
		{"1234", ""},
		{"prop.2025/26:1", ""},
	}
	for _, c := range cases {
		if got := domain.CommitteeCode(c.in); got != c.want {
			t.Errorf("CommitteeCode(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/committees/domain/ -run TestCommitteeCode -v`
Expected: FAIL — the package does not exist yet (`no Go files` or `undefined: domain.CommitteeCode`).

- [ ] **Step 3: Write minimal implementation**

Create `backend/internal/committees/domain/beteckning.go`:

```go
package domain

import "strings"

// canonicalCodes maps an upper-cased committee code to its canonical spelling.
//
// 2002-2006 beteckningar are ALL-CAPS ("UBU14", "JUU9"). Without this, the same
// committee appears twice in a derived list — once per casing — and ~1 070
// voteringar hide under codes no lookup recognises.
var canonicalCodes = map[string]string{
	"AU": "AU", "CU": "CU", "FIU": "FiU", "FÖU": "FöU", "JUU": "JuU",
	"KRU": "KrU", "KU": "KU", "MJU": "MJU", "NU": "NU", "SFU": "SfU",
	"SKU": "SkU", "SOU": "SoU", "TU": "TU", "UBU": "UbU", "UU": "UU",
	// Joint committee: Utrikes- och försvarsutskottet.
	"UFÖU": "UFöU",
	// Abolished in 2006, present in the 2002-2006 record.
	"LU": "LU", "BOU": "BoU",
}

// CommitteeCode extracts the canonical committee code from a beteckning.
//
// A beteckning runs code-then-number ("SoU12", "AU1y"). The code is everything
// up to the first digit. Returns "" when there is no code — a votering on a
// motion rather than a betänkande carries none, and that absence is reported
// rather than guessed at.
func CommitteeCode(beteckning string) string {
	end := -1
	for i, r := range beteckning {
		if r >= '0' && r <= '9' {
			end = i
			break
		}
	}
	raw := beteckning
	if end >= 0 {
		raw = beteckning[:end]
	}
	return Canonical(raw)
}

// Canonical returns the canonical spelling of a bare committee code.
//
// Every committee code ends in "U" (utskott), which is what distinguishes a code
// from any other beteckning prefix: "prop." must not be mistaken for one.
func Canonical(rawCode string) string {
	upper := strings.ToUpper(rawCode)
	if rawCode == "" || !strings.HasSuffix(upper, "U") {
		return ""
	}
	if canon, ok := canonicalCodes[upper]; ok {
		return canon
	}
	// An unknown code is returned as-is rather than dropped. A code we have
	// never seen is a fact about the record, and hiding it would silently
	// remove real voteringar from every derived list.
	return rawCode
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/committees/domain/ -run TestCommitteeCode -v`
Expected: PASS, all cases including `{"prop.2025/26:1", ""}` — `prop.` does not end in `U`, so `Canonical` rejects it.

Add one case to the test for the bare-code entry point, since Task 4 depends on it:

```go
	if got := domain.Canonical("UBU"); got != "UbU" {
		t.Errorf("Canonical(UBU) = %q, want UbU", got)
	}
	if got := domain.Canonical("prop."); got != "" {
		t.Errorf("Canonical(prop.) = %q, want empty", got)
	}
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/committees/domain/beteckning.go backend/internal/committees/domain/beteckning_test.go
git commit -m "feat: canonical committee code from a beteckning

UFöU carries 11 voteringar in 2022-2026 and resolves to nothing in the
current frontend lookup. 2002-2006 beteckningar are ALL-CAPS, which would
split each committee in two once that period is ingested. One parser now
decides a code, normalises the casing, and returns an unknown code as-is
rather than dropping the votering."
```

---

### Task 2: Display names with a fallback that cannot hide a code

**Files:**
- Create: `backend/internal/committees/domain/names.go`
- Create: `backend/internal/committees/domain/committee.go`
- Test: `backend/internal/committees/domain/names_test.go`

**Interfaces:**
- Consumes: `domain.CommitteeCode` from Task 1.
- Produces: `func DisplayName(code string) string`; types `domain.Committee{Code, Name string; ExpenditureAreas []ExpenditureArea}` and `domain.ExpenditureArea{Code, Name string; AmountKsek int64}`.

- [ ] **Step 1: Write the failing test**

Create `backend/internal/committees/domain/names_test.go`:

```go
package domain_test

import (
	"testing"

	"riksdagskollen/internal/committees/domain"
)

func TestDisplayName(t *testing.T) {
	if got := domain.DisplayName("SoU"); got != "Socialutskottet" {
		t.Errorf("DisplayName(SoU) = %q", got)
	}
	if got := domain.DisplayName("UFöU"); got != "Sammansatta utrikes- och försvarsutskottet" {
		t.Errorf("DisplayName(UFöU) = %q", got)
	}
	if got := domain.DisplayName("LU"); got != "Lagutskottet" {
		t.Errorf("DisplayName(LU) = %q", got)
	}
	// An unknown code must fall back to itself, never to "" and never be
	// dropped: the lookup is a display aid, not a gate on what exists.
	if got := domain.DisplayName("ZZU"); got != "ZZU" {
		t.Errorf("DisplayName(ZZU) = %q, want the code itself", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/committees/domain/ -run TestDisplayName -v`
Expected: FAIL with `undefined: domain.DisplayName`.

- [ ] **Step 3: Write minimal implementation**

Create `backend/internal/committees/domain/names.go`:

```go
package domain

// displayNames gives each committee code its Swedish name.
//
// This is a display aid and never a gate. A code absent here still exists in
// the record, so DisplayName falls back to the code itself: "the 15 standing
// committees" is a fact about the present, not about the record, which also
// holds joint committees and ones abolished in 2006.
var displayNames = map[string]string{
	"AU":   "Arbetsmarknadsutskottet",
	"CU":   "Civilutskottet",
	"FiU":  "Finansutskottet",
	"FöU":  "Försvarsutskottet",
	"JuU":  "Justitieutskottet",
	"KrU":  "Kulturutskottet",
	"KU":   "Konstitutionsutskottet",
	"MJU":  "Miljö- och jordbruksutskottet",
	"NU":   "Näringsutskottet",
	"SfU":  "Socialförsäkringsutskottet",
	"SkU":  "Skatteutskottet",
	"SoU":  "Socialutskottet",
	"TU":   "Trafikutskottet",
	"UbU":  "Utbildningsutskottet",
	"UU":   "Utrikesutskottet",
	"UFöU": "Sammansatta utrikes- och försvarsutskottet",
	"LU":   "Lagutskottet",
	"BoU":  "Bostadsutskottet",
}

// DisplayName returns the committee's Swedish name, or the code itself when we
// have no name for it.
func DisplayName(code string) string {
	if name, ok := displayNames[code]; ok {
		return name
	}
	return code
}
```

Create `backend/internal/committees/domain/committee.go`:

```go
package domain

// Committee is one utskott as it appears in the record for a mandate period.
//
// Derived from the beteckningar actually present, not from a fixed list: a
// decision belongs to whoever decided it, so a committee that existed then
// appears then, and one that exists now but decided nothing does not.
type Committee struct {
	Code string `json:"code"`
	Name string `json:"name"`
	// Voteringar is how many distinct vote points the committee decided in the
	// period. A count of the record, not a measure of importance.
	Voteringar int `json:"voteringar"`
	// ExpenditureAreas are the utgiftsområden the committee bereder, per the
	// Bilaga to riksdagsordningen. Empty for a joint committee, which handles
	// no expenditure area of its own.
	ExpenditureAreas []ExpenditureArea `json:"expenditureAreas"`
}

// ExpenditureArea is one utgiftsområde with the amount allocated to it.
//
// Never carries a share of the total or a rank. A figure that invites "who is
// bigger" is where bias enters even when every number is correct, and a share
// would misrepresent committees whose remit exceeds their areas — SkU handles
// all taxation while reading as 1,0 % of expenditure, because revenue is not
// an utgiftsområde.
type ExpenditureArea struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	AmountKsek int64  `json:"amountKsek"`
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd backend && go test ./internal/committees/domain/ -v`
Expected: PASS for both `TestCommitteeCode` and `TestDisplayName`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/committees/domain/
git commit -m "feat: committee display names with a non-hiding fallback

An unknown code falls back to itself rather than to empty. The lookup is a
display aid; making it a gate is what hid UFöU's voteringar. Adds UFöU and
the two committees abolished in 2006."
```

---

### Task 3: Seed the statutory utskott→utgiftsområde allocation

**Files:**
- Create: `backend/migrations/000035_utskott_utgiftsomrade.up.sql`
- Create: `backend/migrations/000035_utskott_utgiftsomrade.down.sql`
- Create: `docs/data-sources/utskott-utgiftsomrade.md`
- Modify: `CLAUDE.md` (mermaid `ext` subgraph)

**Interfaces:**
- Produces: table `utskott_utgiftsomrade(utskott_code TEXT, uo_code TEXT, in_force_from DATE)`, one row per UO.

- [ ] **Step 1: Confirm the migration number is free**

Run: `ls backend/migrations/ | tail -6`
Expected: highest existing number is `000034`. If not, use the next free number consistently across both files and every later reference in this plan.

- [ ] **Step 2: Write the up migration**

Create `backend/migrations/000035_utskott_utgiftsomrade.up.sql`:

```sql
-- Which utskott bereder which utgiftsområde.
--
-- Statutory, not our judgement: the allocation is the Bilaga (tilläggsbestämmelse
-- 7.5.1) to riksdagsordningen (2014:801). It is a strict partition — every
-- utgiftsområde has exactly one utskott and every utskott has at least one, with
-- no orphans either way.
--
-- Keyed on UO *code*, never on name: UO13 and UO20 have been renamed without the
-- allocation changing, so matching on name would silently drop rows.
--
-- Versioned by in-force date rather than per riksmöte. Attribution of a decision
-- is permanent (the beteckning encodes it), so a past vote never needs
-- re-mapping; only the description of what an utskott handles today changes.
CREATE TABLE IF NOT EXISTS utskott_utgiftsomrade (
  utskott_code  TEXT NOT NULL,
  uo_code       TEXT NOT NULL,
  in_force_from DATE NOT NULL,
  PRIMARY KEY (uo_code, in_force_from)
);

INSERT INTO utskott_utgiftsomrade (utskott_code, uo_code, in_force_from) VALUES
  ('KU',  'UO1',  '2014-09-01'),
  ('FiU', 'UO2',  '2014-09-01'),
  ('SkU', 'UO3',  '2014-09-01'),
  ('JuU', 'UO4',  '2014-09-01'),
  ('UU',  'UO5',  '2014-09-01'),
  ('FöU', 'UO6',  '2014-09-01'),
  ('UU',  'UO7',  '2014-09-01'),
  ('SfU', 'UO8',  '2014-09-01'),
  ('SoU', 'UO9',  '2014-09-01'),
  ('SfU', 'UO10', '2014-09-01'),
  ('SfU', 'UO11', '2014-09-01'),
  ('SfU', 'UO12', '2014-09-01'),
  ('AU',  'UO13', '2014-09-01'),
  ('AU',  'UO14', '2014-09-01'),
  ('UbU', 'UO15', '2014-09-01'),
  ('UbU', 'UO16', '2014-09-01'),
  ('KrU', 'UO17', '2014-09-01'),
  ('CU',  'UO18', '2014-09-01'),
  ('NU',  'UO19', '2014-09-01'),
  ('MJU', 'UO20', '2014-09-01'),
  ('NU',  'UO21', '2014-09-01'),
  ('TU',  'UO22', '2014-09-01'),
  ('MJU', 'UO23', '2014-09-01'),
  ('NU',  'UO24', '2014-09-01'),
  ('FiU', 'UO25', '2014-09-01'),
  ('FiU', 'UO26', '2014-09-01'),
  ('FiU', 'UO27', '2014-09-01')
ON CONFLICT (uo_code, in_force_from) DO NOTHING;
```

- [ ] **Step 3: Write the down migration**

Create `backend/migrations/000035_utskott_utgiftsomrade.down.sql`:

```sql
DROP TABLE IF EXISTS utskott_utgiftsomrade;
```

- [ ] **Step 4: Apply and verify the partition holds**

Run:

```bash
make migrate
docker exec swedish-democratic-tracker-postgres-1 psql -U riksdagskollen -d riksdagskollen -c "
  SELECT count(*) AS rows, count(DISTINCT uo_code) AS uo, count(DISTINCT utskott_code) AS utskott
  FROM utskott_utgiftsomrade;
  SELECT ea.code FROM expenditure_areas ea
  LEFT JOIN utskott_utgiftsomrade m ON m.uo_code = ea.code
  WHERE m.uo_code IS NULL;"
```

Expected: `rows = 27`, `uo = 27`, `utskott = 15`, and the second query returns **zero rows** — no orphan UO.

The 15 matters: it means every one of the 15 standing committees bereder at least one utgiftsområde, which is the "no orphans either way" half of the strict partition. A lower number would mean some committee handles no expenditure area, contradicting #84. If any UO is unmatched, the codes in `expenditure_areas` differ from `UO1`…`UO27`; fix the seed to match the existing codes rather than changing `expenditure_areas`.

- [ ] **Step 5: Write the data source doc**

Create `docs/data-sources/utskott-utgiftsomrade.md` from `docs/data-sources/_TEMPLATE.md` with frontmatter:

```yaml
---
id: utskott-utgiftsomrade
name: Utskottens utgiftsområden (riksdagsordningen)
kind: seed
upstream: https://www.riksdagen.se/sv/dokument-och-lagar/dokument/svensk-forfattningssamling/riksdagsordning-2014801_sfs-2014-801/
license: Svensk författningssamling — fri användning
freshness: ändras genom lagändring; kontrolleras vid ny lydelse
last_verified: 2026-08-05
verification_status: verified
verification_notes: ~
used_by:
  - /committees (utgiftsområden per utskott)
  - /committees/:code (KOSTAR-raden)
---
```

Prose sections must state: that the allocation comes from the Bilaga (tb 7.5.1) to riksdagsordningen and **not** from riksdagen.se's committee pages; that it is a strict partition of all 27 UO; that it is keyed on code because UO13 and UO20 have been renamed; and that SFS 2026:1349 amends the bilaga with effect 2026-09-01 without affecting anything the site currently renders.

- [ ] **Step 6: Update the architecture diagram and regenerate the registry**

In `CLAUDE.md`, add to the `ext` subgraph:

```
        RO["Riksdagsordningen 2014:801\nBilaga tb 7.5.1 — utskott→utgiftsområde"]
```

and an edge `CS["committees"] -->|statutory allocation| RO` inside the feature services block.

Run: `cd frontend && npm run sync:data-sources`
Expected: source count increases by one and `utskott-utgiftsomrade` appears in `frontend/src/components/sources/SourceRegistry.generated.ts`.

- [ ] **Step 7: Verify it surfaces**

Run `make run`, open `http://localhost:5173/data` and confirm the new source is listed, then `http://localhost:5173/data/utskott-utgiftsomrade` and confirm the markdown body renders.

- [ ] **Step 8: Commit**

```bash
git add backend/migrations/000035_utskott_utgiftsomrade.up.sql \
        backend/migrations/000035_utskott_utgiftsomrade.down.sql \
        docs/data-sources/utskott-utgiftsomrade.md docs/data-sources/INDEX.md \
        CLAUDE.md frontend/src/components/sources/SourceRegistry.generated.ts
git commit -m "feat: seed the statutory utskott-utgiftsomrade allocation

From the Bilaga (tb 7.5.1) to riksdagsordningen 2014:801, not from
riksdagen.se's committee pages. A strict partition of all 27 UO, keyed on
code because UO13 and UO20 have been renamed without the allocation
changing. Versioned by in-force date, since attribution of a past decision
is permanent."
```

---

### Task 4: Repository — derive committees from the record

**Files:**
- Create: `backend/internal/committees/ports/repository.go`
- Create: `backend/internal/committees/adapters/postgres/repository.go`
- Test: `backend/internal/committees/adapters/postgres/repository_test.go`

**Interfaces:**
- Consumes: `domain.Committee`, `domain.ExpenditureArea`, `domain.CommitteeCode`, `domain.DisplayName`.
- Produces:
  - `ports.Repository` with `ListForPeriod(ctx context.Context, periodCode string) ([]domain.Committee, error)` and `AreasFor(ctx context.Context, utskottCode string, budgetYear int) ([]domain.ExpenditureArea, error)`.

- [ ] **Step 1: Write the port**

Create `backend/internal/committees/ports/repository.go`:

```go
package ports

import (
	"context"

	"riksdagskollen/internal/committees/domain"
)

// Repository reads committees as the record shows them.
type Repository interface {
	// ListForPeriod returns every committee that decided at least one votering
	// in the mandate period, alphabetically by code. Derived from the record:
	// a committee that existed but decided nothing does not appear, and one
	// abolished since does.
	ListForPeriod(ctx context.Context, periodCode string) ([]domain.Committee, error)

	// AreasFor returns the utgiftsområden a committee bereder, with the amount
	// allocated in the given budget year. Never a share of the total.
	AreasFor(ctx context.Context, utskottCode string, budgetYear int) ([]domain.ExpenditureArea, error)
}
```

- [ ] **Step 2: Write the failing test**

Create `backend/internal/committees/adapters/postgres/repository_test.go`:

```go
package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/committees/adapters/postgres"
)

func connectTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration tests")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect test DB: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// The list must come from the record, so UFöU has to appear: it decided 11
// voteringar in 2022-2026 and is absent from any hardcoded set of "the 15
// standing committees".
func TestListForPeriod_IncludesJointCommittee(t *testing.T) {
	pool := connectTestDB(t)
	repo := postgres.NewRepository(pool)

	got, err := repo.ListForPeriod(context.Background(), "2022-2026")
	if err != nil {
		t.Fatalf("ListForPeriod: %v", err)
	}
	if len(got) < 16 {
		t.Errorf("got %d committees, want at least 16 (15 standing + UFöU)", len(got))
	}

	var sawUFoU bool
	for i, c := range got {
		if c.Code == "UFöU" {
			sawUFoU = true
			if c.Name != "Sammansatta utrikes- och försvarsutskottet" {
				t.Errorf("UFöU name = %q", c.Name)
			}
			if c.Voteringar == 0 {
				t.Error("UFöU has 0 voteringar but was derived from the record")
			}
		}
		if i > 0 && got[i-1].Code > c.Code {
			t.Errorf("not alphabetical: %q before %q", got[i-1].Code, c.Code)
		}
	}
	if !sawUFoU {
		t.Error("UFöU missing — the list is gating on a hardcoded set")
	}
}

func TestAreasFor_NoShareOfTotal(t *testing.T) {
	pool := connectTestDB(t)
	repo := postgres.NewRepository(pool)

	got, err := repo.AreasFor(context.Background(), "FiU", 2026)
	if err != nil {
		t.Fatalf("AreasFor: %v", err)
	}
	// FiU bereder UO2, UO25, UO26, UO27.
	if len(got) != 4 {
		t.Errorf("FiU areas = %d, want 4", len(got))
	}
	for _, a := range got {
		if a.Code == "" || a.Name == "" {
			t.Errorf("area missing code or name: %+v", a)
		}
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd backend && TEST_DATABASE_URL="postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable" go test ./internal/committees/adapters/postgres/ -v`
Expected: FAIL with `undefined: postgres.NewRepository`.

- [ ] **Step 4: Write the implementation**

Create `backend/internal/committees/adapters/postgres/repository.go`:

```go
package postgres

import (
	"context"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/committees/domain"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ListForPeriod derives the committees from the beteckningar in the record.
//
// The prefix is taken in SQL and then canonicalised in Go, so the ALL-CAPS
// vintage in 2002-2006 collapses onto the same committee rather than appearing
// as a second one.
func (r *Repository) ListForPeriod(ctx context.Context, periodCode string) ([]domain.Committee, error) {
	rows, err := r.db.Query(ctx, `
		SELECT substring(v.beteckning from '^[A-Za-zÅÄÖåäö]+') AS code,
		       count(DISTINCT v.beteckning || ':' || v.forslagspunkt)  AS voteringar
		FROM votes v
		JOIN mandate_periods p ON v.session = ANY(p.riksmoten)
		WHERE p.code = $1
		  AND substring(v.beteckning from '^[A-Za-zÅÄÖåäö]+') <> ''
		GROUP BY 1`, periodCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Canonicalise in Go, then merge: two casings are one committee.
	merged := map[string]int{}
	for rows.Next() {
		var raw string
		var n int
		if err := rows.Scan(&raw, &n); err != nil {
			return nil, err
		}
		code := domain.Canonical(raw)
		if code == "" {
			continue
		}
		merged[code] += n
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]domain.Committee, 0, len(merged))
	for code, n := range merged {
		out = append(out, domain.Committee{
			Code:             code,
			Name:             domain.DisplayName(code),
			Voteringar:       n,
			ExpenditureAreas: []domain.ExpenditureArea{},
		})
	}
	// Alphabetical by code. Any other order — by volume, by "importance" — is a
	// ranking, and ranking is an editorial act.
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out, nil
}

// AreasFor returns the utgiftsområden a committee bereder with their amounts.
//
// The allocation in force is the newest one whose in_force_from has passed.
func (r *Repository) AreasFor(ctx context.Context, utskottCode string, budgetYear int) ([]domain.ExpenditureArea, error) {
	rows, err := r.db.Query(ctx, `
		WITH in_force AS (
			SELECT uo_code, utskott_code
			FROM utskott_utgiftsomrade m
			WHERE in_force_from = (
				SELECT max(in_force_from) FROM utskott_utgiftsomrade
				WHERE in_force_from <= CURRENT_DATE
			)
		)
		SELECT ea.code, ea.name, COALESCE(ba.amount_ksek, 0)
		FROM in_force f
		JOIN expenditure_areas ea ON ea.code = f.uo_code
		LEFT JOIN budget_years by2 ON by2.year = $2
		LEFT JOIN budget_allocations ba
		       ON ba.expenditure_area_id = ea.id
		      AND ba.budget_year_id = by2.id
		      AND ba.source = 'government'
		      AND ba.party IS NULL
		WHERE f.utskott_code = $1
		ORDER BY (regexp_replace(ea.code, '\D', '', 'g'))::int`, utskottCode, budgetYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.ExpenditureArea{}
	for rows.Next() {
		var a domain.ExpenditureArea
		if err := rows.Scan(&a.Code, &a.Name, &a.AmountKsek); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend && TEST_DATABASE_URL="postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable" go test ./internal/committees/... -v`
Expected: PASS. `TestListForPeriod_IncludesJointCommittee` should report 16 committees for `2022-2026`.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/committees/ports/ backend/internal/committees/adapters/postgres/
git commit -m "feat: derive committees from the record, per mandate period

The list comes from the beteckningar present, not from a fixed set, so
UFöU's 11 voteringar appear and a committee abolished in 2006 appears in
the period it decided in. Codes are canonicalised before merging so the
ALL-CAPS 2002-2006 vintage does not split a committee in two. Alphabetical
by code — any volume-based order would be a ranking."
```

---

### Task 5: Service, HTTP handler, spec, wiring

**Files:**
- Create: `backend/internal/committees/service.go`
- Create: `backend/internal/committees/adapters/http/handler.go`
- Modify: `api/openapi.yaml`
- Modify: `backend/cmd/api/main.go`
- Regenerate: `frontend/src/shared/api-contract.ts`

**Interfaces:**
- Consumes: `ports.Repository`.
- Produces: `GET /api/committees?period=<code>` → `Committee[]`; `GET /api/committees/{code}?period=<code>&year=<int>` → `Committee` with `expenditureAreas` populated.

- [ ] **Step 1: Write the service**

Create `backend/internal/committees/service.go`:

```go
package committees

import (
	"context"

	"riksdagskollen/internal/committees/domain"
	"riksdagskollen/internal/committees/ports"
)

type Service struct {
	repo ports.Repository
}

func NewService(repo ports.Repository) *Service {
	return &Service{repo: repo}
}

// List returns the committees of a mandate period, alphabetically.
func (s *Service) List(ctx context.Context, periodCode string) ([]domain.Committee, error) {
	return s.repo.ListForPeriod(ctx, periodCode)
}

// Get returns one committee with its utgiftsområden for the given budget year.
//
// Returns nil when the committee decided nothing in the period: a page that
// invented a committee out of a URL would claim something the record does not.
func (s *Service) Get(ctx context.Context, periodCode, code string, budgetYear int) (*domain.Committee, error) {
	all, err := s.repo.ListForPeriod(ctx, periodCode)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].Code != code {
			continue
		}
		areas, err := s.repo.AreasFor(ctx, code, budgetYear)
		if err != nil {
			return nil, err
		}
		all[i].ExpenditureAreas = areas
		return &all[i], nil
	}
	return nil, nil
}
```

- [ ] **Step 2: Add the endpoints to the spec**

In `api/openapi.yaml`, add under `paths:`:

```yaml
  /committees:
    get:
      summary: Committees that decided at least one votering in a mandate period
      parameters:
        - name: period
          in: query
          required: true
          schema: { type: string }
          description: Mandate period code, e.g. "2022-2026".
      responses:
        "200":
          description: Committees, alphabetical by code
          content:
            application/json:
              schema:
                type: array
                items: { $ref: "#/components/schemas/Committee" }

  /committees/{code}:
    get:
      summary: One committee with the utgiftsområden it bereder
      parameters:
        - name: code
          in: path
          required: true
          schema: { type: string }
        - name: period
          in: query
          required: true
          schema: { type: string }
        - name: year
          in: query
          required: false
          schema: { type: integer }
          description: Budget year for the amounts. Defaults to the newest held.
      responses:
        "200":
          description: The committee
          content:
            application/json:
              schema: { $ref: "#/components/schemas/Committee" }
        "404":
          description: The committee decided nothing in that period
```

and under `components.schemas:`:

```yaml
    Committee:
      type: object
      required: [code, name, voteringar, expenditureAreas]
      properties:
        code:
          type: string
          description: Canonical committee code, e.g. "SoU". ALL-CAPS vintages are normalised.
        name:
          type: string
          description: >-
            Swedish name, or the code itself when we have no name for it. The
            lookup is a display aid and never a gate on what the record contains.
        voteringar:
          type: integer
          description: >-
            Distinct vote points decided in the period. A count of the record,
            not a measure of importance — never presented as a ranking.
        expenditureAreas:
          type: array
          description: >-
            Utgiftsområden the committee bereder, per the Bilaga to
            riksdagsordningen. Amounts only, never a share of the total: a share
            would understate a committee whose remit exceeds its areas.
          items: { $ref: "#/components/schemas/CommitteeExpenditureArea" }

    CommitteeExpenditureArea:
      type: object
      required: [code, name, amountKsek]
      properties:
        code: { type: string }
        name: { type: string }
        amountKsek: { type: integer, format: int64 }
```

- [ ] **Step 3: Write the handler**

Create `backend/internal/committees/adapters/http/handler.go`:

```go
package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/committees"
)

type Handler struct {
	svc *committees.Service
}

func NewHandler(svc *committees.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r chi.Router) {
	r.Get("/committees", h.list)
	r.Get("/committees/{code}", h.get)
}

const defaultBudgetYear = 2026

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		jsonError(w, "period is required", http.StatusBadRequest)
		return
	}
	cs, err := h.svc.List(r.Context(), period)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, cs)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		jsonError(w, "period is required", http.StatusBadRequest)
		return
	}
	year := defaultBudgetYear
	if y := r.URL.Query().Get("year"); y != "" {
		parsed, err := strconv.Atoi(y)
		if err != nil {
			jsonError(w, "year must be an integer", http.StatusBadRequest)
			return
		}
		year = parsed
	}
	c, err := h.svc.Get(r.Context(), period, chi.URLParam(r, "code"), year)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if c == nil {
		jsonError(w, "committee not found in that period", http.StatusNotFound)
		return
	}
	jsonOK(w, c)
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
```

- [ ] **Step 4: Wire it in main.go**

In `backend/cmd/api/main.go`, alongside the other feature wiring, add imports:

```go
	committeesFeature "riksdagskollen/internal/committees"
	committeesHTTP "riksdagskollen/internal/committees/adapters/http"
	committeesPG "riksdagskollen/internal/committees/adapters/postgres"
```

and in the route registration block:

```go
	committeeSvc := committeesFeature.NewService(committeesPG.NewRepository(db))
	committeesHTTP.NewHandler(committeeSvc).Register(r)
```

Match the surrounding style — find how an existing handler such as the riksdag one is registered and follow it exactly.

- [ ] **Step 5: Build, regenerate types, verify**

```bash
go build -C backend ./...
cd frontend && npm run generate:api && npx tsc --noEmit && npm run lint
```
Expected: all clean, and `Committee` appears in `frontend/src/shared/api-contract.ts`.

- [ ] **Step 6: Verify the endpoints against real data**

```bash
cd backend && DATABASE_URL="postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable" PORT=8099 INITIAL_SYNC=false go run ./cmd/api &
sleep 10
curl -s "http://localhost:8099/api/committees?period=2022-2026" | python3 -m json.tool | head -30
curl -s "http://localhost:8099/api/committees/FiU?period=2022-2026&year=2026" | python3 -m json.tool
curl -s -o /dev/null -w "%{http_code}\n" "http://localhost:8099/api/committees/ZZU?period=2022-2026"
```

Expected: the list holds **16 committees including `UFöU`**, alphabetical by code; `FiU` returns four expenditure areas (UO2, UO25, UO26, UO27) with amounts and **no share field**; the unknown code returns **404**.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/committees/ backend/cmd/api/main.go api/openapi.yaml frontend/src/shared/api-contract.ts
git commit -m "feat: committee endpoints

GET /api/committees lists the committees a mandate period's record shows,
alphabetically. GET /api/committees/{code} adds the utgiftsområden the
committee bereder, amounts only — no share of the total, because a share
understates a committee whose remit exceeds its areas, and no rank."
```

---

## Self-Review

**Spec coverage.** Phase 1 of the spec listed cursor fix, backfill, freshness honesty, and MJU/KrU resolution. Three of those four are already shipped and are recorded above under "What is already done" with the evidence; the fourth (committee resolution) is Tasks 1–2, widened by #97 from a bug fix into deriving the list from the record. The spec's Phase 2 prerequisites — the utskott→UO mapping table and its data source doc — are Task 3, and the backend endpoints the committee page needs are Tasks 4–5. The committee *page* itself is Phase 2 of this plan series, not here.

**Placeholders.** None: every step carries the exact SQL, Go, YAML or command to run, and the two intentional failure points (Task 1 Step 4's `prop.` case, Task 3 Step 4's orphan check) state what to do when they trigger.

**Type consistency.** `domain.CommitteeCode` and `domain.DisplayName` are defined in Tasks 1–2 and used unchanged in Task 4. `domain.Committee` / `domain.ExpenditureArea` field names match the OpenAPI schema (`code`, `name`, `voteringar`, `expenditureAreas`, `amountKsek`). `ports.Repository` methods `ListForPeriod` / `AreasFor` are declared in Task 4 Step 1 and consumed in Task 5's service with the same signatures.

**A gap worth naming.** LOVAT needs party goals per committee, which this plan does not add: `party_goals.relevant_committees` is a `text[]` on an existing table owned by the `goals` feature, so the query belongs there rather than in a new `committees` repository reaching across features. It is the first task of Phase 2, where the row that consumes it is built.

**Honest note on testing.** Backend tasks here carry real Go tests, and the codebase has 13 existing test files plus a `TEST_DATABASE_URL` harness to follow. The frontend has **no component test runner** — only `test:scripts` (`node --test scripts/__tests__/*.test.mjs`) — so Phase 2 and Phase 3 cannot open with a failing component test. Their verification is `tsc` + `lint` + `build` + a scripted browser check, and those plans will say so rather than inventing a test step that cannot run.
