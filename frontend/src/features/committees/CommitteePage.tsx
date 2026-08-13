import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { committeesApi } from "./api";
import { ApiError } from "@/shared/api-client";
import { GoalCard } from "@/shared/GoalCard";
import { PartyBadge } from "@/shared/components";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { formatBudgetAmount } from "@/shared/design";
import { useRecordCoverage } from "@/hooks/useDemocracy";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import type { GoalWithAlignment } from "@/shared/types";

const PAGE_SIZE = 50;

/**
 * Committees whose remit is wider than the utgiftsområden they bereder. Without
 * this note the amounts read as the size of the committee's job, which for
 * these two they are not.
 */
const REMIT_NOTE: Record<string, string> = {
  SkU: "Skatteutskottet bereder statens inkomster, som inte är ett utgiftsområde. Beloppen nedan visar därför bara en del av utskottets område.",
  FiU: "Finansutskottet bereder budgeten som helhet, inte bara utgiftsområdena nedan.",
};

/** A page that says nothing about the record, only that we could not read it.
 *  Carries no SourceMarker: a transport failure has no source. */
function FailureNotice({ children }: { children: React.ReactNode }) {
  return (
    <div className="sdt-page px-8 py-16">
      <p className="text-sm text-on-surface-variant">{children}</p>
      <Link to="/parties" className="text-xs text-primary hover:underline">
        ← Partier
      </Link>
    </div>
  );
}

