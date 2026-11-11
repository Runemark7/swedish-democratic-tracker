import { PARTIES } from "@/shared/design";
import { useMyParty } from "@/contexts/myParty";

/** Mitt parti: kept only in this browser's localStorage, never sent anywhere. */
export function MyPartyPicker() {
  const { myParty, setMyParty } = useMyParty();
  return (
    <div className="flex flex-wrap items-center gap-2 text-xs text-on-surface-variant" data-testid="my-party-picker">
      <label htmlFor="my-party-select" className="uppercase tracking-wider">Mitt parti</label>
      <select
        id="my-party-select"
        value={myParty ?? ""}
        onChange={(e) => setMyParty(e.target.value || null)}
        className="text-sm text-on-surface rounded border px-2 py-1"
        style={{ borderColor: "var(--color-border)", background: "var(--color-surface-lowest)" }}
      >
        <option value="">Inget valt</option>
        {PARTIES.map((p) => <option key={p} value={p}>{p}</option>)}
      </select>
      {myParty && (
        <button type="button" onClick={() => setMyParty(null)} className="font-semibold text-link cursor-pointer">
          Rensa
        </button>
      )}
      <span>Sparas bara i din webbläsare. Markerar, sorterar inte.</span>
    </div>
  );
}
