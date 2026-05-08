import { useQuery } from "@tanstack/react-query";
import { partiesApi } from "./api";
import { PartyBadge } from "@/shared/components";
import { PARTY_COLORS, alignmentColor } from "@/shared/design";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { SectionSource } from "@/components/sources/SectionSource";

export function PartiesPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["parties"],
    queryFn: partiesApi.listParties,
  });

  if (isLoading) {
    return <div className="text-on-surface-variant py-16 text-center text-sm">Laddar...</div>;
  }
  if (error) {
    return <div className="text-contradiction py-16 text-center text-sm">Kunde inte ladda partier.</div>;
  }

  const parties = data ?? [];
  const sorted = [...parties].sort((a, b) => b.avgAlignmentPct - a.avgAlignmentPct);

  return (
    <div>
      <div className="mb-8">
        <h2 className="font-display text-2xl font-extrabold tracking-tight mb-1">
          Partimål & röstning
        </h2>
        <p className="text-sm text-on-surface-variant leading-relaxed max-w-lg">
          Hur väl matchar partiernas röstning i riksdagen deras egna uttalade mål?
          Välj ett parti i menyn ovan för att se detaljerna.
        </p>
      </div>

      {/* Explanation banner */}
      <div className="bg-[#f0f9ff] rounded-lg p-4 mb-6">
        <p className="text-xs text-[#0369a1] leading-relaxed m-0">
          Varje omröstning visar <strong>vilket parti som initierade förslaget</strong> — ett
          NEJ-röst kan bero på att förslaget kom från ett annat parti med annorlunda villkor,
          inte att partiet är emot sakfrågan.
        </p>
      </div>

      {/* Compact summary table */}
      <div className="bg-surface-lowest rounded-xl overflow-hidden">
        {sorted.map((p, i) => {
          const pc = PARTY_COLORS[p.party];
          const avg = Math.round(p.avgAlignmentPct);
          const topics = p.topicBreakdown ?? [];

          return (
            <div
              key={p.party}
              className="flex items-center gap-4 px-5 py-4"
              style={{ borderTop: i > 0 ? "1px solid var(--color-surface-high)" : undefined }}
            >
              <PartyBadge party={p.party} size="lg" />

              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-sm font-semibold text-on-surface">
                    {p.party}
                  </span>
                  <span className="text-xs text-on-surface-variant">
                    {p.totalGoals} mål
                    <SourceMarker sourceId="seed-party-goals" />
                  </span>
                </div>

                {/* Topic tags */}
                {topics.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {topics.slice(0, 4).map((t) => (
                      <span
                        key={t.topic}
                        className="text-[10px] text-on-surface-variant capitalize px-1.5 py-0.5 rounded bg-surface-low"
                      >
                        {t.topic}
                      </span>
                    ))}
                    {topics.length > 4 && (
                      <span className="text-[10px] text-on-surface-variant px-1.5 py-0.5">
                        +{topics.length - 4}
                      </span>
                    )}
                  </div>
                )}
              </div>

              {/* Alignment bar */}
              <div className="flex items-center gap-3 shrink-0 w-[140px]">
                <div className="flex-1 h-2 rounded-full bg-surface-high overflow-hidden">
                  <div
                    className="h-full rounded-full transition-all duration-700"
                    style={{
                      width: `${Math.min(avg, 100)}%`,
                      background: pc?.bg ?? "#666",
                    }}
                  />
                </div>
                <span
                  className="text-sm font-mono font-extrabold min-w-[36px] text-right"
                  style={{ color: alignmentColor(avg) }}
                >
                  {avg}%
                  <SourceMarker sourceId="riksdagen" />
                </span>
              </div>
              <SectionSource sourceIds={["seed-party-goals", "riksdagen"]} />
            </div>
          );
        })}
      </div>

      {parties.length === 0 && (
        <p className="text-on-surface-variant text-center py-16 text-sm">
          Inga partier laddade ännu. Vänta tills datainsamlingen är klar.
        </p>
      )}
    </div>
  );
}
