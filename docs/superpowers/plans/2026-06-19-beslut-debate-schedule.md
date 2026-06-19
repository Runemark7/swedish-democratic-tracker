# Beslut Debate-Schedule + Notis Summary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Surface Riksdagen's scheduled debate/decision dates and plain-Swedish "Beslut i korthet" summary on thin `/beslut/<bet>` pages.

**Architecture:** Extend the live `dokumentstatus` parse in the `votes` feature to capture four extra `dokuppgift` fields, thread them through the existing ad-hoc detail JSON response, and render date-aware labels + the notis summary on `BeslutDetailPage`. No DB migration, no ingestion change, no OpenAPI change — the beslut detail response is an untyped `map[string]any`.

**Tech Stack:** Go 1.26 (chi, pgx — but no DB touched here), React 19 + TypeScript + Vite.

**Note on testing:** This repo has no test framework (per CLAUDE.md). Verification per task is `go build -C backend ./...` (backend) and `cd frontend && npx tsc --noEmit` (frontend), plus a final manual check against the live API. There are no unit-test steps because there is no harness to run them.

---

### Task 1: Backend — domain fields + dokumentstatus parse

**Files:**
- Modify: `backend/internal/votes/domain/vote.go` (DocumentStatus struct, ~line 54-68)
- Modify: `backend/internal/votes/adapters/riksdagen/client.go` (FetchDocumentStatus, ~line 137-225)

- [ ] **Step 1: Add fields to `DocumentStatus`**

In `backend/internal/votes/domain/vote.go`, inside the `DocumentStatus` struct, add four fields after `BodyHTML`:

```go
	BodyHTML string `json:"bodyHtml,omitempty"`
	// Chamber schedule + plain-language summary parsed from dokumentstatus.dokuppgift.
	// All raw Riksdagen values, shown verbatim (fact layer).
	DebattDate string `json:"debattDate,omitempty"` // dokuppgift kod=debattdatumtid, "YYYY-MM-DD HH:MM:SS"
	BeslutDate string `json:"beslutDate,omitempty"` // dokuppgift kod=beslutdatumtid
	StatusText string `json:"statusText,omitempty"` // dokuppgift kod=statustext
	Notis      string `json:"notis,omitempty"`      // dokuppgift kod=notis, "Beslut i korthet" HTML
```

- [ ] **Step 2: Decode `dokuppgift` in the payload struct**

In `backend/internal/votes/adapters/riksdagen/client.go` `FetchDocumentStatus`, add a `Dokuppgift` block to the anonymous `payload` struct, as a sibling of `Dokintressent` (after the `Dokintressent` closing brace, before the final `} \`json:"dokumentstatus"\``):

```go
			Dokuppgift struct {
				Uppgift []struct {
					Kod  string `json:"kod"`
					Text string `json:"text"`
				} `json:"uppgift"`
			} `json:"dokuppgift"`
```

(`dokuppgift.uppgift` is consistently a JSON array for betänkanden — decode directly as a slice.)

- [ ] **Step 3: Populate the new fields after parsing intressenter**

In the same function, after the `for _, i := range ds.Dokintressent.Intressent { ... }` loop and before the `if len(ds.Dokintressent.Intressent) > 0 {` block, add:

```go
	for _, u := range ds.Dokuppgift.Uppgift {
		switch u.Kod {
		case "debattdatumtid":
			status.DebattDate = u.Text
		case "beslutdatumtid":
			status.BeslutDate = u.Text
		case "statustext":
			status.StatusText = u.Text
		case "notis":
			status.Notis = u.Text
		}
	}
```

- [ ] **Step 4: Verify build**

