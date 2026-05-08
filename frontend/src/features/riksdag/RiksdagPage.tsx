import { useState } from "react";
import { Link } from "react-router-dom";
import { useRiksdag } from "@/hooks/useDemocracy";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { Hemicycle, Donut, HBars, DualLine, Pill, GoalBadge, Trend } from "@/components/charts";
import { AgendaList } from "@/components/AgendaList";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { SectionSource } from "@/components/sources/SectionSource";
import type { Authority, LiveVote, Party } from "@/types/democracy";

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

// ── AuthorityRow ──────────────────────────────────────────────────────────────

function AuthorityRow({
  authority,
  color,
  isOpen,
  onToggle,
}: {
  authority: Authority;
  color: string;
  isOpen: boolean;
  onToggle: () => void;
}) {
  const histMax = authority.history.length > 0
    ? Math.max(...authority.history.map((h) => h.expenditureMdkr))
    : authority.expenditureMdkr;

  return (
    <div style={{ borderBottom: "1px solid var(--color-border)" }}>
      {/* Summary row — clickable */}
      <button
        onClick={onToggle}
        style={{
          display: "flex",
          alignItems: "center",
          gap: 12,
          width: "100%",
          background: "none",
          border: "none",
          cursor: "pointer",
          padding: "10px 0",
          textAlign: "left",
        }}
      >
        <span style={{ flex: 1, fontSize: 13, color: "var(--color-fg)", fontFamily: "var(--font-body)" }}>
          {authority.slug
            ? <Link to={`/riksdag/myndigheter/${authority.slug}`} onClick={(e) => e.stopPropagation()} style={{ color: "inherit", textDecoration: "none", borderBottom: "1px dotted var(--color-border)" }}>{authority.name}</Link>
            : authority.name}
        </span>
        <span style={{ fontSize: 11, color: "var(--color-fg-muted)", fontFamily: "var(--font-mono)", whiteSpace: "nowrap" }}>
          {authority.headcount} anst
        </span>
        <span style={{ fontSize: 12, color: "var(--color-fg)", fontFamily: "var(--font-mono)", fontVariantNumeric: "tabular-nums", minWidth: 70, textAlign: "right" }}>
          {authority.expenditureMdkr.toFixed(1)} mdkr
          <SourceMarker sourceId="statskontoret-arsutfall" />
        </span>
        <span style={{ fontSize: 10, color: "var(--color-fg-muted)", marginLeft: 4, flexShrink: 0 }}>
          {isOpen ? "▲" : "▼"}
        </span>
      </button>

      {/* Bar */}
      {!isOpen && (
        <div style={{ paddingBottom: 10 }}>
          <div style={{ height: 5, borderRadius: 3, background: "var(--color-track)", overflow: "hidden" }}>
            <div style={{ height: "100%", width: "100%", borderRadius: 3, background: color, transition: "width 0.4s ease" }} />
          </div>
        </div>
      )}

      {/* Expanded detail */}
      {isOpen && (
        <div style={{ paddingBottom: 20 }}>
          <div style={{ fontSize: 11, color: "var(--color-fg-muted)", marginBottom: 14, fontFamily: "var(--font-mono)" }}>
            {authority.role} · {authority.headcount} anställda
            {authority.ministry && (
              <span style={{ marginLeft: 8, opacity: 0.7 }}>· {authority.ministry}</span>
            )}
          </div>

          {authority.history.length > 0 && authority.headcountHistory && authority.headcountHistory.length > 1 ? (
            /* Dual-line chart when both series are available */
            <DualLine
              series={[
                {
                  label: "Kostnad (mdkr)",
                  color,
                  unit: " mdkr",
                  formatValue: (v) => `${v.toFixed(1)} mdkr`,
                  points: authority.history.map((h) => ({ year: h.year, value: Math.round(h.expenditureMdkr * 10) / 10 })),
                },
                {
                  label: "Anställda",
                  color: "#5a8f6e",
                  unit: "",
                  formatValue: (v) => v.toLocaleString("sv-SE"),
                  points: authority.headcountHistory.map((h) => ({ year: h.year, value: h.headcountInt })),
                },
              ]}
            />
          ) : (
            /* Fallback: cost-only HBars when headcount is missing */
            authority.history.length > 0 && (
              <div>
                <div style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", letterSpacing: "1px", marginBottom: 10, textTransform: "uppercase" }}>
                  Kostnadsutveckling {authority.history[0].year}–{authority.history[authority.history.length - 1].year}
                </div>
                <HBars
                  items={authority.history.map((h) => ({
                    name: String(h.year),
                    value: Math.round(h.expenditureMdkr * 10) / 10,
                    color,
                  }))}
                  max={histMax}
                  unit=" mdkr"
                  height={5}
                  gap={8}
                />
              </div>
            )
          )}
        </div>
      )}
    </div>
  );
}

