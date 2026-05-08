import { Link } from "react-router-dom";
import { PARTY_COLORS, partyShortToName } from "@/shared/design";
import type { Speech } from "@/features/speeches/api";

interface PartySpeechCardProps {
  party: string;
  speech: Speech | null;
}

const WEEKDAY = ["SÖN", "MÅN", "TIS", "ONS", "TOR", "FRE", "LÖR"];

function formatStamp(iso: string): string {
  const d = new Date(iso);
  const wd = WEEKDAY[d.getDay()];
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${wd} ${dd}/${mm}`;
}

export function PartySpeechCard({ party, speech }: PartySpeechCardProps) {
  const color = PARTY_COLORS[party]?.bg ?? "#7c8896";
  const partyName = partyShortToName(party);

  const body = (
    <article
      style={{
        background: "var(--color-sdt-surface)",
        border: "1px solid var(--color-border)",
        padding: 14,
        display: "flex",
        flexDirection: "column",
        gap: 8,
        minHeight: 130,
        height: "100%",
        minWidth: 0,
        overflow: "hidden",
      }}
    >
      <div style={{ display: "flex", alignItems: "center", gap: 8, minWidth: 0 }}>
        <span
          style={{
            display: "inline-block",
            width: 4,
            height: 18,
            background: color,
            flexShrink: 0,
          }}
        />
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            fontWeight: 700,
            color: "var(--color-fg)",
            letterSpacing: "0.05em",
            flexShrink: 0,
          }}
        >
          {party}
        </span>
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            color: "var(--color-fg-muted)",
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
            minWidth: 0,
            flex: 1,
          }}
        >
          {partyName}
        </span>
      </div>

      {speech ? (
        <>
          <div
            style={{
              fontFamily: "var(--font-serif)",
              fontSize: 14,
              color: "var(--color-fg)",
              lineHeight: 1.3,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {speech.politicianName || "Anonym"}
          </div>
          <p
            style={{
              fontFamily: "var(--font-body)",
              fontSize: 12,
              fontStyle: "italic",
              color: "var(--color-fg-muted)",
              lineHeight: 1.4,
              margin: 0,
              display: "-webkit-box",
              WebkitLineClamp: 2,
              WebkitBoxOrient: "vertical",
              overflow: "hidden",
            }}
          >
            "{speech.snippet}"
          </p>
          <div
            style={{
              marginTop: "auto",
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.1em",
              textTransform: "uppercase",
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {formatStamp(speech.date)}
            {speech.topicHeading ? ` · ${speech.topicHeading}` : ""}
          </div>
        </>
      ) : (
        <div
          style={{
            fontFamily: "var(--font-body)",
            fontStyle: "italic",
            fontSize: 12,
            color: "var(--color-fg-muted)",
            marginTop: 6,
          }}
        >
          Inget anförande senaste 30 dagarna
        </div>
      )}
    </article>
  );

  if (speech) {
    return (
      <Link
        to={`/anforanden/${speech.id}`}
        style={{
          textDecoration: "none",
          color: "inherit",
          display: "block",
          height: "100%",
          minWidth: 0,
          overflow: "hidden",
        }}
      >
        {body}
      </Link>
    );
  }
  return body;
}
