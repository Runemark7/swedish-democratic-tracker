import { Link } from "react-router-dom";
import { useRiksdag } from "@/hooks/useDemocracy";
import { Hemicycle, Donut, HBars, Pill } from "@/components/charts";
import type { LiveVote, Party } from "@/types/democracy";

function beslutHref(v: LiveVote): string | null {
  if (!v.beteckning) return null;
  const p = new URLSearchParams({ title: v.title, status: v.status, tag: v.tag, time: v.time });
  return `/beslut/${encodeURIComponent(v.beteckning)}?${p}`;
}

// ── Helpers ───────────────────────────────────────────────────────────────────

function pillTone(status: string): "pass" | "fail" | "pending" | "neutral" {
  if (status === "Bifall") return "pass";
  if (status === "Avslag") return "fail";
  if (status === "Återremiss") return "pending";
  return "neutral";
}

const BUDGET_COLORS = [
  "#0b3d7a",
  "#2d6fa8",
  "#5a9fd0",
  "#8bc0e0",
  "#c8a13b",
  "#d0533f",
  "#7a8390",
  "#b3bcc5",
];

// ── Skeleton ──────────────────────────────────────────────────────────────────

function Skeleton() {
  return (
    <div className="sdt-page" style={{ padding: "40px 32px" }}>
      {[200, 120, 300].map((h, i) => (
        <div
          key={i}
          style={{
            height: h,
            background: "var(--color-track)",
            borderRadius: 4,
            marginBottom: 16,
          }}
        />
      ))}
    </div>
  );
}

// ── RiksdagPage ───────────────────────────────────────────────────────────────

