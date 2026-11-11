import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { motionsApi, type Motion } from "./api";
import { ApiError } from "@/shared/api-client";
import { PartyBadge } from "@/shared/components";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { swedishDate } from "@/shared/dates";
import { VotedAlternative, Reasoning, Reservations } from "./DecisionParts";
import { Debate } from "./Debate";

function Notice({ children }: { children: React.ReactNode }) {
  return (
    <div className="sdt-page px-8 py-16">
      <p className="text-sm text-on-surface-variant">{children}</p>
    </div>
  );
}

/** One label and value from the record, shown as given. */
function Fact({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <dt className="text-xs uppercase tracking-wider text-on-surface-variant">{label}</dt>
      <dd className="text-sm">{children}</dd>
    </div>
  );
}

type Yrkande = Motion["yrkanden"][number];
type Decision = Yrkande["decisions"][number];

type Recurring = NonNullable<Yrkande["recurring"]>;

/** The history of a demand filed again with identical wording. Facts only:
 *  who filed it, in which riksmöte, and how it ended. The rule that grouped
 *  the filings is stated on each step. */
function RecurringBlock({ r }: { r: Recurring }) {
  return (
    <div className="mt-3 rounded border p-3 text-sm" style={{ borderColor: "var(--color-border)" }} data-testid="recurring">
      <p className="font-semibold">Återkommande krav</p>
      <p className="text-xs text-on-surface-variant mb-2">
        Samma ordalydelse (utan hänsyn till blanksteg, skiljetecken och versaler) har lämnats in {r.members.length}{" "}
        gånger.
      </p>
      <ol className="flex flex-col gap-1">
        {r.members.map((m) => (
          <li key={`${m.riksmote}:${m.beteckning}:${m.number}`} data-testid="recurring-step" aria-current={m.current ? "true" : undefined}>
            {m.current ? (
              <strong>
                {m.riksmote}:{m.beteckning} yrkande {m.number}
              </strong>
            ) : (
              <Link
                to={`/motioner/${encodeURIComponent(m.riksmote)}/${encodeURIComponent(m.beteckning)}`}
                className="text-link hover:underline"
              >
                {m.riksmote}:{m.beteckning} yrkande {m.number}
              </Link>
            )}
            {" · "}
            {m.parties.length > 0 ? m.parties.join(", ") : "parti saknas"} · {m.outcome} · {m.rule === "identical_wording" ? r.ruleLabel : m.rule}
            {m.current ? " (denna motion)" : ""}
          </li>
        ))}
      </ol>
      <SourceMarker sourceId="motions" />
    </div>
  );
}

/** The decision chain of one yrkande: where it was decided, how, and by whom.
 *  Facts only: the chamber's decision, the votering's Partilinje per party,
 *  with a link to Riksdagen's betänkande. */
