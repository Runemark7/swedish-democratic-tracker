# Budget Link + Agenda Detail Pages Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a link from the Riksdag budget card to `/budget` and create dedicated detail pages for agenda items at `/agenda/:id`.

**Architecture:** Backend: extend AgendaRepository interface + implementation, add service method, add HTTP handler. Frontend: add `id` to AgendaItem type, make agenda titles linkable, create detail page component, add route.

**Tech Stack:** Go 1.26.1 (chi, pgx), React 19 + TypeScript + React Router, OpenAPI

---

## File Structure

### Backend
- `backend/internal/riksdag/ports/external.go` — Add `GetAgendaItem` to interface
- `backend/internal/riksdag/adapters/postgres/agenda_repository.go` — Implement `GetAgendaItem`
- `backend/internal/riksdag/service.go` — Add `GetAgendaItem` service method
- `backend/internal/riksdag/adapters/http/handler.go` — Add `GET /riksdag/agenda/{id}` route + handler
- `api/openapi.yaml` — Add `RiksdagAgendaItem` schema + `GET /riksdag/agenda/{id}` path

### Frontend
- `frontend/src/types/democracy.ts` — Add `id: number` to `AgendaItem`
- `frontend/src/features/riksdag/api.ts` — Add `getAgendaItem(id)` API method
- `frontend/src/components/AgendaList.tsx` — Wrap titles in `<Link>` to `/agenda/:id`
- `frontend/src/features/riksdag/AgendaDetailPage.tsx` — New detail page component
- `frontend/src/App.tsx` — Add `<Route path="/agenda/:id">`
- `frontend/src/features/riksdag/RiksdagPage.tsx` — Add "Se detaljerad budget →" link

---

## Task 1: Extend AgendaRepository Interface

**Files:**
- Modify: `backend/internal/riksdag/ports/external.go:85-87`

- [ ] **Step 1: Add `GetAgendaItem` to interface**

```go
type AgendaRepository interface {
	ListAgenda(ctx context.Context) ([]domain.AgendaItem, error)
	GetAgendaItem(ctx context.Context, id int) (*domain.AgendaItem, error)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/riksdag/ports/external.go
git commit -m "feat(riksdag): add GetAgendaItem to AgendaRepository interface"
```

---

## Task 2: Implement GetAgendaItem in Postgres Repository

**Files:**
- Modify: `backend/internal/riksdag/adapters/postgres/agenda_repository.go`

- [ ] **Step 1: Add import for pgx errors**

Add to imports:
```go
"github.com/jackc/pgx/v5"
```

- [ ] **Step 2: Add GetAgendaItem method**

Add after `ListAgenda`:
```go
func (r *AgendaRepository) GetAgendaItem(ctx context.Context, id int) (*domain.AgendaItem, error) {
	const q = `
		SELECT id, title, description, source, status
		FROM riksdag_agenda
		WHERE id = $1
	`
	var a domain.AgendaItem
	err := r.db.QueryRow(ctx, q, id).Scan(&a.ID, &a.Title, &a.Description, &a.Source, &a.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &a, nil
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/riksdag/adapters/postgres/agenda_repository.go
git commit -m "feat(riksdag): implement GetAgendaItem in postgres repository"
```

---

## Task 3: Add GetAgendaItem to Service

**Files:**
- Modify: `backend/internal/riksdag/service.go`

- [ ] **Step 1: Add GetAgendaItem method**

Add after `ListAgenda`:
```go
func (s *Service) GetAgendaItem(ctx context.Context, id int) (*domain.AgendaItem, error) {
	if s.agendaRepo == nil {
		return nil, errors.New("agenda repository not configured")
	}
	item, err := s.agendaRepo.GetAgendaItem(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return item, nil
}
```

- [ ] **Step 2: Add pgx import**

Add to imports:
```go
"github.com/jackc/pgx/v5"
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/riksdag/service.go
git commit -m "feat(riksdag): add GetAgendaItem service method"
```

---

## Task 4: Add HTTP Handler for GET /riksdag/agenda/{id}

**Files:**
- Modify: `backend/internal/riksdag/adapters/http/handler.go`

- [ ] **Step 1: Add route**

In `Routes`, add:
```go
r.Get("/riksdag/agenda/{id}", h.getAgendaItem)
```

- [ ] **Step 2: Add handler**

Add after `getAgenda`:
```go
func (h *Handler) getAgendaItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		jsonError(w, "invalid agenda id", http.StatusBadRequest)
		return
	}

	item, err := h.svc.GetAgendaItem(r.Context(), id)
	if err != nil {
		if errors.Is(err, riksdag.ErrNotFound) {
			jsonError(w, "agenda item not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to fetch agenda item", http.StatusInternalServerError)
		return
	}
	jsonOK(w, item)
}
```

