# Speech Detail + Politician Profile Tracking Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire `/anforanden/:id` into a navigation hub (clickable politician/party + image + "Om debatten" + transcript), and rework `/politicians/:id` from 3 flat tabs to 4 themed tabs (Aktivitet · Debatter · Ämnen · Löften) so visitors can track patterns of activity.

**Architecture:** One new backend endpoint `GET /api/documents/:dokId` (12-line chi handler delegating to existing `votes.Service.GetDocumentStatus`). Everything else composes existing endpoints client-side: `useSpeechesByPolitician`, `useSpeechesByDocument`, `politiciansApi.listVotes`, `politiciansApi.getById`. Reuses the `<SpeechRow>` component already shipped.

**Tech Stack:** Go 1.26 (chi), React 19 + Vite + TypeScript, TanStack Query v5. No new dependencies.

**Spec:** `docs/superpowers/specs/2026-05-10-speech-detail-context-design.md`. Read it before Task 1.

**Branching:** main, push per task.

---

## File map

**New files**

| Path | Responsibility |
|---|---|
| (none) | All new logic lives in modifications |

**Modified files**

| Path | Change |
|---|---|
| `backend/internal/votes/adapters/http/handler.go` | Add `Routes` entry + `getDocument` handler. |
| `api/openapi.yaml` | Add `/documents/:dokId` path + `RiksdagDocument` schema. |
| `frontend/src/shared/api-contract.ts` | Regenerated from openapi. |
| `frontend/src/shared/types.ts` | Add `RiksdagDocument` interface. |
| `frontend/src/features/votes/api.ts` | Add `votesApi.getDocument`. |
| `frontend/src/hooks/useDemocracy.ts` | Add `useDocument` hook. |
| `frontend/src/features/speeches/SpeechDetailPage.tsx` | Full rework — politician card, Om debatten, body, Resten av debatten. |
| `frontend/src/features/politicians/PoliticianPage.tsx` | Replace 3-tab strip with 4-tab strip; new Aktivitet, Debatter, Ämnen tabs; preserve Löften body. |

---

## Task 1: Backend — `GET /api/documents/:dokId`

**Files:**
- Modify: `backend/internal/votes/adapters/http/handler.go`

- [ ] **Step 1: Add the route**

Find the existing `Routes` method:

```go
func (h *Handler) Routes(r chi.Router) {
	r.Get("/politicians/{id}/votes", h.listByPolitician)
	r.Get("/votes", h.listAll)
	r.Get("/votes/riksdag-feed", h.riksdagFeed)
	r.Get("/votes/{beteckning}/{punkt}", h.getDetail)
}
```

Add one line:

```go
	r.Get("/documents/{dokId}", h.getDocument)
```

So it becomes:

```go
func (h *Handler) Routes(r chi.Router) {
	r.Get("/politicians/{id}/votes", h.listByPolitician)
	r.Get("/votes", h.listAll)
	r.Get("/votes/riksdag-feed", h.riksdagFeed)
	r.Get("/votes/{beteckning}/{punkt}", h.getDetail)
	r.Get("/documents/{dokId}", h.getDocument)
}
```

- [ ] **Step 2: Add the handler**

Append a new method on `Handler` at the bottom of the file (before any private helpers):

```go
func (h *Handler) getDocument(w http.ResponseWriter, r *http.Request) {
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
	// Strip the heavy bodyHtml — only the beslut detail page needs it.
	resp := map[string]any{
		"dokId":      ds.DokID,
		"type":       ds.Type,
		"title":      ds.Title,
		"subtitle":   ds.Subtitle,
		"summary":    ds.Summary,
		"date":       ds.Date,
		"beteckning": ds.Beteckning,
	}
	jsonOK(w, resp)
}
```

- [ ] **Step 3: Verify `DocumentStatus` has a `Beteckning` field**

```bash
grep -n "Beteckning" /home/rune/Documents/swedish-democratic-tracker/backend/internal/votes/domain/vote.go
```

If absent, add it to the `DocumentStatus` struct in
`backend/internal/votes/domain/vote.go` with the matching JSON tag, and parse it in `backend/internal/votes/adapters/riksdagen/client.go` from `dokumentstatus.dokument.beteckning`. Most-likely path: the field is already present (used by `/votes/{beteckning}/{punkt}` handler).

If genuinely absent:

In `backend/internal/votes/domain/vote.go`, add to the struct:

```go
type DocumentStatus struct {
	DokID      string              `json:"dokId"`
	Title      string              `json:"title"`
	Type       string              `json:"type"`
	References []DocumentReference `json:"references"`
	Date       string
	Subtitle   string
	Summary    string
	Beteckning string  // NEW
	BodyHTML   string  `json:"bodyHtml,omitempty"`
}
```

