# Debate Speakers + Party Detail Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Connect the existing speech corpus to three pages: surface debate speakers on `/beslut/:beteckning`, list a politician's anföranden on `/politicians/:id`, and add a brand-new `/parties/:party` page (hero + Mål · Anföranden · Politiker tabs).

**Architecture:** Three new HTTP endpoints on the existing speeches handler (`/speeches/by-document/:dokId`, `/speeches/by-politician/:intressentId`, `/speeches/by-party/:party`). One shared frontend component (`SpeechRow`) renders identical rows on all three pages. New frontend route `/parties/:party` (`PartyDetailPage`) adds a tab strip; existing `/parties/:party/goals` stays alive.

**Tech Stack:** Go 1.26 (chi, pgx/v5), React 19 + Vite + TypeScript, TanStack Query v5. No new dependencies.

**Spec:** `docs/superpowers/specs/2026-05-09-debate-speakers-and-party-detail-design.md`. Read it before Task 1.

**Branching:** Work directly on `main`. Push at the end of each task.

---

## File map

**New files**

| Path | Responsibility |
|---|---|
| `frontend/src/features/speeches/SpeechRow.tsx` | Shared row component (avatar + name + party chip + date + 2-line snippet + "Läs hela →"). |
| `frontend/src/features/parties/PartyDetailPage.tsx` | New `/parties/:party` route. Hero + 3-tab interface. |

**Modified files**

| Path | Change |
|---|---|
| `backend/internal/speeches/ports/repository.go` | Add `ListByDocument` + `ListByParty` methods. |
| `backend/internal/speeches/adapters/postgres/repository.go` | Implement both new queries. |
| `backend/internal/speeches/service.go` | Pass-through `ListByDocument`, `ListByParty`, expose existing `ListByPolitician`. |
| `backend/internal/speeches/adapters/http/handler.go` | Add 3 routes; extend `PoliticianLookup` with `ImageURLByID`; add `PoliticianImageURL` to `SpeechDTO`. |
| `backend/internal/politicians/service.go` | Add `ImageURLByID`. |
| `api/openapi.yaml` | Add 3 new paths + optional `politicianImageUrl` on `Speech` schema. |
| `frontend/src/shared/api-contract.ts` | Regenerated. |
| `frontend/src/features/speeches/api.ts` | Add `politicianImageUrl` to `Speech`; add `listByDocument`, `listByPolitician`, `listByParty`. |
| `frontend/src/hooks/useDemocracy.ts` | Add `useSpeechesByDocument`, `useSpeechesByPolitician`, `useSpeechesByParty` hooks. |
| `frontend/src/features/votes/BeslutDetailPage.tsx` | Insert `SAMMANFATTNING` block + `DEBATTEN` speakers section above the existing HELA BETÄNKANDET render. |
| `frontend/src/features/politicians/PoliticianPage.tsx` | Add `anforanden` as 3rd tab (existing tabs: `votes`, `promises`). |
| `frontend/src/App.tsx` | Add `<Route path="/parties/:party" element={<PartyDetailPage />} />`. |
| `frontend/src/features/parties/PartiesPage.tsx` | Each party card already links somewhere — verify it links to `/parties/:party` (not `/parties/:party/goals`); update if needed. |

---

## Task 1: Backend — `ListByDocument` + `ListByParty` (port + Postgres)

**Files:**
- Modify: `backend/internal/speeches/ports/repository.go`
- Modify: `backend/internal/speeches/adapters/postgres/repository.go`

- [ ] **Step 1: Extend the port interface**

In `backend/internal/speeches/ports/repository.go`, replace the existing `SpeechRepository` interface with:

```go
type SpeechRepository interface {
	GetByID(ctx context.Context, id int) (*domain.Speech, error)
	ListByPolitician(ctx context.Context, politicianID string) ([]*domain.Speech, error)
	ListRecent(ctx context.Context, limit int) ([]*domain.Speech, error)
	UpsertMany(ctx context.Context, ss []*domain.Speech) error
	ListUnprocessed(ctx context.Context, limit int) ([]*domain.Speech, error)
	MarkProcessed(ctx context.Context, id int) error
	// ListMissingText returns speeches with empty speech_text, ordered
	// newest first so user-facing pages get backfilled first.
	ListMissingText(ctx context.Context, limit int) ([]*domain.Speech, error)
	// UpdateText overwrites speech_text for one row.
	UpdateText(ctx context.Context, id int, text string) error
	// ListByDocument returns speeches whose rel_dok_id matches the
	// given betänkande document id, ordered by anforande_nummer ASC
	// (chronological within the debate).
	ListByDocument(ctx context.Context, dokID string) ([]*domain.Speech, error)
	// ListByParty returns speeches by all members of one party,
	// newest first, capped at limit (default 50, max 200).
	ListByParty(ctx context.Context, party string, limit int) ([]*domain.Speech, error)
}
```

- [ ] **Step 2: Implement `ListByDocument` in Postgres repository**

In `backend/internal/speeches/adapters/postgres/repository.go`, append after the existing `UpdateText` method:

```go
func (r *Repository) ListByDocument(ctx context.Context, dokID string) ([]*domain.Speech, error) {
	if dokID == "" {
		return []*domain.Speech{}, nil
	}
	q := "SELECT " + selectCols + " FROM speeches " +
		"WHERE related_dok_id = $1 " +
		"ORDER BY anforande_nummer::int ASC, id ASC"
	rows, err := r.db.Query(ctx, q, dokID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSpeeches(rows)
}
```

