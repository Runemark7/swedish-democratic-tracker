import { useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import { useRegion, useRegionList, useKommunList, useRegionBudgetHistory } from "@/hooks/useDemocracy";
import { Hemicycle, Pill, Trend, GoalBadge } from "@/components/charts";
import { DeltaIndicator } from "@/features/budget/components/DeltaIndicator";
import { AgendaList } from "@/components/AgendaList";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { BottomSheet } from "@/components/BottomSheet";
import { SwedenKommunMap } from "@/features/municipalities/components/SwedenKommunMap";
import { SourceMarker } from "@/components/sources/SourceMarker";
import type { LiveVote, Party } from "@/types/democracy";
import type { RegionBudgetSnapshot } from "@/shared/types";

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
  const isMobile = useMediaQuery("(max-width: 640px)");
  const [sheetOpen, setSheetOpen] = useState(false);
  const [budgetTab, setBudgetTab] = useState<number | null>(null);

  const { data, isLoading } = useRegion(code ?? "");
  const { data: regionList } = useRegionList();
  const { data: regionKommuner } = useKommunList(code);
  const { data: budgetHistory = [] } = useRegionBudgetHistory(code ?? "");

  if (isLoading || !data) return <Skeleton />;

  const { title, subtitle, ruling, liveVotes, agenda, kpis } = data;

  // ── Budget history: group snapshots by year ────────────────────────────────
  const historyByYear = new Map<number, RegionBudgetSnapshot[]>();
  for (const snap of budgetHistory) {
    const list = historyByYear.get(snap.year) ?? [];
    list.push(snap);
    historyByYear.set(snap.year, list);
  }
  const sortedYears = Array.from(historyByYear.keys()).sort((a, b) => b - a);

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

  // Active budget tab: use state if set, otherwise most recent year with data
  const activeBudgetYear = budgetTab ?? sortedYears[0] ?? null;

  return (
    <>
      <div className="sdt-page">
        {/* ── Hero ─────────────────────────────────────────────────────── */}
        <div
          style={{
            display: "flex",
            alignItems: "flex-end",
            gap: 20,
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
                fontSize: isMobile ? 28 : 52,
                fontWeight: 400,
                letterSpacing: "-1.5px",
                color: "var(--color-fg)",
                lineHeight: 1,
                marginBottom: 8,
              }}
            >
              {title}
            </div>
            <div style={{ display: "flex", alignItems: "center", gap: 12, fontSize: 14, color: "var(--color-fg-muted)", flexWrap: "wrap" }}>
              <span>{subtitle}</span>
              {!isMobile && (
                <>
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
                </>
              )}
            </div>
          </div>
        </div>

        {/* ── KOMMUNER (clickable map) ─────────────────────────────────── */}
        {code && (
          <div
            style={{
              border: "1px solid var(--color-border)",
              margin: isMobile ? "14px 14px 0" : "22px 32px 0",
              background: "var(--color-sdt-surface)",
              padding: isMobile ? "12px 14px" : "16px 20px",
            }}
          >
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.15em",
                color: "var(--color-fg-muted)",
                marginBottom: 10,
              }}
            >
              KOMMUNER · KLICKA FÖR DETALJ
            </div>
            <div style={{ display: "flex", justifyContent: "center" }}>
              <div style={{ width: "100%", maxWidth: isMobile ? 200 : 260 }}>
                <SwedenKommunMap
                  municipalities={regionKommuner ?? []}
                  regionCode={code}
                />
              </div>
            </div>
            {regionKommuner && regionKommuner.length > 0 && (
              <div
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 9,
                  color: "var(--color-fg-muted)",
                  marginTop: 8,
                  textAlign: "center",
                  letterSpacing: "0.1em",
                  textTransform: "uppercase",
                }}
              >
                {regionKommuner.length} kommuner
              </div>
            )}
          </div>
        )}

        {/* ── KPI strip ────────────────────────────────────────────────── */}
        {kpis && kpis.length > 0 && (
          <>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: isMobile ? "repeat(2, 1fr)" : "repeat(3, 1fr)",
                gap: 1,
                border: "1px solid var(--color-border)",
                background: "var(--color-border)",
                margin: isMobile ? "14px 14px 0" : "22px 32px 0",
              }}
            >
              {kpis.map((k) => (
              <div
                key={k.label}
                style={{
                  background: "var(--color-bg)",
                  padding: isMobile ? "14px" : "16px 20px",
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
                      fontSize: isMobile ? 24 : 30,
                      fontVariantNumeric: "tabular-nums",
                      color: "var(--color-fg)",
                      lineHeight: 1,
                    }}
                  >
                    {k.value}
                  </span>
                  <SourceMarker sourceId="kolada" />
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
          </>
        )}

        {/* ── Main grid ────────────────────────────────────────────────── */}
        <div
          style={{
            display: "grid",
            gridTemplateColumns: isMobile ? "1fr" : "1fr 1fr",
            gap: 1,
            border: "1px solid var(--color-border)",
            background: "var(--color-border)",
            margin: isMobile ? "14px 14px 0" : "22px 32px 0",
          }}
        >
          {/* Left card — MANDAT */}
          <div style={{ background: "var(--color-sdt-surface)", padding: isMobile ? 16 : 24 }}>
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
                Mandatdata saknas för denna region.
              </div>
            ) : (
              <Hemicycle groups={hemicycleGroups} width={440} height={150} />
            )}

            {/* Party legend — split by governing / opposition */}
            <div style={{ display: totalSeats === 0 ? "none" : "flex", flexDirection: "column", gap: 8, marginTop: 12 }}>
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
                      minWidth: isMobile ? 48 : 64,
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
            {totalSeats > 0 && (
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
                <SourceMarker sourceId="scb-ltmandat" />
              </div>
            )}
          </div>

          {/* Right card — BUDGET history tab strip */}
          <div style={{ background: "var(--color-sdt-surface)", padding: isMobile ? 16 : 24 }}>
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.15em",
                color: "var(--color-fg-muted)",
                marginBottom: 12,
              }}
            >
              BUDGET · HISTORIK
              <SourceMarker sourceId="scb-kostndrlt" />
            </div>

            {sortedYears.length === 0 ? (
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
                Budgetdata saknas för denna region.
              </div>
            ) : (
              <>
                {/* Year tab strip */}
                <div style={{ display: "flex", gap: 4, flexWrap: "wrap", marginBottom: 16 }}>
                  {sortedYears.map((yr, idx) => {
                    const prevYear = sortedYears[idx + 1] ?? null;
                    const prevSnaps = prevYear != null ? historyByYear.get(prevYear) : null;
                    const curTotal = (historyByYear.get(yr) ?? []).reduce((s, a) => s + a.total_mnkr, 0);
                    const prevTotal = prevSnaps ? prevSnaps.reduce((s, a) => s + a.total_mnkr, 0) : null;
                    const delta = prevTotal != null && prevTotal > 0
                      ? ((curTotal - prevTotal) / prevTotal) * 100
                      : null;
                    const isActive = yr === activeBudgetYear;
                    const tabDeltaColor = delta == null
                      ? "var(--color-fg-muted)"
                      : delta >= 0 ? "#4caf7d" : "#e05c5c";

                    return (
                      <button
                        key={yr}
                        onClick={() => setBudgetTab(yr)}
                        className="px-3 py-1.5 text-xs font-mono font-bold rounded-md transition-all"
                        style={{
                          background: isActive ? "var(--color-primary)" : "var(--color-surface-low)",
                          color: isActive ? "var(--color-on-primary)" : "var(--color-on-surface)",
                        }}
                      >
                        {yr}
                        {delta != null && (
                          <span style={{ marginLeft: 4, fontSize: 10, color: isActive ? "var(--color-on-primary)" : tabDeltaColor }}>
                            {delta >= 0 ? "+" : ""}{delta.toFixed(1)}%
                          </span>
                        )}
                      </button>
                    );
                  })}
                </div>

                {/* Active year — table layout matching /budget */}
                {activeBudgetYear != null && (() => {
                  const areas = historyByYear.get(activeBudgetYear) ?? [];
                  const prevYear = sortedYears[sortedYears.indexOf(activeBudgetYear) + 1] ?? null;
                  const prevAreas = prevYear != null ? (historyByYear.get(prevYear) ?? []) : [];
                  const prevByName = new Map(prevAreas.map(a => [a.area_name, a]));

                  return (
                    <div style={{ borderRadius: 8, overflow: "hidden", border: "1px solid var(--color-surface-high)" }}>
                      {/* Header */}
                      <div style={{
                        display: "grid",
                        gridTemplateColumns: "1fr auto auto",
                        gap: 8,
                        padding: "6px 12px",
                        background: "var(--color-surface-low)",
                        fontFamily: "var(--font-mono)",
                        fontSize: 9,
                        letterSpacing: "0.12em",
                        color: "var(--color-on-surface-variant)",
                        textTransform: "uppercase",
                      }}>
                        <span>Område</span>
                        <span style={{ textAlign: "right", minWidth: 60 }}>{prevYear ?? "—"}</span>
                        <span style={{ textAlign: "right", minWidth: 80 }}>Förändring</span>
                      </div>

                      {areas.map((a, i) => {
                        const prev = prevByName.get(a.area_name);
                        const deltaPct = prev && prev.value_mnkr > 0
                          ? ((a.value_mnkr - prev.value_mnkr) / prev.value_mnkr) * 100
                          : null;
                        return (
                          <Link
                            key={a.area_name}
                            to={`/region/${code}/budget/${encodeURIComponent(a.area_name)}`}
                            style={{
                              display: "grid",
                              gridTemplateColumns: "1fr auto auto",
                              gap: 8,
                              alignItems: "center",
                              padding: "8px 12px",
                              textDecoration: "none",
                              color: "inherit",
                              borderTop: "1px solid var(--color-surface-high)",
                              background: "var(--color-sdt-surface)",
                            }}
                          >
                            <div style={{ display: "flex", alignItems: "center", gap: 8, minWidth: 0 }}>
                              <div style={{
                                width: 10, height: 10, borderRadius: 2, flexShrink: 0,
                                background: BUDGET_COLORS[i % BUDGET_COLORS.length],
                              }} />
                              <span style={{ fontSize: 12, fontWeight: 500, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
                                {a.area_name}
                              </span>
                            </div>
                            <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-on-surface-variant)", textAlign: "right", minWidth: 60 }}>
                              {prev ? `${Math.round(prev.value_mnkr)} mnkr` : "—"}
                            </span>
                            <div style={{ minWidth: 80, display: "flex", justifyContent: "flex-end" }}>
                              {deltaPct != null
                                ? <DeltaIndicator pct={Math.round(deltaPct * 10) / 10} showBar={false} />
                                : <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-on-surface-variant)" }}>—</span>
                              }
                            </div>
                          </Link>
                        );
                      })}

                      {/* Total row */}
                      <div style={{
                        display: "grid",
                        gridTemplateColumns: "1fr auto auto",
                        gap: 8,
                        alignItems: "center",
                        padding: "8px 12px",
                        borderTop: "2px solid var(--color-surface-highest)",
                        background: "var(--color-surface-low)",
                        fontWeight: 700,
                      }}>
                        <span style={{ fontFamily: "var(--font-mono)", fontSize: 11 }}>TOTALT</span>
                        <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-on-surface-variant)", textAlign: "right", minWidth: 60 }}>
                          {prevAreas.length > 0 ? `${Math.round(prevAreas.reduce((s, a) => s + a.value_mnkr, 0))} mnkr` : "—"}
                        </span>
                        <div style={{ minWidth: 80, display: "flex", justifyContent: "flex-end" }}>
                          {(() => {
                            const curTot = areas.reduce((s, a) => s + a.value_mnkr, 0);
                            const prevTot = prevAreas.reduce((s, a) => s + a.value_mnkr, 0);
                            return prevTot > 0
                              ? <DeltaIndicator pct={Math.round(((curTot - prevTot) / prevTot) * 1000) / 10} showBar={false} />
                              : <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-on-surface-variant)" }}>—</span>;
                          })()}
                        </div>
                      </div>
                    </div>
                  );
                })()}
              </>
            )}
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
          <div style={{ background: "var(--color-sdt-surface)", padding: isMobile ? 16 : 24 }}>
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
                      <SourceMarker sourceId="riksdagen" />
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
          <div style={{ background: "var(--color-sdt-surface)", padding: isMobile ? 16 : 24 }}>
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

      {/* Mobile region selector */}
      {isMobile && (
        <>
          <div
            style={{
              position: "fixed",
              bottom: 72,
              left: 14,
              right: 14,
              zIndex: 30,
            }}
          >
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
              BYT REGION ↓
            </button>
          </div>
          <BottomSheet open={sheetOpen} onClose={() => setSheetOpen(false)} title="Välj region">
            {(regionList ?? []).map((r, i) => (
              <button
                key={r.code}
                onClick={() => {
                  setSheetOpen(false);
                  navigate(`/region/${r.code}`);
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
                  borderBottom: i < (regionList?.length ?? 0) - 1 ? "1px solid var(--color-border)" : "none",
                  cursor: "pointer",
                  textAlign: "left",
                }}
              >
                <span style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-accent)", minWidth: 32, letterSpacing: "0.1em" }}>
                  {r.code.toUpperCase().slice(0, 3)}
                </span>
                <span style={{ fontFamily: "var(--font-serif)", fontSize: 16, color: "var(--color-fg)", flex: 1 }}>
                  {r.name}
                </span>
              </button>
            ))}
          </BottomSheet>
        </>
      )}
    </>
  );
}
