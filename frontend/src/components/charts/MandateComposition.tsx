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

// Region/kommun election data only carries the party abbreviation, so map the
// eight Riksdag parties to full names for the list. Falls back to p.name.
const PARTY_FULL_NAMES: Record<string, string> = {
  S: "Socialdemokraterna",
  M: "Moderaterna",
  SD: "Sverigedemokraterna",
  C: "Centerpartiet",
  V: "Vänsterpartiet",
  KD: "Kristdemokraterna",
  L: "Liberalerna",
  MP: "Miljöpartiet",
};

function hexToRgba(hex: string, alpha: number): string {
  if (!/^#[0-9a-fA-F]{6}$/.test(hex)) return `rgba(0,0,0,${alpha})`;
  const r = parseInt(hex.slice(1, 3), 16);
  const g = parseInt(hex.slice(3, 5), 16);
  const b = parseInt(hex.slice(5, 7), 16);
  return `rgba(${r},${g},${b},${alpha})`;
}

function PartyList({ parties }: { parties: Party[] }) {
  return (
    <div>
      {parties.map((p) => {
        const full = PARTY_FULL_NAMES[p.short] ?? p.name;
        const showName = !!full && full !== p.short;
        return (
          <div
            key={p.short}
            style={{
              display: "flex",
              alignItems: "center",
              borderLeft: `5px solid ${p.color}`,
              background: hexToRgba(p.color, 0.07),
              padding: "9px 14px",
              marginBottom: 4,
            }}
          >
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 12,
                fontWeight: 700,
                color: p.color,
                minWidth: 34,
              }}
            >
              {p.short}
            </span>
            {showName && (
              <span
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                  color: "var(--color-fg-muted)",
                  flex: 1,
                  marginLeft: 6,
                  whiteSpace: "nowrap",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                }}
              >
                {full}
              </span>
            )}
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 18,
                fontWeight: 800,
                color: p.color,
                lineHeight: 1,
                marginLeft: showName ? 8 : "auto",
              }}
            >
              {p.seats}
            </span>
          </div>
        );
      })}
    </div>
  );
}

function BlocLabel({
  label,
  total,
  accent,
  totalColor,
}: {
  label: string;
  total: number;
  accent: boolean;
  totalColor: string;
}) {
  return (
    <div
      style={{
        display: "flex",
        justifyContent: "space-between",
        alignItems: "baseline",
        marginBottom: 7,
      }}
    >
      <span
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 7,
          letterSpacing: "0.15em",
          color: accent ? "var(--color-accent)" : "var(--color-fg-muted)",
        }}
      >
        {label}
      </span>
      <span
        style={{
          fontFamily: "var(--font-mono)",
          fontWeight: 800,
          fontSize: 15,
          color: totalColor,
        }}
      >
        {total}
        <span
          style={{
            fontSize: 8,
            fontWeight: 400,
            color: "var(--color-fg-muted)",
            letterSpacing: "0.15em",
            marginLeft: 4,
          }}
        >
          MANDAT
        </span>
      </span>
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
        Mandatdata saknas.
      </div>
    );
  }

  const styreParties = [...ruling.parties, ...(ruling.support ?? [])];
  const oppositionParties = ruling.opposition;
  const styreSeats = styreParties.reduce((s, p) => s + p.seats, 0);
  const oppositionSeats = oppositionParties.reduce((s, p) => s + p.seats, 0);
  const styreColor = styreParties[0]?.color ?? "var(--color-fg)";
  const oppColor = oppositionParties[0]?.color ?? "var(--color-fg)";

  return (
    <div>
      {styreParties.length > 0 && (
        <>
          <BlocLabel label="STYRE" total={styreSeats} accent totalColor={styreColor} />
          <PartyList parties={styreParties} />
        </>
      )}

      {styreParties.length > 0 && oppositionParties.length > 0 && (
        <div style={{ borderTop: "1px solid var(--color-border)", margin: "12px 0" }} />
      )}

      {oppositionParties.length > 0 && (
        <>
          <BlocLabel
            label="OPPOSITION"
            total={oppositionSeats}
            accent={false}
            totalColor={oppColor}
          />
          <PartyList parties={oppositionParties} />
        </>
      )}

      <div style={{ marginTop: 10, textAlign: "right" }}>
        <SourceMarker sourceId={sourceId} />
      </div>
    </div>
  );
}
