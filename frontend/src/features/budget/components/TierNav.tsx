import { TIER_LABELS } from "@/shared/design";
import type { BudgetTier } from "@/shared/types";

const TIERS: { key: BudgetTier; enabled: boolean }[] = [
  { key: "national", enabled: true },
];

export function TierNav({ active }: { active: BudgetTier }) {
  return (
    <div className="flex gap-1 mb-6">
      {TIERS.map((tier) => {
        const isActive = tier.key === active;
        return (
          <button
            key={tier.key}
            disabled={!tier.enabled}
            className="px-4 py-2 text-[13px] font-semibold rounded-lg transition-all"
            style={{
              background: isActive
                ? "var(--color-primary)"
                : tier.enabled
                  ? "var(--color-surface-low)"
                  : "var(--color-surface-high)",
              color: isActive
                ? "var(--color-on-primary)"
                : tier.enabled
                  ? "var(--color-on-surface)"
                  : "var(--color-on-surface-variant)",
              opacity: tier.enabled ? 1 : 0.5,
              cursor: tier.enabled ? "pointer" : "not-allowed",
            }}
            title={!tier.enabled ? "Kommer snart" : undefined}
          >
            {TIER_LABELS[tier.key]}
            {!tier.enabled && (
              <span className="ml-1.5 text-[10px] opacity-70">snart</span>
            )}
          </button>
        );
      })}
    </div>
  );
}
