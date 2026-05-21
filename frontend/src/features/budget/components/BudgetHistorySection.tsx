import { useState, useMemo, useRef, useEffect } from "react";
import { Link } from "react-router-dom";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import type { BudgetSnapshot } from "@/shared/types";

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

interface BudgetHistorySectionProps {
  snapshots: BudgetSnapshot[];
  makeAreaLink?: (areaName: string) => string;
  emptyMessage?: string;
  sourceId: string;
}

// ── ExpandedAreaRow ────────────────────────────────────────────────────────────

interface ExpandedAreaRowProps {
  areaName: string;
  yearSnapshots: { year: number; value_mnkr: number }[];
  makeAreaLink?: (areaName: string) => string;
}

function ExpandedAreaRow({ areaName, yearSnapshots, makeAreaLink }: ExpandedAreaRowProps) {
  const byYear = useMemo(() => {
    const m = new Map<number, number>();
    for (const s of yearSnapshots) m.set(s.year, s.value_mnkr);
    return m;
  }, [yearSnapshots]);

  const oldestYear = useMemo(
    () => Math.min(...yearSnapshots.map((s) => s.year)),
    [yearSnapshots]
  );
  const latestYear = useMemo(
    () => Math.max(...yearSnapshots.map((s) => s.year)),
    [yearSnapshots]
  );
  const sorted = useMemo(
    () => [...yearSnapshots].sort((a, b) => b.value_mnkr - a.value_mnkr),
    [yearSnapshots]
  );
  const maxVal = useMemo(
    () => Math.max(...sorted.map((s) => s.value_mnkr), 1),
    [sorted]
  );

  return (
    <div
      style={{
        padding: "10px 14px 14px",
        background: "var(--color-surface-low)",
        borderTop: "1px solid var(--color-border)",
      }}
    >
      {sorted.map((snap) => {
        const hl = snap.year === latestYear;
        const pct = (snap.value_mnkr / maxVal) * 100;
        const prevVal = byYear.get(snap.year - 1);
        const deltaPct =
          snap.year !== oldestYear && prevVal != null && prevVal > 0
            ? ((snap.value_mnkr - prevVal) / prevVal) * 100
            : null;

        return (
          <div
            key={snap.year}
            style={{ display: "flex", alignItems: "center", gap: 8, padding: "3px 0" }}
          >
            {/* Year */}
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                color: hl ? "var(--color-fg)" : "var(--color-fg-muted)",
                fontWeight: hl ? 700 : 400,
                width: 36,
                textAlign: "right",
                flexShrink: 0,
              }}
            >
              {snap.year}
            </span>

            {/* YoY delta */}
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                width: 52,
                textAlign: "right",
                flexShrink: 0,
                color:
                  deltaPct == null
                    ? "var(--color-fg-muted)"
                    : deltaPct >= 0
                    ? "#4caf7d"
                    : "#e05c5c",
              }}
            >
              {deltaPct == null
                ? "—"
                : `${deltaPct >= 0 ? "+" : ""}${deltaPct.toFixed(1)}%`}
            </span>

            {/* Value */}
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                color: hl ? "var(--color-fg)" : "var(--color-fg-muted)",
                width: 68,
                textAlign: "right",
                flexShrink: 0,
              }}
            >
              {Math.round(snap.value_mnkr).toLocaleString("sv-SE")}
            </span>

            {/* Bar */}
            <div
              style={{
                flex: 1,
                height: 8,
                background: "var(--color-surface-high, #1c1c21)",
                borderRadius: 2,
              }}
            >
              <div
                style={{
                  width: `${pct}%`,
                  height: "100%",
                  borderRadius: 2,
                  background: hl ? "var(--color-accent, #7c9ff5)" : "#2d3a52",
                }}
              />
            </div>
          </div>
        );
      })}

      {makeAreaLink && (
        <div style={{ marginTop: 10 }}>
          <Link
            to={makeAreaLink(areaName)}
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              color: "var(--color-accent, #7c9ff5)",
              textDecoration: "none",
              border: "1px solid rgba(124,159,245,0.3)",
              borderRadius: 4,
              padding: "5px 10px",
              display: "inline-block",
            }}
          >
            → Visa jämförelse med andra regioner
          </Link>
        </div>
      )}
    </div>
  );
}

// ── BudgetHistorySection ───────────────────────────────────────────────────────

