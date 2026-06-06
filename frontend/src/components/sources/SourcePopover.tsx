import { useEffect, useRef, type RefObject } from "react";
import { Link } from "react-router-dom";
import type { SourceEntry } from "./SourceRegistry";

const KIND_LABEL: Record<string, string> = {
  api: "API",
  csv: "CSV/ZIP",
  seed: "Seed-data",
  synthesized: "Härledd",
};

const STATUS_LABEL: Record<string, string> = {
  verified: "Verifierad",
  frozen: "Statisk källa — verifierad en gång",
  unsure: "⚠ Ej verifierad",
};

const STATUS_COLOR: Record<string, string> = {
  verified: "var(--color-fg-muted)",
  frozen: "var(--color-fg-muted)",
  unsure: "var(--color-warn, #c08a2a)",
};

interface SourcePopoverProps {
  entry: SourceEntry;
  anchorRef: RefObject<HTMLElement | null>;
  onClose: () => void;
}

export function SourcePopover({ entry, anchorRef, onClose }: SourcePopoverProps) {
  const popoverRef = useRef<HTMLDivElement>(null);

  // Close on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      const target = e.target as Node;
      if (
        popoverRef.current &&
        !popoverRef.current.contains(target) &&
        anchorRef.current &&
        !anchorRef.current.contains(target)
      ) {
        onClose();
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [anchorRef, onClose]);

  return (
    <div
      ref={popoverRef}
      role="dialog"
      aria-label={`Källa: ${entry.name}`}
      style={{
        position: "absolute",
        top: "calc(100% + 6px)",
        left: 0,
        zIndex: 60,
        width: 280,
        maxWidth: "calc(100vw - 28px)",
        background: "var(--color-sdt-surface)",
        border: "1px solid var(--color-border)",
        boxShadow: "0 8px 24px rgba(0,0,0,0.3)",
        padding: 14,
        display: "flex",
        flexDirection: "column",
        gap: 10,
        fontFamily: "var(--font-body)",
        color: "var(--color-fg)",
      }}
    >
      <div
        style={{
          fontFamily: "var(--font-serif)",
          fontSize: 14,
          fontWeight: 400,
          lineHeight: 1.3,
        }}
      >
        {entry.name}
      </div>
      <div
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 10,
          color: "var(--color-fg-muted)",
          letterSpacing: "0.05em",
          textTransform: "uppercase",
        }}
      >
        {KIND_LABEL[entry.kind] ?? entry.kind} · senast verifierad {entry.lastVerified}
      </div>
      <div
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 11,
          color: STATUS_COLOR[entry.verificationStatus],
          letterSpacing: "0.03em",
        }}
      >
        {STATUS_LABEL[entry.verificationStatus]}
        {entry.verificationNotes ? (
          <div
            style={{
              marginTop: 4,
              fontSize: 11,
              fontFamily: "var(--font-body)",
              color: "var(--color-fg)",
              lineHeight: 1.4,
            }}
          >
            {entry.verificationNotes}
          </div>
        ) : null}
      </div>
      <div
        style={{
          fontSize: 12,
          color: "var(--color-fg)",
          lineHeight: 1.5,
          borderTop: "1px solid var(--color-border)",
          paddingTop: 10,
        }}
      >
        {entry.blurb}
      </div>
      <div style={{ display: "flex", flexDirection: "column", gap: 4, borderTop: "1px solid var(--color-border)", paddingTop: 10 }}>
        <Link
          to={`/data/${entry.id}`}
          onClick={onClose}
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-accent)",
            textDecoration: "none",
            letterSpacing: "0.05em",
          }}
        >
          → Läs hur vi använder datan
        </Link>
        {entry.upstream && (
          <a
            href={entry.upstream}
            target="_blank"
            rel="noopener noreferrer"
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--color-accent)",
              textDecoration: "none",
              letterSpacing: "0.05em",
            }}
          >
            ↗ Öppna upstream-källan
          </a>
        )}
      </div>
    </div>
  );
}
