# Home Page — Riksdag Activity Overview Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace `HomePage.tsx` with a Riksdag-focused activity overview: hero + 5 panels (Beslut idag · Aktuella debatter · Veckans omröstningar · Vad partierna säger · Kommande beslut). Remove the three-chamber identity surface and the unused `ChambersHero` / `StatusCell` / `PulseStrip` blocks.

**Architecture:** Speeches feature already exists end-to-end except the HTTP layer. Add a `ListRecent` query through the speeches port → service → existing Postgres repository, expose it as `GET /api/speeches/recent` and `GET /api/speeches/{id}`, then build a new home page that fetches `useRiksdag()` (existing) plus `useRecentSpeeches()` (new) and renders five panels via a shared `<PanelCard>` chrome.

**Tech Stack:** Go 1.26 (chi, pgx/v5), React 19 + Vite + TypeScript, TanStack Query v5, Tailwind v4. No test framework in repo (per CLAUDE.md) — verification is `go build`, `npx tsc --noEmit`, `curl` against the running backend, and Playwright at 375×812 for visual mobile checks against prod.

**Spec:** `docs/superpowers/specs/2026-05-05-home-page-riksdag-overview-design.md`. Read it before Task 1.

**Branching:** Work directly on `main`. No worktree, no PR. Commit each task individually with a `feat(home):` / `feat(speeches):` / `chore(api):` prefix as relevant; push at the end of each backend or frontend group.

---

## Schema reality check

The `speeches` table has `date DATE NOT NULL` (day-precision only — no `HH:MM`). The spec referenced `occurred_at TIMESTAMPTZ`; that field does not exist. Plan resolution: Panel 2 displays `DD/MM` on the left of each item instead of `HH:MM`. No migration needed. Where `occurred_at` appears in the spec, treat it as `date` (Swedish `YYYY-MM-DD`).

---

## File map (locks decomposition)

**New files**

| Path | Responsibility |
|---|---|
| `backend/internal/speeches/adapters/http/handler.go` | chi routes for `/speeches/recent` and `/speeches/{id}`. JSON DTOs, no business logic. |
| `frontend/src/features/speeches/api.ts` | Typed client for the new endpoints (`speechesApi.listRecent`, `speechesApi.getById`). |
| `frontend/src/features/speeches/SpeechDetailPage.tsx` | Stub detail page: full speech text + politician + party + back link. |
| `frontend/src/features/home/components/PanelCard.tsx` | Shared chrome (border + mono kicker + items + "Visa alla" footer + empty state). |
| `frontend/src/features/home/components/PartySpeechCard.tsx` | Single party card used by Panel 4. |

**Modified files**

| Path | Change |
|---|---|
| `backend/internal/speeches/ports/repository.go` | Add `ListRecent(ctx, limit) ([]*domain.Speech, error)`. |
| `backend/internal/speeches/adapters/postgres/repository.go` | Implement `ListRecent` with `ORDER BY date DESC, id DESC LIMIT $1`. |
| `backend/internal/speeches/service.go` | Pass-through `ListRecent(ctx, limit)`. |
| `backend/cmd/api/main.go` | Construct `speechesHTTP.NewHandler(speechSvc, polRepo)`, mount at `/api`. |
| `api/openapi.yaml` | Add `/speeches/recent`, `/speeches/{id}`, `Speech` schema. |
| `frontend/src/shared/api-contract.ts` | Regenerated from updated OpenAPI. |
| `frontend/src/hooks/useDemocracy.ts` | Add `useRecentSpeeches(limit)` hook. |
| `frontend/src/App.tsx` | Add route `/anforanden/:id` → `<SpeechDetailPage>`. |
| `frontend/src/features/home/HomePage.tsx` | Full rewrite. |

`mockRegion` / `mockKommun` references disappear from `HomePage`. The mock module itself stays (other features may import it).

---

## Task 1: Backend — `ListRecent` port + repository

**Files:**
- Modify: `backend/internal/speeches/ports/repository.go`
- Modify: `backend/internal/speeches/adapters/postgres/repository.go`

- [ ] **Step 1: Add `ListRecent` to the port interface**

Edit `backend/internal/speeches/ports/repository.go`. Add the method to the existing `SpeechRepository` interface:

```go
type SpeechRepository interface {
	GetByID(ctx context.Context, id int) (*domain.Speech, error)
	ListByPolitician(ctx context.Context, politicianID string) ([]*domain.Speech, error)
	ListRecent(ctx context.Context, limit int) ([]*domain.Speech, error)
	UpsertMany(ctx context.Context, ss []*domain.Speech) error
	ListUnprocessed(ctx context.Context, limit int) ([]*domain.Speech, error)
	MarkProcessed(ctx context.Context, id int) error
}
```

- [ ] **Step 2: Implement `ListRecent` on the Postgres repository**

Edit `backend/internal/speeches/adapters/postgres/repository.go`. Append after `ListUnprocessed`:

```go
func (r *Repository) ListRecent(ctx context.Context, limit int) ([]*domain.Speech, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	q := "SELECT " + selectCols + " FROM speeches ORDER BY date DESC, id DESC LIMIT $1"
	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSpeeches(rows)
}
```

The 200-row clamp protects against accidental denial-of-service via `?limit=999999`. Default 100 matches the spec's panel-4 sizing.

