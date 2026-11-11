import { useState } from "react";
import { Link } from "react-router-dom";
import { PartyBadge } from "@/shared/components";
import { swedishDate } from "@/shared/dates";
import { SourceMarker } from "@/components/sources/SourceMarker";
import type { components } from "@/shared/api-contract";

type DebateData = components["schemas"]["Debate"];

const RULE_LABEL: Record<string, string> = {
  rel_dok_id: "kopplad via Riksdagens dokumentreferens",
  avsnittsrubrik_datum: "kopplad via debattens rubrik och datum",
};

const SHOWN = 5;

/** The chamber debate on a betänkande: who spoke, for which party, the start of
 *  each speech verbatim, and links to the full speech and to Riksdagen. Nothing
 *  is selected or summarised; how each speech was tied to the betänkande is
 *  stated. */
export function Debate({ debate, heading = "Debatten" }: { debate: DebateData | null | undefined; heading?: string }) {
  const [all, setAll] = useState(false);
  const speeches = debate?.speeches ?? [];
  const total = debate?.total ?? 0;
  return (
    <section className="mt-4" data-testid="debate">
      <h2 className="text-xs uppercase tracking-wider text-on-surface-variant">
        {heading}
        {total > 0 ? ` · ${total} anföranden` : ""}
        {debate?.date ? ` · ${swedishDate(debate.date)}` : ""}
        <SourceMarker sourceId="riksdagen" />
      </h2>
      {speeches.length === 0 ? (
        <p className="text-sm italic text-on-surface-variant" data-testid="debate-empty">
          Inga anföranden är kopplade till betänkandet.
        </p>
      ) : (
        <>
          <ul>
            {(all ? speeches : speeches.slice(0, SHOWN)).map((s) => (
              <li key={s.speechId} className="py-3 border-b" style={{ borderColor: "var(--color-border)" }} data-testid="debate-speech">
                <div className="flex items-center gap-2 flex-wrap text-sm">
                  <Link to={`/politicians/${encodeURIComponent(s.politicianId)}`} className="text-link hover:underline">{s.name}</Link>
                  <PartyBadge party={s.party} />
                  <span className="text-xs text-on-surface-variant font-mono">{swedishDate(s.date)}</span>
                </div>
                <p className="text-sm mt-1">
                  {s.excerptAvailable ? s.excerpt : <em className="text-on-surface-variant">Anförandetexten är inte hämtad än.</em>}
                </p>
                <p className="text-xs mt-1 flex gap-3 flex-wrap">
                  <Link to={`/anforanden/${s.speechId}`} className="text-link hover:underline">Läs hela anförandet →</Link>
                  <a href={s.riksdagenUrl} className="text-link hover:underline" rel="noreferrer" target="_blank">Hos Riksdagen ↗</a>
                  <span className="text-on-surface-variant">{RULE_LABEL[s.linkRule] ?? s.linkRule}</span>
                </p>
              </li>
            ))}
          </ul>
          {speeches.length > SHOWN && !all ? (
            <button className="mt-2 text-xs text-link" onClick={() => setAll(true)}>
              Visa alla {speeches.length} anföranden →
            </button>
          ) : null}
          {total > speeches.length ? (
            <p className="mt-2 text-xs text-on-surface-variant">Visar {speeches.length} av {total} anföranden.</p>
          ) : null}
        </>
      )}
    </section>
  );
}
