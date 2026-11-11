import { SourceMarker } from "@/components/sources/SourceMarker";
import type { Motion } from "./api";

type Decision = Motion["yrkanden"][number]["decisions"][number];

/** What the reasoning and reservation blocks read; both a yrkande's decision
 *  and a förslagspunkt on the betänkande page have these fields. */
export type DecisionParts = Pick<
  Decision,
  "betankandeSourceUrl" | "votedReservation" | "reasoning" | "reasoningAvailable" | "reservationsAvailable" | "reservations" | "sarskildaYttranden"
>;

export type Statement = Decision["reservations"][number];

function partyList(parties: string[]) {
  return parties.length > 0 ? ` (${parties.join(", ")})` : "";
}

/** "Not available" is stated, with the source: never filled in. */
function NotAvailable({ what, d }: { what: string; d: DecisionParts }) {
  return (
    <p className="mt-2 text-xs text-on-surface-variant" data-testid="not-available">
      {what} är inte tillgänglig från Riksdagen i maskinläsbar form.{" "}
      <a href={d.betankandeSourceUrl} className="text-link hover:underline" rel="noreferrer" target="_blank">
        Läs i betänkandet ↗
      </a>
    </p>
  );
}

/** Which alternative was put to the vote. It names the reservation Riksdagen
 *  names, which may be another party's than the motion's own. */
export function VotedAlternative({ d }: { d: DecisionParts }) {
  const v = d.votedReservation;
  if (!v) return null;
  const parties = v.parties.length > 0 ? v.parties : v.party ? [v.party] : [];
  return (
    <p className="mt-1 text-xs" data-testid="voted-alternative">
      I omröstningen ställdes utskottets förslag mot <strong>reservation {v.number}{partyList(parties)}</strong>. Övriga
      reservationer på punkten sattes inte under omröstning.
    </p>
  );
}

export function Reasoning({ d }: { d: DecisionParts }) {
  if (!d.reasoningAvailable) return <NotAvailable what="Utskottets ställningstagande" d={d} />;
  return (
    <section className="mt-3" data-testid="reasoning" aria-label="Utskottets ställningstagande">
      <h4 className="text-xs uppercase tracking-wider text-on-surface-variant">
        Utskottets ställningstagande
        <SourceMarker sourceId="betankanden" />
      </h4>
      {d.reasoning.split("\n\n").map((para, i) => (
        <p key={i} className="mt-1 text-sm">{para}</p>
      ))}
    </section>
  );
}

function StatementItem({ s, voted, kind }: { s: Statement; voted: boolean; kind: string }) {
  return (
    <li className="rounded border p-2" style={{ borderColor: "var(--color-border)" }} data-testid={kind}>
      <p className="text-sm font-medium">
        {kind === "reservation" ? "Reservation" : "Särskilt yttrande"} {s.number}
        {partyList(s.parties)}
        {voted ? <span className="ml-2 text-xs font-normal">· sattes under omröstning</span> : null}
      </p>
      {s.signers.length > 0 ? <p className="text-xs text-on-surface-variant">av {s.signers.join(", ")}</p> : null}
      {s.text.split("\n\n").map((para, i) => (
        <p key={i} className="mt-1 text-sm">{para}</p>
      ))}
    </li>
  );
}

export function Reservations({ d }: { d: DecisionParts }) {
  if (!d.reservationsAvailable) return <NotAvailable what="Reservationerna" d={d} />;
  if (d.reservations.length === 0 && d.sarskildaYttranden.length === 0) {
    return (
      <p className="mt-2 text-xs text-on-surface-variant" data-testid="no-reservations">
        Betänkandet har ingen reservation på den här punkten.
      </p>
    );
  }
  return (
    <section className="mt-3" data-testid="reservations" aria-label="Reservationer">
      <h4 className="text-xs uppercase tracking-wider text-on-surface-variant">
        Reservationer på punkten
        <SourceMarker sourceId="betankanden" />
      </h4>
      <ul className="mt-1 flex flex-col gap-2">
        {d.reservations.map((r) => (
          <StatementItem key={`r${r.number}`} s={r} kind="reservation" voted={d.votedReservation?.number === r.number} />
        ))}
        {d.sarskildaYttranden.map((r) => (
          <StatementItem key={`y${r.number}`} s={r} kind="sarskilt-yttrande" voted={false} />
        ))}
      </ul>
    </section>
  );
}

