import { PARTY_COLORS, PARTIES } from "@/shared/design";
import type { VoteDetail, PartyVotePosition } from "@/shared/types";

function VoteIcon({ vote }: { vote: string | undefined }) {
  const base: React.CSSProperties = {
    display: "inline-flex",
    alignItems: "center",
    justifyContent: "center",
    width: 32,
    height: 32,
    borderRadius: "50%",
    fontSize: 15,
    fontWeight: 800,
    flexShrink: 0,
  };

  if (vote === "Ja") {
    return <span style={{ ...base, background: "#16a34a", color: "#fff" }}>✓</span>;
  }
  if (vote === "Nej") {
    return <span style={{ ...base, border: "2.5px solid #dc2626", color: "#dc2626" }}>✕</span>;
  }
  return <span style={{ ...base, border: "2.5px solid #d97706", color: "#d97706" }}>○</span>;
}

function PartyChip({ party }: { party: string }) {
  const c = PARTY_COLORS[party];
  return (
    <span
      style={{
        background: c?.bg ?? "#666",
        color: c?.text ?? "#fff",
        padding: "3px 10px",
        borderRadius: 5,
        fontSize: 11,
        fontWeight: 800,
        letterSpacing: "0.05em",
      }}
    >
      {party}
    </span>
  );
}

interface VoteShareCardProps {
  data: VoteDetail;
}

