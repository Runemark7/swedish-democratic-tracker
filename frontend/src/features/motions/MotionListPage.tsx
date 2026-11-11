import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link, useSearchParams } from "react-router-dom";
import { motionPath, motionsApi, type MotionFilters, type MotionListRow } from "./api";
import { PartyBadge } from "@/shared/components";
import { PARTIES } from "@/shared/design";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { swedishDate } from "@/shared/dates";
import { useMyParty, MY_PARTY_RING } from "@/contexts/myParty";
import { MyPartyPicker } from "@/components/MyPartyPicker";

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
const OUTCOMES = ["Avslagen", "Bifallen", "Delvis bifallen", "Besvarad", "Ej avgjord än", "Blandat utfall"];
const KINDS = ["Kommittémotion", "Partimotion", "Enskild motion", "Flerpartimotion"];

/** Riksmöten from the one that is open now back to 2002/03, where the voteringar start. */
function riksmoten(now = new Date()): string[] {
  const start = now.getMonth() >= 8 ? now.getFullYear() : now.getFullYear() - 1;
  const out: string[] = [];
  for (let y = start; y >= 2002; y--) out.push(`${y}/${String((y + 1) % 100).padStart(2, "0")}`);
  return out;
}

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

/** "Delvis bifallen" with the counts behind it, e.g. "Delvis bifallen (2 av 3 yrkanden bifallna)". */
function outcomeLabel(o: MotionListRow["outcome"]): string {
  const kinds = Object.keys(o.counts);
  if (kinds.length <= 1) return o.word;
  const parts = Object.entries(o.counts).map(([w, n]) => `${n} ${w.toLowerCase()}`);
  return `${o.word} (${parts.join(", ")} av ${o.total} yrkanden)`;
}

function Row({ m }: { m: MotionListRow }) {
  const { myParty } = useMyParty();
  const mine = myParty !== null && m.parties.includes(myParty);
  return (
    <li data-my-party={mine ? "true" : undefined}>
      <Link
        to={motionPath(m.riksmote, m.beteckning)}
        data-testid="motion-row"
        className="flex flex-col gap-1.5 py-3.5 px-4 rounded-lg transition-all hover:shadow-ambient group"
        style={{ background: "var(--color-surface-lowest)", boxShadow: mine ? MY_PARTY_RING : undefined }}
      >
        <p className="text-sm font-semibold text-on-surface group-hover:text-link transition-colors">
          {m.title || `Motion ${m.riksmote}:${m.beteckning}`}
          <SourceMarker sourceId="motions" />
        </p>
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-on-surface-variant">
          <span className="font-mono">{m.riksmote}:{m.beteckning}</span>
          <span className="flex gap-1">{m.parties.map((p) => <PartyBadge key={p} party={p} />)}</span>
          <span data-testid="row-kind">{m.kind}</span>
          <span>{m.date ? swedishDate(m.date) : "Datum saknas"}</span>
          <strong className="text-on-surface" data-testid="row-outcome">{outcomeLabel(m.outcome)}</strong>
          <SourceMarker sourceId="betankanden" />
        </div>
      </Link>
    </li>
  );
}

export function MotionListPage() {
  const [params, setParams] = useSearchParams();
  const filters: MotionFilters = {
    q: params.get("q") ?? "",
    party: params.get("party") ?? "",
    utskott: params.get("utskott") ?? "",
    outcome: params.get("outcome") ?? "",
    riksmote: params.get("riksmote") ?? "",
    kind: params.get("kind") ?? "",
  };
  const [text, setText] = useState(filters.q ?? "");
  const filterKey = JSON.stringify(filters);
  // Cursors of the pages after the first, valid only for the filters they came from.
  const [paging, setPaging] = useState<{ key: string; cursors: string[] }>({ key: "", cursors: [] });
  const cursors = paging.key === filterKey ? paging.cursors : [];
  const setCursors = (f: (c: string[]) => string[]) => setPaging({ key: filterKey, cursors: f(cursors) });

  function set(key: keyof MotionFilters, value: string) {
    const next = new URLSearchParams(params);
    if (value) next.set(key, value);
    else next.delete(key);
    setParams(next, { replace: true });
  }

  const page = cursors.length;
  const { data, isLoading, error, isFetching } = useQuery({
    queryKey: ["motions", filterKey, cursors],
    queryFn: () => motionsApi.list(filters, cursors[cursors.length - 1]),
    placeholderData: (prev) => prev,
  });

  return (
    <div>
      <div className="mb-6">
        <h1 className="font-display text-2xl font-extrabold tracking-tight mb-1">Motioner</h1>
        <p className="text-sm text-on-surface-variant leading-relaxed max-w-lg">
          Förslag från riksdagsledamöter och partier, med vad riksdagen beslutade. Senast avgjorda först.
        </p>
      </div>

      <div className="mb-4"><MyPartyPicker /></div>

      <form
        role="search"
        aria-label="Sök bland motioner"
        className="mb-4 flex gap-2"
        onSubmit={(e) => {
          e.preventDefault();
          set("q", text.trim());
        }}
      >
        <input
          type="search"
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Sök i titlar och yrkanden, t.ex. virtuella varor"
          aria-label="Sök i titlar och yrkanden"
          className="flex-1 text-sm rounded border px-3 py-2"
          style={{ borderColor: "var(--color-border)", background: "var(--color-surface-lowest)" }}
        />
        <button type="submit" className="text-sm font-semibold px-4 rounded border cursor-pointer" style={{ borderColor: "var(--color-border)" }}>
          Sök
        </button>
      </form>

      <div className="grid gap-3 mb-6" style={{ gridTemplateColumns: "repeat(auto-fit, minmax(150px, 1fr))" }} data-testid="motion-filters">
        <Select label="Parti" value={filters.party ?? ""} onChange={(v) => set("party", v)} options={PARTIES.map((p) => [p, p])} />
        <Select label="Utskott" value={filters.utskott ?? ""} onChange={(v) => set("utskott", v)} options={UTSKOTT.map(([c, n]) => [c, `${n} (${c})`])} />
        <Select label="Utfall" value={filters.outcome ?? ""} onChange={(v) => set("outcome", v)} options={OUTCOMES.map((o) => [o, o])} />
        <Select label="Riksmöte" value={filters.riksmote ?? ""} onChange={(v) => set("riksmote", v)} options={riksmoten().map((r) => [r, r])} />
        <Select label="Typ" value={filters.kind ?? ""} onChange={(v) => set("kind", v)} options={KINDS.map((k) => [k, k])} />
      </div>

      {isLoading ? (
        <p className="text-sm text-on-surface-variant py-12 text-center">Laddar motioner...</p>
      ) : error ? (
        <p className="text-sm text-contradiction py-12 text-center">Kunde inte ladda motioner.</p>
      ) : (
        <>
          <ul className="space-y-2" aria-busy={isFetching} aria-label="Motioner">
            {(data?.data ?? []).map((m) => <Row key={m.dokId} m={m} />)}
          </ul>
          {(data?.data ?? []).length === 0 ? (
            <p className="text-sm text-on-surface-variant text-center py-12">Inga motioner matchar filtren.</p>
          ) : null}
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
