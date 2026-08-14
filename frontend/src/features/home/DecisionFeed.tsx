import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { votesApi } from "@/features/votes/api";
import { committeesApi } from "@/features/committees/api";
import { useRecordCoverage } from "@/hooks/useDemocracy";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { swedishDate } from "@/shared/dates";
import { useMediaQuery } from "@/hooks/useMediaQuery";

const FEED_COUNT = 40;

/**
 * A complete, dated feed of recent betänkanden across every committee — the
 * front page's "what just happened" surface (#88). The betänkande is the unit
 * rather than the förslagspunkt: a sitting week is ~10-15 betänkanden instead
 * of hundreds of near-identical vote points (#100).
 *
 * The unit is the betänkande, not the decision, and the difference is not
 * cosmetic. Riksdagen's list carries betänkanden that have only been *planned*
 * — on 2026-08-14, 15 of 40, five of them in the first eight rows, two with a
 * justeringsdag four months in the future. It sorts on `datum`, so they arrive
 * at the top. A block headed "senaste besluten" therefore published scheduled
 * documents as decisions that had already happened.
 *
 * `datum` is also not a decision date: on every decided betänkande checked,
 * `beslutsdag` fell one to six days later. So a decided row is dated by
 * `decisionDate`, and a planned row makes no date claim at all.
 *
 * No committee filter and no curation. The span is stated from the items' own
 * decision dates, never as a computed "sitting period" — Riksdagen publishes no
 * sourceable sitting calendar, so a threshold for "in recess" would be an
 * inference we cannot source (#98).
 */