In `backend/internal/votes/adapters/riksdagen/client.go`, find the
`dokumentstatus.dokument` struct definition (already has `Titel`, `Typ`, etc.) and add a `Beteckning` field:

```go
Beteckning string `json:"beteckning"`
```

After parsing, populate `status.Beteckning = ds.Dokument.Beteckning`.

- [ ] **Step 4: Build**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
go build -C backend ./...
```

Expected: empty output, exit 0.

- [ ] **Step 5: Smoke-test (optional, requires `make run`)**

```bash
curl -s 'http://localhost:8080/api/documents/hd01au9' | python3 -m json.tool
```

Expected: JSON with `dokId`, `type: "bet"`, `title`, `beteckning: "AU9"`, etc. If the response is empty, the dokId is wrong; try `hd01au3` or another known betänkande.

- [ ] **Step 6: Commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add backend/internal/votes/
git commit -m "feat(votes): /api/documents/:dokId endpoint for speech + politician tracking"
git push origin main
```

---

## Task 2: OpenAPI + regen

**Files:**
- Modify: `api/openapi.yaml`

- [ ] **Step 1: Add the path entry**

Open `api/openapi.yaml`. Find the existing `/votes/{beteckning}/{punkt}` path entry. Add a new sibling path:

```yaml
  /documents/{dokId}:
    get:
      operationId: getDocument
      tags: [votes]
      summary: Riksdagen document metadata (no body HTML)
      parameters:
        - name: dokId
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Parsed dokumentstatus metadata
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/RiksdagDocument"
        "404":
          $ref: "#/components/responses/NotFound"
```

- [ ] **Step 2: Add the `RiksdagDocument` schema**

Find `components: schemas:` block. Add a new entry (alphabetised):

```yaml
    RiksdagDocument:
      type: object
      required: [dokId, type, title]
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
          description: Present for betänkanden and motions
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
git commit -m "chore(api): /documents/:dokId path + RiksdagDocument schema"
git push origin main
```

---

## Task 3: Frontend — `RiksdagDocument` type + `votesApi.getDocument`

**Files:**
- Modify: `frontend/src/shared/types.ts`
- Modify: `frontend/src/features/votes/api.ts`

- [ ] **Step 1: Add type**

In `frontend/src/shared/types.ts`, append at the bottom:

```ts
export interface RiksdagDocument {
  dokId: string;
  type: string;
  title: string;
  subtitle?: string;
  summary?: string;
  date?: string;
  beteckning?: string;
}
```

- [ ] **Step 2: Add api method**

In `frontend/src/features/votes/api.ts`, replace the file with:

```ts
import { api } from "@/shared/api-client";
import type { VoteDetail, VoteSummaryListResponse, RiksdagDocument } from "@/shared/types";

export const votesApi = {
  list: (params: { page?: number; pageSize?: number } = {}) =>
    api.get<VoteSummaryListResponse>(
      `/votes?page=${params.page ?? 1}&pageSize=${params.pageSize ?? 50}`,
    ),
  getDetail: (beteckning: string, punkt: string) =>
    api.get<VoteDetail>(`/votes/${beteckning}/${punkt}`),
  getDocument: (dokId: string) =>
    api.get<RiksdagDocument>(`/documents/${encodeURIComponent(dokId)}`),
};
```

- [ ] **Step 3: Typecheck**

```bash
cd /home/rune/Documents/swedish-democratic-tracker/frontend
npx tsc --noEmit
```

Expected: empty output.

- [ ] **Step 4: Commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add frontend/src/shared/types.ts frontend/src/features/votes/api.ts
git commit -m "feat(votes): votesApi.getDocument + RiksdagDocument type"
git push origin main
```

---

## Task 4: Frontend — `useDocument` hook

**Files:**
- Modify: `frontend/src/hooks/useDemocracy.ts`

- [ ] **Step 1: Append the hook**

Find the existing `useSpeechesByParty` hook at the bottom of the file. After it, append:

```ts
import type { RiksdagDocument } from "@/shared/types";
import { votesApi } from "@/features/votes/api";

export function useDocument(dokId: string | undefined) {
  return useQuery<RiksdagDocument>({
    queryKey: ["document", dokId],
    queryFn: () => votesApi.getDocument(dokId ?? ""),
    enabled: !!dokId,
    staleTime: 5 * 60 * 1000,
  });
}
```

If `votesApi` and `RiksdagDocument` are already imported elsewhere in the file, do not duplicate the imports — just add the function. Run `grep -n "votesApi\|RiksdagDocument" frontend/src/hooks/useDemocracy.ts` to check.

- [ ] **Step 2: Typecheck**

```bash
cd /home/rune/Documents/swedish-democratic-tracker/frontend
npx tsc --noEmit
```

Empty output expected.

- [ ] **Step 3: Commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add frontend/src/hooks/useDemocracy.ts
git commit -m "feat(hooks): useDocument(dokId) wraps /api/documents/:dokId"
git push origin main
```

