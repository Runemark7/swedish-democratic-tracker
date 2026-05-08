import { Link } from "react-router-dom";
import { SOURCES, type SourceId } from "./SourceRegistry";

interface SectionSourceProps {
  sourceIds: SourceId[];
}

export function SectionSource({ sourceIds }: SectionSourceProps) {
  if (sourceIds.length === 0) return null;

  const entries = sourceIds.map((id) => SOURCES[id]).filter(Boolean);
  if (entries.length === 0) return null;

  return (
    <div
      style={{
        fontFamily: "var(--font-mono)",
        fontSize: 9,
        color: "var(--color-fg-muted)",
        marginTop: 12,
        letterSpacing: "0.05em",
        display: "flex",
        flexWrap: "wrap",
        gap: "2px 8px",
        alignItems: "baseline",
      }}
    >
      <span style={{ textTransform: "uppercase", letterSpacing: "0.1em" }}>Källor:</span>
      {entries.map((e, i) => (
        <span key={e.id}>
          <Link
            to={`/data/${e.id}`}
            style={{
              color: "var(--color-fg-muted)",
              textDecoration: "none",
              borderBottom: "1px dotted var(--color-border)",
            }}
          >
            {e.name}
          </Link>
          {i < entries.length - 1 ? " ·" : ""}
        </span>
      ))}
    </div>
  );
}
