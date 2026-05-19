import { Link, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { speechesApi } from "@/features/speeches/api";
import { politiciansApi } from "@/features/politicians/api";
import { useDocument } from "@/hooks/useDemocracy";
import { useSpeechesByDocument } from "@/hooks/useDemocracy";
import { SpeechRow } from "@/features/speeches/SpeechRow";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { PARTY_COLORS, partyShortToName, DOC_TYPE_LABEL } from "@/shared/design";

const WEEKDAY = ["sön", "mån", "tis", "ons", "tor", "fre", "lör"];

function formatLongStamp(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  const wd = WEEKDAY[d.getDay()];
  const dd = String(d.getDate()).padStart(2, "0");
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const yyyy = d.getFullYear();
  return `${wd} ${dd}/${mm}/${yyyy}`;
}

export function SpeechDetailPage() {
  const { id } = useParams<{ id: string }>();
  const isMobile = useMediaQuery("(max-width: 640px)");
  const speechId = Number(id ?? "0");

  const { data: speech, isLoading, error } = useQuery({
    queryKey: ["speech", speechId],
    queryFn: () => speechesApi.getById(speechId),
    enabled: speechId > 0,
  });

  const { data: politician } = useQuery({
    queryKey: ["politician", speech?.politicianId],
    queryFn: () => politiciansApi.getById(speech?.politicianId ?? ""),
    enabled: !!speech?.politicianId,
    staleTime: 5 * 60 * 1000,
  });

  const relDokId = speech?.relatedDokId;
  const { data: docToShow } = useDocument(relDokId);

  const { data: siblingSpeeches } = useSpeechesByDocument(relDokId);

  if (isLoading) {
    return (
      <div className="sdt-page" style={{ padding: isMobile ? "20px 14px" : "32px" }}>
        <div style={{ color: "var(--color-fg-muted)", fontSize: 13 }}>Laddar anförande…</div>
      </div>
    );
  }

  if (error || !speech) {
    return (
      <div className="sdt-page" style={{ padding: isMobile ? "20px 14px" : "32px" }}>
        <Link to="/" style={{ color: "var(--color-accent)", fontFamily: "var(--font-mono)", fontSize: 11, textDecoration: "none" }}>
          ← Tillbaka
        </Link>
        <div style={{ marginTop: 24, color: "var(--color-pulse)" }}>Kunde inte hämta anförandet.</div>
      </div>
    );
  }

  const pc = PARTY_COLORS[speech.party];
  const partyName = partyShortToName(speech.party);
  const initials =
    (speech.politicianName || "?")
      .split(" ")
      .filter(Boolean)
      .slice(0, 2)
      .map((n) => n[0])
      .join("") || "?";

  const otherSpeeches = (siblingSpeeches ?? []).filter((s) => s.id !== speech.id);
  const docTypeLabel = docToShow ? DOC_TYPE_LABEL[docToShow.type] ?? docToShow.type.toUpperCase() : null;
  const docLink =
    docToShow?.type === "bet" && docToShow.beteckning
      ? `/beslut/${encodeURIComponent(docToShow.beteckning)}`
      : null;

  return (
    <div className="sdt-page" style={{ paddingBottom: 64 }}>
      <div style={{ padding: isMobile ? "16px 14px 0" : "32px 32px 0", maxWidth: 880 }}>
        {/* Back link */}
        <Link
          to="/"
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-accent)",
            textDecoration: "none",
            letterSpacing: "0.1em",
          }}
        >
          ← Tillbaka
        </Link>

        {/* Politician card */}
        <section
          style={{
            marginTop: 16,
            padding: isMobile ? 16 : 20,
            background: pc?.light ?? "var(--color-sdt-surface)",
            border: "1px solid var(--color-border)",
            display: "flex",
            gap: 16,
            alignItems: "center",
            minWidth: 0,
          }}
        >
          {speech.politicianImageUrl ? (
            <img
              src={speech.politicianImageUrl}
              alt=""
              loading="lazy"
              onError={(e) => {
                (e.currentTarget as HTMLImageElement).style.display = "none";
              }}
              style={{
                width: 80,
                height: 80,
                borderRadius: "50%",
                objectFit: "cover",
                flexShrink: 0,
              }}
            />
          ) : (
            <div
              style={{
                width: 80,
                height: 80,
                borderRadius: "50%",
                background: pc?.bg ?? "#888",
                color: pc?.text ?? "#fff",
                fontFamily: "var(--font-mono)",
                fontSize: 22,
                fontWeight: 700,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                flexShrink: 0,
              }}
            >
              {initials}
            </div>
          )}
          <div style={{ minWidth: 0 }}>
            <Link
              to={`/politicians/${encodeURIComponent(speech.politicianId)}`}
              style={{
                fontFamily: "var(--font-serif)",
                fontSize: isMobile ? 22 : 28,
                color: "var(--color-fg)",
                textDecoration: "none",
                lineHeight: 1.1,
                display: "block",
              }}
            >
              {speech.politicianName || "Anonym"}
            </Link>
            <div
              style={{
                marginTop: 6,
                display: "flex",
                gap: 8,
                alignItems: "center",
                flexWrap: "wrap",
              }}
            >
              <Link
                to={`/parties/${encodeURIComponent(speech.party)}`}
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                  fontWeight: 700,
                  padding: "2px 8px",
                  background: pc?.bg ?? "#888",
                  color: pc?.text ?? "#fff",
                  textDecoration: "none",
                  letterSpacing: "0.05em",
                }}
              >
                {speech.party}
              </Link>
              <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-fg-muted)" }}>
                {partyName}
              </span>
              {politician?.constituency && (
                <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-fg-muted)" }}>
                  · {politician.constituency}
                </span>
              )}
            </div>
          </div>
        </section>

        {/* Anförande meta */}
        <div
          style={{
            marginTop: 24,
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "0.15em",
            color: "var(--color-fg-muted)",
            textTransform: "uppercase",
          }}
        >
          ANFÖRANDE · {formatLongStamp(speech.date)}
        </div>
        {speech.topicHeading && (
          <h1
            style={{
              fontFamily: "var(--font-serif)",
              fontStyle: "italic",
              fontSize: isMobile ? 20 : 24,
              fontWeight: 400,
              margin: "8px 0 16px",
              lineHeight: 1.25,
              color: "var(--color-fg)",
            }}
          >
            {speech.topicHeading}
          </h1>
        )}

        {/* Om debatten */}
        {docToShow && (
          <section
            style={{
              padding: isMobile ? "12px 14px" : "14px 18px",
              background: "var(--color-sdt-surface)",
              border: "1px solid var(--color-border)",
              marginBottom: 24,
              display: "flex",
              flexDirection: "column",
              gap: 6,
            }}
          >
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.15em",
                color: "var(--color-fg-muted)",
                textTransform: "uppercase",
              }}
            >
              Om debatten
            </div>
            <div style={{ display: "flex", alignItems: "baseline", gap: 8, flexWrap: "wrap" }}>
              {docTypeLabel && (
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 10,
                    color: "var(--color-fg-muted)",
                    border: "1px solid var(--color-border)",
                    padding: "1px 6px",
                    letterSpacing: "0.05em",
                  }}
                >
                  {docTypeLabel}
                </span>
              )}
              {docLink ? (
                <Link
                  to={docLink}
                  style={{
                    fontFamily: "var(--font-serif)",
                    fontSize: 15,
                    color: "var(--color-fg)",
                    textDecoration: "none",
                    borderBottom: "1px dotted var(--color-border)",
                  }}
                >
                  {docToShow.title}
                </Link>
              ) : (
                <span style={{ fontFamily: "var(--font-serif)", fontSize: 15, color: "var(--color-fg)" }}>
                  {docToShow.title}
                </span>
              )}
            </div>
            {docLink && (
              <Link
                to={docLink}
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                  color: "var(--color-accent)",
                  textDecoration: "none",
                  letterSpacing: "0.05em",
                }}
              >
                → Läs hela betänkandet
              </Link>
            )}
            {relDokId && (
              <Link
                to={`/debatt/${encodeURIComponent(relDokId)}`}
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                  color: "var(--color-accent)",
                  textDecoration: "none",
                  letterSpacing: "0.05em",
                }}
              >
                → Se hela debatten
              </Link>
            )}
          </section>
        )}

        {/* Speech body */}
        {speech.speechText ? (
          <div
            className="riksdagen-doc"
            style={{
              fontSize: 15,
              lineHeight: 1.7,
              color: "var(--color-fg)",
              marginBottom: 32,
            }}
            dangerouslySetInnerHTML={{ __html: speech.speechText }}
          />
        ) : (
          <p
            style={{
              fontSize: 15,
              lineHeight: 1.7,
              color: "var(--color-fg-muted)",
              fontStyle: "italic",
              marginBottom: 32,
            }}
          >
            {speech.snippet || "Anförandetexten är inte tillgänglig."}
          </p>
        )}

        {/* Resten av debatten */}
        {otherSpeeches.length > 0 && (
          <section
            style={{
              borderTop: "1px solid var(--color-border)",
              paddingTop: 24,
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
              Resten av debatten · {otherSpeeches.length} anföranden
            </div>
            <div>
              {otherSpeeches.map((s) => (
                <SpeechRow key={s.id} speech={s} />
              ))}
            </div>
          </section>
        )}
      </div>
    </div>
  );
}