export function DecisionFeed() {
  const isMobile = useMediaQuery("(max-width: 640px)");

  const { data, isLoading, isError } = useQuery({
    queryKey: ["votes-recent", FEED_COUNT],
    queryFn: () => votesApi.listRecent(FEED_COUNT),
    staleTime: 60_000,
  });

  // Same query key as CommitteeIndex, so this shares its cache rather than
  // fetching the list twice on one page.
  const { data: coverage } = useRecordCoverage();
  const period = coverage?.mandate.code;
  const { data: committees } = useQuery({
    queryKey: ["committees", period],
    queryFn: () => committeesApi.listCommittees(period!),
    enabled: !!period,
    staleTime: 5 * 60_000,
  });

  // The feed is unscoped; /committees/:code is resolved against the current
  // mandate period. A code the period does not hold would link to a page
  // truthfully answering "we have no such committee" — so a row only becomes a
  // link once we know the destination exists. The feed already carries rows
  // from the coming riksmöte, so this is not hypothetical.
  const committeeNames = new Map((committees ?? []).map((c) => [c.code, c.name]));

  const items = data ?? [];

  // Span is computed from decision dates only. Including planned rows would
  // stretch it across dates on which nothing was decided.
  let oldest: string | null = null;
  let newest: string | null = null;
  let plannedCount = 0;
  for (const item of items) {
    if (!item.decided) {
      plannedCount += 1;
      continue;
    }
    const d = item.decisionDate;
    if (!d) continue;
    if (!oldest || d < oldest) oldest = d;
    if (!newest || d > newest) newest = d;
  }

  const textStyle: React.CSSProperties = {
    fontFamily: "var(--font-body)",
    fontSize: 13,
    color: "var(--color-fg-muted)",
    margin: 0,
  };
  const metaStyle: React.CSSProperties = {
    fontFamily: "var(--font-mono)",
    fontSize: 10,
    color: "var(--color-fg-muted)",
  };

  return (
    <section style={{ margin: isMobile ? "14px 14px 6px" : "22px 32px 6px" }}>
      <header
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 10,
          letterSpacing: "0.15em",
          color: "var(--color-fg-muted)",
          textTransform: "uppercase",
          marginBottom: 10,
        }}
      >
        SENASTE BETÄNKANDENA
      </header>

      {oldest && newest && (
        <p style={{ ...textStyle, fontSize: 12, marginBottom: 4 }}>
          Betänkandena i listan beslutades {swedishDate(oldest)} – {swedishDate(newest)}.
          {/* Scoped to this list on purpose. "Inget nyare i vår inläsning" would
              contradict the coverage figure rendered further down the same page,
              which counts voteringar rather than betänkanden. */}
          {" "}Inget nyare betänkande finns i vår inläsning av Riksdagens
          dokumentlista.
          <SourceMarker sourceId="riksdagen" />
        </p>
      )}

      {plannedCount > 0 && (
        <p style={{ ...textStyle, fontSize: 12, marginBottom: 14 }}>
          {plannedCount === 1
            ? "Ett av betänkandena nedan är planerat och ännu inte beslutat."
            : `${plannedCount} av betänkandena nedan är planerade och ännu inte beslutade.`}{" "}
          De saknar beslutsdatum tills riksdagen har beslutat.
        </p>
      )}

      {isLoading && <p style={textStyle}>Laddar de senaste betänkandena...</p>}

      {/* A transport failure has no source — it says nothing about what the
          Riksdag has decided, only that we could not ask right now. Shown only
          when there is nothing cached to show: TanStack keeps `data` across a
          failed background refetch, and an error line above a populated list
          would contradict the list beneath it. */}
      {isError && items.length === 0 && (
        <p style={textStyle}>
          Vi kunde inte hämta de senaste betänkandena just nu. Försök igen.
        </p>
      )}

      {/* The upstream Riksdagen list is intermittently empty on an identical
          query (confirmed live, seconds apart, and again for several minutes).
          A zero-length response is therefore our reading failing, never a
          statement that the Riksdag decided nothing. */}
      {!isLoading && !isError && items.length === 0 && (
        <p style={{ ...textStyle, fontStyle: "italic" }}>
          Vi har inga betänkanden inlästa just nu. Det säger inget om vad
          riksdagen faktiskt beslutat — försök igen om en stund.
          <SourceMarker sourceId="riksdagen" />
        </p>
      )}

      {items.length > 0 && (
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            gap: 1,
            background: "var(--color-border)",
            border: "1px solid var(--color-border)",
          }}
        >
          {items.map((item, i) => {
            const committeeName = committeeNames.get(item.organ);
            const rowStyle: React.CSSProperties = {
              background: "var(--color-sdt-surface)",
              padding: isMobile ? "10px 14px" : "10px 16px",
              display: "grid",
              gridTemplateColumns: isMobile
                ? "minmax(0, 1fr) auto"
                : "104px minmax(0, 1fr) auto",
              gap: 10,
              alignItems: "baseline",
              textDecoration: "none",
              color: "inherit",
              minWidth: 0,
            };

            // A decided row is dated by its decision. A planned one states its
            // status instead — showing `datum` there would date an event that
            // has not happened.
            const when = item.decided ? (
              <span style={metaStyle}>{swedishDate(item.decisionDate)}</span>
            ) : (
              <span style={{ ...metaStyle, fontStyle: "italic" }}>
                ännu inte beslutat
              </span>
            );

            const body = (
              <>
                {!isMobile && when}
                <span
                  style={{
                    fontSize: 13,
                    color: "var(--color-fg)",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                    minWidth: 0,
                  }}
                  title={item.title}
                >
                  {item.title}
                  {isMobile && <span style={{ display: "block" }}>{when}</span>}
                </span>
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 11,
                    fontWeight: 700,
                    color: "var(--color-accent)",
                    letterSpacing: "0.05em",
                    whiteSpace: "nowrap",
                  }}
                  title={committeeName ?? item.organ}
                >
                  {/* Named, not just coded: "UbU" is the shorthand this site
                      exists to decode. The code stays alongside it because it
                      is what the beteckning is built from. */}
                  {committeeName ?? item.organ}
                </span>
              </>
            );

            return committeeName ? (
              <Link
                key={`${item.beteckning}-${i}`}
                to={`/committees/${encodeURIComponent(item.organ)}`}
                style={rowStyle}
              >
                {body}
              </Link>
            ) : (
              <div key={`${item.beteckning}-${i}`} style={rowStyle}>
                {body}
              </div>
            );
          })}
        </div>
      )}
    </section>
  );
}
