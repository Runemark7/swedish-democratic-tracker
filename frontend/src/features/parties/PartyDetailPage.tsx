import { useQuery } from "@tanstack/react-query";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { partiesApi } from "./api";
import { committeesApi } from "@/features/committees/api";
import { politiciansApi } from "@/features/politicians/api";
import { regeringApi } from "@/features/regering/api";
import { PARTY_COLORS, partyShortToName } from "@/shared/design";
import { useRecordCoverage, useSpeechesByParty } from "@/hooks/useDemocracy";
import { SpeechRow } from "@/features/speeches/SpeechRow";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { GoalCard } from "@/shared/GoalCard";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { swedishDate } from "@/shared/dates";
import type { components } from "@/shared/api-contract";

type Goal = components["schemas"]["Goal"];

type Tab = "mal" | "anforanden" | "politiker";

const TABS: { key: Tab; label: string }[] = [
  { key: "mal", label: "Mål" },
  { key: "anforanden", label: "Anföranden" },
  { key: "politiker", label: "Politiker" },
];

export function PartyDetailPage() {
  const { party = "" } = useParams<{ party: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const isMobile = useMediaQuery("(max-width: 640px)");

  const tab: Tab =
    (searchParams.get("tab") as Tab | null) && TABS.some((t) => t.key === searchParams.get("tab"))
      ? (searchParams.get("tab") as Tab)
      : "mal";

  const setTab = (next: Tab) => {
    const params = new URLSearchParams(searchParams);
    if (next === "mal") params.delete("tab");
    else params.set("tab", next);
    setSearchParams(params, { replace: true });
  };

  const { data: parties } = useQuery({
    queryKey: ["parties"],
    queryFn: partiesApi.listParties,
    staleTime: 5 * 60 * 1000,
  });
  const summary = parties?.find((p) => p.party === party);

  const { data: goals, isLoading: loadingGoals } = useQuery({
    queryKey: ["party-goals", party],
    queryFn: () => partiesApi.listGoals(party),
    enabled: !!party && tab === "mal",
    staleTime: 60_000,
  });

  // The record is bound to a named mandate period, so the committee list is
  // the one the period's record actually shows — never a hardcoded vintage.
  const { data: coverage, isError: coverageError } = useRecordCoverage();
  const period = coverage?.mandate.code;
  // Without the period the committee query never runs, so it reports neither
  // loading nor error. Treating that silence as "loaded" is what let a coverage
  // outage publish every goal as lacking a committee in our reading.
  const periodPending = !period && !coverageError;

  const { data: committees, isLoading: loadingCommittees, isError: committeesError } = useQuery({
    queryKey: ["committees", period],
    queryFn: () => committeesApi.listCommittees(period!),
    enabled: !!period && tab === "mal",
    staleTime: 5 * 60_000,
  });

  const { data: speeches, isLoading: loadingSpeeches } = useSpeechesByParty(
    tab === "anforanden" ? party : undefined,
    50,
  );

  const { data: politiciansResp, isLoading: loadingPoliticians } = useQuery({
    queryKey: ["politicians", { party, pageSize: 200 }],
    queryFn: () => politiciansApi.list({ party, pageSize: 200 }),
    enabled: !!party && tab === "politiker",
    staleTime: 60_000,
  });

  const { data: ministerGroups } = useQuery({
    queryKey: ["ministers"],
    queryFn: regeringApi.listMinisters,
    staleTime: 5 * 60_000,
  });
  const partyMinisters = (ministerGroups ?? [])
    .flatMap((g) => g.ministers ?? [])
    .filter((m) => m.party === party && m.active);

  const pc = PARTY_COLORS[party];

  // One subject taxonomy: committees. `topic` survives only as a tag on the
  // card. Every committee in the period is listed, not only those where this
  // party has goals — an absent row would read as the subject not existing.
  //
  // A goal whose remit spans several committees appears under each of them, so
  // the section counts deliberately do not sum to the party's total. That is
  // stated on the page rather than left for the reader to infer from arithmetic.
  const goalList: Goal[] = goals ?? [];
  const byCommittee = new Map<string, Goal[]>();
  for (const c of committees ?? []) byCommittee.set(c.code, []);
  const placed = new Set<number>();
  for (const g of goalList) {
    for (const code of g.relevantCommittees ?? []) {
      const bucket = byCommittee.get(code);
      if (bucket) {
        bucket.push(g);
        placed.add(g.id);
      }
    }
  }
  // A goal carrying no committee, or one naming a committee absent from the
  // period's record, would otherwise vanish from a page that still claims to
  // show the party's goals. None exist today; this is here so that if one ever
  // does, it is visible rather than silently dropped.
  const unplaced = goalList.filter((g) => !placed.has(g.id));

  return (
    <div className="sdt-page" style={{ paddingBottom: 64 }}>
      {/* Back link */}
      <div style={{ padding: isMobile ? "16px 14px 0" : "32px 32px 0" }}>
        <Link
          to="/parties"
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-accent)",
            textDecoration: "none",
            letterSpacing: "0.1em",
          }}
        >
          ← Alla partier
        </Link>
      </div>

      {/* Hero */}
      <header
        style={{
          padding: isMobile ? "12px 14px 16px" : "16px 32px 24px",
          display: "flex",
          alignItems: "center",
          gap: 16,
        }}
      >
        <span
          style={{
            display: "inline-block",
            width: 6,
            height: 56,
            background: pc?.bg ?? "#888",
            flexShrink: 0,
          }}
        />
        <div style={{ minWidth: 0 }}>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.1em",
            }}
          >
            {party}
          </div>
          <h1
            style={{
              fontFamily: "var(--font-serif)",
              fontSize: isMobile ? 28 : 40,
              fontWeight: 400,
              margin: 0,
              lineHeight: 1.05,
              color: "var(--color-fg)",
            }}
          >
            {partyShortToName(party)}
          </h1>
          {summary && (
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                color: "var(--color-fg-muted)",
                marginTop: 6,
              }}
            >
              {summary.totalGoals ?? 0} mål
            </div>
          )}
        </div>
      </header>

      {/* Tab strip */}
      <div
        style={{
          display: "flex",
          gap: 0,
          padding: isMobile ? "0 14px" : "0 32px",
          borderBottom: "1px solid var(--color-border)",
          overflowX: "auto",
        }}
      >
        {TABS.map((t) => {
          const active = tab === t.key;
          return (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              style={{
                background: "transparent",
                border: "none",
                padding: "10px 16px",
                fontFamily: "var(--font-mono)",
                fontSize: 12,
                fontWeight: active ? 700 : 500,
                color: active ? "var(--color-fg)" : "var(--color-fg-muted)",
                borderBottom: active ? "2px solid var(--color-fg)" : "2px solid transparent",
                cursor: "pointer",
                whiteSpace: "nowrap",
                marginBottom: -1,
                letterSpacing: "0.05em",
              }}
            >
              {t.label}
            </button>
          );
        })}
      </div>

      {/* Tab body */}
      <main
        style={{
          padding: isMobile ? "16px 14px 0" : "24px 32px 0",
          maxWidth: 920,
        }}
      >
        {tab === "mal" && (
          <section>
            {(loadingGoals || loadingCommittees || periodPending) && (
              <div style={{ fontSize: 13, color: "var(--color-fg-muted)" }}>Laddar mål...</div>
            )}

            {/* Without the committee list there are no areas to order the goals
                by, and every goal falls through to "Utan utskottsområde" — a
                section whose caption says the goal names no committee in our
                reading, which for all 73 of them is false. The goals are still
                facts, so they are shown; only the ordering is missing. */}
            {!loadingGoals && (committeesError || coverageError) && (
              <>
                <p className="text-sm text-on-surface-variant leading-relaxed mb-4">
                  Vi kunde inte hämta utskottsindelningen just nu, så målen visas
                  utan områden. Det säger inget om vilka områden målen rör —
                  försök igen.
                </p>
                <div className="space-y-3">
                  {goalList.map((g) => (
                    <GoalCard key={g.id} goal={g} party={party} />
                  ))}
                </div>
              </>
            )}

            {!loadingGoals && !loadingCommittees && !committeesError && !coverageError && !periodPending && (
              <>
                <p className="text-xs text-on-surface-variant leading-relaxed">
                  {goalList.length} mål från valmanifest och partiprogram, ordnade
                  efter utskottets område. Ett mål som rör flera områden visas under
                  vart och ett, så summan av områdena är större än antalet mål.{" "}
                  <SourceMarker sourceId="seed-party-goals" />
                </p>

                {coverage && (
                  <p className="text-[11px] text-on-surface-variant mt-1 mb-5 leading-relaxed">
                    {/* The record is bound to a named mandate period so votes from
                        the next parliament are never attributed to this one. */}
                    <span className="font-semibold">
                      {coverage.mandate.label}
                      {coverage.mandate.ended ? " (avslutad)" : ""}
                    </span>
                    {" · "}
                    {/* Completeness is a claim, so it is stated rather than implied.
                        The count we hold is read live; Riksdagen's count is fetched
                        on a schedule, so its date is stated too — a denominator of
                        unknown age is what let this figure drift toward looking more
                        complete than it was. */}
                    {coverage.ingested.toLocaleString("sv-SE")}
                    {coverage.expected > 0
                      ? ` av ${coverage.expected.toLocaleString("sv-SE")} omröstningar`
                      : " omröstningar (Riksdagens antal för perioden saknas)"}
                    <SourceMarker sourceId="riksdagen" />
                    {coverage.expected > 0 && coverage.denominatorCheckedAt && (
                      <> · Riksdagens antal hämtat {swedishDate(coverage.denominatorCheckedAt)}</>
                    )}
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

                {(committees ?? []).map((c) => {
                  const gg = byCommittee.get(c.code) ?? [];
                  return (
                    <section key={c.code} className="mb-8">
                      <h3 className="font-display text-base font-semibold text-on-surface mb-3 pb-1">
                        <Link to={`/committees/${c.code}`} className="hover:underline">
                          {c.name}
                        </Link>
                      </h3>
                      {gg.length > 0 ? (
                        <div className="space-y-3">
                          {gg.map((g) => (
                            <GoalCard key={g.id} goal={g} party={party} />
                          ))}
                        </div>
                      ) : (
                        /* Our seed reading is the gap, never the party's silence.
                           "Inga mål inom detta område" would read as the party
                           having nothing to say about the subject. */
                        <p className="text-sm text-on-surface-variant">
                          Vi har inga inlästa mål från {partyShortToName(party)} inom
                          detta område. <SourceMarker sourceId="seed-party-goals" />
                        </p>
                      )}
                    </section>
                  );
                })}

                {unplaced.length > 0 && (
                  <section className="mb-8">
                    <h3 className="font-display text-base font-semibold text-on-surface mb-3 pb-1">
                      Utan utskottsområde
                    </h3>
                    <p className="text-sm text-on-surface-variant mb-3">
                      Dessa mål saknar utskottsområde i vår inläsning, eller pekar på
                      ett utskott som inte finns i periodens register. De visas här så
                      att de inte försvinner ur listan.{" "}
                      <SourceMarker sourceId="seed-party-goals" />
                    </p>
                    <div className="space-y-3">
                      {unplaced.map((g) => (
                        <GoalCard key={g.id} goal={g} party={party} />
                      ))}
                    </div>
                  </section>
                )}

                {goalList.length === 0 && (
                  <p style={{ fontStyle: "italic", color: "var(--color-fg-muted)", fontSize: 13 }}>
                    Vi har inga inlästa mål från {partyShortToName(party)}.{" "}
                    <SourceMarker sourceId="seed-party-goals" />
                  </p>
                )}
              </>
            )}
          </section>
        )}

        {tab === "anforanden" && (
          <section>
            {loadingSpeeches && (
              <div style={{ fontSize: 13, color: "var(--color-fg-muted)" }}>Laddar anföranden...</div>
            )}
            {!loadingSpeeches && (!speeches || speeches.length === 0) && (
              <p style={{ fontStyle: "italic", color: "var(--color-fg-muted)", fontSize: 13 }}>
                Inga registrerade anföranden från {partyShortToName(party)} ännu.
              </p>
            )}
            {speeches && speeches.length > 0 && (
              <div>
                {speeches.map((s) => (
                  <SpeechRow key={s.id} speech={s} />
                ))}
              </div>
            )}
          </section>
        )}

        {tab === "politiker" && (
          <section>
            {loadingPoliticians && (
              <div style={{ fontSize: 13, color: "var(--color-fg-muted)" }}>Laddar politiker...</div>
            )}
            {!loadingPoliticians &&
              (!politiciansResp || politiciansResp.data.length === 0) && (
                <p style={{ fontStyle: "italic", color: "var(--color-fg-muted)", fontSize: 13 }}>
                  Inga politiker hittades för {partyShortToName(party)}.
                </p>
              )}
            {politiciansResp && politiciansResp.data.length > 0 && (
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: isMobile
                    ? "minmax(0, 1fr)"
                    : "repeat(auto-fill, minmax(220px, 1fr))",
                  gap: 8,
                }}
              >
                {politiciansResp.data.map((p) => (
                  <Link
                    key={p.intressentId}
                    to={`/politicians/${p.intressentId}`}
                    style={{
                      background: "var(--color-sdt-surface)",
                      border: "1px solid var(--color-border)",
                      padding: "10px 12px",
                      borderRadius: 4,
                      textDecoration: "none",
                      color: "inherit",
                      display: "flex",
                      alignItems: "center",
                      gap: 10,
                      minWidth: 0,
                    }}
                  >
                    {p.imageUrl ? (
                      <img
                        src={p.imageUrl}
                        alt=""
                        style={{
                          width: 32,
                          height: 32,
                          borderRadius: "50%",
                          objectFit: "cover",
                          flexShrink: 0,
                        }}
                      />
                    ) : (
                      <span
                        style={{
                          width: 32,
                          height: 32,
                          borderRadius: "50%",
                          background: pc?.bg ?? "#888",
                          color: pc?.text ?? "#fff",
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "center",
                          fontSize: 11,
                          fontFamily: "var(--font-mono)",
                          fontWeight: 700,
                          flexShrink: 0,
                        }}
                      >
                        {p.firstName[0]}{p.lastName[0]}
                      </span>
                    )}
                    <div style={{ minWidth: 0 }}>
                      <div
                        style={{
                          fontFamily: "var(--font-serif)",
                          fontSize: 14,
                          color: "var(--color-fg)",
                          overflow: "hidden",
                          textOverflow: "ellipsis",
                          whiteSpace: "nowrap",
                        }}
                      >
                        {p.firstName} {p.lastName}
                      </div>
                      {p.constituency && (
                        <div
                          style={{
                            fontFamily: "var(--font-mono)",
                            fontSize: 10,
                            color: "var(--color-fg-muted)",
                            overflow: "hidden",
                            textOverflow: "ellipsis",
                            whiteSpace: "nowrap",
                          }}
                        >
                          {p.constituency}
                        </div>
                      )}
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </section>
        )}
      </main>

      {partyMinisters.length > 0 && (
        <div
          className="mt-8"
          style={{ padding: isMobile ? "0 14px" : "0 32px", maxWidth: 920 }}
        >
          <h3 className="text-xs font-mono uppercase tracking-widest text-on-surface-variant mb-3">
            Statsråd från partiet
          </h3>
          <div className="bg-surface-lowest rounded-xl overflow-hidden">
            {partyMinisters.map((m, i) => (
              <Link
                key={m.id}
                to={`/regering/${m.id}`}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 12,
                  padding: "12px 20px",
                  borderTop: i > 0 ? "1px solid var(--color-surface-high)" : undefined,
                  textDecoration: "none",
                  color: "inherit",
                  transition: "background 0.15s",
                }}
                onMouseEnter={(e) => { e.currentTarget.style.background = "var(--color-surface-high)"; }}
                onMouseLeave={(e) => { e.currentTarget.style.background = "transparent"; }}
              >
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 14, fontWeight: 600 }}>{m.name}</div>
                  <div style={{ fontSize: 12, color: "var(--color-fg-muted)" }}>{m.title}</div>
                </div>
                <span style={{ fontSize: 12, color: "var(--color-fg-muted)" }}>→</span>
              </Link>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
