import { NavLink } from "react-router-dom";

/** Product name. The old "Tre Kammare" mark implied three chambers. */
export function Wordmark({ size = 17 }: { size?: number }) {
  return (
    <NavLink
      to="/"
      style={{
        display: "inline-flex",
        alignItems: "center",
        textDecoration: "none",
        flexShrink: 0,
        minWidth: 0,
      }}
    >
      <span
        style={{
          fontFamily: "var(--font-serif)",
          fontSize: size,
          fontWeight: 700,
          color: "var(--color-fg)",
          letterSpacing: "-0.02em",
          lineHeight: 1.1,
        }}
      >
        Riksdagskollen
      </span>
    </NavLink>
  );
}
