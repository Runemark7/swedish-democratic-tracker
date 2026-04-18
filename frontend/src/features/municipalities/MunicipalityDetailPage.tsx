import { useParams, NavLink } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { PartyBadge, StatBlock } from "@/shared/components";
import { PARTY_COLORS } from "@/shared/design";
import { municipalitiesApi } from "./api";
import type { ElectionResult, MunicipalityKPIItem } from "@/shared/types";

const KPI_LABELS: Record<string, string> = {
  N00901: "Kommunalskatt",
  N11037: "Nettokostnad förskola per inskrivet barn",
  N15027: "Kostnad grundskola per elev",
  N20043: "Kostnad äldreomsorg per invånare",
  N03010: "Skatteintäkter per invånare",
};
const KPI_UNITS: Record<string, string> = {
  N00901: "%",
  N11037: "kr",
  N15027: "kr",
  N20043: "kr/inv",
  N03010: "kr/inv",
};
const KPI_ORDER = ["N00901", "N11037", "N15027", "N20043", "N03010"];

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

function KPISkeleton() {
  return (
    <div className="space-y-2">
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="flex justify-between items-center py-2 border-b animate-pulse" style={{ borderColor: "var(--color-surface-high)" }}>
          <div className="h-3 rounded bg-surface-high w-48" />
          <div className="h-3 rounded bg-surface-high w-20" />
        </div>
      ))}
    </div>
  );
}

function KPITable({ kpis }: { kpis: MunicipalityKPIItem[] }) {
  // Build map: kpiCode → latest year entry
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
          {KPI_ORDER.map((code, i) => {
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

      {/* Kolada KPI section */}
      <div>
        <p className="text-[10px] uppercase tracking-widest font-semibold text-on-surface-variant mb-2">
          Ekonomiska nyckeltal (Kolada)
        </p>
        {kpisLoading ? (
          <KPISkeleton />
        ) : kpis && kpis.length > 0 ? (
          <KPITable kpis={kpis} />
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