`anforande_nummer::int` casts the text-typed column to int so numeric ordering applies (otherwise `"10"` sorts before `"2"`).

- [ ] **Step 3: Implement `ListByParty`**

Append:

```go
func (r *Repository) ListByParty(ctx context.Context, party string, limit int) ([]*domain.Speech, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if party == "" {
		return []*domain.Speech{}, nil
	}
	q := "SELECT " + selectCols + " FROM speeches " +
		"WHERE party = $1 " +
		"ORDER BY date DESC, id DESC LIMIT $2"
	rows, err := r.db.Query(ctx, q, party, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSpeeches(rows)
}
```

- [ ] **Step 4: Verify backend compiles**

```bash
go build -C backend ./...
```

Expected: empty output, exit 0.

- [ ] **Step 5: Commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add backend/internal/speeches/ports/repository.go backend/internal/speeches/adapters/postgres/repository.go
git commit -m "feat(speeches): add ListByDocument + ListByParty repository methods"
git push origin main
```

---

## Task 2: Backend — service pass-throughs

**Files:**
- Modify: `backend/internal/speeches/service.go`

- [ ] **Step 1: Add two pass-through methods**

After the existing `EnrichMissingText` method, append:

```go
func (s *Service) ListByDocument(ctx context.Context, dokID string) ([]*domain.Speech, error) {
	return s.repo.ListByDocument(ctx, dokID)
}

func (s *Service) ListByParty(ctx context.Context, party string, limit int) ([]*domain.Speech, error) {
	return s.repo.ListByParty(ctx, party, limit)
}
```

(`ListByPolitician` already exists — no change needed.)

- [ ] **Step 2: Verify build**

```bash
go build -C backend ./...
```

- [ ] **Step 3: Commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add backend/internal/speeches/service.go
git commit -m "feat(speeches): expose ListByDocument + ListByParty on service"
git push origin main
```

---

## Task 3: Backend — politicians service `ImageURLByID`

**Files:**
- Modify: `backend/internal/politicians/service.go`

- [ ] **Step 1: Add the method**

After the existing `NameByID`:

```go
func (s *Service) ImageURLByID(ctx context.Context, intressentID string) (string, error) {
	p, err := s.repo.GetByID(ctx, intressentID)
	if err != nil {
		return "", err
	}
	if p == nil {
		return "", nil
	}
	return p.ImageURL, nil
}
```

- [ ] **Step 2: Build + commit**

```bash
go build -C backend ./...

cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add backend/internal/politicians/service.go
git commit -m "feat(politicians): expose ImageURLByID on service"
git push origin main
```

---

## Task 4: Backend — speeches HTTP handler (3 new routes + image URL on DTO)

**Files:**
- Modify: `backend/internal/speeches/adapters/http/handler.go`

- [ ] **Step 1: Extend `PoliticianLookup` interface and `SpeechDTO`**

At the top of `handler.go`, modify the existing `PoliticianLookup` interface:

```go
type PoliticianLookup interface {
	NameByID(ctx context.Context, intressentID string) (string, error)
	ImageURLByID(ctx context.Context, intressentID string) (string, error)
}
```

Add `PoliticianImageURL` to `SpeechDTO`:

```go
type SpeechDTO struct {
	ID                 int       `json:"id"`
	DokID              string    `json:"dokId"`
	AnforandeNummer    string    `json:"anforandeNummer"`
	PoliticianID       string    `json:"politicianId"`
	PoliticianName     string    `json:"politicianName"`
	PoliticianImageURL string    `json:"politicianImageUrl,omitempty"`
	Party              string    `json:"party"`
	Date               time.Time `json:"date"`
	TopicHeading       string    `json:"topicHeading"`
	Snippet            string    `json:"snippet"`
	SpeechText         string    `json:"speechText,omitempty"`
}
```

- [ ] **Step 2: Populate `PoliticianImageURL` in `toDTO`**

In the existing `toDTO` method, after fetching `name`:

```go
func (h *Handler) toDTO(ctx context.Context, s *domain.Speech, includeFullText bool) SpeechDTO {
	name, _ := h.pol.NameByID(ctx, s.PoliticianID)
	imageURL, _ := h.pol.ImageURLByID(ctx, s.PoliticianID)
	text := s.SpeechText
	if strings.TrimSpace(text) == "" {
		text = ""
	}
	dto := SpeechDTO{
		ID:                 s.ID,
		DokID:              s.DokID,
		AnforandeNummer:    s.AnforandeNummer,
		PoliticianID:       s.PoliticianID,
		PoliticianName:     name,
		PoliticianImageURL: imageURL,
		Party:              s.Party,
		Date:               s.Date,
		TopicHeading:       s.TopicHeading,
		Snippet:            snippetFrom(stripHTML(text), 180),
	}
	if includeFullText {
		dto.SpeechText = text
	}
	return dto
}
```

- [ ] **Step 3: Add 3 new routes**

In the `Routes` method:

```go
func (h *Handler) Routes(r chi.Router) {
	r.Get("/speeches/recent", h.listRecent)
	r.Get("/speeches/{id}", h.getByID)
	r.Get("/speeches/by-document/{dokId}", h.listByDocument)
	r.Get("/speeches/by-politician/{intressentId}", h.listByPolitician)
	r.Get("/speeches/by-party/{party}", h.listByParty)
}
```

- [ ] **Step 4: Add the 3 handler functions**

After the existing `getByID` method, append:

