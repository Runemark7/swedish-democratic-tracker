import { useState } from "react";
import type { AgendaItem } from "@/types/democracy";

const STATUS_LABEL: Record<AgendaItem["status"], string> = {
  active: "Aktiv",
  in_progress: "Pågående",
  completed: "Genomförd",
};
const STATUS_COLOR: Record<AgendaItem["status"], string> = {
  active: "var(--color-accent-2)",
  in_progress: "#5a9fd0",
  completed: "#4caf50",
};

export function AgendaList({ items }: { items: AgendaItem[] }) {
  const [openIdx, setOpenIdx] = useState<number | null>(null);
  return (
    <ol style={{ listStyle: "none", margin: 0, padding: 0 }}>
      {items.map((item, i) => {
        const isOpen = openIdx === i;
        return (
          <li
            key={i}
            style={{ borderBottom: i < items.length - 1 ? "1px solid var(--color-border)" : "none" }}
          >
            <button
              onClick={() => setOpenIdx(isOpen ? null : i)}
              style={{
                display: "flex",
                alignItems: "baseline",
                gap: 12,
                width: "100%",
                padding: "10px 0",
                background: "none",
                border: "none",
                cursor: "pointer",
                textAlign: "left",
              }}
            >
              <span
                style={{
                  fontFamily: "var(--font-serif)",
                  fontStyle: "italic",
                  color: "var(--color-accent-2)",
                  fontSize: 22,
                  lineHeight: 1,
                  minWidth: 20,
                  flexShrink: 0,
                }}
              >
                {i + 1}
              </span>
              <span style={{ flex: 1, fontSize: 14, color: "var(--color-fg)", lineHeight: 1.4 }}>
                {item.title}
              </span>
              <span style={{ fontSize: 11, color: "var(--color-fg-muted)", marginLeft: 8, flexShrink: 0 }}>
                {isOpen ? "▲" : "▼"}
              </span>
            </button>
            {isOpen && (
              <div style={{ padding: "0 0 14px 32px", display: "flex", flexDirection: "column", gap: 8 }}>
                <p style={{ margin: 0, fontSize: 13, color: "var(--color-fg-muted)", lineHeight: 1.5 }}>
                  {item.description}
                </p>
                <div style={{ display: "flex", alignItems: "center", gap: 10, flexWrap: "wrap" }}>
                  <span style={{ fontSize: 11, color: "var(--color-fg-muted)", fontFamily: "var(--font-mono)" }}>
                    Källa: {item.source}
                  </span>
                  <span
                    style={{
                      fontSize: 10,
                      fontFamily: "var(--font-mono)",
                      letterSpacing: "1px",
                      textTransform: "uppercase",
                      color: STATUS_COLOR[item.status],
                      border: `1px solid ${STATUS_COLOR[item.status]}`,
                      padding: "2px 6px",
                      borderRadius: 2,
                    }}
                  >
                    {STATUS_LABEL[item.status]}
                  </span>
                </div>
              </div>
            )}
          </li>
        );
      })}
    </ol>
  );
}
