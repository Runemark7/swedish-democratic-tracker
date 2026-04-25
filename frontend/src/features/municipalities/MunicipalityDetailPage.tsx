import { Link, useParams, useNavigate } from "react-router-dom";
import { Hemicycle, Donut, HBars, Pill, Trend, TargetBar } from "@/components/charts";
import { useKommun, useKommunList } from "@/hooks/useDemocracy";
import type { LiveVote } from "@/types/democracy";

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

function beslutHref(v: LiveVote): string | null {
  if (!v.beteckning) return null;
  const p = new URLSearchParams({ title: v.title, status: v.status, tag: v.tag, time: v.time });
  return `/beslut/${encodeURIComponent(v.beteckning)}?${p}`;
}

function pillTone(status: LiveVote["status"]): "pass" | "fail" | "pending" | "neutral" {
  if (status === "Bifall") return "pass";
  if (status === "Avslag") return "fail";
  if (status === "Återremiss") return "pending";
  return "neutral";
}

export function MunicipalityDetailPage() {
  const { code } = useParams<{ code: string }>();
  const navigate = useNavigate();

  const { data, isLoading } = useKommun(code ?? "");
  const { data: allKommuner } = useKommunList();

  if (isLoading || !data) {
    return (
      <div
        style={{
          background: "var(--color-bg)",
          color: "var(--color-fg)",
          minHeight: "100vh",
          padding: "40px 32px",
        }}
      >
        {/* Hero skeleton */}
        <div style={{ display: "flex", alignItems: "baseline", gap: 20, marginBottom: 22 }}>
          <div
            style={{
              width: 60,
              height: 75,
              borderRadius: 4,
              background: "var(--color-track)",
              animation: "d3pulse 2s ease infinite",
            }}
          />
          <div>
            <div
              style={{
                width: 80,
                height: 10,
                borderRadius: 4,
                background: "var(--color-track)",
                marginBottom: 8,
                animation: "d3pulse 2s ease infinite",
              }}
            />
            <div
              style={{
                width: 300,
                height: 44,
                borderRadius: 4,
                background: "var(--color-track)",
                animation: "d3pulse 2s ease infinite",
              }}
            />
          </div>
        </div>
        {/* KPI strip skeleton */}
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(3, 1fr)",
            margin: "22px 0 0",
            gap: 1,
            background: "var(--color-border)",
            border: "1px solid var(--color-border)",
          }}
        >
          {[0, 1, 2].map((i) => (
            <div
              key={i}
              style={{
                background: "var(--color-sdt-surface)",
                padding: "20px 24px",
                height: 96,
                animation: "d3pulse 2s ease infinite",
              }}
            />
          ))}
        </div>
      </div>
    );
  }

  const governing = data.ruling.parties;
  const opposition = data.ruling.opposition;
  const allParties = [...opposition, ...governing];
  const totalSeats = allParties.reduce((s, p) => s + p.seats, 0);
  const govSeats = governing.reduce((s, p) => s + p.seats, 0);

  const kpis = data.kpis ?? [];
  const budgetAreas = data.budget.areas;
  const donutSegments = budgetAreas.map((a, i) => ({
    color: BUDGET_COLORS[i % BUDGET_COLORS.length],
    value: a.pct,
    name: a.name,
  }));
  const hbarsItems = budgetAreas.map((a, i) => ({
    name: a.name,
    value: a.pct,
    pct: a.pct,
    amount: `${a.value.toFixed(2)} mdkr`,
    color: BUDGET_COLORS[i % BUDGET_COLORS.length],
  }));

  return (
    <div
      style={{
        background: "var(--color-bg)",
        color: "var(--color-fg)",
        minHeight: "100vh",
        fontFamily: "var(--font-body)",
      }}
    >
      {/* ── Hero ──────────────────────────────────────────────────────── */}
      <div style={{ padding: "40px 32px 0" }}>
        <div style={{ display: "flex", alignItems: "baseline", gap: 20 }}>
          {/* Roman numeral */}
          <span
            style={{
              fontFamily: "var(--font-serif)",
              fontSize: 88,
              fontWeight: 400,
              letterSpacing: "-3px",
              lineHeight: 0.85,
              color: "var(--color-accent-2)",
              fontStyle: "italic",
              userSelect: "none",
            }}
          >
            III
          </span>

          {/* Title block */}
          <div>
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "2px",
                color: "var(--color-fg-muted)",
                textTransform: "uppercase",
                marginBottom: 4,
              }}
            >
              KAMMARE TRE
            </div>
            <h1
              style={{
                fontFamily: "var(--font-serif)",
                fontSize: 52,
                fontWeight: 400,
                letterSpacing: "-1.5px",
                lineHeight: 1,
                margin: 0,
                color: "var(--color-fg)",
              }}
            >
              {data.title}
            </h1>
            {/* Subtitle row */}
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: 12,
                marginTop: 8,
                flexWrap: "wrap",
              }}
            >
              <span
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                  color: "var(--color-fg-muted)",
                }}
              >
                {data.subtitle}
              </span>
              {data.population && (
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 11,
                    color: "var(--color-fg-muted)",
                  }}
                >
                  · {data.population}
                </span>
              )}
              {/* BYT KOMMUN inline select */}
              <span style={{ position: "relative", display: "inline-flex", alignItems: "center" }}>
                <select
                  value={code ?? ""}
                  onChange={(e) => {
                    if (e.target.value) navigate(`/kommun/${e.target.value}`);
                  }}
                  style={{
                    position: "absolute",
                    inset: 0,
                    opacity: 0,
                    cursor: "pointer",
                    width: "100%",
                  }}
                  aria-label="Byt kommun"
                >
                  {(allKommuner ?? []).map((k) => (
                    <option key={k.code} value={k.code}>
                      {k.name}
                    </option>
                  ))}
                </select>
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 11,
                    color: "var(--color-accent)",
                    letterSpacing: "1px",
                    cursor: "pointer",
                    textDecoration: "underline",
                    textUnderlineOffset: 3,
                  }}
                >
                  ↓ BYT KOMMUN
                </span>
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* ── KPI strip ──────────────────────────────────────────────────── */}
      {kpis.length > 0 && (
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(3, 1fr)",
            margin: "22px 32px 0",
            gap: 1,
            background: "var(--color-border)",
            border: "1px solid var(--color-border)",
          }}
        >
          {kpis.slice(0, 3).map((kpi) => (
            <div
              key={kpi.label}
              style={{
                background: "var(--color-sdt-surface)",
                padding: "20px 24px",
              }}
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
                {kpi.label}
              </div>
              <div
                style={{
                  fontFamily: "var(--font-serif)",
                  fontSize: 30,
                  fontWeight: 400,
                  color: "var(--color-fg)",
                  lineHeight: 1,
                  marginBottom: 8,
                }}
              >
                {kpi.value}
              </div>
              <div style={{ marginBottom: 12 }}>
                <Trend trend={kpi.trend} delta={kpi.delta} />
              </div>
              <TargetBar
                value={kpi.raw}
                target={kpi.target}
                worseHigher={kpi.worseHigher}
                unit={kpi.unit}
                height={28}
              />
            </div>
          ))}
        </div>
      )}

      {/* ── Main 2-column grid ─────────────────────────────────────────── */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr",
          gap: 1,
          background: "var(--color-border)",
          border: "1px solid var(--color-border)",
          margin: "22px 32px 0",
        }}
      >
        {/* Left — MANDAT */}
        <div style={{ background: "var(--color-sdt-surface)", padding: "24px 28px" }}>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              letterSpacing: "2px",
              color: "var(--color-fg-muted)",
              textTransform: "uppercase",
              marginBottom: 16,
            }}
          >
            MANDAT · KAMMARENS SAMMANSÄTTNING
          </div>

          <Hemicycle groups={allParties.map((p) => ({ color: p.color, count: p.seats }))} width={440} height={150} />

          {/* Party legend — split by governing / opposition */}
          <div style={{ display: "flex", flexDirection: "column", gap: 8, marginTop: 16 }}>
            {[
              { label: "STYRE", parties: governing },
              { label: "OPPOSITION", parties: opposition },
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
                    <div key={p.short} style={{ display: "flex", alignItems: "center", gap: 5 }}>
                      <span
                        style={{
                          width: 8,
                          height: 8,
                          borderRadius: "50%",
                          background: p.color,
                          flexShrink: 0,
                        }}
                      />
                      <span
                        style={{
                          fontFamily: "var(--font-mono)",
                          fontSize: 10,
                          color: "var(--color-fg-muted)",
                        }}
                      >
                        <span style={{ color: "var(--color-fg)", fontWeight: 600 }}>{p.short}</span>{" "}
                        {p.seats}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>

          {/* Coalition name + majority */}
          <div style={{ marginTop: 14 }}>
            <span
              style={{
                fontFamily: "var(--font-serif)",
                fontStyle: "italic",
                fontSize: 14,
                color: "var(--color-fg-muted)",
              }}
            >
              {data.ruling.type}
            </span>
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                color: "var(--color-fg-muted)",
                marginLeft: 12,
              }}
            >
              MAJORITET {govSeats}/{totalSeats}
            </span>
          </div>
        </div>

        {/* Right — BUDGET */}
        <div style={{ background: "var(--color-sdt-surface)", padding: "24px 28px" }}>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              letterSpacing: "2px",
              color: "var(--color-fg-muted)",
              textTransform: "uppercase",
              marginBottom: 16,
            }}
          >
            BUDGET {data.budget.year} · {data.budget.total}
          </div>

          <div style={{ display: "flex", gap: 20, alignItems: "flex-start" }}>
            <Donut
              segments={donutSegments}
              size={160}
              thickness={24}
              label={data.budget.total}
              sublabel={data.budget.year}
            />
            <div style={{ flex: 1, minWidth: 0 }}>
              <HBars items={hbarsItems} unit="" height={5} gap={9} />
            </div>
          </div>
        </div>
      </div>

      {/* ── Bottom grid ────────────────────────────────────────────────── */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1.3fr 1fr",
          gap: 1,
          background: "var(--color-border)",
          border: "1px solid var(--color-border)",
          borderTop: "none",
          margin: "1px 32px 28px",
        }}
      >
        {/* Left — PULS */}
        <div style={{ background: "var(--color-sdt-surface)", padding: "24px 28px" }}>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              letterSpacing: "2px",
              color: "var(--color-fg-muted)",
              textTransform: "uppercase",
              marginBottom: 16,
              display: "flex",
              alignItems: "center",
              gap: 8,
            }}
          >
            <span
              className="animate-pulse-dot"
              style={{
                width: 6,
                height: 6,
                borderRadius: "50%",
                background: "var(--color-pulse)",
                display: "inline-block",
                flexShrink: 0,
              }}
            />
            <div>
              <div>RIKSDAG · RELEVANTA BESLUT</div>
              <div
                style={{
                  fontSize: 9,
                  letterSpacing: "0.08em",
                  opacity: 0.6,
                  marginTop: 2,
                  textTransform: "none",
                }}
              >
                Riksdagsbeslut som berör kommunal nivå
              </div>
            </div>
          </div>

          <div style={{ display: "flex", flexDirection: "column", gap: 0 }}>
            {data.liveVotes.map((vote, i) => {
              const href = beslutHref(vote);
              const row = (
                <div
                  style={{
                    display: "flex",
                    alignItems: "flex-start",
                    gap: 12,
                    padding: "10px 0",
                    borderBottom: i < data.liveVotes.length - 1 ? "1px solid var(--color-border)" : "none",
                    cursor: href ? "pointer" : "default",
                  }}
                >
                  <span
                    style={{
                      fontFamily: "var(--font-mono)",
                      fontSize: 10,
                      color: "var(--color-fg-muted)",
                      whiteSpace: "nowrap",
                      paddingTop: 2,
                      minWidth: 64,
                    }}
                  >
                    {vote.time}
                  </span>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 13, color: "var(--color-fg)", marginBottom: 4, lineHeight: 1.35 }}>
                      {vote.title}
                    </div>
                    <div style={{ display: "flex", gap: 6, alignItems: "center" }}>
                      <span
                        style={{
                          fontFamily: "var(--font-mono)",
                          fontSize: 10,
                          color: "var(--color-fg-muted)",
                          background: "var(--color-track)",
                          borderRadius: 2,
                          padding: "1px 6px",
                        }}
                      >
                        {vote.tag}
                      </span>
                      <Pill tone={pillTone(vote.status)}>{vote.status}</Pill>
                      {vote.margin && (
                        <span style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg-muted)" }}>
                          {vote.margin}
                        </span>
                      )}
                    </div>
                  </div>
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

        {/* Right — AGENDA */}
        <div style={{ background: "var(--color-sdt-surface)", padding: "24px 28px" }}>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              letterSpacing: "2px",
              color: "var(--color-fg-muted)",
              textTransform: "uppercase",
              marginBottom: 16,
            }}
          >
            AGENDA · PRIORITERINGAR
          </div>

          <ol style={{ listStyle: "none", margin: 0, padding: 0 }}>
            {data.agenda.map((item, i) => (
              <li
                key={i}
                style={{
                  display: "flex",
                  alignItems: "baseline",
                  gap: 14,
                  padding: "10px 0",
                  borderBottom:
                    i < data.agenda.length - 1
                      ? "1px solid var(--color-border)"
                      : "none",
                }}
              >
                <span
                  style={{
                    fontFamily: "var(--font-serif)",
                    fontStyle: "italic",
                    fontSize: 22,
                    fontWeight: 400,
                    color: "var(--color-accent-2)",
                    lineHeight: 1,
                    minWidth: 20,
                    textAlign: "right",
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
