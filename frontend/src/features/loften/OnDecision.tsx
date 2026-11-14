import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { PartyBadge } from "@/shared/components";
import { partyShortToName } from "@/shared/design";
import { MY_PARTY_RING, useMyParty } from "@/contexts/myParty";
import { loftenApi, type LoefteRef, type LoftePartySummary } from "./api";
import { MANDATPERIODER } from "./labels";

/** The period the party summaries count; the 2022–2026 archive is never mixed in. */
const CURRENT_PERIOD = MANDATPERIODER[0];

const betankandeKey = (riksmote: string, beteckning: string) => `${riksmote}:${beteckning}`;

/** "berör N löften": links to exactly the Löften with a Bekräftad koppling
 *  to a votering of this betänkande. Renders nothing for zero. */
export function BerorLoften({ riksmote, beteckning, count }: { riksmote: string; beteckning: string; count: number }) {
  if (!count || count <= 0) return null;
  return (
    <Link
      to={`/loften?betankande=${encodeURIComponent(betankandeKey(riksmote, beteckning))}`}
      className="text-xs text-on-surface-variant hover:text-link underline-offset-2 hover:underline"
      data-testid="beror-loften"
    >
      berör {count} {count === 1 ? "löfte" : "löften"}
    </Link>
  );
}

/** The Löften a decision touches, as plain links. No status: a Löfte is listed
 *  here because a person confirmed a vote in this decision tests it. */
export function LoefteRefList({ items, heading }: { items: LoefteRef[]; heading: string }) {
  // A response from before this field existed carries none: show nothing.
  if (!items || items.length === 0) return null;
  return (
    <section className="mt-6" data-testid="decision-loften" aria-label={heading}>
      <h2 className="text-lg font-semibold mb-1">
        {heading}
        <SourceMarker sourceId="loften-valmanifest" />
      </h2>
      <p className="text-xs text-on-surface-variant mb-2">
        Löften där redaktören bekräftat att en omröstning här prövar löftet. Det säger inget om hur partiet röstade.
      </p>
      <ul className="space-y-1">
        {items.map((l) => (
          <li key={l.id} className="text-sm flex items-center gap-2" data-testid="decision-loefte">
            <PartyBadge party={l.party} />
            <Link to={`/loften/${l.id}`} className="text-link hover:underline">{l.text}</Link>
          </li>
        ))}
      </ul>
    </section>
  );
}

/** Eight tiles in the order the API gives (alphabetical by party code). Mitt
 *  parti gets a ring; it is never moved, filtered or ranked. */
export function PartyLoefteTiles({ tiles }: { tiles: LoftePartySummary[] }) {
  const { myParty } = useMyParty();
  return (
    <section aria-label="Löften per parti" data-testid="party-tiles">
      <div className="flex items-baseline justify-between mb-2">
        <h2 className="text-xs uppercase tracking-wider text-on-surface-variant">
          Löften per parti
          <SourceMarker sourceId="loften-valmanifest" />
        </h2>
        <Link to="/loften" className="text-xs font-semibold text-link">Alla löften →</Link>
      </div>
      <ul className="grid gap-2" style={{ gridTemplateColumns: "repeat(auto-fill, minmax(190px, 1fr))" }}>
        {tiles.map((t) => (
          <li key={t.party}>
            <Link
              to={`/loften?party=${t.party}`}
              data-testid="party-tile"
              data-party={t.party}
              data-my-party={t.party === myParty ? "true" : undefined}
              className="block rounded-lg p-3 bg-surface-lowest hover:underline"
              style={t.party === myParty ? { boxShadow: MY_PARTY_RING } : undefined}
            >
              <span className="flex items-center gap-2 text-sm font-semibold">
                <PartyBadge party={t.party} /> {partyShortToName(t.party)}
              </span>
              <CountsLine t={t} />
            </Link>
          </li>
        ))}
      </ul>
    </section>
  );
}

function CountsLine({ t }: { t: LoftePartySummary }) {
  const own = t.total - t.arkiv;
  if (own === 0) {
    return <span className="block mt-1 text-xs text-on-surface-variant">Inga löften publicerade ännu</span>;
  }
  return (
    <span className="block mt-1 text-xs text-on-surface-variant">
      {own} löften · Röstat för {t.for} · Röstat emot {t.emot} · Blandat {t.blandat} · Inte prövat än {t.inteProvat}
    </span>
  );
}

/** A party's Löftesstatus counts, linking to its Löften. */
export function PartyLoefteSummary({ party }: { party: string }) {
  const { data } = useQuery({ queryKey: ["loften-summary", CURRENT_PERIOD], queryFn: () => loftenApi.summary(CURRENT_PERIOD), staleTime: 60_000 });
  const t = data?.find((x) => x.party === party);
  if (!t) return null;
  return (
    <p className="text-sm mb-4" data-testid="party-loefte-summary">
      <Link to={`/loften?party=${party}`} className="text-link underline">Löften</Link>
      {": "}
      {t.total - t.arkiv === 0 ? "inga publicerade löften för 2026–2030 ännu" : (
        <>
          {t.total - t.arkiv} · Röstat för {t.for} · Röstat emot {t.emot} · Blandat {t.blandat} · Inte prövat än {t.inteProvat}
        </>
      )}
      <SourceMarker sourceId="loften-valmanifest" />
    </p>
  );
}
