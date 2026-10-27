import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { PartyBadge } from "@/shared/components";
import { partyShortToName } from "@/shared/design";
import { partiesApi } from "@/features/parties/api";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { DecisionFeed } from "./DecisionFeed";

export function HomePage() {
  const isMobile = useMediaQuery("(max-width: 640px)");

  // Read from the record rather than hardcoded. A literal roster keeps
  // asserting the same eight parties after the record changes — the same class
  // of defect as a hardcoded mandate period, which the committees fetcher
  // forbids one directory away.
  //
  // Alphabetical, not by mandate size: an editorial ordering is exactly what
  // the party index exists to avoid (#100: "8 badges, alphabetical, no ranking").
  const { data: parties } = useQuery({
    queryKey: ["parties"],
    queryFn: partiesApi.listParties,
    staleTime: 5 * 60_000,
  });
  const partyCodes = [...(parties ?? [])]
    .map((p) => p.party)
    .sort((a, b) => a.localeCompare(b, "sv"));

  const partyGridStyle: React.CSSProperties = {
    display: "grid",
    gridTemplateColumns: isMobile
      ? "minmax(0, 1fr) minmax(0, 1fr)"
      : "repeat(4, minmax(0, 1fr))",
    gap: 1,
    background: "var(--color-border)",
    border: "1px solid var(--color-border)",
  };

  return (
    <div className="sdt-page">
      {/* ── Hero ─────────────────────────────────────────────────────── */}
      <header
        style={{
          padding: isMobile ? "20px 14px 16px" : "40px 32px 24px",
          display: "flex",
          flexDirection: "column",
          gap: 8,
        }}
      >
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "2px",
            color: "var(--color-fg-muted)",
            textTransform: "uppercase",
          }}
        >
          RIKSDAGSKOLLEN · ÖPPEN DEMOKRATI
        </div>
        <h1
          style={{
            fontFamily: "var(--font-serif)",
            fontSize: isMobile ? 28 : 44,
            fontWeight: 400,
            letterSpacing: "-1.2px",
            color: "var(--color-fg)",
            margin: 0,
            lineHeight: 1.05,
          }}
        >
          Vad beslutar riksdagen — och håller de vad de lovar?
        </h1>
        <p
          style={{
            fontFamily: "var(--font-body)",
            fontSize: 13,
            color: "var(--color-fg-muted)",
            margin: 0,
          }}
        >
          Röster, löften och beslut — med källa, utan vinkel.{" "}
          <Link to="/om-sajten" style={{ color: "var(--color-accent)" }}>
            Om sajten
          </Link>
        </p>
      </header>

      <DecisionFeed />

      <section
        style={{
          margin: isMobile ? "20px 14px 6px" : "32px 32px 6px",
        }}
      >
        <header
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "0.15em",
            color: "var(--color-fg-muted)",
            textTransform: "uppercase",
            marginBottom: 10,
          }}
        >
          PARTIERNA
          <SourceMarker sourceId="seed-party-goals" />
        </header>
        <div style={partyGridStyle}>
          {partyCodes.map((p) => (
            <Link
              key={p}
              to={`/parties/${p}`}
              style={{
                background: "var(--color-sdt-surface)",
                padding: isMobile ? "14px" : "16px",
                display: "flex",
                alignItems: "center",
                gap: 10,
                textDecoration: "none",
                color: "inherit",
                minWidth: 0,
              }}
            >
              <PartyBadge party={p} />
              <span
                style={{
                  fontFamily: "var(--font-body)",
                  fontSize: 12,
                  color: "var(--color-fg-muted)",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                  minWidth: 0,
                }}
              >
                {partyShortToName(p)}
              </span>
            </Link>
          ))}
        </div>
      </section>

    </div>
  );
}
