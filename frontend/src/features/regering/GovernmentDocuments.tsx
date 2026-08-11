import type { AgendaItem } from "@/types/democracy";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { swedishDate } from "@/shared/dates";

/**
 * The government's programme and budget documents, as we have registered them.
 *
 * Replaces a list of five hand-picked points, each with a description we
 * paraphrased and a status we assigned. Selecting whole documents is far less
 * bias-prone than selecting points within them: the reader gets the primary
 * source and extracts from it themselves, and can argue with which documents we
 * list — an argument that is visible and answerable, unlike a shaded paraphrase.
 *
 * Nothing here is summarised, scored or given a progress label.
 */
export function GovernmentDocuments({ items }: { items: AgendaItem[] }) {
  if (items.length === 0) {
    return (
      <p style={{ fontSize: 13, color: "var(--color-fg-muted)", fontStyle: "italic", margin: 0 }}>
        Vi har inga registrerade regeringsdokument just nu.{" "}
        <SourceMarker sourceId="national-agenda" />
      </p>
    );
  }

  return (
    <ul style={{ listStyle: "none", margin: 0, padding: 0 }}>
      {items.map((item, i) => (
        <li
          key={item.id}
          style={{
            borderTop: i > 0 ? "1px solid var(--color-border)" : "none",
            padding: "14px 0",
          }}
        >
          <a
            href={item.url}
            target="_blank"
            rel="noopener noreferrer"
            style={{
              fontFamily: "var(--font-serif)",
              fontSize: 16,
              color: "var(--color-fg)",
              textDecoration: "none",
              fontWeight: 600,
            }}
          >
            {item.title} <span style={{ color: "var(--color-accent)" }}>→</span>
          </a>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--color-fg-muted)",
              marginTop: 5,
              letterSpacing: "0.04em",
            }}
          >
            {item.source}
            {item.published && <> · {swedishDate(item.published)}</>}
            <SourceMarker sourceId="national-agenda" />
          </div>
          <div style={{ fontSize: 12, color: "var(--color-fg-muted)", marginTop: 3 }}>
            {item.issuer}
          </div>
        </li>
      ))}
    </ul>
  );
}
