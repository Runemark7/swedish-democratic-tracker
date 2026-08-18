import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { budgetApi } from "./api";
import { TierNav } from "./components/TierNav";
import { DeltaIndicator } from "./components/DeltaIndicator";
import { DocumentLinks } from "./components/DocumentLinks";
import { formatBudgetAmount } from "@/shared/design";
import { SourceMarker } from "@/components/sources/SourceMarker";
import type { components } from "@/shared/api-contract";

type BudgetDocumentRef = components["schemas"]["BudgetDocumentRef"];

function budgetDocRefs(year: number): BudgetDocumentRef[] {
  const session = `${year - 1}/${String(year % 100).padStart(2, "0")}`;
  const offset = year - 2013;
  const code = offset >= 1 && offset <= 35
    ? "H" + "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"[offset]
    : "";
  return [
    {
      type: "proposition",
      title: `Prop. ${session}:1`,
      url: code
        ? `https://data.riksdagen.se/dokument/${code}031.html`
        : `https://data.riksdagen.se/dokumentlista/?doktyp=prop&rm=${encodeURIComponent(session)}&bet=1&utformat=html&sz=50`,
      description: `Budgetpropositionen för ${year}`,
    },
    {
      type: "committee_report",
      title: `FiU — ${session}`,
      url: `https://data.riksdagen.se/dokumentlista/?doktyp=bet&rm=${encodeURIComponent(session)}&org=FiU&utformat=html&sz=50`,
      description: `Finansutskottets betänkanden ${session}`,
    },
  ];
}

type SortMode = "default" | "increase" | "decrease";

