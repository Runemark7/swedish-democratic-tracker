import { Link } from "react-router-dom";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { MANDATE_LABEL } from "./labels";
import type { Loefte, Valmanifest } from "./api";

/**
 * A party's Valmanifest, listed as sources: title, link, year and the party's
 * own marking (utkast or slutligt). Nothing here ranks or summarises them.
 */
export function ValmanifestList({ items }: { items: Valmanifest[] }) {
  if (items.length === 0) {
    return (
      <p className="text-sm text-on-surface-variant">
        Vi har inget registrerat valmanifest för partiet. Det betyder inte att
        partiet saknar ett.
      </p>
    );
  }
  return (
    <ul data-testid="valmanifest-list" className="space-y-2">
      {items.map((v) => (
        <li key={v.id} data-testid="valmanifest-row" className="text-sm">
          <a href={v.url} target="_blank" rel="noopener noreferrer" className="text-link underline">
            {v.title}
          </a>
          <span className="text-xs text-on-surface-variant">
            {" "}· {v.electionYear} · {v.state === "draft" ? "Utkast" : "Slutligt"}
            {v.publishedOn ? ` · ${v.publishedOn.slice(0, 10)}` : ""}
          </span>
        </li>
      ))}
    </ul>
  );
}

export function ArchiveNote() {
  return (
    <p data-testid="archive-note" className="text-xs text-on-surface-variant leading-relaxed">
      Arkiv {MANDATE_LABEL["2022-2026"]}: löften från valet 2022. Vi visar källan men
      ingen Löftesstatus, eftersom vi inte har kopplat dem till omröstningar.{" "}
      <SourceMarker sourceId="seed-party-goals" />
    </p>
  );
}

export function LoefteCard({ l, showParty }: { l: Loefte; showParty?: boolean }) {
  return (
    <article data-testid="loefte-card" className="bg-surface-lowest rounded-xl p-5">
      <Link to={`/loften/${l.id}`} className="text-sm font-semibold text-on-surface hover:underline">
        {showParty ? `${l.party}: ` : ""}
        {l.text}
      </Link>
      <div className="text-xs text-on-surface-variant mt-2">
        {l.archived ? "Arkiv " : ""}
        {MANDATE_LABEL[l.mandatePeriod] ?? l.mandatePeriod}
        {l.sourceDocument ? ` · ${l.sourceDocument}` : ""}
        {l.utskott.length > 0 ? ` · ${l.utskott.join(", ")}` : ""}{" "}
        <SourceMarker sourceId="seed-party-goals" />
      </div>
    </article>
  );
}
