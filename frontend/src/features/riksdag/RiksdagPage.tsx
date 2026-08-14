import { useState } from "react";
import { Link } from "react-router-dom";
import { useRiksdag, useRiksdagBudgetHistory } from "@/hooks/useDemocracy";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { MandateComposition, Pill, Trend } from "@/components/charts";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { BudgetHistorySection } from "@/features/budget/components/BudgetHistorySection";
import type { Authority, LiveVote } from "@/types/democracy";

function beslutHref(v: LiveVote): string | null {
  if (!v.beteckning) return null;
  const p = new URLSearchParams({ title: v.title, status: v.status, tag: v.tag, time: v.time });
  return `/beslut/${encodeURIComponent(v.beteckning)}?${p}`;
}

// National KPI label → registered data-source id. KPIs are ingested by the
// national-kpis worker; each label maps to the SCB table it comes from.
const NATIONAL_KPI_SOURCE: Record<string, string> = {
  "Arbetslöshet": "scb-aku-arbetsloshet",
  "Inflation (KPI)": "scb-kpi-inflation",
};

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

// ── MYNDIGHETER ───────────────────────────────────────────────────────────────

// Latest year-over-year change in expenditure from an authority's history.
function authorityYoY(a: Authority): number | null {
  if (a.history.length < 2) return null;
  const sorted = [...a.history].sort((x, y) => x.year - y.year);
  const latest = sorted[sorted.length - 1];
  const prev = sorted[sorted.length - 2];
  if (!prev || prev.expenditureMdkr === 0) return null;
  return ((latest.expenditureMdkr - prev.expenditureMdkr) / prev.expenditureMdkr) * 100;
}

function Stat({ label, value, source }: { label: string; value: string; source?: boolean }) {
  return (
    <div>
      <div
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 8,
          letterSpacing: "0.12em",
          color: "var(--color-fg-muted)",
          textTransform: "uppercase",
          marginBottom: 3,
        }}
      >
        {label}
      </div>
      <div style={{ fontFamily: "var(--font-mono)", fontSize: 15, fontWeight: 700, color: "var(--color-fg)" }}>
        {value}
        {source && <SourceMarker sourceId="statskontoret-arsutfall" />}
      </div>
    </div>
  );
}

function MyndighetRow({
  authority,
  color,
  share,
  isOpen,
  onToggle,
}: {
  authority: Authority;
  color: string;
  share: number;
  isOpen: boolean;
  onToggle: () => void;
}) {
  const yoy = authorityYoY(authority);
  return (
    <div style={{ borderTop: "1px solid var(--color-surface-high)" }}>
      <button
        onClick={onToggle}
        style={{
          display: "grid",
          gridTemplateColumns: "1fr auto auto auto 14px",
          gap: 8,
          alignItems: "center",
          padding: "8px 12px",
          width: "100%",
          background: isOpen ? "var(--color-surface-low)" : "var(--color-sdt-surface)",
          border: "none",
          cursor: "pointer",
          textAlign: "left",
          color: "inherit",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 8, minWidth: 0 }}>
          <div style={{ width: 10, height: 10, borderRadius: 2, flexShrink: 0, background: color }} />
          <span style={{ fontSize: 12, fontWeight: 500, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
            {authority.name}
          </span>
        </div>
        <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-on-surface-variant)", textAlign: "right", minWidth: 72 }}>
          {authority.expenditureMdkr.toFixed(1)} mdkr
        </span>
        <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-fg-muted)", textAlign: "right", minWidth: 44 }}>
          {share.toFixed(1)}%
        </span>
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            textAlign: "right",
            minWidth: 64,
            color: yoy == null ? "var(--color-fg-muted)" : yoy >= 0 ? "#4caf7d" : "#e05c5c",
          }}
        >
          {yoy == null ? "—" : `${yoy >= 0 ? "+" : ""}${yoy.toFixed(1)}%`}
        </span>
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 13,
            color: "var(--color-fg-muted)",
            minWidth: 12,
            transform: isOpen ? "rotate(90deg)" : "none",
            display: "inline-block",
            transition: "transform 0.15s",
          }}
        >
          ›
        </span>
      </button>

      {isOpen && (
        <div style={{ padding: "10px 14px 14px", background: "var(--color-surface-low)", borderTop: "1px solid var(--color-border)" }}>
          <div style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-fg-muted)", marginBottom: 12 }}>
            {authority.role}
            {authority.ministry && <span style={{ opacity: 0.7 }}> · {authority.ministry}</span>}
          </div>
          <div style={{ display: "flex", gap: 28, marginBottom: 14, flexWrap: "wrap" }}>
            <Stat label="Aktuellt år" value={String(authority.year)} />
            <Stat label="Anställda" value={authority.headcount} />
            <Stat label="Kostnad" value={`${authority.expenditureMdkr.toFixed(1)} mdkr`} source />
          </div>
          {authority.slug && (
            <Link
              to={`/riksdag/myndigheter/${authority.slug}`}
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                color: "var(--color-accent)",
                textDecoration: "none",
                border: "1px solid var(--color-border)",
                borderRadius: 4,
                padding: "5px 10px",
                display: "inline-block",
              }}
            >
              Läs mer →
            </Link>
          )}
        </div>
      )}
    </div>
  );
}

