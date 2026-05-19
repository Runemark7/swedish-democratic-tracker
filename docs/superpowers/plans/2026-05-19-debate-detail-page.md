# Debate Detail Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `/debatt/:dokId` page showing the full interpellation text, debate participants, and all speeches — linked from the "Om debatten" block on `SpeechDetailPage`.

**Architecture:** Backend adds `GET /documents/:dokId/full` endpoint returning `bodyHtml` + full `intressenter` list alongside a new `Intressent` domain type. Frontend adds `DebateDetailPage`, a `useDocumentFull` hook, and a "Se hela debatten" link in `SpeechDetailPage`.

**Tech Stack:** Go 1.26 (chi, pgx), React 19, TypeScript 5, TanStack Query v5, Tailwind CSS v4

---

## File Map

| File | Action |
|------|--------|
| `backend/internal/votes/domain/vote.go` | Add `Intressent` struct; add `Intressenter []Intressent` to `DocumentStatus` |
| `backend/internal/votes/adapters/riksdagen/client.go` | Parse full intressenter list in `FetchDocumentStatus` |
| `backend/internal/votes/adapters/http/handler.go` | Add `getDocumentFull` handler; register `GET /documents/{dokId}/full` before `GET /documents/{dokId}` |
| `api/openapi.yaml` | Add `Intressent` + `RiksdagDocumentFull` schemas; add `/documents/{dokId}/full` path |
| `frontend/src/shared/types.ts` | Add `Intressent` interface + `RiksdagDocumentFull` interface |
| `frontend/src/features/votes/api.ts` | Add `getDocumentFull` method |
| `frontend/src/hooks/useDemocracy.ts` | Add `useDocumentFull` hook |
| `frontend/src/features/speeches/DebateDetailPage.tsx` | Create new page |
| `frontend/src/features/speeches/SpeechDetailPage.tsx` | Add "→ Se hela debatten" link in "Om debatten" block |
| `frontend/src/App.tsx` | Import `DebateDetailPage`; add `/debatt/:dokId` route |

---

### Task 1: Add `Intressent` domain type + update Riksdagen client

**Files:**
- Modify: `backend/internal/votes/domain/vote.go`
- Modify: `backend/internal/votes/adapters/riksdagen/client.go`

- [ ] **Step 1: Add `Intressent` struct and field to `DocumentStatus`**

In `backend/internal/votes/domain/vote.go`, add the `Intressent` struct and the `Intressenter` field to `DocumentStatus`. The final `DocumentStatus` struct (starting at line 46) becomes:

```go
// Intressent is a politician formally attached to a Riksdagen document.
type Intressent struct {
	IntressentID string `json:"intressentId"`
	Name         string `json:"name"`
	Party        string `json:"party"`
	Role         string `json:"role"`
}

// DocumentStatus represents the parsed response from /dokumentstatus/{dok_id}.json.
type DocumentStatus struct {
	DokID        string              `json:"dokId"`
	Title        string              `json:"title"`
	Type         string              `json:"type"`
	References   []DocumentReference `json:"references"`
	Intressenter []Intressent        `json:"intressenter,omitempty"`
	Date         string
	Subtitle     string
	Summary      string
	Beteckning   string
	// BodyHTML is the full document content from /dokumentstatus.dokument.html.
	// Inline-rendered on the BeslutDetailPage so users see the proposal,
	// motivation and debate transcript without leaving the site.
	BodyHTML string `json:"bodyHtml,omitempty"`
}
```

- [ ] **Step 2: Update `FetchDocumentStatus` to parse the full intressent list**

In `backend/internal/votes/adapters/riksdagen/client.go`, update the anonymous struct inside `FetchDocumentStatus` for `Dokintressent`. Replace the current struct (which only has `Partibet`) with one that captures all fields:

```go
Dokintressent struct {
    Intressent []struct {
        IntressentID string `json:"intressent_id"`
        Namn         string `json:"namn"`
        Partibet     string `json:"partibet"`
        Roll         string `json:"roll"`
    } `json:"intressent"`
} `json:"dokintressent"`
```

