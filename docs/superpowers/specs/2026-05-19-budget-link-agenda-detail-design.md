# Budget Link + Agenda Detail Pages — Design Spec

**Date:** 2026-05-19
**Status:** Approved

## Problem

The `/riksdag` page shows budget and agenda summaries but provides no navigation to detailed views:

1. **Budget card** shows a donut chart + horizontal bars but has no link to the full `/budget` page with year-over-year comparisons
2. **Agenda items** are expandable inline (title + description + status) but have no dedicated detail page for sharing or deeper context

## Architecture

### Backend changes

**1. Add `GetAgendaItem` to repository**

File: `backend/internal/riksdag/adapters/postgres/agenda_repository.go`

Add method:
```go
func (r *AgendaRepository) GetAgendaItem(ctx context.Context, id int) (*domain.AgendaItem, error)
```

Query:
```sql
SELECT id, title, description, source, status
FROM riksdag_agenda
WHERE id = $1
```

Returns `ErrNotFound` when no rows.

**2. Update `AgendaRepository` interface**

File: `backend/internal/riksdag/ports/external.go`

Add method to interface:
```go
type AgendaRepository interface {
	ListAgenda(ctx context.Context) ([]domain.AgendaItem, error)
	GetAgendaItem(ctx context.Context, id int) (*domain.AgendaItem, error)  // NEW
}
```

**3. Add `GetAgendaItem` to service**

File: `backend/internal/riksdag/service.go`

Add method:
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

Note: `ErrNotFound` is reused (currently says "authority not found" but functions as generic 404).

**3. Add `GET /riksdag/agenda/{id}` endpoint**

File: `backend/internal/riksdag/adapters/http/handler.go`

Add route:
```go
r.Get("/riksdag/agenda/{id}", h.getAgendaItem)
```

Handler parses `id` param, calls `svc.GetAgendaItem`, returns 404 if not found.

**4. Update `api/openapi.yaml`**

Add `GET /riksdag/agenda/{id}` path with `RiksdagAgendaItem` schema (reuses existing fields: `id`, `title`, `description`, `source`, `status`).

### Frontend changes

**1. Add `id` to `AgendaItem` type**

File: `frontend/src/types/democracy.ts`

```typescript
export interface AgendaItem {
  id: number;  // NEW
  title: string;
  description: string;
  source: string;
  status: "active" | "in_progress" | "completed";
}
```

**2. Update `AgendaList` component**

File: `frontend/src/components/AgendaList.tsx`

- Make title a `<Link to={`/agenda/${item.id}`}>` instead of plain text
- Keep expand/collapse on the chevron (▼/▲) or the whole row (decide in implementation)
- Preserve existing styling

**3. Add `getAgendaItem` API method**

File: `frontend/src/features/riksdag/api.ts`

```typescript
getAgendaItem: (id: number) => fetch(`/api/riksdag/agenda/${id}`)...
```

**4. Create `AgendaDetailPage`**

File: `frontend/src/features/riksdag/AgendaDetailPage.tsx`

Layout (top → bottom):

| Block | Content |
|---|---|
| Breadcrumb | `← Riksdagen` link |
| Title | Large serif heading: agenda item title |
| Status badge | Colored badge (active/in_progress/completed) |
| Description | Full description text |
| Source | "Källa: {source}" with `SourceMarker` |
| Related | Optional: other agenda items |

Uses `useQuery` with `["agenda-item", id]` key.

**5. Add route**

File: `frontend/src/App.tsx`

```tsx
<Route path="/agenda/:id" element={<AgendaDetailPage />} />
```

**6. Add budget link**

File: `frontend/src/features/riksdag/RiksdagPage.tsx`

In the BUDGET card (line 444-484), add below the chart:

```tsx
<div style={{ marginTop: 16, textAlign: "right" }}>
  <Link to="/budget" style={{ fontSize: 13, color: "var(--color-accent-2)" }}>
    Se detaljerad budget →
  </Link>
</div>
```

## Data Flow

```
User clicks agenda title
  → React Router navigates to /agenda/:id
  → AgendaDetailPage mounts
  → useQuery calls GET /api/riksdag/agenda/:id
  → Backend: handler → service → repository → PostgreSQL
  → Renders detail page
```

## Error Handling

- **404**: Show "Agendapunkt hittades inte" with link back to `/riksdag`
- **500**: Show generic error with retry button
- **Loading**: Skeleton similar to existing pages

## Files Changed

### Backend
- `backend/internal/riksdag/adapters/postgres/agenda_repository.go` — add `GetAgendaItem`
- `backend/internal/riksdag/ports/repository.go` — add interface method
- `backend/internal/riksdag/service.go` — add `GetAgendaItem`
- `backend/internal/riksdag/adapters/http/handler.go` — add route + handler
- `api/openapi.yaml` — add path + schema

### Frontend
- `frontend/src/types/democracy.ts` — add `id` to `AgendaItem`
- `frontend/src/components/AgendaList.tsx` — make titles clickable links
- `frontend/src/features/riksdag/api.ts` — add `getAgendaItem`
- `frontend/src/features/riksdag/AgendaDetailPage.tsx` — new page
- `frontend/src/App.tsx` — add route
- `frontend/src/features/riksdag/RiksdagPage.tsx` — add budget link

## Testing

- [ ] `go build -C backend ./...` passes
- [ ] `cd frontend && npx tsc --noEmit` passes
- [ ] Agenda item links navigate to detail page
- [ ] Budget link navigates to `/budget`
- [ ] 404 for non-existent agenda IDs

## OpenAPI Schema

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
```
