# Handoff: Tre Kammare — Swedish Democratic Tracker Redesign

## Overview

This package contains the design for a UX-focused redesign of the Swedish Democratic Tracker (sdt.runevibe.se). The goal is to help Swedish citizens **see what's happening right now** across the three levels of Swedish democracy — **Riksdag** (parliament, national), **Region** (regional, e.g. vård & kollektivtrafik) and **Kommun** (municipal, e.g. skola & omsorg) — with a clear level-first mental model.

The visual concept is called **"Tre Kammare"** (Three Chambers): each level of government is a chamber with its own live pulse of decisions. Copy is strictly neutral Swedish.

---

## About the Design Files

The files in `references/` are **design references** created as an HTML/JSX prototype (React 18 + Babel-standalone + inline styles). **They are not production code to copy directly.** Your job is to **recreate these designs in the target codebase** (React 19 + Vite 6 + TypeScript 5, Tailwind CSS v4, React Router v7, TanStack Query v5) using its established patterns.

What to carry over:
- **Visual system** (colors, type, spacing, layout, motion) — pixel-accurate
- **Component structure & data shape** — use as blueprint
- **Interactions, copy, and information hierarchy** — preserve exactly

What to throw away:
- Inline `style={{…}}` objects → **port to Tailwind utility classes + CSS variables for themeable tokens**
- `window.SDT_DATA` global → **TanStack Query hooks fetching from your API**
- Babel-standalone JSX → **proper TSX modules**

---

## Fidelity

**High-fidelity.** Pixel-perfect mockups with final colors, typography, spacing, motion, and interaction copy. Recreate exactly, substituting only the implementation technology.

---

## Target Stack Guidance

- **React 19 + TS 5** — prefer server/client split where it makes sense; types below
- **Vite 6** — no special notes
- **Tailwind CSS v4** — use `@theme` in a global CSS file to declare the design tokens listed below; expose them as CSS vars so both light/dark themes work
- **React Router v7** — file-based or declarative routes as you prefer; see **Routes** section
- **TanStack Query v5** — one query per data resource (see **Data Layer**); stale-time 30s for "live" data so the UI feels up-to-date without hammering the server

---

## Routes

| Path | Component | Purpose |
|---|---|---|
| `/` | `HomePage` | "Tre kammare" landing — concentric hero + 3 level status cards + 24h pulse strip |
| `/riksdag` | `RiksdagPage` | National parliament — mandates, budget, live votes, agenda |
| `/region` | `RegionIndexPage` | Region selector (21 regions). Defaults to user's region if geolocated, else Stockholm |
| `/region/:regionId` | `RegionPage` | Single region — KPIs with target bars, ruling coalition, budget, live decisions |
| `/kommun` | `KommunIndexPage` | Kommun selector (290 kommuner). Defaults to user's kommun, else Malmö |
| `/kommun/:kommunId` | `KommunPage` | Single kommun — same structure as Region |
| `/sok` | `SearchPage` | Cross-level exploration by topic or party |

The `:regionId` and `:kommunId` params should be slugs (e.g. `stockholm`, `malmo`, `vastra-gotaland`).

---

## Data Layer

Define TypeScript types for the shapes below, then one TanStack Query hook per level.

```typescript
// types/democracy.ts

export interface Party {
  name: string;        // "Moderaterna"
  short: string;       // "M"
  seats: number;
  color: string;       // hex — use official SE party colors (already in data.jsx)
}

export interface Ruling {
  type: string;        // "Regeringskoalition (stöd av SD)" / "Blågrön majoritet" / etc.
  parties: Party[];    // governing parties
  support?: Party[];   // parties providing support (e.g. SD to the current regering)
  opposition: Party[];
}

export interface LiveVote {
  time: string;        // "Idag 14:20" / "Igår 16:40" / "17 apr" — localized SE
  title: string;
  status: 'Bifall' | 'Avslag' | 'Återremiss' | 'Bordlagd';
  margin?: string;     // "196–149" (Riksdag only)
  tag: string;         // topic: "Välfärd" / "Skatt" / "Vård" / ...
}

export interface BudgetArea { name: string; value: number; pct: number }
export interface Budget { total: string; year: string; areas: BudgetArea[] }

export interface Kpi {
  label: string;
  value: string;       // formatted for display: "3 h 12 min"
  raw: number;         // numeric value for bar math
  target: number;
  worseHigher: boolean;// true if higher = worse (e.g. wait time)
  unit: string;        // "h" / "%" / "v" / ""
  trend: 'up' | 'down' | 'flat';
  delta: string;       // "+8 min" / "−0,5 pp"
  note: string;        // "Mål: under 3 h"
}

export interface Authority {
  name: string;        // "Försäkringskassan"
  role: string;
  headcount: string;
  budget: string;
}

export interface LevelData {
  title: string;
  subtitle: string;
  population?: string;
  ruling: Ruling;
  liveVotes: LiveVote[];
  budget: Budget;
  agenda: string[];
  kpis?: Kpi[];        // Region + Kommun only
  authorities?: Authority[];  // Riksdag only
}
```

