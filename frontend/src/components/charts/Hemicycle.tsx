interface HemicycleGroup {
  color: string;
  count: number;
}

interface HemicycleProps {
  groups: HemicycleGroup[];
  width?: number;
  height?: number;
}

export function Hemicycle({ groups, width = 320, height = 140 }: HemicycleProps) {
  const ROWS = 6;
  const total = groups.reduce((sum, g) => sum + g.count, 0);

  if (total === 0) {
    return <svg width={width} height={height} aria-hidden="true" />;
  }

  // Build row seat counts
  const rawRowCounts: number[] = [];
  for (let r = 0; r < ROWS; r++) {
    rawRowCounts.push(Math.round((total / ROWS) * (0.7 + (0.6 * (r + 1)) / ROWS)));
  }

  const rawSum = rawRowCounts.reduce((a, b) => a + b, 0);
  const scale = total / rawSum;
  const adjustedRowCounts: number[] = rawRowCounts.map((c) =>
    Math.max(1, Math.round(c * scale))
  );

  // Flatten all groups into a flat color array
  const flatColors: string[] = [];
  for (const g of groups) {
    for (let i = 0; i < g.count; i++) {
      flatColors.push(g.color);
    }
  }

  const cx = width / 2;
  const cy = height;
  const rMin = height * 0.35;
  const rMax = height * 0.95;
  const dotR = Math.max(2, Math.min(4.5, width / 90));

  // Assign seats to rows: left-to-right across all rows, filling row by row
  const circles: { x: number; y: number; color: string }[] = [];
  let colorIdx = 0;

  for (let r = 0; r < ROWS; r++) {
    const n = adjustedRowCounts[r];
    const rowR = rMin + ((rMax - rMin) * r) / (ROWS - 1);

    for (let s = 0; s < n; s++) {
      if (colorIdx >= flatColors.length) break;

      const t = n === 1 ? 0.5 : s / (n - 1);
      const angle = Math.PI * (1 - t); // PI to 0 (left to right)
      const x = cx + rowR * Math.cos(angle);
      const y = cy - rowR * Math.sin(angle);

      circles.push({ x, y, color: flatColors[colorIdx] });
      colorIdx++;
    }
  }

  return (
    <svg
      width={width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      role="img"
      aria-labelledby="hemicycle-title hemicycle-desc"
    >
      <title id="hemicycle-title">Parlamentssäten</title>
      <desc id="hemicycle-desc">
        Halvrundsdiagram som visar fördelning av {total} parlamentssäten
      </desc>
      {circles.map((c, i) => (
        <circle key={i} cx={c.x} cy={c.y} r={dotR} fill={c.color} />
      ))}
    </svg>
  );
}
