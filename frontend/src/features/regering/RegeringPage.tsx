import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { regeringApi } from "./api";
import { PartyBadge } from "@/shared/components";

export function RegeringPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["ministers"],
    queryFn: regeringApi.listMinisters,
  });

  if (isLoading) return <div className="text-on-surface-variant py-16 text-center text-sm">Laddar...</div>;
  if (error)    return <div className="text-contradiction py-16 text-center text-sm">Kunde inte ladda regeringen.</div>;

  const groups = data ?? [];

  return (
    <div>
      <div className="mb-8">
        <h2 className="font-display text-2xl font-extrabold tracking-tight mb-1">Regeringen</h2>
        <p className="text-sm text-on-surface-variant leading-relaxed max-w-lg">
          Aktuella statsråd, deras portföljer och propositioner till riksdagen.
        </p>
      </div>

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
    </div>
  );
}