```go
func (h *Handler) listByDocument(w http.ResponseWriter, r *http.Request) {
	dokID := chi.URLParam(r, "dokId")
	ss, err := h.svc.ListByDocument(r.Context(), dokID)
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

func (h *Handler) listByPolitician(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "intressentId")
	ss, err := h.svc.ListByPolitician(r.Context(), id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	limit := 20
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if len(ss) > limit {
		ss = ss[:limit]
	}
	out := make([]SpeechDTO, 0, len(ss))
	for _, s := range ss {
		out = append(out, h.toDTO(r.Context(), s, false))
	}
	jsonOK(w, out)
}

func (h *Handler) listByParty(w http.ResponseWriter, r *http.Request) {
	party := chi.URLParam(r, "party")
	limit := 50
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			limit = n
		}
	}
	ss, err := h.svc.ListByParty(r.Context(), party, limit)
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
```

- [ ] **Step 5: Build + smoke-test**

```bash
go build -C backend ./...
```

Expected: empty output. (Run `make run` in another terminal if you want to curl-test the endpoints.)

- [ ] **Step 6: Commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add backend/internal/speeches/adapters/http/handler.go
git commit -m "feat(speeches): add /speeches/by-{document,politician,party} routes + politicianImageUrl on DTO"
git push origin main
```

---

## Task 5: OpenAPI + regen frontend types

**Files:**
- Modify: `api/openapi.yaml`

- [ ] **Step 1: Add 3 path entries**

Find the existing `/speeches/recent` and `/speeches/{id}` entries in `api/openapi.yaml`. Append three more inside `paths:`:

```yaml
  /speeches/by-document/{dokId}:
    get:
      summary: Speeches whose rel_dok_id matches a betänkande
      tags: [speeches]
      parameters:
        - in: path
          name: dokId
          required: true
          schema:
            type: string
      responses:
        "200":
          description: List of speeches in chronological debate order
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/Speech"
  /speeches/by-politician/{intressentId}:
    get:
      summary: Recent speeches by one politician
      tags: [speeches]
      parameters:
        - in: path
          name: intressentId
          required: true
          schema:
            type: string
        - in: query
          name: limit
          schema:
            type: integer
            default: 20
            minimum: 1
            maximum: 200
      responses:
        "200":
          description: Speeches by one politician, newest first
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/Speech"
  /speeches/by-party/{party}:
    get:
      summary: Recent speeches by all members of one party
      tags: [speeches]
      parameters:
        - in: path
          name: party
          required: true
          schema:
            type: string
        - in: query
          name: limit
          schema:
            type: integer
            default: 50
            minimum: 1
            maximum: 200
      responses:
        "200":
          description: Speeches by party members, newest first
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/Speech"
```

- [ ] **Step 2: Add `politicianImageUrl` to the `Speech` schema**

Find the existing `Speech:` schema under `components: schemas:`. Add one property after `politicianName`:

```yaml
        politicianImageUrl:
          type: string
          description: URL to politician portrait (Riksdagen CDN). May be empty.
```

- [ ] **Step 3: Regenerate frontend types**

```bash
cd /home/rune/Documents/swedish-democratic-tracker/frontend
npm run generate:api
```

Expected output ends with `🚀 ../api/openapi.yaml → src/shared/api-contract.ts`.

- [ ] **Step 4: Commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add api/openapi.yaml frontend/src/shared/api-contract.ts
git commit -m "chore(api): add /speeches/by-{document,politician,party} + politicianImageUrl"
git push origin main
```

---

## Task 6: Frontend — `speechesApi` + `Speech` type

**Files:**
- Modify: `frontend/src/features/speeches/api.ts`

- [ ] **Step 1: Replace contents**

Replace the file with:

```ts
import { api } from "@/shared/api-client";

export interface Speech {
  id: number;
  dokId: string;
  anforandeNummer?: string;
  politicianId: string;
  politicianName: string;
  politicianImageUrl?: string;
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
  listByDocument: (dokId: string) =>
    api.get<Speech[]>(`/speeches/by-document/${encodeURIComponent(dokId)}`),
  listByPolitician: (intressentId: string, limit = 20) =>
    api.get<Speech[]>(`/speeches/by-politician/${encodeURIComponent(intressentId)}?limit=${limit}`),
  listByParty: (party: string, limit = 50) =>
    api.get<Speech[]>(`/speeches/by-party/${encodeURIComponent(party)}?limit=${limit}`),
};
```

- [ ] **Step 2: Verify typecheck**

```bash
cd /home/rune/Documents/swedish-democratic-tracker/frontend
npx tsc --noEmit
```

Expected: empty output.

- [ ] **Step 3: Commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add frontend/src/features/speeches/api.ts
git commit -m "feat(speeches): add listByDocument/Politician/Party + politicianImageUrl"
git push origin main
```

---

## Task 7: Frontend — three new query hooks

**Files:**
- Modify: `frontend/src/hooks/useDemocracy.ts`

- [ ] **Step 1: Append three hooks at the bottom of the file**

Find the existing `useRecentSpeeches` hook. After it, add:

```ts
export function useSpeechesByDocument(dokId: string | undefined) {
  return useQuery<Speech[]>({
    queryKey: ["speeches-by-document", dokId],
    queryFn: () => speechesApi.listByDocument(dokId ?? ""),
    enabled: !!dokId,
    staleTime: 60_000,
  });
}

export function useSpeechesByPolitician(
  intressentId: string | undefined,
  limit = 20,
) {
  return useQuery<Speech[]>({
    queryKey: ["speeches-by-politician", intressentId, limit],
    queryFn: () => speechesApi.listByPolitician(intressentId ?? "", limit),
    enabled: !!intressentId,
    staleTime: 60_000,
  });
}