export function CommitteePage() {
  const { code = "" } = useParams<{ code: string }>();
  const isMobile = useMediaQuery("(max-width: 640px)");
  // The pager is keyed to a committee, and the key is stored with it. Carried
  // across a change of :code, page 4 of a 120-votering committee shows an empty
  // list under a pager reading "151–120 av 120" — unreachable until
  // committee→committee navigation exists, and wrong the day it does. Adjusted
  // during render rather than in an effect, so the discarded page is never
  // rendered or fetched.
  const [pager, setPager] = useState({ code, page: 0 });
  const page = pager.code === code ? pager.page : 0;
  const setPage = (next: (p: number) => number) =>
    setPager({ code, page: next(page) });
  if (pager.code !== code) setPager({ code, page: 0 });

  const {
    data: coverage,
    isError: coverageError,
    isLoading: loadingCoverage,
  } = useRecordCoverage();
  const period = coverage?.mandate.code;

  const {
    data: committee,
    isLoading: loadingCommittee,
    error: committeeError,
  } = useQuery({
    queryKey: ["committee", code, period],
    queryFn: () => committeesApi.getCommittee(code, period!),
    enabled: !!code && !!period,
    staleTime: 5 * 60_000,
  });

  const { data: goals, isLoading: loadingGoals, isError: goalsError } = useQuery({
    queryKey: ["committee-goals", code],
    queryFn: () => committeesApi.listGoals(code),
    enabled: !!code,
    staleTime: 60_000,
  });

  const { data: votesPage, isLoading: loadingVotes, isError: votesError } = useQuery({
    queryKey: ["committee-votes", code, period, page],
    queryFn: () => committeesApi.listVotes(code, period!, PAGE_SIZE, page * PAGE_SIZE),
    enabled: !!code && !!period,
    staleTime: 60_000,
  });

  // Coverage names the mandate period every query below is scoped to, so it is
  // resolved first. Without this the committee query stays disabled, which in
  // TanStack v5 means isLoading is false and data undefined — indistinguishable
  // from "no such committee" unless coverage is checked on its own.
  if (coverageError) {
    return (
      <FailureNotice>
        Vi kunde inte läsa vilken mandatperiod som gäller, så utskottet kan inte
        visas just nu.
      </FailureNotice>
    );
  }
  if (loadingCoverage || !period || loadingCommittee) {
    return <p className="sdt-page px-8 py-16 text-sm text-on-surface-variant">Laddar utskott...</p>;
  }
  // 404 is the one failure that licenses a claim about the register: the
  // backend looked and holds no such committee in this period. Every other
  // failure — 500, 503, a dropped connection — says only that we could not
  // ask. Reporting those as "this committee does not exist" turns a transient
  // outage into a false statement about parliament.
  if (committeeError instanceof ApiError && committeeError.status === 404) {
    return (
      <div className="sdt-page px-8 py-16">
        <p className="text-sm text-on-surface-variant">
          Vi har inget utskott med koden {code} i den här mandatperiodens register.{" "}
          <SourceMarker sourceId="riksdagen" />
        </p>
        <Link to="/parties" className="text-xs text-primary hover:underline">
          ← Partier
        </Link>
      </div>
    );
  }
  if (committeeError || !committee) {
    return <FailureNotice>Vi kunde inte hämta utskottet just nu. Försök igen.</FailureNotice>;
  }

  // Alphabetical by party. Any other order would rank the parties, which is an
  // editorial act the site does not perform.
  const goalList: GoalWithAlignment[] = goals ?? [];
  const byParty = new Map<string, GoalWithAlignment[]>();
  for (const g of goalList) {
    const bucket = byParty.get(g.party) ?? [];
    bucket.push(g);
    byParty.set(g.party, bucket);
  }
  const parties = [...byParty.keys()].sort();

  const votes = votesPage?.items ?? [];
  const total = votesPage?.total ?? 0;

  const areas = committee.expenditureAreas ?? [];
  const remitNote = REMIT_NOTE[committee.code];
  const pad = isMobile ? "0 14px" : "0 32px";

  return (
    <div className="sdt-page" style={{ paddingBottom: 64 }}>
      <div style={{ padding: isMobile ? "16px 14px 0" : "32px 32px 0" }}>
        <Link to="/parties" className="text-xs text-primary hover:underline">
          ← Partier
        </Link>
      </div>

      <header style={{ padding: isMobile ? "12px 14px 16px" : "16px 32px 24px" }}>
        {/* The name falls back to the code in the API, so a committee we have no
            name for still renders rather than showing "undefined". */}
        <h1 className="font-display text-3xl font-extrabold tracking-tight m-0">
          {committee.name}
        </h1>
        <p className="text-xs text-on-surface-variant mt-2">
          {committee.code} · {committee.voteringar.toLocaleString("sv-SE")} voteringar under{" "}
          {coverage?.mandate.label ?? "perioden"} <SourceMarker sourceId="riksdagen" />
          {/* This count is our holdings, not Riksdagen's total. The party page
              states its shortfall next to the same record; without the same
              caveat here the number reads as complete. */}
          {coverage && coverage.ingested < coverage.expected && (
            <>
              {" "}
              Vi saknar ett fåtal av periodens omröstningar.{" "}
              <Link to="/data" className="underline">
                Vad saknas?
              </Link>
            </>
          )}
          <br />
          {/* Stated so the number is not read as a measure of importance, and
              so the grain is named: /parties/:party counts the same record per
              förslagspunkt and reports 2 558 where the committee pages sum to
              2 576. The 18-row gap is the 18 double-decided förslagspunkter,
              and a reader adding committee totals up must be able to see why. */}
          <span className="text-[11px]">
            En räkning av registret, inte ett mått på utskottets betydelse.
            Räknat per votering. En förslagspunkt som avgjordes av två
            voteringar räknas som två.
          </span>
        </p>
      </header>

      <main style={{ padding: pad, maxWidth: 920 }}>
        {/* ── LOVAT ─────────────────────────────────────────────────── */}
        <section className="mb-10">
          <h2 className="font-display text-lg font-bold mb-1">
            Partimål inom {committee.name}s område
          </h2>
          {/* Committee-level, not point-level: relevant_committees says the goal
              touches this area, never that it addresses a particular decision. */}
          <p className="text-xs text-on-surface-variant mb-4">
            Mål som partierna själva formulerat och som rör utskottets område.
          </p>

          {loadingGoals && (
            <p className="text-sm text-on-surface-variant">Laddar partimål...</p>
          )}

          {/* A failed fetch is not an empty result. Without this the caption
              below states we hold no goals for an area we may hold nine for,
              on the strength of a dropped request. No source marker: nothing
              here comes from a source. */}
          {goalsError && (
            <p className="text-sm text-on-surface-variant">
              Vi kunde inte hämta partimålen just nu. Det säger inget om vad
              partierna vill — försök igen.
            </p>
          )}

          {!loadingGoals && !goalsError && goalList.length === 0 && (
            /* Our seed reading is the gap, never the parties' silence. */
            <p className="text-sm text-on-surface-variant">
              Vi har inga inlästa partimål inom detta område.{" "}
              <SourceMarker sourceId="seed-party-goals" />
            </p>
          )}

          {parties.map((p) => (
            <div key={p} className="mb-6">
              <div className="flex items-center gap-2 mb-2">
                <PartyBadge party={p} />
              </div>
              <div className="space-y-3">
                {(byParty.get(p) ?? []).map((g) => (
                  <GoalCard key={g.id} goal={g} party={p} />
                ))}
              </div>
            </div>
          ))}
        </section>

        {/* ── RÖSTAT ────────────────────────────────────────────────── */}
        <section className="mb-10">
          <h2 className="font-display text-lg font-bold mb-1">
            Så röstade partierna i utskottets ärenden
          </h2>
          {/* Positions are stated, never scored. The promise sits above and the
              reader draws the connection; the site does not draw it for them. */}
          <p className="text-xs text-on-surface-variant mb-4">
            Varje rad är en förslagspunkt och den ståndpunkt partiets ledamöter
            tog. Utan beslutsdatum: registret innehåller inget sådant datum.{" "}
            <SourceMarker sourceId="riksdagen" />
          </p>

          {loadingVotes && <p className="text-sm text-on-surface-variant">Laddar omröstningar...</p>}

          {/* Stated separately from the empty case, which would otherwise
              contradict the header above it: "inga inlästa omröstningar" under
              a heading reading "AU · 116 voteringar". */}
          {votesError && (
            <p className="text-sm text-on-surface-variant">
              Vi kunde inte hämta omröstningarna just nu. Det säger inget om vad
              utskottet beslutat — försök igen.
            </p>
          )}

          {!loadingVotes && !votesError && total === 0 && (
            <p className="text-sm text-on-surface-variant">
              Vi har inga inlästa omröstningar för detta utskott i perioden.{" "}
              <SourceMarker sourceId="riksdagen" />
            </p>
          )}

          {votes.map((v) => (
            <article
              key={v.voteringId}
              className="rounded-xl px-5 py-4 mb-3"
              style={{ background: "var(--color-surface-lowest)" }}
            >
              <div className="flex items-baseline gap-2 flex-wrap">
                <Link
                  to={`/votes/${v.beteckning}/${v.forslagspunkt}`}
                  className="text-sm font-semibold text-on-surface hover:underline"
                >
                  {v.beteckning} punkt {v.forslagspunkt}
                </Link>
                <span className="text-[11px] text-on-surface-variant">{v.riksmote}</span>
              </div>
              {v.documentTitle && (
                <p className="text-xs text-on-surface-variant mt-1">{v.documentTitle}</p>
              )}
              {/* From the API, which sees the committee's whole record. Counted
                  within this page instead, a pair split across two pages —
                  NU7 punkt 2 in 2025/26 is one — leaves both halves unmarked,
                  and four parties read as reversing their vote for no reason. */}
              {v.decidedByMultipleVoteringar && (
                <p className="text-[11px] text-on-surface-variant mt-1">
                  Denna förslagspunkt avgjordes av två separata voteringar. Raderna
                  visar olika omröstningar, inte samma omröstning två gånger.
                </p>
              )}
              <div className="flex flex-wrap gap-x-4 gap-y-1 mt-3">
                {v.partyPositions.map((pp) => (
                  <span key={pp.party} className="text-xs text-on-surface-variant">
                    <span className="font-semibold text-on-surface">{pp.party}</span>{" "}
                    {pp.position}
                  </span>
                ))}
              </div>
            </article>
          ))}

          {total > PAGE_SIZE && (
            <div className="flex items-center gap-4 mt-4">
              <button
                onClick={() => setPage((p) => Math.max(0, p - 1))}
                disabled={page === 0}
                className="text-xs text-primary disabled:text-on-surface-variant disabled:cursor-default hover:underline cursor-pointer"
              >
                ← Föregående
              </button>
              <span className="text-[11px] text-on-surface-variant">
                {page * PAGE_SIZE + 1}–{Math.min((page + 1) * PAGE_SIZE, total)} av {total}
              </span>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={(page + 1) * PAGE_SIZE >= total}
                className="text-xs text-primary disabled:text-on-surface-variant disabled:cursor-default hover:underline cursor-pointer"
              >
                Nästa →
              </button>
            </div>
          )}
        </section>

        {/* ── KOSTAR ────────────────────────────────────────────────── */}
        <section className="mb-10">
          <h2 className="font-display text-lg font-bold mb-1">
            Utgiftsområden som utskottet bereder
          </h2>
          {/* Amounts only. A share of the total would understate a committee
              whose remit exceeds its areas, and a ranking would imply that one
              area matters more than another. */}
          {/* Two sources, one sentence each, each beside what it actually
              sources. The bilaga says which utgiftsområden this committee
              bereder and contains no amounts at all; the amounts are our
              seeded budget figures, which every other budget surface marks
              seed-budget-data. Marking the amounts as the bilaga's credited
              a law with numbers it does not carry. */}
          <p className="text-xs text-on-surface-variant mb-4">
            Vilka utgiftsområden utskottet bereder står i riksdagsordningens
            bilaga. <SourceMarker sourceId="utskott-utgiftsomrade" />
            <br />
            Beloppen kommer från den beslutade statsbudgeten.{" "}
            <SourceMarker sourceId="seed-budget-data" />
            {committee.budgetYear > 0 && (
              <> Belopp för budgetåret {committee.budgetYear}.</>
            )}
          </p>

          {remitNote && (
            <p className="text-xs text-on-surface-variant mb-4 leading-relaxed">{remitNote}</p>
          )}

          {areas.length === 0 ? (
            /* Our table is the gap, never the committee's remit. "Detta
               utskott bereder inget utgiftsområde" states a fact about the
               world that only our reading of the bilaga supports. */
            <p className="text-sm text-on-surface-variant">
              Vi har inget utgiftsområde inläst för detta utskott.{" "}
              <SourceMarker sourceId="utskott-utgiftsomrade" />
            </p>
          ) : (
            <ul className="list-none m-0 p-0">
              {areas.map((a) => (
                <li
                  key={a.code}
                  className="flex justify-between gap-4 py-2"
                  style={{ borderBottom: "1px solid var(--color-border)" }}
                >
                  <span className="text-sm text-on-surface">
                    <span className="text-on-surface-variant">{a.code}</span> {a.name}
                  </span>
                  <span className="text-sm font-semibold text-on-surface whitespace-nowrap">
                    {formatBudgetAmount(a.amountKsek)}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </section>
      </main>
    </div>
  );
}