---

## Task 5: Speech detail page — full rework

**Files:**
- Modify: `frontend/src/features/speeches/SpeechDetailPage.tsx`

This task replaces the page wholesale. Read the existing file first so you know what existed (back link, basic header, body) — most of it is being restructured.

- [ ] **Step 1: Replace the file contents**

Write the entire file:

```tsx
import { Link, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { speechesApi } from "@/features/speeches/api";
import { politiciansApi } from "@/features/politicians/api";
import { useDocument } from "@/hooks/useDemocracy";
import { useSpeechesByDocument } from "@/hooks/useDemocracy";
import { SpeechRow } from "@/features/speeches/SpeechRow";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { PARTY_COLORS, partyShortToName } from "@/shared/design";

const WEEKDAY = ["sön", "mån", "tis", "ons", "tor", "fre", "lör"];

function formatLongStamp(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  const wd = WEEKDAY[d.getDay()];
  const dd = String(d.getDate()).padStart(2, "0");
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const yyyy = d.getFullYear();
  return `${wd} ${dd}/${mm}/${yyyy}`;
}

const TYPE_LABEL: Record<string, string> = {
  bet: "Betänkande",
  ip: "Interpellation",
  mot: "Motion",
  prop: "Proposition",
  prot: "Protokoll",
  fr: "Skriftlig fråga",
};

export function SpeechDetailPage() {
  const { id } = useParams<{ id: string }>();
  const isMobile = useMediaQuery("(max-width: 640px)");
  const speechId = Number(id ?? "0");

  const { data: speech, isLoading, error } = useQuery({
    queryKey: ["speech", speechId],
    queryFn: () => speechesApi.getById(speechId),
    enabled: speechId > 0,
  });

  const { data: politician } = useQuery({
    queryKey: ["politician", speech?.politicianId],
    queryFn: () => politiciansApi.getById(speech?.politicianId ?? ""),
    enabled: !!speech?.politicianId,
    staleTime: 5 * 60 * 1000,
  });

  const { data: document } = useDocument(speech?.dokId && undefined /* see hook arg */ ? speech.dokId : undefined);
  // Speech model uses `dokId` for the protokoll/parent doc and may not
  // expose rel_dok_id directly — use anforandeNummer + dokId pair when
  // anforandelista doesn't surface relatedDokId. Fall back gracefully.
  const relDokId = (speech as unknown as { relatedDokId?: string })?.relatedDokId;
  const { data: doc2 } = useDocument(relDokId);
  const docToShow = doc2 ?? document;

  const { data: siblingSpeeches } = useSpeechesByDocument(relDokId);

  if (isLoading) {
    return (
      <div className="sdt-page" style={{ padding: isMobile ? "20px 14px" : "32px" }}>
        <div style={{ color: "var(--color-fg-muted)", fontSize: 13 }}>Laddar anförande…</div>
      </div>
    );
  }

  if (error || !speech) {
    return (
      <div className="sdt-page" style={{ padding: isMobile ? "20px 14px" : "32px" }}>
        <Link to="/" style={{ color: "var(--color-accent)", fontFamily: "var(--font-mono)", fontSize: 11, textDecoration: "none" }}>
          ← Tillbaka
        </Link>
        <div style={{ marginTop: 24, color: "var(--color-pulse)" }}>Kunde inte hämta anförandet.</div>
      </div>
    );
  }

  const pc = PARTY_COLORS[speech.party];
  const partyName = partyShortToName(speech.party);
  const initials =
    (speech.politicianName || "?")
      .split(" ")
      .filter(Boolean)
      .slice(0, 2)
      .map((n) => n[0])
      .join("") || "?";

  const otherSpeeches = (siblingSpeeches ?? []).filter((s) => s.id !== speech.id);
  const docTypeLabel = docToShow ? TYPE_LABEL[docToShow.type] ?? docToShow.type.toUpperCase() : null;
  const docLink =
    docToShow?.type === "bet" && docToShow.beteckning
      ? `/beslut/${encodeURIComponent(docToShow.beteckning)}`
      : null;

  return (
    <div className="sdt-page" style={{ paddingBottom: 64 }}>
      <div style={{ padding: isMobile ? "16px 14px 0" : "32px 32px 0", maxWidth: 880 }}>
        {/* Back link */}
        <Link
          to="/"
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-accent)",
            textDecoration: "none",
            letterSpacing: "0.1em",
          }}
        >
          ← Tillbaka
        </Link>

        {/* Politician card */}
        <section
          style={{
            marginTop: 16,
            padding: isMobile ? 16 : 20,
            background: pc?.light ?? "var(--color-sdt-surface)",
            border: "1px solid var(--color-border)",
            display: "flex",
            gap: 16,
            alignItems: "center",
            minWidth: 0,
          }}
        >
          {speech.politicianImageUrl ? (
            <img
              src={speech.politicianImageUrl}
              alt=""
              loading="lazy"
              onError={(e) => {
                (e.currentTarget as HTMLImageElement).style.display = "none";
              }}
              style={{
                width: 80,
                height: 80,
                borderRadius: "50%",
                objectFit: "cover",
                flexShrink: 0,
              }}
            />
          ) : (
            <div
              style={{
                width: 80,
                height: 80,
                borderRadius: "50%",
                background: pc?.bg ?? "#888",
                color: pc?.text ?? "#fff",
                fontFamily: "var(--font-mono)",
                fontSize: 22,
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
          <div style={{ minWidth: 0 }}>
            <Link
              to={`/politicians/${encodeURIComponent(speech.politicianId)}`}
              style={{
                fontFamily: "var(--font-serif)",
                fontSize: isMobile ? 22 : 28,
                color: "var(--color-fg)",
                textDecoration: "none",
                lineHeight: 1.1,
                display: "block",
              }}
            >
              {speech.politicianName || "Anonym"}
            </Link>
            <div
              style={{
                marginTop: 6,
                display: "flex",
                gap: 8,
                alignItems: "center",
                flexWrap: "wrap",
              }}
            >
              <Link
                to={`/parties/${encodeURIComponent(speech.party)}`}
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                  fontWeight: 700,
                  padding: "2px 8px",
                  background: pc?.bg ?? "#888",
                  color: pc?.text ?? "#fff",
                  textDecoration: "none",
                  letterSpacing: "0.05em",
                }}
              >
                {speech.party}
              </Link>
              <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-fg-muted)" }}>
                {partyName}
              </span>
              {politician?.constituency && (
                <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-fg-muted)" }}>
                  · {politician.constituency}
                </span>
              )}
            </div>
          </div>
        </section>

        {/* Anförande meta */}
        <div
          style={{
            marginTop: 24,
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "0.15em",
            color: "var(--color-fg-muted)",
            textTransform: "uppercase",
          }}
        >
          ANFÖRANDE · {formatLongStamp(speech.date)}
        </div>
        {speech.topicHeading && (
          <h1
            style={{
              fontFamily: "var(--font-serif)",
              fontStyle: "italic",
              fontSize: isMobile ? 20 : 24,
              fontWeight: 400,
              margin: "8px 0 16px",
              lineHeight: 1.25,
              color: "var(--color-fg)",
            }}
          >
            {speech.topicHeading}
          </h1>
        )}

        {/* Om debatten */}
        {docToShow && (
          <section
            style={{
              padding: isMobile ? "12px 14px" : "14px 18px",
              background: "var(--color-sdt-surface)",
              border: "1px solid var(--color-border)",
              marginBottom: 24,
              display: "flex",
              flexDirection: "column",
              gap: 6,
            }}
          >
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.15em",
                color: "var(--color-fg-muted)",
                textTransform: "uppercase",
              }}
            >
              Om debatten
            </div>
            <div style={{ display: "flex", alignItems: "baseline", gap: 8, flexWrap: "wrap" }}>
              {docTypeLabel && (
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 10,
                    color: "var(--color-fg-muted)",
                    border: "1px solid var(--color-border)",
                    padding: "1px 6px",
                    letterSpacing: "0.05em",
                  }}
                >
                  {docTypeLabel}
                </span>
              )}
              {docLink ? (
                <Link
                  to={docLink}
                  style={{
                    fontFamily: "var(--font-serif)",
                    fontSize: 15,
                    color: "var(--color-fg)",
                    textDecoration: "none",
                    borderBottom: "1px dotted var(--color-border)",
                  }}
                >
                  {docToShow.title}
                </Link>
              ) : (
                <span style={{ fontFamily: "var(--font-serif)", fontSize: 15, color: "var(--color-fg)" }}>
                  {docToShow.title}
                </span>
              )}
            </div>
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

        {/* Speech body */}
        {speech.speechText ? (
          <div
            className="riksdagen-doc"
            style={{
              fontSize: 15,
              lineHeight: 1.7,
              color: "var(--color-fg)",
              marginBottom: 32,
            }}
            dangerouslySetInnerHTML={{ __html: speech.speechText }}
          />
        ) : (
          <p
            style={{
              fontSize: 15,
              lineHeight: 1.7,
              color: "var(--color-fg-muted)",
              fontStyle: "italic",
              marginBottom: 32,
            }}
          >
            {speech.snippet || "Anförandetexten är inte tillgänglig."}
          </p>
        )}

        {/* Resten av debatten */}
        {otherSpeeches.length > 0 && (
          <section
            style={{
              borderTop: "1px solid var(--color-border)",
              paddingTop: 24,
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
              Resten av debatten · {otherSpeeches.length} anföranden
            </div>
            <div>
              {otherSpeeches.map((s) => (
                <SpeechRow key={s.id} speech={s} />
              ))}
            </div>
          </section>
        )}
      </div>
    </div>
  );
}
```

