import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";
import { Link, useLocation, useParams } from "react-router-dom";
import { loftenApi } from "./api";
import { ArchiveNote, StatusBadge } from "./components";
import { MANDATE_LABEL } from "./labels";
import { betankandePath } from "@/features/votes/betankandeApi";
import { SourceMarker } from "@/components/sources/SourceMarker";

/** One Löfte. Archived 2022 goals open here under their old id. */
export function LoefteDetailPage() {
  const { id = "" } = useParams<{ id: string }>();
  const { data: l, isLoading, isError } = useQuery({
    queryKey: ["loefte", id],
    queryFn: () => loftenApi.get(Number(id)),
    retry: false,
  });

  const { hash } = useLocation();
  useEffect(() => {
    if (l && hash) document.getElementById(hash.slice(1))?.scrollIntoView();
  }, [l, hash]);

  const manifests = useQuery({
    queryKey: ["valmanifest", l?.party],
    queryFn: () => loftenApi.valmanifest(l!.party),
    enabled: !!l?.valmanifestId,
    staleTime: 60_000,
  });
  const manifest = manifests.data?.find((v) => v.id === l?.valmanifestId);

  return (
    <div className="sdt-page" style={{ padding: "24px 32px 64px", maxWidth: 760 }}>
      <Link to="/loften" className="text-xs text-link">← Löften</Link>
      {isLoading && <p className="text-sm mt-4">Laddar...</p>}
      {isError && <p className="text-sm mt-4">Löftet finns inte.</p>}
      {l && (
        <>
          <h1 className="font-display text-2xl mt-2 mb-3" data-testid="loefte-text">{l.text}</h1>
          <p className="text-xs text-on-surface-variant mb-4">
            <Link to={`/parties/${l.party}?tab=loften`} className="underline">{l.party}</Link>
            {" · "}{l.archived ? "Arkiv " : ""}{MANDATE_LABEL[l.mandatePeriod] ?? l.mandatePeriod}
            {l.utskott.length > 0 ? ` · ${l.utskott.join(", ")}` : ""}
          </p>
          {l.archived && <ArchiveNote />}
          <section className="my-4" data-testid="loefte-status-section">
            <h2 className="font-display text-lg mb-1">Löftesstatus</h2>
            <p><StatusBadge status={l.status} /> <Link to="/metod" className="text-xs text-link underline">Så räknar vi</Link></p>
            {l.status && (
              <p className="text-xs text-on-surface-variant mt-1" data-testid="status-changed">
                {l.statusChangedAt
                  ? `Senast ändrad ${l.statusChangedAt.slice(0, 10)}. ${l.statusReason ?? ""}`
                  : "Statusen har inte ändrats sedan Löftet publicerades."}
              </p>
            )}
            {l.status === "Inte prövat än" && (
              <p className="text-xs text-on-surface-variant mt-1">
                Inte prövat än betyder att inget parti-ställningstagande i en bekräftad omröstning går att räkna
                för eller emot. Omröstningar där partiet avstod eller var delat visas nedan men räknas inte.
              </p>
            )}
          </section>
          {l.quote && <blockquote className="border-l-2 pl-3 my-4 text-sm" data-testid="loefte-quote">{l.quote}</blockquote>}
          <p className="text-sm" data-testid="loefte-source">
            Källa:{" "}
            {manifest ? (
              <span data-testid="loefte-valmanifest">{manifest.title} ({manifest.electionYear})</span>
            ) : (
              l.sourceDocument || "Valmanifest"
            )}
            {l.locator ? `, ${l.locator}` : ""}{" "}
            {l.sourceUrl && <a href={l.sourceUrl} className="underline" target="_blank" rel="noopener noreferrer" data-testid="loefte-source-link">Öppna originalet</a>}{" "}
            <SourceMarker sourceId="seed-party-goals" />
          </p>
          {(l.kopplingar ?? []).length > 0 && (
            <section className="mt-8" data-testid="kopplingar">
              <h2 className="font-display text-lg mb-3">Omröstningarna bakom statusen</h2>
              <ul className="space-y-4">
                {l.kopplingar!.map((k) => (
                  <li key={k.id} id={`koppling-${k.id}`} data-testid="koppling" className="bg-surface-lowest rounded-xl p-4 text-sm">
                    <div className="font-semibold">
                      <Link to={betankandePath(k.session, k.beteckning, k.forslagspunkt)} className="underline">
                        {k.betankande || `${k.beteckning} punkt ${k.forslagspunkt}`}
                      </Link>
                    </div>
                    <div className="text-xs text-on-surface-variant">
                      {k.session}{k.voteDate ? ` · ${k.voteDate}` : ""} · {l.party}: Ja {k.ja}, Nej {k.nej}, Avstår {k.avstar}, Frånvarande {k.franvarande}
                      {" · "}Partilinje: {k.partilinje || "saknas"} · Bekräftad {k.confirmedAt.slice(0, 10)}{" "}
                      <Link to={`/kopplingar/${k.id}`} className="underline">Länk</Link>{" "}
                      <SourceMarker sourceId="voteringar" />
                    </div>
                    <p className="mt-2" data-testid="koppling-explanation">{k.explanation}</p>
                  </li>
                ))}
              </ul>
            </section>
          )}
        </>
      )}
    </div>
  );
}
