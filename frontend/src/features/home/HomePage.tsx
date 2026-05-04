import { Link } from "react-router-dom";
import { useRiksdag } from "@/hooks/useDemocracy";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { mockRegion, mockKommun } from "@/mock/democracy";
import { StackBar, Pill } from "@/components/charts";
import type { LevelData, Party } from "@/types/democracy";

// ── Helpers ──────────────────────────────────────────────────────────────────

function pillTone(status: string): "pass" | "fail" | "pending" | "neutral" {
  if (status === "Bifall") return "pass";
  if (status === "Avslag") return "fail";
  if (status === "Återremiss") return "pending";
  return "neutral";
}

function allParties(data: LevelData): Party[] {
  const ruling = data.ruling.parties ?? [];
  const support = data.ruling.support ?? [];
  const opp = data.ruling.opposition ?? [];
  return [...ruling, ...support, ...opp];
}

// ── ChambersHero ─────────────────────────────────────────────────────────────

const rings = [
  {
    tag: "I",
    label: "Riksdagen",
    sub: "349 MANDAT · NATIONELLT",
    inset: 0,
    pulses: 4,
    activeToday: 3,
  },
  {
    tag: "II",
    label: "Regionen",
    sub: "21 REGIONER · VÅRD & TRAFIK",
    inset: 52,
    pulses: 3,
    activeToday: 1,
  },
  {
    tag: "III",
    label: "Kommunen",
    sub: "290 KOMMUNER · SKOLA, OMSORG",
    inset: 104,
    pulses: 5,
    activeToday: 2,
  },
] as const;

interface ChambersHeroProps {
  riksdagParties: Party[];
  regionParties: Party[];
  kommunParties: Party[];
}

function ChambersHero({ riksdagParties, regionParties, kommunParties }: ChambersHeroProps) {
  const partyStrips = [riksdagParties, regionParties, kommunParties];

  return (
    <div style={{ position: "relative", height: 320, margin: "0 40px" }}>
      {rings.map((ring, i) => (
        <div
          key={ring.tag}
          style={{
            position: "absolute",
            top: ring.inset,
            left: ring.inset + 80,
            right: ring.inset + 80,
            bottom: ring.inset,
            border: "1px solid var(--color-border)",
            borderRadius: 3,
            ...(i === 2 ? { background: "var(--color-plane)" } : {}),
            ...(i === 0
              ? { boxShadow: "0 30px 60px -20px rgba(0,0,0,0.4)" }
              : {}),
          }}
        >
          {/* Label tab — top-left edge */}
          <div
            style={{
              position: "absolute",
              top: 0,
              left: 20,
              transform: "translateY(-50%)",
              display: "flex",
              alignItems: "baseline",
              gap: 8,
              background: "var(--color-bg)",
              padding: "0 6px",
            }}
          >
            <span
              style={{
                fontFamily: "var(--font-serif)",
                fontStyle: "italic",
                color: "var(--color-accent-2)",
                fontSize: 14,
                lineHeight: 1,
              }}
            >
              {ring.tag}
            </span>
            <span
              style={{
                fontFamily: "var(--font-serif)",
                fontSize: 18,
                color: "var(--color-fg)",
                lineHeight: 1,
              }}
            >
              {ring.label}
            </span>
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                color: "var(--color-fg-muted)",
                letterSpacing: "0.12em",
              }}
            >
              {ring.sub}
            </span>
          </div>

          {/* Pulse dots — bottom edge */}
          {Array.from({ length: ring.pulses }).map((_, j) => (
            <div
              key={j}
              className={j < ring.activeToday ? "animate-pulse-dot" : undefined}
              style={{
                position: "absolute",
                bottom: -4,
                left: `${10 + (j / ring.pulses) * 78}%`,
                width: 9,
                height: 9,
                borderRadius: "50%",
                background:
                  j < ring.activeToday
                    ? "var(--color-pulse)"
                    : "var(--color-accent)",
                border: "2px solid var(--color-bg)",
              }}
            />
          ))}

          {/* Ruling-party color strip — top-right */}
          <div
            style={{
              position: "absolute",
              top: 14,
              right: 14,
              display: "flex",
              gap: 3,
            }}
          >
            {partyStrips[i].slice(0, 2).map((p, k) => (
              <div
                key={k}
                style={{
                  width: 18,
                  height: 4,
                  borderRadius: 1,
                  background: p.color,
                }}
              />
            ))}
          </div>
        </div>
      ))}

      {/* Center legend */}
      <div
        style={{
          position: "absolute",
          top: "50%",
          left: "50%",
          transform: "translate(-50%, -50%)",
          textAlign: "center",
          pointerEvents: "none",
        }}
      >
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "0.15em",
            color: "var(--color-fg-muted)",
            marginBottom: 8,
          }}
        >
          12 PÅGÅENDE · 6 BESLUT IDAG
        </div>
        <div style={{ display: "flex", gap: 12, justifyContent: "center", alignItems: "center" }}>
          <span style={{ display: "flex", alignItems: "center", gap: 4 }}>
            <span
              style={{
                width: 8,
                height: 8,
                borderRadius: "50%",
                background: "var(--color-accent)",
                display: "inline-block",
              }}
            />
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 9,
                color: "var(--color-fg-muted)",
                letterSpacing: "0.1em",
              }}
            >
              PÅGÅR
            </span>
          </span>
          <span style={{ display: "flex", alignItems: "center", gap: 4 }}>
            <span
              className="animate-pulse-dot"
              style={{
                width: 8,
                height: 8,
                borderRadius: "50%",
                background: "var(--color-pulse)",
                display: "inline-block",
              }}
            />
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 9,
                color: "var(--color-fg-muted)",
                letterSpacing: "0.1em",
              }}
            >
              BESLUT IDAG
            </span>
          </span>
        </div>
      </div>
    </div>
  );
}

