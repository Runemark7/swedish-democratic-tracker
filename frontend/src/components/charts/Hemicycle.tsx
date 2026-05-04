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
  const total = groups.reduce((sum, g) => sum + g.count, 0);

  if (total === 0) {
    return (
      <svg
        viewBox={`0 0 ${width} ${height}`}
        style={{ width: "100%", height: "auto" }}
        aria-hidden="true"
      />
    );
  }

  // Scale row count to council size so small councils don't look sparse
  const ROWS = total <= 30 ? 4 : total <= 60 ? 5 : 6;

  // Largest Remainder Method — guarantees adjustedRowCounts sums exactly to total
  const weights = Array.from({ length: ROWS }, (_, r) => 0.7 + (0.6 * (r + 1)) / ROWS);
  const totalWeight = weights.reduce((a, b) => a + b, 0);
  const idealCounts = weights.map((w) => Math.max(1, (w / totalWeight) * total));
  const floorCounts = idealCounts.map((v) => Math.floor(v));
  const remaining = total - floorCounts.reduce((a, b) => a + b, 0);
  const remainders = idealCounts
    .map((v, i) => ({ idx: i, rem: v - floorCounts[i] }))
    .sort((a, b) => b.rem - a.rem);
  for (let i = 0; i < remaining; i++) floorCounts[remainders[i].idx]++;
  const adjustedRowCounts = floorCounts;

  // Flatten all groups into a flat color array
  const flatColors: string[] = [];
  for (const g of groups) {
    for (let i = 0; i < g.count; i++) flatColors.push(g.color);
  }

  const cx = width / 2;
  const cy = height;
  const rMin = height * 0.35;
  const rMax = height * 0.95;
  // Larger dots for small councils so the chart fills the space
  const dotR = ROWS <= 4
    ? Math.max(3, Math.min(6, width / 65))
    : Math.max(2, Math.min(4.5, width / 90));

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
      viewBox={`0 0 ${width} ${height}`}
      style={{ width: "100%", height: "auto" }}
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
