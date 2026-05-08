import { useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import { Hemicycle, Donut, HBars, Pill, Trend, GoalBadge } from "@/components/charts";
import { AgendaList } from "@/components/AgendaList";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { useKommun, useKommunList } from "@/hooks/useDemocracy";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { BottomSheet } from "@/components/BottomSheet";
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
  const isMobile = useMediaQuery("(max-width: 640px)");
  const [sheetOpen, setSheetOpen] = useState(false);

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
      <div style={{ padding: isMobile ? "18px 14px 0" : "40px 32px 0" }}>
        <div style={{ display: "flex", alignItems: "baseline", gap: 20 }}>
          {/* Roman numeral */}
          <span
            style={{
              fontFamily: "var(--font-serif)",
              fontSize: isMobile ? 56 : 88,
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
                fontSize: isMobile ? 28 : 52,
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
              {/* Desktop: invisible-select overlay */}
              {!isMobile && (
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
              )}
              {/* Mobile: opens BottomSheet */}
              {isMobile && (
                <button
                  onClick={() => setSheetOpen(true)}
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 11,
                    color: "var(--color-accent)",
                    background: "none",
                    border: "none",
                    cursor: "pointer",
                    letterSpacing: "0.08em",
                    padding: 0,
                  }}
                >
                  ↓ BYT KOMMUN
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* ── KPI strip ──────────────────────────────────────────────────── */}
      {kpis.length > 0 && (
        <div style={{ marginBottom: 0 }}>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: isMobile ? "repeat(2, 1fr)" : "repeat(3, 1fr)",
              margin: isMobile ? "14px 14px 0" : "22px 32px 0",
              gap: 1,
              background: "var(--color-border)",
              border: "1px solid var(--color-border)",
            }}
          >
            {kpis.map((kpi) => (
              <div
                key={kpi.label}
                style={{
                  background: "var(--color-sdt-surface)",
                  padding: isMobile ? "14px" : "20px 24px",
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
                    fontSize: isMobile ? 24 : 30,
                    fontWeight: 400,
                    color: "var(--color-fg)",
                    lineHeight: 1,
                    marginBottom: 8,
                    display: "flex",
                    alignItems: "baseline",
                    gap: 4,
                  }}
                >
                  {kpi.value}
                  <SourceMarker sourceId="kolada" />
                </div>
                <div style={{ marginBottom: 12 }}>
                  <Trend trend={kpi.trend} delta={kpi.delta} worseHigher={kpi.worseHigher} />
                </div>
                <GoalBadge
                  raw={kpi.raw}
                  target={kpi.target}
                  worseHigher={kpi.worseHigher}
                  unit={kpi.unit}
                  note={kpi.note}
                  sourceUrl={kpi.sourceUrl}
                />
                {kpi.description && (
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
                    {kpi.description}
                  </div>
                )}
              </div>
            ))}
          </div>
          <div style={{ padding: isMobile ? "0 14px" : "0 32px", paddingTop: 0 }}>
          </div>
        </div>
      )}

      {/* ── Main 2-column grid ─────────────────────────────────────────── */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: isMobile ? "1fr" : "1fr 1fr",
          gap: 1,
          background: "var(--color-border)",
          border: "1px solid var(--color-border)",
          margin: isMobile ? "14px 14px 0" : "22px 32px 0",
        }}
      >
        {/* Left — MANDAT */}
        <div style={{ background: "var(--color-sdt-surface)", padding: isMobile ? 16 : "24px 28px" }}>
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

          {totalSeats === 0 ? (
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
              Mandatdata saknas för denna kommun.
            </div>
          ) : (
            <Hemicycle groups={allParties.map((p) => ({ color: p.color, count: p.seats }))} width={440} height={150} />
          )}

          {/* Party legend — split by governing / opposition */}
          <div style={{ display: totalSeats === 0 ? "none" : "flex", flexDirection: "column", gap: 8, marginTop: 16 }}>
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
                    minWidth: isMobile ? 48 : 64,
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
          {totalSeats > 0 && (
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
                  display: "inline-flex",
                  alignItems: "center",
                  gap: 4,
                }}
              >
                MAJORITET {govSeats}/{totalSeats}
                <SourceMarker sourceId="scb-kfmandat" />
              </span>
            </div>
          )}
          <div style={{ marginTop: 16 }}>
          </div>
        </div>

        {/* Right — BUDGET */}
        <div style={{ background: "var(--color-sdt-surface)", padding: isMobile ? 16 : "24px 28px", display: "flex", flexDirection: "column" }}>
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
              gap: 4,
            }}
          >
            <span>
              BUDGET {data.budget.year}
              {data.budget.total ? ` · ${data.budget.total}` : ""}
            </span>
            <SourceMarker sourceId="kolada-spending" />
          </div>

          <div style={{ flex: 1 }}>
            {budgetAreas.length === 0 ? (
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
                Budgetdata saknas för denna kommun.
                <div style={{ fontSize: 9, marginTop: 6, opacity: 0.7 }}>
                  Källa: Kolada — endast utvalda kommuner stöds f.n.
                </div>
              </div>
            ) : (
              <div style={{ display: "flex", gap: 20, alignItems: isMobile ? "center" : "flex-start", flexDirection: isMobile ? "column" : "row" }}>
                <Donut
                  segments={donutSegments}
                  size={isMobile ? 140 : 160}
                  thickness={isMobile ? 16 : 24}
                  label={data.budget.total}
                  sublabel={data.budget.year}
                />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <HBars items={hbarsItems} unit="" height={5} gap={9} />
                </div>
              </div>
            )}
          </div>
          <div style={{ marginTop: 16 }}>
          </div>
        </div>
      </div>

      {/* ── Bottom grid ────────────────────────────────────────────────── */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: isMobile ? "1fr" : "1.3fr 1fr",
          gap: 1,
          background: "var(--color-border)",
          border: "1px solid var(--color-border)",
          borderTop: "none",
          margin: isMobile ? "1px 14px 28px" : "1px 32px 28px",
        }}
      >
        {/* Left — PULS */}
        <div style={{ background: "var(--color-sdt-surface)", padding: isMobile ? 16 : "24px 28px", display: "flex", flexDirection: "column" }}>
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

          <div style={{ display: "flex", flexDirection: "column", gap: 0, flex: 1 }}>
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
                    <div style={{ fontSize: 13, color: "var(--color-fg)", marginBottom: 4, lineHeight: 1.35, display: "flex", alignItems: "baseline", gap: 4 }}>
                      {vote.title}
                      <SourceMarker sourceId="riksdagen" />
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
          <div style={{ marginTop: 16 }}>
          </div>
        </div>

        {/* Right — AGENDA */}
        <div style={{ background: "var(--color-sdt-surface)", padding: isMobile ? 16 : "24px 28px", display: "flex", flexDirection: "column" }}>
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

          <div style={{ flex: 1 }}>
            <AgendaList items={data.agenda} />
          </div>
          <div style={{ marginTop: 16 }}>
          </div>
        </div>
      </div>

      {/* Mobile floating BYT KOMMUN button + BottomSheet */}
      {isMobile && (
        <>
          <div style={{ position: "fixed", bottom: 72, left: 14, right: 14, zIndex: 30 }}>
            <button
              onClick={() => setSheetOpen(true)}
              style={{
                width: "100%",
                height: 48,
                borderRadius: 999,
                border: "1px solid var(--color-border)",
                background: "var(--color-sdt-surface)",
                color: "var(--color-fg)",
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                letterSpacing: "1px",
                textTransform: "uppercase",
                cursor: "pointer",
              }}
            >
              BYT KOMMUN ↓
            </button>
          </div>
          <BottomSheet open={sheetOpen} onClose={() => setSheetOpen(false)} title="Välj kommun">
            {(allKommuner ?? []).map((k, i) => (
              <button
                key={k.code}
                onClick={() => {
                  setSheetOpen(false);
                  navigate(`/kommun/${k.code}`);
                }}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 12,
                  width: "100%",
                  height: 56,
                  padding: "0 20px",
                  background: "none",
                  border: "none",
                  borderBottom: i < (allKommuner?.length ?? 0) - 1 ? "1px solid var(--color-border)" : "none",
                  cursor: "pointer",
                  textAlign: "left",
                }}
              >
                <span style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-accent)", minWidth: 32, letterSpacing: "0.1em" }}>
                  {k.code.toUpperCase().slice(0, 3)}
                </span>
                <span style={{ fontFamily: "var(--font-serif)", fontSize: 16, color: "var(--color-fg)", flex: 1 }}>
                  {k.name}
                </span>
              </button>
            ))}
          </BottomSheet>
        </>
      )}
    </div>
  );
}
