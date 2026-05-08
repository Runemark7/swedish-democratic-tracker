import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { politiciansApi } from "./api";
import { PartyBadge } from "@/shared/components";
import { PARTY_COLORS } from "@/shared/design";
import { SectionSource } from "@/components/sources/SectionSource";
import type { PoliticianSummary } from "@/shared/types";

function PoliticianCard({ p }: { p: PoliticianSummary }) {
  const pc = PARTY_COLORS[p.party];

  return (
    <Link
      to={`/politicians/${p.intressentId}`}
      className="flex items-center gap-4 bg-surface-lowest rounded-xl p-4 transition-all hover:shadow-ambient group"
    >
      {/* Avatar */}
      {p.imageUrl ? (
        <img src={p.imageUrl} alt={p.firstName} className="w-12 h-12 rounded-full object-cover shrink-0" />
      ) : (
        <div
          className="w-12 h-12 rounded-full flex items-center justify-center shrink-0 text-sm font-extrabold"
          style={{ background: pc?.bg ?? "#666", color: pc?.text ?? "#fff" }}
        >
          {p.firstName[0]}{p.lastName[0]}
        </div>
      )}

      {/* Name + constituency */}
      <div className="flex-1 min-w-0">
        <p className="font-semibold text-sm text-on-surface truncate group-hover:text-primary transition-colors">
          {p.firstName} {p.lastName}
        </p>
        <p className="text-xs text-on-surface-variant truncate">{p.constituency}</p>
      </div>

      {/* Party badge */}
      <PartyBadge party={p.party} />
    </Link>
  );
}

export function PoliticiansPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");

  const isSearching = search.trim().length > 0;

  const { data, isLoading } = useQuery({
    queryKey: ["politicians", page, isSearching],
    queryFn: () => politiciansApi.list({ page: isSearching ? 1 : page, pageSize: isSearching ? 400 : 50 }),
  });

  const allPoliticians = data?.data ?? [];

  const filtered = isSearching
    ? allPoliticians.filter((p) => {
        const q = search.trim().toLowerCase();
        return (
          `${p.firstName} ${p.lastName}`.toLowerCase().includes(q) ||
          (p.constituency ?? "").toLowerCase().includes(q)
        );
      })
    : allPoliticians;

  const totalPages = isSearching ? 1 : (data ? Math.ceil(data.total / data.pageSize) : 1);

  return (
    <div>
      <div className="mb-6">
        <h2 className="font-display text-2xl font-extrabold tracking-tight mb-1">
          Enskilda politiker
        </h2>
        <p className="text-sm text-on-surface-variant leading-relaxed">
          Sök bland riksdagens ledamöter och se deras röstningshistorik.
        </p>
      </div>

      {/* ── Search ────────────────────────────────────────────────── */}
      <div className="mb-5">
        <input
          type="search"
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1); }}
          placeholder="Sök namn eller valkrets..."
          className="w-full sm:w-80 px-4 py-2 rounded-lg text-sm bg-surface-lowest text-on-surface placeholder:text-on-surface-variant outline-none focus:ring-2 focus:ring-primary/30"
          style={{ border: "1px solid var(--color-surface-high)" }}
        />
      </div>

      {/* ── Politician list ────────────────────────────────────────── */}
      {isLoading ? (
        <div className="text-on-surface-variant text-center py-16 text-sm">Laddar...</div>
      ) : (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-6">
            {filtered.map((p) => (
              <PoliticianCard key={p.intressentId} p={p} />
            ))}
          </div>

          {filtered.length === 0 && (
            <p className="text-on-surface-variant text-center py-16 text-sm">
              Inga politiker hittades.
            </p>
          )}

          <SectionSource sourceIds={["riksdagen"]} />

          {/* Pagination — hidden when searching */}
          {!isSearching && totalPages > 1 && (
            <div className="flex items-center justify-center gap-3">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1 rounded-lg text-sm bg-surface-lowest text-on-surface-variant disabled:opacity-40 cursor-pointer"
              >
                ←
              </button>
              <span className="text-sm text-on-surface-variant font-mono">
                {page} / {totalPages}
              </span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="px-3 py-1 rounded-lg text-sm bg-surface-lowest text-on-surface-variant disabled:opacity-40 cursor-pointer"
              >
                →
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
