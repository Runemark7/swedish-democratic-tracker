import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { budgetApi } from "./api";
import { TierNav } from "./components/TierNav";
import { DeltaIndicator } from "./components/DeltaIndicator";
import { DocumentLinks } from "./components/DocumentLinks";
import { formatBudgetAmount } from "@/shared/design";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { SectionSource } from "@/components/sources/SectionSource";

export function AreaHistoryPage() {
  const { code } = useParams<{ code: string }>();

  const { data: timeSeries, isLoading, error } = useQuery({
    queryKey: ["budget-area", code],
    queryFn: () => budgetApi.getAreaTimeSeries(code!),
    enabled: !!code,
    staleTime: 5 * 60 * 1000,
  });

  if (isLoading) {
    return <div className="text-on-surface-variant py-16 text-center text-sm">Laddar...</div>;
  }
  if (error || !timeSeries) {
    return <div className="text-contradiction py-16 text-center text-sm">Kunde inte ladda data.</div>;
  }

  const entries = timeSeries.entries;

  return (
    <div>
      <TierNav active="national" />

      <Link
        to="/budget"
        className="text-xs text-on-surface-variant hover:text-on-surface transition-colors no-underline mb-4 inline-block"
      >
        &larr; Tillbaka till statsbudgeten
      </Link>

      <div className="mb-8">
        <div className="flex items-center gap-3 mb-1">
          <span className="px-2.5 py-0.5 rounded text-xs font-mono font-bold bg-surface-low text-on-surface-variant">
            {timeSeries.area.code}
          </span>
          <h2 className="font-display text-2xl font-extrabold tracking-tight m-0">
            {timeSeries.area.name}
          </h2>
        </div>
        {timeSeries.area.description && (
          <p className="text-sm text-on-surface-variant leading-relaxed max-w-lg mt-1">
            {timeSeries.area.description}
          </p>
        )}
      </div>

      <div className="bg-surface-lowest rounded-xl overflow-hidden">
        {/* Header */}
        <div
          className="flex items-center gap-3 px-5 py-3 text-[10px] font-semibold uppercase tracking-widest text-on-surface-variant"
          style={{ background: "var(--color-surface-low)" }}
        >
          <div className="w-16">&Aring;r</div>
          <div className="flex-1">Belopp</div>
          <div className="w-36 text-right">F&ouml;r&auml;ndring</div>
        </div>

        {entries.map((entry, i) => {
          const prev = i > 0 ? entries[i - 1] : null;
          const deltaPct = prev && prev.amountKsek > 0
            ? ((entry.amountKsek - prev.amountKsek) / prev.amountKsek) * 100
            : null;

          return (
            <div
              key={entry.year}
              style={{
                borderTop: i > 0 ? "1px solid var(--color-surface-high)" : undefined,
              }}
            >
              <div className="flex items-center gap-3 px-5 py-3">
                <div className="w-16 text-sm font-mono font-bold text-on-surface">
                  {entry.year}
                </div>
                <div className="flex-1">
                  <span className="text-sm font-mono font-semibold text-on-surface">
                    {formatBudgetAmount(entry.amountKsek)}
                    <SourceMarker sourceId="seed-budget-data" />
                  </span>
                </div>
                <div className="w-36 flex justify-end">
                  {deltaPct !== null ? (
                    <DeltaIndicator pct={deltaPct} />
                  ) : (
                    <span className="text-[10px] text-on-surface-variant">&mdash;</span>
                  )}
                </div>
              </div>
              {entry.documents && entry.documents.length > 0 && (
                <div
                  className="px-5 pb-3 pt-0"
                  style={{ paddingLeft: "calc(1.25rem + 4rem + 0.75rem)" }}
                >
                  <DocumentLinks documents={entry.documents} compact />
                  <SectionSource sourceIds={["seed-budget-data"]} />
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