- [ ] **Step 3: Verify the backend compiles**

Run from repo root:

```bash
go build -C backend ./...
```

Expected: empty output, exit code 0. If the build complains that `ListRecent` is missing from a different `SpeechRepository` implementation (e.g. an in-memory test fake), add the method there too — but the codebase currently only has the Postgres impl, so this should not happen.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/speeches/ports/repository.go backend/internal/speeches/adapters/postgres/repository.go
git commit -m "feat(speeches): add ListRecent repository method"
```

---

## Task 2: Backend — service pass-through

**Files:**
- Modify: `backend/internal/speeches/service.go`

- [ ] **Step 1: Add `ListRecent` to the service**

Append a method on `*Service`:

```go
func (s *Service) ListRecent(ctx context.Context, limit int) ([]*domain.Speech, error) {
	return s.repo.ListRecent(ctx, limit)
}
```

- [ ] **Step 2: Verify compile**

```bash
go build -C backend ./...
```

Expected: empty output, exit 0.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/speeches/service.go
git commit -m "feat(speeches): expose ListRecent on service"
```

---

## Task 3: Backend — HTTP handler for `/speeches/recent` and `/speeches/{id}`

**Files:**
- Create: `backend/internal/speeches/adapters/http/handler.go`

The handler joins `speeches` → `politicians` (via `politician_id` → `intressent_id`) so the JSON payload includes `politicianName` directly. To keep dependencies hexagonal, the handler accepts a `politicians.Service`-shaped lookup interface rather than reaching into the politicians postgres adapter.

- [ ] **Step 1: Create the handler file**

Create `backend/internal/speeches/adapters/http/handler.go`:

```go
package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/speeches"
	"riksdagskollen/internal/speeches/domain"
)

// PoliticianLookup resolves intressent_id → display name. The speeches
// handler depends on this minimal surface (not the full politicians
// service) to stay decoupled.
type PoliticianLookup interface {
	NameByID(ctx context.Context, intressentID string) (string, error)
}

type Handler struct {
	svc *speeches.Service
	pol PoliticianLookup
}

func NewHandler(svc *speeches.Service, pol PoliticianLookup) *Handler {
	return &Handler{svc: svc, pol: pol}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/speeches/recent", h.listRecent)
	r.Get("/speeches/{id}", h.getByID)
}

// SpeechDTO is the JSON shape returned to the frontend.
type SpeechDTO struct {
	ID              int       `json:"id"`
	DokID           string    `json:"dokId"`
	AnforandeNummer string    `json:"anforandeNummer"`
	PoliticianID    string    `json:"politicianId"`
	PoliticianName  string    `json:"politicianName"`
	Party           string    `json:"party"`
	Date            time.Time `json:"date"`
	TopicHeading    string    `json:"topicHeading"`
	Snippet         string    `json:"snippet"`
	SpeechText      string    `json:"speechText,omitempty"`
}

func (h *Handler) listRecent(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			limit = n
		}
	}
	ss, err := h.svc.ListRecent(r.Context(), limit)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]SpeechDTO, 0, len(ss))
	for _, s := range ss {
		out = append(out, h.toDTO(r.Context(), s, false))
	}
	jsonOK(w, out)
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	s, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s == nil {
		jsonError(w, "speech not found", http.StatusNotFound)
		return
	}
	jsonOK(w, h.toDTO(r.Context(), s, true))
}

func (h *Handler) toDTO(ctx context.Context, s *domain.Speech, includeFullText bool) SpeechDTO {
	name, _ := h.pol.NameByID(ctx, s.PoliticianID)
	dto := SpeechDTO{
		ID:              s.ID,
		DokID:           s.DokID,
		AnforandeNummer: s.AnforandeNummer,
		PoliticianID:    s.PoliticianID,
		PoliticianName:  name,
		Party:           s.Party,
		Date:            s.Date,
		TopicHeading:    s.TopicHeading,
		Snippet:         snippetFrom(s.SpeechText, 180),
	}
	if includeFullText {
		dto.SpeechText = s.SpeechText
	}
	return dto
}

func snippetFrom(text string, maxRunes int) string {
	if text == "" {
		return ""
	}
	r := []rune(text)
	if len(r) <= maxRunes {
		return text
	}
	return string(r[:maxRunes]) + "…"
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
```

- [ ] **Step 2: Add `NameByID` to the politicians service**

The handler depends on `PoliticianLookup`. Confirm `politicians.Service` already exposes a method that fits, or add one. Run:

```bash
grep -n "func (s \*Service) " backend/internal/politicians/service.go
```

If you don't see a method that returns just the display name by `intressent_id`, add this to `backend/internal/politicians/service.go`:

```go
func (s *Service) NameByID(ctx context.Context, intressentID string) (string, error) {
	p, err := s.repo.GetByID(ctx, intressentID)
	if err != nil {
		return "", err
	}
	if p == nil {
		return "", nil
	}
	return p.FirstName + " " + p.LastName, nil
}
```

The exact field names (`FirstName`, `LastName`) come from `backend/internal/politicians/domain/politician.go` — open it to confirm before pasting.

- [ ] **Step 3: Verify compile**

```bash
go build -C backend ./...
```

