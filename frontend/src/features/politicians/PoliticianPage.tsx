import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams, Link } from "react-router-dom";
import { politiciansApi } from "./api";
import { PartyBadge, ProposalOriginTag, TopicTag, SpecificityBadge } from "@/shared/components";
import { PARTY_COLORS, committeeFromBeteckning } from "@/shared/design";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { SectionSource } from "@/components/sources/SectionSource";
import type { Vote, PromiseWithMatches } from "@/shared/types";

const VOTE_STYLES: Record<string, { color: string; bg: string }> = {
  Ja:          { color: "#16a34a", bg: "#f0fdf4" },
  Nej:         { color: "#dc2626", bg: "#fef2f2" },
  "Avstår":    { color: "#d97706", bg: "#fffbeb" },
  Frånvarande: { color: "#9ca3af", bg: "#f9fafb" },
};

function VoteRow({ vote }: { vote: Vote }) {
  const style = VOTE_STYLES[vote.voteResult] ?? VOTE_STYLES.Frånvarande;
  const committee = committeeFromBeteckning(vote.beteckning);

  return (
    <Link
      to={`/votes/${vote.beteckning}/${vote.forslagspunkt}`}
      className="flex items-start justify-between gap-4 py-3.5 px-4 rounded-lg transition-all hover:shadow-ambient group"
      style={{ background: "var(--color-surface-lowest)" }}
    >
      <div className="flex-1 min-w-0">
        <p className="text-sm font-semibold text-on-surface truncate group-hover:text-primary transition-colors">
          {vote.documentTitle || `${vote.beteckning} punkt ${vote.forslagspunkt}`}
          <SourceMarker sourceId="riksdagen" />
        </p>
        <div className="flex flex-wrap items-center gap-2 mt-1.5">
          <span className="text-[11px] font-mono text-on-surface-variant">{vote.beteckning}</span>
          {committee && (
            <span className="text-[11px] text-on-surface-variant">· {committee}</span>
          )}
          {(vote.proposedByParty || vote.proposalType) && (
            <ProposalOriginTag proposedBy={vote.proposedByParty} proposalType={vote.proposalType} />
          )}
        </div>
      </div>
      <span
        className="text-xs font-extrabold px-2.5 py-1 rounded shrink-0"
        style={{ background: style.bg, color: style.color }}
      >
        {vote.voteResult}
      </span>
    </Link>
  );
}

const ALIGNMENT_STYLES: Record<string, { color: string; bg: string; label: string }> = {
  supports:    { color: "#16a34a", bg: "#f0fdf4", label: "Stödjer" },
  contradicts: { color: "#dc2626", bg: "#fef2f2", label: "Motsäger" },
  unclear:     { color: "#9ca3af", bg: "#f9fafb", label: "Oklart" },
};

