export interface DualLineSeries {
  label: string;
  color: string;
  unit: string;
  formatValue?: (v: number) => string;
  points: { year: number; value: number }[];
}

interface DualLineProps {
  series: DualLineSeries[];
  height?: number;
}

export function DualLine({ series, height = 110 }: DualLineProps) {
  if (series.length === 0 || series[0].points.length === 0) return null;

  // Collect all years across series for unified X-axis
  const allYears = Array.from(
    new Set(series.flatMap((s) => s.points.map((p) => p.year)))
  ).sort((a, b) => a - b);

  const padLeft = 0;
  const padRight = 80;  // room for end-of-line value labels
  const padTop = 8;
  const padBottom = 22; // room for year labels
  const chartW = 600;   // SVG viewBox width (scales with container)
  const chartH = height;
  const plotW = chartW - padLeft - padRight;
  const plotH = chartH - padTop - padBottom;

  const xPos = (year: number) => {
    const idx = allYears.indexOf(year);
    return padLeft + (idx / (allYears.length - 1)) * plotW;
  };

  const normalize = (s: DualLineSeries) => {
    const vals = s.points.map((p) => p.value);
    const min = Math.min(...vals);
    const max = Math.max(...vals);
    return (v: number) =>
      padTop + plotH - ((v - min) / (max - min || 1)) * plotH;
  };

  const fmt = (s: DualLineSeries, v: number) =>
    s.formatValue ? s.formatValue(v) : `${v}${s.unit}`;

  return (
    <div>
      <svg
        viewBox={`0 0 ${chartW} ${chartH + padBottom}`}
        style={{ width: "100%", display: "block", overflow: "visible" }}
        aria-label="Historikdiagram"
      >
        {/* Vertical grid lines per year */}
        {allYears.map((yr) => (
          <line
            key={yr}
            x1={xPos(yr)}
            x2={xPos(yr)}
            y1={padTop}
            y2={padTop + plotH}
            stroke="var(--color-track)"
            strokeWidth={1}
          />
        ))}

        {/* Series lines + dots + end labels */}
        {series.map((s) => {
          const yPos = normalize(s);
          const pts = s.points
            .slice()
            .sort((a, b) => a.year - b.year);
          const polyPoints = pts
            .map((p) => `${xPos(p.year)},${yPos(p.value)}`)
            .join(" ");
          const last = pts[pts.length - 1];

          return (
            <g key={s.label}>
              <polyline
                points={polyPoints}
                fill="none"
                stroke={s.color}
                strokeWidth={2}
                strokeLinejoin="round"
                strokeLinecap="round"
              />
              {pts.map((p) => (
                <circle
                  key={p.year}
                  cx={xPos(p.year)}
                  cy={yPos(p.value)}
                  r={2.5}
                  fill={s.color}
                />
              ))}
              {/* End label */}
              <text
                x={xPos(last.year) + 8}
                y={yPos(last.value) + 4}
                fill={s.color}
                fontSize={10}
                fontFamily="var(--font-mono)"
              >
                {fmt(s, last.value)}
              </text>
            </g>
          );
        })}

        {/* X-axis year labels — show first, last, and every ~3rd */}
        {allYears.map((yr, i) => {
          const show =
            i === 0 ||
            i === allYears.length - 1 ||
            i % 3 === 0;
          if (!show) return null;
          return (
            <text
              key={yr}
              x={xPos(yr)}
              y={padTop + plotH + 14}
              textAnchor="middle"
              fill="var(--color-fg-muted)"
              fontSize={9}
              fontFamily="var(--font-mono)"
            >
              {yr}
            </text>
          );
        })}
      </svg>

      {/* Legend */}
      <div style={{ display: "flex", gap: 16, marginTop: 6 }}>
        {series.map((s) => (
          <div key={s.label} style={{ display: "flex", alignItems: "center", gap: 5 }}>
            <div style={{ width: 14, height: 3, background: s.color, borderRadius: 2 }} />
            <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)" }}>
              {s.label}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