// ── StatusStrip ───────────────────────────────────────────────────────────────

interface StatusCellProps {
  tag: string;
  data: LevelData;
  to: string;
}

function StatusCell({ tag, data, to }: StatusCellProps) {
  const parties = allParties(data);
  const totalSeats = parties.reduce((s, p) => s + p.seats, 0);
  const rulingSeats = data.ruling.parties.reduce((s, p) => s + p.seats, 0);

  return (
    <div
      style={{
        background: "var(--color-bg)",
        padding: 22,
        display: "flex",
        flexDirection: "column",
        gap: 14,
      }}
    >
      {/* Header */}
      <div style={{ display: "flex", alignItems: "baseline", gap: 10 }}>
        <span
          style={{
            fontFamily: "var(--font-serif)",
            fontStyle: "italic",
            color: "var(--color-accent-2)",
            fontSize: 20,
          }}
        >
          {tag}
        </span>
        <span
          style={{
            fontFamily: "var(--font-serif)",
            fontSize: 22,
            color: "var(--color-fg)",
          }}
        >
          {data.title}
        </span>
        <span
          className="animate-pulse-dot"
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            color: "var(--color-pulse)",
            letterSpacing: "0.1em",
            marginLeft: "auto",
          }}
        >
          ● {data.liveVotes.filter((v) => v.time.startsWith("Idag")).length} IDAG
        </span>
      </div>

      {/* Party stack bar */}
      <StackBar
        segments={parties.map((p) => ({ color: p.color, value: p.seats, name: p.name, short: p.short }))}
        height={8}
        rounded={false}
      />

      {/* Styre */}
      <div>
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 9,
            letterSpacing: "0.15em",
            color: "var(--color-fg-muted)",
            marginBottom: 6,
          }}
        >
          STYRE
        </div>
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            color: "var(--color-fg-muted)",
            marginBottom: 6,
          }}
        >
          {data.ruling.type}
        </div>
        <div style={{ display: "flex", flexWrap: "wrap", gap: "4px 6px", marginBottom: 6 }}>
          {data.ruling.parties.map((p) => (
            <span
              key={p.short}
              style={{
                display: "inline-flex",
                alignItems: "center",
                gap: 4,
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                color: "var(--color-fg)",
              }}
            >
              <span
                style={{
                  width: 7,
                  height: 7,
                  borderRadius: "50%",
                  background: p.color,
                  display: "inline-block",
                  flexShrink: 0,
                }}
              />
              {p.short}{" "}
              <span style={{ color: "var(--color-fg-muted)" }}>{p.seats}</span>
            </span>
          ))}
        </div>
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 9,
            color: "var(--color-fg-muted)",
            letterSpacing: "0.12em",
          }}
        >
          MAJORITET {rulingSeats}/{totalSeats}
        </div>
      </div>

      {/* 2 latest live votes */}
      <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        {data.liveVotes.slice(0, 2).map((v, i) => (
          <div key={i} style={{ display: "flex", alignItems: "center", gap: 8 }}>
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                color: "var(--color-fg-muted)",
                flexShrink: 0,
                minWidth: 60,
              }}
            >
              {v.time}
            </span>
            <span
              style={{
                fontSize: 12,
                color: "var(--color-fg)",
                flex: 1,
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
              }}
            >
              {v.title}
            </span>
            <Pill tone={pillTone(v.status)}>{v.status}</Pill>
          </div>
        ))}
      </div>

      {/* Footer link */}
      <Link
        to={to}
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 10,
          letterSpacing: "0.15em",
          color: "var(--color-accent)",
          textDecoration: "none",
          marginTop: 2,
        }}
      >
        GÅ TILL KAMMARE →
      </Link>
    </div>
  );
}

// ── PulsStrip ────────────────────────────────────────────────────────────────

const PULSE_ACTIVE = new Set([5, 12, 19, 26, 31, 38, 42]);
const PULSE_LEVELS: Record<number, number> = { 12: 0, 31: 0, 19: 1, 38: 1, 5: 2, 26: 2, 42: 2 };
const LEVEL_COLORS = [
  "var(--color-pulse)",
  "var(--color-accent)",
  "var(--color-accent-2)",
];

