import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { motionsApi, motionPath } from "@/features/motions/api";
import { SourceMarker } from "@/components/sources/SourceMarker";

/** The motions a member signed, each with its outcome and a link to its page. */
export function MemberMotions({ politicianId }: { politicianId: string }) {
  const { data, isLoading } = useQuery({
    queryKey: ["politician-motions", politicianId],
    queryFn: () => motionsApi.list({ author: politicianId }),
  });
  const rows = data?.data ?? [];
  return (
    <section aria-label="Motioner" data-testid="member-motions">
      <h2 className="text-lg font-semibold mb-2">
        Motioner
        <SourceMarker sourceId="riksdagen" />
      </h2>
      {isLoading ? (
        <p className="text-sm text-on-surface-variant">Laddar...</p>
      ) : rows.length === 0 ? (
        <p className="text-sm italic text-on-surface-variant py-4">Inga motioner registrerade.</p>
      ) : (
        <ul className="m-0 p-0 list-none">
          {rows.map((m) => (
            <li key={m.dokId} style={{ padding: "10px 0", borderBottom: "1px solid var(--color-border)" }}>
              <Link to={motionPath(m.riksmote, m.beteckning)} className="text-sm hover:text-link">
                {m.riksmote}:{m.beteckning} {m.title}
              </Link>
              <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg-muted)" }}>
                {m.kind} · Utfall: {m.outcome.word}
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
