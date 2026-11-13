import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link, Navigate, useLocation, useParams } from "react-router-dom";
import { betankandeApi, betankandePath, voteSharePath, type BetankandePage as Page, type BetankandePunkt } from "./betankandeApi";
import { PartyBadge } from "@/shared/components";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { swedishDate } from "@/shared/dates";
import { motionPath } from "@/features/motions/api";
import { Reasoning, Reservations, VotedAlternative } from "@/features/motions/DecisionParts";
import { LoefteRefList } from "@/features/loften/OnDecision";
import { Debate } from "@/features/motions/Debate";

function Notice({ children }: { children: React.ReactNode }) {
  return (
    <div className="sdt-page px-8 py-16">
      <p className="text-sm text-on-surface-variant">{children}</p>
    </div>
  );
}

/** Decision state as Riksdagen states it. A planned betänkande says so and
 *  shows the planned date; the date is never our own record timestamp. */
function DecisionStatus({ p }: { p: Page }) {
  if (p.decided === null || p.decided === undefined) {
    return <p className="text-sm text-on-surface-variant" data-testid="decision-status">Beslutsstatus kunde inte hämtas från Riksdagen just nu.</p>;
  }
  if (!p.decided) {
    return (
      <p className="text-sm" data-testid="decision-status">
        <strong>{p.decisionText || "Ej avgjort än"}</strong>
        {p.decisionDate ? <span> · planerad beslutsdag {swedishDate(p.decisionDate)}</span> : null}
        <SourceMarker sourceId="riksdagen" />
      </p>
    );
  }
  return (
    <p className="text-sm" data-testid="decision-status">
      <strong>Avgjort</strong>
      {p.decisionDate ? <span> {swedishDate(p.decisionDate)}</span> : null}
      <SourceMarker sourceId="riksdagen" />
    </p>
  );
}