**Note:** the `relatedDokId` field on `Speech` was added in the prior data
provenance work. If TypeScript complains it doesn't exist, open
`frontend/src/features/speeches/api.ts` and ensure the `Speech`
interface has `relatedDokId?: string`. If missing, add it. Backend
already returns the field per `domain.Speech.RelatedDokID`.

The double `useDocument` call is paranoia — the `document` variable
is bound to `undefined` in this snippet because Speech doesn't expose
`dokId`-as-related (only `dokId` = protokoll). Simplify by removing
the unused first call:

Replace:
```tsx
  const { data: document } = useDocument(speech?.dokId && undefined /* ... */ ? speech.dokId : undefined);
  const relDokId = (speech as unknown as { relatedDokId?: string })?.relatedDokId;
  const { data: doc2 } = useDocument(relDokId);
  const docToShow = doc2 ?? document;
```

with:
```tsx
  const relDokId = speech?.relatedDokId;
  const { data: docToShow } = useDocument(relDokId);
```

(After confirming `Speech.relatedDokId` is a typed field via Step 2 below.)

- [ ] **Step 2: Confirm `Speech.relatedDokId` is in the type**

```bash
grep -n "relatedDokId" /home/rune/Documents/swedish-democratic-tracker/frontend/src/features/speeches/api.ts
```

