import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams, Link } from "react-router-dom";
import { partiesApi } from "./api";
import {
  AlignmentRing, PartyBadge, StatBlock,
  SpecificityBadge, TopicTag, ConsequencePanel,
} from "@/shared/components";
import { PARTY_COLORS, TOPIC_LABELS } from "@/shared/design";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { useRecordCoverage } from "@/hooks/useDemocracy";
import type { GoalWithAlignment } from "@/shared/types";

function GoalCard({ goal, party }: { goal: GoalWithAlignment; party: string }) {
  const [open, setOpen] = useState(false);
  const matched = goal.relevantVotes ?? 0;
  const scored = goal.scoredVotes ?? 0;
  const aligned = goal.alignedVotes ?? 0;
  const pct = goal.alignmentPct;

  return (
    <div
      className="rounded-xl overflow-hidden transition-all"
      style={{
        background: open ? "var(--color-surface-low)" : "var(--color-surface-lowest)",
        borderLeft: "3px solid var(--color-surface-high)",
      }}
    >
      {/* Header (clickable) */}
      <button
        onClick={() => setOpen(!open)}
        className="w-full text-left px-5 py-4 cursor-pointer"
      >
        <div className="flex justify-between items-start gap-4">
          <div className="flex-1 min-w-0">
            <div className="flex gap-2 items-center flex-wrap mb-2">
              <SpecificityBadge specificity={goal.specificity} />
              <TopicTag topic={goal.topic} />
            </div>
            <p className="text-sm font-semibold text-on-surface leading-snug">
              &ldquo;{goal.goalText}&rdquo;
            </p>
            <div className="flex items-center gap-3 mt-2 text-xs text-on-surface-variant flex-wrap">
              <span>{goal.sourceDocument} <SourceMarker sourceId="seed-party-goals" /></span>
              {scored > 0 ? (
                <span>{aligned} av {scored} relevanta röster i linje <SourceMarker sourceId="riksdagen" /></span>
              ) : matched > 0 ? (
                <span>
                  {matched === 1
                    ? "1 omröstning matchad, riktning ej fastställd"
                    : `${matched} omröstningar matchade, riktning ej fastställd`}{" "}
                  <SourceMarker sourceId="riksdagen" />
                </span>
              ) : (
                <span>Ej prövad i någon omröstning ännu <SourceMarker sourceId="riksdagen" /></span>
              )}
            </div>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            {pct !== null && pct !== undefined ? (
              <AlignmentRing pct={pct} />
            ) : (
              <span className="text-sm text-on-surface-variant" title="Riktning ej fastställd">—</span>
            )}
            <span
              className="text-lg text-on-surface-variant transition-transform duration-300"
              style={{ transform: open ? "rotate(180deg)" : "rotate(0deg)" }}
            >
              ▾
            </span>
          </div>
        </div>
      </button>

      {/* Expanded detail */}
      {open && (
        <div className="px-5 pb-5">
          <div className="h-px bg-surface-high mb-4" />
          {matched > 0 ? (
            <div>
              <p className="text-[11px] text-on-surface-variant uppercase tracking-widest font-semibold mb-3">
                Så röstade partiets ledamöter
              </p>
              <Link
                to={`/parties/${party}/goals/${goal.id}/votes`}
                className="inline-block text-xs text-primary font-semibold hover:underline mb-3"
              >
                Visa alla omröstningar →
              </Link>
            </div>
          ) : (
            <p className="text-sm text-on-surface-variant text-center py-4">
              Inga relevanta omröstningar har matchats till detta mål ännu.
            </p>
          )}
          <ConsequencePanel topic={goal.topic} />
        </div>
      )}
    </div>
  );
}