// ── RiksdagPage ───────────────────────────────────────────────────────────────

export function RiksdagPage() {
  const isMobile = useMediaQuery("(max-width: 640px)");
  const { data, isLoading } = useRiksdag();
  const [openAuthority, setOpenAuthority] = useState<number | null>(null);

  if (isLoading || !data) return <Skeleton />;

  const { ruling, liveVotes, budget, agenda, kpis, authorities } = data;

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
          padding: isMobile ? "18px 14px 0" : "40px 32px 0",
        }}
      >
        <span
          style={{
            fontFamily: "var(--font-serif)",
            fontStyle: "italic",
            color: "var(--color-accent-2)",
            fontSize: isMobile ? 56 : 88,
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
              fontSize: isMobile ? 28 : 52,
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
              display: isMobile ? "none" : undefined,
            }}
          >
            Sveriges nationella parlament — 349 ledamöter · Mandatperiod 2022–2026
          </div>
        </div>
      </div>

      {/* ── KPI strip ────────────────────────────────────────────────── */}
      {kpis && kpis.length > 0 && (
        <div
          style={{
            display: "grid",
            gridTemplateColumns: isMobile ? "repeat(2, 1fr)" : "repeat(4, 1fr)",
            gap: 1,
            border: "1px solid var(--color-border)",
            background: "var(--color-border)",
            margin: isMobile ? "14px 14px 0" : "22px 32px 0",
          }}
        >
          {kpis.map((k) => (
            <div
              key={k.label}
              style={{ background: "var(--color-sdt-surface)", padding: isMobile ? 14 : "20px 24px" }}
            >
              <div
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 10,
                  letterSpacing: "1.5px",
                  color: "var(--color-fg-muted)",
                  textTransform: "uppercase",
                  marginBottom: 6,
                }}
              >
                {k.label}
              </div>
              <div
                style={{
                  fontFamily: "var(--font-serif)",
                  fontSize: isMobile ? 24 : 30,
                  fontWeight: 400,
                  color: "var(--color-fg)",
                  lineHeight: 1,
                  marginBottom: 8,
                }}
              >
                {k.value}
                <SourceMarker sourceId="seed-budget-data" />
              </div>
              <div style={{ marginBottom: 12 }}>
                <Trend trend={k.trend} delta={k.delta} worseHigher={k.worseHigher} />
              </div>
              <GoalBadge
                raw={k.raw}
                target={k.target}
                worseHigher={k.worseHigher}
                unit={k.unit}
                note={k.note}
                sourceUrl={k.sourceUrl}
              />
              {k.description && (
                <div
                  style={{
                    marginTop: 12,
                    fontFamily: "var(--font-sans)",
                    fontSize: 11,
                    lineHeight: 1.5,
                    color: "var(--color-fg-muted)",
                    borderTop: "1px solid var(--color-border)",
                    paddingTop: 10,
                    display: isMobile ? "none" : undefined,
                  }}
                >
                  {k.description}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* ── Main grid ────────────────────────────────────────────────── */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: isMobile ? "1fr" : "1fr 1fr",
          gap: 1,
          border: "1px solid var(--color-border)",
          margin: isMobile ? "14px 14px 0" : "22px 32px 0",
          background: "var(--color-border)",
        }}
      >
        {/* Left card — MANDAT */}
        <div
          style={{
            background: "var(--color-sdt-surface)",
            padding: isMobile ? 16 : 24,
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
              <SourceMarker sourceId="riksdagen" />
            </span>
          </div>
          <SectionSource sourceIds={["riksdagen"]} />
        </div>

        {/* Right card — BUDGET */}
        <div
          style={{
            background: "var(--color-sdt-surface)",
            padding: isMobile ? 16 : 24,
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
            <SourceMarker sourceId="seed-budget-data" />
          </div>
          <div style={{ display: "flex", gap: 20, alignItems: isMobile ? "center" : "flex-start", flexDirection: isMobile ? "column" : "row" }}>
            <Donut
              segments={budgetSegments}
              size={isMobile ? 140 : 150}
              thickness={isMobile ? 16 : 18}
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
          <SectionSource sourceIds={["seed-budget-data"]} />
        </div>
      </div>

      {/* ── Bottom grid ──────────────────────────────────────────────── */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: isMobile ? "1fr" : "1.3fr 1fr",
          gap: 1,
          border: "1px solid var(--color-border)",
          borderTop: "none",
          margin: isMobile ? "1px 14px 28px" : "1px 32px 28px",
          background: "var(--color-border)",
        }}
      >
        {/* Left card — PULS */}
        <div
          style={{
            background: "var(--color-sdt-surface)",
            padding: isMobile ? 16 : 24,
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
                    gridTemplateColumns: isMobile ? "18px 60px minmax(0,1fr) auto" : "18px 70px minmax(0,1fr) auto auto",
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
                    <SourceMarker sourceId="riksdagen" />
                  </span>
                  {/* pill */}
                  <Pill tone={pillTone(v.status)}>{v.status}</Pill>
                  {/* margin */}
                  {!isMobile && (
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
                  )}
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
          <SectionSource sourceIds={["riksdagen"]} />
        </div>

        {/* Right card — AGENDA */}
        <div
          style={{
            background: "var(--color-sdt-surface)",
            padding: isMobile ? 16 : 24,
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
          <AgendaList items={agenda} />
          <SectionSource sourceIds={["derived-agenda"]} />
        </div>
      </div>

      {/* ── MYNDIGHETER ──────────────────────────────────────────────── */}
      {authorities && authorities.length > 0 && (() => {
        const totalMdkr = authorities.reduce((s, a) => s + a.expenditureMdkr, 0);
        const donutSegments = authorities.map((a, i) => ({
          color: BUDGET_COLORS[i % BUDGET_COLORS.length],
          value: a.expenditureMdkr,
          name: a.name,
        }));
        return (
          <div
            style={{
              border: "1px solid var(--color-border)",
              borderTop: "none",
              margin: isMobile ? "0 14px 32px" : "0 32px 32px",
              background: "var(--color-sdt-surface)",
              padding: isMobile ? "16px" : "24px 28px",
            }}
          >
            {/* Header */}
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.15em",
                color: "var(--color-fg-muted)",
                marginBottom: 24,
              }}
            >
              MYNDIGHETER · STATLIGA DRIFTKOSTNADER {authorities[0].year}
            </div>

            {/* Two-column: donut left, list right */}
            <div style={{ display: "flex", gap: 40, alignItems: isMobile ? "stretch" : "flex-start", flexDirection: isMobile ? "column" : "row" }}>

              {/* Left: donut + legend */}
              <div style={{ flexShrink: 0, display: "flex", flexDirection: "column", alignItems: "center", gap: 16 }}>
                <Donut
                  segments={donutSegments}
                  size={isMobile ? 160 : 200}
                  thickness={isMobile ? 24 : 32}
                  label={`${Math.round(totalMdkr)} mdkr`}
                  sublabel="totalt"
                />
                <div style={{ display: isMobile ? "none" : "flex", flexDirection: "column", gap: 5, width: 200 }}>
                  {authorities.map((a, i) => (
                    <div key={i} style={{ display: "flex", alignItems: "center", gap: 6 }}>
                      <div style={{ width: 8, height: 8, borderRadius: "50%", background: BUDGET_COLORS[i % BUDGET_COLORS.length], flexShrink: 0 }} />
                      <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
                        {a.name}
                      </span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Right: expandable agency rows */}
              <div style={{ flex: 1, minWidth: 0 }}>
                {authorities.map((a, i) => (
                  <AuthorityRow
                    key={a.name}
                    authority={a}
                    color={BUDGET_COLORS[i % BUDGET_COLORS.length]}
                    isOpen={openAuthority === i}
                    onToggle={() => setOpenAuthority(openAuthority === i ? null : i)}
                  />
                ))}
              </div>
            </div>

            {/* Footer with source link */}
            <SectionSource sourceIds={["statskontoret-arsutfall"]} />
          </div>
        );
      })()}
    </div>
  );
}
