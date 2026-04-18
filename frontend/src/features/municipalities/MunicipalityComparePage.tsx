import { useState, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { municipalitiesApi } from "./api";
import type { MunicipalityKPIItem } from "@/shared/types";

const ALL_KPI_LABELS: Record<string, string> = {
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
const ALL_KPI_UNITS: Record<string, string> = {
  N00900: "%", N11037: "kr", N15027: "kr", N20043: "kr/inv",
  N03010: "kr/inv", N03007: "kr/inv", N03102: "%", N03106: "%",
  N03040: "kr/inv", N03132: "kr/inv",
};
const KPI_ORDER = ["N00900", "N11037", "N15027", "N20043", "N03010", "N03007", "N03102", "N03106", "N03040", "N03132"];

function latestByKPI(items: MunicipalityKPIItem[]): Map<string, MunicipalityKPIItem> {
  const map = new Map<string, MunicipalityKPIItem>();
  for (const item of items) {
    const existing = map.get(item.kpi);
    if (!existing || item.year > existing.year) map.set(item.kpi, item);
  }
  return map;
}

function diffColor(pct: number): string {
  const abs = Math.abs(pct);
  if (abs >= 50) return "rgba(239,68,68,0.12)";
  if (abs >= 25) return "rgba(245,158,11,0.12)";
  return "transparent";
}

function diffBadgeStyle(pct: number): React.CSSProperties {
  const abs = Math.abs(pct);
  if (abs >= 50) return { color: "#ef4444", fontWeight: 700 };
  if (abs >= 25) return { color: "#d97706", fontWeight: 600 };
  return { color: "var(--color-on-surface-variant)" };
}

export function MunicipalityComparePage() {
  const [codeA, setCodeA] = useState("");
  const [codeB, setCodeB] = useState("");

  const { data: allMuns } = useQuery({
    queryKey: ["municipalities-all"],
    queryFn: () => municipalitiesApi.listMunicipalities(),
    staleTime: 60 * 60 * 1000,
  });

  const sorted = useMemo(
    () => [...(allMuns ?? [])].sort((a, b) => a.name.localeCompare(b.name, "sv")),
    [allMuns]
  );

  const { data: kpisA, isLoading: loadingA } = useQuery({
    queryKey: ["municipality-kpi", codeA],
    queryFn: () => municipalitiesApi.getMunicipalityKPIs(codeA),
    enabled: !!codeA,
    staleTime: 10 * 60 * 1000,
  });

  const { data: kpisB, isLoading: loadingB } = useQuery({
    queryKey: ["municipality-kpi", codeB],
    queryFn: () => municipalitiesApi.getMunicipalityKPIs(codeB),
    enabled: !!codeB,
    staleTime: 10 * 60 * 1000,
  });

  const munA = sorted.find((m) => m.code === codeA);
  const munB = sorted.find((m) => m.code === codeB);
  const latestA = kpisA ? latestByKPI(kpisA) : null;
  const latestB = kpisB ? latestByKPI(kpisB) : null;
  const ready = !!codeA && !!codeB && latestA && latestB && !loadingA && !loadingB;

  return (
    <div className="space-y-6">
      <div>
        <h2 className="font-display text-xl font-extrabold tracking-tight text-on-surface mb-1">
          Jämför kommuner
        </h2>
        <p className="text-sm text-on-surface-variant">
          Välj två kommuner för att jämföra nyckeltal. Rader markerade i{" "}
          <span className="text-amber-500 font-semibold">gult</span> skiljer sig med {">"} 25 % och{" "}
          <span className="text-red-500 font-semibold">rött</span> med {">"} 50 %.
        </p>
      </div>

      {/* Selectors */}
      <div className="grid grid-cols-2 gap-4">
        {[
          { label: "Kommun A", value: codeA, onChange: setCodeA },
          { label: "Kommun B", value: codeB, onChange: setCodeB },
        ].map(({ label, value, onChange }) => (
          <div key={label}>
            <p className="text-[10px] uppercase tracking-widest font-semibold text-on-surface-variant mb-1.5">
              {label}
            </p>
            <select
              value={value}
              onChange={(e) => onChange(e.target.value)}
              className="w-full rounded-lg px-3 py-2 text-sm border"
              style={{
                background: "var(--color-surface-low)",
                borderColor: "var(--color-surface-high)",
                color: "var(--color-on-surface)",
              }}
            >
              <option value="">Välj kommun…</option>
              {sorted.map((m) => (
                <option key={m.code} value={m.code}>
                  {m.name}
                </option>
              ))}
            </select>
          </div>
        ))}
      </div>

      {/* Loading */}
      {(loadingA || loadingB) && codeA && codeB && (
        <p className="text-sm text-on-surface-variant animate-pulse">Hämtar nyckeltal…</p>
      )}

      {/* Comparison table */}
      {ready && (
        <div className="rounded-xl overflow-hidden border" style={{ borderColor: "var(--color-surface-high)" }}>
          <table className="w-full text-sm">
            <thead>
              <tr
                className="text-[10px] uppercase tracking-widest text-on-surface-variant"
                style={{ background: "var(--color-surface-low)" }}
              >
                <th className="px-4 py-2.5 text-left font-semibold">Nyckeltal</th>
                <th className="px-4 py-2.5 text-right font-semibold">{munA?.name ?? codeA}</th>
                <th className="px-4 py-2.5 text-right font-semibold">{munB?.name ?? codeB}</th>
                <th className="px-4 py-2.5 text-right font-semibold pr-4">Skillnad</th>
              </tr>
            </thead>
            <tbody>
              {KPI_ORDER.map((code, i) => {
                const a = latestA!.get(code);
                const b = latestB!.get(code);
                const unit = ALL_KPI_UNITS[code] ?? "";
                const label = ALL_KPI_LABELS[code] ?? code;
                const hasA = a && a.status !== "M";
                const hasB = b && b.status !== "M";
                const diffPct =
                  hasA && hasB && a!.value !== 0
                    ? ((b!.value - a!.value) / Math.abs(a!.value)) * 100
                    : null;
                const rowBg = diffPct != null ? diffColor(diffPct) : "transparent";
                return (
                  <tr
                    key={code}
                    className="border-t"
                    style={{
                      borderColor: "var(--color-surface-high)",
                      background:
                        rowBg !== "transparent"
                          ? rowBg
                          : i % 2 === 0
                          ? "var(--color-surface-lowest)"
                          : "var(--color-surface-low)",
                    }}
                  >
                    <td className="px-4 py-2.5 text-on-surface-variant text-xs">{label}</td>
                    <td className="px-4 py-2.5 text-right font-mono text-on-surface text-xs">
                      {hasA
                        ? `${a!.value.toLocaleString("sv-SE", { maximumFractionDigits: 1 })} ${unit}`
                        : <span className="text-on-surface-variant">–</span>}
                    </td>
                    <td className="px-4 py-2.5 text-right font-mono text-on-surface text-xs">
                      {hasB
                        ? `${b!.value.toLocaleString("sv-SE", { maximumFractionDigits: 1 })} ${unit}`
                        : <span className="text-on-surface-variant">–</span>}
                    </td>
                    <td className="px-4 py-2.5 text-right pr-4 font-mono text-xs">
                      {diffPct != null ? (
                        <span style={diffBadgeStyle(diffPct)}>
                          {diffPct >= 0 ? "+" : ""}
                          {diffPct.toFixed(1)}%
                        </span>
                      ) : (
                        <span className="text-on-surface-variant">–</span>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {!codeA || !codeB ? (
        <p className="text-sm text-on-surface-variant text-center py-8">
          Välj två kommuner ovan för att se jämförelsen.
        </p>
      ) : null}
    </div>
  );
}