Then replace the population block at the bottom of the function (currently lines ~193–210) with:

```go
for _, i := range ds.Dokintressent.Intressent {
    status.Intressenter = append(status.Intressenter, domain.Intressent{
        IntressentID: i.IntressentID,
        Name:         i.Namn,
        Party:        i.Partibet,
        Role:         i.Roll,
    })
}
// Preserve existing behaviour: first intressent's party feeds proposal-origin tracing via References.
if len(ds.Dokintressent.Intressent) > 0 {
    status.References = append(status.References, domain.DocumentReference{
        RefDokTyp: ds.Dokument.Typ,
        RefDokID:  ds.Dokument.DokID,
        PartyBet:  ds.Dokintressent.Intressent[0].Partibet,
    })
}
return status, nil
```

- [ ] **Step 3: Verify backend compiles**

```bash
go build -C backend ./...
```

Expected: no output, exit 0.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/votes/domain/vote.go \
        backend/internal/votes/adapters/riksdagen/client.go
git commit -m "feat(votes): Intressent domain type + parse full intressenter in FetchDocumentStatus"
```

---

### Task 2: Add `/documents/:dokId/full` HTTP endpoint

**Files:**
- Modify: `backend/internal/votes/adapters/http/handler.go`

- [ ] **Step 1: Register the new route — before the existing `/documents/{dokId}` route**

In `handler.go` `Routes` method, the current lines are:

```go
r.Get("/politicians/{id}/votes", h.listByPolitician)
r.Get("/votes", h.listAll)
r.Get("/votes/riksdag-feed", h.riksdagFeed)
r.Get("/votes/{beteckning}/{punkt}", h.getDetail)
r.Get("/documents/{dokId}", h.getDocument)
```

Change to:

```go
r.Get("/politicians/{id}/votes", h.listByPolitician)
r.Get("/votes", h.listAll)
r.Get("/votes/riksdag-feed", h.riksdagFeed)
r.Get("/votes/{beteckning}/{punkt}", h.getDetail)
r.Get("/documents/{dokId}/full", h.getDocumentFull)
r.Get("/documents/{dokId}", h.getDocument)
```

- [ ] **Step 2: Add the `getDocumentFull` handler**

Add this method to `handler.go`, after the existing `getDocument` method (after line 284):

```go
func (h *Handler) getDocumentFull(w http.ResponseWriter, r *http.Request) {
	dokID := chi.URLParam(r, "dokId")
	if dokID == "" {
		jsonError(w, "dokId required", http.StatusBadRequest)
		return
	}
	ds, err := h.svc.GetDocumentStatus(r.Context(), dokID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ds == nil {
		jsonError(w, "document not found", http.StatusNotFound)
		return
	}
	intressenter := ds.Intressenter
	if intressenter == nil {
		intressenter = []domain.Intressent{}
	}
	resp := map[string]any{
		"dokId":        ds.DokID,
		"type":         ds.Type,
		"title":        ds.Title,
		"subtitle":     ds.Subtitle,
		"summary":      ds.Summary,
		"date":         ds.Date,
		"beteckning":   ds.Beteckning,
		"bodyHtml":     ds.BodyHTML,
		"intressenter": intressenter,
	}
	jsonOK(w, resp)
}
```

- [ ] **Step 3: Verify backend compiles**

```bash
go build -C backend ./...
```

Expected: no output, exit 0.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/votes/adapters/http/handler.go
git commit -m "feat(votes): GET /documents/:dokId/full endpoint — bodyHtml + intressenter"
```

---

### Task 3: Update API spec + frontend types

**Files:**
- Modify: `api/openapi.yaml`
- Modify: `frontend/src/shared/types.ts`

- [ ] **Step 1: Add `Intressent` and `RiksdagDocumentFull` to openapi.yaml**

In `api/openapi.yaml`, after the `RiksdagDocument` schema block (after line 1542), insert:

```yaml
    Intressent:
      type: object
      required: [intressentId, name, party, role]
      properties:
        intressentId:
          type: string
        name:
          type: string
        party:
          type: string
          description: Party abbreviation (e.g. "S", "M")
        role:
          type: string
          description: Role in this document (e.g. "Upphovsman", "Svar")

    RiksdagDocumentFull:
      type: object
      required: [dokId, type, title, intressenter]
      properties:
        dokId:
          type: string
        type:
          type: string
          description: Riksdagen document type (bet, ip, mot, prop, etc.)
        title:
          type: string
        subtitle:
          type: string
        summary:
          type: string
        date:
          type: string
        beteckning:
          type: string
        bodyHtml:
          type: string
          description: Full document body from riksdagen dokumentstatus HTML field
        intressenter:
          type: array
          items:
            $ref: "#/components/schemas/Intressent"
```

- [ ] **Step 2: Add `/documents/{dokId}/full` path to openapi.yaml**

In `api/openapi.yaml`, before the existing `/documents/{dokId}:` entry (before line 313), insert:

```yaml
  /documents/{dokId}/full:
    get:
      operationId: getDocumentFull
      tags: [votes]
      summary: Full Riksdagen document — body HTML + intressenter list
      parameters:
        - name: dokId
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Full document with body and participants
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/RiksdagDocumentFull"
        "404":
          $ref: "#/components/responses/NotFound"

```

- [ ] **Step 3: Add `Intressent` and `RiksdagDocumentFull` to `frontend/src/shared/types.ts`**

After the `RiksdagDocument` interface (after line 367), add:

```ts
export interface Intressent {
  intressentId: string;
  name: string;
  party: string;
  role: string;
}

export interface RiksdagDocumentFull extends RiksdagDocument {
  bodyHtml?: string;
  intressenter: Intressent[];
}
```

- [ ] **Step 4: Regenerate frontend API contract**

```bash
cd frontend && npm run generate:api
```

Expected: `frontend/src/shared/api-contract.ts` updated with no errors.

- [ ] **Step 5: Commit**

```bash
git add api/openapi.yaml frontend/src/shared/types.ts frontend/src/shared/api-contract.ts
git commit -m "feat(api): RiksdagDocumentFull schema + /documents/:dokId/full path"
```

---

### Task 4: Add `votesApi.getDocumentFull` + `useDocumentFull` hook

**Files:**
- Modify: `frontend/src/features/votes/api.ts`
- Modify: `frontend/src/hooks/useDemocracy.ts`

- [ ] **Step 1: Add `getDocumentFull` to `votesApi`**

In `frontend/src/features/votes/api.ts`, update imports and add the new method. The full file becomes:

```ts
import { api } from "@/shared/api-client";
import type { VoteDetail, VoteSummaryListResponse, RiksdagDocument, RiksdagDocumentFull } from "@/shared/types";

export const votesApi = {
  list: (params: { page?: number; pageSize?: number } = {}) =>
    api.get<VoteSummaryListResponse>(
      `/votes?page=${params.page ?? 1}&pageSize=${params.pageSize ?? 50}`,
    ),
  getDetail: (beteckning: string, punkt: string) =>
    api.get<VoteDetail>(`/votes/${beteckning}/${punkt}`),
  getDocument: (dokId: string) =>
    api.get<RiksdagDocument>(`/documents/${encodeURIComponent(dokId)}`),
  getDocumentFull: (dokId: string) =>
    api.get<RiksdagDocumentFull>(`/documents/${encodeURIComponent(dokId)}/full`),
};
```

- [ ] **Step 2: Add `useDocumentFull` to `useDemocracy.ts`**

At the end of `frontend/src/hooks/useDemocracy.ts` (after the `useDocument` function, after line 610), add:

```ts
export function useDocumentFull(dokId: string | undefined) {
  return useQuery<RiksdagDocumentFull>({
    queryKey: ["document-full", dokId],
    queryFn: () => votesApi.getDocumentFull(dokId ?? ""),
    enabled: !!dokId,
    staleTime: 5 * 60 * 1000,
  });
}
```

Also ensure `RiksdagDocumentFull` is imported at the top of `useDemocracy.ts`. The existing import on line 8 already imports from `@/shared/types`; extend it:

```ts
import type { ElectionResult, RegionSummary, MunicipalitySummary, MunicipalityKPIItem, RiksdagDocument, RiksdagDocumentFull } from "@/shared/types";
```

- [ ] **Step 3: Type-check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/votes/api.ts frontend/src/hooks/useDemocracy.ts
git commit -m "feat(frontend): votesApi.getDocumentFull + useDocumentFull hook"
```

---

### Task 5: Create `DebateDetailPage`

**Files:**
- Create: `frontend/src/features/speeches/DebateDetailPage.tsx`

- [ ] **Step 1: Create the page**

Create `frontend/src/features/speeches/DebateDetailPage.tsx` with the following content:

```tsx
import { useNavigate, useParams } from "react-router-dom";
import { Link } from "react-router-dom";
import { useDocumentFull, useSpeechesByDocument } from "@/hooks/useDemocracy";
import { SpeechRow } from "@/features/speeches/SpeechRow";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { PARTY_COLORS } from "@/shared/design";

const TYPE_LABEL: Record<string, string> = {
  bet: "Betänkande",
  ip: "Interpellation",
  mot: "Motion",
  prop: "Proposition",
  prot: "Protokoll",
  fr: "Skriftlig fråga",
};

function formatDate(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  const dd = String(d.getDate()).padStart(2, "0");
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const yyyy = d.getFullYear();
  return `${dd}/${mm}/${yyyy}`;
}

const backBtnStyle: React.CSSProperties = {
  background: "none",
  border: "none",
  cursor: "pointer",
  padding: 0,
  fontFamily: "var(--font-mono)",
  fontSize: 11,
  color: "var(--color-accent)",
  letterSpacing: "0.1em",
};

export function DebateDetailPage() {
  const { dokId } = useParams<{ dokId: string }>();
  const navigate = useNavigate();
  const isMobile = useMediaQuery("(max-width: 640px)");

  const { data: doc, isLoading, error } = useDocumentFull(dokId);
  const { data: speeches } = useSpeechesByDocument(dokId);

  if (isLoading) {
    return (
      <div className="sdt-page" style={{ padding: isMobile ? "20px 14px" : "32px" }}>
        <div style={{ color: "var(--color-fg-muted)", fontSize: 13 }}>Laddar debatt…</div>
      </div>
    );
  }

  if (error || !doc) {
    return (
      <div className="sdt-page" style={{ padding: isMobile ? "20px 14px" : "32px" }}>
        <button onClick={() => navigate(-1)} style={backBtnStyle}>← Tillbaka</button>
        <div style={{ marginTop: 24, color: "var(--color-pulse)" }}>Kunde inte hämta debatten.</div>
      </div>
    );
  }

  const typeLabel = TYPE_LABEL[doc.type] ?? doc.type.toUpperCase();
  const intressenter = doc.intressenter ?? [];

  return (
    <div className="sdt-page" style={{ paddingBottom: 64 }}>
      <div style={{ padding: isMobile ? "16px 14px 0" : "32px 32px 0", maxWidth: 880 }}>
        {/* Back */}
        <button onClick={() => navigate(-1)} style={backBtnStyle}>← Tillbaka</button>

        {/* Header */}
        <div style={{ marginTop: 16, marginBottom: 24 }}>
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
            {typeLabel}{doc.date ? ` · ${formatDate(doc.date)}` : ""}
          </div>
          <h1
            style={{
              fontFamily: "var(--font-serif)",
              fontStyle: "italic",
              fontSize: isMobile ? 20 : 26,
              fontWeight: 400,
              margin: "0 0 8px",
              lineHeight: 1.25,
              color: "var(--color-fg)",
            }}
          >
            {doc.title}
          </h1>
          {doc.subtitle && (
            <div
              style={{
                fontFamily: "var(--font-body)",
                fontSize: 13,
                color: "var(--color-fg-muted)",
                marginBottom: 8,
              }}
            >
              {doc.subtitle}
            </div>
          )}
          {doc.summary && (
            <div
              style={{
                fontFamily: "var(--font-body)",
                fontSize: 14,
                lineHeight: 1.65,
                color: "var(--color-fg)",
                background: "var(--color-sdt-surface)",
                border: "1px solid var(--color-border)",
                padding: isMobile ? "12px 14px" : "14px 18px",
              }}
            >
              {doc.summary}
            </div>
          )}
        </div>

        {/* Debattdeltagare */}
        {intressenter.length > 0 && (
          <section style={{ marginBottom: 32 }}>
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.15em",
                color: "var(--color-fg-muted)",
                textTransform: "uppercase",
                marginBottom: 12,
              }}
            >
              Debattdeltagare
            </div>
            <div style={{ display: "flex", flexWrap: "wrap", gap: 10 }}>
              {intressenter.map((p) => {
                const pc = PARTY_COLORS[p.party];
                const initials =
                  p.name
                    .split(" ")
                    .filter(Boolean)
                    .slice(0, 2)
                    .map((n) => n[0])
                    .join("") || "?";
                return (
                  <Link
                    key={p.intressentId}
                    to={`/politicians/${encodeURIComponent(p.intressentId)}`}
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 8,
                      padding: "6px 10px",
                      background: "var(--color-sdt-surface)",
                      border: "1px solid var(--color-border)",
                      textDecoration: "none",
                    }}
                  >
                    <div
                      style={{
                        width: 28,
                        height: 28,
                        borderRadius: "50%",
                        background: pc?.bg ?? "#888",
                        color: pc?.text ?? "#fff",
                        fontFamily: "var(--font-mono)",
                        fontSize: 11,
                        fontWeight: 700,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        flexShrink: 0,
                      }}
                    >
                      {initials}
                    </div>
                    <div style={{ minWidth: 0 }}>
                      <div
                        style={{
                          fontFamily: "var(--font-serif)",
                          fontSize: 13,
                          color: "var(--color-fg)",
                          whiteSpace: "nowrap",
                          overflow: "hidden",
                          textOverflow: "ellipsis",
                        }}
                      >
                        {p.name}
                      </div>
                      <div style={{ display: "flex", gap: 4, alignItems: "center", marginTop: 2 }}>
                        <span
                          style={{
                            fontFamily: "var(--font-mono)",
                            fontSize: 9,
                            fontWeight: 700,
                            padding: "1px 5px",
                            background: pc?.bg ?? "#888",
                            color: pc?.text ?? "#fff",
                            letterSpacing: "0.05em",
                          }}
                        >
                          {p.party}
                        </span>
                        {p.role && (
                          <span
                            style={{
                              fontFamily: "var(--font-mono)",
                              fontSize: 9,
                              color: "var(--color-fg-muted)",
                              letterSpacing: "0.04em",
                            }}
                          >
                            {p.role}
                          </span>
                        )}
                      </div>
                    </div>
                  </Link>
                );
              })}
            </div>
          </section>
        )}

        {/* Document body */}
        {doc.bodyHtml && (
          <section style={{ marginBottom: 40 }}>
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.15em",
                color: "var(--color-fg-muted)",
                textTransform: "uppercase",
                marginBottom: 12,
              }}
            >
              Innehåll
            </div>
            <div
              className="riksdagen-doc"
              style={{ fontSize: 15, lineHeight: 1.7, color: "var(--color-fg)" }}
              dangerouslySetInnerHTML={{ __html: doc.bodyHtml }}
            />
          </section>
        )}

        {/* Alla anföranden */}
        {speeches && speeches.length > 0 && (
          <section
            style={{
              borderTop: "1px solid var(--color-border)",
              paddingTop: 24,
              marginBottom: 32,
            }}
          >
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.15em",
                color: "var(--color-fg-muted)",
                textTransform: "uppercase",
                marginBottom: 12,
              }}
            >
              Anföranden i debatten · {speeches.length}
            </div>
            <div>
              {speeches.map((s) => (
                <SpeechRow key={s.id} speech={s} />
              ))}
            </div>
          </section>
        )}

        {/* Source */}
        <div style={{ borderTop: "1px solid var(--color-border)", paddingTop: 16 }}>
          <a
            href={`https://data.riksdagen.se/dokument/${doc.dokId}`}
            target="_blank"
            rel="noopener noreferrer"
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--color-fg-muted)",
              textDecoration: "none",
              letterSpacing: "0.05em",
            }}
          >
            Källa: riksdagen.se →
          </a>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Type-check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/features/speeches/DebateDetailPage.tsx
