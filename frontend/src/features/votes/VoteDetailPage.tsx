import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { votesApi } from "./api";
import { PartyBadge, VoteBar, ProposalOriginTag } from "@/shared/components";
import { PARTY_COLORS, committeeFromBeteckning } from "@/shared/design";
import type { PartyVotePosition } from "@/shared/types";

interface CommitteeProposal {
  punkt: string;
  rubrik: string;
  forslag: string;
}

async function fetchProposalDescription(dokId: string, punkt: string): Promise<CommitteeProposal | null> {
  try {
    const res = await fetch(`https://data.riksdagen.se/utskottsforslag/${dokId}.json`);
    if (!res.ok) return null;
    const data = await res.json();
    const proposals: any[] = [].concat(
      data?.utskottsforslag?.dokutskottsforslag?.utskottsforslag ?? []
    );
    const match = proposals.find((p: any) => String(p.punkt) === String(punkt));
    if (!match) return null;
    return {
      punkt: match.punkt,
      rubrik: match.rubrik ?? "",
      forslag: (match.forslag ?? "")
        .replace(/<BR\s*\/?>/gi, "\n")
        .replace(/<[^>]+>/g, "")
        .trim(),
    };
  } catch {
    return null;
  }
}

function PartyBreakdownCard({ pos }: { pos: PartyVotePosition }) {
  const total = pos.jaCount + pos.nejCount + pos.avstarCount + pos.franvarandeCount;
  const pc = PARTY_COLORS[pos.party];

  return (
    <div className="bg-surface-lowest rounded-xl p-5">
      <div className="flex items-center gap-3 mb-3">
        <PartyBadge party={pos.party} size="lg" />
        {pos.dominantVote && (
          <span
            className="text-[10px] font-extrabold px-2 py-0.5 rounded"
            style={{
              background:
                pos.dominantVote === "Ja" ? "#f0fdf4" :
                pos.dominantVote === "Nej" ? "#fef2f2" :
                pos.dominantVote === "Avstår" ? "#fffbeb" : "#f9fafb",
              color:
                pos.dominantVote === "Ja" ? "#16a34a" :
                pos.dominantVote === "Nej" ? "#dc2626" :
                pos.dominantVote === "Avstår" ? "#d97706" : "#9ca3af",
            }}
          >
            Majoritet: {pos.dominantVote}
          </span>
        )}
      </div>
      <VoteBar
        ja={pos.jaCount}
        nej={pos.nejCount}
        avstar={pos.avstarCount}
        franvarande={pos.franvarandeCount}
        total={total}
        partyColor={pc?.bg}
      />
    </div>
  );
}

function ShareDropdown({ beteckning, punkt }: { beteckning: string; punkt: string }) {
  const [open, setOpen] = useState(false);
  const [copied, setCopied] = useState(false);
  const shareUrl = `${window.location.origin}/votes/${beteckning}/${punkt}/dela`;

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(shareUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // silently fail
    }
    setOpen(false);
  };

  return (
    <div style={{ position: "relative", flexShrink: 0 }}>
      <button
        onClick={() => setOpen((o) => !o)}
        style={{
          display: "inline-flex",
          alignItems: "center",
          gap: 6,
          padding: "6px 12px",
          borderRadius: 8,
          border: "1px solid var(--color-border)",
          background: open ? "var(--color-sdt-surface)" : "transparent",
          color: "var(--color-fg-muted)",
          fontFamily: "var(--font-mono)",
          fontSize: 11,
          letterSpacing: "0.08em",
          cursor: "pointer",
          transition: "background 0.15s",
        }}
      >
        ↗ Dela
      </button>

      {open && (
        <>
          <div
            style={{ position: "fixed", inset: 0, zIndex: 40 }}
            onClick={() => setOpen(false)}
          />
          <div
            style={{
              position: "absolute",
              top: "calc(100% + 8px)",
              right: 0,
              zIndex: 50,
              background: "var(--color-bg)",
              border: "1px solid var(--color-border)",
              borderRadius: 10,
              padding: 8,
              minWidth: 230,
              boxShadow: "0 4px 24px rgba(0,0,0,0.15)",
              display: "flex",
              flexDirection: "column",
              gap: 2,
            }}
          >
            <button
              onClick={copy}
              style={{
                display: "block",
                width: "100%",
                textAlign: "left",
                padding: "9px 12px",
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                letterSpacing: "0.06em",
                background: "transparent",
                border: "none",
                borderRadius: 6,
                cursor: "pointer",
                color: copied ? "#16a34a" : "var(--color-fg)",
                transition: "background 0.1s",
              }}
              onMouseEnter={(e) => (e.currentTarget.style.background = "var(--color-sdt-surface)")}
              onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
            >
              {copied ? "✓  Länk kopierad!" : "⧉  Kopiera länk"}
            </button>
            <a
              href={shareUrl}
              target="_blank"
              rel="noopener noreferrer"
              onClick={() => setOpen(false)}
              style={{
                display: "block",
                padding: "9px 12px",
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                letterSpacing: "0.06em",
                borderRadius: 6,
                color: "var(--color-fg-muted)",
                textDecoration: "none",
                transition: "background 0.1s",
              }}
              onMouseEnter={(e) => (e.currentTarget.style.background = "var(--color-sdt-surface)")}
              onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
            >
              □  Öppna dela-sida →
            </a>
          </div>
        </>
      )}
    </div>
  );
}

