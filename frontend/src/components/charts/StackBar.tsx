interface Segment {
  color: string;
  value: number;
  name?: string;
  short?: string;
}

interface StackBarProps {
  segments: Segment[];
  height?: number;
  rounded?: boolean;
  showLabels?: boolean;
}

export function StackBar({
  segments,
  height = 14,
  rounded = true,
  showLabels = false,
}: StackBarProps) {
  const total = segments.reduce((sum, s) => sum + s.value, 0);
  const borderRadius = rounded ? height / 2 : 2;

  return (
    <div style={{ width: '100%' }}>
      <div
        style={{
          display: 'flex',
          width: '100%',
          height,
          borderRadius,
          background: 'var(--color-track)',
          overflow: 'hidden',
        }}
      >
        {segments.map((seg, i) => {
          const widthPct = total > 0 ? (seg.value / total) * 100 : 0;
          return (
            <div
              key={i}
              title={seg.name ?? seg.short}
              style={{
                width: `${widthPct}%`,
                background: seg.color,
                transition: 'width 0.4s',
                flexShrink: 0,
              }}
            />
          );
        })}
      </div>

      {showLabels && (
        <div
          style={{
            display: 'flex',
            flexWrap: 'wrap',
            gap: '4px 8px',
            marginTop: '6px',
          }}
        >
          {segments.map((seg, i) => (
            <span
              key={i}
              style={{
                fontSize: '11px',
                color: seg.color,
                fontVariantNumeric: 'tabular-nums',
              }}
            >
              {seg.short ?? seg.name} {seg.value}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}
