import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { regeringApi } from "./api";
import { PartyBadge } from "@/shared/components";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { riksdagApi } from "@/features/riksdag/api";
import { GovernmentDocuments } from "./GovernmentDocuments";

export function RegeringPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["ministers"],
    queryFn: regeringApi.listMinisters,
  });
  // Government documents live here rather than on the front page: they are
  // government-level material, and the utskott partition the rest of the site
  // is organised around does not fit them.
  const { data: documents } = useQuery({
    queryKey: ["government-documents"],
    queryFn: riksdagApi.getAgenda,
  });

  // No early return on the ministers query. The document index below is a
  // separate source, and bailing out of the whole page when the minister list is
  // unavailable would take the documents down with it — the reader would see
  // nothing rather than the half we still hold.
  const groups = data ?? [];

  return (
    <div>
      <div className="mb-8">
        <h2 className="font-display text-2xl font-extrabold tracking-tight mb-1">
          Regeringen
          <SourceMarker sourceId="riksdagen" />
        </h2>
        <p className="text-sm text-on-surface-variant leading-relaxed max-w-lg">
          Aktuella statsråd, deras portföljer och propositioner till riksdagen.
        </p>
      </div>

      {isLoading && (
        <p className="text-on-surface-variant text-sm py-4">Laddar statsråd...</p>
      )}
      {error && (
        <p className="text-contradiction text-sm py-4">
          Kunde inte ladda statsråden. Regeringens dokument nedan påverkas inte.
        </p>
      )}

      <div className="space-y-8">
        {groups.map((group) => (
          <div key={group.departmentCode}>
            <h3 className="text-xs font-mono uppercase tracking-widest text-on-surface-variant mb-3">
              {group.department}
            </h3>
            <div className="bg-surface-lowest rounded-xl overflow-hidden">
              {group.ministers?.map((m, i) => (
                <Link
                  key={m.id}
                  to={`/regering/${m.id}`}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: 16,
                    padding: "14px 20px",
                    borderTop: i > 0 ? "1px solid var(--color-surface-high)" : undefined,
                    textDecoration: "none",
                    color: "inherit",
                    transition: "background 0.15s",
                  }}
                  onMouseEnter={(e) => { e.currentTarget.style.background = "var(--color-surface-high)"; }}
                  onMouseLeave={(e) => { e.currentTarget.style.background = "transparent"; }}
                >
                  <PartyBadge party={m.party ?? ""} size="sm" />
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: 14, fontWeight: 600 }}>{m.name}</div>
                    <div style={{ fontSize: 12, color: "var(--color-fg-muted)" }}>{m.title}</div>
                  </div>
                  <span style={{ fontSize: 12, color: "var(--color-fg-muted)" }}>→</span>
                </Link>
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* ── Government documents ──────────────────────────────────────── */}
      <section className="mt-10">
        <h3 className="text-xs font-mono uppercase tracking-widest text-on-surface-variant mb-3">
          Regeringens dokument
        </h3>
        {/* The selection rule is stated because the list is ours. A reader can
            argue with which documents we registered; they cannot argue with a
            paraphrase we never wrote. */}
        <p className="text-sm text-on-surface-variant leading-relaxed max-w-xl mb-4">
          Regeringens program- och budgetdokument för mandatperioden, som vi har
          registrerat dem. Vi länkar dokumenten i sin helhet och sammanfattar dem
          inte. Saknas ett dokument betyder det att vi inte har registrerat det.
        </p>
        <div className="bg-surface-lowest rounded-xl px-5 py-2">
          <GovernmentDocuments items={documents ?? []} />
        </div>
      </section>
    </div>
  );
}