git commit -m "feat(speeches): DebateDetailPage at /debatt/:dokId"
```

---

### Task 6: Wire up link in `SpeechDetailPage` + register route in `App.tsx`

**Files:**
- Modify: `frontend/src/features/speeches/SpeechDetailPage.tsx`
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: Add "→ Se hela debatten" link in `SpeechDetailPage`**

In `frontend/src/features/speeches/SpeechDetailPage.tsx`, the "Om debatten" section currently ends with (lines ~293–308):

```tsx
            {docLink && (
              <Link
                to={docLink}
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                  color: "var(--color-accent)",
                  textDecoration: "none",
                  letterSpacing: "0.05em",
                }}
              >
                → Läs hela betänkandet
              </Link>
            )}
          </section>
        )}
```

Replace with:

```tsx
            {docLink && (
              <Link
                to={docLink}
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                  color: "var(--color-accent)",
                  textDecoration: "none",
                  letterSpacing: "0.05em",
                }}
              >
                → Läs hela betänkandet
              </Link>
            )}
            {relDokId && (
              <Link
                to={`/debatt/${encodeURIComponent(relDokId)}`}
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                  color: "var(--color-accent)",
                  textDecoration: "none",
                  letterSpacing: "0.05em",
                }}
              >
                → Se hela debatten
              </Link>
            )}
          </section>
        )}
