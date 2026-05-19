import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { riksdagApi } from "./api";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { useMediaQuery } from "@/hooks/useMediaQuery";

const STATUS_LABEL: Record<string, string> = {
  active: "Aktiv",
  in_progress: "Pågående",
  completed: "Genomförd",
};

const STATUS_COLOR: Record<string, string> = {
  active: "var(--color-accent-2)",
  in_progress: "#5a9fd0",
  completed: "#4caf50",
};

export function AgendaDetailPage() {
  const isMobile = useMediaQuery("(max-width: 640px)");
  const { id } = useParams<{ id: string }>();
  const numId = id ? parseInt(id, 10) : 0;

  const { data: item, isLoading, error } = useQuery({
    queryKey: ["agenda-item", numId],
    queryFn: () => riksdagApi.getAgendaItem(numId),
    enabled: numId > 0,
  });

  if (isLoading) {
    return (
      <div className="sdt-page" style={{ padding: "40px 32px" }}>
        <div style={{ height: 200, background: "var(--color-track)", borderRadius: 4 }} />
      </div>
    );
  }

  if (error || !item) {
    return (
      <div className="sdt-page" style={{ padding: "40px 32px" }}>
        <div style={{ fontSize: 14, color: "var(--color-fg-muted)" }}>
          Agendapunkt hittades inte.
        </div>
        <Link to="/riksdag" style={{ fontSize: 13, color: "var(--color-accent-2)", marginTop: 12, display: "inline-block" }}>
          ← Tillbaka till Riksdagen
        </Link>
      </div>
    );
  }

  return (
    <div className="sdt-page" style={{ padding: isMobile ? "18px 14px" : "40px 32px" }}>
      {/* Breadcrumb */}
      <div style={{ marginBottom: 24 }}>
        <Link to="/riksdag" style={{ fontSize: 13, color: "var(--color-accent-2)", textDecoration: "none" }}>
          ← Riksdagen
        </Link>
      </div>

      {/* Title */}
      <h1 style={{
        fontFamily: "var(--font-serif)",
        fontSize: isMobile ? 28 : 42,
        fontWeight: 400,
        letterSpacing: "-1px",
        color: "var(--color-fg)",
        marginBottom: 16,
        lineHeight: 1.1,
      }}>
        {item.title}
      </h1>

      {/* Status badge */}
      <div style={{ marginBottom: 24 }}>
        <span style={{
          fontSize: 11,
          fontFamily: "var(--font-mono)",
          letterSpacing: "1px",
          textTransform: "uppercase",
          color: STATUS_COLOR[item.status] || "var(--color-fg-muted)",
          border: `1px solid ${STATUS_COLOR[item.status] || "var(--color-fg-muted)"}`,
          padding: "4px 10px",
          borderRadius: 2,
        }}>
          {STATUS_LABEL[item.status] || item.status}
        </span>
      </div>

      {/* Description */}
      <div style={{
        fontSize: 15,
        lineHeight: 1.6,
        color: "var(--color-fg)",
        marginBottom: 32,
        maxWidth: 640,
      }}>
        {item.description}
      </div>

      {/* Source */}
      <div style={{
        fontSize: 12,
        color: "var(--color-fg-muted)",
        fontFamily: "var(--font-mono)",
        borderTop: "1px solid var(--color-border)",
        paddingTop: 16,
      }}>
        Källa: {item.source}
        <SourceMarker sourceId="derived-agenda" />
      </div>
    </div>
  );
}
