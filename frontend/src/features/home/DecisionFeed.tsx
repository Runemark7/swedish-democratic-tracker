import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { votesApi } from "@/features/votes/api";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { swedishDate } from "@/shared/dates";
import { COMMITTEES, committeeFromBeteckning } from "@/shared/design";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import type { components } from "@/shared/api-contract";

type RecentBetankande = components["schemas"]["RecentBetankande"];

const FEED_COUNT = 40;
const PREVIEW = 8;

/** A row is decided only when the API says so and Riksdagen's own fields agree.
 *  `beslutsdag` is filled for planned items (the scheduled day), so a future
 *  date or status "planerat" is not a decision even if `decided` is stale.
 *  `decided` itself is the backend Decided() reading of `beslutad`. */
function rowDecided(item: { decided: boolean; status: string; decisionDate: string }): boolean {
  if (item.status.trim().toLowerCase() === "planerat") return false;
  if (item.decisionDate && item.decisionDate > localISODate()) return false;
  return item.decided;
}

function localISODate(d = new Date()): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

function committeeLabel(organ: string): string {
  return COMMITTEES[organ] ?? committeeFromBeteckning(organ) ?? organ;
}

/** Beteckning numbering restarts each riksmöte, so the link carries `rm`. */
function beslutHref(item: RecentBetankande): string {
  const path = `/beslut/${encodeURIComponent(item.beteckning)}`;
  if (!item.riksmote) return path;
  return `${path}?rm=${encodeURIComponent(item.riksmote)}`;
}

/**
 * Recent betänkanden, split into what the chamber has decided and what is
 * still planned. Every row links to the betänkande, not to the committee.
 */
export function DecisionFeed() {
  const isMobile = useMediaQuery("(max-width: 640px)");
  const [showAllDecided, setShowAllDecided] = useState(false);
  const [showAllUpcoming, setShowAllUpcoming] = useState(false);

  const { data, isLoading, isError } = useQuery({
    queryKey: ["votes-recent", FEED_COUNT],
    queryFn: () => votesApi.listRecent(FEED_COUNT),
    staleTime: 60_000,
  });

  const items = data ?? [];
  const decided = items.filter(rowDecided).sort((a, b) => b.decisionDate.localeCompare(a.decisionDate));
  const upcoming = items
    .filter((item) => !rowDecided(item))
    .sort((a, b) => {
      if (!a.decisionDate && !b.decisionDate) return 0;
      if (!a.decisionDate) return 1;
      if (!b.decisionDate) return -1;
      return a.decisionDate.localeCompare(b.decisionDate);
    });

  let oldest: string | null = null;
  let newest: string | null = null;
  for (const item of decided) {
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

  return (
    <section style={{ margin: isMobile ? "14px 14px 6px" : "22px 32px 6px" }}>
      {isLoading && <p style={textStyle}>Laddar de senaste betänkandena...</p>}

      {isError && items.length === 0 && (
        <p style={textStyle}>
          Vi kunde inte hämta de senaste betänkandena just nu. Försök igen.
        </p>
      )}

      {!isLoading && !isError && items.length === 0 && (
        <p style={{ ...textStyle, fontStyle: "italic" }}>
          Vi har inga betänkanden inlästa just nu. Det säger inget om vad
          riksdagen faktiskt beslutat — försök igen om en stund.
          <SourceMarker sourceId="riksdagen" />
        </p>
      )}

      {items.length > 0 && (
        <>
          <FeedBlock
            heading="Beslutat"
            isMobile={isMobile}
            items={showAllDecided ? decided : decided.slice(0, PREVIEW)}
            hiddenCount={showAllDecided ? 0 : Math.max(0, decided.length - PREVIEW)}
            onShowAll={() => setShowAllDecided(true)}
            showAllLabel={`Visa alla beslutade i den här inläsningen (${decided.length})`}
            intro={
              decided.length === 0 ? (
                <p style={{ ...textStyle, fontSize: 12, marginBottom: 10 }}>
                  Inget betänkande i den här inläsningen är markerat som beslutat.
                  <SourceMarker sourceId="riksdagen" />
                </p>
              ) : oldest && newest ? (
                <p style={{ ...textStyle, fontSize: 12, marginBottom: 10 }}>
                  Beslutade {swedishDate(oldest)} – {swedishDate(newest)}. Inget
                  nyare beslutat betänkande finns i vår inläsning av Riksdagens
                  dokumentlista.
                  <SourceMarker sourceId="riksdagen" />
                </p>
              ) : null
            }
          />
          <FeedBlock
            heading="Kommande"
            isMobile={isMobile}
            items={showAllUpcoming ? upcoming : upcoming.slice(0, PREVIEW)}
            hiddenCount={showAllUpcoming ? 0 : Math.max(0, upcoming.length - PREVIEW)}
            onShowAll={() => setShowAllUpcoming(true)}
            showAllLabel={`Visa alla kommande i den här inläsningen (${upcoming.length})`}
            intro={
              <p style={{ ...textStyle, fontSize: 12, marginBottom: 10 }}>
                {upcoming.length === 0
                  ? "Inga planerade betänkanden i den här inläsningen."
                  : "Ännu inte beslutade. Ett ifyllt beslutsdatum kan vara dagen kammaren planerar att besluta."}
                <SourceMarker sourceId="riksdagen" />
              </p>
            }
          />
        </>
      )}
    </section>
  );
}

function FeedBlock({
  heading,
  intro,
  items,
  isMobile,
  hiddenCount,
  onShowAll,
  showAllLabel,
}: {
  heading: string;
  intro: React.ReactNode;
  items: RecentBetankande[];
  isMobile: boolean;
  hiddenCount: number;
  onShowAll: () => void;
  showAllLabel: string;
}) {
  const metaStyle: React.CSSProperties = {
    fontFamily: "var(--font-mono)",
    fontSize: 10,
    color: "var(--color-fg-muted)",
  };

  return (
    <div style={{ marginBottom: 28 }}>
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
        {heading}
      </header>
      {intro}
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
          {items.map((item, index) => {
            const name = committeeLabel(item.organ);
            const decided = rowDecided(item);
            const whenLabel = item.decisionDate ? swedishDate(item.decisionDate) : "datum saknas";
            const when = (
              <span
                style={{ ...metaStyle, fontStyle: decided ? "normal" : "italic" }}
                title={
                  decided
                    ? "Beslutsdag"
                    : item.decisionDate
                      ? `Planerad beslutsdag ${swedishDate(item.decisionDate)}`
                      : "Ännu inte beslutat"
                }
              >
                {whenLabel}
              </span>
            );
            return (
              <Link
                key={`${item.riksmote}-${item.beteckning}-${index}`}
                to={beslutHref(item)}
                style={{
                  background: "var(--color-sdt-surface)",
                  padding: isMobile ? "10px 14px" : "10px 16px",
                  display: "grid",
                  gridTemplateColumns: isMobile ? "minmax(0, 1fr) auto" : "132px minmax(0, 1fr) auto",
                  gap: 10,
                  alignItems: "baseline",
                  textDecoration: "none",
                  color: "inherit",
                  minWidth: 0,
                }}
              >
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
                  title={item.organ}
                >
                  {name}
                </span>
              </Link>
            );
          })}
        </div>
      )}
      {hiddenCount > 0 && (
        <button
          type="button"
          onClick={onShowAll}
          style={{
            marginTop: 10,
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            letterSpacing: "0.04em",
            color: "var(--color-accent)",
            background: "none",
            border: "none",
            padding: 0,
            cursor: "pointer",
          }}
        >
          {showAllLabel}
        </button>
      )}
    </div>
  );
}
