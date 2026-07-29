import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { PARTY_COLORS, SPECIFICITY_CONFIG, TOPIC_LABELS, alignmentColor, deltaColor, formatBudgetAmount } from "./design";
import { DeltaIndicator } from "@/features/budget/components/DeltaIndicator";
import { contextApi } from "@/features/context/api";
import type { BudgetTrendEntry, FundingRow, RelatedBudgetArea } from "./types";

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

// ── AlignmentRing (SVG donut) ─────────────────────────────────────────
export function AlignmentRing({ pct, size = 56, color }: { pct: number; size?: number; color?: string }) {
  const r = (size - 6) / 2;
  const circ = 2 * Math.PI * r;
  const offset = circ - (pct / 100) * circ;
  const ringColor = color ?? alignmentColor(pct);
  return (
    <div className="relative shrink-0" style={{ width: size, height: size }}>
      <svg width={size} height={size} className="-rotate-90">
        <circle cx={size / 2} cy={size / 2} r={r} fill="none" stroke="var(--color-surface-high)" strokeWidth="4" />
        <circle
          cx={size / 2} cy={size / 2} r={r} fill="none" stroke={ringColor}
          strokeWidth="4" strokeDasharray={circ} strokeDashoffset={offset}
          strokeLinecap="round" className="transition-all duration-700 ease-out"
        />
      </svg>
      <div
        className="absolute inset-0 flex items-center justify-center font-mono font-extrabold"
        style={{ fontSize: size * 0.23, color: ringColor }}
      >
        {Math.round(pct)}%
      </div>
    </div>
  );
}

// ── StatBlock ─────────────────────────────────────────────────────────
export function StatBlock({ value, label, accent, small }: {
  value: string | number;
  label: string;
  accent?: string;
  small?: boolean;
}) {
  return (
    <div className="text-center">
      <div
        className={`font-mono font-extrabold ${small ? "text-xl" : "text-3xl"}`}
        style={{ color: accent ?? "var(--color-on-surface)" }}
      >
        {value}
      </div>
      <div className="text-[10px] text-on-surface-variant uppercase tracking-widest mt-0.5">
        {label}
      </div>
    </div>
  );
}

// ── TopicTag ──────────────────────────────────────────────────────────
export function TopicTag({ topic }: { topic: string }) {
  return (
    <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-surface-low text-on-surface-variant capitalize">
      {TOPIC_LABELS[topic] ?? topic}
    </span>
  );
}

// ── SpecificityBadge ──────────────────────────────────────────────────
export function SpecificityBadge({ specificity }: { specificity: string }) {
  const cfg = SPECIFICITY_CONFIG[specificity] ?? { label: specificity, color: "#9ca3af" };
  return (
    <span
      className="px-2 py-0.5 rounded text-[10px] font-semibold"
      style={{ background: `${cfg.color}15`, color: cfg.color }}
    >
      {cfg.label}
    </span>
  );
}

// ── ProposalOriginTag ─────────────────────────────────────────────────
export function ProposalOriginTag({ proposedBy, proposalType }: {
  proposedBy?: string;
  proposalType?: string;
}) {
  if (!proposedBy && !proposalType) return null;
  const partyColor = proposedBy ? PARTY_COLORS[proposedBy] : null;
  const typeLabel = proposalType === "prop" ? "Proposition" : proposalType === "mot" ? "Motion" : proposalType === "bet" ? "Betänkande" : null;

  return (
    <div className="flex items-center gap-2 px-2.5 py-1.5 rounded-md bg-surface-low text-[11px]">
      <span className="text-on-surface-variant">Initierat av:</span>
      {partyColor ? (
        <span
          className="px-1.5 py-px rounded text-[10px] font-bold"
          style={{ background: partyColor.bg, color: partyColor.text }}
        >
          {proposedBy}
        </span>
      ) : (
        <span className="font-semibold text-on-surface">{proposedBy ?? "Okänt"}</span>
      )}
      {typeLabel && (
        <span className="px-1.5 py-px rounded bg-surface-high text-on-surface-variant text-[10px] font-semibold">
          {typeLabel}
        </span>
      )}
    </div>
  );
}