Expected: empty output, exit 0.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/speeches/adapters/http/handler.go backend/internal/politicians/service.go
git commit -m "feat(speeches): add HTTP handler for /speeches/recent and /speeches/{id}"
```

---

## Task 4: Backend — wire handler in `cmd/api/main.go`

**Files:**
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: Add the import alias near the other speeches imports**

Find the existing block around lines 27-30:

```go
// Feature: speeches
speechesPG "riksdagskollen/internal/speeches/adapters/postgres"
speechesRD "riksdagskollen/internal/speeches/adapters/riksdagen"
"riksdagskollen/internal/speeches"
```

Add a new line below `"riksdagskollen/internal/speeches"`:

```go
speechesHTTP "riksdagskollen/internal/speeches/adapters/http"
```

- [ ] **Step 2: Construct and mount the handler**

Find where the other handlers are wired (around line 134 where `speechSvc := speeches.NewService(...)` is). Right after:

```go
speechSvc := speeches.NewService(speechRepo, speechRD)
```

add:

```go
speechHandler := speechesHTTP.NewHandler(speechSvc, polSvc)
```

Note: `polSvc` is the politicians service variable already constructed earlier in `main.go` — confirm it implements `NameByID(ctx, intressentID) (string, error)` from Task 3 Step 2. If `polSvc` is named differently locally, use the actual name.

Then find the route-mounting block (it'll look like `r.Route("/api", func(r chi.Router) { ... })` or a sequence of `polHandler.Routes(r)` calls). Add:

```go
speechHandler.Routes(r)
```

next to the other `*Handler.Routes(r)` invocations under `/api`. The route mount point must match — copy the surrounding structure exactly.

- [ ] **Step 3: Verify compile**

```bash
go build -C backend ./...
```

Expected: empty output, exit 0.

- [ ] **Step 4: Smoke-test against the running stack**

In a second terminal:

```bash
make run
```

Wait for `backend` to log `migrations applied`. Then in a third terminal:

```bash
curl -s 'http://localhost:8080/api/speeches/recent?limit=3' | python3 -m json.tool | head -40
```

Expected: a JSON array of three speeches with fields `id`, `politicianName`, `party`, `date`, `topicHeading`, `snippet`. If the array is empty, the speeches table genuinely has no rows yet (the worker may not have run yet) — that's still a pass for this task as long as the response is `[]` and HTTP 200.

Also test the by-id endpoint:

```bash
curl -s -i 'http://localhost:8080/api/speeches/999999' | head -3
```

Expected: `HTTP/1.1 404 Not Found`.

- [ ] **Step 5: Commit**

```bash
git add backend/cmd/api/main.go
git commit -m "feat(speeches): wire HTTP handler for recent + by-id"
```

---

## Task 5: OpenAPI spec — `/speeches/recent`, `/speeches/{id}`

**Files:**
- Modify: `api/openapi.yaml`

- [ ] **Step 1: Add the path definitions**

Open `api/openapi.yaml`. Under `paths:` add two entries (alphabetised among existing paths):

```yaml
  /speeches/recent:
    get:
      summary: Recent speeches across all politicians
      tags: [speeches]
      parameters:
        - in: query
          name: limit
          schema:
            type: integer
            default: 100
            minimum: 1
            maximum: 200
      responses:
        '200':
          description: List of recent speeches
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Speech'
  /speeches/{id}:
    get:
      summary: Single speech by id (full text)
      tags: [speeches]
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: Full speech detail
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Speech'
        '404':
          description: Speech not found
```

- [ ] **Step 2: Add the `Speech` schema component**

Under `components: schemas:` add:

```yaml
    Speech:
      type: object
      required: [id, dokId, politicianId, politicianName, party, date, snippet]
      properties:
        id:
          type: integer
        dokId:
          type: string
        anforandeNummer:
          type: string
        politicianId:
          type: string
        politicianName:
          type: string
        party:
          type: string
        date:
          type: string
          format: date-time
        topicHeading:
          type: string
        snippet:
          type: string
        speechText:
          type: string
          description: Full text — only included on /speeches/{id}, omitted from /speeches/recent
```

- [ ] **Step 3: Regenerate the frontend types**

```bash
cd frontend && npm run generate:api
```

Expected: `frontend/src/shared/api-contract.ts` is rewritten and now contains a `Speech` definition. Run `git diff --stat frontend/src/shared/api-contract.ts` to confirm changes; the diff should reference `/speeches/recent` and `/speeches/{id}`.

- [ ] **Step 4: Commit**

```bash
git add api/openapi.yaml frontend/src/shared/api-contract.ts
git commit -m "chore(api): add /speeches/recent + /speeches/{id} to OpenAPI"
```

---

## Task 6: Frontend — `speechesApi` client

**Files:**
- Create: `frontend/src/features/speeches/api.ts`

- [ ] **Step 1: Create the api file**

```ts
import { api } from "@/shared/api-client";

export interface Speech {
  id: number;
  dokId: string;
  anforandeNummer?: string;
  politicianId: string;
  politicianName: string;
  party: string;
  date: string; // ISO date-time
  topicHeading?: string;
  snippet: string;
  speechText?: string;
}