- [ ] **Step 3: Add strconv import**

Add to imports:
```go
"strconv"
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/riksdag/adapters/http/handler.go
git commit -m "feat(riksdag): add GET /riksdag/agenda/{id} endpoint"
```

---

## Task 5: Update OpenAPI Spec

**Files:**
- Modify: `api/openapi.yaml`

- [ ] **Step 1: Add RiksdagAgendaItem schema**

Add to `components/schemas`:
```yaml
RiksdagAgendaItem:
  type: object
  properties:
    id:
      type: integer
    title:
      type: string
    description:
      type: string
    source:
      type: string
    status:
      type: string
      enum: [active, in_progress, completed]
  required: [id, title, description, source, status]
```

- [ ] **Step 2: Add GET /riksdag/agenda/{id} path**

Add to `paths`:
```yaml
/riksdag/agenda/{id}:
  get:
    summary: Get a single agenda item
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: integer
    responses:
      '200':
        description: Agenda item details
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/RiksdagAgendaItem'
      '404':
        description: Agenda item not found
```

- [ ] **Step 3: Commit**

```bash
git add api/openapi.yaml
git commit -m "docs(api): add RiksdagAgendaItem schema and GET /riksdag/agenda/{id}"
```

---

## Task 6: Add id to AgendaItem Type

**Files:**
- Modify: `frontend/src/types/democracy.ts`

- [ ] **Step 1: Add id field**

Change:
```typescript
export interface AgendaItem {
  id: number;  // NEW
  title: string;
  description: string;
  source: string;
  status: "active" | "in_progress" | "completed";
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/types/democracy.ts
git commit -m "feat(types): add id field to AgendaItem"
```

---

## Task 7: Add getAgendaItem API Method

**Files:**
- Modify: `frontend/src/features/riksdag/api.ts`

- [ ] **Step 1: Add method to riksdagApi**

Add to `riksdagApi` object:
```typescript
getAgendaItem: (id: number) =>
  fetch(`/api/riksdag/agenda/${id}`).then((r) => {
    if (!r.ok) throw new Error(`riksdag/agenda/${id}: ${r.status}`);
    return r.json() as Promise<AgendaItem>;
  }),
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/features/riksdag/api.ts
git commit -m "feat(api): add getAgendaItem method"
```

---

## Task 8: Make AgendaList Titles Clickable

**Files:**
- Modify: `frontend/src/components/AgendaList.tsx`

- [ ] **Step 1: Add Link import**

```typescript
import { Link } from "react-router-dom";
```

- [ ] **Step 2: Wrap title in Link**

Change the title span to:
```tsx
<span style={{ flex: 1, fontSize: 14, color: "var(--color-fg)", lineHeight: 1.4 }}>
  <Link to={`/agenda/${item.id}`} style={{ color: "inherit", textDecoration: "none", borderBottom: "1px dotted var(--color-border)" }}>
    {item.title}
  </Link>
</span>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/AgendaList.tsx
git commit -m "feat(ui): make agenda titles link to detail pages"
```

---

## Task 9: Create AgendaDetailPage Component

**Files:**
- Create: `frontend/src/features/riksdag/AgendaDetailPage.tsx`

- [ ] **Step 1: Create the component**