// ── ConsequencePanel ──────────────────────────────────────────────────
export function ConsequencePanel({ topic }: { topic: string }) {
  const [open, setOpen] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ["context", topic],
    queryFn: () => contextApi.getTopic(topic),
    enabled: open,
    staleTime: 10 * 60 * 1000,
  });

  return (
    <div className="mt-3 rounded-lg overflow-hidden border border-surface-high">
      <button
        onClick={() => setOpen((o) => !o)}
        className="w-full flex items-center justify-between px-4 py-2.5 bg-surface-low text-xs font-semibold text-on-surface-variant hover:bg-surface-high transition-colors"
      >
        <span>Visa konsekvenser</span>
        <span className="transition-transform duration-200" style={{ transform: open ? "rotate(180deg)" : "none" }}>
          ▾
        </span>
      </button>

      {open && (
        <div className="px-4 py-4 space-y-5">
          {isLoading && (
            <p className="text-xs text-on-surface-variant text-center py-2">Laddar budgetdata…</p>
          )}

          {data && (
            <>
              {/* Description + committee chips */}
              <div>
                <p className="text-xs text-on-surface-variant leading-relaxed mb-2">{data.description}</p>
                <div className="flex flex-wrap gap-1.5">
                  {data.committees.map((c: string) => (
                    <span key={c} className="px-2 py-0.5 rounded bg-surface-high text-[10px] font-semibold text-on-surface-variant">
                      {c}
                    </span>
                  ))}
                </div>
              </div>

              {/* Budget area trend bars */}
              {data.relatedAreas.length > 0 && (
                <div>
                  <p className="text-[10px] uppercase tracking-widest text-on-surface-variant font-semibold mb-3">
                    Budgetutveckling
                  </p>
                  <div className="space-y-4">
                    {data.relatedAreas.map((area: RelatedBudgetArea) => {
                      const maxAmount = Math.max(...area.trend.map((t: BudgetTrendEntry) => t.amountKsek), 1);
                      return (
                        <div key={area.code}>
                          <div className="flex items-baseline justify-between mb-1.5">
                            <span className="text-xs font-semibold text-on-surface">{area.name}</span>
                            <div className="flex items-center gap-3">
                              <DeltaIndicator pct={area.latestDeltaPct} showBar={false} />
                              <span className="text-[10px] text-on-surface-variant">
                                {area.shareOfBudgetPct.toFixed(1)}% av statsbudgeten
                              </span>
                            </div>
                          </div>
                          <div className="flex items-end gap-px">
                            {area.trend.map((entry: BudgetTrendEntry) => {
                              const barPx = Math.max(1, Math.round((entry.amountKsek / maxAmount) * 40));
                              return (
                                <div key={entry.year} className="flex-1 flex flex-col items-center gap-0.5">
                                  <div
                                    className="w-full rounded-t transition-all duration-300"
                                    style={{
                                      height: `${barPx}px`,
                                      background: "var(--color-primary)",
                                      opacity: 0.75,
                                    }}
                                  />
                                  <span className="text-[9px] text-on-surface-variant">{entry.year}</span>
                                </div>
                              );
                            })}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}

              {/* Funding context */}
              {data.fundingContext && (
                <div>
                  <p className="text-[10px] uppercase tracking-widest text-on-surface-variant font-semibold mb-2">
                    Finansiering {data.fundingContext.year}
                  </p>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <p className="text-[10px] text-on-surface-variant mb-1.5">Ökade mest</p>
                      <div className="space-y-1">
                        {data.fundingContext.topIncreases.map((row: FundingRow) => (
                          <div key={row.code} className="flex items-center justify-between text-[11px]">
                            <span className="text-on-surface truncate mr-2">{row.name}</span>
                            <span className="font-mono font-bold shrink-0" style={{ color: deltaColor(row.deltaPct) }}>
                              +{formatBudgetAmount(row.deltaKsek)}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                    <div>
                      <p className="text-[10px] text-on-surface-variant mb-1.5">Minskade mest</p>
                      <div className="space-y-1">
                        {data.fundingContext.topDecreases.map((row: FundingRow) => (
                          <div key={row.code} className="flex items-center justify-between text-[11px]">
                            <span className="text-on-surface truncate mr-2">{row.name}</span>
                            <span className="font-mono font-bold shrink-0" style={{ color: deltaColor(row.deltaPct) }}>
                              {formatBudgetAmount(row.deltaKsek)}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                  {data.fundingContext.totalBudgetDeltaPct !== 0 && (
                    <div className="mt-3 pt-2.5 border-t border-surface-high flex items-center gap-2 text-[11px] text-on-surface-variant">
                      <span>Total budgetförändring:</span>
                      <DeltaIndicator pct={data.fundingContext.totalBudgetDeltaPct} showBar={false} />
                    </div>
                  )}
                </div>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
}

// ── VoteBar (stacked horizontal bar) ──────────────────────────────────
export function VoteBar({ ja, nej, avstar, franvarande, total, partyColor }: {
  ja: number; nej: number; avstar: number; franvarande: number;
  total: number; partyColor?: string;
}) {
  const pct = (n: number) => total ? `${(n / total) * 100}%` : "0%";
  return (
    <div className="space-y-1.5">
      <div className="flex h-5 rounded overflow-hidden bg-surface-high">
        {ja > 0 && (
          <div className="flex items-center justify-center" style={{ width: pct(ja), background: partyColor ?? "#16a34a" }}>
            {ja / total > 0.1 && <span className="text-[10px] font-bold text-white">{ja}</span>}
          </div>
        )}
        {nej > 0 && (
          <div className="flex items-center justify-center bg-red-400" style={{ width: pct(nej) }}>
            {nej / total > 0.1 && <span className="text-[10px] font-bold text-white">{nej}</span>}
          </div>
        )}
        {avstar > 0 && (
          <div className="flex items-center justify-center bg-amber-400" style={{ width: pct(avstar) }}>
            {avstar / total > 0.1 && <span className="text-[10px] font-bold text-on-surface">{avstar}</span>}
          </div>
        )}
        {franvarande > 0 && (
          <div className="flex items-center justify-center bg-surface-highest" style={{ width: pct(franvarande) }}>
            {franvarande / total > 0.1 && <span className="text-[10px] font-bold text-on-surface-variant">{franvarande}</span>}
          </div>
        )}
      </div>
      <div className="flex gap-3.5 flex-wrap">
        {[
          { label: "Ja", count: ja, color: partyColor ?? "#16a34a" },
          { label: "Nej", count: nej, color: "#ef4444" },
          { label: "Avstår", count: avstar, color: "#fbbf24" },
          { label: "Frånv.", count: franvarande, color: "#d1d5db" },
        ].map(item => (
          <div key={item.label} className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-sm" style={{ background: item.color }} />
            <span className="text-[10px] text-on-surface-variant">{item.label}: {item.count}/{total}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
