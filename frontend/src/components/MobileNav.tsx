import { NavLink } from "react-router-dom";
import { Wordmark } from "@/components/Wordmark";
import { PRIMARY_NAV, SECONDARY_NAV, isActive, isLokaltSection } from "@/shared/nav";

interface MobileNavProps {
  pathname: string;
  theme: "dark" | "light";
  toggleTheme: () => void;
}

export function MobileNav({ pathname, theme, toggleTheme }: MobileNavProps) {
  const topRowStyle: React.CSSProperties = {
    position: "sticky",
    top: 0,
    zIndex: 40,
    height: 56,
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    padding: "0 14px",
    background: "color-mix(in oklab, var(--color-bg) 92%, transparent)",
    backdropFilter: "blur(20px) saturate(160%)",
    WebkitBackdropFilter: "blur(20px) saturate(160%)",
    borderBottom: "1px solid var(--color-border)",
  };

  const pillRowStyle: React.CSSProperties = {
    position: "sticky",
    top: 56,
    zIndex: 39,
    background: "color-mix(in oklab, var(--color-bg) 92%, transparent)",
    backdropFilter: "blur(20px) saturate(160%)",
    WebkitBackdropFilter: "blur(20px) saturate(160%)",
    borderBottom: "1px solid var(--color-border)",
  };

  const lokaltActive = isLokaltSection(pathname);

  return (
    <div className="mobile-only" style={{ minWidth: 0, overflow: "hidden" }}>
      <div style={topRowStyle}>
        {/* Balances the theme button so the wordmark stays centered.
            The old left control was a hamburger that did not open a menu. */}
        <span style={{ width: 36, flexShrink: 0 }} aria-hidden="true" />

        <div
          style={{
            position: "absolute",
            left: "50%",
            transform: "translateX(-50%)",
            maxWidth: "calc(100% - 100px)",
          }}
        >
          <Wordmark size={16} />
        </div>

        <button
          onClick={toggleTheme}
          aria-label={theme === "dark" ? "Byt till ljust läge" : "Byt till mörkt läge"}
          style={{
            width: 36,
            height: 36,
            borderRadius: "50%",
            border: "1px solid var(--color-border)",
            background: "transparent",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            cursor: "pointer",
            color: "var(--color-fg-muted)",
            fontSize: 16,
            flexShrink: 0,
          }}
        >
          {theme === "dark" ? "☀" : "☽"}
        </button>
      </div>

      <div style={pillRowStyle}>
        <div
          className="mobile-scroll-pills"
          style={{ gap: 6, padding: "6px 14px", height: 44, alignItems: "center" }}
        >
          {PRIMARY_NAV.map((item) => {
            const active = isActive(item, pathname);
            return (
              <NavLink
                key={item.to}
                to={item.to}
                style={{
                  display: "inline-flex",
                  alignItems: "center",
                  height: 32,
                  padding: "0 14px",
                  borderRadius: 999,
                  border: active ? "none" : "1px solid var(--color-border)",
                  background: active ? "var(--color-fg)" : "transparent",
                  color: active ? "var(--color-bg)" : "var(--color-fg-muted)",
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                  letterSpacing: "1.1px",
                  textTransform: "uppercase",
                  textDecoration: "none",
                  whiteSpace: "nowrap",
                  flexShrink: 0,
                }}
              >
                {item.label}
              </NavLink>
            );
          })}
          <NavLink
            to="/lokalt"
            style={{
              display: "inline-flex",
              alignItems: "center",
              height: 32,
              padding: "0 14px",
              borderRadius: 999,
              border: lokaltActive ? "none" : "1px solid var(--color-accent)",
              background: lokaltActive ? "var(--color-fg)" : "transparent",
              color: lokaltActive ? "var(--color-bg)" : "var(--color-accent)",
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              letterSpacing: "1.1px",
              textTransform: "uppercase",
              textDecoration: "none",
              whiteSpace: "nowrap",
              flexShrink: 0,
            }}
          >
            Lokalt
          </NavLink>
          {SECONDARY_NAV.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              style={{
                display: "inline-flex",
                alignItems: "center",
                height: 32,
                padding: "0 8px",
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                letterSpacing: "0.08em",
                textTransform: "uppercase",
                textDecoration: "none",
                whiteSpace: "nowrap",
                flexShrink: 0,
                color: isActive(item, pathname) ? "var(--color-fg)" : "var(--color-fg-muted)",
              }}
            >
              {item.label}
            </NavLink>
          ))}
        </div>
      </div>
    </div>
  );
}
