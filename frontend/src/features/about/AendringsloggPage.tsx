import { useInfiniteQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";
import { swedishDate } from "@/shared/dates";

type Entry = components["schemas"]["AendringsloggEntry"];

const PAGE = 50;

const ACTION: Record<Entry["action"], string> = {
  upsert: "Skapad eller ändrad",
  undo: "Ångrad",
};

/** The public, read-only Ändringslogg: what changed, when, and on whose decision. */
export function AendringsloggPage() {
  const q = useInfiniteQuery({
    queryKey: ["aendringslogg"],
    initialPageParam: undefined as number | undefined,
    queryFn: ({ pageParam }) =>
      api.get<{ entries: Entry[] }>(
        `/aendringslogg?limit=${PAGE}${pageParam ? `&before=${pageParam}` : ""}`,
      ),
    getNextPageParam: (last) =>
      last.entries.length < PAGE ? undefined : last.entries[last.entries.length - 1].id,
  });
  const entries = q.data?.pages.flatMap((p) => p.entries) ?? [];

  return (
    <div className="sdt-page" style={{ padding: "24px 32px 64px", maxWidth: 760 }}>
      <h1 className="font-display text-3xl mb-2">Ändringslogg</h1>
      <p className="text-sm leading-relaxed mb-4" data-testid="andringslogg-intro">
        Här står varje ändring i våra kuraterade uppgifter: vad som ändrades, när och på vems
        uppdrag. Loggen kan bara byggas på, aldrig skrivas om. Så här hanteras fel:{" "}
        <Link to="/metod#rattelser" className="text-link underline">Rättelser</Link>.
      </p>
      {q.isPending && <p className="text-sm">Laddar…</p>}
      {q.isError && <p className="text-sm" role="alert">Loggen kunde inte hämtas just nu.</p>}
      {q.isSuccess && entries.length === 0 && <p className="text-sm">Inga ändringar än.</p>}
      <ol className="space-y-3 text-sm" data-testid="andringslogg-list">
        {entries.map((e) => (
          <li key={e.id} className="border-t pt-2" style={{ borderColor: "var(--color-border)" }}>
            <div>
              <strong>{ACTION[e.action] ?? e.action}</strong> · {e.entity} {e.entityId} · post {e.id}
              {e.undoes != null && <> (ångrar post {e.undoes})</>}
            </div>
            <div className="text-xs" style={{ color: "var(--color-fg-muted)" }}>
              {swedishDate(e.createdAt)} · {e.onBehalfOf}
              {e.chatRef && <> · referens {e.chatRef}</>}
              {e.redacted && <> · en sida av ändringen är inte offentlig</>}
            </div>
          </li>
        ))}
      </ol>
      {q.hasNextPage && (
        <button
          type="button"
          className="text-link underline text-sm mt-4"
          onClick={() => q.fetchNextPage()}
          disabled={q.isFetchingNextPage}
        >
          Visa äldre
        </button>
      )}
    </div>
  );
}
