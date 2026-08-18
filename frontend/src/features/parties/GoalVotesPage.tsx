import { useQuery } from "@tanstack/react-query";
import { useParams, Link } from "react-router-dom";
import { partiesApi } from "./api";
import {
  PartyBadge, TopicTag, SpecificityBadge,
  ProposalOriginTag,
} from "@/shared/components";
import { PARTY_COLORS, committeeFromBeteckning } from "@/shared/design";
import { SourceMarker } from "@/components/sources/SourceMarker";
import type { components } from "@/shared/api-contract";

type GoalVoteMatch = components["schemas"]["GoalVoteMatch"];

function VoteMatchCard({ match }: { match: GoalVoteMatch }) {
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
          {/* No source marker. The relevance figure is our keyword score, not
              anything Riksdagen published; marking it "riksdagen" attributed
              our own computation to the record. Its provenance is stated once
              above the list instead. */}
          <span className="text-[10px] font-mono text-on-surface-variant">
            {relevance}% relevans
          </span>
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

  const { goal } = data ?? { goal: null };
  // `?? []` and not just the destructuring default: the API answered `matches:
  // null` for a goal with no hits, and `null` survives a default that only
  // fires on `undefined`. That crashed the page whose whole job is to say we
  // found nothing.
  const matches = data?.matches ?? [];

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
              <p className="text-xs text-on-surface-variant">{goal.sourceDocument} <SourceMarker sourceId="seed-party-goals" /></p>
            </div>
          </div>
        </div>
      )}

      {/* ── Info banner ────────────────────────────────────────────── */}
      {/* Same voice as the shared GoalCard: the keywords are ours, and a hit is
          not the record having tested the goal. "Omröstningar som matchats till
          detta mål" said the opposite — that something matched them — and on a
          goal with no hits it promised rows that are not below it. */}
      <div className="bg-[#f0f9ff] rounded-lg p-4 mb-5">
        <p className="text-xs text-[#0369a1] leading-relaxed m-0">
          Nedan visas omröstningar vars titel nämner våra nyckelord. Nyckelorden
          är våra, inte partiets, och en träff betyder inte att omröstningen
          prövade målet.
        </p>
      </div>

      {/* ── Vote matches ───────────────────────────────────────────── */}
      {/* No count. A number of hits reads as a score for the goal — how much
          the party has been tested on it — when it only counts titles our
          keywords happened to land on. */}
      <h3 className="font-display text-base font-semibold text-on-surface mb-1">
        Omröstningar vars titel nämner våra nyckelord
      </h3>
      {/* Stated once, here, because the hits and the percentage are ours and
          nothing on the record says a votering belongs to this goal. */}
      <p className="text-xs text-on-surface-variant mb-3">
        Träffarna och deras relevanssiffra är våra, beräknade på nyckelord vi
        valt. <SourceMarker sourceId="seed-party-goals" />
      </p>
      <div className="space-y-3">
        {matches.map((m, i) => (
          <VoteMatchCard key={i} match={m} />
        ))}
      </div>

      {matches.length === 0 && (
        /* Our search is the gap, never the record. This page is now reachable
           for any goal carrying keywords, not only ones with hits, so the
           no-hit case is on the page rather than hypothetical. */
        <p className="text-on-surface-variant text-center py-16 text-sm">
          Våra nyckelord gav ingen träff bland omröstningarnas titlar. Det
          betyder inte att målet är oprövat — bara att vår sökning inte hittade
          något. <SourceMarker sourceId="seed-party-goals" />
        </p>
      )}
    </div>
  );
}