export function RiksdagPage() {
  const { data, isLoading } = useRiksdag();

  if (isLoading || !data) return <Skeleton />;

  const { ruling, liveVotes, budget, agenda } = data;

  // Build hemicycle groups: opposition left → support → ruling right
  const hemicycleGroups = [
    ...ruling.opposition.map((p: Party) => ({ color: p.color, count: p.seats })),
    ...(ruling.support ?? []).map((p: Party) => ({ color: p.color, count: p.seats })),
    ...ruling.parties.map((p: Party) => ({ color: p.color, count: p.seats })),
  ];

  const allParties = [
    ...ruling.parties,
    ...(ruling.support ?? []),
    ...ruling.opposition,
  ];
  const totalSeats = allParties.reduce((s, p) => s + p.seats, 0);
  const rulingSeats = ruling.parties.reduce((s, p) => s + p.seats, 0);

  const budgetSegments = budget.areas.map((a, i) => ({
    name: a.name,
    value: a.value,
    color: BUDGET_COLORS[i % BUDGET_COLORS.length],
    pct: a.pct,
  }));

  return (
    <div className="sdt-page">
      {/* ── Hero ─────────────────────────────────────────────────────── */}
      <div
        style={{
          display: "flex",
          alignItems: "flex-end",
          gap: 28,
          padding: "40px 32px 0",
        }}
      >
        <span
          style={{
            fontFamily: "var(--font-serif)",
            fontStyle: "italic",
            color: "var(--color-accent-2)",
            fontSize: 88,
            fontWeight: 400,
            letterSpacing: "-3px",
            lineHeight: 0.85,
            flexShrink: 0,
          }}
        >
          I
        </span>
        <div>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              letterSpacing: "2px",
              color: "var(--color-fg-muted)",
              marginBottom: 6,
            }}
          >
            KAMMARE ETT
          </div>
          <div
            style={{
              fontFamily: "var(--font-serif)",
              fontSize: 52,
              fontWeight: 400,
              letterSpacing: "-1.5px",
              color: "var(--color-fg)",
              lineHeight: 1,
              marginBottom: 8,
            }}
          >
            Riksdagen
          </div>
          <div
            style={{
              fontSize: 14,
              color: "var(--color-fg-muted)",
            }}
          >
            Sveriges nationella parlament — 349 ledamöter · Mandatperiod 2022–2026
          </div>
        </div>
      </div>

      {/* ── Main grid ────────────────────────────────────────────────── */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr",
          gap: 1,
          border: "1px solid var(--color-border)",
          margin: "22px 32px 0",
          background: "var(--color-border)",
        }}
      >
        {/* Left card — MANDAT */}
        <div
          style={{
            background: "var(--color-sdt-surface)",
            padding: 24,
          }}
        >
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              letterSpacing: "0.15em",
              color: "var(--color-fg-muted)",
              marginBottom: 16,
            }}
          >
            MANDAT · KAMMARENS SAMMANSÄTTNING
          </div>

          <Hemicycle groups={hemicycleGroups} width={440} height={150} />

          {/* Party legend */}
          <div
            style={{
              display: "flex",
              flexWrap: "wrap",
              gap: "6px 14px",
              marginTop: 12,
            }}
          >
            {allParties.map((p) => (
              <span
                key={p.short}
                style={{
                  display: "inline-flex",
                  alignItems: "center",
                  gap: 5,
                  fontFamily: "var(--font-mono)",
                  fontSize: 10,
                  color: "var(--color-fg-muted)",
                }}
              >
                <span
                  style={{
                    width: 8,
                    height: 8,
                    borderRadius: "50%",
                    background: p.color,
                    display: "inline-block",
                    flexShrink: 0,
                  }}
                />
                {p.short}{" "}
                <span style={{ color: "var(--color-fg)" }}>{p.seats}</span>
              </span>
            ))}
          </div>

          {/* Coalition */}
          <div
            style={{
              borderTop: "1px solid var(--color-border)",
              marginTop: 14,
              paddingTop: 14,
            }}
          >
            <span
              style={{
                fontFamily: "var(--font-serif)",
                fontStyle: "italic",
                fontSize: 18,
                color: "var(--color-fg)",
                marginRight: 12,
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
          </div>
        </div>

        {/* Right card — BUDGET */}
        <div
          style={{
            background: "var(--color-sdt-surface)",
            padding: 24,
          }}
        >
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              letterSpacing: "0.15em",
              color: "var(--color-fg-muted)",
              marginBottom: 16,
            }}
          >
            BUDGET {budget.year} · {budget.total}
          </div>
          <div style={{ display: "flex", gap: 20, alignItems: "flex-start" }}>
            <Donut
              segments={budgetSegments}
              size={150}
              thickness={18}
              label={budget.total}
              sublabel={budget.year}
            />
            <div style={{ flex: 1, minWidth: 0, paddingTop: 4 }}>
              <HBars
                items={budget.areas.map((a, i) => ({
                  name: a.name,
                  value: a.value,
                  color: BUDGET_COLORS[i % BUDGET_COLORS.length],
                  pct: a.pct,
                }))}
                height={6}
                gap={10}
              />
            </div>
          </div>
        </div>
      </div>

      {/* ── Bottom grid ──────────────────────────────────────────────── */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1.3fr 1fr",
          gap: 1,
          border: "1px solid var(--color-border)",
          borderTop: "none",
          margin: "1px 32px 28px",
          background: "var(--color-border)",
        }}
      >
        {/* Left card — PULS */}
        <div
          style={{
            background: "var(--color-sdt-surface)",
            padding: 24,
          }}
        >
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: 8,
              marginBottom: 16,
            }}
          >
            <span
              className="animate-pulse-dot"
              style={{
                width: 8,
                height: 8,
                borderRadius: "50%",
                background: "var(--color-pulse)",
                display: "inline-block",
                flexShrink: 0,
              }}
            />
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.15em",
                color: "var(--color-fg-muted)",
              }}
            >
              PULS · AKTUELLA BESLUT
            </span>
          </div>

          <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
            {liveVotes.map((v, i) => {
              const href = beslutHref(v);
              const row = (
                <div
                  style={{
                    display: "grid",
                    gridTemplateColumns: "18px 70px minmax(0,1fr) auto auto",
                    gap: 12,
                    alignItems: "center",
                    cursor: href ? "pointer" : "default",
                  }}
                >
                  {/* dot */}
                  <span
                    style={{
                      width: 7,
                      height: 7,
                      borderRadius: "50%",
                      background: i < 2 ? "var(--color-pulse)" : "var(--color-accent)",
                      display: "inline-block",
                      flexShrink: 0,
                    }}
                  />
                  {/* time */}
                  <span
                    style={{
                      fontFamily: "var(--font-mono)",
                      fontSize: 10,
                      color: "var(--color-fg-muted)",
                    }}
                  >
                    {v.time}
                  </span>
                  {/* title + tag */}
                  <span
                    style={{
                      fontSize: 13,
                      color: "var(--color-fg)",
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                      whiteSpace: "nowrap",
                    }}
                    title={v.title}
                  >
                    {v.title}
                    {v.tag && (
                      <span
                        style={{
                          fontFamily: "var(--font-mono)",
                          fontSize: 9,
                          color: "var(--color-fg-muted)",
                          marginLeft: 6,
                          verticalAlign: "middle",
                        }}
                      >
                        [{v.tag}]
                      </span>
                    )}
                  </span>
                  {/* pill */}
                  <Pill tone={pillTone(v.status)}>{v.status}</Pill>
                  {/* margin */}
                  <span
                    style={{
                      fontFamily: "var(--font-mono)",
                      fontSize: 10,
                      color: "var(--color-fg-muted)",
                      textAlign: "right",
                    }}
                  >
                    {v.margin ?? ""}
                  </span>
                </div>
              );
              return href ? (
                <Link key={i} to={href} style={{ textDecoration: "none", color: "inherit" }}>
                  {row}
                </Link>
              ) : (
                <div key={i}>{row}</div>
              );
            })}
          </div>
        </div>

        {/* Right card — AGENDA */}
        <div
          style={{
            background: "var(--color-sdt-surface)",
            padding: 24,
          }}
        >
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              letterSpacing: "0.15em",
              color: "var(--color-fg-muted)",
              marginBottom: 16,
            }}
          >
            AGENDA · STYRETS PRIORITERINGAR
          </div>
          <ol style={{ listStyle: "none", margin: 0, padding: 0, display: "flex", flexDirection: "column", gap: 12 }}>
            {agenda.map((item, i) => (
              <li
                key={i}
                style={{
                  display: "flex",
                  alignItems: "baseline",
                  gap: 12,
                }}
              >
                <span
                  style={{
                    fontFamily: "var(--font-serif)",
                    fontStyle: "italic",
                    color: "var(--color-accent-2)",
                    fontSize: 22,
                    lineHeight: 1,
                    minWidth: 20,
                    flexShrink: 0,
                  }}
                >
                  {i + 1}
                </span>
                <span
                  style={{
                    fontSize: 14,
                    color: "var(--color-fg)",
                    lineHeight: 1.4,
                  }}
                >
                  {item}
                </span>
              </li>
            ))}
          </ol>
        </div>
      </div>
    </div>
  );
}