```tsx
import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { riksdagApi } from "./api";
import { SourceMarker } from "@/components/sources/SourceMarker";

const STATUS_LABEL: Record<string, string> = {
  active: "Aktiv",
  in_progress: "Pågående",
  completed: "Genomförd",
};

const STATUS_COLOR: Record<string, string> = {
  active: "var(--color-accent-2)",
  in_progress: "#5a9fd0",
  completed: "#4caf50",
};

export function AgendaDetailPage() {
  const { id } = useParams<{ id: string }>();
  const numId = id ? parseInt(id, 10) : 0;

  const { data: item, isLoading, error } = useQuery({
    queryKey: ["agenda-item", numId],
    queryFn: () => riksdagApi.getAgendaItem(numId),
    enabled: numId > 0,
  });

  if (isLoading) {
    return (
      <div className="sdt-page" style={{ padding: "40px 32px" }}>
        <div style={{ height: 200, background: "var(--color-track)", borderRadius: 4 }} />
      </div>
    );
  }

  if (error || !item) {
    return (
      <div className="sdt-page" style={{ padding: "40px 32px" }}>
        <div style={{ fontSize: 14, color: "var(--color-fg-muted)" }}>
          Agendapunkt hittades inte.
        </div>
        <Link to="/riksdag" style={{ fontSize: 13, color: "var(--color-accent-2)", marginTop: 12, display: "inline-block" }}>
          ← Tillbaka till Riksdagen
        </Link>
      </div>
    );
  }

  return (
    <div className="sdt-page" style={{ padding: isMobile ? "18px 14px" : "40px 32px" }}>
      {/* Breadcrumb */}
      <div style={{ marginBottom: 24 }}>
        <Link to="/riksdag" style={{ fontSize: 13, color: "var(--color-accent-2)", textDecoration: "none" }}>
          ← Riksdagen
        </Link>
      </div>

      {/* Title */}
      <h1 style={{
        fontFamily: "var(--font-serif)",
        fontSize: isMobile ? 28 : 42,
        fontWeight: 400,
        letterSpacing: "-1px",
        color: "var(--color-fg)",
        marginBottom: 16,
        lineHeight: 1.1,
      }}>
        {item.title}
      </h1>

      {/* Status badge */}
      <div style={{ marginBottom: 24 }}>
        <span style={{
          fontSize: 11,
          fontFamily: "var(--font-mono)",
          letterSpacing: "1px",
          textTransform: "uppercase",
          color: STATUS_COLOR[item.status] || "var(--color-fg-muted)",
          border: `1px solid ${STATUS_COLOR[item.status] || "var(--color-fg-muted)"}`,
          padding: "4px 10px",
          borderRadius: 2,
        }}>
          {STATUS_LABEL[item.status] || item.status}
        </span>
      </div>

      {/* Description */}
      <div style={{
        fontSize: 15,
        lineHeight: 1.6,
        color: "var(--color-fg)",
        marginBottom: 32,
        maxWidth: 640,
      }}>
        {item.description}
      </div>

      {/* Source */}
      <div style={{
        fontSize: 12,
        color: "var(--color-fg-muted)",
        fontFamily: "var(--font-mono)",
        borderTop: "1px solid var(--color-border)",
        paddingTop: 16,
      }}>
        Källa: {item.source}
        <SourceMarker sourceId="derived-agenda" />
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/features/riksdag/AgendaDetailPage.tsx
git commit -m "feat(riksdag): create AgendaDetailPage component"
```

---

## Task 10: Add Route for /agenda/:id

**Files:**
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: Import AgendaDetailPage**

Add import:
```typescript
import { AgendaDetailPage } from "./features/riksdag/AgendaDetailPage";
```

- [ ] **Step 2: Add route**

Add to routes:
```tsx
<Route path="/agenda/:id" element={<AgendaDetailPage />} />
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/App.tsx
git commit -m "feat(routing): add /agenda/:id route"
```

---

## Task 11: Add Budget Link to RiksdagPage

**Files:**
- Modify: `frontend/src/features/riksdag/RiksdagPage.tsx`

- [ ] **Step 1: Add Link import**

Ensure `Link` is imported (already imported at top).

- [ ] **Step 2: Add budget link**

After the budget chart section (after `</div>` that closes the flex container at line ~483), add:
```tsx
<div style={{ marginTop: 16, textAlign: "right" }}>
  <Link to="/budget" style={{ fontSize: 13, color: "var(--color-accent-2)", textDecoration: "none", borderBottom: "1px dotted var(--color-border)" }}>
    Se detaljerad budget →
  </Link>
</div>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/features/riksdag/RiksdagPage.tsx
git commit -m "feat(riksdag): add link to detailed budget page"
```

---

## Task 12: Verification

- [ ] **Step 1: Build backend**

```bash
go build -C backend ./...
```
Expected: No errors

- [ ] **Step 2: Typecheck frontend**

```bash
cd frontend && npx tsc --noEmit
```
Expected: No errors

- [ ] **Step 3: Verify no missing pieces**

Checklist:
- [ ] Agenda items show `id` in API response
- [ ] `/agenda/1` route works
- [ ] Budget link visible on `/riksdag`
- [ ] 404 shown for invalid agenda IDs

- [ ] **Step 4: Final commit**

```bash
git commit --allow-empty -m "feat: budget link + agenda detail pages complete"
```

---

## Spec Coverage Check

| Spec Requirement | Task |
|---|---|
| Budget card link to `/budget` | Task 11 |
| Agenda items have clickable titles | Task 8 |
| `GET /riksdag/agenda/{id}` endpoint | Tasks 1-4 |
| `id` field in AgendaItem type | Task 6 |
| AgendaDetailPage component | Task 9 |
| `/agenda/:id` route | Task 10 |
| OpenAPI schema updated | Task 5 |
| Error handling (404) | Task 9 |
| Build verification | Task 12 |

## Placeholder Scan

- No TBD/TODO/fill-in found
- All code blocks contain complete implementations
- All file paths are exact
- All commands have expected output

## Type Consistency

- `id` is `int` in Go, `number` in TypeScript
- `status` enum matches: `[active, in_progress, completed]`
- `ErrNotFound` reused from existing service error
- `pgx.ErrNoRows` properly handled
