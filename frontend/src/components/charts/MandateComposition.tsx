import type { Party } from "@/types/democracy";
import { SourceMarker } from "@/components/sources/SourceMarker";

interface MandateCompositionProps {
  ruling: {
    parties: Party[];
    support?: Party[];
    opposition: Party[];
    type: string;
  };
  totalSeats: number;
  sourceId?: string;
}

function hexToRgba(hex: string, alpha: number): string {
  if (!/^#[0-9a-fA-F]{6}$/.test(hex)) return `rgba(0,0,0,${alpha})`;
  const r = parseInt(hex.slice(1, 3), 16);
  const g = parseInt(hex.slice(3, 5), 16);
  const b = parseInt(hex.slice(5, 7), 16);
  return `rgba(${r},${g},${b},${alpha})`;
}

function PartyRow({
  parties,
  size,
}: {
  parties: Party[];
  size: "large" | "small";
}) {
  return (
    <div style={{ display: "flex", gap: 5 }}>
      {parties.map((p) => (
        <div
          key={p.short}
          style={{
            borderLeft: `5px solid ${p.color}`,
            padding: size === "large" ? "9px 14px" : "7px 10px",
            background: hexToRgba(p.color, 0.07),
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            flex: 1,
            minWidth: 0,
          }}
        >
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: size === "large" ? 12 : 11,
              fontWeight: 700,
              color: p.color,
            }}
          >
            {p.short}
          </span>
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: size === "large" ? 22 : 17,
              fontWeight: size === "large" ? 800 : 700,
              color: p.color,
              lineHeight: 1,
            }}
          >
            {p.seats}
          </span>
        </div>
      ))}
    </div>
  );
}

export function MandateComposition({
  ruling,
  totalSeats,
  sourceId = "scb-ltmandat",
}: MandateCompositionProps) {
  if (totalSeats === 0) {
    return (
      <div
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 11,
          color: "var(--color-fg-muted)",
          padding: "24px 0",
          textAlign: "center",
          lineHeight: 1.5,
        }}
      >
        Mandatdata saknas för denna region.
      </div>
    );
  }

  const styreParties = [...ruling.parties, ...(ruling.support ?? [])];
  const oppositionParties = ruling.opposition;
  const rulingSeats = styreParties.reduce((s, p) => s + p.seats, 0);

  return (
    <div>
      {/* STYRE */}
      <div
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 7,
          letterSpacing: "0.15em",
          color: "var(--color-accent)",
          marginBottom: 7,
        }}
      >
        STYRE
      </div>
      {styreParties.length > 0 && <PartyRow parties={styreParties} size="large" />}

      <div
        style={{ borderTop: "1px solid var(--color-border)", margin: "12px 0" }}
      />

      {/* OPPOSITION */}
      <div
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 7,
          letterSpacing: "0.15em",
          color: "var(--color-fg-muted)",
          marginBottom: 7,
        }}
      >
        OPPOSITION
      </div>
      {oppositionParties.length > 0 && <PartyRow parties={oppositionParties} size="small" />}

      {/* Coalition footer */}
      <div
        style={{
          borderTop: "1px solid var(--color-border)",
          marginTop: 14,
          paddingTop: 14,
          display: "flex",
          alignItems: "baseline",
          gap: 12,
        }}
      >
        <span
          style={{
            fontFamily: "var(--font-serif)",
            fontStyle: "italic",
            fontSize: 18,
            color: "var(--color-fg)",
          }}
        >
          {ruling.type}
        </span>
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            color: "var(--color-fg-muted)",
            letterSpacing: "0.1em",
          }}
        >
          MAJORITET {rulingSeats}/{totalSeats}
        </span>
        <SourceMarker sourceId={sourceId} />
      </div>
    </div>
  );
}