Query hooks:

```typescript
// hooks/useDemocracy.ts
export const useRiksdag     = () => useQuery({ queryKey: ['riksdag'],         queryFn: fetchRiksdag,    staleTime: 30_000 });
export const useRegion      = (id: string) => useQuery({ queryKey: ['region', id], queryFn: () => fetchRegion(id), staleTime: 30_000 });
export const useKommun      = (id: string) => useQuery({ queryKey: ['kommun', id], queryFn: () => fetchKommun(id), staleTime: 30_000 });
export const useRegionList  = () => useQuery({ queryKey: ['regions'], queryFn: fetchRegions });
export const useKommunList  = () => useQuery({ queryKey: ['kommuner'], queryFn: fetchKommuner });
```

Seed data that exactly matches the prototype is in `references/data.jsx` — port it to a typed `mock-data.ts` for first implementation, then replace `fetchX` with real API calls (Riksdagens öppna data, SCB, kommun-API:er).

---

## Design Tokens

Port these to `@theme` in your Tailwind v4 global CSS. All colors use a light/dark pair — the prototype defaults to **dark**.

### Colors

```css
@theme {
  /* Dark (default) */
  --color-bg:         #0a0e14;
  --color-surface:    #10151d;
  --color-fg:         #eef1f5;
  --color-fg-muted:   #7c8896;
  --color-border:     rgb(238 241 245 / 0.14);
  --color-track:      rgb(238 241 245 / 0.08);
  --color-accent:     #6fb0f2;   /* cool civic blue */
  --color-accent-2:   #e0bc5e;   /* gold for level numerals */
  --color-pulse:      #e87560;   /* "beslut idag" red-orange */
  --color-up:         #5fbb7f;
  --color-down:       #e07a62;
  --color-plane:      rgb(238 241 245 / 0.04);
}

[data-theme="light"] {
  --color-bg:         #f5f3ee;
  --color-surface:    #ffffff;
  --color-fg:         #0e1620;
  --color-fg-muted:   #5a6470;
  --color-border:     rgb(14 22 32 / 0.14);
  --color-track:      rgb(14 22 32 / 0.06);
  --color-accent:     #0b3d7a;
  --color-accent-2:   #c8a13b;
  --color-pulse:      #d0533f;
  --color-up:         #2d7a4a;
  --color-down:       #b8391c;
  --color-plane:      rgb(14 22 32 / 0.04);
}
```

### Typography

Three families — load via `@fontsource` or Google Fonts:

| Role | Family | Fallback |
|---|---|---|
| **Display** (headlines, level numerals, coalition names) | `"GT Sectra"` or `"Tiempos Headline"` | `Georgia, serif` |
| **Body / UI** | `"Söhne"` or `"Inter"` | `system-ui, sans-serif` |
| **Mono** (labels, timestamps, stats) | `"Berkeley Mono"` or `"JetBrains Mono"` | `ui-monospace, monospace` |

The display face is used italic for level numerals (I, II, III) in gold accent color.

### Spacing / Sizing

- Page gutter: `32px` on 1280-wide artboards
- Panel padding: `20–22px`
- Section grids: 1px dividers between cards (use `gap: 1px; background: var(--color-border)` trick)
- Label uppercase letter-spacing: `0.1em–0.2em`
- Border radius: `2–3px` on inner elements, `6px` on buttons, `999px` on chips

### Motion

- Pulse dots: `d3pulse` keyframe — opacity 1 ↔ 0.35, 2–2.4s infinite ease
- Panel transitions: 300ms `ease-in-out`
- Number deltas: animate with TanStack Query's `previousData` (optional polish)

---

## Party Colors (authoritative)

```typescript
export const PARTY_COLORS = {
  S:  '#E8112D',  // Socialdemokraterna
  M:  '#1E88E5',  // Moderaterna
  SD: '#DDD014',  // Sverigedemokraterna
  V:  '#AF0000',  // Vänsterpartiet
  C:  '#009933',  // Centerpartiet
  KD: '#00558E',  // Kristdemokraterna
  L:  '#006AB3',  // Liberalerna
  MP: '#83CF39',  // Miljöpartiet
} as const;
```

---

## Screens

Artboards are designed at **1280 × 880** but the layout is fluid; at ≥1024 use side-by-side grid, ≤1024 stack vertically.

### 1. `HomePage` (`/`)