export function BudgetHistorySection({
  snapshots,
  makeAreaLink,
  emptyMessage = "Budgetdata saknas.",
  sourceId,
}: BudgetHistorySectionProps) {
  const isMobile = useMediaQuery("(max-width: 640px)");

  // Group snapshots by year.
  const historyByYear = useMemo(() => {
    const m = new Map<number, BudgetSnapshot[]>();
    for (const snap of snapshots) {
      const list = m.get(snap.year) ?? [];
      list.push(snap);
      m.set(snap.year, list);
    }
    return m;
  }, [snapshots]);

  const sortedYears = useMemo(
    () => [...historyByYear.keys()].sort((a, b) => b - a),
    [historyByYear]
  );

  const [expandedArea, setExpandedArea] = useState<string | null>(null);

  // Most recent year data, sorted by value descending
  const latestYear = sortedYears[0] ?? null;
  const previousYear = sortedYears[1] ?? null;

  const latestAreas = useMemo(
    () =>
      latestYear != null
        ? [...(historyByYear.get(latestYear) ?? [])].sort(
            (a, b) => b.value_mnkr - a.value_mnkr
          )
        : [],
    [historyByYear, latestYear]
  );

  const prevByName = useMemo(() => {
    const prevAreas =
      previousYear != null ? (historyByYear.get(previousYear) ?? []) : [];
    return new Map(prevAreas.map((a) => [a.area_name, a]));
  }, [historyByYear, previousYear]);

  // ── Total trend line chart ────────────────────────────────────────────────────
  const svgRef = useRef<SVGSVGElement>(null);
  const [svgWidth, setSvgWidth] = useState<number>(600);

  useEffect(() => {
    const el = svgRef.current;
    if (!el) return;
    const obs = new ResizeObserver((entries) => {
      const w = entries[0]?.contentRect.width;
      if (w && w > 0) setSvgWidth(w);
    });
    obs.observe(el);
    const initial = el.getBoundingClientRect().width;
    if (initial > 0) setSvgWidth(initial);
    return () => obs.disconnect();
  }, []);

  const chartH = 110;
  const padL = 44, padR = 12, padT = 10, padB = 22;
  const innerW = svgWidth - padL - padR;
  const innerH = chartH - padT - padB;

  const totalTrend = useMemo((): { year: number; total: number }[] | null => {
    if (sortedYears.length < 2) return null;
    const ascending = [...sortedYears].sort((a, b) => a - b);
    return ascending.map((yr) => ({
      year: yr,
      total: (historyByYear.get(yr) ?? []).reduce((s, a) => s + a.value_mnkr, 0),
    }));
  }, [sortedYears, historyByYear]);

  const trendMax = totalTrend
    ? Math.max(...totalTrend.map((p) => p.total), 1)
    : 1;

  const xScale = (year: number, years: number[]): number => {
    if (years.length < 2) return padL;
    return (
      padL +
      ((year - years[0]) / (years[years.length - 1] - years[0])) * innerW
    );
  };

  const yScale = (val: number): number =>
    padT + innerH - (val / trendMax) * innerH;

  // Headline KPI
  const latestTotal =
    latestYear != null
      ? (historyByYear.get(latestYear) ?? []).reduce(
          (s, a) => s + a.value_mnkr,
          0
        )
      : 0;
  const prevTotal =
    previousYear != null
      ? (historyByYear.get(previousYear) ?? []).reduce(
          (s, a) => s + a.value_mnkr,
          0
        )
      : null;
  const yoyDeltaPct =
    prevTotal != null && prevTotal > 0
      ? ((latestTotal - prevTotal) / prevTotal) * 100
      : null;

  return (
    <div style={{ background: "var(--color-sdt-surface)", padding: isMobile ? 16 : 24 }}>
      {/* 1. Section header */}
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
        <SourceMarker sourceId={sourceId} />
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
          {emptyMessage}
        </div>
      ) : (
        <>
          {/* 2. Total trend line chart */}
          {totalTrend && (
            <div style={{ marginBottom: 20 }}>
              {/* Headline KPI */}
              <div
                style={{
                  display: "flex",
                  alignItems: "baseline",
                  gap: 8,
                  marginBottom: 8,
                }}
              >
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 20,
                    fontWeight: 700,
                    color: "var(--color-fg)",
                  }}
                >
                  {Math.round(latestTotal).toLocaleString("sv-SE")} mnkr
                </span>
                {yoyDeltaPct != null && (
                  <span
                    style={{
                      fontFamily: "var(--font-mono)",
                      fontSize: 11,
                      color: yoyDeltaPct >= 0 ? "#4caf7d" : "#e05c5c",
                    }}
                  >
                    {yoyDeltaPct >= 0 ? "+" : ""}
                    {yoyDeltaPct.toFixed(1)}%
                  </span>
                )}
                {latestYear != null && (
                  <span
                    style={{
                      fontFamily: "var(--font-mono)",
                      fontSize: 10,
                      color: "var(--color-fg-muted)",
                    }}
                  >
                    {latestYear}
                  </span>
                )}
              </div>

              {/* Line chart */}
              <svg
                ref={svgRef}
                width="100%"
                height={chartH}
                style={{ display: "block", overflow: "visible" }}
              >
                {/* Y-axis */}
                <line
                  x1={padL} y1={padT}
                  x2={padL} y2={padT + innerH}
                  stroke="var(--color-border)"
                  strokeWidth={1}
                />
                {/* X-axis */}
                <line
                  x1={padL} y1={padT + innerH}
                  x2={padL + innerW} y2={padT + innerH}
                  stroke="var(--color-border)"
                  strokeWidth={1}
                />

                {/* Y-axis 3 ticks */}
                {[0, 0.5, 1].map((frac) => {
                  const val = Math.round(frac * trendMax);
                  const y = yScale(frac * trendMax);
                  return (
                    <g key={frac}>
                      <line
                        x1={padL - 3} y1={y}
                        x2={padL} y2={y}
                        stroke="var(--color-border)"
                        strokeWidth={1}
                      />
                      <text
                        x={padL - 6}
                        y={y + 3}
                        textAnchor="end"
                        fontSize={9}
                        fontFamily="var(--font-mono)"
                        fill="var(--color-fg-muted)"
                      >
                        {val >= 1000 ? `${Math.round(val / 1000)}k` : String(val)}
                      </text>
                    </g>
                  );
                })}

                {/* X-axis year labels */}
                {totalTrend.map((pt) => {
                  const x = xScale(pt.year, totalTrend.map((p) => p.year));
                  return (
                    <g key={pt.year}>
                      <line
                        x1={x} y1={padT + innerH}
                        x2={x} y2={padT + innerH + 4}
                        stroke="var(--color-border)"
                        strokeWidth={1}
                      />
                      <text
                        x={x}
                        y={padT + innerH + 14}
                        textAnchor="middle"
                        fontSize={9}
                        fontFamily="var(--font-mono)"
                        fill="var(--color-fg-muted)"
                      >
                        {pt.year}
                      </text>
                    </g>
                  );
                })}

                {/* Line */}
                {(() => {
                  const years = totalTrend.map((p) => p.year);
                  const pts = totalTrend
                    .map(
                      (p) =>
                        `${xScale(p.year, years).toFixed(1)},${yScale(p.total).toFixed(1)}`
                    )
                    .join(" ");
                  return (
                    <polyline
                      points={pts}
                      fill="none"
                      stroke="var(--color-accent, #7c9ff5)"
                      strokeWidth={2}
                      strokeLinejoin="round"
                    />
                  );
                })()}

                {/* Dots */}
                {totalTrend.map((pt) => {
                  const years = totalTrend.map((p) => p.year);
                  return (
                    <circle
                      key={pt.year}
                      cx={xScale(pt.year, years)}
                      cy={yScale(pt.total)}
                      r={3}
                      fill="var(--color-accent, #7c9ff5)"
                    />
                  );
                })}
              </svg>
            </div>
          )}

          {/* 3. Area rows + header */}
          <div
            style={{
              borderRadius: 8,
              overflow: "hidden",
              border: "1px solid var(--color-surface-high)",
            }}
          >
            {/* Header */}
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr auto auto auto",
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
              <span>Område</span>
              <span style={{ textAlign: "right", minWidth: 72 }}>Senaste</span>
              <span style={{ textAlign: "right", minWidth: 64 }}>Förändring</span>
              <span style={{ minWidth: 12 }} />
            </div>

            {latestAreas.map((a, i) => {
              const prev = prevByName.get(a.area_name);
              const deltaPct =
                prev && prev.value_mnkr > 0
                  ? ((a.value_mnkr - prev.value_mnkr) / prev.value_mnkr) * 100
                  : null;
              const isExpanded = expandedArea === a.area_name;

              const handleToggle = () => {
                setExpandedArea(isExpanded ? null : a.area_name);
              };

              return (
                <div
                  key={a.area_name}
                  style={{ borderTop: "1px solid var(--color-surface-high)" }}
                >
                  {/* Collapsed row */}
                  <button
                    onClick={handleToggle}
                    style={{
                      display: "grid",
                      gridTemplateColumns: "1fr auto auto auto",
                      gap: 8,
                      alignItems: "center",
                      padding: "8px 12px",
                      width: "100%",
                      background: isExpanded
                        ? "var(--color-surface-low)"
                        : "var(--color-sdt-surface)",
                      border: "none",
                      cursor: "pointer",
                      textAlign: "left",
                      color: "inherit",
                    }}
                  >
                    {/* Color dot + name */}
                    <div
                      style={{
                        display: "flex",
                        alignItems: "center",
                        gap: 8,
                        minWidth: 0,
                      }}
                    >
                      <div
                        style={{
                          width: 10,
                          height: 10,
                          borderRadius: 2,
                          flexShrink: 0,
                          background: BUDGET_COLORS[i % BUDGET_COLORS.length],
                        }}
                      />
                      <span
                        style={{
                          fontSize: 12,
                          fontWeight: 500,
                          whiteSpace: "nowrap",
                          overflow: "hidden",
                          textOverflow: "ellipsis",
                        }}
                      >
                        {a.area_name}
                      </span>
                    </div>

                    {/* Latest value */}
                    <span
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 11,
                        color: "var(--color-on-surface-variant)",
                        textAlign: "right",
                        minWidth: 72,
                      }}
                    >
                      {Math.round(a.value_mnkr)} mnkr
                    </span>

                    {/* YoY delta */}
                    <span
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 11,
                        textAlign: "right",
                        minWidth: 64,
                        color:
                          deltaPct == null
                            ? "var(--color-fg-muted)"
                            : deltaPct >= 0
                            ? "#4caf7d"
                            : "#e05c5c",
                      }}
                    >
                      {deltaPct != null
                        ? `${deltaPct >= 0 ? "+" : ""}${deltaPct.toFixed(1)}%`
                        : "—"}
                    </span>

                    {/* Chevron */}
                    <span
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 13,
                        color: "var(--color-fg-muted)",
                        minWidth: 12,
                        transform: isExpanded ? "rotate(90deg)" : "none",
                        display: "inline-block",
                        transition: "transform 0.15s",
                      }}
                    >
                      ›
                    </span>
                  </button>

                  {/* Expanded content */}
                  {isExpanded && (
                    <ExpandedAreaRow
                      areaName={a.area_name}
                      yearSnapshots={[...historyByYear.entries()].flatMap(
                        ([yr, areas]) => {
                          const snap = areas.find(
                            (s) => s.area_name === a.area_name
                          );
                          return snap
                            ? [{ year: yr, value_mnkr: snap.value_mnkr }]
                            : [];
                        }
                      )}
                      makeAreaLink={makeAreaLink}
                    />
                  )}
                </div>
              );
            })}

            {/* 4. Total row */}
            {latestYear != null && (
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr auto auto auto",
                  gap: 8,
                  alignItems: "center",
                  padding: "8px 12px",
                  borderTop: "2px solid var(--color-surface-highest)",
                  background: "var(--color-surface-low)",
                  fontWeight: 700,
                }}
              >
                <span style={{ fontFamily: "var(--font-mono)", fontSize: 11 }}>
                  TOTALT
                </span>
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 11,
                    color: "var(--color-fg)",
                    textAlign: "right",
                    minWidth: 72,
                  }}
                >
                  {Math.round(latestTotal).toLocaleString("sv-SE")} mnkr
                </span>
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 11,
                    textAlign: "right",
                    minWidth: 64,
                    color:
                      yoyDeltaPct == null
                        ? "var(--color-fg-muted)"
                        : yoyDeltaPct >= 0
                        ? "#4caf7d"
                        : "#e05c5c",
                  }}
                >
                  {yoyDeltaPct != null
                    ? `${yoyDeltaPct >= 0 ? "+" : ""}${yoyDeltaPct.toFixed(1)}%`
                    : "—"}
                </span>
                <span style={{ minWidth: 12 }} />
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}
