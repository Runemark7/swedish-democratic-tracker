import { useParams, NavLink } from "react-router-dom";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { PartyBadge, StatBlock } from "@/shared/components";
import { PARTY_COLORS } from "@/shared/design";
import { municipalitiesApi } from "./api";
import type { ElectionResult, MunicipalityKPIItem } from "@/shared/types";

const KPI_LABELS: Record<string, string> = {
  N00900: "Kommunalskatt (total, inkl. landsting)",
  N11037: "Nettokostnad förskola per inskrivet barn",
  N15027: "Kostnad grundskola per elev",
  N20043: "Kostnad äldreomsorg per invånare",
  N03010: "Skatteintäkter per invånare",
  N03007: "Årets resultat",
  N03102: "Årets resultat som andel av skatt & statsbidrag",
  N03106: "Soliditet",
  N03040: "Skulder totalt",
  N03132: "Nettoinvesteringar totalt",
};
const KPI_UNITS: Record<string, string> = {
  N00900: "%",
  N11037: "kr",
  N15027: "kr",
  N20043: "kr/inv",
  N03010: "kr/inv",
  N03007: "kr/inv",
  N03102: "%",
  N03106: "%",
  N03040: "kr/inv",
  N03132: "kr/inv",
};
const VERKSAMHET_KPI_ORDER = ["N00900", "N11037", "N15027", "N20043", "N03010"];
const BUDGET_KPI_ORDER = ["N03007", "N03102", "N03106", "N03040", "N03132"];

