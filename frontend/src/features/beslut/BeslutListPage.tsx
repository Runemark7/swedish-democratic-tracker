import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link, useSearchParams } from "react-router-dom";
import { beslutApi, type BeslutFilters, type BeslutListRow } from "./api";
import { betankandePath } from "@/features/votes/betankandeApi";
import { PartyBadge } from "@/shared/components";
import { PARTIES } from "@/shared/design";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { swedishDate } from "@/shared/dates";

const UTSKOTT: [string, string][] = [
  ["AU", "Arbetsmarknadsutskottet"],
  ["CU", "Civilutskottet"],
  ["FiU", "Finansutskottet"],
  ["FöU", "Försvarsutskottet"],
  ["JuU", "Justitieutskottet"],
  ["KrU", "Kulturutskottet"],
  ["KU", "Konstitutionsutskottet"],
  ["MJU", "Miljö- och jordbruksutskottet"],
  ["NU", "Näringsutskottet"],
  ["SfU", "Socialförsäkringsutskottet"],
  ["SkU", "Skatteutskottet"],
  ["SoU", "Socialutskottet"],
  ["TU", "Trafikutskottet"],
  ["UbU", "Utbildningsutskottet"],
  ["UU", "Utrikesutskottet"],
];
const STATES: [string, string][] = [
  ["decided", "Fattade beslut"],
  ["upcoming", "Kommande beslut"],
];

function Select({
  label, value, onChange, options,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  options: [string, string][];
}) {
  return (
    <label className="flex flex-col gap-1 text-xs uppercase tracking-wider text-on-surface-variant">
      {label}
      <select
        aria-label={label}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="text-sm normal-case tracking-normal text-on-surface rounded border px-2 py-1.5"
        style={{ borderColor: "var(--color-border)", background: "var(--color-surface-lowest)" }}
      >
        <option value="">Alla</option>
        {options.map(([v, l]) => (
          <option key={v} value={v}>{l}</option>
        ))}
      </select>
    </label>
  );
}

/** Each party's Partilinje in the betänkande, a link to the party page. Fact only. */
function PartyMarks({ row }: { row: BeslutListRow }) {
  if (row.parties.length === 0) return null;
  return (
    <ul className="flex flex-wrap gap-x-3 gap-y-1.5" aria-label="Partilinje per parti">
      {PARTIES.map((p) => {
        const line = row.parties.find((x) => x.party === p);
        return (
          <li key={p}>
            <Link
              to={`/parties/${p}`}
              data-testid="party-mark"
              data-party={p}
              className="inline-flex items-center gap-1.5 text-xs text-on-surface-variant hover:text-link"
              title={`Partilinje för ${p}`}
            >
              <PartyBadge party={p} />
              <span className="font-mono">{line?.partilinje || "–"}</span>
            </Link>
          </li>
        );
      })}
    </ul>
  );
}

function Row({ r }: { r: BeslutListRow }) {
  return (
    <li
      data-testid={r.decided ? "beslut-row" : "beslut-row-upcoming"}
      className={`flex flex-col gap-2 py-3.5 px-4 rounded-lg ${r.decided ? "" : "border border-dashed"}`}
      style={{
        background: r.decided ? "var(--color-surface-lowest)" : "transparent",
        borderColor: "var(--color-border)",
      }}
    >
      <Link to={betankandePath(r.riksmote, r.beteckning)} className="group">
        <p className="text-sm font-semibold text-on-surface group-hover:text-link transition-colors">
          {r.title || `${r.riksmote}:${r.beteckning}`}
          <SourceMarker sourceId="betankanden" />
        </p>
      </Link>
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-on-surface-variant">
        <span className="font-mono">{r.riksmote}:{r.beteckning}</span>
        <span>{r.utskottName}</span>
        <span data-testid="row-date">
          {r.decided ? "Beslutad " : "Planerad beslutsdag "}
          {r.date ? swedishDate(r.date) : "saknas"}
        </span>
        <strong className="text-on-surface" data-testid="row-outcome">{r.outcome}</strong>
      </div>
      <PartyMarks row={r} />
    </li>
  );
}

