import { useQuery } from "@tanstack/react-query";
import { useParams, Link } from "react-router-dom";
import { partiesApi } from "./api";
import {
  AlignmentRing, PartyBadge, TopicTag, SpecificityBadge,
  ProposalOriginTag,
} from "@/shared/components";
import { PARTY_COLORS, alignmentColor, committeeFromBeteckning } from "@/shared/design";
import type { GoalVoteMatch } from "@/shared/types";

const DIRECTION_STYLES: Record<string, { color: string; bg: string; label: string }> = {
  Ja:          { color: "#16a34a", bg: "#f0fdf4", label: "Ja" },
  Nej:         { color: "#dc2626", bg: "#fef2f2", label: "Nej" },
  "Avstår":    { color: "#d97706", bg: "#fffbeb", label: "Avstår" },
};

function VoteMatchCard({ match }: { match: GoalVoteMatch }) {
  const dir = DIRECTION_STYLES[match.alignedDirection] ?? { color: "#9ca3af", bg: "#f9fafb", label: match.alignedDirection };
  const relevance = Math.round(match.relevanceScore * 100);
  const committee = committeeFromBeteckning(match.beteckning);

  return (
    <Link
      to={`/votes/${match.beteckning}/${match.forslagspunkt}`}
      className="block bg-surface-lowest rounded-xl p-5 transition-all hover:shadow-ambient group"
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <p className="text-sm font-semibold text-on-surface leading-snug group-hover:text-primary transition-colors">
            {match.documentTitle || `${match.beteckning} punkt ${match.forslagspunkt}`}
          </p>

          <div className="flex flex-wrap items-center gap-2 mt-2">
            <span className="text-[11px] font-mono text-on-surface-variant px-2 py-0.5 rounded bg-surface-low">
              {match.beteckning}/{match.forslagspunkt}
            </span>
            {committee && (
              <span className="text-[11px] text-on-surface-variant">· {committee}</span>
            )}
            {(match.proposedByParty || match.proposalType) && (
              <ProposalOriginTag proposedBy={match.proposedByParty} proposalType={match.proposalType} />
            )}
          </div>

          {match.contextNote && (
            <p className="text-xs text-on-surface-variant italic mt-2 leading-relaxed">
              {match.contextNote}
            </p>
          )}
        </div>

        <div className="flex flex-col items-end gap-2 shrink-0">
          <span
            className="text-[10px] font-extrabold px-2.5 py-1 rounded"
            style={{ background: dir.bg, color: dir.color }}
          >
            {dir.label}
          </span>
          <span className="text-[10px] font-mono text-on-surface-variant">
            {relevance}% relevans
          </span>
          {match.partyAlignmentPct != null && (
            <div className="flex items-center gap-1.5">
              <div
                className="w-1.5 h-1.5 rounded-full"
                style={{ background: alignmentColor(match.partyAlignmentPct) }}
              />
              <span className="text-[10px] font-mono font-semibold text-on-surface-variant">
                {match.partyAlignmentPct.toFixed(0)}% röstade rätt
              </span>
            </div>
          )}
        </div>
      </div>
    </Link>
  );
}

export function GoalVotesPage() {
  const { party = "", goalId = "" } = useParams();
  const pc = PARTY_COLORS[party];

  const { data, isLoading, error } = useQuery({
    queryKey: ["goal-votes", party, goalId],
    queryFn: () => partiesApi.getGoalVotes(party, Number(goalId)),
    enabled: !!party && !!goalId,
  });

  if (isLoading) {
    return <div className="text-on-surface-variant py-16 text-center text-sm">Laddar...</div>;
  }
  if (error) {
    return <div className="text-contradiction py-16 text-center text-sm">Kunde inte ladda röstningar.</div>;
  }

  const { goal, matches } = data ?? { goal: null, matches: [] };

  return (
    <div>
      {/* ── Goal header ────────────────────────────────────────────── */}
      {goal && (
        <div
          className="rounded-xl p-6 mt-3 mb-5"
          style={{ background: pc?.light ?? "#f5f5f5" }}
        >
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-3 mb-2">
                <PartyBadge party={party} size="lg" />
                <TopicTag topic={goal.topic} />
                <SpecificityBadge specificity={goal.specificity} />
              </div>
              <p className="text-sm font-semibold text-on-surface leading-snug mb-2">
                &ldquo;{goal.goalText}&rdquo;
              </p>
              <p className="text-xs text-on-surface-variant">{goal.sourceDocument}</p>
            </div>
            {goal.alignmentPct != null && (
              <AlignmentRing pct={goal.alignmentPct} size={64} color={pc?.bg} />
            )}
          </div>
        </div>
      )}

      {/* ── Info banner ────────────────────────────────────────────── */}
      <div className="bg-[#f0f9ff] rounded-lg p-4 mb-5">
        <p className="text-xs text-[#0369a1] leading-relaxed m-0">
          Nedan visas de riksdagsomröstningar som matchats till detta mål.
          Varje omröstning visar <strong>vilket parti som initierade förslaget</strong>.
        </p>
      </div>

      {/* ── Vote matches ───────────────────────────────────────────── */}
      <h3 className="font-display text-base font-semibold text-on-surface mb-3">
        Matchade omröstningar ({matches.length})
      </h3>
      <div className="space-y-3">
        {matches.map((m, i) => (
          <VoteMatchCard key={i} match={m} />
        ))}
      </div>

      {matches.length === 0 && (
        <p className="text-on-surface-variant text-center py-16 text-sm">
          Inga relevanta omröstningar hittades för detta mål.
        </p>
      )}
    </div>
  );
}