function PunktBlock({ page, k }: { page: Page; k: BetankandePunkt }) {
  const parts = { ...k, betankandeSourceUrl: page.sourceUrl };
  return (
    <section id={`punkt-${k.number}`} className="rounded border p-4" style={{ borderColor: "var(--color-border)" }} data-testid="punkt">
      <h2 className="text-base font-medium">
        Förslagspunkt {k.number}{k.rubrik ? ` – ${k.rubrik}` : ""}
      </h2>
      <p className="mt-1 text-sm" data-testid="decision-mode">
        <strong>{page.decided === false ? "Ej avgjort än" : k.modeLabel}</strong>
        {k.alternativ ? <span> – {k.alternativ}</span> : null}
      </p>
      {k.text ? <p className="mt-2 text-sm whitespace-pre-line">{k.text}</p> : null}

      {k.mode === "votering" && k.voteringStored && k.parties.length > 0 ? (
        <table className="mt-3 text-xs" data-testid="party-table">
          <caption className="text-left text-on-surface-variant mb-1">Partilinje i omröstningen</caption>
          <thead>
            <tr className="text-left text-on-surface-variant">
              <th className="pr-4">Parti</th><th className="pr-4">Partilinje</th>
              <th className="pr-3">Ja</th><th className="pr-3">Nej</th><th className="pr-3">Avstår</th><th>Frånvarande</th>
            </tr>
          </thead>
          <tbody>
            {k.parties.map((p) => (
              <tr key={p.party}>
                <td className="pr-4"><PartyBadge party={p.party} /></td>
                <td className="pr-4">{p.partilinje || "Ingen"}</td>
                <td className="pr-3">{p.ja}</td><td className="pr-3">{p.nej}</td><td className="pr-3">{p.avstar}</td><td>{p.franvarande}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : null}
      {k.mode === "votering" && !k.voteringStored ? (
        <p className="mt-1 text-xs text-on-surface-variant">Omröstningen finns ännu inte inläst hos oss.</p>
      ) : null}

      {k.voteringStored ? (
        <div className="mt-3" data-testid="avvikare">
          <h3 className="text-xs uppercase tracking-wider text-on-surface-variant">Avvikare</h3>
          {k.avvikare.length === 0 ? (
            <p className="text-xs text-on-surface-variant">Ingen ledamot röstade mot sin Partilinje.</p>
          ) : (
            <ul className="mt-1 text-sm">
              {k.avvikare.map((a) => (
                <li key={a.politicianId}>
                  <Link to={`/politicians/${encodeURIComponent(a.politicianId)}`} className="text-link hover:underline">{a.name}</Link>{" "}
                  <PartyBadge party={a.party} />{" "}
                  <span className="text-xs text-on-surface-variant">röstade {a.voteResult} (Partilinje {a.partilinje})</span>
                </li>
              ))}
            </ul>
          )}
        </div>
      ) : null}

      {page.held ? (
        <>
          <VotedAlternative d={parts} />
          <Reasoning d={parts} />
          <Reservations d={parts} />
        </>
      ) : (
        <p className="mt-2 text-xs text-on-surface-variant" data-testid="not-available">
          Förslagstext, reservationer och utskottets ställningstagande är inte inlästa för det här betänkandet.
          {page.sourceUrl ? <> <a href={page.sourceUrl} className="text-link hover:underline" rel="noreferrer" target="_blank">Läs i betänkandet ↗</a></> : null}
        </p>
      )}

      <div className="mt-3" data-testid="punkt-motions">
        <h3 className="text-xs uppercase tracking-wider text-on-surface-variant">
          Motioner som avgjordes här<SourceMarker sourceId="betankanden" />
        </h3>
        {k.motions.length === 0 ? (
          <p className="text-xs text-on-surface-variant">Inga motioner är kopplade till punkten.</p>
        ) : (
          <ul className="mt-1 text-sm">
            {k.motions.map((m) => (
              <li key={m.dokId}>
                <Link to={motionPath(m.riksmote, m.beteckning)} className="text-link hover:underline">
                  {m.riksmote}:{m.beteckning}
                </Link>{" "}
                {m.title}
                {m.parties.length > 0 ? ` (${m.parties.join(", ")})` : ""}
                <span className="text-xs text-on-surface-variant"> · yrkande {m.yrkanden.join(", ")}</span>
              </li>
            ))}
          </ul>
        )}
      </div>
      <p className="mt-2 text-xs">
        {/* A plain anchor: the share page must be a fresh load (see App.tsx). */}
        <a href={voteSharePath(page.riksmote, page.beteckning, k.number)} className="text-link hover:underline">
          Dela omröstningen
        </a>
      </p>
    </section>
  );
}

export function BetankandePage() {
  const params = useParams<{ y?: string; yy?: string; riksmote?: string; beteckning: string }>();
  // /beslut/2025/26/SkU16 is canonical; /beslut/2025%2F26/SkU16 is accepted too.
  const riksmote = params.riksmote ?? `${params.y}/${params.yy}`;
  const beteckning = params.beteckning ?? "";
  const { hash } = useLocation();
  const { data, isLoading, isError } = useQuery({
    queryKey: ["betankande", riksmote, beteckning],
    queryFn: () => betankandeApi.get(riksmote, beteckning),
    retry: false,
  });

  useEffect(() => {
    if (data && hash) document.getElementById(hash.slice(1))?.scrollIntoView();
  }, [data, hash]);

  if (isLoading) return <Notice>Hämtar betänkandet…</Notice>;
  if (isError || !data) return <Notice>Betänkande {riksmote}:{beteckning} finns inte i vårt register.</Notice>;
  // One address per betänkande: /beslut/2025/26/sfu18 moves to …/SfU18.
  if (data.beteckning !== beteckning) return <Navigate to={betankandePath(data.riksmote, data.beteckning) + hash} replace />;

  return (
    <div className="sdt-page px-8 py-10" style={{ maxWidth: 880 }}>
      <p className="text-xs uppercase tracking-wider text-on-surface-variant">Betänkande {data.riksmote}:{data.beteckning}</p>
      <h1 className="text-3xl mt-1" style={{ fontFamily: "var(--font-serif)" }}>{data.title || `${data.riksmote}:${data.beteckning}`}</h1>
      <div className="mt-2"><DecisionStatus p={data} /></div>
      {data.sourceUrl ? (
        <p className="mt-1 text-xs">
          <a href={data.sourceUrl} className="text-link hover:underline" rel="noreferrer" target="_blank">Läs betänkandet hos Riksdagen ↗</a>
        </p>
      ) : null}

      <div className="mt-6 flex flex-col gap-4">
        {data.punkter.length === 0 ? (
          <p className="text-sm text-on-surface-variant" data-testid="no-punkter">
            {data.decided === false
              ? "Förslagspunkterna är inte inlästa än; betänkandet är inte avgjort."
              : "Inga förslagspunkter är inlästa för det här betänkandet."}
          </p>
        ) : (
          data.punkter.map((k) => <PunktBlock key={`${k.number}-${k.voteringId}`} page={data} k={k} />)
        )}
      </div>

      <LoefteRefList items={data.loften} heading={`Berör ${data.loefteCount} ${data.loefteCount === 1 ? "löfte" : "löften"}`} />
      <div className="mt-8"><Debate debate={data.debate} /></div>
    </div>
  );
}
