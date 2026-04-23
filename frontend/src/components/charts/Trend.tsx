type TrendDirection = 'up' | 'down' | 'flat';

interface TrendProps {
  trend: TrendDirection;
  delta: string;
}

const trendConfig: Record<TrendDirection, { symbol: string; color: string }> = {
  up: { symbol: '↑', color: 'var(--color-up)' },
  down: { symbol: '↓', color: 'var(--color-down)' },
  flat: { symbol: '→', color: 'var(--color-fg-muted)' },
};

export function Trend({ trend, delta }: TrendProps) {
  const { symbol, color } = trendConfig[trend];

  return (
    <span
      style={{
        fontSize: '12px',
        fontVariantNumeric: 'tabular-nums',
        color,
      }}
    >
      {symbol} {delta}
    </span>
  );
}
