import { Link, useParams } from "react-router-dom";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { SOURCES } from "@/components/sources/SourceRegistry";
import { Markdown } from "./markdown";

const RAW_MARKDOWN = import.meta.glob("../../data-sources/*.md", {
  query: "?raw",
  import: "default",
  eager: true,
}) as Record<string, string>;

const KIND_LABEL: Record<string, string> = {
  api: "API",
  csv: "CSV/ZIP",
  seed: "Seed-data",
  synthesized: "Härledd",
};

const STATUS_LABEL: Record<string, string> = {
  verified: "Verifierad",
  frozen: "Statisk — verifierad en gång",
  unsure: "⚠ Ej verifierad",
};

const STATUS_BG: Record<string, { bg: string; fg: string }> = {
  verified: { bg: "transparent", fg: "var(--color-fg)" },
  frozen: { bg: "transparent", fg: "var(--color-fg)" },
  unsure: { bg: "var(--color-warn, #c08a2a)", fg: "var(--color-bg)" },
};

function getMarkdown(id: string): string | null {
  // Vite glob keys look like "../../data-sources/scb-kfmandat.md"
  const key = Object.keys(RAW_MARKDOWN).find((k) => k.endsWith(`/${id}.md`));
  if (!key) return null;
  const raw = RAW_MARKDOWN[key];
  // Strip frontmatter (between leading --- and second ---)
  const m = raw.match(/^---\n[\s\S]*?\n---\n([\s\S]*)$/);
  return m ? m[1] : raw;
}

export function DataSourcePage() {
  const { id } = useParams<{ id: string }>();
  const isMobile = useMediaQuery("(max-width: 640px)");
  const entry = id ? SOURCES[id] : undefined;
  const md = id ? getMarkdown(id) : null;

  if (!entry || !md) {
    return (
      <div className="sdt-page">
        <div style={{ padding: isMobile ? "20px 14px" : "32px" }}>
          <Link to="/data" style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-accent)", textDecoration: "none", letterSpacing: "0.1em" }}>
            ← Tillbaka till alla källor
          </Link>
          <div style={{ marginTop: 24, color: "var(--color-fg-muted)" }}>
            Källan "{id}" hittades inte.
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="sdt-page">
      <div style={{ padding: isMobile ? "20px 14px 28px" : "32px", maxWidth: 760 }}>
        <Link
          to="/data"
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-accent)",
            textDecoration: "none",
            letterSpacing: "0.1em",
          }}
        >
          ← Tillbaka till alla källor
        </Link>

        <header style={{ marginTop: 18, marginBottom: 20 }}>
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
            {KIND_LABEL[entry.kind]} · senast verifierad {entry.lastVerified}
          </div>
          <h1
            style={{
              fontFamily: "var(--font-serif)",
              fontSize: isMobile ? 28 : 36,
              margin: 0,
              color: "var(--color-fg)",
              lineHeight: 1.1,
              fontWeight: 400,
            }}
          >
            {entry.name}
          </h1>
        </header>

        <div
          style={{
            background: STATUS_BG[entry.verificationStatus].bg,
            color: STATUS_BG[entry.verificationStatus].fg,
            border: "1px solid var(--color-border)",
            padding: "10px 14px",
            marginBottom: 14,
            fontSize: 12,
            display: "flex",
            flexDirection: "column",
            gap: 4,
          }}
        >
          <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, letterSpacing: "0.15em", textTransform: "uppercase" }}>
            {STATUS_LABEL[entry.verificationStatus]} · senast {entry.lastVerified}
          </div>
          {entry.verificationNotes ? (
            <div style={{ fontSize: 12, lineHeight: 1.5 }}>{entry.verificationNotes}</div>
          ) : null}
        </div>

        <div
          style={{
            background: "var(--color-sdt-surface)",
            border: "1px solid var(--color-border)",
            padding: 14,
            marginBottom: 24,
            display: "grid",
            gridTemplateColumns: isMobile ? "1fr" : "auto 1fr",
            gap: "8px 14px",
            fontSize: 12,
            color: "var(--color-fg)",
          }}
        >
          {entry.upstream && (
            <>
              <span style={{ fontFamily: "var(--font-mono)", color: "var(--color-fg-muted)", letterSpacing: "0.1em", textTransform: "uppercase", fontSize: 10 }}>Upstream</span>
              <a
                href={entry.upstream}
                target="_blank"
                rel="noopener noreferrer"
                style={{ color: "var(--color-accent)", textDecoration: "none", overflowWrap: "anywhere" }}
              >
                {entry.upstream}
              </a>
            </>
          )}
          {entry.license && (
            <>
              <span style={{ fontFamily: "var(--font-mono)", color: "var(--color-fg-muted)", letterSpacing: "0.1em", textTransform: "uppercase", fontSize: 10 }}>Licens</span>
              <span>{entry.license}</span>
            </>
          )}
          {entry.freshness && (
            <>
              <span style={{ fontFamily: "var(--font-mono)", color: "var(--color-fg-muted)", letterSpacing: "0.1em", textTransform: "uppercase", fontSize: 10 }}>Färskhet</span>
              <span>{entry.freshness}</span>
            </>
          )}
        </div>

        <Markdown source={md} />
      </div>
    </div>
  );
}