If absent, add the field to the `Speech` interface in
`frontend/src/features/speeches/api.ts`:

```ts
export interface Speech {
  id: number;
  dokId: string;
  anforandeNummer?: string;
  politicianId: string;
  politicianName: string;
  politicianImageUrl?: string;
  party: string;
  date: string;
  topicHeading?: string;
  snippet: string;
  speechText?: string;
  relatedDokId?: string;
}
```

Backend already returns the field; this just makes the TS type honest.

- [ ] **Step 3: Typecheck**

```bash
cd /home/rune/Documents/swedish-democratic-tracker/frontend
npx tsc --noEmit
```

Empty output expected. If errors mention `relatedDokId` or
`useSpeechesByDocument` argument types, fix per the messages — the
hook accepts `string | undefined` and gates with `enabled`.

- [ ] **Step 4: Commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add frontend/src/features/speeches/SpeechDetailPage.tsx frontend/src/features/speeches/api.ts
git commit -m "feat(speech-detail): politician card + Om debatten + transcript"
git push origin main
```

---

## Task 6: Politician page — restructure tabs

**Files:**
- Modify: `frontend/src/features/politicians/PoliticianPage.tsx`

The existing page has 3 tabs: `votes`, `promises`, `anforanden`. Change to 4 tabs: `aktivitet`, `debatter`, `amnen`, `loften`.

This is a moderate edit. Approach: keep the hero exactly as-is, swap the tab strip, replace the three tab bodies with four.

- [ ] **Step 1: Add imports**

At the top of `PoliticianPage.tsx`, alongside existing imports:

```tsx
import { useSearchParams } from "react-router-dom";
import { useDocument } from "@/hooks/useDemocracy";
import { SpeechRow } from "@/features/speeches/SpeechRow";
import type { Speech } from "@/features/speeches/api";
```

The existing imports already include `useSpeechesByPolitician` (added
in the previous Anföranden tab work) and the politicians API. If
`useSpeechesByPolitician` is missing, add:

```tsx
import { useSpeechesByPolitician } from "@/hooks/useDemocracy";
```

- [ ] **Step 2: Replace `activeTab` state with URL-driven tab**

Find the existing line (currently around line 103):

```tsx
const [activeTab, setActiveTab] = useState<"votes" | "promises" | "anforanden">("votes");
```

Replace with:

```tsx
type Tab = "aktivitet" | "debatter" | "amnen" | "loften";
const ALL_TABS: Tab[] = ["aktivitet", "debatter", "amnen", "loften"];

