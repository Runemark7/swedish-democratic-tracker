export interface GoalBadgeProps {
  raw: number;
  target: number;
  unit: string;
  worseHigher?: boolean;
  note?: string;
  sourceUrl?: string;
}

export function GoalBadge({ raw, target, unit, worseHigher = false, note, sourceUrl }: GoalBadgeProps) {
  const gap = Math.abs(raw - target);
  const onTarget = gap < 0.05;
  const isGood = worseHigher ? raw <= target : raw >= target;
  const over = raw > target;
  const direction = over ? "över" : "under";
  const gapLabel = unit === "%" ? "pp" : unit;
  const label = onTarget
    ? "På mål"
    : `${gap % 1 === 0 ? gap : gap.toFixed(1)} ${gapLabel} ${direction} mål`;

  return (
    <div style={{ marginTop: 10 }}>
      <div
        style={{
          display: "inline-flex",
          alignItems: "center",
          gap: 6,
          padding: "4px 10px",
          borderRadius: 4,
          background: onTarget
            ? "color-mix(in oklch, var(--color-fg-muted) 12%, transparent)"
            : isGood
              ? "color-mix(in oklch, var(--color-up) 14%, transparent)"
              : "color-mix(in oklch, var(--color-down) 14%, transparent)",
          color: onTarget
            ? "var(--color-fg-muted)"
            : isGood
              ? "var(--color-up)"
              : "var(--color-down)",
          fontFamily: "var(--font-mono)",
          fontSize: 11,
          letterSpacing: "0.04em",
        }}
      >
        <span style={{ fontSize: 8 }}>●</span>
        {label}
      </div>
      {sourceUrl && note && (
        <a
          href={sourceUrl}
          target="_blank"
          rel="noopener noreferrer"
          style={{
            display: "block",
            marginTop: 6,
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "0.04em",
            color: "var(--color-fg-muted)",
            textDecoration: "none",
          }}
        >
          {note} →
        </a>
      )}
    </div>
  );
}