function DecisionBlock({ d }: { d: Decision }) {
  return (
    <div className="mt-3 rounded border p-3 text-sm" style={{ borderColor: "var(--color-border)" }} data-testid="decision">
      <p>
        Avgjordes i{" "}
        <a href={d.betankandeSourceUrl} className="text-link hover:underline" rel="noreferrer" target="_blank">
          betänkande {d.betankandeRiksmote}:{d.betankandeBeteckning}
        </a>
        {d.betankandeTitle ? ` ${d.betankandeTitle}` : ""}, förslagspunkt {d.punkt}
        {d.rubrik ? ` "${d.rubrik}"` : ""}
        <SourceMarker sourceId="betankanden" />
      </p>
      <p className="mt-1" data-testid="decision-mode">
        <strong>{d.modeLabel}</strong>
        {d.alternativ ? <span> – {d.alternativ}</span> : null}
      </p>
      {d.mode === "votering" && d.voteringStored && d.parties.length > 0 ? (
        <table className="mt-2 text-xs" data-testid="party-table">
          <caption className="text-left text-on-surface-variant mb-1">Partilinje i omröstningen</caption>
          <thead>
            <tr className="text-left text-on-surface-variant">
              <th className="pr-4">Parti</th>
              <th className="pr-4">Partilinje</th>
              <th className="pr-3">Ja</th>
              <th className="pr-3">Nej</th>
              <th className="pr-3">Avstår</th>
              <th>Frånvarande</th>
            </tr>
          </thead>
          <tbody>
            {d.parties.map((p) => (
              <tr key={p.party}>
                <td className="pr-4"><PartyBadge party={p.party} /></td>
                <td className="pr-4">{p.partilinje || "Ingen"}</td>
                <td className="pr-3">{p.ja}</td>
                <td className="pr-3">{p.nej}</td>
                <td className="pr-3">{p.avstar}</td>
                <td>{p.franvarande}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : null}
      <VotedAlternative d={d} />
      <Reasoning d={d} />
      <Reservations d={d} />
      <Debate debate={d.debate} />
      {d.mode === "votering" && !d.voteringStored ? (
        <p className="mt-1 text-xs text-on-surface-variant">Omröstningen finns ännu inte inläst hos oss.</p>
      ) : null}
    </div>
  );
}

export function MotionPage() {
  const { riksmote = "", beteckning = "" } = useParams<{ riksmote: string; beteckning: string }>();
  const { data: motion, isLoading, error } = useQuery({
    queryKey: ["motion", riksmote, beteckning],
    queryFn: () => motionsApi.get(riksmote, beteckning),
    retry: false,
  });

  if (isLoading) return <Notice>Laddar...</Notice>;
  if (error instanceof ApiError && (error.status === 404 || error.status === 400)) {
    return <Notice>Motion {riksmote}:{beteckning} finns inte i vårt register.</Notice>;
  }
  if (error || !motion) return <Notice>Kunde inte ladda motionen.</Notice>;

  return (
    <article className="sdt-page px-8 py-8" style={{ maxWidth: 860 }}>
      <p className="font-mono text-xs uppercase tracking-wider text-on-surface-variant">
        Motion {motion.riksmote}:{motion.beteckning}
      </p>
      <h1 className="text-2xl font-bold my-2">
        {motion.title}
        <SourceMarker sourceId="motions" />
      </h1>

      <p className="text-sm mb-1" data-testid="motion-kind">
        <strong>{motion.kind}</strong>
        {motion.kindExplanation ? <span> – {motion.kindExplanation}</span> : null}
      </p>

      <p className="text-sm mb-1" data-testid="motion-outcome">
        Utfall: <strong>{motion.outcome.word}</strong>
        {Object.keys(motion.outcome.counts).length > 1
          ? ` (${Object.entries(motion.outcome.counts)
              .map(([w, n]) => `${n} ${w.toLowerCase()}`)
              .join(", ")} av ${motion.outcome.total} yrkanden)`
          : ""}
        <SourceMarker sourceId="betankanden" />
      </p>

      <dl className="grid gap-3 my-6" style={{ gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))" }}>
        <Fact label="Inlämnad">{motion.date ? swedishDate(motion.date) : "Datum saknas"}</Fact>
        <Fact label="Utskott">
          {motion.utskottName || motion.utskott || "Ej angivet"}
          {motion.utskott && motion.utskottName ? ` (${motion.utskott})` : ""}
        </Fact>
        <Fact label="Behandlas i">{motion.behandlasI || "Ej angivet"}</Fact>
      </dl>

      <section aria-labelledby="authors-heading" className="mb-8">
        <h2 id="authors-heading" className="text-lg font-semibold mb-2">
          Undertecknare
          <SourceMarker sourceId="motions" />
        </h2>
        <p className="text-xs text-on-surface-variant mb-2">
          Partiet är det som stod på motionen när den lämnades in.
        </p>
        <ul className="flex flex-col gap-1" aria-label="Undertecknare">
          {motion.authors.map((a) => (
            <li key={a.politicianId || a.name} className="flex items-center gap-2 text-sm">
              {a.politicianId ? (
                <Link to={`/politicians/${a.politicianId}`} className="text-link hover:underline">
                  {a.name}
                </Link>
              ) : (
                <span>{a.name}</span>
              )}
              {a.party && a.party !== "-" ? <PartyBadge party={a.party} /> : null}
            </li>
          ))}
        </ul>
      </section>

      <section aria-labelledby="yrkanden-heading" className="mb-8">
        <h2 id="yrkanden-heading" className="text-lg font-semibold mb-2">
          Yrkanden
          <SourceMarker sourceId="motions" />
        </h2>
        {motion.yrkanden.length === 0 ? (
          <p className="text-sm text-on-surface-variant">Riksdagen anger inga yrkanden för motionen.</p>
        ) : (
          <ol className="flex flex-col gap-4" aria-label="Yrkanden">
            {motion.yrkanden.map((y) => (
              <li
                key={y.number}
                className="rounded border p-4"
                style={{ borderColor: "var(--color-border)" }}
                data-testid="yrkande"
              >
                <p className="font-mono text-xs text-on-surface-variant mb-1">Yrkande {y.number}</p>
                <p className="text-sm">{y.wording}</p>
                <p className="text-sm mt-2" data-testid="yrkande-outcome">
                  Utfall: <strong>{y.outcome}</strong>
                </p>
                {y.recurring ? <RecurringBlock r={y.recurring} /> : null}
                {y.decisions.map((d) => (
                  <DecisionBlock key={`${d.betankandeRiksmote}:${d.betankandeBeteckning}:${d.punkt}`} d={d} />
                ))}
                {y.decisions.length === 0 ? (
                  <p className="text-xs text-on-surface-variant mt-1">
                    Vi har inte hittat någon förslagspunkt som avgjort yrkandet.
                  </p>
                ) : null}
                <p className="text-xs text-on-surface-variant mt-2">
                  Utskottets förslag: {y.utskottProposal || "ej angivet"} · Kammarens beslut:{" "}
                  {y.chamberDecision || "ej angivet"}
                  {y.behandlasI ? ` · Behandlas i ${y.behandlasI}` : ""}
                </p>
              </li>
            ))}
          </ol>
        )}
      </section>

      <p className="text-sm">
        <a href={motion.sourceUrl} className="text-link hover:underline" rel="noreferrer" target="_blank">
          Läs motionen hos Riksdagen ↗
        </a>
      </p>
    </article>
  );
}
