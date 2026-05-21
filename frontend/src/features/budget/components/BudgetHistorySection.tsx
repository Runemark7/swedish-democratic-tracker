import { useState, useMemo, useRef, useEffect } from "react";
import { Link } from "react-router-dom";
import { DeltaIndicator } from "./DeltaIndicator";
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

export function BudgetHistorySection({
  snapshots,
  makeAreaLink,
  emptyMessage = "Budgetdata saknas.",
  sourceId,
}: BudgetHistorySectionProps) {
  const isMobile = useMediaQuery("(max-width: 640px)");

  // Group snapshots by year.
  const historyByYear = new Map<number, BudgetSnapshot[]>();
  for (const snap of snapshots) {
    const list = historyByYear.get(snap.year) ?? [];
    list.push(snap);
    historyByYear.set(snap.year, list);
  }
  const sortedYears = [...historyByYear.keys()].sort((a, b) => b - a);

  const [activeTab, setActiveTab] = useState<number | null>(null);
  const activeYear = activeTab ?? sortedYears[0] ?? null;

  // --- Chart state ---
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
    // Initialise with actual rendered width
    const initial = el.getBoundingClientRect().width;
    if (initial > 0) setSvgWidth(initial);
    return () => obs.disconnect();
  }, []);

  // Chart geometry constants
  const padL = 48, padR = 16, padT = 12, padB = 28;
  const chartH = isMobile ? 140 : 180;

  const chartData = useMemo(() => {
    if (sortedYears.length < 2) return null;

    // Years ascending for x-axis
    const allYears = [...sortedYears].sort((a, b) => a - b);

    // Area order = order in the most recent year's data (index → color matches table)
    const mostRecentAreas = historyByYear.get(sortedYears[0]) ?? [];
    const areaNames = mostRecentAreas.map((a) => a.area_name);

    // Build lookup: area_name → year → value_mnkr
    const byArea = new Map<string, Map<number, number>>();
    for (const snap of snapshots) {
      let m = byArea.get(snap.area_name);
      if (!m) { m = new Map(); byArea.set(snap.area_name, m); }
      m.set(snap.year, snap.value_mnkr);
    }

    const maxVal = Math.max(...snapshots.map((s) => s.value_mnkr), 1);

    return { allYears, areaNames, byArea, maxVal };
  }, [snapshots, sortedYears, historyByYear]);

  const innerW = svgWidth - padL - padR;
  const innerH = chartH - padT - padB;

  const xScale = (year: number, allYears: number[]): number => {
    if (allYears.length < 2) return padL;
    return padL + ((year - allYears[0]) / (allYears[allYears.length - 1] - allYears[0])) * innerW;
  };

  const yScale = (val: number, maxVal: number): number =>
    padT + innerH - (val / maxVal) * innerH;

  return (
    <div style={{ background: "var(--color-sdt-surface)", padding: isMobile ? 16 : 24 }}>
      {/* Section label */}
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

      {/* Multi-line trend chart (only when ≥2 years) */}
      {chartData && (
        <div style={{ marginBottom: 16 }}>
          <svg
            ref={svgRef}
            width="100%"
            height={chartH}
            style={{ display: "block", overflow: "visible" }}
          >
            {/* Y-axis line */}
            <line
              x1={padL} y1={padT}
              x2={padL} y2={padT + innerH}
              stroke="var(--color-border)"
              strokeWidth={1}
            />
            {/* X-axis line */}
            <line
              x1={padL} y1={padT + innerH}
              x2={padL + innerW} y2={padT + innerH}
              stroke="var(--color-border)"
              strokeWidth={1}
            />

            {/* Y-axis labels: 0, max/2, max */}
            {[0, 0.5, 1].map((frac) => {
              const val = Math.round(frac * chartData.maxVal);
              const y = yScale(frac * chartData.maxVal, chartData.maxVal);
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

            {/* X-axis year labels + tick marks */}
            {chartData.allYears.map((yr) => {
              const x = xScale(yr, chartData.allYears);
              return (
                <g key={yr}>
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
                    {yr}
                  </text>
                </g>
              );
            })}

            {/* Lines + dots per area */}
            {chartData.areaNames.map((areaName, i) => {
              const color = BUDGET_COLORS[i % BUDGET_COLORS.length];
              const valByYear = chartData.byArea.get(areaName);
              if (!valByYear) return null;

              // Build polyline segments (skip gaps)
              const points = chartData.allYears
                .filter((yr) => valByYear.has(yr))
                .map((yr) => {
                  const val = valByYear.get(yr)!;
                  return `${xScale(yr, chartData.allYears).toFixed(1)},${yScale(val, chartData.maxVal).toFixed(1)}`;
                })
                .join(" ");

              return (
                <g key={areaName}>
                  {points && (
                    <polyline
                      points={points}
                      fill="none"
                      stroke={color}
                      strokeWidth={2}
                      strokeLinejoin="round"
                    />
                  )}
                  {chartData.allYears
                    .filter((yr) => valByYear.has(yr))
                    .map((yr) => {
                      const val = valByYear.get(yr)!;
                      return (
                        <circle
                          key={yr}
                          cx={xScale(yr, chartData.allYears)}
                          cy={yScale(val, chartData.maxVal)}
                          r={3}
                          fill={color}
                        />
                      );
                    })}
                </g>
              );
            })}
          </svg>

          {/* Legend */}
          <div
            style={{
              display: "flex",
              flexWrap: "wrap",
              gap: "4px 12px",
              marginTop: 6,
            }}
          >
            {chartData.areaNames.map((areaName, i) => (
              <div
                key={areaName}
                style={{ display: "flex", alignItems: "center", gap: 4 }}
              >
                <div
                  style={{
                    width: 8,
                    height: 8,
                    borderRadius: 2,
                    flexShrink: 0,
                    background: BUDGET_COLORS[i % BUDGET_COLORS.length],
                  }}
                />
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 10,
                    color: "var(--color-fg-muted)",
                    whiteSpace: "nowrap",
                  }}
                >
                  {areaName}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

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
          {/* Year tabs */}
          <div style={{ display: "flex", gap: 4, flexWrap: "wrap", marginBottom: 16 }}>
            {sortedYears.map((yr, idx) => {
              const prevYear = sortedYears[idx + 1] ?? null;
              const prevSnaps = prevYear != null ? (historyByYear.get(prevYear) ?? []) : null;
              const curTotal = (historyByYear.get(yr) ?? []).reduce((s, a) => s + a.value_mnkr, 0);
              const prevTotal = prevSnaps ? prevSnaps.reduce((s, a) => s + a.value_mnkr, 0) : null;
              const delta =
                prevTotal != null && prevTotal > 0
                  ? ((curTotal - prevTotal) / prevTotal) * 100
                  : null;
              const isActive = yr === activeYear;
              const tabDeltaColor =
                delta == null
                  ? "var(--color-fg-muted)"
                  : delta >= 0
                  ? "#4caf7d"
                  : "#e05c5c";

              return (
                <button
                  key={yr}
                  onClick={() => setActiveTab(yr)}
                  className="px-3 py-1.5 text-xs font-mono font-bold rounded-md transition-all"
                  style={{
                    background: isActive
                      ? "var(--color-primary)"
                      : "var(--color-surface-low)",
                    color: isActive
                      ? "var(--color-on-primary)"
                      : "var(--color-on-surface)",
                  }}
                >
                  {yr}
                  {delta != null && (
                    <span
                      style={{
                        marginLeft: 4,
                        fontSize: 10,
                        color: isActive ? "var(--color-on-primary)" : tabDeltaColor,
                      }}
                    >
                      {delta >= 0 ? "+" : ""}
                      {delta.toFixed(1)}%
                    </span>
                  )}
                </button>
              );
            })}
          </div>

          {/* Active year table */}
          {activeYear != null &&
            (() => {
              const areas = historyByYear.get(activeYear) ?? [];
              const prevYear =
                sortedYears[sortedYears.indexOf(activeYear) + 1] ?? null;
              const prevAreas =
                prevYear != null ? (historyByYear.get(prevYear) ?? []) : [];
              const prevByName = new Map(prevAreas.map((a) => [a.area_name, a]));

              return (
                <div
                  style={{
                    borderRadius: 8,
                    overflow: "hidden",
                    border: "1px solid var(--color-surface-high)",
                  }}
                >
                  {/* Header row */}
                  <div
                    style={{
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
                    }}
                  >
                    <span>Område</span>
                    <span style={{ textAlign: "right", minWidth: 60 }}>
                      {prevYear ?? "—"}
                    </span>
                    <span style={{ textAlign: "right", minWidth: 80 }}>
                      Förändring
                    </span>
                  </div>

                  {/* Area rows */}
                  {areas.map((a, i) => {
                    const prev = prevByName.get(a.area_name);
                    const deltaPct =
                      prev && prev.value_mnkr > 0
                        ? ((a.value_mnkr - prev.value_mnkr) / prev.value_mnkr) * 100
                        : null;

                    const rowStyle = {
                      display: "grid",
                      gridTemplateColumns: "1fr auto auto",
                      gap: 8,
                      alignItems: "center",
                      padding: "8px 12px",
                      textDecoration: "none" as const,
                      color: "inherit",
                      borderTop: "1px solid var(--color-surface-high)",
                      background: "var(--color-sdt-surface)",
                    };

                    const inner = (
                      <>
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
                        <span
                          style={{
                            fontFamily: "var(--font-mono)",
                            fontSize: 11,
                            color: "var(--color-on-surface-variant)",
                            textAlign: "right",
                            minWidth: 60,
                          }}
                        >
                          {prev ? `${Math.round(prev.value_mnkr)} mnkr` : "—"}
                        </span>
                        <div
                          style={{
                            minWidth: 80,
                            display: "flex",
                            justifyContent: "flex-end",
                          }}
                        >
                          {deltaPct != null ? (
                            <DeltaIndicator
                              pct={Math.round(deltaPct * 10) / 10}
                              showBar={false}
                            />
                          ) : (
                            <span
                              style={{
                                fontFamily: "var(--font-mono)",
                                fontSize: 11,
                                color: "var(--color-on-surface-variant)",
                              }}
                            >
                              —
                            </span>
                          )}
                        </div>
                      </>
                    );

                    return makeAreaLink ? (
                      <Link
                        key={a.area_name}
                        to={makeAreaLink(a.area_name)}
                        style={rowStyle}
                      >
                        {inner}
                      </Link>
                    ) : (
                      <div key={a.area_name} style={rowStyle}>
                        {inner}
                      </div>
                    );
                  })}

                  {/* Total row */}
                  <div
                    style={{
                      display: "grid",
                      gridTemplateColumns: "1fr auto auto",
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
                        color: "var(--color-on-surface-variant)",
                        textAlign: "right",
                        minWidth: 60,
                      }}
                    >
                      {prevAreas.length > 0
                        ? `${Math.round(
                            prevAreas.reduce((s, a) => s + a.value_mnkr, 0)
                          )} mnkr`
                        : "—"}
                    </span>
                    <div
                      style={{
                        minWidth: 80,
                        display: "flex",
                        justifyContent: "flex-end",
                      }}
                    >
                      {(() => {
                        const curTot = areas.reduce((s, a) => s + a.value_mnkr, 0);
                        const prevTot = prevAreas.reduce(
                          (s, a) => s + a.value_mnkr,
                          0
                        );
                        return prevTot > 0 ? (
                          <DeltaIndicator
                            pct={
                              Math.round(
                                ((curTot - prevTot) / prevTot) * 1000
                              ) / 10
                            }
                            showBar={false}
                          />
                        ) : (
                          <span
                            style={{
                              fontFamily: "var(--font-mono)",
                              fontSize: 11,
                              color: "var(--color-on-surface-variant)",
                            }}
                          >
                            —
                          </span>
                        );
                      })()}
                    </div>
                  </div>
                </div>
              );
            })()}
        </>
      )}
    </div>
  );
}
