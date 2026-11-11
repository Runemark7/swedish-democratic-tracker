import { Link } from "react-router-dom";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { swedishDate } from "@/shared/dates";
import { betankandePath, type Motion } from "./api";


function Step({ title, id, children }: { title: string; id: string; children: React.ReactNode }) {
  return (
    <li className="rounded border p-3 text-sm" style={{ borderColor: "var(--color-border)" }} data-testid="timeline-step" data-step={id}>
      <p className="text-xs uppercase tracking-wider text-on-surface-variant">{title}</p>
      <div className="mt-1 flex flex-col gap-1">{children}</div>
    </li>
  );
}

const linkCls = "text-link hover:underline";

/** The motion's path in order: filed, utskott, chamber debate, decision. Each
 *  step states a fact from the record and links out; a step with no record
 *  says so instead of being left out. */
export function Timeline({ motion }: { motion: Motion }) {
  const decisions = motion.yrkanden.flatMap((y) => y.decisions);
  const seen = new Set<string>();
  const unique = decisions.filter((d) => {
    const k = `${d.betankandeRiksmote}:${d.betankandeBeteckning}:${d.punkt}`;
    if (seen.has(k)) return false;
    seen.add(k);
    return true;
  });
  const debate = unique.map((d) => d.debate).find((x) => x && x.total > 0);
  const decided = motion.outcome.word !== "Ej avgjord än" || unique.length > 0;

  return (
    <section aria-labelledby="timeline-heading" className="my-6" data-testid="timeline">
      <h2 id="timeline-heading" className="text-lg font-semibold mb-2">Motionens väg</h2>
      <ol className="grid gap-3" style={{ gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))" }}>
        <Step title="Inlämnad" id="filed">
          <span>{motion.date ? swedishDate(motion.date) : "Datum saknas"}</span>
          <a href={motion.sourceUrl} className={linkCls} rel="noreferrer" target="_blank">Motionen hos Riksdagen ↗</a>
        </Step>
        <Step title="I utskottet" id="utskott">
          <span>{motion.utskottName || motion.utskott || "Utskott ej angivet"}</span>
          {unique.length > 0 ? (
            unique.map((d) => (
              <Link key={`${d.betankandeRiksmote}:${d.betankandeBeteckning}:${d.punkt}`} to={betankandePath(d)} className={linkCls}>
                Betänkande {d.betankandeRiksmote}:{d.betankandeBeteckning}, punkt {d.punkt}
              </Link>
            ))
          ) : (
            <span className="text-on-surface-variant">
              {motion.behandlasI ? `Behandlas i ${motion.behandlasI}` : "Betänkande ej angivet"}
            </span>
          )}
        </Step>
        <Step title="I kammaren" id="chamber">
          {debate ? (
            <>
              <span>Debatt{debate.date ? ` ${swedishDate(debate.date)}` : ""} · {debate.total} anföranden</span>
              <a href="#yrkanden-heading" className={linkCls}>Se debatten nedan</a>
            </>
          ) : (
            <span className="text-on-surface-variant">Inga anföranden är kopplade till betänkandet.</span>
          )}
        </Step>
        <Step title="Beslut" id="decision">
          {decided ? (
            <>
              <strong>{motion.outcome.word}</strong>
              {unique.map((d) => (
                <span key={`${d.betankandeRiksmote}:${d.betankandeBeteckning}:${d.punkt}`}>
                  {d.modeLabel}{" "}
                  {d.voteringUrl ? (
                    <a href={d.voteringUrl} className={linkCls} rel="noreferrer" target="_blank">Omröstningen hos Riksdagen ↗</a>
                  ) : null}
                </span>
              ))}
            </>
          ) : (
            <strong>Ej avgjord än</strong>
          )}
        </Step>
        <Step title="Uppföljning" id="followup">
          <a href="#questions-heading" className={linkCls}>Frågor och interpellationer</a>
        </Step>
      </ol>
      <SourceMarker sourceId="motions" />
    </section>
  );
}
