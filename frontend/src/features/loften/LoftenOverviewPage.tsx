import { useQuery } from "@tanstack/react-query";
import { Link, useSearchParams } from "react-router-dom";
import { loftenApi } from "./api";
import { ArchiveNote, LoefteCard, MandatperiodSelect, ValmanifestList } from "./components";
import { loftenHitCount } from "./hitCount";
import { MANDATE_LABEL, STATUS_FILTERS, mandatperiodFromQuery } from "./labels";
import { betankandePath } from "@/features/votes/betankandeApi";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { TemporaryError } from "@/shared/components";

const PARTIES = ["C", "KD", "L", "M", "MP", "S", "SD", "V"];
const EMPTY_SUMMARY = { total: 0, for: 0, emot: 0, blandat: 0, inteProvat: 0, arkiv: 0 };
const UTSKOTT = ["AU", "CU", "FiU", "FöU", "JuU", "KU", "KrU", "MJU", "NU", "SfU", "SkU", "SoU", "TU", "UbU", "UU"];

/** Löften across parties, with Löftesstatus. Reached from Partier, not from the top menu. */
export function LoftenOverviewPage() {
  const [sp, setSp] = useSearchParams();
  const party = sp.get("party") ?? "";
  const utskott = sp.get("utskott") ?? "";
  const status = sp.get("status") ?? "";
  const betankande = sp.get("betankande") ?? "";
  const mandatperiod = mandatperiodFromQuery(sp.get("mandatperiod"));
  const set = (k: string, v: string) => {
    const n = new URLSearchParams(sp);
    if (v) n.set(k, v);
    else n.delete(k);
    setSp(n, { replace: true });
  };

  const loften = useQuery({
    queryKey: ["loften", party, utskott, status, betankande, mandatperiod],
    queryFn: () => loftenApi.list({ party, utskott, status, betankande, mandatperiod }),
    staleTime: 60_000,
  });
  const summary = useQuery({
    queryKey: ["loften-summary", mandatperiod],
    queryFn: () => loftenApi.summary(mandatperiod),
    staleTime: 60_000,
  });
  const manifest = useQuery({ queryKey: ["valmanifest"], queryFn: () => loftenApi.valmanifest(), staleTime: 60_000 });

  return (
    <div className="sdt-page" style={{ padding: "24px 32px 64px", maxWidth: 920 }}>
      <Link to="/parties" className="text-xs text-link">← Partier</Link>
      <h1 className="font-display text-3xl mt-2 mb-2">Löften</h1>
      <p className="text-sm text-on-surface-variant mb-2">
        Löften är konkreta uttalanden i partiernas valmanifest. Löftesstatus
        visar bara hur partiet röstade i de omröstningar redaktören kopplat till
        Löftet, aldrig om löftet infriats. Vi låter dig dra slutsatserna.{" "}
        <Link to="/metod" className="text-link underline">Så räknar vi</Link>
      </p>

      <section data-testid="loften-summary" className="my-4 overflow-x-auto">
        <table className="text-xs w-full">
          <caption data-testid="loften-summary-period" className="text-left text-on-surface-variant mb-1">
            {mandatperiod ? `Mandatperiod ${MANDATE_LABEL[mandatperiod] ?? mandatperiod}` : "Alla mandatperioder"}
          </caption>
          <thead>
            <tr className="text-left text-on-surface-variant">
              <th>Parti</th><th>Röstat för</th><th>Röstat emot</th><th>Blandat</th><th>Inte prövat än</th><th>Arkiv</th>
            </tr>
          </thead>
          <tbody>
            {summary.data && PARTIES.map((code) => {
              // A party the API leaves out has no public Löften at all.
              const r = summary.data.find((x) => x.party === code) ?? EMPTY_SUMMARY;
              // Archived Löften carry no Löftesstatus, so only the rest count here.
              const published = r.total - r.arkiv;
              return (
                <tr key={code} data-testid={`summary-${code}`}>
                  <td><Link to={`/loften?party=${code}${mandatperiod ? `&mandatperiod=${mandatperiod}` : ""}`} className="underline">{code}</Link></td>
                  {published === 0 ? (
                    // No data is not the same as zero Löften with a status.
                    <td colSpan={4} className="text-on-surface-variant" data-testid="summary-no-loften">Inga löften publicerade ännu</td>
                  ) : (
                    [r.for, r.emot, r.blandat, r.inteProvat].map((n, i) => (
                      <td key={i} data-testid="summary-status-count">{n}</td>
                    ))
                  )}
                  <td data-testid="summary-arkiv">{r.arkiv}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
        <SourceMarker sourceId="loften-valmanifest" />
      </section>

      <div className="flex gap-3 flex-wrap text-sm mb-4" data-testid="loften-filters">
        <label>Parti{" "}
          <select value={party} onChange={(e) => set("party", e.target.value)}>
            <option value="">Alla</option>
            {PARTIES.map((p) => <option key={p} value={p}>{p}</option>)}
          </select>
        </label>
        <label>Utskott{" "}
          <select value={utskott} onChange={(e) => set("utskott", e.target.value)}>
            <option value="">Alla</option>
            {UTSKOTT.map((u) => <option key={u} value={u}>{u}</option>)}
          </select>
        </label>
        <MandatperiodSelect value={mandatperiod} onChange={(v) => set("mandatperiod", v)} />
        <label>Löftesstatus{" "}
          <select value={status} onChange={(e) => set("status", e.target.value)}>
            <option value="">Alla</option>
            {STATUS_FILTERS.map((s) => <option key={s.slug} value={s.slug}>{s.label}</option>)}
          </select>
        </label>
      </div>

      {betankande && (
        <p data-testid="betankande-filter" className="text-sm mb-3">
          Löften som berörs av betänkande{" "}
          <Link to={betankandePath(betankande.split(":")[0], betankande.split(":")[1] ?? "")} className="text-link underline">{betankande}</Link>.{" "}
          <button type="button" className="text-link underline" onClick={() => set("betankande", "")}>Visa alla</button>
        </p>
      )}

      {loften.data && (
        <p data-testid="loften-hit-count" aria-live="polite" className="text-sm font-medium mb-2">
          {loftenHitCount(loften.data.length, party, mandatperiod)}
        </p>
      )}

      <ArchiveNote />

      {loften.isLoading && <p className="text-sm">Laddar...</p>}
      {loften.isError && <TemporaryError what="Löftena" onRetry={() => void loften.refetch()} />}
      {loften.data?.length === 0 && <p className="text-sm text-on-surface-variant">Inga löften matchar filtret.</p>}
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