**Purpose**: In 5 seconds, answer "what's happening in Swedish politics right now?" and route the user into a specific level.

**Layout (top → bottom)**:
1. **Top nav** — logo lockup ("Tre Kammare"), 5 pills (Start · Riksdag · Region · Kommun · Sök), live date + pulse on right
2. **Hero** — oversized display headline *"Tre **kammare**, ett samhälle."* (second word in gold italic). Mono eyebrow above: "VEM BESTÄMMER · VAR · JUST NU"
3. **Concentric-frames diagram** — three nested bordered rectangles representing Riksdag → Region → Kommun containment. Each has a tab-label on its top-left edge (level numeral in gold italic + name + mono sub), a ruling-party color strip on top-right, and 3–5 pulse dots along the bottom edge (red = beslut idag, blue = pågår). Center shows total counts.
4. **3-column status strip** — one column per level. Each has:
   - Level numeral + title + "● N IDAG" (mono, in pulse color)
   - Full stack-bar (flat) showing parliament composition
   - STYRE label + coalition name + colored party chips + "MAJORITET N/total"
   - 2 latest votes with time, title, status pill
   - "GÅ TILL KAMMARE →" link in accent color
5. **24h pulse strip** — 48 cells = 30min buckets for the last 24h. Empty = track color. Hit cells are colored by level (I = pulse red, II = accent blue, III = accent gold). Labels: 00:00 / 06:00 / 12:00 / 18:00 / NU

**Exact component reference**: `D3Home` in `references/direction3-novel.jsx`.

### 2. `RiksdagPage` (`/riksdag`) · `RegionPage` (`/region/:id`) · `KommunPage` (`/kommun/:id`)

