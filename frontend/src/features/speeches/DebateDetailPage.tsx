import { useNavigate, useParams } from "react-router-dom";
import { Link } from "react-router-dom";
import { useDocumentFull, useSpeechesByDocument } from "@/hooks/useDemocracy";
import { SpeechRow } from "@/features/speeches/SpeechRow";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { PARTY_COLORS } from "@/shared/design";

const TYPE_LABEL: Record<string, string> = {
  bet: "Betänkande",
  ip: "Interpellation",
  mot: "Motion",
  prop: "Proposition",
  prot: "Protokoll",
  fr: "Skriftlig fråga",
};

function formatDate(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  const dd = String(d.getDate()).padStart(2, "0");
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const yyyy = d.getFullYear();
  return `${dd}/${mm}/${yyyy}`;
}

const backBtnStyle: React.CSSProperties = {
  background: "none",
  border: "none",
  cursor: "pointer",
  padding: 0,
  fontFamily: "var(--font-mono)",
  fontSize: 11,
  color: "var(--color-accent)",
  letterSpacing: "0.1em",
};

export function DebateDetailPage() {
  const { dokId } = useParams<{ dokId: string }>();
  const navigate = useNavigate();
  const isMobile = useMediaQuery("(max-width: 640px)");

  const { data: doc, isLoading, error } = useDocumentFull(dokId);
  const { data: speeches } = useSpeechesByDocument(dokId);

  if (isLoading) {
    return (
      <div className="sdt-page" style={{ padding: isMobile ? "20px 14px" : "32px" }}>
        <div style={{ color: "var(--color-fg-muted)", fontSize: 13 }}>Laddar debatt…</div>
      </div>
    );
  }

  if (error || !doc) {
    return (
      <div className="sdt-page" style={{ padding: isMobile ? "20px 14px" : "32px" }}>
        <button onClick={() => navigate(-1)} style={backBtnStyle}>← Tillbaka</button>
        <div style={{ marginTop: 24, color: "var(--color-pulse)" }}>Kunde inte hämta debatten.</div>
      </div>
    );
  }

  const typeLabel = TYPE_LABEL[doc.type] ?? doc.type.toUpperCase();
  const intressenter = doc.intressenter ?? [];

  return (
    <div className="sdt-page" style={{ paddingBottom: 64 }}>
      <div style={{ padding: isMobile ? "16px 14px 0" : "32px 32px 0", maxWidth: 880 }}>
        {/* Back */}
        <button onClick={() => navigate(-1)} style={backBtnStyle}>← Tillbaka</button>

        {/* Header */}
        <div style={{ marginTop: 16, marginBottom: 24 }}>
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
            {typeLabel}{doc.date ? ` · ${formatDate(doc.date)}` : ""}
          </div>
          <h1
            style={{
              fontFamily: "var(--font-serif)",
              fontStyle: "italic",
              fontSize: isMobile ? 20 : 26,
              fontWeight: 400,
              margin: "0 0 8px",
              lineHeight: 1.25,
              color: "var(--color-fg)",
            }}
          >
            {doc.title}
          </h1>
          {doc.subtitle && (
            <div
              style={{
                fontFamily: "var(--font-body)",
                fontSize: 13,
                color: "var(--color-fg-muted)",
                marginBottom: 8,
              }}
            >
              {doc.subtitle}
            </div>
          )}
          {doc.summary && (
            <div
              style={{
                fontFamily: "var(--font-body)",
                fontSize: 14,
                lineHeight: 1.65,
                color: "var(--color-fg)",
                background: "var(--color-sdt-surface)",
                border: "1px solid var(--color-border)",
                padding: isMobile ? "12px 14px" : "14px 18px",
              }}
            >
              {doc.summary}
            </div>
          )}
        </div>

        {/* Debattdeltagare */}
        {intressenter.length > 0 && (
          <section style={{ marginBottom: 32 }}>
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.15em",
                color: "var(--color-fg-muted)",
                textTransform: "uppercase",
                marginBottom: 12,
              }}
            >
              Debattdeltagare
            </div>
            <div style={{ display: "flex", flexWrap: "wrap", gap: 10 }}>
              {intressenter.map((p) => {
                const pc = PARTY_COLORS[p.party];
                const initials =
                  p.name
                    .split(" ")
                    .filter(Boolean)
                    .slice(0, 2)
                    .map((n) => n[0])
                    .join("") || "?";
                return (
                  <Link
                    key={p.intressentId}
                    to={`/politicians/${encodeURIComponent(p.intressentId)}`}
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 8,
                      padding: "6px 10px",
                      background: "var(--color-sdt-surface)",
                      border: "1px solid var(--color-border)",
                      textDecoration: "none",
                    }}
                  >
                    <div
                      style={{
                        width: 28,
                        height: 28,
                        borderRadius: "50%",
                        background: pc?.bg ?? "#888",
                        color: pc?.text ?? "#fff",
                        fontFamily: "var(--font-mono)",
                        fontSize: 11,
                        fontWeight: 700,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        flexShrink: 0,
                      }}
                    >
                      {initials}
                    </div>
                    <div style={{ minWidth: 0 }}>
                      <div
                        style={{
                          fontFamily: "var(--font-serif)",
                          fontSize: 13,
                          color: "var(--color-fg)",
                          whiteSpace: "nowrap",
                          overflow: "hidden",
                          textOverflow: "ellipsis",
                        }}
                      >
                        {p.name}
                      </div>
                      <div style={{ display: "flex", gap: 4, alignItems: "center", marginTop: 2 }}>
                        <span
                          style={{
                            fontFamily: "var(--font-mono)",
                            fontSize: 9,
                            fontWeight: 700,
                            padding: "1px 5px",
                            background: pc?.bg ?? "#888",
                            color: pc?.text ?? "#fff",
                            letterSpacing: "0.05em",
                          }}
                        >
                          {p.party}
                        </span>
                        {p.role && (
                          <span
                            style={{
                              fontFamily: "var(--font-mono)",
                              fontSize: 9,
                              color: "var(--color-fg-muted)",
                              letterSpacing: "0.04em",
                            }}
                          >
                            {p.role}
                          </span>
                        )}
                      </div>
                    </div>
                  </Link>
                );
              })}
            </div>
          </section>
        )}

        {/* Document body */}
        {doc.bodyHtml && (
          <section style={{ marginBottom: 40 }}>
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.15em",
                color: "var(--color-fg-muted)",
                textTransform: "uppercase",
                marginBottom: 12,
              }}
            >
              Innehåll
            </div>
            <div
              className="riksdagen-doc"
              style={{ fontSize: 15, lineHeight: 1.7, color: "var(--color-fg)" }}
              dangerouslySetInnerHTML={{ __html: doc.bodyHtml }}
            />
          </section>
        )}

        {/* Alla anföranden */}
        {speeches && speeches.length > 0 && (
          <section
            style={{
              borderTop: "1px solid var(--color-border)",
              paddingTop: 24,
              marginBottom: 32,
            }}
          >
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.15em",
                color: "var(--color-fg-muted)",
                textTransform: "uppercase",
                marginBottom: 12,
              }}
            >
              Anföranden i debatten · {speeches.length}
            </div>
            <div>
              {speeches.map((s) => (
                <SpeechRow key={s.id} speech={s} />
              ))}
            </div>
          </section>
        )}

        {/* Source */}
        <div style={{ borderTop: "1px solid var(--color-border)", paddingTop: 16 }}>
          <a
            href={`https://data.riksdagen.se/dokument/${doc.dokId}`}
            target="_blank"
            rel="noopener noreferrer"
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--color-fg-muted)",
              textDecoration: "none",
              letterSpacing: "0.05em",
            }}
          >
            Källa: riksdagen.se →
          </a>
        </div>
      </div>
    </div>
  );
}