export function VoteShareCard({ data }: VoteShareCardProps) {
  const { partyBreakdown, documentTitle, session, beteckning, forslagspunkt } = data;

  const posMap: Record<string, PartyVotePosition> = {};
  partyBreakdown.forEach((p) => { posMap[p.party] = p; });

  const allKnown = PARTIES as string[];
  const extra = partyBreakdown.map((p) => p.party).filter((p) => !allKnown.includes(p));
  const orderedParties = [...allKnown, ...extra].filter((p) => !!posMap[p]);

  const jaParties = orderedParties.filter((p) => posMap[p]?.dominantVote === "Ja");
  const nejParties = orderedParties.filter((p) => posMap[p]?.dominantVote === "Nej");
  const avstarParties = orderedParties.filter((p) => posMap[p]?.dominantVote === "Avstår");

  const totalJa = partyBreakdown.reduce((s, p) => s + p.jaCount, 0);
  const totalNej = partyBreakdown.reduce((s, p) => s + p.nejCount, 0);
  const status = data.status ?? (totalJa >= totalNej ? "Bifall" : "Avslag");
  const passed = status === "Bifall";
  const statusColor = passed ? "#16a34a" : "#dc2626";

  const activeSides = [jaParties, avstarParties, nejParties].filter((g) => g.length > 0);
  const gridCols = activeSides.length === 1 ? "1fr" : activeSides.length === 2 ? "1fr 1fr" : "1fr 1fr 1fr";

  return (
    <div
      id="vote-share-card"
      style={{
        background: "#ffffff",
        color: "#111827",
        borderRadius: 16,
        padding: "32px 36px",
        width: "100%",
        maxWidth: 620,
        fontFamily: "system-ui, -apple-system, 'Helvetica Neue', sans-serif",
        boxSizing: "border-box",
      }}
    >
      {/* Header */}
      <p
        style={{
          fontFamily: "monospace",
          fontSize: 10,
          fontWeight: 700,
          letterSpacing: "0.18em",
          textTransform: "uppercase",
          color: "#9ca3af",
          margin: "0 0 10px",
        }}
      >
        Så röstade partierna om
      </p>
      <h2
        style={{
          fontSize: 22,
          fontWeight: 800,
          lineHeight: 1.25,
          margin: "0 0 8px",
          color: "#111827",
        }}
      >
        {documentTitle || `${beteckning} punkt ${forslagspunkt}`}
      </h2>
      <p
        style={{
          fontFamily: "monospace",
          fontSize: 11,
          color: "#6b7280",
          margin: "0 0 28px",
          letterSpacing: "0.04em",
        }}
      >
        Riksdagsomröstning
        {session ? ` · ${session}` : ""}
        {beteckning ? ` · ${beteckning}` : ""}
        {forslagspunkt ? ` punkt ${forslagspunkt}` : ""}
      </p>

      {/* Party vote grid */}
      <div
        style={{
          background: "#f9fafb",
          borderRadius: 12,
          padding: "20px 24px",
          marginBottom: 20,
          display: "flex",
          flexWrap: "wrap",
          gap: 16,
          justifyContent: "center",
        }}
      >
        {orderedParties.map((party) => {
          const pos = posMap[party];
          const c = PARTY_COLORS[party];
          return (
            <div
              key={party}
              style={{
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                gap: 8,
                flex: "0 0 auto",
              }}
            >
              <span
                style={{
                  background: c?.bg ?? "#666",
                  color: c?.text ?? "#fff",
                  padding: "4px 10px",
                  borderRadius: 6,
                  fontSize: 12,
                  fontWeight: 800,
                  letterSpacing: "0.05em",
                }}
              >
                {party}
              </span>
              <VoteIcon vote={pos?.dominantVote} />
            </div>
          );
        })}
      </div>

      {/* Legend */}
      <div
        style={{
          display: "flex",
          gap: 20,
          marginBottom: 20,
          flexWrap: "wrap",
        }}
      >
        {[
          { color: "#16a34a", label: "Röstade Ja", filled: true },
          { color: "#d97706", label: "Avstår", filled: false },
          { color: "#dc2626", label: "Röstade Nej", filled: false },
        ].map((item) => (
          <div
            key={item.label}
            style={{
              display: "flex",
              alignItems: "center",
              gap: 6,
              fontSize: 11,
              color: "#6b7280",
            }}
          >
            <span
              style={{
                display: "inline-block",
                width: 14,
                height: 14,
                borderRadius: "50%",
                background: item.filled ? item.color : "transparent",
                border: item.filled ? "none" : `2px solid ${item.color}`,
                flexShrink: 0,
              }}
            />
            {item.label}
          </div>
        ))}
      </div>

      {/* Yes / Abstain / No split */}
      {activeSides.length > 0 && (
        <div
          style={{
            display: "grid",
            gridTemplateColumns: gridCols,
            gap: 10,
            marginBottom: 20,
          }}
        >
          {jaParties.length > 0 && (
            <div
              style={{
                padding: "14px 16px",
                background: "#f0fdf4",
                borderRadius: 10,
                border: "1px solid #bbf7d0",
              }}
            >
              <p
                style={{
                  fontFamily: "monospace",
                  fontSize: 10,
                  fontWeight: 700,
                  color: "#15803d",
                  letterSpacing: "0.12em",
                  textTransform: "uppercase",
                  margin: "0 0 10px",
                }}
              >
                ✓ JA-SIDAN
              </p>
              <div style={{ display: "flex", flexWrap: "wrap", gap: 5 }}>
                {jaParties.map((p) => <PartyChip key={p} party={p} />)}
              </div>
            </div>
          )}
          {avstarParties.length > 0 && (
            <div
              style={{
                padding: "14px 16px",
                background: "#fffbeb",
                borderRadius: 10,
                border: "1px solid #fde68a",
              }}
            >
              <p
                style={{
                  fontFamily: "monospace",
                  fontSize: 10,
                  fontWeight: 700,
                  color: "#b45309",
                  letterSpacing: "0.12em",
                  textTransform: "uppercase",
                  margin: "0 0 10px",
                }}
              >
                ○ AVSTÅR
              </p>
              <div style={{ display: "flex", flexWrap: "wrap", gap: 5 }}>
                {avstarParties.map((p) => <PartyChip key={p} party={p} />)}
              </div>
            </div>
          )}
          {nejParties.length > 0 && (
            <div
              style={{
                padding: "14px 16px",
                background: "#fef2f2",
                borderRadius: 10,
                border: "1px solid #fecaca",
              }}
            >
              <p
                style={{
                  fontFamily: "monospace",
                  fontSize: 10,
                  fontWeight: 700,
                  color: "#dc2626",
                  letterSpacing: "0.12em",
                  textTransform: "uppercase",
                  margin: "0 0 10px",
                }}
              >
                ✕ NEJ-SIDAN
              </p>
              <div style={{ display: "flex", flexWrap: "wrap", gap: 5 }}>
                {nejParties.map((p) => <PartyChip key={p} party={p} />)}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Result banner */}
      {(data.status || partyBreakdown.length > 0) && (
        <div
          style={{
            padding: "14px 20px",
            background: passed ? "#f0fdf4" : "#fef2f2",
            borderRadius: 10,
            border: `1px solid ${passed ? "#bbf7d0" : "#fecaca"}`,
            display: "flex",
            alignItems: "center",
            gap: 14,
            marginBottom: 24,
          }}
        >
          <span
            style={{
              width: 36,
              height: 36,
              borderRadius: "50%",
              background: statusColor,
              display: "inline-flex",
              alignItems: "center",
              justifyContent: "center",
              color: "#fff",
              fontSize: 18,
              fontWeight: 800,
              flexShrink: 0,
            }}
          >
            {passed ? "✓" : "✕"}
          </span>
          <div>
            <p
              style={{
                fontFamily: "monospace",
                fontSize: 10,
                color: "#6b7280",
                margin: "0 0 3px",
                letterSpacing: "0.12em",
                textTransform: "uppercase",
              }}
            >
              Beslutet i Riksdagen
            </p>
            <p style={{ fontSize: 15, fontWeight: 700, margin: 0, color: statusColor }}>
              Förslaget {passed ? "gick igenom" : "avslogs"} · Majoriteten röstade{" "}
              {passed ? "JA" : "NEJ"}
            </p>
          </div>
        </div>
      )}

      {/* Footer */}
      <div
        style={{
          paddingTop: 16,
          borderTop: "1px solid #e5e7eb",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <span
          style={{
            fontFamily: "monospace",
            fontSize: 12,
            fontWeight: 700,
            color: "#374151",
            letterSpacing: "0.06em",
          }}
        >
          Tre Kammare
        </span>
        <span style={{ fontFamily: "monospace", fontSize: 10, color: "#9ca3af" }}>
          Källa: riksdagen.se{beteckning ? ` · ${beteckning}` : ""}
        </span>
      </div>
    </div>
  );
}
