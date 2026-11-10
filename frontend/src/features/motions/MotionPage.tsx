import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { motionsApi } from "./api";
import { ApiError } from "@/shared/api-client";
import { PartyBadge } from "@/shared/components";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { swedishDate } from "@/shared/dates";

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