const [searchParams, setSearchParams] = useSearchParams();
const tabParam = searchParams.get("tab") as Tab | null;
const activeTab: Tab = tabParam && ALL_TABS.includes(tabParam) ? tabParam : "aktivitet";
const setActiveTab = (next: Tab) => {
  const params = new URLSearchParams(searchParams);
  if (next === "aktivitet") params.delete("tab");
  else params.set("tab", next);
  setSearchParams(params, { replace: true });
};
```

- [ ] **Step 3: Replace the speeches query limit**

Find the existing speeches query in this file (added during the
previous Anföranden tab work):

```tsx
const { data: speeches, isLoading: loadingSpeeches } = useSpeechesByPolitician(
  id,
  50,
);
```

Increase the limit so Debatter and Ämnen aggregations have enough
data:

```tsx
const { data: speeches, isLoading: loadingSpeeches } = useSpeechesByPolitician(
  id,
  200,
);
```

- [ ] **Step 4: Replace tab strip JSX**

Find the existing loop that renders the three tabs. Replace it with:

```tsx
{ALL_TABS.map((tab) => {
  const label =
    tab === "aktivitet" ? "Aktivitet"
    : tab === "debatter" ? "Debatter"
    : tab === "amnen" ? "Ämnen"
    : "Löften";
  const active = activeTab === tab;
  return (
    <button
      key={tab}
      onClick={() => setActiveTab(tab)}
      className="px-4 py-2.5 text-sm font-medium transition-colors cursor-pointer"
      style={{
        borderBottom: active ? "2px solid var(--color-on-surface)" : "2px solid transparent",
        color: active ? "var(--color-on-surface)" : "var(--color-on-surface-variant)",
        fontWeight: active ? 700 : 500,
        marginBottom: "-1px",
      }}
    >
      {label}
    </button>
  );
})}
```

(Uses the same `<button>` styling as the existing tabs.)

- [ ] **Step 5: Replace tab bodies**

Find the existing three blocks:

```tsx
{activeTab === "votes" && ( ... )}
{activeTab === "promises" && ( ... )}
{activeTab === "anforanden" && ( ... )}
```

Replace all three with the four blocks below. Keep references to
existing variables (`votes`, `loadingVotes`, `promises`, `loadingPromises`,
`speeches`, `loadingSpeeches`).

```tsx
{activeTab === "aktivitet" && (
  <ActivityFeed votes={votes} speeches={speeches ?? []} loading={loadingVotes || loadingSpeeches} />
)}

{activeTab === "debatter" && (
  <DebatterPanel speeches={speeches ?? []} loading={loadingSpeeches} />
)}

{activeTab === "amnen" && (
  <AmnenPanel speeches={speeches ?? []} loading={loadingSpeeches} />
)}

{activeTab === "loften" && (
  <div>
    {/* keep the existing löften body inline here — copy from the
       previous activeTab === "promises" block, which already renders
       loadingPromises and promises. Do not refactor its internals. */}
  </div>
)}
```

You'll move the existing löften JSX (currently under `activeTab === "promises"`)
into the `loften` block as-is. The promises query hook stays in the
component body.

- [ ] **Step 6: Add the three new sub-components above `PoliticianPage`**

Above the `export function PoliticianPage()` declaration, add three
sub-components used by the tabs:

```tsx
type ActivityEvent =
  | { kind: "vote"; date: string; vote: Vote }
  | { kind: "speech"; date: string; speech: Speech };

function buildActivity(votes: Vote[], speeches: Speech[]): ActivityEvent[] {
  const events: ActivityEvent[] = [];
  for (const v of votes) {
    events.push({ kind: "vote", date: v.date ?? "", vote: v });
  }
  for (const s of speeches) {
    events.push({ kind: "speech", date: s.date, speech: s });
  }
  events.sort((a, b) => (b.date || "").localeCompare(a.date || ""));
  return events.slice(0, 50);
}

function ActivityFeed({
  votes,
  speeches,
  loading,
}: {
  votes: Vote[];
  speeches: Speech[];
  loading: boolean;
}) {
  if (loading) {
    return <div className="text-on-surface-variant text-sm py-6 text-center">Laddar...</div>;
  }
  const events = buildActivity(votes, speeches);
  if (events.length === 0) {
    return (
      <p className="text-sm italic text-on-surface-variant py-4">
        Ingen registrerad aktivitet ännu.
      </p>
    );
  }
  return (
    <div className="rounded-xl p-4" style={{ background: "var(--color-surface-lowest)" }}>
      <div className="text-[10px] font-mono uppercase tracking-widest text-on-surface-variant mb-3">
        Senaste 50 händelser
      </div>
      <div>
        {events.map((e, i) =>
          e.kind === "speech" ? (
            <SpeechRow key={`s-${e.speech.id}`} speech={e.speech} hidePolitician />
          ) : (
            <ActivityVoteRow key={`v-${i}`} vote={e.vote} />
          ),
        )}
      </div>
    </div>
  );
}

