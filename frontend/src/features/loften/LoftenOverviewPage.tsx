import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { loftenApi } from "./api";
import { ArchiveNote, LoefteCard, ValmanifestList } from "./components";

const PARTIES = ["C", "KD", "L", "M", "MP", "S", "SD", "V"];

/** Löften across parties. Reached from Partier, not from the top menu. */
export function LoftenOverviewPage() {
  const loften = useQuery({ queryKey: ["loften"], queryFn: () => loftenApi.list(), staleTime: 60_000 });
  const manifest = useQuery({ queryKey: ["valmanifest"], queryFn: () => loftenApi.valmanifest(), staleTime: 60_000 });

  return (
    <div className="sdt-page" style={{ padding: "24px 32px 64px", maxWidth: 920 }}>
      <Link to="/parties" className="text-xs text-link">← Partier</Link>
      <h1 className="font-display text-3xl mt-2 mb-2">Löften</h1>
      <p className="text-sm text-on-surface-variant mb-6">
        Löften är konkreta uttalanden i partiernas valmanifest. Vi visar vad som
        står och var, och låter dig dra slutsatserna.
      </p>

      <ArchiveNote />

      {loften.isLoading && <p className="text-sm">Laddar...</p>}
      {loften.isError && <p className="text-sm text-on-surface-variant">Kunde inte ladda löften. Försök igen.</p>}
      {PARTIES.map((p) => {
        const rows = (loften.data ?? []).filter((l) => l.party === p);
        if (rows.length === 0) return null;
        return (
          <section key={p} className="mt-8" data-testid={`loften-party-${p}`}>
            <h2 className="font-display text-lg mb-3">
              <Link to={`/parties/${p}?tab=loften`} className="hover:underline">{p}</Link>
            </h2>
            <div className="space-y-3">
              {rows.map((l) => <LoefteCard key={l.id} l={l} />)}
            </div>
          </section>
        );
      })}

      <section id="valmanifest" className="mt-12">
        <h2 className="font-display text-lg mb-3">Valmanifest (källor)</h2>
        {manifest.data && <ValmanifestList items={manifest.data} />}
      </section>
    </div>
  );
}