interface PulseStripProps {
  compact?: boolean;
}

function PulseStrip({ compact }: PulseStripProps) {
  return (
    <div style={{ padding: compact ? "16px 14px 24px" : "22px 32px 28px" }}>
      {/* Header */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 16,
          marginBottom: 12,
          flexWrap: "wrap",
        }}
      >
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "0.15em",
            color: "var(--color-fg-muted)",
          }}
        >
          DYGNETS PULS — 9 BESLUT I 3 KAMMARE
        </span>
        <span style={{ display: "flex", gap: 12, alignItems: "center" }}>
          {(["I", "II", "III"] as const).map((tag, i) => (
            <span
              key={tag}
              style={{
                display: "inline-flex",
                alignItems: "center",
                gap: 4,
                fontFamily: "var(--font-mono)",
                fontSize: 9,
                color: "var(--color-fg-muted)",
              }}
            >
              <span
                style={{
                  display: "inline-block",
                  width: 10,
                  height: 10,
                  background: LEVEL_COLORS[2 - i],
                }}
              />
              {tag}
            </span>
          ))}
        </span>
      </div>

      {/* 48 cells */}
      <div style={{ display: "flex", gap: 2 }}>
        {Array.from({ length: 48 }).map((_, idx) => {
          const active = PULSE_ACTIVE.has(idx);
          const level = PULSE_LEVELS[idx];
          const bg = active ? LEVEL_COLORS[level] : "var(--color-track)";
          return (
            <div
              key={idx}
              style={{
                flex: 1,
                height: 28,
                borderRadius: 1,
                background: bg,
              }}
            />
          );
        })}
      </div>

      {/* Time labels */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          marginTop: 6,
          fontFamily: "var(--font-mono)",
          fontSize: 9,
          color: "var(--color-fg-muted)",
          letterSpacing: "0.08em",
        }}
      >
        {["00:00", "06:00", "12:00", "18:00", "NU"].map((t) => (
          <span key={t}>{t}</span>
        ))}
      </div>
    </div>
  );
}

// ── HomePage ─────────────────────────────────────────────────────────────────

export function HomePage() {
  const { data: riksdagData, isLoading } = useRiksdag();
  const isMobile = useMediaQuery("(max-width: 640px)");

  if (isLoading || !riksdagData) {
    return (
      <div className="sdt-page" style={{ padding: "80px 32px" }}>
        <div
          style={{
            height: 60,
            background: "var(--color-track)",
            borderRadius: 4,
            marginBottom: 24,
            maxWidth: 480,
          }}
        />
        <div
          style={{
            height: 320,
            background: "var(--color-track)",
            borderRadius: 4,
          }}
        />
      </div>
    );
  }

  return (
    <div className="sdt-page">
      {/* ── Section 1: Hero text ─────────────────────────────────────── */}
      <div style={{ padding: isMobile ? "40px 14px 20px" : "80px 32px 24px" }}>
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "2.5px",
            color: "var(--color-fg-muted)",
            textTransform: "uppercase",
            marginBottom: 16,
          }}
        >
          VEM BESTÄMMER · VAR · JUST NU
        </div>
        <h1
          style={{
            fontFamily: "var(--font-serif)",
            fontSize: isMobile ? 36 : 60,
            fontWeight: 400,
            letterSpacing: "-1.8px",
            color: "var(--color-fg)",
            margin: 0,
            lineHeight: 1.05,
          }}
        >
          Tre{" "}
          <em
            style={{
              fontStyle: "italic",
              color: "var(--color-accent-2)",
            }}
          >
            kammare
          </em>
          , ett samhälle.
        </h1>
      </div>

      {/* ── Section 2: ChambersHero ──────────────────────────────────── */}
      {!isMobile && (
        <ChambersHero
          riksdagParties={riksdagData.ruling.parties}
          regionParties={mockRegion.ruling.parties.slice(0, 2)}
          kommunParties={mockKommun.ruling.parties.slice(0, 2)}
        />
      )}

      {/* ── Section 3: Three-column status strip ─────────────────────── */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: isMobile ? "1fr" : "repeat(3, 1fr)",
          gap: 1,
          background: "var(--color-border)",
          margin: isMobile ? "14px 14px 0" : "24px 32px 0",
          borderTop: "1px solid var(--color-border)",
          borderBottom: "1px solid var(--color-border)",
        }}
      >
        <StatusCell tag="I" data={riksdagData} to="/riksdag" />
        <StatusCell tag="II" data={mockRegion} to="/region" />
        <StatusCell tag="III" data={mockKommun} to="/kommun" />
      </div>

      {/* ── Section 4: 24h pulse strip ───────────────────────────────── */}
      <PulseStrip compact={isMobile} />
    </div>
  );
}