function ActivityVoteRow({ vote }: { vote: Vote }) {
  return (
    <Link
      to={`/votes/${vote.beteckning}/${vote.forslagspunkt}`}
      style={{
        display: "flex",
        gap: 12,
        padding: "12px 0",
        borderBottom: "1px solid var(--color-border)",
        alignItems: "flex-start",
        textDecoration: "none",
        color: "inherit",
        minWidth: 0,
      }}
    >
      <span
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 9,
          letterSpacing: "0.05em",
          background: "var(--color-track)",
          color: "var(--color-fg-muted)",
          padding: "2px 6px",
          flexShrink: 0,
          height: 18,
        }}
      >
        RÖST
      </span>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            fontSize: 13,
            color: "var(--color-fg)",
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
            marginBottom: 4,
          }}
          title={vote.documentTitle ?? `${vote.beteckning} punkt ${vote.forslagspunkt}`}
        >
          {vote.documentTitle || `${vote.beteckning} punkt ${vote.forslagspunkt}`}
        </div>
        <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg-muted)" }}>
          {vote.beteckning} · {vote.voteResult}
          {vote.date ? ` · ${vote.date}` : ""}
        </div>
      </div>
    </Link>
  );
}

function DebatterPanel({ speeches, loading }: { speeches: Speech[]; loading: boolean }) {
  // Group by relatedDokId, sort by count desc, take top 20.
  const groups = new Map<string, { count: number; minDate: string; maxDate: string }>();
  for (const s of speeches) {
    if (!s.relatedDokId) continue;
    const g = groups.get(s.relatedDokId);
    if (!g) {
      groups.set(s.relatedDokId, { count: 1, minDate: s.date, maxDate: s.date });
    } else {
      g.count++;
      if (s.date < g.minDate) g.minDate = s.date;
      if (s.date > g.maxDate) g.maxDate = s.date;
    }
  }
  const top = Array.from(groups.entries())
    .sort((a, b) => b[1].count - a[1].count)
    .slice(0, 20);

  if (loading) {
    return <div className="text-on-surface-variant text-sm py-6 text-center">Laddar...</div>;
  }
  if (top.length === 0) {
    return (
      <p className="text-sm italic text-on-surface-variant py-4">
        Inga debatter har kopplats till anförandena ännu.
      </p>
    );
  }
  return (
    <div className="rounded-xl p-4" style={{ background: "var(--color-surface-lowest)" }}>
      <div className="text-[10px] font-mono uppercase tracking-widest text-on-surface-variant mb-3">
        {top.length} aktiva debatter
      </div>
      <div>
        {top.map(([dokId, info]) => (
          <DebatterRow key={dokId} dokId={dokId} count={info.count} minDate={info.minDate} maxDate={info.maxDate} />
        ))}
      </div>
    </div>
  );
}

function DebatterRow({
  dokId,
  count,
  minDate,
  maxDate,
}: {
  dokId: string;
  count: number;
  minDate: string;
  maxDate: string;
}) {
  const { data: doc } = useDocument(dokId);
  const title = doc?.title || dokId;
  const typeLabel = doc?.type ? doc.type.toUpperCase() : "";
  const link = doc?.type === "bet" && doc.beteckning ? `/beslut/${encodeURIComponent(doc.beteckning)}` : null;
  const range = minDate === maxDate ? minDate.slice(0, 10) : `${minDate.slice(0, 10)} → ${maxDate.slice(0, 10)}`;

  const body = (
    <div
      style={{
        display: "flex",
        gap: 12,
        padding: "12px 0",
        borderBottom: "1px solid var(--color-border)",
        alignItems: "flex-start",
        minWidth: 0,
      }}
    >
      {typeLabel && (
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 9,
            letterSpacing: "0.05em",
            background: "var(--color-track)",
            color: "var(--color-fg-muted)",
            padding: "2px 6px",
            flexShrink: 0,
            height: 18,
          }}
        >
          {typeLabel}
        </span>
      )}
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            fontSize: 14,
            color: "var(--color-fg)",
            fontFamily: "var(--font-serif)",
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
            marginBottom: 4,
          }}
          title={title}
        >
          {title}
        </div>
        <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg-muted)" }}>
          {count} anförande{count === 1 ? "" : "n"} · {range}
        </div>
      </div>
    </div>
  );

  return link ? (
    <Link to={link} style={{ textDecoration: "none", color: "inherit", display: "block" }}>
      {body}
    </Link>
  ) : (
    body
  );
}