export const speechesApi = {
  listRecent: (limit = 100) =>
    api.get<Speech[]>(`/speeches/recent?limit=${limit}`),
  getById: (id: number) =>
    api.get<Speech>(`/speeches/${id}`),
};
```

This duplicates the generated `Speech` shape on purpose — the generated types live in `api-contract.ts` but are awkward to import (deep nested `paths['...']` traversal). A flat domain type is fine and stays in sync via this manual mirror.

- [ ] **Step 2: Verify typecheck**

```bash
cd frontend && npx tsc --noEmit
```

Expected: empty output, exit 0.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/features/speeches/api.ts
git commit -m "feat(speeches): add typed API client"
```

---

## Task 7: Frontend — `useRecentSpeeches` hook

**Files:**
- Modify: `frontend/src/hooks/useDemocracy.ts`

- [ ] **Step 1: Add an import**

At the top of `useDemocracy.ts`, alongside the other feature-api imports:

```ts
import { speechesApi, type Speech } from "@/features/speeches/api";
```

- [ ] **Step 2: Append the hook to the bottom of the file**

```ts
// ── Recent speeches ───────────────────────────────────────────────────────────
export function useRecentSpeeches(limit = 100) {
  return useQuery<Speech[]>({
    queryKey: ["speeches-recent", limit],
    queryFn: () => speechesApi.listRecent(limit),
    staleTime: 60_000,
  });
}
```

- [ ] **Step 3: Verify typecheck**

```bash
cd frontend && npx tsc --noEmit
```

Expected: empty output.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/hooks/useDemocracy.ts
git commit -m "feat(speeches): add useRecentSpeeches hook"
```

---

## Task 8: Frontend — `<PanelCard>` shared chrome

**Files:**
- Create: `frontend/src/features/home/components/PanelCard.tsx`

- [ ] **Step 1: Create the component**

```tsx
import { Link } from "react-router-dom";
import { useMediaQuery } from "@/hooks/useMediaQuery";

interface PanelCardProps {
  title: string; // e.g. "BESLUT IDAG"
  showAllHref?: string; // omit to hide the footer link
  showAllLabel?: string; // default "Visa alla →"
  emptyText?: string;
  isEmpty?: boolean;
  children?: React.ReactNode;
}

export function PanelCard({
  title,
  showAllHref,
  showAllLabel = "Visa alla →",
  emptyText,
  isEmpty,
  children,
}: PanelCardProps) {
  const isMobile = useMediaQuery("(max-width: 640px)");

  return (
    <section
      style={{
        background: "var(--color-sdt-surface)",
        border: "1px solid var(--color-border)",
        padding: isMobile ? 16 : 24,
        display: "flex",
        flexDirection: "column",
        gap: 12,
        minWidth: 0,
      }}
    >
      <header
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 10,
          letterSpacing: "0.15em",
          color: "var(--color-fg-muted)",
          textTransform: "uppercase",
        }}
      >
        {title}
      </header>

      {isEmpty ? (
        <div
          style={{
            fontFamily: "var(--font-body)",
            fontStyle: "italic",
            fontSize: 13,
            color: "var(--color-fg-muted)",
            padding: "8px 0",
          }}
        >
          {emptyText ?? "Ingen aktivitet just nu."}
        </div>
      ) : (
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          {children}
        </div>
      )}

      {showAllHref && !isEmpty && (
        <Link
          to={showAllHref}
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            letterSpacing: "0.1em",
            color: "var(--color-accent)",
            textDecoration: "none",
            marginTop: "auto",
          }}
        >
          {showAllLabel}
        </Link>
      )}
    </section>
  );
}
```

- [ ] **Step 2: Verify typecheck**

```bash
cd frontend && npx tsc --noEmit
```

Expected: empty output.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/features/home/components/PanelCard.tsx
git commit -m "feat(home): PanelCard shared chrome component"
```

---

## Task 9: Frontend — `<PartySpeechCard>` for Panel 4

**Files:**
- Create: `frontend/src/features/home/components/PartySpeechCard.tsx`

- [ ] **Step 1: Create the card**

```tsx
import { Link } from "react-router-dom";
import { PARTY_COLORS, partyShortToName } from "@/shared/design";
import type { Speech } from "@/features/speeches/api";

interface PartySpeechCardProps {
  party: string; // short code: S, M, SD, ...
  speech: Speech | null; // null → empty card
}

const WEEKDAY = ["SÖN", "MÅN", "TIS", "ONS", "TOR", "FRE", "LÖR"];

function formatStamp(iso: string): string {
  const d = new Date(iso);
  const wd = WEEKDAY[d.getDay()];
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${wd} ${dd}/${mm}`;
}

