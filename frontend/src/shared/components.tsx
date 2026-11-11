import { PARTY_COLORS } from "./design";

// ── PartyBadge ────────────────────────────────────────────────────────
export function PartyBadge({ party, size = "sm" }: { party: string; size?: "sm" | "lg" }) {
  const c = PARTY_COLORS[party] ?? { bg: "#666", text: "#fff" };
  const cls = size === "lg" ? "px-3.5 py-1 text-sm" : "px-2.5 py-0.5 text-xs";
  return (
    <span
      className={`${cls} rounded font-bold tracking-wider inline-block`}
      style={{ background: c.bg, color: c.text }}
    >
      {party}
    </span>
  );
}
