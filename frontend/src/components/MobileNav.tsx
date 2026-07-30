import { NavLink, useLocation } from "react-router-dom";

type NavSection = "start" | "riksdag" | "region" | "kommun" | "regering" | "sok";

interface MobileNavProps {
  section: NavSection;
  isRiksdagSection: boolean;
  theme: "dark" | "light";
  toggleTheme: () => void;
}

const navPills = [
  { key: "start",    to: "/",         label: "◆ START"     },
  { key: "riksdag",  to: "/riksdag",  label: "I RIKSDAG"   },
  { key: "region",   to: "/region",   label: "II REGION"   },
  { key: "kommun",   to: "/kommun",   label: "III KOMMUN"  },
  { key: "regering", to: "/regering", label: "IV REGERING" },
  { key: "sok",      to: "/sok",      label: "⌕ SÖK"       },
] as const;

const riksdagTabs = [
  { to: "/parties",     label: "Partier",     matches: (p: string) => p.startsWith("/parties") },
  { to: "/votes",       label: "Omröstningar",matches: (p: string) => p.startsWith("/votes") },
  { to: "/budget",      label: "Budget",      matches: (p: string) => p.startsWith("/budget") },
  { to: "/politicians", label: "Politiker",   matches: (p: string) => p.startsWith("/politicians") },
  { to: "/manifestos",  label: "Manifest",    matches: (p: string) => p.startsWith("/manifestos") },
];

export function MobileNav({ section, isRiksdagSection, theme, toggleTheme }: MobileNavProps) {
  const location = useLocation();

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

  return (
    <div className="mobile-only" style={{ minWidth: 0, overflow: "hidden" }}>
      {/* Row 1 — logo + theme toggle */}
      <div style={topRowStyle}>
        {/* Hamburger stub */}
        <button
          aria-label="Meny"
          style={{
            width: 36,
            height: 36,
            borderRadius: 8,
            border: "1px solid var(--color-border)",
            background: "transparent",
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            gap: 5,
            cursor: "pointer",
            flexShrink: 0,
          }}
        >
          {[0, 1, 2].map((i) => (
            <span
              key={i}
              style={{
                display: "block",
                width: 16,
                height: 2,
                borderRadius: 1,
                background: "var(--color-fg-muted)",
              }}
            />
          ))}
        </button>

        {/* Centered wordmark */}
        <NavLink
          to="/"
          style={{
            position: "absolute",
            left: "50%",
            transform: "translateX(-50%)",
            display: "flex",
            alignItems: "center",
            gap: 8,
            textDecoration: "none",
          }}
        >
          <svg width="18" height="14" viewBox="0 0 20 16" fill="none" aria-hidden="true">
            <rect x="0" y="0"   width="20" height="3" rx="1.5" fill="var(--color-fg)" fillOpacity="0.45" />
            <rect x="0" y="6.5" width="20" height="3" rx="1.5" fill="var(--color-fg)" fillOpacity="0.70" />
            <rect x="0" y="13"  width="20" height="3" rx="1.5" fill="var(--color-fg)" fillOpacity="1"    />
          </svg>
          <span
            style={{
              fontFamily: "var(--font-serif)",
              fontSize: 18,
              fontWeight: 700,
              color: "var(--color-fg)",
              letterSpacing: "0.01em",
            }}
          >
            Tre{" "}
            <em style={{ fontStyle: "italic", color: "var(--color-accent-2)" }}>Kammare</em>
          </span>
        </NavLink>

        {/* Theme toggle */}
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

      {/* Row 2 — chamber pills */}
      <div style={pillRowStyle}>
        <div
          className="mobile-scroll-pills"
          style={{ gap: 6, padding: "6px 14px", height: 44, alignItems: "center" }}
        >
          {navPills.map((pill) => {
            const active = section === pill.key;
            return (
              <NavLink
                key={pill.key}
                to={pill.to}
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
                {pill.label}
              </NavLink>
            );
          })}
        </div>

        {/* Row 2b — Riksdag sub-tabs */}
        {isRiksdagSection && (
          <div
            className="mobile-scroll-pills"
            style={{
              borderTop: "1px solid var(--color-border)",
              gap: 0,
              padding: "0 14px",
            }}
          >
            {riksdagTabs.map((tab) => {
              const active = tab.matches(location.pathname);
              return (
                <NavLink
                  key={tab.to}
                  to={tab.to}
                  style={{
                    display: "inline-flex",
                    alignItems: "center",
                    height: 40,
                    padding: "0 14px",
                    fontFamily: "var(--font-body)",
                    fontSize: 13,
                    fontWeight: active ? 700 : 500,
                    color: active ? "var(--color-fg)" : "var(--color-fg-muted)",
                    borderBottom: active
                      ? "2px solid var(--color-fg)"
                      : "2px solid transparent",
                    textDecoration: "none",
                    whiteSpace: "nowrap",
                    flexShrink: 0,
                  }}
                >
                  {tab.label}
                </NavLink>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
