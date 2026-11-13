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

/** Shown when our own API failed for a reason that can pass (rate limit,
 *  server error, dropped connection) and the automatic retries ran out. It
 *  says nothing about whether the record exists. */
export function TemporaryError({ what, onRetry }: { what: string; onRetry: () => void }) {
  return (
    <div role="alert" data-testid="temporary-error" className="text-sm">
      <p>Tillfälligt fel: {what} kunde inte hämtas just nu. Det säger inget om vad som finns i vårt register.</p>
      <button
        type="button"
        onClick={onRetry}
        className="mt-3 underline text-link"
        style={{ background: "none", border: 0, padding: 0, cursor: "pointer", font: "inherit" }}
      >
        Försök igen
      </button>
    </div>
  );
}