export function PartySpeechCard({ party, speech }: PartySpeechCardProps) {
  const color = PARTY_COLORS[party] ?? "#7c8896";
  const partyName = partyShortToName(party);

  const body = (
    <article
      style={{
        background: "var(--color-sdt-surface)",
        border: "1px solid var(--color-border)",
        padding: 14,
        display: "flex",
        flexDirection: "column",
        gap: 8,
        minHeight: 130,
        height: "100%",
      }}
    >
      <div style={{ display: "flex", alignItems: "center", gap: 8, minWidth: 0 }}>
        <span
          style={{
            display: "inline-block",
            width: 4,
            height: 18,
            background: color,
            flexShrink: 0,
          }}
        />
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            fontWeight: 700,
            color: "var(--color-fg)",
            letterSpacing: "0.05em",
          }}
        >
          {party}
        </span>
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            color: "var(--color-fg-muted)",
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
            minWidth: 0,
          }}
        >
          {partyName}
        </span>
      </div>

      {speech ? (
        <>
          <div
            style={{
              fontFamily: "var(--font-serif)",
              fontSize: 14,
              color: "var(--color-fg)",
              lineHeight: 1.3,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {speech.politicianName || "Anonym"}
          </div>
          <p
            style={{
              fontFamily: "var(--font-body)",
              fontSize: 12,
              fontStyle: "italic",
              color: "var(--color-fg-muted)",
              lineHeight: 1.4,
              margin: 0,
              display: "-webkit-box",
              WebkitLineClamp: 2,
              WebkitBoxOrient: "vertical",
              overflow: "hidden",
            }}
          >
            "{speech.snippet}"
          </p>
          <div
            style={{
              marginTop: "auto",
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.1em",
              textTransform: "uppercase",
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {formatStamp(speech.date)}
            {speech.topicHeading ? ` · ${speech.topicHeading}` : ""}
          </div>
        </>
      ) : (
        <div
          style={{
            fontFamily: "var(--font-body)",
            fontStyle: "italic",
            fontSize: 12,
            color: "var(--color-fg-muted)",
            marginTop: 6,
          }}
        >
          Inget anförande senaste 30 dagarna
        </div>
      )}
    </article>
  );

  if (speech) {
    return (
      <Link
        to={`/anforanden/${speech.id}`}
        style={{ textDecoration: "none", color: "inherit", display: "block", height: "100%" }}
      >
        {body}
      </Link>
    );
  }
  return body;
}
```

- [ ] **Step 2: Confirm `PARTY_COLORS` and `partyShortToName` exist in `@/shared/design`**

```bash
grep -nE "PARTY_COLORS|partyShortToName" frontend/src/shared/design.ts
```

Expected: matches for both. If `partyShortToName` does not exist, add it to `frontend/src/shared/design.ts`:

```ts
const PARTY_NAMES: Record<string, string> = {
  S: "Socialdemokraterna",
  M: "Moderaterna",
  SD: "Sverigedemokraterna",
  C: "Centerpartiet",
  V: "Vänsterpartiet",
  KD: "Kristdemokraterna",
  L: "Liberalerna",
  MP: "Miljöpartiet",
};

export function partyShortToName(short: string): string {
  return PARTY_NAMES[short] ?? short;
}
```

- [ ] **Step 3: Verify typecheck**

```bash
cd frontend && npx tsc --noEmit
```

Expected: empty output. If it complains about a missing export, fix the export in `design.ts` first.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/home/components/PartySpeechCard.tsx frontend/src/shared/design.ts
git commit -m "feat(home): PartySpeechCard component for Panel 4"
```

---

## Task 10: Frontend — `<SpeechDetailPage>` stub + route

**Files:**
- Create: `frontend/src/features/speeches/SpeechDetailPage.tsx`
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: Create the stub page**

```tsx
import { Link, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { speechesApi } from "@/features/speeches/api";
import { useMediaQuery } from "@/hooks/useMediaQuery";

export function SpeechDetailPage() {
  const { id } = useParams<{ id: string }>();
  const isMobile = useMediaQuery("(max-width: 640px)");
  const speechId = Number(id ?? "0");
  const { data, isLoading, error } = useQuery({
    queryKey: ["speech", speechId],
    queryFn: () => speechesApi.getById(speechId),
    enabled: speechId > 0,
  });

  return (
    <div className="sdt-page">
      <div style={{ padding: isMobile ? "18px 14px" : "32px" }}>
        <Link
          to="/"
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            letterSpacing: "0.1em",
            color: "var(--color-accent)",
            textDecoration: "none",
          }}
        >
          ← Tillbaka
        </Link>
        {isLoading && (
          <div style={{ marginTop: 24, color: "var(--color-fg-muted)" }}>Laddar…</div>
        )}
        {error && (
          <div style={{ marginTop: 24, color: "var(--color-pulse)" }}>Kunde inte hämta anförandet.</div>
        )}
        {data && (
          <article style={{ marginTop: 24, maxWidth: 720 }}>
            <header style={{ marginBottom: 16 }}>
              <div
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 10,
                  letterSpacing: "0.15em",
                  color: "var(--color-fg-muted)",
                  textTransform: "uppercase",
                  marginBottom: 6,
                }}
              >
                ANFÖRANDE · {data.party} · {new Date(data.date).toLocaleDateString("sv-SE")}
              </div>
              <h1
                style={{
                  fontFamily: "var(--font-serif)",
                  fontSize: isMobile ? 24 : 32,
                  margin: 0,
                  color: "var(--color-fg)",
                  lineHeight: 1.2,
                }}
              >
                {data.politicianName}
              </h1>
              {data.topicHeading && (
                <div style={{ fontSize: 14, color: "var(--color-fg-muted)", marginTop: 4 }}>
                  {data.topicHeading}
                </div>
              )}
            </header>
            <p
              style={{
                fontSize: 15,
                lineHeight: 1.7,
                color: "var(--color-fg)",
                whiteSpace: "pre-wrap",
              }}
            >
              {data.speechText || data.snippet}
            </p>
          </article>
        )}
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Add the route**

In `frontend/src/App.tsx`, find the existing `<Route path="/sok"` line and add (alphabetised among the routes):

```tsx
<Route path="/anforanden/:id"                     element={<SpeechDetailPage />} />
```

Add the import at the top with the other feature-page imports:

```tsx
import { SpeechDetailPage } from "./features/speeches/SpeechDetailPage";
```

- [ ] **Step 3: Verify typecheck**

```bash
cd frontend && npx tsc --noEmit
```

Expected: empty output.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/speeches/SpeechDetailPage.tsx frontend/src/App.tsx
git commit -m "feat(speeches): stub SpeechDetailPage + route /anforanden/:id"
```

---

## Task 11: Frontend — rewrite `HomePage`

**Files:**
- Modify: `frontend/src/features/home/HomePage.tsx`

This is the biggest task. Replace the entire file contents. The skeleton block at the top (loading state) is preserved in spirit but inlined — there's no separate skeleton component.

- [ ] **Step 1: Replace `HomePage.tsx` contents**

```tsx
import { Link } from "react-router-dom";
import { useRiksdag, useRecentSpeeches } from "@/hooks/useDemocracy";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { Pill } from "@/components/charts";
import { AgendaList } from "@/components/AgendaList";
import { PanelCard } from "./components/PanelCard";
import { PartySpeechCard } from "./components/PartySpeechCard";
import type { LiveVote } from "@/types/democracy";
import type { Speech } from "@/features/speeches/api";

const PARTY_ORDER = ["S", "M", "SD", "C", "V", "KD", "L", "MP"] as const;

function pillTone(status: string): "pass" | "fail" | "pending" | "neutral" {
  if (status === "Bifall") return "pass";
  if (status === "Avslag") return "fail";
  if (status === "Återremiss") return "pending";
  return "neutral";
}

function beslutHref(v: LiveVote): string | null {
  if (!v.beteckning) return null;
  const p = new URLSearchParams({ title: v.title, status: v.status, tag: v.tag, time: v.time });
  return `/beslut/${encodeURIComponent(v.beteckning)}?${p}`;
}

function isToday(time: string): boolean {
  return time.toLowerCase().startsWith("idag");
}

function isThisWeek(time: string): boolean {
  // mock + feed conventions: "Idag HH:MM", "Igår HH:MM", or "DD MMM"
  const t = time.toLowerCase();
  if (t.startsWith("idag") || t.startsWith("igår")) return true;
  // best-effort 7-day check on "DD MMM" — assume current month for v1
  const m = t.match(/^(\d{1,2})\s+(\S+)/);
  if (!m) return false;
  const day = Number(m[1]);
  const now = new Date();
  const diffDays = Math.abs(now.getDate() - day);
  return diffDays <= 7;
}

function pickLatestPerParty(speeches: Speech[]): Record<string, Speech | null> {
  const out: Record<string, Speech | null> = {};
  for (const p of PARTY_ORDER) out[p] = null;
  for (const s of speeches) {
    if (!(s.party in out)) continue;
    const existing = out[s.party];
    if (!existing || s.date > existing.date) out[s.party] = s;
  }
  return out;
}

export function HomePage() {
  const isMobile = useMediaQuery("(max-width: 640px)");
  const { data: riksdag, isLoading: riksdagLoading } = useRiksdag();
  const { data: speeches, isLoading: speechesLoading } = useRecentSpeeches(100);

  if (riksdagLoading || !riksdag) {
    return (
      <div className="sdt-page" style={{ padding: isMobile ? "20px 14px" : "40px 32px" }}>
        <div
          style={{
            height: 60,
            background: "var(--color-track)",
            borderRadius: 4,
            marginBottom: 24,
            maxWidth: 480,
          }}
        />
        <div style={{ height: 320, background: "var(--color-track)", borderRadius: 4 }} />
      </div>
    );
  }

  const liveVotes = riksdag.liveVotes ?? [];
  const agenda = riksdag.agenda ?? [];

  const todayVotes = liveVotes.filter((v) => isToday(v.time)).slice(0, 5);
  const weekVotes = liveVotes.filter((v) => isThisWeek(v.time)).slice(0, 5);

  const recentSpeeches = (speeches ?? []).slice(0, 5);
  const latestByParty = pickLatestPerParty(speeches ?? []);

  // Layout: 2-col grid on desktop for panels 1+2 and 3+5; full width for 4.
  // Mobile: single column, order 1, 2, 3, 5, 4.
  const panelGrid: React.CSSProperties = {
    display: "grid",
    gap: 1,
    gridTemplateColumns: isMobile ? "1fr" : "1fr 1fr",
    background: "var(--color-border)",
    border: "1px solid var(--color-border)",
    margin: isMobile ? "14px 14px 0" : "22px 32px 0",
  };

  const partyGridStyle: React.CSSProperties = {
    display: "grid",
    gridTemplateColumns: isMobile ? "1fr 1fr" : "repeat(4, 1fr)",
    gap: 12,
    margin: isMobile ? "14px 14px 28px" : "22px 32px 28px",
  };

  return (
    <div className="sdt-page">
      {/* ── Hero ─────────────────────────────────────────────────────── */}
      <header
        style={{
          padding: isMobile ? "20px 14px 16px" : "40px 32px 24px",
          display: "flex",
          flexDirection: "column",
          gap: 8,
        }}
      >
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "2px",
            color: "var(--color-fg-muted)",
            textTransform: "uppercase",
          }}
        >
          KAMMARE ETT · RIKSDAGEN
        </div>
        <h1
          style={{
            fontFamily: "var(--font-serif)",
            fontSize: isMobile ? 28 : 44,
            fontWeight: 400,
            letterSpacing: "-1.2px",
            color: "var(--color-fg)",
            margin: 0,
            lineHeight: 1.05,
          }}
        >
          Vad fokuserar riksdagen på?
        </h1>
        <p
          style={{
            fontFamily: "var(--font-body)",
            fontSize: 13,
            color: "var(--color-fg-muted)",
            margin: 0,
            maxWidth: 540,
          }}
        >
          Beslut, debatter och vad partierna driver — i realtid.
        </p>
      </header>

      {/* ── Panels 1 + 2 (row) ───────────────────────────────────────── */}
      <div style={panelGrid}>
        {/* Panel 1 — Beslut idag */}
        <PanelCard
          title="BESLUT IDAG"
          showAllHref="/votes"
          isEmpty={todayVotes.length === 0}
          emptyText="Inga beslut idag — kammaren kan ha sommaruppehåll."
        >
          {todayVotes.map((v, i) => {
            const href = beslutHref(v);
            const row = (
              <div
                key={i}
                style={{
                  display: "grid",
                  gridTemplateColumns: "12px 1fr auto",
                  gap: 10,
                  alignItems: "center",
                  minWidth: 0,
                }}
              >
                <span
                  style={{
                    width: 7,
                    height: 7,
                    borderRadius: "50%",
                    background: i < 2 ? "var(--color-pulse)" : "var(--color-accent)",
                    display: "inline-block",
                  }}
                />
                <span
                  style={{
                    fontSize: 13,
                    color: "var(--color-fg)",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                    minWidth: 0,
                  }}
                  title={v.title}
                >
                  {v.title}
                </span>
                <Pill tone={pillTone(v.status)}>{v.status}</Pill>
              </div>
            );
            return href ? (
              <Link key={i} to={href} style={{ textDecoration: "none", color: "inherit" }}>
                {row}
              </Link>
            ) : (
              <div key={i}>{row}</div>
            );
          })}
        </PanelCard>

        {/* Panel 2 — Aktuella debatter */}
        <PanelCard
          title="AKTUELLA DEBATTER"
          showAllHref="/politicians"
          isEmpty={!speechesLoading && recentSpeeches.length === 0}
          emptyText="Inga registrerade debatter ännu."
        >
          {recentSpeeches.map((s) => {
            const date = new Date(s.date);
            const stamp = `${String(date.getDate()).padStart(2, "0")}/${String(date.getMonth() + 1).padStart(2, "0")}`;
            return (
              <Link
                key={s.id}
                to={`/anforanden/${s.id}`}
                style={{
                  display: "grid",
                  gridTemplateColumns: "44px 1fr",
                  gap: 10,
                  alignItems: "baseline",
                  textDecoration: "none",
                  color: "inherit",
                  minWidth: 0,
                }}
              >
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 10,
                    color: "var(--color-fg-muted)",
                  }}
                >
                  {stamp}
                </span>
                <span
                  style={{
                    fontSize: 13,
                    color: "var(--color-fg)",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                    minWidth: 0,
                  }}
                  title={s.politicianName}
                >
                  <strong style={{ fontWeight: 600 }}>{s.politicianName || "Anonym"}</strong>
                  <span style={{ color: "var(--color-fg-muted)", marginLeft: 6 }}>
                    [{s.party}]
                  </span>
                  {s.topicHeading && (
                    <span style={{ marginLeft: 8, color: "var(--color-fg-muted)" }}>
                      · {s.topicHeading}
                    </span>
                  )}
                </span>
              </Link>
            );
          })}
        </PanelCard>
      </div>

      {/* ── Panels 3 + 5 (row) ───────────────────────────────────────── */}
      <div style={{ ...panelGrid, marginTop: 1, borderTop: "none" }}>
        {/* Panel 3 — Veckans omröstningar */}
        <PanelCard
          title="VECKANS OMRÖSTNINGAR"
          showAllHref="/votes"
          isEmpty={weekVotes.length === 0}
          emptyText="Inga omröstningar denna vecka."
        >
          {weekVotes.map((v, i) => {
            const href = beslutHref(v);
            const row = (
              <div
                key={i}
                style={{
                  display: "grid",
                  gridTemplateColumns: "60px 1fr auto",
                  gap: 10,
                  alignItems: "center",
                  minWidth: 0,
                }}
              >
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 10,
                    color: "var(--color-fg-muted)",
                  }}
                >
                  {v.time}
                </span>
                <span
                  style={{
                    fontSize: 13,
                    color: "var(--color-fg)",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                    minWidth: 0,
                  }}
                  title={v.title}
                >
                  {v.title}
                </span>
                <Pill tone={pillTone(v.status)}>{v.status}</Pill>
              </div>
            );
            return href ? (
              <Link key={i} to={href} style={{ textDecoration: "none", color: "inherit" }}>
                {row}
              </Link>
            ) : (
              <div key={i}>{row}</div>
            );
          })}
        </PanelCard>

        {/* Panel 5 — Kommande beslut */}
        <PanelCard
          title="KOMMANDE BESLUT"
          showAllHref="/votes"
          isEmpty={agenda.length === 0}
          emptyText="Ingen agenda publicerad."
        >
          <AgendaList items={agenda.slice(0, 5)} />
        </PanelCard>
      </div>

      {/* ── Panel 4 — Vad partierna säger (full width) ───────────────── */}
      <section
        style={{
          margin: isMobile ? "14px 14px 6px" : "22px 32px 6px",
        }}
      >
        <header
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "0.15em",
            color: "var(--color-fg-muted)",
            textTransform: "uppercase",
          }}
        >
          VAD PARTIERNA SÄGER
        </header>
      </section>
      <div style={partyGridStyle}>
        {PARTY_ORDER.map((p) => (
          <PartySpeechCard key={p} party={p} speech={latestByParty[p]} />
        ))}
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify typecheck**

```bash
cd frontend && npx tsc --noEmit
```

Expected: empty output. If `LiveVote` import is unused after edits, delete it.

- [ ] **Step 3: Lint**

```bash
cd frontend && npm run lint
```

Expected: zero errors. Warnings are tolerable but try to fix `react-hooks/exhaustive-deps` if any appear.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/home/HomePage.tsx
git commit -m "feat(home): rewrite HomePage as Riksdag activity overview

5 panels (Beslut idag, Aktuella debatter, Veckans omröstningar,
Vad partierna säger, Kommande beslut) sourced from useRiksdag +
useRecentSpeeches. Drops ChambersHero, StatusCell grid, and
PulseStrip. Mobile order: 1,2,3,5,4. Desktop: 2-col panel pairs +
full-width party grid."
```

---

## Task 12: Verification

- [ ] **Step 1: Run the full stack**

```bash
make run
```

Wait for `frontend on :5173` and `backend on :8080` log lines.

- [ ] **Step 2: Local sanity — desktop**

Open `http://localhost:5173/` in a desktop browser. Confirm:
- Hero text reads `KAMMARE ETT · RIKSDAGEN` / `Vad fokuserar riksdagen på?` / subtitle.
- Two rows of 2 panels (1+2 then 3+5), then 8-card grid (Panel 4).
- No `ChambersHero` boxes, no `StatusCell` cells, no `PulseStrip` timeline strip.

- [ ] **Step 3: Local sanity — mobile**

In Chrome DevTools, switch to iPhone 14 (375 × 812). Confirm:
- All panels stack as a single column.
- Panel order from top: 1, 2, 3, 5, 4.
- No horizontal scroll. Verify with the document.documentElement.scrollWidth check from earlier audits if needed.

- [ ] **Step 4: Push and wait for deploy**

```bash
git push origin main
```

Watch the GitHub Actions deploy update `deploy/charts/riksdagskollen/values-prod.yaml` (it auto-bumps the image tag), then ArgoCD rolls. Usually 4-7 minutes.

- [ ] **Step 5: Production verification at 375 × 812**

Use the Playwright MCP browser at 375 × 812 against `https://sdt.runevibe.se/`. Run:

```js
() => ({
  url: location.pathname,
  overflowPx: document.documentElement.scrollWidth - document.documentElement.clientWidth,
});
```

Expected: `overflowPx: 0`.

Spot-check three click paths:
1. Panel 1 first item → lands on `/votes/{...}`.
2. Panel 2 first item → lands on `/anforanden/{id}` showing date + politician + speech text.
3. Panel 4 any party card with a speech → same `/anforanden/{id}` route.

- [ ] **Step 6: Empty-state verification**

In Chrome DevTools network tab, throttle `/api/speeches/recent?limit=100` to return `[]` (right-click request → "Block request URL"). Reload `/`. Confirm:
- Panel 2 shows the italic `Inga registrerade debatter ännu.` line.
- Panel 4 renders 8 cards each with `Inget anförande senaste 30 dagarna`.

Unblock and reload to restore.

- [ ] **Step 7: Final note**

If verification surfaces any unforeseen overflow or routing bug, fix it in a follow-up commit on `main`. The plan does not bundle additional fixes here because the scope is the panel rewrite.

---

## Self-review notes

- **Spec coverage:** Hero (Tasks 11), Panel 1 (11), Panel 2 (11 + 6/7 for hook + endpoint), Panel 3 (11), Panel 4 (9 + 11), Panel 5 (11), backend gaps (1-5), routing (10), TypeScript regen (5). All covered.
- **Schema deviation:** documented at the top — `date` not `occurred_at`. Item shape adjusted to `DD/MM` instead of `HH:MM`.
- **Empty states:** verified per panel with copy.
- **Mobile order 1,2,3,5,4:** achieved by placing panel 4 last in JSX after both 2-col rows; the desktop grid still pairs 1+2 and 3+5 correctly.
- **No tests:** project has no test framework (per CLAUDE.md). Plan substitutes `go build`, `npx tsc --noEmit`, `npm run lint`, `curl`, and Playwright visual checks.
