import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { votesApi } from "./api";
import { ProposalOriginTag } from "@/shared/components";
import { committeeFromBeteckning } from "@/shared/design";

export function VotesPage() {
  const [page, setPage] = useState(1);

  const { data, isLoading, error } = useQuery({
    queryKey: ["votes", page],
    queryFn: () => votesApi.list({ page, pageSize: 50 }),
  });

  if (isLoading) {
    return <div className="text-on-surface-variant py-16 text-center text-sm">Laddar omröstningar...</div>;
  }
  if (error) {
    return <div className="text-contradiction py-16 text-center text-sm">Kunde inte ladda omröstningar.</div>;
  }

  const votes = data?.data ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.ceil(total / 50);

  return (
    <div>
      <div className="mb-8">
        <h2 className="font-display text-2xl font-extrabold tracking-tight mb-1">
          Omröstningar
        </h2>
        <p className="text-sm text-on-surface-variant leading-relaxed max-w-lg">
          Alla berikade riksdagsomröstningar med förslagens ursprung.
        </p>
      </div>

      <div className="space-y-2">
        {votes.map((v) => {
          const committee = committeeFromBeteckning(v.beteckning);
          return (
            <Link
              key={`${v.beteckning}-${v.forslagspunkt}`}
              to={`/votes/${v.beteckning}/${v.forslagspunkt}`}
              className="flex items-start justify-between gap-4 py-3.5 px-4 rounded-lg transition-all hover:shadow-ambient group"
              style={{ background: "var(--color-surface-lowest)" }}
            >
              <div className="flex-1 min-w-0">
                <p className="text-sm font-semibold text-on-surface truncate group-hover:text-primary transition-colors">
                  {v.documentTitle || `${v.beteckning} punkt ${v.forslagspunkt}`}
                </p>
                <div className="flex flex-wrap items-center gap-2 mt-1.5">
                  <span className="text-[11px] font-mono text-on-surface-variant">
                    {v.beteckning} / punkt {v.forslagspunkt}
                  </span>
                  {committee && (
                    <span className="text-[11px] text-on-surface-variant">· {committee}</span>
                  )}
                  {(v.proposedByParty || v.proposalType) && (
                    <ProposalOriginTag proposedBy={v.proposedByParty} proposalType={v.proposalType} />
                  )}
                </div>
              </div>
            </Link>
          );
        })}
      </div>

      {votes.length === 0 && (
        <p className="text-on-surface-variant text-center py-16 text-sm">
          Inga berikade omröstningar finns ännu. Vänta tills datainsamlingen är klar.
        </p>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-3 mt-6">
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page <= 1}
            className="text-xs font-semibold text-primary disabled:text-on-surface-variant disabled:cursor-not-allowed cursor-pointer"
          >
            ← Föregående
          </button>
          <span className="text-xs text-on-surface-variant font-mono">
            {page} / {totalPages}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page >= totalPages}
            className="text-xs font-semibold text-primary disabled:text-on-surface-variant disabled:cursor-not-allowed cursor-pointer"
          >
            Nästa →
          </button>
        </div>
      )}
    </div>
  );
}