function MandateBar({ results }: { results: ElectionResult[] }) {
  if (!results.length) return null;
  const total = results[0].totalMandates;
  return (
    <div className="space-y-1.5">
      <div className="flex h-6 rounded overflow-hidden bg-surface-high">
        {results.map((r) => {
          const pct = (r.mandates / total) * 100;
          const color = PARTY_COLORS[r.party]?.bg ?? "#9ca3af";
          return (
            <div
              key={r.party}
              title={`${r.party}: ${r.mandates} mandat (${r.votePct.toFixed(1)}%)`}
              style={{ width: `${pct}%`, background: color, minWidth: pct > 0.5 ? 2 : 0 }}
            />
          );
        })}
      </div>
      <div className="flex flex-wrap gap-x-3 gap-y-1.5">
        {results.map((r) => {
          const color = PARTY_COLORS[r.party]?.bg ?? "#9ca3af";
          return (
            <div key={r.party} className="flex items-center gap-1">
              <div className="w-2.5 h-2.5 rounded-sm" style={{ background: color }} />
              <span className="text-[11px] text-on-surface-variant">
                <span className="font-bold text-on-surface">{r.party}</span>
                {" "}{r.mandates}m · {r.votePct.toFixed(1)}%
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function UncertainMark({ status }: { status?: string }) {
  if (status !== "U") return null;
  return (
    <span
      className="ml-1 text-amber-500 font-bold cursor-help"
      title="Kolada flaggar detta värde som osäkert"
    >
      *
    </span>
  );
}

function KoladaSource({ munCode }: { munCode: string }) {
  return (
    <p className="text-[10px] text-on-surface-variant mt-2">
      Källa: Kolada · hämtas live ·{" "}
      <a
        href={`https://www.kolada.se/verktyg/fri-sokning/?kpis=&municipalityId=${munCode}`}
        target="_blank"
        rel="noopener noreferrer"
        className="underline hover:text-on-surface"
      >
        verifiera på kolada.se
      </a>
      {" · osäkra värden markeras med "}
      <span className="text-amber-500 font-bold">*</span>
    </p>
  );
}

function KPISkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="space-y-2">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex justify-between items-center py-2 border-b animate-pulse" style={{ borderColor: "var(--color-surface-high)" }}>
          <div className="h-3 rounded bg-surface-high w-48" />
          <div className="h-3 rounded bg-surface-high w-20" />
        </div>
      ))}
    </div>
  );
}

function KPITable({ kpis, kpiOrder }: { kpis: MunicipalityKPIItem[]; kpiOrder: string[] }) {
  // Build sorted-by-year map per KPI code
  const byCode = new Map<string, MunicipalityKPIItem[]>();
  for (const item of kpis) {
    if (!byCode.has(item.kpi)) byCode.set(item.kpi, []);
    byCode.get(item.kpi)!.push(item);
  }
  for (const arr of byCode.values()) arr.sort((a, b) => a.year - b.year);

  return (
    <div className="rounded-xl overflow-hidden border" style={{ borderColor: "var(--color-surface-high)" }}>
      <table className="w-full text-sm">
        <thead>
          <tr className="text-[10px] uppercase tracking-widest text-on-surface-variant" style={{ background: "var(--color-surface-low)" }}>
            <th className="px-4 py-2.5 text-left font-semibold">Nyckeltal</th>
            <th className="px-4 py-2.5 text-right font-semibold">Senaste värde</th>
            <th className="px-4 py-2.5 text-right font-semibold">Förändring</th>
            <th className="px-4 py-2.5 text-right font-semibold pr-4">År</th>
          </tr>
        </thead>
        <tbody>
          {kpiOrder.map((code, i) => {
            const years = byCode.get(code) ?? [];
            const item = years[years.length - 1];
            const prev = years.length >= 2 ? years[years.length - 2] : null;
            const label = KPI_LABELS[code] ?? code;
            const unit = KPI_UNITS[code] ?? "";
            const hasValue = item && item.status !== "M";
            const hasPrev = prev && prev.status !== "M";
            const signsCross = !!(hasValue && hasPrev &&
              ((item!.value > 0 && prev!.value < 0) || (item!.value < 0 && prev!.value > 0)));
            const rawDeltaPct = hasValue && hasPrev && prev!.value !== 0 && !signsCross
              ? ((item!.value - prev!.value) / Math.abs(prev!.value)) * 100
              : null;
            const useAbsDelta = signsCross || (rawDeltaPct != null && Math.abs(rawDeltaPct) > 100);
            const deltaPct = useAbsDelta ? null : rawDeltaPct;
            const deltaAbs = useAbsDelta && hasValue && hasPrev ? item!.value - prev!.value : null;
            return (
              <tr
                key={code}
                className="border-t"
                style={{
                  borderColor: "var(--color-surface-high)",
                  background: i % 2 === 0 ? "var(--color-surface-lowest)" : "var(--color-surface-low)",
                }}
              >
                <td className="px-4 py-2.5 text-on-surface-variant text-xs">{label}</td>
                <td className="px-4 py-2.5 text-right font-mono text-on-surface font-semibold text-xs">
                  {hasValue ? (
                    <>
                      {item!.value.toLocaleString("sv-SE", { maximumFractionDigits: 1 })} {unit}
                      <UncertainMark status={item!.status} />
                    </>
                  ) : (
                    <span className="text-on-surface-variant">–</span>
                  )}
                </td>
                <td className="px-4 py-2.5 text-right font-mono text-xs">
                  {deltaPct != null ? (
                    <span className="text-on-surface-variant">
                      {deltaPct > 0 ? "↑" : "↓"}{Math.abs(deltaPct).toFixed(1)}%
                    </span>
                  ) : deltaAbs != null ? (
                    <span className="text-on-surface-variant">
                      {deltaAbs > 0 ? "+" : ""}{deltaAbs.toLocaleString("sv-SE", { maximumFractionDigits: 2 })} {unit === "%" ? "pp" : unit}
                    </span>
                  ) : <span className="text-on-surface-variant">–</span>}
                  {hasPrev && (
                    <span className="text-on-surface-variant text-[10px] ml-1">
                      vs {prev!.year}
                    </span>
                  )}
                </td>
                <td className="px-4 py-2.5 text-right pr-4 text-xs text-on-surface-variant font-mono">
                  {item ? item.year : ""}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

const SPENDING_LABELS: Record<string, string> = {
  N11004: "Förskola",
  N15028: "Grundskola",
  N17014: "Gymnasieskola",
  N20014: "Äldreomsorg",
  N30005: "Individ- & familjeomsorg",
  N07037: "Infrastruktur & skydd",
  N09022: "Kultur & fritid",
  N05011: "Nämnd & administration",
};
const SPENDING_ORDER = ["N20014", "N15028", "N30005", "N11004", "N17014", "N07037", "N09022", "N05011"];
const SPENDING_COLORS: Record<string, string> = {
  N11004: "#6366f1",
  N15028: "#0ea5e9",
  N17014: "#14b8a6",
  N20014: "#f59e0b",
  N30005: "#ef4444",
  N07037: "#8b5cf6",
  N09022: "#10b981",
  N05011: "#94a3b8",
};
const SPENDING_YEARS = [2019, 2020, 2021, 2022, 2023];

function donutSlicePath(cx: number, cy: number, R: number, r: number, startDeg: number, endDeg: number): string {
  const toRad = (d: number) => (d - 90) * (Math.PI / 180);
  const [sx, sy] = [cx + R * Math.cos(toRad(startDeg)), cy + R * Math.sin(toRad(startDeg))];
  const [ex, ey] = [cx + R * Math.cos(toRad(endDeg)), cy + R * Math.sin(toRad(endDeg))];
  const [iex, iey] = [cx + r * Math.cos(toRad(endDeg)), cy + r * Math.sin(toRad(endDeg))];
  const [isx, isy] = [cx + r * Math.cos(toRad(startDeg)), cy + r * Math.sin(toRad(startDeg))];
  const large = endDeg - startDeg > 180 ? 1 : 0;
  return `M ${sx} ${sy} A ${R} ${R} 0 ${large} 1 ${ex} ${ey} L ${iex} ${iey} A ${r} ${r} 0 ${large} 0 ${isx} ${isy} Z`;
}

function SpendingChart({ items, year }: { items: MunicipalityKPIItem[]; year: number }) {
  const byCode: Record<string, { value: number; status?: string }> = {};
  for (const item of items) {
    if (item.year === year && item.status !== "M" && item.value > 0) {
      byCode[item.kpi] = { value: item.value, status: item.status };
    }
  }
  const total = SPENDING_ORDER.reduce((s, c) => s + (byCode[c]?.value ?? 0), 0);
  if (total === 0) return <p className="text-sm text-on-surface-variant py-2">Data saknas för {year}.</p>;

  const cx = 100, cy = 100, R = 88, r = 58;
  const GAP = 1.5;
  let cursor = 0;
  const slices = SPENDING_ORDER.map((code) => {
    const entry = byCode[code];
    const val = entry?.value ?? 0;
    const pct = val / total;
    const deg = pct * 360;
    const start = cursor + GAP / 2;
    const end = cursor + deg - GAP / 2;
    cursor += deg;
    return { code, val, pct, start, end, status: entry?.status };
  }).filter((s) => s.val > 0);

  return (
    <div className="flex gap-6 items-center">
      <svg viewBox="0 0 200 200" className="shrink-0" style={{ width: 180, height: 180 }}>
        {slices.map((s) => (
          <path
            key={s.code}
            d={donutSlicePath(cx, cy, R, r, s.start, s.end)}
            fill={SPENDING_COLORS[s.code]}
          />
        ))}
        <text x={cx} y={cy - 7} textAnchor="middle" fontSize="11" fill="currentColor" opacity="0.5">totalt</text>
        <text x={cx} y={cy + 8} textAnchor="middle" fontSize="13" fontWeight="700" fill="currentColor">
          {(total / 1000).toFixed(0)}k
        </text>
        <text x={cx} y={cy + 22} textAnchor="middle" fontSize="9" fill="currentColor" opacity="0.4">kr/inv</text>
      </svg>
      <div className="flex flex-col gap-1.5 flex-1 min-w-0">
        {slices.map((s) => (
          <div key={s.code} className="flex items-center gap-2">
            <div className="w-2.5 h-2.5 rounded-sm shrink-0" style={{ background: SPENDING_COLORS[s.code] }} />
            <span className="text-xs text-on-surface-variant truncate flex-1">
              {SPENDING_LABELS[s.code]}
              <UncertainMark status={s.status} />
            </span>
            <span className="text-xs font-mono text-on-surface shrink-0">{s.val.toLocaleString("sv-SE", { maximumFractionDigits: 0 })}</span>
            <span className="text-[11px] text-on-surface-variant shrink-0 w-8 text-right">{(s.pct * 100).toFixed(0)}%</span>
          </div>
        ))}
        <p className="text-[10px] text-on-surface-variant pt-1 border-t" style={{ borderColor: "var(--color-surface-high)" }}>
          kr/inv · {year}
        </p>
      </div>
    </div>
  );
}

function PopulationChart({ entries }: { entries: { year: number; population: number }[] }) {
  if (!entries.length) return null;
  const sorted = [...entries].sort((a, b) => a.year - b.year);
  const W = 600, H = 100, padX = 10, padY = 12, bottom = 20;
  const plotH = H - padY - bottom;
  const min = Math.min(...sorted.map((e) => e.population));
  const max = Math.max(...sorted.map((e) => e.population));
  const range = max - min || 1;
  const xs = sorted.map((_, i) => padX + (i / (sorted.length - 1 || 1)) * (W - padX * 2));
  const ys = sorted.map((e) => padY + plotH - ((e.population - min) / range) * plotH);
  const points = xs.map((x, i) => `${x},${ys[i]}`).join(" ");

  return (
    <svg viewBox={`0 0 ${W} ${H}`} className="w-full" style={{ overflow: "visible" }}>
      <polyline
        points={points}
        fill="none"
        stroke="var(--color-primary)"
        strokeWidth="2"
        strokeLinejoin="round"
        strokeLinecap="round"
      />
      {sorted.map((e, i) => (
        <g key={e.year}>
          <circle cx={xs[i]} cy={ys[i]} r="3.5" fill="var(--color-primary)" />
          <text
            x={xs[i]}
            y={ys[i] - 7}
            textAnchor="middle"
            fontSize="9"
            fill="currentColor"
            opacity="0.6"
          >
            {e.population >= 1000 ? `${(e.population / 1000).toFixed(1)}k` : e.population}
          </text>
          <text
            x={xs[i]}
            y={H - 4}
            textAnchor="middle"
            fontSize="9"
            fill="currentColor"
            opacity="0.5"
          >
            {e.year}
          </text>
        </g>
      ))}
    </svg>
  );
}

export function MunicipalityDetailPage() {
  const { code } = useParams<{ code: string }>();

  const { data: mun, isLoading, error } = useQuery({
    queryKey: ["municipality", code],
    queryFn: () => municipalitiesApi.getMunicipality(code!),
    enabled: !!code,
    staleTime: 60 * 60 * 1000,
  });

  const { data: kpis, isLoading: kpisLoading } = useQuery({
    queryKey: ["municipality-kpi", code],
    queryFn: () => municipalitiesApi.getMunicipalityKPIs(code!),
    enabled: !!code,
    staleTime: 10 * 60 * 1000,
  });

  const { data: popTrend, isLoading: popLoading } = useQuery({
    queryKey: ["municipality-pop-trend", code],
    queryFn: () => municipalitiesApi.getPopulationTrend(code!),
    enabled: !!code,
    staleTime: 10 * 60 * 1000,
  });

  const { data: spending, isLoading: spendingLoading } = useQuery({
    queryKey: ["municipality-spending", code],
    queryFn: () => municipalitiesApi.getMunicipalitySpending(code!),
    enabled: !!code,
    staleTime: 10 * 60 * 1000,
  });

  const availableSpendingYears = spending
    ? [...new Set(spending.map((s) => s.year))].sort((a, b) => b - a)
    : [];
  const [selectedSpendingYear, setSelectedSpendingYear] = useState<number | null>(null);
  const activeSpendingYear = selectedSpendingYear ?? availableSpendingYears[0] ?? 2023;

  if (isLoading) {
    return <p className="text-sm text-on-surface-variant py-16 text-center">Laddar kommundata…</p>;
  }
  if (error || !mun) {
    return (
      <div className="py-8 text-center">
        <p className="text-sm text-red-500">Kommunen hittades inte.</p>
        <NavLink to="/kommun" className="text-sm text-primary hover:underline mt-2 inline-block">
          ← Alla kommuner
        </NavLink>
      </div>
    );
  }

  const electionResults = mun.electionResults ?? [];
  const governingParties = mun.governingParties ?? [];
  const totalVotePct = electionResults.reduce((s, r) => s + r.votePct, 0);

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 text-xs text-on-surface-variant">
        <NavLink to="/kommun" className="hover:text-on-surface transition-colors">
          ← Kommuner
        </NavLink>
        <span>/</span>
        <NavLink
          to={`/region/${mun.regionCode}`}
          className="hover:text-on-surface transition-colors"
        >
          {mun.regionName}
        </NavLink>
      </div>

      {/* Header */}
      <div className="rounded-xl p-5 border" style={{ background: "var(--color-surface-lowest)", borderColor: "var(--color-surface-high)" }}>
        <div className="flex items-start justify-between mb-3">
          <div>
            <h2 className="font-display text-2xl font-extrabold tracking-tight text-on-surface">
              {mun.name}
            </h2>
            <p className="text-sm text-on-surface-variant">{mun.regionName}</p>
          </div>
          <span className="text-xs font-mono text-on-surface-variant bg-surface-high px-2 py-1 rounded">
            {mun.code}
          </span>
        </div>

        <div className="flex gap-6 pt-3 border-t" style={{ borderColor: "var(--color-surface-high)" }}>
          <StatBlock value={mun.population.toLocaleString("sv-SE")} label="Invånare" small />
          <StatBlock value={mun.totalMandates} label="Mandat" small />
          <StatBlock value={mun.electionYear} label="Senaste val" small />
          <StatBlock value={governingParties.join("+")} label="Styre" small />
        </div>
      </div>

      {/* Governing coalition */}
      {governingParties.length > 0 && (
        <div>
          <p className="text-[10px] uppercase tracking-widest font-semibold text-on-surface-variant mb-2">
            Styrande koalition
          </p>
          <div className="flex gap-2 flex-wrap">
            {governingParties.map((p) => (
              <PartyBadge key={p} party={p} size="lg" />
            ))}
          </div>
        </div>
      )}

      {/* Election results */}
      {electionResults.length > 0 ? (
        <>
          <div>
            <p className="text-[10px] uppercase tracking-widest font-semibold text-on-surface-variant mb-3">
              Kommunalval {mun.electionYear} — mandatfördelning
            </p>
            <div className="rounded-xl p-4 border" style={{ background: "var(--color-surface-lowest)", borderColor: "var(--color-surface-high)" }}>
              <MandateBar results={electionResults} />
            </div>
          </div>

          <div>
            <p className="text-[10px] uppercase tracking-widest font-semibold text-on-surface-variant mb-2">
              Partiresultat
            </p>
            <div className="rounded-xl overflow-hidden border" style={{ borderColor: "var(--color-surface-high)" }}>
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-[10px] uppercase tracking-widest text-on-surface-variant" style={{ background: "var(--color-surface-low)" }}>
                    <th className="px-4 py-2.5 text-left font-semibold">Parti</th>
                    <th className="px-4 py-2.5 text-right font-semibold">Röster %</th>
                    <th className="px-4 py-2.5 text-right font-semibold">Mandat</th>
                    <th className="px-4 py-2.5 text-right font-semibold pr-4">Andel</th>
                  </tr>
                </thead>
                <tbody>
                  {electionResults.map((r, i) => {
                    const color = PARTY_COLORS[r.party]?.bg ?? "#9ca3af";
                    const isGov = governingParties.includes(r.party);
                    const mandatePct = ((r.mandates / r.totalMandates) * 100).toFixed(1);
                    return (
                      <tr
                        key={r.party}
                        className="border-t"
                        style={{
                          borderColor: "var(--color-surface-high)",
                          background: i % 2 === 0 ? "var(--color-surface-lowest)" : "var(--color-surface-low)",
                        }}
                      >
                        <td className="px-4 py-2.5">
                          <div className="flex items-center gap-2">
                            <span className="w-2.5 h-2.5 rounded-sm shrink-0" style={{ background: color }} />
                            <span className="font-semibold text-on-surface">{r.party}</span>
                            {isGov && (
                              <span className="text-[9px] px-1 py-px rounded font-bold" style={{ background: color + "22", color }}>
                                STYRE
                              </span>
                            )}
                          </div>
                        </td>
                        <td className="px-4 py-2.5 text-right font-mono text-on-surface">{r.votePct.toFixed(1)}%</td>
                        <td className="px-4 py-2.5 text-right font-mono text-on-surface font-bold">{r.mandates}</td>
                        <td className="px-4 py-2.5 text-right pr-4">
                          <div className="flex items-center justify-end gap-2">
                            <div className="h-1.5 rounded-full" style={{ width: `${Math.max(2, parseFloat(mandatePct))}px`, background: color }} />
                            <span className="font-mono text-xs text-on-surface-variant">{mandatePct}%</span>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
                <tfoot>
                  <tr className="border-t text-[11px] font-semibold text-on-surface-variant" style={{ borderColor: "var(--color-surface-high)", background: "var(--color-surface-low)" }}>
                    <td className="px-4 py-2">Totalt</td>
                    <td className="px-4 py-2 text-right font-mono">{totalVotePct.toFixed(1)}%</td>
                    <td className="px-4 py-2 text-right font-mono">{mun.totalMandates}</td>
                    <td className="px-4 py-2 text-right pr-4 font-mono">100%</td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </div>
        </>
      ) : (
        <p className="text-sm text-on-surface-variant py-4">
          Valdata saknas för denna kommun i nuläget.
        </p>
      )}

      {/* Var går pengarna? */}
      <div>
        <p className="text-[10px] uppercase tracking-widest font-semibold text-on-surface-variant mb-3">
          Var går pengarna? (Kolada)
        </p>
        {spendingLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="flex items-center gap-3 animate-pulse">
                <div className="h-3 rounded bg-surface-high w-44 shrink-0" />
                <div className="flex-1 h-5 rounded bg-surface-high" />
                <div className="h-3 rounded bg-surface-high w-24 shrink-0" />
              </div>
            ))}
          </div>
        ) : spending && spending.length > 0 ? (
          <div className="rounded-xl p-4 border space-y-3" style={{ background: "var(--color-surface-lowest)", borderColor: "var(--color-surface-high)" }}>
            <div className="flex gap-1 flex-wrap">
              {SPENDING_YEARS.map((y) => (
                <button
                  key={y}
                  onClick={() => setSelectedSpendingYear(y)}
                  className="text-[11px] font-mono px-2.5 py-1 rounded transition-colors"
                  style={{
                    background: activeSpendingYear === y ? "var(--color-primary)" : "var(--color-surface-high)",
                    color: activeSpendingYear === y ? "var(--color-on-primary, #fff)" : "var(--color-on-surface-variant)",
                  }}
                >
                  {y}
                </button>
              ))}
            </div>
            <SpendingChart items={spending} year={activeSpendingYear} />
            <KoladaSource munCode={mun.code} />
          </div>
        ) : (
          <p className="text-sm text-on-surface-variant py-2">Data ej tillgänglig.</p>
        )}
      </div>

      {/* Verksamhetsnyckeltal */}
      <div>
        <p className="text-[10px] uppercase tracking-widest font-semibold text-on-surface-variant mb-2">
          Verksamhetsnyckeltal (Kolada)
        </p>
        {kpisLoading ? (
          <KPISkeleton rows={5} />
        ) : kpis && kpis.length > 0 ? (
          <>
            <KPITable kpis={kpis} kpiOrder={VERKSAMHET_KPI_ORDER} />
            <KoladaSource munCode={mun.code} />
          </>
        ) : (
          <p className="text-sm text-on-surface-variant py-2">Data ej tillgänglig.</p>
        )}
      </div>

      {/* Budget & Ekonomi */}
      <div>
        <p className="text-[10px] uppercase tracking-widest font-semibold text-on-surface-variant mb-2">
          Budget & Ekonomi (Kolada)
        </p>
        {kpisLoading ? (
          <KPISkeleton rows={5} />
        ) : kpis && kpis.length > 0 ? (
          <>
            <KPITable kpis={kpis} kpiOrder={BUDGET_KPI_ORDER} />
            <KoladaSource munCode={mun.code} />
          </>
        ) : (
          <p className="text-sm text-on-surface-variant py-2">Data ej tillgänglig.</p>
        )}
      </div>

      {/* Population trend */}
      <div>
        <p className="text-[10px] uppercase tracking-widest font-semibold text-on-surface-variant mb-3">
          Befolkningsutveckling (SCB)
        </p>
        {popLoading ? (
          <div className="h-24 rounded animate-pulse bg-surface-high" />
        ) : popTrend && popTrend.length > 0 ? (
          <div className="rounded-xl p-4 border" style={{ background: "var(--color-surface-lowest)", borderColor: "var(--color-surface-high)" }}>
            <PopulationChart entries={popTrend} />
            <p className="text-[10px] text-on-surface-variant mt-2">
              Källa: SCB BefolkningNy ·{" "}
              <a
                href="https://www.statistikdatabasen.scb.se/pxweb/sv/ssd/START__BE__BE0101__BE0101A/BefolkningNy/"
                target="_blank"
                rel="noopener noreferrer"
                className="underline hover:text-on-surface"
              >
                api.scb.se
              </a>
            </p>
          </div>
        ) : (
          <p className="text-sm text-on-surface-variant py-2">Data ej tillgänglig.</p>
        )}
      </div>

      {/* Offentligabeslut section */}
      <div className="rounded-xl p-5 border" style={{ background: "var(--color-surface-lowest)", borderColor: "var(--color-surface-high)" }}>
        <p className="text-[10px] uppercase tracking-widest font-semibold text-on-surface-variant mb-2">
          Kommunfullmäktiges beslut
        </p>
        <p className="text-sm text-on-surface-variant mb-4 leading-relaxed">
          Offentligabeslut.se samlar protokoll från kommunfullmäktige sedan 2018.
        </p>
        <a
          href={`https://offentligabeslut.se/?q=${encodeURIComponent(mun.name)}`}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-2 text-sm font-medium px-4 py-2 rounded-lg bg-surface-low text-on-surface hover:bg-surface transition-colors"
        >
          Se beslut för {mun.name} →
        </a>
      </div>
    </div>
  );
}