export function useSpeechesByParty(
  party: string | undefined,
  limit = 50,
) {
  return useQuery<Speech[]>({
    queryKey: ["speeches-by-party", party, limit],
    queryFn: () => speechesApi.listByParty(party ?? "", limit),
    enabled: !!party,
    staleTime: 60_000,
  });
}
```

- [ ] **Step 2: Typecheck + commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker/frontend
npx tsc --noEmit

cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add frontend/src/hooks/useDemocracy.ts
git commit -m "feat(speeches): add useSpeechesBy{Document,Politician,Party} hooks"
git push origin main
```

---

## Task 8: Frontend — `<SpeechRow>` shared component

**Files:**
- Create: `frontend/src/features/speeches/SpeechRow.tsx`

- [ ] **Step 1: Create the component**

```tsx
import { Link } from "react-router-dom";
import { PARTY_COLORS } from "@/shared/design";
import type { Speech } from "@/features/speeches/api";

interface SpeechRowProps {
  speech: Speech;
  /** Hide politician name link if rendering on that politician's own page. */
  hidePolitician?: boolean;
}

function formatStamp(iso: string): string {
  const d = new Date(iso);
  const dd = String(d.getDate()).padStart(2, "0");
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  return `${dd}/${mm}`;
}

export function SpeechRow({ speech, hidePolitician }: SpeechRowProps) {
  const pc = PARTY_COLORS[speech.party];
  const initials =
    (speech.politicianName || "?")
      .split(" ")
      .filter(Boolean)
      .slice(0, 2)
      .map((n) => n[0])
      .join("") || "?";

  return (
    <article
      style={{
        display: "flex",
        gap: 12,
        padding: "12px 0",
        borderBottom: "1px solid var(--color-border)",
        alignItems: "flex-start",
        minWidth: 0,
      }}
    >
      {/* Avatar */}
      {speech.politicianImageUrl ? (
        <img
          src={speech.politicianImageUrl}
          alt=""
          loading="lazy"
          onError={(e) => {
            (e.currentTarget as HTMLImageElement).style.display = "none";
          }}
          style={{
            width: 40,
            height: 40,
            borderRadius: "50%",
            objectFit: "cover",
            flexShrink: 0,
          }}
        />
      ) : (
        <div
          style={{
            width: 40,
            height: 40,
            borderRadius: "50%",
            background: pc?.bg ?? "#888",
            color: pc?.text ?? "#fff",
            fontFamily: "var(--font-mono)",
            fontSize: 13,
            fontWeight: 700,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            flexShrink: 0,
          }}
        >
          {initials}
        </div>
      )}

      {/* Body */}
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 8,
            marginBottom: 4,
            flexWrap: "wrap",
          }}
        >
          {!hidePolitician && (
            <Link
              to={`/politicians/${encodeURIComponent(speech.politicianId)}`}
              style={{
                fontFamily: "var(--font-serif)",
                fontSize: 14,
                color: "var(--color-fg)",
                textDecoration: "none",
              }}
            >
              {speech.politicianName || "Anonym"}
            </Link>
          )}
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              fontWeight: 700,
              padding: "1px 6px",
              borderRadius: 2,
              background: pc?.bg ?? "#888",
              color: pc?.text ?? "#fff",
              letterSpacing: "0.05em",
            }}
          >
            {speech.party}
          </span>
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              color: "var(--color-fg-muted)",
              marginLeft: "auto",
            }}
          >
            {formatStamp(speech.date)}
          </span>
        </div>
        {speech.topicHeading && (
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.05em",
              marginBottom: 4,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {speech.topicHeading}
          </div>
        )}
        <p
          style={{
            fontFamily: "var(--font-body)",
            fontSize: 13,
            lineHeight: 1.5,
            color: "var(--color-fg)",
            margin: "0 0 6px",
            display: "-webkit-box",
            WebkitLineClamp: 2,
            WebkitBoxOrient: "vertical",
            overflow: "hidden",
          }}
        >
          {speech.snippet || "Anförandetexten är inte tillgänglig än."}
        </p>
        <Link
          to={`/anforanden/${speech.id}`}
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-accent)",
            textDecoration: "none",
            letterSpacing: "0.05em",
          }}
        >
          Läs hela →
        </Link>
      </div>
    </article>
  );
}
```

- [ ] **Step 2: Typecheck + commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker/frontend
npx tsc --noEmit

cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add frontend/src/features/speeches/SpeechRow.tsx
git commit -m "feat(speeches): SpeechRow shared component"
git push origin main
```

---

## Task 9: Frontend — `BeslutDetailPage` speakers section + summary block

**Files:**
- Modify: `frontend/src/features/votes/BeslutDetailPage.tsx`

- [ ] **Step 1: Add imports**

At the top of the file, alongside the existing imports:

```tsx
import { useState } from "react";
import { useSpeechesByDocument } from "@/hooks/useDemocracy";
import { SpeechRow } from "@/features/speeches/SpeechRow";
```

- [ ] **Step 2: Fetch speeches inside the component**

Inside `BeslutDetailPage`, after the existing `useQuery` for `["beslut", beteckning]`, add:

```tsx
  const dokIdForSpeeches = data?.dokId;
  const { data: speeches } = useSpeechesByDocument(dokIdForSpeeches);
  const [expandSpeakers, setExpandSpeakers] = useState(false);