function PromiseCard({ promise }: { promise: PromiseWithMatches }) {
  return (
    <div className="bg-surface-lowest rounded-xl p-5">
      <div className="flex flex-wrap items-center gap-2 mb-2">
        <TopicTag topic={promise.topic} />
        <SpecificityBadge specificity={promise.specificity} />
        {promise.speechDate && (
          <span className="text-[11px] text-on-surface-variant">{promise.speechDate}</span>
        )}
      </div>
      <p className="text-sm text-on-surface leading-relaxed mb-3">&ldquo;{promise.promiseText}&rdquo;</p>
      {promise.speechTopic && (
        <p className="text-xs text-on-surface-variant italic mb-3">Ämne: {promise.speechTopic}</p>
      )}
      {(promise.voteMatches ?? []).length > 0 && (
        <div className="space-y-2 pt-3" style={{ borderTop: "1px solid var(--color-surface-high)" }}>
          {promise.voteMatches!.map((m) => {
            const s = ALIGNMENT_STYLES[m.alignment] ?? ALIGNMENT_STYLES.unclear;
            return (
              <div key={m.voteId} className="flex items-start justify-between gap-3">
                <Link
                  to={`/votes/${m.beteckning}/${m.forslagspunkt}`}
                  className="text-xs text-on-surface-variant hover:text-primary transition-colors truncate flex-1"
                >
                  {m.documentTitle || `${m.beteckning} punkt ${m.forslagspunkt}`}
                </Link>
                <span
                  className="text-[10px] font-extrabold px-2 py-0.5 rounded shrink-0"
                  style={{ background: s.bg, color: s.color }}
                >
                  {s.label}
                </span>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

export function PoliticianPage() {
  const { id = "" } = useParams();
  const [activeTab, setActiveTab] = useState<"votes" | "promises">("votes");

  const { data: politician, isLoading: loadingPolitician } = useQuery({
    queryKey: ["politician", id],
    queryFn: () => politiciansApi.getById(id),
    enabled: !!id,
  });

  const { data: votesData, isLoading: loadingVotes } = useQuery({
    queryKey: ["politician-votes", id],
    queryFn: () => politiciansApi.listVotes(id, { pageSize: 100 }),
    enabled: !!id,
  });

  const { data: promises, isLoading: loadingPromises } = useQuery({
    queryKey: ["politician-promises", id],
    queryFn: () => politiciansApi.listPromises(id),
    enabled: !!id && activeTab === "promises",
  });

  if (loadingPolitician) {
    return <div className="text-on-surface-variant py-16 text-center text-sm">Laddar...</div>;
  }
  if (!politician) {
    return <div className="text-contradiction py-16 text-center text-sm">Politiker hittades inte.</div>;
  }

  const pc = PARTY_COLORS[politician.party];
  const votes = votesData?.data ?? [];

  const ja = votes.filter((v) => v.voteResult === "Ja").length;
  const nej = votes.filter((v) => v.voteResult === "Nej").length;
  const avstar = votes.filter((v) => v.voteResult === "Avstår").length;
  const franvarande = votes.filter((v) => v.voteResult === "Frånvarande").length;

  return (
    <div>
      {/* ── Profile header ─────────────────────────────────────────── */}
      <div
        className="rounded-xl p-6 mt-3 mb-5"
        style={{ background: pc?.light ?? "#f5f5f5" }}
      >
        <div className="flex items-center gap-5">
          {politician.imageUrl ? (
            <img
              src={politician.imageUrl}
              alt={politician.firstName}
              className="w-20 h-20 rounded-full object-cover shrink-0"
              loading="lazy"
            />
          ) : (
            <div
              className="w-20 h-20 rounded-full flex items-center justify-center shrink-0 text-xl font-extrabold"
              style={{ background: pc?.bg ?? "#666", color: pc?.text ?? "#fff" }}
            >
              {politician.firstName[0]}{politician.lastName[0]}
            </div>
          )}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 mb-1">
              <h2 className="font-display text-[22px] font-extrabold tracking-tight m-0">
                {politician.firstName} {politician.lastName}
              </h2>
              <PartyBadge party={politician.party} size="lg" />
            </div>
            {politician.constituency && (
              <p className="text-sm text-on-surface-variant">{politician.constituency}</p>
            )}
          </div>
        </div>

        {/* Stats row */}
        {votes.length > 0 && (
          <div className="flex gap-6 mt-5 pt-4" style={{ borderTop: "1px solid rgba(0,0,0,0.06)" }}>
            {[
              { label: "Ja", value: ja, color: "#16a34a" },
              { label: "Nej", value: nej, color: "#dc2626" },
              { label: "Avstår", value: avstar, color: "#d97706" },
              { label: "Frånv.", value: franvarande, color: "#9ca3af" },
            ].map((s) => (
              <div key={s.label} className="text-center">
                <div className="font-mono font-extrabold text-lg" style={{ color: s.color }}>
                  {s.value}
                  <SourceMarker sourceId="riksdagen" />
                </div>
                <div className="text-[10px] text-on-surface-variant uppercase tracking-widest">
                  {s.label}
                </div>
              </div>
            ))}
            <div className="text-center">
              <div className="font-mono font-extrabold text-lg text-on-surface">
                {votes.length}
                <SourceMarker sourceId="riksdagen" />
              </div>
              <div className="text-[10px] text-on-surface-variant uppercase tracking-widest">
                Totalt
              </div>
            </div>
          </div>
        )}
      </div>

      {/* ── Tabs ───────────────────────────────────────────────────── */}
      <div className="flex gap-1 mb-5" style={{ borderBottom: "1px solid var(--color-surface-high)" }}>
        {(["votes", "promises"] as const).map((tab) => {
          const label = tab === "votes" ? "Röstningshistorik" : "Löften";
          const active = activeTab === tab;
          return (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className="px-4 py-2.5 text-sm font-medium transition-colors cursor-pointer"
              style={{
                borderBottom: active ? "2px solid var(--color-on-surface)" : "2px solid transparent",
                color: active ? "var(--color-on-surface)" : "var(--color-on-surface-variant)",
                fontWeight: active ? 700 : 500,
                marginBottom: "-1px",
              }}
            >
              {label}
            </button>
          );
        })}
      </div>

      {/* ── Vote history ───────────────────────────────────────────── */}
      {activeTab === "votes" && (
        loadingVotes ? (
          <div className="text-on-surface-variant text-center py-12 text-sm">Laddar röstningar...</div>
        ) : (
          <>
            <div className="space-y-2">
              {votes.map((v) => (
                <VoteRow key={v.id} vote={v} />
              ))}
            </div>
            {votes.length > 0 && (
              <div className="mt-4">
                <SectionSource sourceIds={["riksdagen"]} />
              </div>
            )}
            {votes.length === 0 && (
              <p className="text-on-surface-variant text-center py-16 text-sm">
                Inga röstningar hittades.
              </p>
            )}
          </>
        )
      )}

      {/* ── Promises ───────────────────────────────────────────────── */}
      {activeTab === "promises" && (
        loadingPromises ? (
          <div className="text-on-surface-variant text-center py-12 text-sm">Laddar löften...</div>
        ) : (
          <>
            <div className="space-y-3">
              {(promises ?? []).map((p) => (
                <PromiseCard key={p.id} promise={p} />
              ))}
            </div>
            {(promises ?? []).length === 0 && (
              <p className="text-on-surface-variant text-center py-16 text-sm">
                Inga löften extraherade ännu.
              </p>
            )}
          </>
        )
      )}
    </div>
  );
}
