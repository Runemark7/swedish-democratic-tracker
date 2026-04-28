type TrendDirection = 'up' | 'down' | 'flat';

interface TrendProps {
  trend: TrendDirection;
  delta: string;
  worseHigher?: boolean;
}

const trendSymbol: Record<TrendDirection, string> = {
  up: '↑', down: '↓', flat: '→',
};

export function Trend({ trend, delta, worseHigher = false }: TrendProps) {
  const symbol = trendSymbol[trend];
  const isGoodChange =
    trend === 'flat' ? null :
    trend === 'up' ? !worseHigher :
    worseHigher;
  const color =
    isGoodChange === null ? 'var(--color-fg-muted)' :
    isGoodChange ? 'var(--color-up)' : 'var(--color-down)';

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
