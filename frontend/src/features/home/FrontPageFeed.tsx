import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { frontPageApi, type FrontPage } from "./api";
import { betankandePath } from "@/features/votes/betankandeApi";
import { motionPath } from "@/features/motions/api";
import { PartyBadge } from "@/shared/components";
import { PARTIES } from "@/shared/design";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { swedishDate } from "@/shared/dates";

type Beslut = FrontPage["decided"][number];
type Motion = FrontPage["motions"][number];

const tag = "text-xs text-on-surface-variant hover:text-link underline-offset-2 hover:underline";

function PartyMarks({ row }: { row: Beslut }) {
  if (row.parties.length === 0) return null;
  return (
    <ul className="flex flex-wrap gap-x-3 gap-y-1.5" aria-label="Partilinje per parti">
      {PARTIES.map((p) => (
        <li key={p}>
          <Link
            to={`/parties/${p}`}
            data-testid="party-mark"
            data-party={p}
            className="inline-flex items-center gap-1.5 text-xs text-on-surface-variant hover:text-link"
            title={`Partilinje för ${p}`}
          >
            <PartyBadge party={p} />
            <span className="font-mono">{row.parties.find((x) => x.party === p)?.partilinje || "–"}</span>
          </Link>
        </li>
      ))}
    </ul>
  );
}

function BeslutRow({ r }: { r: Beslut }) {
  const href = betankandePath(r.riksmote, r.beteckning);
  return (
    <li
      data-testid={r.decided ? "front-decided" : "front-upcoming"}
      className={`flex flex-col gap-2 py-3.5 px-4 rounded-lg ${r.decided ? "" : "border border-dashed"}`}
      style={{ background: r.decided ? "var(--color-surface-lowest)" : "transparent", borderColor: "var(--color-border)" }}
    >
      <Link to={href} className="text-sm font-semibold text-on-surface hover:text-link transition-colors">
        {r.title || `${r.riksmote}:${r.beteckning}`}
        <SourceMarker sourceId="betankanden" />
      </Link>
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
        <span className="text-xs text-on-surface-variant" data-testid="row-date">
          {r.decided ? "Beslutad " : "Planerad beslutsdag "}
          {r.date ? swedishDate(r.date) : "saknas"}
        </span>
        <Link to={`/beslut?utskott=${encodeURIComponent(r.utskott)}`} className={tag} data-testid="utskott-tag">
          {r.utskottName}
        </Link>
        <Link to={href} className="text-xs font-bold text-on-surface" data-testid="row-outcome">
          {r.outcome}
        </Link>
      </div>
      <PartyMarks row={r} />
    </li>
  );
}

function MotionRow({ m }: { m: Motion }) {
  const outcomeQ = encodeURIComponent(m.outcome.word);
  return (
    <li data-testid="front-motion" className="flex flex-col gap-1.5 py-3.5 px-4 rounded-lg" style={{ background: "var(--color-surface-lowest)" }}>
      <Link to={motionPath(m.riksmote, m.beteckning)} className="text-sm font-semibold text-on-surface hover:text-link transition-colors">
        {m.title || `Motion ${m.riksmote}:${m.beteckning}`}
        <SourceMarker sourceId="motions" />
      </Link>
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
        <span className="text-xs text-on-surface-variant">Avgjord {swedishDate(m.decidedAt)}</span>
        <span className="flex gap-1">
          {m.parties.map((p) => (
            <Link key={p} to={`/parties/${p}`} data-testid="motion-party" aria-label={`Parti ${p}`}>
              <PartyBadge party={p} />
            </Link>
          ))}
        </span>
        <Link to={`/motioner?kind=${encodeURIComponent(m.kind)}`} className={tag} data-testid="motion-kind">{m.kind}</Link>
        <Link to={`/motioner?outcome=${outcomeQ}`} className="text-xs font-bold text-on-surface" data-testid="motion-outcome">{m.outcome.word}</Link>
        {m.handledIn.map((h) => {
          const [rm, bet] = h.split(":");
          return (
            <Link key={h} to={betankandePath(rm, bet)} className={`${tag} font-mono`} data-testid="motion-handled-in">{h}</Link>
          );
        })}
      </div>
    </li>
  );
}

function Heading({ children, to, linkText }: { children: React.ReactNode; to: string; linkText: string }) {
  return (
    <div className="flex items-baseline justify-between mb-2">
      <h2 className="text-xs uppercase tracking-wider text-on-surface-variant">{children}</h2>
      <Link to={to} className="text-xs font-semibold text-link">{linkText} →</Link>
    </div>
  );
}

/** The front page feed: decided betänkanden, recent motion outcomes and the
 *  next planned betänkanden, each in its own section. Every row, tag and
 *  party mark is a link. */
export function FrontPageFeed() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["frontpage"],
    queryFn: frontPageApi.get,
    staleTime: 60_000,
  });
  if (isLoading) return <p className="text-sm text-on-surface-variant px-8 py-6">Laddar de senaste besluten...</p>;
  if (error || !data) return <p className="text-sm text-on-surface-variant px-8 py-6">Vi kunde inte hämta de senaste besluten just nu. Försök igen.</p>;

  const { coverage } = data;
  return (
    <div className="px-4 sm:px-8 py-4 space-y-8">
      <section aria-label="Senaste besluten">
        <Heading to="/beslut?state=decided" linkText="Alla beslut">Senaste besluten</Heading>
        {coverage.stale && (
          <p data-testid="coverage" className="text-xs text-on-surface-variant mb-2">
            {coverage.newest && coverage.oldest
              ? `Besluten nedan avser ${swedishDate(coverage.oldest)} – ${swedishDate(coverage.newest)}. Inget nyare beslut finns i vår inläsning av Riksdagens dokumentlista.`
              : "Inget beslutat betänkande finns i vår inläsning av Riksdagens dokumentlista just nu."}
            <SourceMarker sourceId="riksdagen" />
          </p>
        )}
        <ul className="space-y-2">{data.decided.map((r) => <BeslutRow key={`${r.riksmote}:${r.beteckning}`} r={r} />)}</ul>
      </section>

      <section aria-label="Senaste motionsutfall">
        <Heading to="/motioner" linkText="Alla motioner">Senaste motionsutfall</Heading>
        {data.motions.length === 0 ? (
          <p className="text-xs text-on-surface-variant">Inga avgjorda motioner i vår inläsning ännu.</p>
        ) : (
          <ul className="space-y-2">{data.motions.map((m) => <MotionRow key={m.dokId} m={m} />)}</ul>
        )}
      </section>

      <section aria-label="Kommande beslut">
        <Heading to="/beslut?state=upcoming" linkText="Alla kommande beslut">Kommande beslut</Heading>
        <p className="text-xs text-on-surface-variant mb-2">Ännu inte beslutade. Datumet är den planerade beslutsdagen.</p>
        {data.upcoming.length === 0 ? (
          <p className="text-xs text-on-surface-variant">Inga planerade betänkanden i vår inläsning.</p>
        ) : (
          <ul className="space-y-2">{data.upcoming.map((r) => <BeslutRow key={`${r.riksmote}:${r.beteckning}`} r={r} />)}</ul>
        )}
      </section>
    </div>
  );
}
