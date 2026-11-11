import { Link } from "react-router-dom";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { betankandePath } from "@/features/votes/betankandeApi";
import { motionPath } from "@/features/motions/api";
import { swedishDate } from "@/shared/dates";
import { partyShortToName } from "@/shared/design";
import type { PartyRecord } from "./api";

/** Recent decisions with the party's Partilinje; each row links to the betänkande page. */
export function PartyDecisions({ party, record }: { party: string; record: PartyRecord }) {
  if (record.decisions.length === 0) {
    return (
      <p className="text-sm italic text-on-surface-variant">
        Inga avgjorda beslut hittades för {partyShortToName(party)} i vår inläsning.
      </p>
    );
  }
  return (
    <ul className="space-y-2" data-testid="party-decisions">
      {record.decisions.map((d) => (
        <li key={`${d.riksmote}:${d.beteckning}`} data-testid="party-decision-row">
          <Link
            to={betankandePath(d.riksmote, d.beteckning)}
            className="flex flex-col gap-1.5 py-3.5 px-4 rounded-lg hover:shadow-ambient group"
            style={{ background: "var(--color-surface-lowest)" }}
          >
            <p className="text-sm font-semibold text-on-surface group-hover:text-link transition-colors">
              {d.title || `${d.riksmote}:${d.beteckning}`}
              <SourceMarker sourceId="betankanden" />
            </p>
            <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-on-surface-variant">
              <span className="font-mono">{d.riksmote}:{d.beteckning}</span>
              <span>{d.utskottName}</span>
              <span>{d.date ? `Beslutad ${swedishDate(d.date)}` : "Beslutsdatum saknas"}</span>
              <span>{d.outcome}</span>
              <strong className="text-on-surface" data-testid="party-decision-partilinje">
                {d.partilinje ? `Partilinje: ${d.partilinje}` : "Ingen omröstning med partiets röster"}
              </strong>
            </div>
          </Link>
        </li>
      ))}
    </ul>
  );
}

/** The party's Kommittémotioner and what became of them. */
export function PartyMotions({ party, record }: { party: string; record: PartyRecord }) {
  if (record.motions.length === 0) {
    return (
      <p className="text-sm italic text-on-surface-variant">
        Inga kommittémotioner från {partyShortToName(party)} i vår inläsning.
      </p>
    );
  }
  return (
    <ul className="space-y-2" data-testid="party-motions">
      {record.motions.map((m) => (
        <li key={m.dokId} data-testid="party-motion-row">
          <Link
            to={motionPath(m.riksmote, m.beteckning)}
            className="flex flex-col gap-1.5 py-3.5 px-4 rounded-lg hover:shadow-ambient group"
            style={{ background: "var(--color-surface-lowest)" }}
          >
            <p className="text-sm font-semibold text-on-surface group-hover:text-link transition-colors">
              {m.title || `Motion ${m.riksmote}:${m.beteckning}`}
              <SourceMarker sourceId="motions" />
            </p>
            <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-on-surface-variant">
              <span className="font-mono">{m.riksmote}:{m.beteckning}</span>
              <span>{m.utskottName}</span>
              <span>{m.date ? swedishDate(m.date) : "Datum saknas"}</span>
              <strong className="text-on-surface" data-testid="party-motion-outcome">{m.outcome.word}</strong>
              <SourceMarker sourceId="betankanden" />
            </div>
          </Link>
        </li>
      ))}
    </ul>
  );
}