```

- [ ] **Step 3: Insert `SAMMANFATTNING` block**

Find the existing block where `data?.summary` is rendered (around the existing summary section). Replace it with this version that always renders a section, with a muted empty state when missing:

```tsx
        {/* ── SAMMANFATTNING ──────────────────────────────────────── */}
        <section
          style={{
            background: "var(--color-sdt-surface)",
            border: "1px solid var(--color-border)",
            borderRadius: 4,
            padding: "16px 20px",
            marginBottom: 16,
          }}
        >
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              letterSpacing: "0.15em",
              textTransform: "uppercase",
              color: "var(--color-fg-muted)",
              marginBottom: 8,
            }}
          >
            Sammanfattning
          </div>
          {data?.summary ? (
            <p
              style={{
                fontFamily: "var(--font-body)",
                fontSize: 14,
                lineHeight: 1.6,
                color: "var(--color-fg)",
                margin: 0,
              }}
            >
              {data.summary}
            </p>
          ) : (
            <p
              style={{
                fontFamily: "var(--font-body)",
                fontSize: 13,
                fontStyle: "italic",
                color: "var(--color-fg-muted)",
                margin: 0,
              }}
            >
              Riksdagen har inte publicerat någon sammanfattning för det här beslutet.
            </p>
          )}
        </section>
```

If a previous summary block already exists, remove it.

- [ ] **Step 4: Insert `DEBATTEN` speakers section**

Place this section after the party-vote breakdown card and before the `Riksdagen link` block:

```tsx
        {/* ── DEBATTEN — speakers ─────────────────────────────────── */}
        <section
          style={{
            background: "var(--color-sdt-surface)",
            border: "1px solid var(--color-border)",
            borderRadius: 4,
            padding: "20px 24px",
            marginBottom: 16,
          }}
        >
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              letterSpacing: "0.15em",
              textTransform: "uppercase",
              color: "var(--color-fg-muted)",
              marginBottom: 12,
            }}
          >
            Debatten {speeches && speeches.length > 0 ? `· ${speeches.length} talare` : ""}
          </div>
          {(!speeches || speeches.length === 0) && (
            <p
              style={{
                fontFamily: "var(--font-body)",
                fontStyle: "italic",
                fontSize: 13,
                color: "var(--color-fg-muted)",
                margin: 0,
              }}
            >
              Inga registrerade anföranden för det här beslutet ännu.
            </p>
          )}
          {speeches && speeches.length > 0 && (
            <>
              {(expandSpeakers ? speeches : speeches.slice(0, 5)).map((s) => (
                <SpeechRow key={s.id} speech={s} />
              ))}
              {speeches.length > 5 && !expandSpeakers && (
                <button
                  onClick={() => setExpandSpeakers(true)}
                  style={{
                    marginTop: 12,
                    fontFamily: "var(--font-mono)",
                    fontSize: 11,
                    color: "var(--color-accent)",
                    background: "none",
                    border: "none",
                    cursor: "pointer",
                    letterSpacing: "0.05em",
                    padding: 0,
                  }}
                >
                  Visa alla {speeches.length} anföranden →
                </button>
              )}
            </>
          )}
        </section>
```

- [ ] **Step 5: Typecheck + commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker/frontend
npx tsc --noEmit

cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add frontend/src/features/votes/BeslutDetailPage.tsx
git commit -m "feat(beslut): SAMMANFATTNING block + DEBATTEN speakers section"
git push origin main
```

---

## Task 10: Frontend — `PoliticianPage` anföranden tab

**Files:**
- Modify: `frontend/src/features/politicians/PoliticianPage.tsx`

- [ ] **Step 1: Add imports**

```tsx
import { useSpeechesByPolitician } from "@/hooks/useDemocracy";
import { SpeechRow } from "@/features/speeches/SpeechRow";
```

- [ ] **Step 2: Extend the `activeTab` union**

Replace:

```tsx
const [activeTab, setActiveTab] = useState<"votes" | "promises">("votes");
```

with:

```tsx
const [activeTab, setActiveTab] = useState<"votes" | "promises" | "anforanden">("votes");
```

- [ ] **Step 3: Add the speeches query**

Below the existing `loadingPromises` query:

```tsx
  const { data: speeches, isLoading: loadingSpeeches } = useSpeechesByPolitician(
    id,
    50,
  );
```

- [ ] **Step 4: Update the tab strip**

Find the existing tab loop. Replace:

```tsx
{(["votes", "promises"] as const).map((tab) => {
  const label = tab === "votes" ? "Röstningshistorik" : "Löften";
```

with:

```tsx
{(["votes", "promises", "anforanden"] as const).map((tab) => {
  const label =
    tab === "votes" ? "Röstningshistorik"
    : tab === "promises" ? "Löften"
    : "Anföranden";
```

- [ ] **Step 5: Add the anföranden tab body**

Find the existing tab-body section (after the tab strip closes). Append a third conditional block:

```tsx
      {activeTab === "anforanden" && (
        <div className="rounded-xl p-4" style={{ background: "var(--color-surface-lowest)" }}>
          <div className="text-[10px] font-mono uppercase tracking-widest text-on-surface-variant mb-3">
            Senaste anföranden i kammaren
          </div>
          {loadingSpeeches && (
            <div className="text-on-surface-variant text-sm py-6 text-center">Laddar...</div>
          )}
          {!loadingSpeeches && (!speeches || speeches.length === 0) && (
            <p className="text-sm italic text-on-surface-variant py-4">
              Inga registrerade anföranden ännu.
            </p>
          )}
          {speeches && speeches.length > 0 && (
            <div>
              {speeches.map((s) => (
                <SpeechRow key={s.id} speech={s} hidePolitician />
              ))}
            </div>
          )}
        </div>
      )}
```

