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
  const latest = new Map<string, MunicipalityKPIItem>();
  for (const item of kpis) {
    const existing = latest.get(item.kpi);
    if (!existing || item.year > existing.year) latest.set(item.kpi, item);
  }

  return (
    <div className="rounded-xl overflow-hidden border" style={{ borderColor: "var(--color-surface-high)" }}>
      <table className="w-full text-sm">
        <thead>
          <tr className="text-[10px] uppercase tracking-widest text-on-surface-variant" style={{ background: "var(--color-surface-low)" }}>
            <th className="px-4 py-2.5 text-left font-semibold">Nyckeltal</th>
            <th className="px-4 py-2.5 text-right font-semibold">Värde</th>
            <th className="px-4 py-2.5 text-right font-semibold pr-4">År</th>
          </tr>
        </thead>
        <tbody>
          {kpiOrder.map((code, i) => {
            const item = latest.get(code);
            const label = KPI_LABELS[code] ?? code;
            const unit = KPI_UNITS[code] ?? "";
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
                <td className="px-4 py-2.5 text-right font-mono text-on-surface font-semibold">
                  {item && item.status !== "M"
                    ? `${item.value.toLocaleString("sv-SE", { maximumFractionDigits: 1 })} ${unit}`
                    : <span className="text-on-surface-variant text-xs">–</span>}
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

function SpendingChart({ items, year }: { items: MunicipalityKPIItem[]; year: number }) {
  const byCode: Record<string, number> = {};
  for (const item of items) {
    if (item.year === year && item.value > 0) byCode[item.kpi] = item.value;
  }
  const total = SPENDING_ORDER.reduce((s, c) => s + (byCode[c] ?? 0), 0);
  if (total === 0) return <p className="text-sm text-on-surface-variant py-2">Data saknas för {year}.</p>;

  return (
    <div className="space-y-2">
      {SPENDING_ORDER.map((code) => {
        const val = byCode[code] ?? 0;
        const pct = total > 0 ? (val / total) * 100 : 0;
        const color = SPENDING_COLORS[code];
        const label = SPENDING_LABELS[code];
        return (
          <div key={code} className="flex items-center gap-3">
            <span className="text-xs text-on-surface-variant w-44 shrink-0 text-right">{label}</span>
            <div className="flex-1 h-5 rounded-sm overflow-hidden" style={{ background: "var(--color-surface-high)" }}>
              <div
                className="h-full rounded-sm transition-all duration-300"
                style={{ width: `${pct}%`, background: color }}
              />
            </div>
            <span className="text-xs font-mono text-on-surface w-28 shrink-0">
              {val > 0 ? `${val.toLocaleString("sv-SE", { maximumFractionDigits: 0 })} kr/inv` : "–"}
            </span>
            <span className="text-[11px] text-on-surface-variant w-10 shrink-0 text-right">
              {pct > 0 ? `${pct.toFixed(0)}%` : ""}
            </span>
          </div>
        );
      })}
      <p className="text-[10px] text-on-surface-variant pt-1">
        Summa redovisade sektorer: {total.toLocaleString("sv-SE", { maximumFractionDigits: 0 })} kr/inv
      </p>
    </div>
  );
}

function PopulationChart({ entries }: { entries: { year: number; population: number }[] }) {
  if (!entries.length) return null;
  const sorted = [...entries].sort((a, b) => a.year - b.year);
  const max = Math.max(...sorted.map((e) => e.population));
  return (
    <div className="flex items-end gap-2 h-24">
      {sorted.map((e) => {
        const pct = (e.population / max) * 100;
        return (
          <div key={e.year} className="flex flex-col items-center gap-1 flex-1">
            <span className="text-[10px] font-mono text-on-surface-variant">
              {e.population >= 1000
                ? `${(e.population / 1000).toFixed(0)}k`
                : e.population}
            </span>
            <div
              className="w-full rounded-t"
              style={{ height: `${Math.max(4, pct * 0.7)}px`, background: "var(--color-primary)" }}
            />
            <span className="text-[10px] text-on-surface-variant">{e.year}</span>
          </div>
        );
      })}
    </div>
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

      {/* Verksamhetsnyckeltal */}
      <div>
        <p className="text-[10px] uppercase tracking-widest font-semibold text-on-surface-variant mb-2">
          Verksamhetsnyckeltal (Kolada)
        </p>
        {kpisLoading ? (
          <KPISkeleton rows={5} />
        ) : kpis && kpis.length > 0 ? (
          <KPITable kpis={kpis} kpiOrder={VERKSAMHET_KPI_ORDER} />
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
          <KPITable kpis={kpis} kpiOrder={BUDGET_KPI_ORDER} />
        ) : (
          <p className="text-sm text-on-surface-variant py-2">Data ej tillgänglig.</p>
        )}
      </div>

      {/* Population trend section */}
      <div>
        <p className="text-[10px] uppercase tracking-widest font-semibold text-on-surface-variant mb-3">
          Befolkningsutveckling (SCB)
        </p>
        {popLoading ? (
          <div className="h-24 rounded animate-pulse bg-surface-high" />
        ) : popTrend && popTrend.length > 0 ? (
          <div className="rounded-xl p-4 border" style={{ background: "var(--color-surface-lowest)", borderColor: "var(--color-surface-high)" }}>
            <PopulationChart entries={popTrend} />
          </div>
        ) : (
          <p className="text-sm text-on-surface-variant py-2">Data ej tillgänglig.</p>
        )}
      </div>

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