function MyndigheterCard({
  authorities,
  stateBudgetMdkr,
}: {
  authorities: Authority[];
  // Statens budget total (mdkr) for the latest year — used as the *honest*
  // denominator for "Andel". The 10 agencies on this card cover only ~7% of
  // the budget; the old "sum of shown / sum of shown = 100%" was misleading.
  stateBudgetMdkr: number | null;
}) {
  const [open, setOpen] = useState<number | null>(null);
  const total = authorities.reduce((s, a) => s + a.expenditureMdkr, 0);
  const year = authorities[0]?.year;
  const denom = stateBudgetMdkr ?? 0;
  const totalSharePct = denom > 0 ? (total / denom) * 100 : null;

  return (
    <div>
      <div
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 10,
          letterSpacing: "0.15em",
          color: "var(--color-fg-muted)",
          marginBottom: 12,
        }}
      >
        MYNDIGHETER · TOPP 10 EFTER DRIFTKOSTNAD{year ? ` ${year}` : ""}
        <SourceMarker sourceId="statskontoret-arsutfall" />
      </div>

      {authorities.length === 0 ? (
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-fg-muted)",
            padding: "24px 0",
            textAlign: "center",
          }}
        >
          Myndighetsdata saknas.
        </div>
      ) : (
        <div style={{ borderRadius: 8, overflow: "hidden", border: "1px solid var(--color-surface-high)" }}>
          {/* Header */}
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr auto auto auto 14px",
              gap: 8,
              padding: "6px 12px",
              background: "var(--color-surface-low)",
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              letterSpacing: "0.12em",
              color: "var(--color-on-surface-variant)",
              textTransform: "uppercase",
            }}
          >
            <span>Myndighet</span>
            <span style={{ textAlign: "right", minWidth: 72 }}>Senaste</span>
            <span style={{ textAlign: "right", minWidth: 44 }}>Andel</span>
            <span style={{ textAlign: "right", minWidth: 64 }}>Förändring</span>
            <span style={{ minWidth: 12 }} />
          </div>

          {authorities.map((a, i) => (
            <MyndighetRow
              key={a.slug || a.name}
              authority={a}
              color={BUDGET_COLORS[i % BUDGET_COLORS.length]}
              share={denom > 0 ? (a.expenditureMdkr / denom) * 100 : 0}
              isOpen={open === i}
              onToggle={() => setOpen(open === i ? null : i)}
            />
          ))}

          {/* Total row */}
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr auto auto auto 14px",
              gap: 8,
              alignItems: "center",
              padding: "8px 12px",
              borderTop: "2px solid var(--color-surface-highest)",
              background: "var(--color-surface-low)",
              fontWeight: 700,
            }}
          >
            <span style={{ fontFamily: "var(--font-mono)", fontSize: 11 }}>SUMMA TOPP-10</span>
            <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-fg)", textAlign: "right", minWidth: 72 }}>
              {total.toFixed(1)} mdkr
            </span>
            <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-fg-muted)", textAlign: "right", minWidth: 44 }}>
              {totalSharePct != null ? `${totalSharePct.toFixed(1)}%` : "—"}
            </span>
            <span style={{ minWidth: 64 }} />
            <span style={{ minWidth: 12 }} />
          </div>
        </div>
      )}

      {denom > 0 && (
        <div style={{ marginTop: 10, fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", letterSpacing: "0.08em" }}>
          Andel = av statens budget {Math.round(denom).toLocaleString("sv-SE")} mdkr ({year}).
        </div>
      )}

      <div style={{ marginTop: 14, textAlign: "right" }}>
        <Link
          to="/riksdag/myndigheter"
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-accent)",
            textDecoration: "none",
            border: "1px solid var(--color-border)",
            borderRadius: 4,
            padding: "6px 12px",
            display: "inline-block",
          }}
        >
          Läs om fler myndigheter →
        </Link>
      </div>
    </div>
  );
}