All three share the **same layout skeleton** (`D3LevelPage` in the reference). Differences:
- **Riksdag**: no KPI strip (doesn't fit national data well); has `authorities` section instead
- **Region / Kommun**: KPI strip with 3 target bars at the top

**Layout**:

1. **Top nav** (active state matches the level)
2. **Hero** — giant display numeral (I/II/III) in gold italic at `88px`, next to a block with:
   - Mono eyebrow: "KAMMARE ETT/TVÅ/TRE"
   - Display title: "Riksdagen" / "Region Stockholm" / "Malmö kommun" — `52px`
   - Subtitle with context (mandatperiod, population, …)
   - For Region/Kommun: a "↓ BYT REGION/KOMMUN" affordance in accent color (click → open selector)
3. **KPI strip** (Region + Kommun only) — 3 cards side-by-side separated by 1px dividers. Each card:
   - Mono label (uppercase)
   - Display value (30px, tabular nums) + trend indicator (↑/↓/→ with delta, colored)
   - **Target bar** (see charts spec) with scale `[lo, target, hi]` and green "good range" band — see `SDTTargetBar` in `references/charts.jsx`
4. **Two-up grid — Mandat + Budget**
   - **Left card (Mandat)**: Mono label, **hemicycle** chart of parliament seats (6 rows, 60–349 seats), party legend row below (dot + short + seats), then coalition name in display italic + "MAJORITET N/total" mono
   - **Right card (Budget)**: Mono label with year + total, then donut (size 150, thickness 18, center label = total amount) on the left and top-5 horizontal bar list on the right
5. **Bottom grid — Puls + Agenda** (1.3fr / 1fr)
   - **Puls**: live-vote list. Each row: red dot (today) or blue (earlier) · mono timestamp · title + tag · status pill · mono margin
   - **Agenda**: numbered list (gold display italic numerals) of 3–5 priority items from the ruling coalition

### 3. `SearchPage` (`/sok`)

**Purpose**: Let users start from a topic or party and see cross-level results.

**Layout**:
1. Nav
2. Hero — "Vad **händer** om…" display headline (gold italic word)
3. Search input (mono placeholder with example terms)
4. Two-up grid:
   - **Ämnen** (left): 2-col grid of topic cards. Each: topic name, level numerals (I/II/III in gold) indicating which levels touch it, count of related items
   - **Partier** (right): flex-wrap row of party pills (color dot + short + seat count), then "SENASTE FRÅN DIN KOMMUN" list of 3 items

Exact reference: `D3Search` in `references/direction3-novel.jsx`.

---

## Chart Primitives

Port these from `references/charts.jsx` to TSX components under `src/components/charts/`. All are inline SVG and theme-aware via CSS vars.

| Component | Props | Purpose |
|---|---|---|
| `<Donut>` | `segments: {color, value}[]`, `size=180`, `thickness=28`, `label`, `sublabel` | Budget breakdown. Auto-sizes center label to fit the hole. |
| `<StackBar>` | `segments: {color, value, name?}[]`, `height=14`, `rounded=true` | Parliament composition as single horizontal bar |
| `<HBars>` | `items: {name, value, color?, pct?}[]`, `max?`, `unit=""` | Ranked horizontal bar list |
| `<Hemicycle>` | `groups: {color, count}[]`, `width=320`, `height=140` | Parliament seating arc (6 rows, left = opposition → right = government) |
| `<Sparkline>` | `points: number[]`, `width`, `height`, `color`, `fill` | Trend line with optional filled area |
| `<Trend>` | `trend: 'up'\|'down'\|'flat'`, `delta: string` | Arrow + delta label |
| `<Pill>` | `tone: 'neutral'\|'pass'\|'fail'\|'pending'`, `children` | Status tag |
| `<TargetBar>` | `value, target, min?, max?, worseHigher?, unit?` | KPI progress vs goal with "good range" band, target tick, current marker |

Study the reference implementations carefully — `<TargetBar>` in particular encodes the "am I on track?" logic (whether higher is better) via the `worseHigher` flag, and the "good range" band + marker color reflect it.

---

## Interactions

- **Level pills in top nav** → React Router navigation
- **"GÅ TILL KAMMARE →"** on homepage cards → navigate to respective level page
- **Region/Kommun switcher** → opens a dropdown/dialog with a searchable list of all 21 regions or 290 kommuner, with user's location pre-selected if available
- **Pulse dots & live-vote rows** → click to open a detail drawer with full decision text, full vote breakdown per party, and links to official source
- **Hover states**: slight border brightening + cursor pointer on interactive elements; no scale transforms

---

## Theme Toggle

Implement as a user-pref stored in `localStorage` with key `sdt-theme`. Default to **dark** (Tre Kammare's canonical mode). Expose a minimal toggle in the top nav's right side — sun/moon glyph. Respect `prefers-color-scheme` if no pref is stored.

Set `data-theme="light"` on `<html>` to switch.

---

## State Management

Almost entirely server state (TanStack Query). Client state:
- Theme preference (`localStorage` + `useState` in a small context)
- Selected region/kommun slug (URL param, not state)
- Search input (local state)
- Dialog open/close (local state)

No Redux / Zustand needed.

---

## Accessibility

- All charts need a `<title>` and `<desc>` element for screen readers (e.g. hemicycle: "Riksdagens sammansättning: S 107, M 68, SD 73, …")
- Party chips have `aria-label="Moderaterna, 68 mandat"`
- KPI target bars have `aria-valuenow / valuemin / valuemax` + a visually-hidden sentence: "Väntetid akut är 3 tim 12 min, 12 minuter över målet 3 tim"
- Pulse animations should respect `prefers-reduced-motion`

---

## Files in this Bundle

- `screenshots/` — reference captures of all 5 pages (dark theme) + home in light theme. Use these to verify visual fidelity as you build.
  - `01-home-dark.png` — HomePage
  - `02-riksdag-dark.png` — Riksdag level page
  - `03-region-dark.png` — Region level page
  - `04-kommun-dark.png` — Kommun level page
  - `05-search-dark.png` — Search / cross-level explore
  - `06-home-light.png` — HomePage, light theme
- `references/prototype.html` — runnable entry (open in a browser to see the full canvas)
- `references/direction3-novel.jsx` — all page components (`D3Home`, `D3Riksdag`, `D3Region`, `D3Kommun`, `D3Search`, `D3LevelPage`, `D3Nav`, `D3ChambersHero`)
- `references/charts.jsx` — all chart primitives (port these first — everything else depends on them)
- `references/data.jsx` — seed data exactly matching the prototype. Use for your TS types and your mock API during implementation.

---

## Suggested Implementation Order

1. **Design tokens** — get Tailwind `@theme` set up, fonts loaded, dark/light toggle working
2. **Chart primitives** — port `references/charts.jsx` to `src/components/charts/*.tsx`. Test each in isolation.
3. **Types + mock data** — port `references/data.jsx` to `src/mock/democracy.ts` with the types above
4. **TanStack Query hooks** — point at mock data first; real APIs later
5. **Layout shell** — top nav + `<Outlet />` + theme provider
6. **Pages** — Riksdag first (most complete data), then Region, Kommun, Home, Search

---

## Open Questions to Clarify With Product

1. **Data sources**: Riksdagens öppna data API for national votes, SCB for stats — confirmed? Kommun/region APIs vary wildly; pick a gov-provided aggregator if possible
2. **Kommun/region switcher UX**: full 290-item searchable list vs. map-based picker?
3. **Authorities section on Riksdag**: is it valuable to show Försäkringskassan etc. here, or does it belong on a separate `/myndigheter` page?
4. **Historical data**: only "today" + sparkline trends, or full history drill-down?