function AmnenPanel({ speeches, loading }: { speeches: Speech[]; loading: boolean }) {
  const counts = new Map<string, { count: number; lastDate: string }>();
  for (const s of speeches) {
    const t = (s.topicHeading ?? "").trim();
    if (!t) continue;
    const c = counts.get(t);
    if (!c) {
      counts.set(t, { count: 1, lastDate: s.date });
    } else {
      c.count++;
      if (s.date > c.lastDate) c.lastDate = s.date;
    }
  }
  const top = Array.from(counts.entries())
    .sort((a, b) => b[1].count - a[1].count)
    .slice(0, 10);

  if (loading) {
    return <div className="text-on-surface-variant text-sm py-6 text-center">Laddar...</div>;
  }
  if (top.length === 0) {
    return (
      <p className="text-sm italic text-on-surface-variant py-4">
        Inga teman går att utläsa ännu.
      </p>
    );
  }
  return (
    <div className="rounded-xl p-4" style={{ background: "var(--color-surface-lowest)" }}>
      <div className="text-[10px] font-mono uppercase tracking-widest text-on-surface-variant mb-3">
        Återkommande teman
      </div>
      <div>
        {top.map(([topic, info]) => (
          <div
            key={topic}
            style={{
              padding: "10px 0",
              borderBottom: "1px solid var(--color-border)",
              minWidth: 0,
            }}
          >
            <div
              style={{
                fontSize: 13,
                color: "var(--color-fg)",
                marginBottom: 4,
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
              }}
              title={topic}
            >
              {topic}
            </div>
            <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg-muted)" }}>
              {info.count} anförande{info.count === 1 ? "" : "n"} · senast {info.lastDate.slice(0, 10)}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
```

- [ ] **Step 7: Typecheck**

```bash
cd /home/rune/Documents/swedish-democratic-tracker/frontend
npx tsc --noEmit
```

Expected: empty output.

If TS complains that `Vote` doesn't have `date`, look at the existing
import — `Vote` comes from `@/shared/types`. Open the type and add
`date?: string` if missing (the field is already in the API response;
just may not be in the TS interface). Backend `domain.Vote.Date`
ships in JSON as `date`.

- [ ] **Step 8: Commit**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
git pull --rebase origin main
git add frontend/src/features/politicians/PoliticianPage.tsx
git commit -m "feat(politicians): tab rework — Aktivitet · Debatter · Ämnen · Löften"
git push origin main
```

---

## Task 7: Verification

- [ ] **Step 1: Local typecheck + builds**

```bash
cd /home/rune/Documents/swedish-democratic-tracker
go build -C backend ./...
cd frontend && npx tsc --noEmit && npm run build
```

Expected: backend silent, tsc silent, vite reports `✓ built in ...`.

- [ ] **Step 2: Local smoke (`make run`)**

Open in browser:
- `/anforanden/974` — politician avatar 80×80, name links to `/politicians/:id`, party chip links to `/parties/:party`, "Om debatten" card renders with the related document title (links to /beslut/:beteckning when bet), "Resten av debatten" lists other speakers excluding current.
- `/politicians/{some-active-id}` — 4 tabs visible (Aktivitet · Debatter · Ämnen · Löften). Default Aktivitet shows up to 50 mixed votes + speeches newest first. Switch to Debatter — top 20 grouped debates with title from the documents endpoint. Switch to Ämnen — top 10 topic strings. Switch to Löften — same content as old promises tab.
- `/politicians/:id?tab=debatter` opens directly on Debatter tab.

- [ ] **Step 3: Curl smoke**

```bash
curl -s 'http://localhost:8080/api/documents/hd01au9' | python3 -m json.tool
```

Expected: JSON with `dokId`, `type: "bet"`, `title`, `beteckning: "AU9"`, etc.

- [ ] **Step 4: Mobile prod check**

Wait for CI green + ArgoCD rollout (~5-10 min). Use Playwright at 375 × 812 against:
- `https://sdt.runevibe.se/anforanden/974` — `overflowPx === 0`, politician card visible, transcript list renders.
- `https://sdt.runevibe.se/politicians/<some-id>` — `overflowPx === 0`, 4 tabs visible, Aktivitet renders mixed feed.
- `https://sdt.runevibe.se/politicians/<some-id>?tab=debatter` — Debatter tab renders directly on load.

- [ ] **Step 5: Tear down**

```bash
docker compose -f docker-compose.dev.yml down
```

---

## Self-review notes

- **Spec coverage:**
  - Backend `/api/documents/:dokId` → Tasks 1, 2.
  - Frontend `RiksdagDocument` + `votesApi.getDocument` → Task 3.
  - `useDocument` hook → Task 4.
  - SpeechDetailPage politician card + Om debatten + Resten av debatten → Task 5.
  - PoliticianPage tab rework → Task 6.
  - Verification → Task 7.
- **Type consistency:** `Speech.relatedDokId` declared in Task 5 Step 2, used by Tasks 5 and 6. `RiksdagDocument` declared in Task 3, used by Task 4 and consumed by Tasks 5 and 6 via the `useDocument` hook.
- **No new tests:** project still has no unit-test framework. Verification is `go build`, `tsc`, `npm run build`, curl, and Playwright at 375 × 812.
- **Existing tab body preservation:** Task 6 Step 5 explicitly reuses the existing löften JSX inline rather than refactoring. The extra ActivityVoteRow component is small and sub-100 lines so it stays in the page file rather than splitting into its own file.
- **`/politicians/:id` URL state:** mirrors the pattern from `PartyDetailPage` (already shipped). Default tab is `aktivitet` (mapped to no query param). Replace-mode setSearchParams preserves browser history sanity.
