import { useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import { useRegion, useRegionList, useKommunList, useRegionBudgetHistory, useRegionKpiRanks } from "@/hooks/useDemocracy";
import { MandateComposition, Pill } from "@/components/charts";
import { BudgetHistorySection } from "@/features/budget/components/BudgetHistorySection";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { BottomSheet } from "@/components/BottomSheet";
import { SwedenKommunMap } from "@/features/municipalities/components/SwedenKommunMap";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { REGION_KPI_META } from "@/features/regions/regionKpiMeta";
import { REGION_BUDGET_DOCS } from "@/features/regions/regionBudgetDocs";
import type { LiveVote } from "@/types/democracy";

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

  const { data, isLoading } = useRegion(code ?? "");
  const { data: regionList } = useRegionList();
  const { data: regionKommuner } = useKommunList(code);
  const { data: budgetHistory = [] } = useRegionBudgetHistory(code ?? "");
  const { data: kpiRanks = [] } = useRegionKpiRanks(code ?? "");
  const rankMap = new Map(kpiRanks.map((r) => [r.kpi, r]));

  if (isLoading || !data) return <Skeleton />;

  const { title, subtitle, ruling, liveVotes, kpis } = data;
  const budgetDoc = code ? REGION_BUDGET_DOCS[code] : undefined;

  const allParties = [
    ...ruling.parties,
    ...(ruling.support ?? []),
    ...ruling.opposition,
  ];
  const totalSeats = allParties.reduce((s, p) => s + p.seats, 0);
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
                    marginBottom: 8,
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
                </div>
                {(() => {
                  const rank = rankMap.get(k.kpiId);
                  if (!rank) return null;
                  const meta = REGION_KPI_META[k.kpiId];
                  const unit = meta?.unit ?? k.unit;
                  const diff = k.raw - rank.mean;
                  return (
                    <Link
                      to={`/region/${code}/kpi/${k.kpiId}`}
                      style={{ textDecoration: "none", display: "inline-flex", alignItems: "baseline", gap: 7, marginBottom: 10, flexWrap: "wrap" }}
                    >
                      <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "#b91c1c", fontWeight: 700 }}>
                        #{rank.rank} AV {rank.total}
                      </span>
                      <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)" }}>·</span>
                      <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)" }}>
                        <span style={{ fontWeight: 700, color: "var(--color-fg)" }}>
                          {diff >= 0 ? "+" : ""}{diff.toFixed(2)}{unit}
                        </span>
                        {" "}{diff >= 0 ? "ÖVER" : "UNDER"} MEDEL →
                      </span>
                    </Link>
                  );
                })()}
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
            gridTemplateColumns: isMobile ? "1fr" : "1fr 2fr",
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

            <MandateComposition
              ruling={ruling}
              totalSeats={totalSeats}
            />
          </div>

          {/* Right card — BUDGET */}
          <div style={{ background: "var(--color-sdt-surface)", padding: isMobile ? 16 : 24 }}>
            <BudgetHistorySection
              snapshots={budgetHistory}
              makeAreaLink={(a) => `/region/${code}/budget/${encodeURIComponent(a)}`}
              emptyMessage="Budgetdata saknas för denna region."
              sourceId="scb-kostndrlt"
            />
          </div>

        </div>

        {/* ── Bottom grid ──────────────────────────────────────────────── */}
        <div
          style={{
            display: "grid",
            gridTemplateColumns: isMobile ? "1fr" : budgetDoc ? "1.3fr 1fr" : "1fr",
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

          {/* Right card — REGIONENS PLAN (link to official primary source) */}
          {budgetDoc && (
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
                REGIONENS PLAN
              </div>
              <p style={{ margin: "0 0 16px", fontSize: 13, color: "var(--color-fg-muted)", lineHeight: 1.5 }}>
                {budgetDoc.goals && budgetDoc.goals.length > 0
                  ? "Regionens egna övergripande mål, ordagrant ur den beslutade planen. Vi tolkar dem inte — läs hela planen och dra dina egna slutsatser."
                  : "Regionens faktiska årsplan beslutas av regionfullmäktige. Vi länkar det officiella dokumentet i stället för att tolka det — läs och dra dina egna slutsatser."}
              </p>
              {budgetDoc.goals && budgetDoc.goals.length > 0 && (
                <div style={{ marginBottom: 18 }}>
                  {(budgetDoc.goalsLabel || budgetDoc.period) && (
                    <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, letterSpacing: "0.1em", color: "var(--color-fg-muted)", marginBottom: 10 }}>
                      {[budgetDoc.goalsLabel, budgetDoc.period].filter(Boolean).join(" · ").toUpperCase()}
                    </div>
                  )}
                  <ol style={{ listStyle: "none", margin: 0, padding: 0, display: "flex", flexDirection: "column", gap: 10 }}>
                    {budgetDoc.goals.map((g, i) => (
                      <li key={i} style={{ display: "flex", gap: 10, alignItems: "baseline" }}>
                        <span style={{ fontFamily: "var(--font-serif)", fontStyle: "italic", color: "var(--color-accent-2)", fontSize: 18, lineHeight: 1, minWidth: 18, flexShrink: 0 }}>
                          {i + 1}
                        </span>
                        <span style={{ fontSize: 13, color: "var(--color-fg)", lineHeight: 1.4 }}>
                          {g.title}
                          {g.description && (
                            <span style={{ display: "block", fontSize: 12, color: "var(--color-fg-muted)", marginTop: 2, lineHeight: 1.4 }}>
                              {g.description}
                            </span>
                          )}
                        </span>
                      </li>
                    ))}
                  </ol>
                </div>
              )}
              <a
                href={budgetDoc.url}
                target="_blank"
                rel="noopener noreferrer"
                style={{
                  display: "inline-flex",
                  alignItems: "baseline",
                  gap: 8,
                  fontSize: 14,
                  color: "var(--color-accent-2)",
                  textDecoration: "none",
                  borderBottom: "1px solid var(--color-accent-2)",
                  paddingBottom: 2,
                }}
              >
                {budgetDoc.goals && budgetDoc.goals.length > 0 ? `Läs hela: ${budgetDoc.label}` : budgetDoc.label}
                <span style={{ fontSize: 11 }}>↗</span>
              </a>
              <div style={{ marginTop: 14, fontSize: 11, color: "var(--color-fg-muted)", fontFamily: "var(--font-mono)" }}>
                Källa: {title}
                <SourceMarker sourceId="region-budget-plan" />
              </div>
            </div>
          )}
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
