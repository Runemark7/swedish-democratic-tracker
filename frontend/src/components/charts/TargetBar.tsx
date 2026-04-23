export interface TargetBarProps {
  value: number;
  target: number;
  min?: number;
  max?: number;
  worseHigher?: boolean;
  unit?: string;
  height?: number;
}

export function TargetBar({
  value,
  target,
  min,
  max,
  worseHigher = false,
  unit = '',
  height = 30,
}: TargetBarProps) {
  const lo = min != null ? min : Math.min(value, target) * 0.85;
  const hi = max != null ? max : Math.max(value, target) * 1.15;

  const pct = (v: number) =>
    Math.max(0, Math.min(1, (v - lo) / (hi - lo))) * 100;

  const goodStart = worseHigher ? 0 : pct(target);
  const goodEnd = worseHigher ? pct(target) : 100;
  const isGood = worseHigher ? value <= target : value >= target;

  const valuePct = pct(value);
  const targetPct = pct(target);

  const markerSize = height * 0.6;
  const trackHeight = Math.round(height * 0.28);
  const trackTop = (height - trackHeight) / 2;

  const descriptiveText = `${isGood ? 'On track' : 'Below target'}: current value ${value}${unit}, target ${target}${unit}`;

  return (
    <div
      role="meter"
      aria-valuenow={value}
      aria-valuemin={lo}
      aria-valuemax={hi}
      aria-label={descriptiveText}
      style={{ position: 'relative', width: '100%', height }}
    >
      {/* Visually hidden descriptive text */}
      <span
        style={{
          position: 'absolute',
          width: 1,
          height: 1,
          padding: 0,
          margin: -1,
          overflow: 'hidden',
          clip: 'rect(0,0,0,0)',
          whiteSpace: 'nowrap',
          border: 0,
        }}
      >
        {descriptiveText}
      </span>

      {/* Baseline track */}
      <div
        style={{
          position: 'absolute',
          left: 0,
          right: 0,
          top: trackTop,
          height: trackHeight,
          borderRadius: trackHeight / 2,
          background: 'var(--color-track)',
          overflow: 'hidden',
        }}
      >
        {/* Good-range highlight */}
        <div
          style={{
            position: 'absolute',
            top: 0,
            bottom: 0,
            left: `${goodStart}%`,
            width: `${goodEnd - goodStart}%`,
            background: 'color-mix(in oklch, var(--color-up) 38%, transparent)',
          }}
        />
      </div>

      {/* Target tick */}
      <div
        style={{
          position: 'absolute',
          top: trackTop - 2,
          bottom: trackTop - 2,
          left: `${targetPct}%`,
          width: 2,
          marginLeft: -1,
          background: 'var(--color-fg-muted)',
          height: trackHeight + 4,
        }}
      />

      {/* Current value marker (circle) */}
      <div
        style={{
          position: 'absolute',
          top: '50%',
          left: `${valuePct}%`,
          width: markerSize,
          height: markerSize,
          marginLeft: -(markerSize / 2),
          marginTop: -(markerSize / 2),
          borderRadius: '50%',
          background: isGood ? 'var(--color-up)' : 'var(--color-down)',
          border: '2px solid var(--color-bg)',
          boxSizing: 'border-box',
        }}
      />

      {/* Scale labels */}
      <div
        style={{
          position: 'absolute',
          top: height,
          left: 0,
          right: 0,
          display: 'flex',
          justifyContent: 'space-between',
          fontSize: 10,
          fontVariantNumeric: 'tabular-nums',
          color: 'var(--color-fg-muted)',
          paddingTop: 2,
          pointerEvents: 'none',
        }}
      >
        <span>{lo}{unit}</span>
        <span style={{ color: 'var(--color-fg)', fontWeight: 'bold' }}>
          {target}{unit}
        </span>
        <span>{hi}{unit}</span>
      </div>
    </div>
  );
}
