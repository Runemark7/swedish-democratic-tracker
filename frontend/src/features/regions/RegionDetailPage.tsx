import { useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import { useRegion, useRegionList } from "@/hooks/useDemocracy";
import { Hemicycle, Donut, HBars, Pill, Trend, TargetBar } from "@/components/charts";
import { AgendaList } from "@/components/AgendaList";
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

// ── RegionDetailPage ──────────────────────────────────────────────────────────

export function RegionDetailPage() {
  const { code } = useParams<{ code: string }>();
  const navigate = useNavigate();
  const [selectOpen, setSelectOpen] = useState(false);

  const { data, isLoading } = useRegion(code ?? "");
  const { data: regionList } = useRegionList();

  if (isLoading || !data) return <Skeleton />;

  const { title, subtitle, ruling, liveVotes, budget, agenda, kpis } = data;

  // Build hemicycle groups: opposition left → support → governing right
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

  // Donut label: first word + rest split
  const totalParts = budget.total.split(" ");
  const donutLabel = totalParts[0];
  const donutSublabel = totalParts.slice(1).join(" ");

  return (
    <div className="sdt-page">
      {/* ── Hero ─────────────────────────────────────────────────────── */}
      <div
        style={{
          display: "flex",
          alignItems: "flex-end",
          gap: 20,
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
          II
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
            KAMMARE TVÅ
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
            {title}
          </div>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: 12,
              fontSize: 14,
              color: "var(--color-fg-muted)",
            }}
          >
            <span>{subtitle}</span>
            <button
              onClick={() => setSelectOpen((v) => !v)}
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                color: "var(--color-accent)",
                background: "none",
                border: "none",
                cursor: "pointer",
                padding: 0,
                letterSpacing: "0.08em",
              }}
            >
              ↓ BYT REGION
            </button>
            {selectOpen && regionList && (
              <select
                defaultValue={code}
                onChange={(e) => {
                  setSelectOpen(false);
                  navigate(`/region/${e.target.value}`);
                }}
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                  background: "var(--color-sdt-surface)",
                  color: "var(--color-fg)",
                  border: "1px solid var(--color-border)",
                  borderRadius: 3,
                  padding: "3px 6px",
                  cursor: "pointer",
                }}
              >
                {regionList.map((r) => (
                  <option key={r.code} value={r.code}>
                    {r.name}
                  </option>
                ))}
              </select>
            )}
          </div>
        </div>
      </div>

      {/* ── KPI strip ────────────────────────────────────────────────── */}
      {kpis && kpis.length > 0 && (
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(3, 1fr)",
            gap: 1,
            border: "1px solid var(--color-border)",
            background: "var(--color-border)",
            margin: "22px 32px 0",
          }}
        >
          {kpis.map((k) => (
            <div
              key={k.label}
              style={{
                background: "var(--color-bg)",
                padding: "16px 20px",
              }}
            >
              <div
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 10,
                  letterSpacing: "1.5px",
                  color: "var(--color-fg-muted)",
                  marginBottom: 8,
                }}
              >
                {k.label}
              </div>
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 8,
                  marginBottom: 12,
                }}
              >
                <span
                  style={{
                    fontFamily: "var(--font-serif)",
                    fontSize: 30,
                    fontVariantNumeric: "tabular-nums",
                    color: "var(--color-fg)",
                    lineHeight: 1,
                  }}
                >
                  {k.value}
                </span>
                <Trend trend={k.trend} delta={k.delta} />
              </div>
              <TargetBar
                value={k.raw}
                target={k.target}
                worseHigher={k.worseHigher}
                unit={k.unit}
              />
            </div>
          ))}
        </div>
      )}

      {/* ── Main grid ────────────────────────────────────────────────── */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr",
          gap: 1,
          border: "1px solid var(--color-border)",
          background: "var(--color-border)",
          margin: "22px 32px 0",
        }}
      >
        {/* Left card — MANDAT */}
        <div style={{ background: "var(--color-sdt-surface)", padding: 24 }}>
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

          {/* Party legend — split by governing / opposition */}
          <div style={{ display: "flex", flexDirection: "column", gap: 8, marginTop: 12 }}>
            {[
              { label: "STYRE", parties: [...ruling.parties, ...(ruling.support ?? [])] },
              { label: "OPPOSITION", parties: ruling.opposition },
            ].map(({ label, parties }) => (
              <div key={label} style={{ display: "flex", alignItems: "center", gap: 10, flexWrap: "wrap" }}>
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 8,
                    letterSpacing: "0.15em",
                    color: label === "STYRE" ? "var(--color-accent)" : "var(--color-fg-muted)",
                    opacity: label === "STYRE" ? 1 : 0.7,
                    minWidth: 64,
                    flexShrink: 0,
                  }}
                >
                  {label}
                </span>
                <div style={{ display: "flex", flexWrap: "wrap", gap: "4px 12px" }}>
                  {parties.map((p) => (
                    <span
                      key={p.short}
                      style={{
                        display: "inline-flex",
                        alignItems: "center",
                        gap: 4,
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
              </div>
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
        <div style={{ background: "var(--color-sdt-surface)", padding: 24 }}>
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
              label={donutLabel}
              sublabel={donutSublabel}
            />
            <div style={{ flex: 1, minWidth: 0, paddingTop: 4 }}>
              <HBars
                items={budgetSegments.slice(0, 5)}
                height={6}
                gap={10}
                unit=" mdkr"
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
        <div style={{ background: "var(--color-sdt-surface)", padding: 24 }}>
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
            <div>
              <div
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 10,
                  letterSpacing: "0.15em",
                  color: "var(--color-fg-muted)",
                }}
              >
                RIKSDAG · RELEVANTA BESLUT
              </div>
              <div
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 9,
                  letterSpacing: "0.08em",
                  color: "var(--color-fg-muted)",
                  opacity: 0.6,
                  marginTop: 2,
                }}
              >
                Riksdagsbeslut som berör regional nivå
              </div>
            </div>
          </div>

          <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
            {liveVotes.map((v, i) => {
              const href = beslutHref(v);
              const row = (
                <div
                  style={{
                    display: "grid",
                    gridTemplateColumns: "18px 70px 1fr auto",
                    gap: 12,
                    alignItems: "center",
                    cursor: href ? "pointer" : "default",
                  }}
                >
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
                  <span style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg-muted)" }}>
                    {v.time}
                  </span>
                  <span
                    style={{ fontSize: 13, color: "var(--color-fg)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}
                    title={v.title}
                  >
                    {v.title}
                    {v.tag && (
                      <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", marginLeft: 6, verticalAlign: "middle" }}>
                        [{v.tag}]
                      </span>
                    )}
                  </span>
                  <Pill tone={pillTone(v.status)}>{v.status}</Pill>
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
        <div style={{ background: "var(--color-sdt-surface)", padding: 24 }}>
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
        </div>
      </div>
    </div>
  );
}
