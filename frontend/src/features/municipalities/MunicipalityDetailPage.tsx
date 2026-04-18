import { useParams, NavLink } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { PartyBadge, StatBlock } from "@/shared/components";
import { PARTY_COLORS } from "@/shared/design";
import { municipalitiesApi } from "./api";
import type { ElectionResult } from "@/shared/types";

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

export function MunicipalityDetailPage() {
  const { code } = useParams<{ code: string }>();

  const { data: mun, isLoading, error } = useQuery({
    queryKey: ["municipality", code],
    queryFn: () => municipalitiesApi.getMunicipality(code!),
    enabled: !!code,
    staleTime: 60 * 60 * 1000,
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

  const totalVotePct = mun.electionResults.reduce((s, r) => s + r.votePct, 0);

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
          <StatBlock value={mun.governingParties.join("+")} label="Styre" small />
        </div>
      </div>

      {/* Governing coalition */}
      <div>
        <p className="text-[10px] uppercase tracking-widest font-semibold text-on-surface-variant mb-2">
          Styrande koalition
        </p>
        <div className="flex gap-2 flex-wrap">
          {mun.governingParties.map((p) => (
            <PartyBadge key={p} party={p} size="lg" />
          ))}
        </div>
      </div>

      {/* Election results */}
      {mun.electionResults.length > 0 ? (
        <>
          <div>
            <p className="text-[10px] uppercase tracking-widest font-semibold text-on-surface-variant mb-3">
              Kommunalval {mun.electionYear} — mandatfördelning
            </p>
            <div className="rounded-xl p-4 border" style={{ background: "var(--color-surface-lowest)", borderColor: "var(--color-surface-high)" }}>
              <MandateBar results={mun.electionResults} />
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
                  {mun.electionResults.map((r, i) => {
                    const color = PARTY_COLORS[r.party]?.bg ?? "#9ca3af";
                    const isGov = mun.governingParties.includes(r.party);
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
    </div>
  );
}
