# Municipal KPI reference

This page documents every Kolada KPI and SCB dataset used by Riksdagskollen's
municipality pages. Use it to audit a specific number: each row links to the
upstream source so you can verify the value we display matches what Kolada /
SCB actually publish.

All KPIs are fetched live from [api.kolada.se](https://api.kolada.se/v3) on
each request. There is no caching layer and no stored copy in our database —
the source of truth is always the upstream API.

---

## Data quality conventions

Kolada returns a `status` flag on every value:

| Flag | Meaning | Our handling |
|---|---|---|
| (empty) / `N` | Normal, validated value | Displayed as-is |
| `U` | Uncertain / provisional | Displayed with amber `*`, tooltip explains |
| `M` | Missing for this municipality+year | Row hidden / shown as `–` |

We only extract `gender="T"` (total) values. If a KPI is reported only with
gender disaggregation and has no total row, the value will be missing —
let us know if you spot this for any of the KPIs below.

---

## Detail page KPIs (Verksamhet + Budget & Ekonomi)

Used on `/kommun/{code}` and `/kommun/jämför`.

| Code | Title | Unit | Kolada |
|---|---|---|---|
| N00900 | Kommunalskatt (total, inkl. landsting) | % | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N00900) |
| N11037 | Nettokostnad förskola per inskrivet barn | kr | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N11037) |
| N15027 | Kostnad grundskola per elev | kr | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N15027) |
| N20043 | Kostnad äldreomsorg per invånare | kr/inv | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N20043) |
| N03010 | Skatteintäkter per invånare | kr/inv | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N03010) |
| N03007 | Årets resultat | kr/inv | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N03007) |
| N03102 | Årets resultat som andel av skatt & statsbidrag | % | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N03102) |
| N03106 | Soliditet | % | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N03106) |
| N03040 | Skulder totalt | kr/inv | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N03040) |
| N03132 | Nettoinvesteringar totalt | kr/inv | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N03132) |

**Wired in**: `backend/internal/regions/service.go` (`defaultKPIs`).

---

## Spending KPIs ("Var går pengarna?")

Used to build the donut chart on the detail page. All in `kr/inv`.
Years fetched: 2019, 2020, 2021, 2022, 2023.

| Code | Sector | Kolada |
|---|---|---|
| N11004 | Förskola | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N11004) |
| N15028 | Grundskola | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N15028) |
| N17014 | Gymnasieskola | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N17014) |
| N20014 | Äldreomsorg | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N20014) |
| N30005 | Individ- & familjeomsorg | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N30005) |
| N07037 | Infrastruktur & skydd | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N07037) |
| N09022 | Kultur & fritid | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N09022) |
| N05011 | Nämnd & administration | [kolada.se](https://www.kolada.se/verktyg/fri-sokning/?kpis=N05011) |

**Wired in**: `backend/internal/regions/service.go` (`spendingKPIs`).

---

## Population trend

Befolkningsutveckling chart on the detail page.

| Source | Dataset | API |
|---|---|---|
| SCB | `BE/BE0101/BE0101A/BefolkningNy`, `ContentsCode=BE0101N1` (Folkmängd) | [api.scb.se](https://www.statistikdatabasen.scb.se/pxweb/sv/ssd/START__BE__BE0101__BE0101A/BefolkningNy/) |

**Wired in**: `backend/internal/regions/adapters/scb/client.go`.

---

## Reporting a discrepancy

If a number on the site doesn't match what kolada.se shows for the same
municipality and year:

1. Note the municipality code, KPI code, and year.
2. Click the Kolada link for that KPI above.
3. Open an issue at
   [github.com/AlexanderRunemark/swedish-democratic-tracker/issues](https://github.com/AlexanderRunemark/swedish-democratic-tracker/issues)
   with both values (ours vs Kolada's).

Large year-over-year changes for small municipalities (especially
Nettoinvesteringar and Årets resultat) are usually real — a single
construction project or one-off cost can swing the per-capita figure by
>100%. The source links exist precisely so you can verify this.
