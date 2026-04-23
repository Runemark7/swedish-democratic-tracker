export interface DonutProps {
  segments: { color: string; value: number; name?: string }[];
  size?: number;
  thickness?: number;
  label?: string;
  sublabel?: string;
  center?: boolean;
}

export function Donut({
  segments,
  size = 180,
  thickness = 28,
  label,
  sublabel,
  center = true,
}: DonutProps) {
  const radius = (size - thickness) / 2;
  const circ = 2 * Math.PI * radius;
  const total = segments.reduce((a, s) => a + s.value, 0);

  const inner = (size - thickness * 2) - 18;
  const maxLabelChars = Math.max(4, String(label ?? '').length);
  const byWidth = inner / (maxLabelChars * 0.58);
  const labelSize = Math.max(11, Math.min(size * 0.16, byWidth));
  const subSize = Math.max(9, Math.min(size * 0.075, labelSize * 0.5));

  let offset = 0;
  const arcs = segments.map((s, i) => {
    const frac = total > 0 ? s.value / total : 0;
    const dash = frac * circ;
    const el = (
      <circle
        key={i}
        cx={size / 2}
        cy={size / 2}
        r={radius}
        fill="none"
        stroke={s.color}
        strokeWidth={thickness}
        strokeDasharray={`${dash} ${circ - dash}`}
        strokeDashoffset={-offset}
      >
        {s.name && <title>{s.name}: {s.value}</title>}
      </circle>
    );
    offset += dash;
    return el;
  });

  const titleText = segments.map(s => `${s.name ?? 'Segment'}: ${s.value}`).join(', ');

  return (
    <div style={{ position: 'relative', width: size, height: size }}>
      <svg
        width={size}
        height={size}
        viewBox={`0 0 ${size} ${size}`}
        style={{ transform: 'rotate(-90deg)' }}
        aria-label={titleText}
        role="img"
      >
        <title>{titleText}</title>
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="var(--color-track)"
          strokeWidth={thickness}
        />
        {arcs}
      </svg>
      {center && (
        <div
          style={{
            position: 'absolute',
            inset: 0,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            textAlign: 'center',
            pointerEvents: 'none',
            padding: thickness,
          }}
        >
          {label && (
            <span
              style={{
                fontSize: labelSize,
                color: 'var(--color-fg)',
                fontVariantNumeric: 'tabular-nums',
                lineHeight: 1.1,
                wordBreak: 'break-all',
              }}
            >
              {label}
            </span>
          )}
          {sublabel && (
            <span
              style={{
                fontSize: subSize,
                color: 'var(--color-fg-muted)',
                textTransform: 'uppercase',
                lineHeight: 1.2,
                marginTop: 2,
              }}
            >
              {sublabel}
            </span>
          )}
        </div>
      )}
    </div>
  );
}
