export interface HBarsProps {
  items: { name: string; value: number; color?: string; pct?: number; amount?: string }[];
  max?: number;
  unit?: string;
  height?: number;
  gap?: number;
}

export function HBars({
  items,
  max,
  unit = '',
  height = 6,
  gap = 10,
}: HBarsProps) {
  const computedMax = max ?? Math.max(...items.map(i => i.value), 1);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap }}>
      {items.map((item, i) => {
        const fraction = Math.max(0, Math.min(1, item.value / computedMax));
        const displayValue =
          item.pct !== undefined
            ? item.amount ? `${item.pct}%  ${item.amount}` : `${item.pct}%`
            : `${item.value}${unit}`;

        return (
          <div key={i}>
            {/* Label row */}
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                fontSize: 12,
                marginBottom: 4,
              }}
            >
              <span style={{ color: 'var(--color-fg)' }}>{item.name}</span>
              <span style={{ color: 'var(--color-fg-muted)', fontVariantNumeric: 'tabular-nums' }}>
                {displayValue}
              </span>
            </div>
            {/* Bar */}
            <div
              style={{
                height,
                borderRadius: height / 2,
                background: 'var(--color-track)',
                overflow: 'hidden',
              }}
            >
              <div
                style={{
                  height: '100%',
                  width: `${fraction * 100}%`,
                  borderRadius: height / 2,
                  background: item.color ?? 'var(--color-accent)',
                  transition: 'width 0.4s ease',
                }}
              />
            </div>
          </div>
        );
      })}
    </div>
  );
}
