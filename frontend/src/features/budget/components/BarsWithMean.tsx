export interface BarsWithMeanProps {
  points: { region_code: string; pct: number }[];
  mean: number;
  maxValue: number;
  highlightCode: string;
  codeToName: Map<string, string>;
  isMobile: boolean;
}

export function BarsWithMean({
  points,
  mean,
  maxValue,
  highlightCode,
  codeToName,
  isMobile,
}: BarsWithMeanProps) {
  const LABEL_W = isMobile ? 88 : 140;
  const VAL_W = 44;
  const FIXED_W = LABEL_W + VAL_W;
  const BAR_H = 6;
  const ROW_GAP = 9;

  const meanPct = (mean / maxValue) * 100;
  // mean line left = fixedW * (1 - meanPct/100) + meanPct%
  // (see comment: derived so left == fixedW when meanPct==0, == 100% when meanPct==100)
  const meanLineLeft = `calc(${FIXED_W * (1 - meanPct / 100)}px + ${meanPct}%)`;

  return (
    <div style={{ position: "relative", paddingTop: 20 }}>
      {/* Mean label above the line */}
      <div
        style={{
          position: "absolute",
          left: meanLineLeft,
          top: 0,
          transform: "translateX(-50%)",
          fontFamily: "var(--font-mono)",
          fontSize: 8,
          color: "#d97706",
          letterSpacing: "0.06em",
          whiteSpace: "nowrap",
          pointerEvents: "none",
        }}
      >
        MEDEL {mean.toFixed(1)}%
      </div>

      {/* Vertical mean line spanning all rows */}
      <div
        style={{
          position: "absolute",
          left: meanLineLeft,
          top: 20,
          bottom: 0,
          width: 0,
          borderLeft: "1.5px dashed #d97706",
          pointerEvents: "none",
          zIndex: 1,
        }}
      />

      {/* Bar rows */}
      {points.map((p) => {
        const name = codeToName.get(p.region_code) ?? p.region_code;
        const value = Math.round(p.pct * 10) / 10;
        const isHighlighted = p.region_code === highlightCode;
        const barColor = isHighlighted
          ? "var(--color-accent, #0b3d7a)"
          : "var(--color-fg-muted, #7a8390)";
        const barW = `${Math.min((p.pct / maxValue) * 100, 100)}%`;
        return (
          <div
            key={p.region_code}
            style={{ display: "flex", alignItems: "center", marginBottom: ROW_GAP }}
          >
            <div
              style={{
                width: LABEL_W,
                fontFamily: "var(--font-mono)",
                fontSize: isMobile ? 9 : 10,
                color: isHighlighted ? "var(--color-fg)" : "var(--color-fg-muted)",
                fontWeight: isHighlighted ? 600 : 400,
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
                paddingRight: 6,
                flexShrink: 0,
              }}
            >
              {name}
            </div>
            <div
              style={{
                width: VAL_W,
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                color: "var(--color-fg-muted)",
                textAlign: "right",
                paddingRight: 8,
                fontVariantNumeric: "tabular-nums",
                flexShrink: 0,
              }}
            >
              {value.toFixed(1)}%
            </div>
            <div
              style={{
                flex: 1,
                height: BAR_H,
                background: "var(--color-track, #e5e7eb)",
                borderRadius: BAR_H / 2,
                overflow: "hidden",
              }}
            >
              <div
                style={{
                  width: barW,
                  height: "100%",
                  background: barColor,
                  borderRadius: BAR_H / 2,
                }}
              />
            </div>
          </div>
        );
      })}
    </div>
  );
}
