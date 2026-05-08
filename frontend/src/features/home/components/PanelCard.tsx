import { Link } from "react-router-dom";
import { useMediaQuery } from "@/hooks/useMediaQuery";

interface PanelCardProps {
  title: string;
  showAllHref?: string;
  showAllLabel?: string;
  emptyText?: string;
  isEmpty?: boolean;
  children?: React.ReactNode;
}

export function PanelCard({
  title,
  showAllHref,
  showAllLabel = "Visa alla →",
  emptyText,
  isEmpty,
  children,
}: PanelCardProps) {
  const isMobile = useMediaQuery("(max-width: 640px)");

  return (
    <section
      style={{
        background: "var(--color-sdt-surface)",
        padding: isMobile ? 16 : 24,
        display: "flex",
        flexDirection: "column",
        gap: 12,
        minWidth: 0,
        overflow: "hidden",
      }}
    >
      <header
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 10,
          letterSpacing: "0.15em",
          color: "var(--color-fg-muted)",
          textTransform: "uppercase",
        }}
      >
        {title}
      </header>

      {isEmpty ? (
        <div
          style={{
            fontFamily: "var(--font-body)",
            fontStyle: "italic",
            fontSize: 13,
            color: "var(--color-fg-muted)",
            padding: "8px 0",
          }}
        >
          {emptyText ?? "Ingen aktivitet just nu."}
        </div>
      ) : (
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          {children}
        </div>
      )}

      {showAllHref && !isEmpty && (
        <Link
          to={showAllHref}
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            letterSpacing: "0.1em",
            color: "var(--color-accent)",
            textDecoration: "none",
            marginTop: "auto",
          }}
        >
          {showAllLabel}
        </Link>
      )}
    </section>
  );
}
