import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { partiesApi } from "./api";
import { PartyBadge } from "@/shared/components";
import { SourceMarker } from "@/components/sources/SourceMarker";

export function PartiesPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["parties"],
    queryFn: partiesApi.listParties,
  });

  const { data: metaList } = useQuery({
    queryKey: ["party-meta"],
    queryFn: partiesApi.listPartyMeta,
  });
  const metaByCode = Object.fromEntries((metaList ?? []).map((m) => [m.code ?? "", m]));

  if (isLoading) {
    return <div className="text-on-surface-variant py-16 text-center text-sm">Laddar...</div>;
  }
  if (error) {
    return <div className="text-contradiction py-16 text-center text-sm">Kunde inte ladda partier.</div>;
  }

  const parties = data ?? [];
  // Alphabetical. The alignment percentage that previously ordered this list
  // was retired, and ordering parties by any derived figure is a ranking —
  // an editorial act the site does not perform.
  const sorted = [...parties].sort((a, b) => a.party.localeCompare(b.party, "sv"));

  return (
    <div>
      <div className="mb-8">
        <h2 className="font-display text-2xl font-extrabold tracking-tight mb-1">
          Partimål & röstning
        </h2>
        <p className="text-sm text-on-surface-variant leading-relaxed max-w-lg">
          Varje partis uttalade mål, bredvid hur partiet faktiskt röstade i
          riksdagen. Välj ett parti för att se målen och omröstningarna.
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
          const topics = p.topicBreakdown ?? [];

          return (
            <Link
              key={p.party}
              to={`/parties/${p.party}`}
              style={{
                display: "flex",
                alignItems: "center",
                gap: 16,
                padding: "16px 20px",
                borderTop: i > 0 ? "1px solid var(--color-surface-high)" : undefined,
                textDecoration: "none",
                color: "inherit",
                transition: "background 0.15s",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = "var(--color-surface-high)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = "transparent";
              }}
            >
              <PartyBadge party={p.party} size="lg" />

              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 4 }}>
                  <span style={{ fontSize: 14, fontWeight: 600, color: "var(--color-fg)" }}>
                    {p.party}
                  </span>
                  {metaByCode[p.party]?.ideology && (
                    <span style={{ fontSize: 10, color: "var(--color-fg-muted)", marginLeft: 8 }}>
                      {metaByCode[p.party].ideology}
                    </span>
                  )}
                  <span style={{ fontSize: 12, color: "var(--color-fg-muted)" }}>
                    {p.totalGoals} mål
                    <SourceMarker sourceId="seed-party-goals" />
                  </span>
                </div>

                {/* Topic tags */}
                {topics.length > 0 && (
                  <div style={{ display: "flex", flexWrap: "wrap", gap: 4 }}>
                    {topics.slice(0, 4).map((t) => (
                      <span
                        key={t.topic}
                        style={{
                          fontSize: 10,
                          color: "var(--color-fg-muted)",
                          textTransform: "capitalize",
                          padding: "2px 6px",
                          borderRadius: 3,
                          background: "var(--color-surface-low)",
                        }}
                      >
                        {t.topic}
                      </span>
                    ))}
                    {topics.length > 4 && (
                      <span style={{ fontSize: 10, color: "var(--color-fg-muted)", padding: "2px 6px" }}>
                        +{topics.length - 4}
                      </span>
                    )}
                  </div>
                )}
              </div>

            </Link>
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
