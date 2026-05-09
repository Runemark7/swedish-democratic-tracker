import { Link } from "react-router-dom";
import { PARTY_COLORS } from "@/shared/design";
import type { Speech } from "@/features/speeches/api";

interface SpeechRowProps {
  speech: Speech;
  /** Hide politician name link if rendering on that politician's own page. */
  hidePolitician?: boolean;
}

function formatStamp(iso: string): string {
  const d = new Date(iso);
  const dd = String(d.getDate()).padStart(2, "0");
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  return `${dd}/${mm}`;
}

export function SpeechRow({ speech, hidePolitician }: SpeechRowProps) {
  const pc = PARTY_COLORS[speech.party];
  const initials =
    (speech.politicianName || "?")
      .split(" ")
      .filter(Boolean)
      .slice(0, 2)
      .map((n) => n[0])
      .join("") || "?";

  return (
    <article
      style={{
        display: "flex",
        gap: 12,
        padding: "12px 0",
        borderBottom: "1px solid var(--color-border)",
        alignItems: "flex-start",
        minWidth: 0,
      }}
    >
      {/* Avatar */}
      {speech.politicianImageUrl ? (
        <img
          src={speech.politicianImageUrl}
          alt=""
          loading="lazy"
          onError={(e) => {
            (e.currentTarget as HTMLImageElement).style.display = "none";
          }}
          style={{
            width: 40,
            height: 40,
            borderRadius: "50%",
            objectFit: "cover",
            flexShrink: 0,
          }}
        />
      ) : (
        <div
          style={{
            width: 40,
            height: 40,
            borderRadius: "50%",
            background: pc?.bg ?? "#888",
            color: pc?.text ?? "#fff",
            fontFamily: "var(--font-mono)",
            fontSize: 13,
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

      {/* Body */}
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 8,
            marginBottom: 4,
            flexWrap: "wrap",
          }}
        >
          {!hidePolitician && (
            <Link
              to={`/politicians/${encodeURIComponent(speech.politicianId)}`}
              style={{
                fontFamily: "var(--font-serif)",
                fontSize: 14,
                color: "var(--color-fg)",
                textDecoration: "none",
              }}
            >
              {speech.politicianName || "Anonym"}
            </Link>
          )}
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              fontWeight: 700,
              padding: "1px 6px",
              borderRadius: 2,
              background: pc?.bg ?? "#888",
              color: pc?.text ?? "#fff",
              letterSpacing: "0.05em",
            }}
          >
            {speech.party}
          </span>
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              color: "var(--color-fg-muted)",
              marginLeft: "auto",
            }}
          >
            {formatStamp(speech.date)}
          </span>
        </div>
        {speech.topicHeading && (
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.05em",
              marginBottom: 4,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {speech.topicHeading}
          </div>
        )}
        <p
          style={{
            fontFamily: "var(--font-body)",
            fontSize: 13,
            lineHeight: 1.5,
            color: "var(--color-fg)",
            margin: "0 0 6px",
            display: "-webkit-box",
            WebkitLineClamp: 2,
            WebkitBoxOrient: "vertical",
            overflow: "hidden",
          }}
        >
          {speech.snippet || "Anförandetexten är inte tillgänglig än."}
        </p>
        <Link
          to={`/anforanden/${speech.id}`}
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-accent)",
            textDecoration: "none",
            letterSpacing: "0.05em",
          }}
        >
          Läs hela →
        </Link>
      </div>
    </article>
  );
}
