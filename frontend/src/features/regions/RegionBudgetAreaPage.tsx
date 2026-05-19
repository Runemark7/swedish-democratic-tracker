import { Link, useParams } from "react-router-dom";
import {
  useAreaAcrossRegions,
  useRegionBudgetHistory,
  useRegionList,
} from "@/hooks/useDemocracy";
import { HBars } from "@/components/charts";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import type { RegionBudgetSnapshot } from "@/shared/types";

// ── Skeleton ──────────────────────────────────────────────────────────────────

function Skeleton() {
  return (
    <div className="sdt-page" style={{ padding: "40px 32px" }}>
      {[80, 200, 320].map((h, i) => (
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

// ── MiniLineChart ─────────────────────────────────────────────────────────────
// Simple SVG polyline for the 4-year trend within a region.

interface MiniLineChartProps {
  points: { year: number; value: number }[];
}

function MiniLineChart({ points }: MiniLineChartProps) {
  if (points.length < 2) {
    return (
      <div
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 11,
          color: "var(--color-fg-muted)",
          padding: "16px 0",
        }}
      >
        Data saknas
      </div>
    );
  }

  const W = 480;
  const H = 90;
  const PAD_LEFT = 48;
  const PAD_RIGHT = 12;
  const PAD_TOP = 12;
  const PAD_BOTTOM = 24;

  const innerW = W - PAD_LEFT - PAD_RIGHT;
  const innerH = H - PAD_TOP - PAD_BOTTOM;

  const sorted = [...points].sort((a, b) => a.year - b.year);
  const values = sorted.map((p) => p.value);
  const minV = Math.min(...values);
  const maxV = Math.max(...values);
  const range = maxV - minV || 1;

  const toX = (i: number) =>
    PAD_LEFT + (sorted.length === 1 ? innerW / 2 : (i / (sorted.length - 1)) * innerW);
  const toY = (v: number) => PAD_TOP + innerH - ((v - minV) / range) * innerH;

  const pathD = sorted
    .map((p, i) => `${i === 0 ? "M" : "L"} ${toX(i).toFixed(1)} ${toY(p.value).toFixed(1)}`)
    .join(" ");

  // Area fill path
  const areaD =
    pathD +
    ` L ${toX(sorted.length - 1).toFixed(1)} ${(PAD_TOP + innerH).toFixed(1)}` +
    ` L ${toX(0).toFixed(1)} ${(PAD_TOP + innerH).toFixed(1)} Z`;

  return (
    <svg
      viewBox={`0 0 ${W} ${H}`}
      width="100%"
      style={{ display: "block", overflow: "visible" }}
      aria-hidden="true"
    >
      <defs>
        <linearGradient id="area-grad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="var(--color-primary, #2d6fa8)" stopOpacity="0.18" />
          <stop offset="100%" stopColor="var(--color-primary, #2d6fa8)" stopOpacity="0" />
        </linearGradient>
      </defs>

      {/* Area fill */}
      <path d={areaD} fill="url(#area-grad)" />

      {/* Polyline */}
      <path
        d={pathD}
        fill="none"
        stroke="var(--color-primary, #2d6fa8)"
        strokeWidth="2"
        strokeLinejoin="round"
        strokeLinecap="round"
      />

      {/* Data dots + year labels */}
      {sorted.map((p, i) => (
        <g key={p.year}>
          <circle
            cx={toX(i)}
            cy={toY(p.value)}
            r={3}
            fill="var(--color-primary, #2d6fa8)"
          />
          {/* Year label */}
          <text
            x={toX(i)}
            y={H - 4}
            textAnchor="middle"
            fontSize={9}
            fill="var(--color-fg-muted)"
            fontFamily="var(--font-mono)"
          >
            {p.year}
          </text>
          {/* Value label above dot */}
          <text
            x={toX(i)}
            y={toY(p.value) - 7}
            textAnchor="middle"
            fontSize={9}
            fill="var(--color-fg)"
            fontFamily="var(--font-mono)"
          >
            {Math.round(p.value).toLocaleString("sv-SE")}
          </text>
        </g>
      ))}

      {/* Y-axis ticks */}
      {[minV, (minV + maxV) / 2, maxV].map((v) => (
        <text
          key={v}
          x={PAD_LEFT - 4}
          y={toY(v) + 3}
          textAnchor="end"
          fontSize={8}
          fill="var(--color-fg-muted)"
          fontFamily="var(--font-mono)"
        >
          {Math.round(v).toLocaleString("sv-SE")}
        </text>
      ))}
    </svg>
  );
}

// ── RegionBudgetAreaPage ──────────────────────────────────────────────────────

export function RegionBudgetAreaPage() {
  const { code = "", areaName: rawArea = "" } = useParams<{
    code: string;
    areaName: string;
  }>();
  const areaName = decodeURIComponent(rawArea);
  const isMobile = useMediaQuery("(max-width: 640px)");

  const { data: regionList, isLoading: loadingRegions } = useRegionList();
  const { data: historySnaps = [], isLoading: loadingHistory } =
    useRegionBudgetHistory(code);
  const { data: areaPoints = [], isLoading: loadingArea } =
    useAreaAcrossRegions(areaName);

  const isLoading = loadingRegions || loadingHistory || loadingArea;

  // Build region code → name map
  const regionName = regionList?.find((r) => r.code === code)?.name ?? code;
  const codeToName = new Map(regionList?.map((r) => [r.code, r.name]) ?? []);

  // ── 4-year trend for this region + area ──────────────────────────────────
  const trendPoints: { year: number; value: number }[] = historySnaps
    .filter((s: RegionBudgetSnapshot) => s.area_name === areaName)
    .map((s: RegionBudgetSnapshot) => ({ year: s.year, value: s.value_mnkr }))
    .sort((a, b) => a.year - b.year);

  // ── Cross-region comparison (% of total budget) ──────────────────────────
  const sortedPoints = [...areaPoints].sort((a, b) => b.pct - a.pct);
  const mean =
    areaPoints.length > 0
      ? areaPoints.reduce((s, p) => s + p.pct, 0) / areaPoints.length
      : 0;
  const maxValue = sortedPoints[0]?.pct ?? 1;

  const hbarItems = sortedPoints.map((p) => ({
    name: codeToName.get(p.region_code) ?? p.region_code,
    value: Math.round(p.pct * 10) / 10,
    color:
      p.region_code === code
        ? "var(--color-accent, #0b3d7a)"
        : "var(--color-fg-muted, #7a8390)",
  }));

  if (isLoading) return <Skeleton />;

  const pad = isMobile ? "14px" : "32px";

  return (
    <div className="sdt-page">
      {/* ── Breadcrumb ──────────────────────────────────────────────── */}
      <div
        style={{
          padding: isMobile ? "14px 14px 0" : "28px 32px 0",
          display: "flex",
          alignItems: "center",
          gap: 8,
        }}
      >
        <Link
          to={`/region/${code}`}
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            letterSpacing: "0.1em",
            color: "var(--color-accent)",
            textDecoration: "none",
            textTransform: "uppercase",
          }}
        >
          ← {regionName}
        </Link>
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-fg-muted)",
            letterSpacing: "0.08em",
          }}
        >
          / {areaName}
        </span>
      </div>

      {/* ── Section: 4-year trend ────────────────────────────────────── */}
      <div
        style={{
          border: "1px solid var(--color-border)",
          margin: isMobile ? "14px 14px 0" : "20px 32px 0",
          background: "var(--color-sdt-surface)",
          padding: isMobile ? "14px" : "20px 24px",
        }}
      >
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "0.15em",
            color: "var(--color-fg-muted)",
            marginBottom: 4,
            textTransform: "uppercase",
          }}
        >
          {areaName} · {regionName}
        </div>
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 9,
            letterSpacing: "0.08em",
            color: "var(--color-fg-muted)",
            marginBottom: 16,
            opacity: 0.7,
          }}
        >
          Nettokostnad mnkr — 4-årsperiod
        </div>

        {trendPoints.length === 0 ? (
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--color-fg-muted)",
              padding: "20px 0",
              textAlign: "center",
            }}
          >
            Data saknas
          </div>
        ) : (
          <div style={{ maxWidth: 520 }}>
            <MiniLineChart points={trendPoints} />
          </div>
        )}
      </div>

      {/* ── Section: Cross-region comparison ────────────────────────── */}
      <div
        style={{
          border: "1px solid var(--color-border)",
          borderTop: "none",
          margin: isMobile ? "0 14px 28px" : "0 32px 28px",
          background: "var(--color-sdt-surface)",
          padding: isMobile ? "14px" : "20px 24px",
        }}
      >
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "0.15em",
            color: "var(--color-fg-muted)",
            marginBottom: 4,
            textTransform: "uppercase",
          }}
        >
          Jämförelse med alla regioner
        </div>
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 9,
            letterSpacing: "0.08em",
            color: "var(--color-fg-muted)",
            marginBottom: 20,
            opacity: 0.7,
          }}
        >
          % av total budget, senaste år · {regionName} markerad
        </div>

        {areaPoints.length === 0 ? (
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--color-fg-muted)",
              padding: "20px 0",
              textAlign: "center",
            }}
          >
            Data saknas
          </div>
        ) : (
          <>
            <HBars
              items={hbarItems}
              max={maxValue}
              height={6}
              gap={8}
              formatValue={(v) => `${v.toFixed(1)}%`}
            />

            {/* Mean annotation */}
            <div
              style={{
                marginTop: 16,
                paddingTop: 12,
                borderTop: "1px dashed var(--color-border)",
                display: "flex",
                alignItems: "center",
                gap: 8,
              }}
            >
              <span
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 10,
                  color: "var(--color-fg-muted)",
                  letterSpacing: "0.08em",
                }}
              >
                MEDELVÄRDE:
              </span>
              <span
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 10,
                  color: "var(--color-fg)",
                  fontVariantNumeric: "tabular-nums",
                }}
              >
                {mean.toFixed(1)}%
              </span>
              {/* Visual mean bar */}
              <div
                style={{
                  flex: 1,
                  position: "relative",
                  height: 4,
                  background: "var(--color-track)",
                  borderRadius: 2,
                  overflow: "hidden",
                }}
              >
                <div
                  style={{
                    position: "absolute",
                    left: 0,
                    top: 0,
                    height: "100%",
                    width: `${(mean / maxValue) * 100}%`,
                    background: "var(--color-fg-muted)",
                    opacity: 0.4,
                    borderRadius: 2,
                  }}
                />
              </div>
            </div>
          </>
        )}
      </div>

      {/* ── Legend ──────────────────────────────────────────────────── */}
      <div
        style={{
          padding: `0 ${pad} 16px`,
          display: "flex",
          gap: 16,
          flexWrap: "wrap",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
          <span
            style={{
              display: "inline-block",
              width: 12,
              height: 4,
              borderRadius: 2,
              background: "var(--color-accent, #0b3d7a)",
            }}
          />
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.08em",
            }}
          >
            {regionName.toUpperCase()}
          </span>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
          <span
            style={{
              display: "inline-block",
              width: 12,
              height: 4,
              borderRadius: 2,
              background: "var(--color-fg-muted, #7a8390)",
            }}
          />
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.08em",
            }}
          >
            ÖVRIGA REGIONER
          </span>
        </div>
      </div>
    </div>
  );
}