// ── RiksdagPage ───────────────────────────────────────────────────────────────

export function RiksdagPage() {
  const isMobile = useMediaQuery("(max-width: 640px)");
  const { data, isLoading } = useRiksdag();
  const { data: budgetHistory = [] } = useRiksdagBudgetHistory();

  if (isLoading || !data) return <Skeleton />;

  const { ruling, liveVotes = [], kpis, authorities } = data;

  const totalSeats = [...ruling.parties, ...(ruling.support ?? []), ...ruling.opposition].reduce((s, p) => s + p.seats, 0);

  // Statens budget total (mdkr) for the latest year present in budgetHistory.
  // BudgetSnapshot.total_mnkr is the full state-budget total for that year, so
  // any row of the latest year works as the source.
  const latestBudgetYear = budgetHistory.length > 0 ? Math.max(...budgetHistory.map((s) => s.year)) : 0;
  const stateBudgetMdkr =
    latestBudgetYear > 0
      ? (budgetHistory.find((s) => s.year === latestBudgetYear)?.total_mnkr ?? 0) / 1000
      : null;

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
                {NATIONAL_KPI_SOURCE[k.label] && (
                  <SourceMarker sourceId={NATIONAL_KPI_SOURCE[k.label]} />
                )}
              </div>
              <div style={{ marginBottom: 12 }}>
                <Trend trend={k.trend} delta={k.delta} worseHigher={k.worseHigher} />
              </div>
              {k.note && (
                <div
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 9,
                    color: "var(--color-fg-muted)",
                    letterSpacing: "0.06em",
                  }}
                >
                  {k.sourceUrl ? (
                    <a
                      href={k.sourceUrl}
                      target="_blank"
                      rel="noreferrer"
                      style={{ color: "inherit" }}
                    >
                      {k.note}
                    </a>
                  ) : (
                    k.note
                  )}
                </div>
              )}
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

          <MandateComposition ruling={ruling} totalSeats={totalSeats} sourceId="riksdagen" />
        </div>

        {/* Right card — MYNDIGHETER */}
        <div
          style={{
            background: "var(--color-sdt-surface)",
            padding: isMobile ? 16 : 24,
          }}
        >
          <MyndigheterCard authorities={authorities ?? []} stateBudgetMdkr={stateBudgetMdkr} />
        </div>
      </div>

      {/* ── Budget (full-width) ──────────────────────────────────────── */}
      {budgetHistory.length > 0 && (
        <div
          style={{
            border: "1px solid var(--color-border)",
            borderTop: "none",
            margin: isMobile ? "0 14px 0" : "0 32px 0",
          }}
        >
          <BudgetHistorySection snapshots={budgetHistory} sourceId="seed-budget-data" />
        </div>
      )}

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
        </div>

        {/* The government-document index moved to /regering. It was headed
            "STYRETS PRIORITERINGAR", which named our own selection of five
            points as the government's priorities, and it is government-level
            material rather than something about the Riksdag. */}
      </div>
    </div>
  );
}
