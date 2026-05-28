import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { regeringApi } from "./api";
import { PartyBadge } from "@/shared/components";

export function MinisterDetailPage() {
  const { id = "" } = useParams<{ id: string }>();

  const { data, isLoading, error } = useQuery({
    queryKey: ["minister", id],
    queryFn: () => regeringApi.getMinister(id),
    enabled: !!id,
  });

  if (isLoading) return <div className="text-on-surface-variant py-16 text-center text-sm">Laddar...</div>;
  if (error || !data?.minister) return <div className="text-contradiction py-16 text-center text-sm">Kunde inte ladda statsrådet.</div>;

  const { minister: m, proposals } = data;

  return (
    <div className="max-w-2xl mx-auto">
      <div className="mb-6">
        <Link to="/regering" style={{ fontSize: 12, color: "var(--color-fg-muted)", textDecoration: "none" }}>
          ← Regeringen
        </Link>
        <div className="mt-4 flex items-center gap-3">
          <PartyBadge party={m.party ?? ""} size="lg" />
          <div>
            <h2 className="font-display text-2xl font-extrabold tracking-tight">{m.name}</h2>
            <p className="text-sm text-on-surface-variant">{m.title}</p>
            <p className="text-xs text-on-surface-variant mt-0.5 font-mono uppercase tracking-wider">
              {m.department}
            </p>
          </div>
        </div>
      </div>

      {m.bio && (
        <div className="bg-surface-lowest rounded-xl p-4 mb-6 text-sm leading-relaxed">
          {m.bio}
        </div>
      )}

      {m.politicianId && (
        <div className="mb-6">
          <Link
            to={`/politicians/${m.politicianId}`}
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: 6,
              fontSize: 13,
              color: "var(--color-accent)",
              textDecoration: "none",
              padding: "8px 16px",
              borderRadius: 8,
              border: "1px solid var(--color-border)",
            }}
          >
            Även riksdagsledamot — visa rösthistorik →
          </Link>
        </div>
      )}

      <div>
        <h3 className="text-xs font-mono uppercase tracking-widest text-on-surface-variant mb-3">
          Propositioner från departementet
        </h3>
        {(!proposals || proposals.length === 0) ? (
          <p className="text-sm text-on-surface-variant py-4">Inga propositioner laddade ännu.</p>
        ) : (
          <div className="bg-surface-lowest rounded-xl overflow-hidden">
            {proposals.map((p, i) => (
              <a
                key={p.id}
                href={p.url}
                target="_blank"
                rel="noopener noreferrer"
                style={{
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                  gap: 12,
                  padding: "12px 20px",
                  borderTop: i > 0 ? "1px solid var(--color-surface-high)" : undefined,
                  textDecoration: "none",
                  color: "inherit",
                  transition: "background 0.15s",
                }}
                onMouseEnter={(e) => { e.currentTarget.style.background = "var(--color-surface-high)"; }}
                onMouseLeave={(e) => { e.currentTarget.style.background = "transparent"; }}
              >
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 13, fontWeight: 500, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
                    {p.title}
                  </div>
                  <div style={{ fontSize: 11, color: "var(--color-fg-muted)", marginTop: 2 }}>
                    {p.riksdagYear}{p.publishedAt ? ` · ${new Date(p.publishedAt).toLocaleDateString("sv-SE")}` : ""}
                  </div>
                </div>
                <span style={{ fontSize: 11, color: "var(--color-fg-muted)", flexShrink: 0 }}>↗</span>
              </a>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