export function BudgetPage() {
  const { data: years } = useQuery({
    queryKey: ["budget-years"],
    queryFn: budgetApi.listYears,
    staleTime: 5 * 60 * 1000,
  });

  const availableYears = (years ?? []).map((y) => y.year).sort((a, b) => b - a);
  const [compareYear, setCompareYear] = useState<number | null>(null);
  const [baseYear, setBaseYear] = useState<number | null>(null);
  const [sortMode, setSortMode] = useState<SortMode>("default");

  // Default: compare latest two years
  const cy = compareYear ?? availableYears[0] ?? 2025;
  const by = baseYear ?? availableYears[1] ?? 2024;

  const { data: comparison, isLoading, error } = useQuery({
    queryKey: ["budget-compare", by, cy],
    queryFn: () => budgetApi.compareYears(by, cy),
    enabled: !!by && !!cy,
    staleTime: 5 * 60 * 1000,
  });

  const rows = comparison?.rows ?? [];
  const sortedRows = [...rows].sort((a, b) => {
    if (sortMode === "increase") return b.deltaPct - a.deltaPct;
    if (sortMode === "decrease") return a.deltaPct - b.deltaPct;
    return a.area.sortOrder - b.area.sortOrder;
  });

  return (
    <div>
      <TierNav active="national" />

      <div className="mb-8">
        <h2 className="font-display text-2xl font-extrabold tracking-tight mb-1">
          Statsbudgeten
        </h2>
        <p className="text-sm text-on-surface-variant leading-relaxed max-w-lg">
          Hur fördelas statens pengar mellan olika utgiftsomr&aring;den?
          J&auml;mf&ouml;r &aring;r f&ouml;r &aring;r f&ouml;r att se vad som prioriteras.
        </p>
      </div>

      {/* Year selectors */}
      {availableYears.length > 0 && (
        <div className="flex items-start gap-4 mb-6 flex-wrap" style={{ minWidth: 0 }}>
          <div className="flex items-center gap-2" style={{ minWidth: 0, maxWidth: "100%" }}>
            <span className="text-xs text-on-surface-variant font-medium uppercase tracking-wider">
              Fr&aring;n
            </span>
            <div className="flex gap-1 overflow-x-auto" style={{ scrollbarWidth: "none", maxWidth: "100%" }}>
              {availableYears.map((y) => (
                <button
                  key={`base-${y}`}
                  onClick={() => setBaseYear(y)}
                  className="px-3 py-1.5 text-xs font-mono font-bold rounded-md transition-all shrink-0"
                  style={{
                    background: y === by ? "var(--color-primary)" : "var(--color-surface-low)",
                    color: y === by ? "var(--color-on-primary)" : "var(--color-on-surface)",
                  }}
                >
                  {y}
                </button>
              ))}
            </div>
          </div>
          <span className="text-on-surface-variant">&rarr;</span>
          <div className="flex items-center gap-2" style={{ minWidth: 0, maxWidth: "100%" }}>
            <span className="text-xs text-on-surface-variant font-medium uppercase tracking-wider">
              Till
            </span>
            <div className="flex gap-1 overflow-x-auto" style={{ scrollbarWidth: "none", maxWidth: "100%" }}>
              {availableYears.map((y) => (
                <button
                  key={`comp-${y}`}
                  onClick={() => setCompareYear(y)}
                  className="px-3 py-1.5 text-xs font-mono font-bold rounded-md transition-all shrink-0"
                  style={{
                    background: y === cy ? "var(--color-primary)" : "var(--color-surface-low)",
                    color: y === cy ? "var(--color-on-primary)" : "var(--color-on-surface)",
                  }}
                >
                  {y}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Sort controls */}
      <div className="flex gap-2 mb-4 overflow-x-auto" style={{ scrollbarWidth: "none" }}>
        {([
          { mode: "default" as SortMode, label: "UO-ordning" },
          { mode: "increase" as SortMode, label: "St\u00f6rst \u00f6kning" },
          { mode: "decrease" as SortMode, label: "St\u00f6rst minskning" },
        ]).map(({ mode, label }) => (
          <button
            key={mode}
            onClick={() => setSortMode(mode)}
            className="px-3 py-1 text-[11px] font-semibold rounded transition-all"
            style={{
              background: sortMode === mode ? "var(--color-surface-highest)" : "transparent",
              color: sortMode === mode ? "var(--color-on-surface)" : "var(--color-on-surface-variant)",
            }}
          >
            {label}
          </button>
        ))}
      </div>

      {isLoading && (
        <div className="text-on-surface-variant py-16 text-center text-sm">Laddar...</div>
      )}
      {error && (
        <div className="text-contradiction py-16 text-center text-sm">
          Kunde inte ladda budgetdata.
        </div>
      )}

      {comparison && (
        <>
          {/* Total summary */}
          <div className="bg-surface-lowest rounded-xl p-5 mb-6">
            <div className="flex items-center justify-between flex-wrap gap-4">
              <div>
                <div className="text-[10px] text-on-surface-variant uppercase tracking-widest mb-1">
                  Total statsbudget {cy}
                </div>
                <div className="font-mono text-2xl font-extrabold text-on-surface">
                  {formatBudgetAmount(comparison.compareTotalKsek)}
                  <SourceMarker sourceId="seed-budget-data" />
                </div>
              </div>
              <div className="text-right">
                <div className="text-[10px] text-on-surface-variant uppercase tracking-widest mb-1">
                  F&ouml;r&auml;ndring fr&aring;n {by}
                </div>
                <DeltaIndicator pct={comparison.totalDeltaPct} showBar={false} />
                <div className="text-[10px] text-on-surface-variant mt-0.5">
                  {comparison.totalDeltaKsek > 0 ? "+" : ""}
                  {formatBudgetAmount(comparison.totalDeltaKsek)}
                  <SourceMarker sourceId="seed-budget-data" />
                </div>
              </div>
            </div>
            <div className="mt-3 pt-3" style={{ borderTop: "1px solid var(--color-surface-high)" }}>
              <div className="text-[10px] text-on-surface-variant uppercase tracking-widest mb-1.5">
                K&auml;lldokument
              </div>
              <DocumentLinks documents={budgetDocRefs(cy)} compact />
            </div>
          </div>

          {/* Comparison table */}
          <div className="bg-surface-lowest rounded-xl overflow-hidden">
            {/* Header */}
            <div
              className="flex items-center gap-3 px-5 py-3 text-[10px] font-semibold uppercase tracking-widest text-on-surface-variant"
              style={{ background: "var(--color-surface-low)" }}
            >
              <div className="w-12">UO</div>
              <div className="flex-1">Omr&aring;de</div>
              <div className="w-24 text-right">{by}</div>
              <div className="w-24 text-right">{cy}</div>
              <div className="w-36 text-right">F&ouml;r&auml;ndring</div>
            </div>

            {/* Rows */}
            {sortedRows.map((row, i) => (
              <Link
                key={row.area.code}
                to={`/budget/areas/${row.area.code}`}
                className="flex items-center gap-3 px-5 py-3 transition-colors hover:bg-surface-low no-underline text-inherit"
                style={{
                  borderTop: i > 0 ? "1px solid var(--color-surface-high)" : undefined,
                }}
              >
                <div className="w-12 text-xs font-mono font-bold text-on-surface-variant">
                  {row.area.code}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium text-on-surface truncate">
                    {row.area.name}
                  </div>
                </div>
                <div className="w-24 text-right text-xs font-mono text-on-surface-variant">
                  {formatBudgetAmount(row.baseAmountKsek)}
                  <SourceMarker sourceId="seed-budget-data" />
                </div>
                <div className="w-24 text-right text-xs font-mono font-semibold text-on-surface">
                  {formatBudgetAmount(row.compareAmountKsek)}
                  <SourceMarker sourceId="seed-budget-data" />
                </div>
                <div className="w-36 flex justify-end">
                  <DeltaIndicator pct={row.deltaPct} />
                </div>
              </Link>
            ))}

            {/* Total row */}
            <div
              className="flex items-center gap-3 px-5 py-3 font-bold"
              style={{
                borderTop: "2px solid var(--color-surface-highest)",
                background: "var(--color-surface-low)",
              }}
            >
              <div className="w-12" />
              <div className="flex-1 text-sm text-on-surface">TOTALT</div>
              <div className="w-24 text-right text-xs font-mono text-on-surface-variant">
                {formatBudgetAmount(comparison.baseTotalKsek)}
                <SourceMarker sourceId="seed-budget-data" />
              </div>
              <div className="w-24 text-right text-xs font-mono text-on-surface">
                {formatBudgetAmount(comparison.compareTotalKsek)}
                <SourceMarker sourceId="seed-budget-data" />
              </div>
              <div className="w-36 flex justify-end">
                <DeltaIndicator pct={comparison.totalDeltaPct} />
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