Run: `go build -C backend ./...`
Expected: no output, exit 0.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/votes/domain/vote.go backend/internal/votes/adapters/riksdagen/client.go
git commit -m "feat(votes): parse dokuppgift debate/beslut dates + notis from dokumentstatus"
```

---

### Task 2: Backend — expose fields in detail response

**Files:**
- Modify: `backend/internal/votes/adapters/http/handler.go` (two response branches, ~line 178-183 and ~line 208-213)

- [ ] **Step 1: Add fields to the no-votes (metadata) branch**

In `backend/internal/votes/adapters/http/handler.go`, the `if len(vv) == 0 {` branch has this block:

```go
		if ds, err := h.svc.GetDocumentStatus(r.Context(), info.DokID); err == nil && ds != nil {
			resp["subtitle"] = ds.Subtitle
			resp["summary"] = ds.Summary
			resp["bodyHtml"] = ds.BodyHTML
		}
```

Replace it with:

```go
		if ds, err := h.svc.GetDocumentStatus(r.Context(), info.DokID); err == nil && ds != nil {
			resp["subtitle"] = ds.Subtitle
			resp["summary"] = ds.Summary
			resp["bodyHtml"] = ds.BodyHTML
			resp["debattDate"] = ds.DebattDate
			resp["beslutDate"] = ds.BeslutDate
			resp["statusText"] = ds.StatusText
			resp["notis"] = ds.Notis
		}
```

- [ ] **Step 2: Add fields to the votes-in-DB branch**

Further down, the votes-in-DB branch has:

```go
	if ds, err := h.svc.GetDocumentStatus(r.Context(), meta.DokID); err == nil && ds != nil {
		resp["date"] = ds.Date
		resp["subtitle"] = ds.Subtitle
		resp["summary"] = ds.Summary
		resp["bodyHtml"] = ds.BodyHTML
	}
```

Replace it with:

```go
	if ds, err := h.svc.GetDocumentStatus(r.Context(), meta.DokID); err == nil && ds != nil {
		resp["date"] = ds.Date
		resp["subtitle"] = ds.Subtitle
		resp["summary"] = ds.Summary
		resp["bodyHtml"] = ds.BodyHTML
		resp["debattDate"] = ds.DebattDate
		resp["beslutDate"] = ds.BeslutDate
		resp["statusText"] = ds.StatusText
		resp["notis"] = ds.Notis
	}
```

- [ ] **Step 3: Verify build**

Run: `go build -C backend ./...`
Expected: no output, exit 0.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/votes/adapters/http/handler.go
git commit -m "feat(votes): include debate/beslut dates + notis in beslut detail response"
```

---

### Task 3: Frontend — types + date-aware Debatten label

**Files:**
- Modify: `frontend/src/shared/types.ts` (VoteDetail interface, ~line 93-111)
- Modify: `frontend/src/features/votes/BeslutDetailPage.tsx` (helper + empty-Debatten branch, ~line 454-466)

- [ ] **Step 1: Add optional fields to `VoteDetail`**

In `frontend/src/shared/types.ts`, inside `interface VoteDetail`, add after `bodyHtml?: string;`:

```ts
  /** Riksdagen dokuppgift — raw scheduled debate/decision datetimes
   *  ("YYYY-MM-DD HH:MM:SS") and lifecycle text. Fact layer, shown verbatim. */
  debattDate?: string;
  beslutDate?: string;
  statusText?: string;
  /** "Beslut i korthet" — plain-Swedish summary (HTML) from Riksdagen. */
  notis?: string;
```

- [ ] **Step 2: Add a Swedish date helper to BeslutDetailPage**

In `frontend/src/features/votes/BeslutDetailPage.tsx`, after the `committeeLabel` function (ends ~line 31), add:

```tsx
// Parse a Riksdagen "YYYY-MM-DD HH:MM:SS" datetime to a Date (date part only).
// Returns null for empty/unparseable input.
function parseRiksdagDate(raw?: string): Date | null {
  if (!raw) return null;
  const datePart = raw.slice(0, 10);
  const d = new Date(`${datePart}T00:00:00`);
  return isNaN(d.getTime()) ? null : d;
}

const SV_DATE = new Intl.DateTimeFormat("sv-SE", {
  day: "numeric",
  month: "long",
  year: "numeric",
});

// Label for the empty Debatten section, derived purely from sourced dates.
// Never asserts "no debate happened" — states the date + that no anföranden
// are registered, leaving interpretation to the reader.
function debattFallbackLabel(debattDate?: string): string {
  const d = parseRiksdagDate(debattDate);
  if (!d) return "Debatt ännu inte planerad";
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  if (d.getTime() > today.getTime()) {
    return `Debatt planerad: ${SV_DATE.format(d)}`;
  }
  return `Debatten hölls ${SV_DATE.format(d)} · inga anföranden registrerade här`;
}
```

- [ ] **Step 3: Use the label in the empty-Debatten branch**

In the same file, the Debatten section currently renders this when there are no speeches:

```tsx
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
```

Replace the inner text `Inga registrerade anföranden för det här beslutet ännu.` with:

```tsx
              {debattFallbackLabel(data?.debattDate)}
```

- [ ] **Step 4: Verify typecheck**

Run: `cd frontend && npx tsc --noEmit`
Expected: no output, exit 0.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/shared/types.ts frontend/src/features/votes/BeslutDetailPage.tsx
git commit -m "feat(beslut): show scheduled debate date in empty Debatten section"
```

---

### Task 4: Frontend — prefer notis in Sammanfattning card

**Files:**
- Modify: `frontend/src/features/votes/BeslutDetailPage.tsx` (Sammanfattning section, ~line 329-353)

- [ ] **Step 1: Render notis HTML when present, else fall back**

In `frontend/src/features/votes/BeslutDetailPage.tsx`, the Sammanfattning section body currently is:

```tsx
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
```

Replace the whole `{data?.summary ? (...) : (...)}` expression with a three-way choice that prefers `notis` (Riksdagen's plain-Swedish "Beslut i korthet", trusted government HTML — rendered via the existing `DocumentBody` wrapper):

```tsx
          {data?.notis ? (
            <DocumentBody html={data.notis} />
          ) : data?.summary ? (
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
```

(`DocumentBody` is already defined in this file, ~line 88, and applies the `riksdagen-doc` style wrapper. The default `DocumentBody` has `marginBottom: 16` which is acceptable inside the card; no extra style needed.)

- [ ] **Step 2: Verify typecheck**

Run: `cd frontend && npx tsc --noEmit`
Expected: no output, exit 0.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/features/votes/BeslutDetailPage.tsx
git commit -m "feat(beslut): prefer 'Beslut i korthet' notis in Sammanfattning card"
```

---

### Task 5: Final verification

- [ ] **Step 1: Full backend build**

Run: `go build -C backend ./...`
Expected: exit 0.

- [ ] **Step 2: Full frontend typecheck + lint**

Run: `cd frontend && npx tsc --noEmit && npm run lint`
Expected: exit 0 (lint may warn but must not error on touched files).

- [ ] **Step 3: Manual smoke test against the live API**

With the dev stack running (`make run`), confirm the response carries the new fields:

Run:
```bash
curl -s "https://data.riksdagen.se/dokumentstatus/hd01ku45.json" | grep -o '"kod":"debattdatumtid"'
```
Expected: `"kod":"debattdatumtid"` (confirms upstream field still present).

Then in the browser open `/beslut/KU45?...` and confirm:
- Sammanfattning card shows the "Beslut i korthet" text (KU bedömer att…), not the empty-summary line.
- Debatten section shows `Debatten hölls 17 juni 2026 · inga anföranden registrerade här` (KU45 debate date is in the past).

For an upcoming betänkande (status "klart för beslut", future debattdatumtid), the Debatten section should read `Debatt planerad: <date>`.

- [ ] **Step 4: Confirm done**

All five tasks committed; build + typecheck green; manual check matches expected labels.