- [ ] **Step 6: Typecheck + commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker/frontend
npx tsc --noEmit

cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add frontend/src/features/politicians/PoliticianPage.tsx
git commit -m "feat(politicians): add Anföranden tab to PoliticianPage"
git push origin main
```

---

## Task 11: Frontend — `PartyDetailPage` + route

**Files:**
- Create: `frontend/src/features/parties/PartyDetailPage.tsx`
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/features/parties/PartiesPage.tsx` (only the link target if it currently goes to `/parties/:party/goals`).

- [ ] **Step 1: Create `PartyDetailPage.tsx`**

```tsx
import { useQuery } from "@tanstack/react-query";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { partiesApi } from "./api";
import { politiciansApi } from "@/features/politicians/api";
import { PARTY_COLORS, partyShortToName } from "@/shared/design";
import { useSpeechesByParty } from "@/hooks/useDemocracy";
import { SpeechRow } from "@/features/speeches/SpeechRow";
import { useMediaQuery } from "@/hooks/useMediaQuery";

type Tab = "mal" | "anforanden" | "politiker";

const TABS: { key: Tab; label: string }[] = [
  { key: "mal", label: "Mål" },
  { key: "anforanden", label: "Anföranden" },
  { key: "politiker", label: "Politiker" },
];

export function PartyDetailPage() {
  const { party = "" } = useParams<{ party: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const isMobile = useMediaQuery("(max-width: 640px)");

  const tab: Tab =
    (searchParams.get("tab") as Tab | null) && TABS.some((t) => t.key === searchParams.get("tab"))
      ? (searchParams.get("tab") as Tab)
      : "mal";

  const setTab = (next: Tab) => {
    const params = new URLSearchParams(searchParams);
    if (next === "mal") params.delete("tab");
    else params.set("tab", next);
    setSearchParams(params, { replace: true });
  };

  const { data: parties } = useQuery({
    queryKey: ["parties"],
    queryFn: partiesApi.listParties,
    staleTime: 5 * 60 * 1000,
  });
  const summary = parties?.find((p) => p.party === party);

  const { data: goals, isLoading: loadingGoals } = useQuery({
    queryKey: ["party-goals", party],
    queryFn: () => partiesApi.listGoals(party),
    enabled: !!party && tab === "mal",
    staleTime: 60_000,
  });

  const { data: speeches, isLoading: loadingSpeeches } = useSpeechesByParty(
    tab === "anforanden" ? party : undefined,
    50,
  );

  const { data: politiciansResp, isLoading: loadingPoliticians } = useQuery({
    queryKey: ["politicians", { party, pageSize: 200 }],
    queryFn: () => politiciansApi.list({ party, pageSize: 200 }),
    enabled: !!party && tab === "politiker",
    staleTime: 60_000,
  });

  const pc = PARTY_COLORS[party];

  return (
    <div className="sdt-page" style={{ paddingBottom: 64 }}>
      {/* Back link */}
      <div style={{ padding: isMobile ? "16px 14px 0" : "32px 32px 0" }}>
        <Link
          to="/parties"
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-accent)",
            textDecoration: "none",
            letterSpacing: "0.1em",
          }}
        >
          ← Alla partier
        </Link>
      </div>

      {/* Hero */}
      <header
        style={{
          padding: isMobile ? "12px 14px 16px" : "16px 32px 24px",
          display: "flex",
          alignItems: "center",
          gap: 16,
        }}
      >
        <span
          style={{
            display: "inline-block",
            width: 6,
            height: 56,
            background: pc?.bg ?? "#888",
            flexShrink: 0,
          }}
        />
        <div style={{ minWidth: 0 }}>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.1em",
            }}
          >
            {party}
          </div>
          <h1
            style={{
              fontFamily: "var(--font-serif)",
              fontSize: isMobile ? 28 : 40,
              fontWeight: 400,
              margin: 0,
              lineHeight: 1.05,
              color: "var(--color-fg)",
            }}
          >
            {partyShortToName(party)}
          </h1>
          {summary && (
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                color: "var(--color-fg-muted)",
                marginTop: 6,
              }}
            >
              {summary.totalGoals ?? 0} mål · {summary.alignedPct ?? 0}% i linje
            </div>
          )}
        </div>
      </header>

      {/* Tab strip */}
      <div
        style={{
          display: "flex",
          gap: 0,
          padding: isMobile ? "0 14px" : "0 32px",
          borderBottom: "1px solid var(--color-border)",
          overflowX: "auto",
        }}
      >
        {TABS.map((t) => {
          const active = tab === t.key;
          return (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              style={{
                background: "transparent",
                border: "none",
                padding: "10px 16px",
                fontFamily: "var(--font-mono)",
                fontSize: 12,
                fontWeight: active ? 700 : 500,
                color: active ? "var(--color-fg)" : "var(--color-fg-muted)",
                borderBottom: active ? "2px solid var(--color-fg)" : "2px solid transparent",
                cursor: "pointer",
                whiteSpace: "nowrap",
                marginBottom: -1,
                letterSpacing: "0.05em",
              }}
            >
              {t.label}
            </button>
          );
        })}
      </div>

      {/* Tab body */}
      <main
        style={{
          padding: isMobile ? "16px 14px 0" : "24px 32px 0",
          maxWidth: 920,
        }}
      >
        {tab === "mal" && (
          <section>
            {loadingGoals && (
              <div style={{ fontSize: 13, color: "var(--color-fg-muted)" }}>Laddar mål...</div>
            )}
            {!loadingGoals && (!goals || goals.length === 0) && (
              <p style={{ fontStyle: "italic", color: "var(--color-fg-muted)", fontSize: 13 }}>
                Inga registrerade mål för {partyShortToName(party)}.
              </p>
            )}
            {goals && goals.length > 0 && (
              <ul style={{ listStyle: "none", margin: 0, padding: 0 }}>
                {goals.map((g) => (
                  <li
                    key={g.id}
                    style={{
                      borderBottom: "1px solid var(--color-border)",
                      padding: "12px 0",
                    }}
                  >
                    <div style={{ fontFamily: "var(--font-serif)", fontSize: 15, color: "var(--color-fg)" }}>
                      {g.goalText}
                    </div>
                    <div
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 11,
                        color: "var(--color-fg-muted)",
                        marginTop: 4,
                      }}
                    >
                      {g.topic} · {g.alignmentPct ?? 0}% i linje · {g.relevantVotes ?? 0} röster
                    </div>
                  </li>
                ))}
              </ul>
            )}
            <Link
              to={`/parties/${party}/goals`}
              style={{
                display: "inline-block",
                marginTop: 16,
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                color: "var(--color-accent)",
                textDecoration: "none",
              }}
            >
              Visa fullständig målvy →
            </Link>
          </section>
        )}

        {tab === "anforanden" && (
          <section>
            {loadingSpeeches && (
              <div style={{ fontSize: 13, color: "var(--color-fg-muted)" }}>Laddar anföranden...</div>
            )}
            {!loadingSpeeches && (!speeches || speeches.length === 0) && (
              <p style={{ fontStyle: "italic", color: "var(--color-fg-muted)", fontSize: 13 }}>
                Inga registrerade anföranden från {partyShortToName(party)} ännu.
              </p>
            )}
            {speeches && speeches.length > 0 && (
              <div>
                {speeches.map((s) => (
                  <SpeechRow key={s.id} speech={s} />
                ))}
              </div>
            )}
          </section>
        )}

        {tab === "politiker" && (
          <section>
            {loadingPoliticians && (
              <div style={{ fontSize: 13, color: "var(--color-fg-muted)" }}>Laddar politiker...</div>
            )}
            {!loadingPoliticians &&
              (!politiciansResp || politiciansResp.data.length === 0) && (
                <p style={{ fontStyle: "italic", color: "var(--color-fg-muted)", fontSize: 13 }}>
                  Inga politiker hittades för {partyShortToName(party)}.
                </p>
              )}
            {politiciansResp && politiciansResp.data.length > 0 && (
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: isMobile
                    ? "minmax(0, 1fr)"
                    : "repeat(auto-fill, minmax(220px, 1fr))",
                  gap: 8,
                }}
              >
                {politiciansResp.data.map((p) => (
                  <Link
                    key={p.intressentId}
                    to={`/politicians/${p.intressentId}`}
                    style={{
                      background: "var(--color-sdt-surface)",
                      border: "1px solid var(--color-border)",
                      padding: "10px 12px",
                      borderRadius: 4,
                      textDecoration: "none",
                      color: "inherit",
                      display: "flex",
                      alignItems: "center",
                      gap: 10,
                      minWidth: 0,
                    }}
                  >
                    {p.imageUrl ? (
                      <img
                        src={p.imageUrl}
                        alt=""
                        style={{
                          width: 32,
                          height: 32,
                          borderRadius: "50%",
                          objectFit: "cover",
                          flexShrink: 0,
                        }}
                      />
                    ) : (
                      <span
                        style={{
                          width: 32,
                          height: 32,
                          borderRadius: "50%",
                          background: pc?.bg ?? "#888",
                          color: pc?.text ?? "#fff",
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "center",
                          fontSize: 11,
                          fontFamily: "var(--font-mono)",
                          fontWeight: 700,
                          flexShrink: 0,
                        }}
                      >
                        {p.firstName[0]}{p.lastName[0]}
                      </span>
                    )}
                    <div style={{ minWidth: 0 }}>
                      <div
                        style={{
                          fontFamily: "var(--font-serif)",
                          fontSize: 14,
                          color: "var(--color-fg)",
                          overflow: "hidden",
                          textOverflow: "ellipsis",
                          whiteSpace: "nowrap",
                        }}
                      >
                        {p.firstName} {p.lastName}
                      </div>
                      {p.constituency && (
                        <div
                          style={{
                            fontFamily: "var(--font-mono)",
                            fontSize: 10,
                            color: "var(--color-fg-muted)",
                            overflow: "hidden",
                            textOverflow: "ellipsis",
                            whiteSpace: "nowrap",
                          }}
                        >
                          {p.constituency}
                        </div>
                      )}
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </section>
        )}
      </main>
    </div>
  );
}
```

