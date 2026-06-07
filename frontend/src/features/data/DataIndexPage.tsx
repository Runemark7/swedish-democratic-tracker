import { useState } from "react";
import { Link } from "react-router-dom";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { SOURCES, type SourceKind } from "@/components/sources/SourceRegistry";

const KIND_LABEL: Record<SourceKind | "all", string> = {
  all: "Alla",
  api: "API",
  csv: "CSV",
  seed: "Seed",
  synthesized: "Härledd",
};

const KIND_ORDER: ("all" | SourceKind)[] = ["all", "api", "csv", "seed", "synthesized"];

const STATUS_LABEL: Record<string, string> = {
  verified: "VERIFIERAD",
  frozen: "STATISK",
  unsure: "EJ VERIFIERAD",
};

const STATUS_BG: Record<string, { bg: string; fg: string }> = {
  verified: { bg: "transparent", fg: "var(--color-fg-muted)" },
  frozen: { bg: "transparent", fg: "var(--color-fg-muted)" },
  unsure: { bg: "var(--color-warn, #c08a2a)", fg: "var(--color-bg)" },
};

export function DataIndexPage() {
  const isMobile = useMediaQuery("(max-width: 640px)");
  const [filter, setFilter] = useState<"all" | SourceKind>("all");

  const all = Object.values(SOURCES).sort((a, b) => a.name.localeCompare(b.name, "sv"));
  const filtered = filter === "all" ? all : all.filter((e) => e.kind === filter);

  return (
    <div className="sdt-page">
      <header style={{ padding: isMobile ? "20px 14px 16px" : "40px 32px 24px" }}>
        <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, letterSpacing: "2px", color: "var(--color-fg-muted)", textTransform: "uppercase", marginBottom: 8 }}>
          DATAKÄLLOR
        </div>
        <h1 style={{ fontFamily: "var(--font-serif)", fontSize: isMobile ? 28 : 44, margin: 0, fontWeight: 400, lineHeight: 1.05 }}>
          Var kommer datan ifrån?
        </h1>
        <p style={{ fontSize: 13, color: "var(--color-fg-muted)", margin: "8px 0 0", maxWidth: 540 }}>
          Varje siffra på sidan är spårbar. {all.length} källor — API:er, hand-kuraterade migrationer, härledda värden.
        </p>
      </header>

      <div
        style={{
          display: "flex",
          gap: 8,
          padding: isMobile ? "0 14px" : "0 32px",
          flexWrap: "wrap",
          marginBottom: 16,
        }}
      >
        {KIND_ORDER.map((k) => {
          const active = filter === k;
          return (
            <button
              key={k}
              onClick={() => setFilter(k)}
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                padding: "6px 12px",
                border: "1px solid var(--color-border)",
                background: active ? "var(--color-fg)" : "transparent",
                color: active ? "var(--color-bg)" : "var(--color-fg-muted)",
                cursor: "pointer",
                letterSpacing: "0.1em",
                textTransform: "uppercase",
              }}
            >
              {KIND_LABEL[k]}
            </button>
          );
        })}
      </div>

      <div
        style={{
          display: "grid",
          gridTemplateColumns: isMobile ? "minmax(0, 1fr)" : "minmax(0, 1fr) minmax(0, 1fr)",
          gap: 1,
          background: "var(--color-border)",
          border: "1px solid var(--color-border)",
          margin: isMobile ? "0 14px 28px" : "0 32px 28px",
        }}
      >
        {filtered.map((e) => (
          <Link
            key={e.id}
            to={`/data/${e.id}`}
            style={{
              background: "var(--color-sdt-surface)",
              padding: isMobile ? 16 : 20,
              textDecoration: "none",
              color: "inherit",
              display: "flex",
              flexDirection: "column",
              gap: 6,
              minWidth: 0,
            }}
          >
            <div style={{ display: "flex", alignItems: "center", gap: 8, minWidth: 0 }}>
              <span style={{ fontFamily: "var(--font-serif)", fontSize: 18, color: "var(--color-fg)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap", flex: 1, minWidth: 0 }}>
                {e.name}
              </span>
              <span
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 9,
                  color: STATUS_BG[e.verificationStatus].fg,
                  background: STATUS_BG[e.verificationStatus].bg,
                  letterSpacing: "0.1em",
                  textTransform: "uppercase",
                  border: "1px solid var(--color-border)",
                  padding: "1px 6px",
                  flexShrink: 0,
                }}
              >
                {STATUS_LABEL[e.verificationStatus]}
              </span>
              <span
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 9,
                  color: "var(--color-fg-muted)",
                  letterSpacing: "0.1em",
                  textTransform: "uppercase",
                  border: "1px solid var(--color-border)",
                  padding: "1px 6px",
                  flexShrink: 0,
                }}
              >
                {KIND_LABEL[e.kind]}
              </span>
            </div>
            <p style={{ fontSize: 12, color: "var(--color-fg-muted)", margin: 0, lineHeight: 1.5, display: "-webkit-box", WebkitLineClamp: 2, WebkitBoxOrient: "vertical", overflow: "hidden" }}>
              {e.blurb}
            </p>
            <div style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", letterSpacing: "0.1em", textTransform: "uppercase", marginTop: 4 }}>
              SENAST VERIFIERAD {e.lastVerified}
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