export function PartyGoalsPage() {
  const { party = "" } = useParams();
  const pc = PARTY_COLORS[party];

  const { data, isLoading, error } = useQuery({
    queryKey: ["party-goals", party],
    queryFn: () => partiesApi.listGoals(party),
    enabled: !!party,
  });
  // Must sit above the early returns: a hook called conditionally breaks the
  // rules of hooks once isLoading flips.
  const { data: coverage } = useRecordCoverage();

  if (isLoading) return <div className="text-on-surface-variant py-16 text-center text-sm">Laddar mål...</div>;
  if (error) return <div className="text-contradiction py-16 text-center text-sm">Kunde inte ladda mål.</div>;

  const goals = data ?? [];

  // Neutral aggregates only — raw facts, no good/bad bucketing. Goals whose
  // direction could not be determined are excluded from the average rather than
  // counted as 0: that would assert "not in line" where we simply do not know.
  const totalScored = goals.reduce((s, g) => s + (g.scoredVotes ?? 0), 0);
  const scoredGoals = goals.filter(
    (g) => g.alignmentPct !== null && g.alignmentPct !== undefined,
  );
  const avgAlignment = scoredGoals.length
    ? Math.round(
        scoredGoals.reduce((s, g) => s + (g.alignmentPct ?? 0), 0) / scoredGoals.length,
      )
    : null;

  // Group by topic
  const byTopic = goals.reduce<Record<string, GoalWithAlignment[]>>((acc, g) => {
    (acc[g.topic] ??= []).push(g);
    return acc;
  }, {});

  return (
    <div>
      {/* ── Party summary header ───────────────────────────────────── */}
      <div
        className="rounded-xl p-6 mt-3 mb-5"
        style={{ background: pc?.light ?? "#f5f5f5" }}
      >
        <div className="flex justify-between items-center flex-wrap gap-4">
          <div>
            <div className="flex items-center gap-3 mb-1.5">
              <PartyBadge party={party} size="lg" />
              <h2 className="font-display text-[22px] font-extrabold tracking-tight m-0">
                {party}
              </h2>
            </div>
            <p className="text-xs text-on-surface-variant">{goals.length} mål från valmanifest och partiprogram</p>
            {coverage && (
              <p className="text-[11px] text-on-surface-variant mt-1 leading-relaxed">
                {/* The record is bound to a named mandate period so votes from
                    the next parliament are never attributed to this one. */}
                <span className="font-semibold">
                  {coverage.mandate.label}
                  {coverage.mandate.ended ? " (avslutad)" : ""}
                </span>
                {" · "}
                {/* Completeness is a claim, so it is stated rather than implied. */}
                {coverage.ingested.toLocaleString("sv-SE")} av{" "}
                {coverage.expected.toLocaleString("sv-SE")} omröstningar
                <SourceMarker sourceId="riksdagen" />
                {coverage.ingested < coverage.expected && (
                  <>
                    {" "}
                    <Link to="/data" className="underline">
                      Vad saknas?
                    </Link>
                  </>
                )}
              </p>
            )}
          </div>
          <div className="flex gap-7 items-center">
            <div className="text-center">
              {avgAlignment !== null ? (
                <AlignmentRing pct={avgAlignment} size={64} color={pc?.bg} />
              ) : (
                <div
                  className="text-2xl text-on-surface-variant flex items-center justify-center"
                  style={{ width: 64, height: 64 }}
                  title="Ingen omröstning med fastställd riktning ännu"
                >
                  —
                </div>
              )}
              <div className="text-[10px] text-on-surface-variant mt-1 uppercase tracking-wider">Snitt i linje</div>
            </div>
            <StatBlock value={goals.length} label="Mål" small />
            <StatBlock value={totalScored} label="Bedömda röster" small />
          </div>
        </div>
      </div>

      {/* ── Explanation banner ──────────────────────────────────────── */}
      <div className="bg-[#f0f9ff] rounded-lg p-4 mb-5">
        <p className="text-xs text-[#0369a1] leading-relaxed m-0">
          Nedan visas partiets uttalade mål matchade mot relevanta riksdagsomröstningar.
          Varje omröstning visar <strong>vilket parti som initierade förslaget</strong> — ett NEJ-röst
          kan bero på att förslaget kom från ett annat parti med annorlunda villkor.
        </p>
      </div>

      {/* ── Goal cards grouped by topic ────────────────────────────── */}
      {Object.entries(byTopic).map(([topic, gg]) => (
        <section key={topic} className="mb-8">
          <h3 className="font-display text-base font-semibold text-on-surface-variant capitalize mb-3 pb-1">
            {TOPIC_LABELS[topic] ?? topic}
          </h3>
          <div className="space-y-3">
            {gg.map((g) => (
              <GoalCard key={g.id} goal={g} party={party} />
            ))}
          </div>
        </section>
      ))}

      {goals.length === 0 && (
        <p className="text-on-surface-variant text-center py-16 text-sm">
          Inga mål har lagts in för {party} ännu.
        </p>
      )}
    </div>
  );
}