```

- [ ] **Step 2: Register route in `App.tsx`**

In `frontend/src/App.tsx`, add the import near the other speech imports (after line 24):

```tsx
import { DebateDetailPage } from "./features/speeches/DebateDetailPage";
```

In the `<Routes>` block, add the route after the existing `/anforanden/:id` route (after line 345):

```tsx
          <Route path="/anforanden/:id"                        element={<SpeechDetailPage />} />
          <Route path="/debatt/:dokId"                         element={<DebateDetailPage />} />
```

- [ ] **Step 3: Final type-check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 4: Final backend build check**

```bash
go build -C backend ./...
```

Expected: no output, exit 0.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/features/speeches/SpeechDetailPage.tsx frontend/src/App.tsx
git commit -m "feat(speeches): link SpeechDetailPage → /debatt/:dokId + register route"
```

---

## Self-Review Checklist

- [x] Task 1 adds `Intressent` type and updates client to parse name/id/role — not just party
- [x] Task 2 new endpoint returns `intressenter` as `[]` (not null) via nil guard
- [x] Task 3 openapi path added before existing `/documents/{dokId}` path — no ambiguity
- [x] Task 4 imports `RiksdagDocumentFull` in both `api.ts` and `useDemocracy.ts`
- [x] Task 5 `DebateDetailPage` uses `doc.intressenter ?? []` — safe against missing field
- [x] Task 6 `relDokId` is already declared at line 50 of `SpeechDetailPage` — no new variable needed
- [x] `useSpeechesByDocument(dokId)` on `DebateDetailPage` fetches speeches by `rel_dok_id = dokId` — correct, this is the interpellation's dokId
- [x] External link uses `https://data.riksdagen.se/dokument/${doc.dokId}` — reliable, no beteckning dependency
- [x] No tests exist in codebase — `go build` + `tsc --noEmit` used as verification steps per CLAUDE.md
