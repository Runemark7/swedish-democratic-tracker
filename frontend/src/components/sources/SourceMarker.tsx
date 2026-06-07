import { useState, useRef, useEffect } from "react";
import { SOURCES, type SourceId } from "./SourceRegistry";
import { SourcePopover } from "./SourcePopover";

const KIND_GLYPH: Record<string, string> = {
  api: "ⓘ",
  csv: "⤓",
  seed: "✎",
  synthesized: "≈",
};

interface SourceMarkerProps {
  sourceId: SourceId;
  label?: string; // optional accessible label override
}

export function SourceMarker({ sourceId, label }: SourceMarkerProps) {
  const [open, setOpen] = useState(false);
  const buttonRef = useRef<HTMLButtonElement>(null);

  const entry = SOURCES[sourceId];

  // Close on Escape
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open]);

  if (!entry) {
    // Defensive — registry is the source of truth; render an inert glyph.
    return (
      <span
        aria-hidden="true"
        style={{
          marginLeft: 4,
          fontSize: 11,
          color: "var(--color-fg-muted)",
          opacity: 0.5,
        }}
      >
        —
      </span>
    );
  }

  const glyph = KIND_GLYPH[entry.kind] ?? "ⓘ";
  const markerColor =
    entry.verificationStatus === "unsure"
      ? "var(--color-warn, #c08a2a)"
      : "var(--color-fg-muted)";
  const ariaSuffix = entry.verificationStatus === "unsure" ? " (ej verifierad)" : "";

  return (
    <span style={{ position: "relative", display: "inline" }}>
      <button
        ref={buttonRef}
        onClick={(e) => {
          e.preventDefault();
          e.stopPropagation();
          setOpen((v) => !v);
        }}
        aria-label={label ?? `Källa: ${entry.name}${ariaSuffix}`}
        aria-expanded={open}
        style={{
          display: "inline-flex",
          alignItems: "center",
          justifyContent: "center",
          marginLeft: 4,
          padding: 0,
          width: 16,
          height: 16,
          minWidth: 24,
          minHeight: 24,
          background: "transparent",
          border: "none",
          color: markerColor,
          fontSize: 11,
          lineHeight: 1,
          cursor: "pointer",
          verticalAlign: "baseline",
        }}
      >
        {glyph}
      </button>
      {open && (
        <SourcePopover
          entry={entry}
          anchorRef={buttonRef}
          onClose={() => setOpen(false)}
        />
      )}
    </span>
  );
}
