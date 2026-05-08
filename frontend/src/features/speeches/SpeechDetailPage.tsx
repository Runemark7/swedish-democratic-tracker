import { Link, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { speechesApi } from "@/features/speeches/api";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { SectionSource } from "@/components/sources/SectionSource";

export function SpeechDetailPage() {
  const { id } = useParams<{ id: string }>();
  const isMobile = useMediaQuery("(max-width: 640px)");
  const speechId = Number(id ?? "0");
  const { data, isLoading, error } = useQuery({
    queryKey: ["speech", speechId],
    queryFn: () => speechesApi.getById(speechId),
    enabled: speechId > 0,
  });

  return (
    <div className="sdt-page">
      <div style={{ padding: isMobile ? "18px 14px" : "32px" }}>
        <Link
          to="/"
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            letterSpacing: "0.1em",
            color: "var(--color-accent)",
            textDecoration: "none",
          }}
        >
          ← Tillbaka
        </Link>
        {isLoading && (
          <div style={{ marginTop: 24, color: "var(--color-fg-muted)" }}>Laddar…</div>
        )}
        {error && (
          <div style={{ marginTop: 24, color: "var(--color-pulse)" }}>
            Kunde inte hämta anförandet.
          </div>
        )}
        {data && (
          <article style={{ marginTop: 24, maxWidth: 720 }}>
            <header style={{ marginBottom: 16 }}>
              <div
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 10,
                  letterSpacing: "0.15em",
                  color: "var(--color-fg-muted)",
                  textTransform: "uppercase",
                  marginBottom: 6,
                }}
              >
                ANFÖRANDE · {data.party} <SourceMarker sourceId="riksdagen" /> · {new Date(data.date).toLocaleDateString("sv-SE")} <SourceMarker sourceId="riksdagen" />
              </div>
              <h1
                style={{
                  fontFamily: "var(--font-serif)",
                  fontSize: isMobile ? 24 : 32,
                  margin: 0,
                  color: "var(--color-fg)",
                  lineHeight: 1.2,
                }}
              >
                {data.politicianName}
              </h1>
              {data.topicHeading && (
                <div style={{ fontSize: 14, color: "var(--color-fg-muted)", marginTop: 4 }}>
                  {data.topicHeading}
                </div>
              )}
            </header>
            <p
              style={{
                fontSize: 15,
                lineHeight: 1.7,
                color: "var(--color-fg)",
                whiteSpace: "pre-wrap",
              }}
            >
              {data.speechText || data.snippet}
            </p>
            <div style={{ marginTop: 24 }}>
              <SectionSource sourceIds={["riksdagen"]} />
            </div>
          </article>
        )}
      </div>
    </div>
  );
}