- [ ] **Step 2: Add the route in `App.tsx`**

Add an import alongside the other feature pages:

```tsx
import { PartyDetailPage } from "./features/parties/PartyDetailPage";
```

In the `<Routes>` block, add the route **before** `<Route path="/parties/:party/goals" ...>` so the more-specific path doesn't get shadowed (React Router v7 picks specific over generic, but adding it first keeps the file readable):

```tsx
<Route path="/parties/:party"                          element={<PartyDetailPage />} />
```

- [ ] **Step 3: Verify `PartiesPage` cards link to `/parties/:party`**

Open `frontend/src/features/parties/PartiesPage.tsx`. Find the card link target — if it currently uses `/parties/:party/goals`, change it to `/parties/:party`. Show this exact diff snippet:

```tsx
// before
<Link to={`/parties/${p.party}/goals`}>...</Link>
// after
<Link to={`/parties/${p.party}`}>...</Link>
```

Only change the link target. Do not restructure the card.

If the existing target is already `/parties/:party` or some other path, leave it.

- [ ] **Step 4: Verify `partiesApi.listGoals` exists with that exact name**

```bash
grep -n "listGoals" /home/rune/Documents/swedish-democratic-tracker/frontend/src/features/parties/api.ts
```

If the method is named differently (e.g. `getGoals`), use the actual name in `PartyDetailPage.tsx` Step 1 above. Type-check will catch a wrong reference.

