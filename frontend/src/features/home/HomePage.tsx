import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { PartyBadge } from "@/shared/components";
import { partyShortToName } from "@/shared/design";
import { partiesApi } from "@/features/parties/api";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { DecisionFeed } from "./DecisionFeed";
import { CommitteeIndex } from "./CommitteeIndex";

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
          Röster, löften och beslut — med källa, utan vinkel.
        </p>
      </header>

      {/* ── Mission banner ───────────────────────────────────────────── */}
      <div
        style={{
          margin: isMobile ? "0 14px 20px" : "0 32px 20px",
          borderLeft: "3px solid var(--color-accent)",
          background: "var(--color-sdt-surface)",
          padding: isMobile ? "16px 16px" : "20px 28px",
          display: "flex",
          flexDirection: "column",
          gap: 12,
        }}
      >
        <p
          style={{
            fontFamily: "var(--font-body)",
            fontSize: isMobile ? 13 : 14,
            color: "var(--color-fg)",
            margin: 0,
            lineHeight: 1.7,
          }}
        >
          Demokrati kräver att du vet vad dina folkvalda gör. Den informationen
          finns — men ingen serverar den åt dig, och de som borde går inte att
          lita på. Partierna är partiska i sin natur; deras jobb är att vinna
          din röst, inte att upplysa dig. Public service jagar ofta det som ger
          rubriker, inte det som är viktigt men tråkigt. Kvar finns ett tomrum:
          den oglamorösa, avgörande informationen om vad som faktiskt händer —
          den som ingen tjänar på att ge dig.
        </p>
        <p
          style={{
            fontFamily: "var(--font-body)",
            fontSize: isMobile ? 13 : 14,
            color: "var(--color-fg)",
            margin: 0,
            lineHeight: 1.7,
          }}
        >
          Det tomrummet finns Riksdagskollen för att fylla. Vi samlar vad dina
          representanter faktiskt gör — röster, löften, beslut — och visar det
          med källa, utan att vinkla. Ingen agenda. Ingen knuff åt något håll.
          Bara fakta, framlagda så att du själv kan döma.
        </p>
        <p
          style={{
            fontFamily: "var(--font-body)",
            fontSize: isMobile ? 12 : 13,
            color: "var(--color-fg-muted)",
            margin: 0,
            lineHeight: 1.7,
            fontStyle: "italic",
          }}
        >
          Ansvaret att förstå landar på dig. Vi gör det möjligt att bära. Läs
          på, dra dina egna slutsatser. Det är så det ska gå till.
        </p>
      </div>

      {/* ── Feed → committee index → party index ─────────────────────── */}
      <DecisionFeed />
      <CommitteeIndex />

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
