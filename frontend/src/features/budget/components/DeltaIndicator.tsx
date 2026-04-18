import { deltaColor } from "@/shared/design";

export function DeltaIndicator({ pct, showBar = true }: { pct: number; showBar?: boolean }) {
  const color = deltaColor(pct);
  const arrow = pct > 0 ? "\u25b2" : pct < 0 ? "\u25bc" : "";
  const sign = pct > 0 ? "+" : "";
  const barWidth = Math.min(Math.abs(pct), 30); // Cap visual at 30%

  return (
    <div className="flex items-center gap-2">
      {showBar && (
        <div className="w-16 h-1.5 rounded-full bg-surface-high overflow-hidden">
          <div
            className="h-full rounded-full transition-all duration-500"
            style={{ width: `${(barWidth / 30) * 100}%`, background: color }}
          />
        </div>
      )}
      <span
        className="text-xs font-mono font-bold whitespace-nowrap"
        style={{ color }}
      >
        {sign}{pct.toFixed(1)}% {arrow}
      </span>
    </div>
  );
}
