import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { loftenApi } from "./api";
import { ArchiveNote } from "./components";
import { MANDATE_LABEL } from "./labels";
import { SourceMarker } from "@/components/sources/SourceMarker";

/** One Löfte. Archived 2022 goals open here under their old id. */
export function LoefteDetailPage() {
  const { id = "" } = useParams<{ id: string }>();
  const { data: l, isLoading, isError } = useQuery({
    queryKey: ["loefte", id],
    queryFn: () => loftenApi.get(Number(id)),
    retry: false,
  });

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
          {l.quote && <blockquote className="border-l-2 pl-3 my-4 text-sm">{l.quote}</blockquote>}
          <p className="text-sm">
            Källa: {l.sourceDocument || "Valmanifest"}
            {l.locator ? `, ${l.locator}` : ""}{" "}
            {l.sourceUrl && <a href={l.sourceUrl} className="underline" target="_blank" rel="noopener noreferrer">Öppna dokumentet</a>}{" "}
            <SourceMarker sourceId="seed-party-goals" />
          </p>
        </>
      )}
    </div>
  );
}