- [ ] **Step 5: Verify `partyShortToName` is exported from `@/shared/design`**

```bash
grep -n "partyShortToName" /home/rune/Documents/swedish-democratic-tracker/frontend/src/shared/design.ts
```

Expected: existing export from earlier work. If missing, add (this should not happen but check):

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

- [ ] **Step 6: Typecheck**

```bash
cd /home/rune/Documents/swedish-democratic-tracker/frontend
npx tsc --noEmit
```

Expected: empty output. If errors mention missing `politiciansResp.data` shape, inspect the actual `politiciansApi.list` return type — it returns `{ data, total, pageSize, page }` per existing code; the Step 1 usage is correct, but adjust if the type ever shifts.

- [ ] **Step 7: Commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add frontend/src/features/parties/PartyDetailPage.tsx frontend/src/App.tsx frontend/src/features/parties/PartiesPage.tsx
git commit -m "feat(parties): /parties/:party route — hero + Mål · Anföranden · Politiker tabs"
git push origin main
```

---

## Task 12: Verification

- [ ] **Step 1: Local build + typecheck**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
go build -C backend ./...
cd frontend && npx tsc --noEmit && npm run build
```

Expected: backend silent (exit 0), tsc silent, vite build prints `✓ built in ...`.

- [ ] **Step 2: Local smoke**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
make run
```

Open in a browser:
- `http://localhost:5173/beslut/AU9?title=...` — confirm SAMMANFATTNING block (or muted empty line) and DEBATTEN section (or "Inga registrerade anföranden..." line) appear above HELA BETÄNKANDET. The party-vote breakdown still renders.
- `http://localhost:5173/politicians/{some-intressent-id}` — confirm a third "Anföranden" tab is selectable; switching to it shows up to 50 SpeechRow entries (or empty-state line).
- `http://localhost:5173/parties/S` — confirm hero, three tabs, and tab body switch via URL query (e.g. `?tab=anforanden`).
- `http://localhost:5173/parties/S?tab=politiker` — politicians grid renders.

- [ ] **Step 3: Mobile prod check (after CI deploy)**

```bash
git push origin main  # already done by previous tasks; ensure latest is pushed
```

Wait for CI green + ArgoCD rollout (~5-10 min). Use Playwright at 375 × 812 against:
- `https://sdt.runevibe.se/parties/S` — `overflowPx === 0`, three tabs visible, tap-friendly.
- `https://sdt.runevibe.se/beslut/AU9?title=...` — SpeechRow list renders, no horizontal scroll.

- [ ] **Step 4: Speakers data sanity**

For an AU-class beslut with a known debate (e.g. AU9 from earlier audits), the speakers list should match the names visible on `riksdagen.se`'s debate page. If the list is dominated by interpellation speeches (off-topic), file a follow-up to filter on `kammaraktivitet` per the spec's open question.

- [ ] **Step 5: Tear down dev stack**

```bash
docker compose -f docker-compose.dev.yml down
```

---

## Self-review notes

- **Spec coverage:**
  - Section 1 (Backend endpoints) → Tasks 1, 2, 4.
  - Section 1 (politicians.Service.ImageURLByID + PoliticianLookup extension) → Tasks 3, 4.
  - SpeechRow shared component (Section "Components") → Task 8.
  - BeslutDetailPage SAMMANFATTNING + DEBATTEN → Task 9.
  - PoliticianPage anföranden card → Task 10 (implemented as a third tab to match the page's existing tab pattern; spec said "card" but the page is tab-driven, so the tab approach is consistent).
  - PartyDetailPage with Mål · Anföranden · Politiker tabs → Task 11.
  - URL query for tab state (`?tab=...`) → Task 11.
  - Empty-state copy lines → Tasks 9, 10, 11.
  - OpenAPI + regen → Task 5.
  - Verification → Task 12.
- **No new tests:** project has no unit-test framework. Verification leans on `go build`, `tsc --noEmit`, `npm run build`, and Playwright at 375 × 812.
- **Type consistency:** `Speech` shape (with new `politicianImageUrl`) defined in Task 6, used by Tasks 8, 9, 10, 11. `PoliticianLookup` extension in Task 4 implemented by `politicians.Service` in Task 3 — both declared in the same plan run.
- **Spec ambiguity addressed:** the spec said "PoliticianPage card", this plan implements it as a tab to match the page's existing tab UX. If the user prefers a card, swap the tab body with an inline section above the existing tabs — single-file edit.
- **`PartiesPage` link target:** Task 11 Step 3 contains a conditional verification rather than a fixed change, since the current target may already be correct. The verification step makes the necessary edit explicit.