export function VoteDetailPage() {
  const { beteckning = "", punkt = "" } = useParams();

  const { data, isLoading, error } = useQuery({
    queryKey: ["vote-detail", beteckning, punkt],
    queryFn: () => votesApi.getDetail(beteckning, punkt),
    enabled: !!beteckning && !!punkt,
  });

  const { data: proposal } = useQuery({
    queryKey: ["proposal-description", data?.dokId, punkt],
    queryFn: () => fetchProposalDescription(data!.dokId!, punkt),
    enabled: !!data?.dokId,
    staleTime: 30 * 60 * 1000,
  });

  if (isLoading) {
    return <div className="text-on-surface-variant py-16 text-center text-sm">Laddar omröstning...</div>;
  }
  if (error || !data) {
    return <div className="text-contradiction py-16 text-center text-sm">Omröstning hittades inte.</div>;
  }

  const committee = committeeFromBeteckning(beteckning);
  const breakdown = (data.partyBreakdown ?? []).sort(
    (a, b) => (b.jaCount + b.nejCount) - (a.jaCount + a.nejCount)
  );

  return (
    <div>
      {/* ── Vote header ────────────────────────────────────────────── */}
      <div className="bg-surface-lowest rounded-xl p-6 mt-3 mb-5">
        <div className="flex items-start justify-between gap-4 mb-3">
          <h2 className="font-display text-[22px] font-extrabold tracking-tight flex-1">
            {data.documentTitle || `${beteckning} punkt ${punkt}`}
          </h2>
          <ShareDropdown beteckning={beteckning} punkt={punkt} />
        </div>
        <div className="flex flex-wrap items-center gap-2 mb-3">
          <span className="text-[11px] font-mono font-semibold text-on-surface-variant px-2 py-0.5 rounded bg-surface-low">
            {beteckning} / punkt {punkt}
          </span>
          {committee && (
            <span className="text-[11px] text-on-surface-variant px-2 py-0.5 rounded bg-surface-low">
              {committee}
            </span>
          )}
          {data.session && (
            <span className="text-[11px] text-on-surface-variant px-2 py-0.5 rounded bg-surface-low">
              Riksmöte {data.session}
            </span>
          )}
        </div>
        {(data.proposedByParty || data.proposalType) && (
          <ProposalOriginTag proposedBy={data.proposedByParty} proposalType={data.proposalType} />
        )}
      </div>

      {/* ── Proposal description ───────────────────────────────────── */}
      {proposal && (proposal.rubrik || proposal.forslag) && (
        <div className="bg-surface-lowest rounded-xl p-6 mb-5">
          <p className="text-[10px] uppercase tracking-widest text-on-surface-variant font-semibold mb-3">
            Vad handlar voteringen om?
          </p>
          {proposal.rubrik && (
            <h3 className="text-base font-semibold text-on-surface mb-2 leading-snug">
              {proposal.rubrik}
            </h3>
          )}
          {proposal.forslag && (
            <p className="text-sm text-on-surface-variant leading-relaxed whitespace-pre-line">
              {proposal.forslag}
            </p>
          )}
        </div>
      )}

      {/* ── Party breakdown ────────────────────────────────────────── */}
      <h3 className="font-display text-base font-semibold text-on-surface mb-3">
        Partiernas röster
      </h3>
      <div className="space-y-3">
        {breakdown.map((pos) => (
          <PartyBreakdownCard key={pos.party} pos={pos} />
        ))}
      </div>

      {breakdown.length === 0 && (
        <p className="text-on-surface-variant text-center py-16 text-sm">
          Ingen röstdata tillgänglig.
        </p>
      )}

      {/* ── Debate link ────────────────────────────────────────────── */}
      {data.dokId && (
        <div className="mt-8 p-5 bg-surface-lowest rounded-xl">
          <h2 className="text-[13px] font-semibold uppercase tracking-wider text-on-surface-variant mb-2">
            Debatt i kammaren
          </h2>
          <p className="text-sm text-on-surface-variant mb-4 leading-relaxed">
            Riksdagen.se har video och transkript från debatten om detta betänkande.
          </p>
          <a
            href={`https://www.riksdagen.se/sv/dokument-och-lagar/dokument/betankande/_${data.dokId}/`}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 text-sm font-medium px-4 py-2 rounded-lg bg-surface-low text-on-surface hover:bg-surface transition-colors"
          >
            Se hela debatten på riksdagen.se →
          </a>
        </div>
      )}
    </div>
  );
}