export function BeslutListPage() {
  const [params, setParams] = useSearchParams();
  const filters: BeslutFilters = {
    state: params.get("state") ?? "",
    utskott: params.get("utskott") ?? "",
    party: params.get("party") ?? "",
  };
  const filterKey = JSON.stringify(filters);
  // Cursors of the pages after the first, valid only for the filters they came from.
  const [paging, setPaging] = useState<{ key: string; cursors: string[] }>({ key: "", cursors: [] });
  const cursors = paging.key === filterKey ? paging.cursors : [];
  const setCursors = (f: (c: string[]) => string[]) => setPaging({ key: filterKey, cursors: f(cursors) });

  function set(key: keyof BeslutFilters, value: string) {
    const next = new URLSearchParams(params);
    if (value) next.set(key, value);
    else next.delete(key);
    setParams(next, { replace: true });
  }

  const page = cursors.length;
  const { data, isLoading, error, isFetching } = useQuery({
    queryKey: ["beslut", filterKey, cursors],
    queryFn: () => beslutApi.list(filters, cursors[cursors.length - 1]),
    placeholderData: (prev) => prev,
  });
  const rows = data?.data ?? [];
  const upcoming = rows.filter((r) => !r.decided);
  const decided = rows.filter((r) => r.decided);

  return (
    <div>
      <div className="mb-6">
        <h1 className="font-display text-2xl font-extrabold tracking-tight mb-1">Beslut</h1>
        <p className="text-sm text-on-surface-variant leading-relaxed max-w-lg">
          Riksdagens betänkanden. Kommande beslut först, därefter fattade beslut, senaste först.
          Varje parti visar hur flertalet av dess ledamöter röstade.
        </p>
      </div>

      <div className="grid gap-3 mb-6" style={{ gridTemplateColumns: "repeat(auto-fit, minmax(150px, 1fr))" }} data-testid="beslut-filters">
        <Select label="Status" value={filters.state ?? ""} onChange={(v) => set("state", v)} options={STATES} />
        <Select label="Utskott" value={filters.utskott ?? ""} onChange={(v) => set("utskott", v)} options={UTSKOTT.map(([c, n]) => [c, `${n} (${c})`])} />
        <Select label="Parti" value={filters.party ?? ""} onChange={(v) => set("party", v)} options={PARTIES.map((p) => [p, p])} />
      </div>
      {filters.party && (
        <p className="text-xs text-on-surface-variant mb-4">
          Parti-filtret visar betänkanden som avgör en motion från partiet eller där partiet har en reservation.
        </p>
      )}

      {isLoading ? (
        <p className="text-sm text-on-surface-variant py-12 text-center">Laddar beslut...</p>
      ) : error ? (
        <p className="text-sm text-contradiction py-12 text-center">Kunde inte ladda beslut.</p>
      ) : (
        <>
          <div aria-busy={isFetching} className="space-y-6">
            {upcoming.length > 0 && (
              <section aria-label="Kommande beslut">
                <h2 className="text-xs uppercase tracking-wider text-on-surface-variant mb-2">Kommande beslut</h2>
                <ul className="space-y-2">{upcoming.map((r) => <Row key={`${r.riksmote}:${r.beteckning}`} r={r} />)}</ul>
              </section>
            )}
            {decided.length > 0 && (
              <section aria-label="Fattade beslut">
                <h2 className="text-xs uppercase tracking-wider text-on-surface-variant mb-2">Fattade beslut</h2>
                <ul className="space-y-2">{decided.map((r) => <Row key={`${r.riksmote}:${r.beteckning}`} r={r} />)}</ul>
              </section>
            )}
          </div>
          {rows.length === 0 && (
            <p className="text-sm text-on-surface-variant text-center py-12">Inga beslut matchar filtren.</p>
          )}
          <div className="flex items-center justify-center gap-4 mt-6">
            <button
              onClick={() => setCursors((c) => c.slice(0, -1))}
              disabled={page === 0}
              className="text-xs font-semibold text-link disabled:text-on-surface-variant disabled:cursor-not-allowed cursor-pointer"
            >
              ← Föregående
            </button>
            <span className="text-xs text-on-surface-variant font-mono">Sida {page + 1}</span>
            <button
              onClick={() => data?.nextCursor && setCursors((c) => [...c, data.nextCursor as string])}
              disabled={!data?.nextCursor}
              className="text-xs font-semibold text-link disabled:text-on-surface-variant disabled:cursor-not-allowed cursor-pointer"
            >
              Nästa →
            </button>
          </div>
        </>
      )}
    </div>
  );
}
