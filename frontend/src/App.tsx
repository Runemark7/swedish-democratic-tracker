import { Routes, Route, NavLink, useLocation, Link } from "react-router-dom";
import { PartiesPage } from "./features/parties/PartiesPage";
import { PartyDetailPage } from "./features/parties/PartyDetailPage";
import { PartyGoalsPage } from "./features/parties/PartyGoalsPage";
import { GoalVotesPage } from "./features/parties/GoalVotesPage";
import { PoliticiansPage } from "./features/politicians/PoliticiansPage";
import { PoliticianPage } from "./features/politicians/PoliticianPage";
import { VotesPage } from "./features/votes/VotesPage";
import { VoteDetailPage } from "./features/votes/VoteDetailPage";
import { VoteSharePage } from "./features/votes/VoteSharePage";
import { BeslutDetailPage } from "./features/votes/BeslutDetailPage";
import { BudgetPage } from "./features/budget/BudgetPage";
import { AreaHistoryPage } from "./features/budget/AreaHistoryPage";
import { ManifestosPage } from "./features/manifestos/ManifestosPage";
import { RegionLandingPage } from "./features/regions/RegionLandingPage";
import { RegionDetailPage } from "./features/regions/RegionDetailPage";
import { MunicipalityLandingPage } from "./features/municipalities/MunicipalityLandingPage";
import { MunicipalityDetailPage } from "./features/municipalities/MunicipalityDetailPage";
import { MunicipalityComparePage } from "./features/municipalities/MunicipalityComparePage";
import { HomePage } from "./features/home/HomePage";
import { RiksdagPage } from "./features/riksdag/RiksdagPage";
import { AuthorityDetailPage } from "./features/riksdag/AuthorityDetailPage";
import { SearchPage } from "./features/search/SearchPage";
import { SpeechDetailPage } from "./features/speeches/SpeechDetailPage";
import { DataIndexPage } from "./features/data/DataIndexPage";
import { DataSourcePage } from "./features/data/DataSourcePage";
import { useTheme } from "./contexts/ThemeContext";
import { useMediaQuery } from "./hooks/useMediaQuery";
import { MobileNav } from "./components/MobileNav";

// ── Section detection ─────────────────────────────────────────────────────
type NavSection = "start" | "riksdag" | "region" | "kommun" | "sok";

function sectionFromPath(pathname: string): NavSection {
  if (
    pathname.startsWith("/riksdag") ||
    pathname.startsWith("/parties") ||
    pathname.startsWith("/votes") ||
    pathname.startsWith("/budget") ||
    pathname.startsWith("/politicians") ||
    pathname.startsWith("/manifestos") ||
    pathname === "/"
  ) {
    if (pathname === "/") return "start";
    return "riksdag";
  }
  if (pathname.startsWith("/region")) return "region";
  if (pathname.startsWith("/kommun")) return "kommun";
  if (pathname.startsWith("/sok")) return "sok";
  return "start";
}

function isRiksdagSection(pathname: string): boolean {
  return (
    pathname.startsWith("/riksdag") ||
    pathname.startsWith("/parties") ||
    pathname.startsWith("/votes") ||
    pathname.startsWith("/budget") ||
    pathname.startsWith("/politicians") ||
    pathname.startsWith("/manifestos")
  );
}

// ── Live date string ───────────────────────────────────────────────────────
function liveDateStr(): string {
  const d = new Date();
  const dd = String(d.getDate()).padStart(2, "0");
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const yyyy = d.getFullYear();
  return `${dd}.${mm}.${yyyy}`;
}

// ── Three-bar logo icon ────────────────────────────────────────────────────
function TreKammareLogo() {
  return (
    <NavLink
      to="/"
      style={{
        display: "flex",
        alignItems: "center",
        gap: 10,
        textDecoration: "none",
        flexShrink: 0,
      }}
    >
      <svg
        width="20"
        height="16"
        viewBox="0 0 20 16"
        fill="none"
        aria-hidden="true"
      >
        <rect x="0" y="0"  width="20" height="3" rx="1.5" fill="var(--color-fg)" fillOpacity="0.45" />
        <rect x="0" y="6.5" width="20" height="3" rx="1.5" fill="var(--color-fg)" fillOpacity="0.70" />
        <rect x="0" y="13" width="20" height="3" rx="1.5" fill="var(--color-fg)" fillOpacity="1"    />
      </svg>
      <span
        style={{
          fontFamily: "var(--font-serif)",
          fontSize: 17,
          fontWeight: 700,
          color: "var(--color-fg)",
          letterSpacing: "0.01em",
        }}
      >
        Tre{" "}
        <em
          style={{
            fontStyle: "italic",
            color: "var(--color-accent-2)",
          }}
        >
          Kammare
        </em>
      </span>
    </NavLink>
  );
}

// ── Nav pill ───────────────────────────────────────────────────────────────
interface PillNavProps {
  to: string;
  label: string;
  active: boolean;
}

function PillNav({ to, label, active }: PillNavProps) {
  return (
    <NavLink
      to={to}
      style={{
        display: "inline-flex",
        alignItems: "center",
        padding: "6px 14px",
        borderRadius: 999,
        border: active ? "none" : "1px solid var(--color-border)",
        background: active ? "var(--color-fg)" : "transparent",
        color: active ? "var(--color-bg)" : "var(--color-fg-muted)",
        fontFamily: "var(--font-mono)",
        fontSize: 11,
        letterSpacing: "1px",
        textTransform: "uppercase",
        textDecoration: "none",
        whiteSpace: "nowrap",
        cursor: "pointer",
        transition: "background 0.15s, color 0.15s",
      }}
    >
      {label}
    </NavLink>
  );
}

// ── Main App ───────────────────────────────────────────────────────────────
export default function App() {
  const location = useLocation();
  const { theme, toggleTheme } = useTheme();

  // Share-only pages render without app chrome
  if (location.pathname.endsWith("/dela")) {
    return (
      <Routes>
        <Route path="/votes/:beteckning/:punkt/dela" element={<VoteSharePage />} />
      </Routes>
    );
  }

  const isMobile = useMediaQuery("(max-width: 640px)");
  const section = sectionFromPath(location.pathname);
  const showRiksdagTabs = isRiksdagSection(location.pathname);

  const navPills = [
    { key: "start",   to: "/",        label: "◆ START"     },
    { key: "riksdag", to: "/riksdag",  label: "I RIKSDAG"   },
    { key: "region",  to: "/region",  label: "II REGION"   },
    { key: "kommun",  to: "/kommun",  label: "III KOMMUN"  },
    { key: "sok",     to: "/sok",     label: "⌕ SÖK"       },
  ] as const;

  return (
    <div className="sdt-page">
      {/* ── Mobile nav (≤640 px) ──────────────────────────────────────── */}
      <MobileNav
        section={section}
        isRiksdagSection={showRiksdagTabs}
        theme={theme}
        toggleTheme={toggleTheme}
      />

      {/* ── Desktop top nav (>640 px) ─────────────────────────────────── */}
      <nav
        className="desktop-only"
        style={{
          background: "var(--color-bg)",
          borderBottom: "1px solid var(--color-border)",
          padding: "0 32px",
          position: "sticky",
          top: 0,
          zIndex: 10,
        }}
      >
        {/* Pill row */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 16,
            height: 60,
          }}
        >
          <TreKammareLogo />

          <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
            {navPills.map((pill) => (
              <PillNav
                key={pill.key}
                to={pill.to}
                label={pill.label}
                active={section === pill.key}
              />
            ))}
          </div>

          <div style={{ display: "flex", alignItems: "center", gap: 16, flexShrink: 0 }}>
            {/* Live indicator */}
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: 6,
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                letterSpacing: "0.08em",
                textTransform: "uppercase",
                color: "var(--color-fg-muted)",
              }}
            >
              <span
                className="animate-pulse-dot"
                style={{
                  width: 7,
                  height: 7,
                  borderRadius: "50%",
                  background: "var(--color-pulse)",
                  display: "inline-block",
                  flexShrink: 0,
                }}
              />
              LIVE · {liveDateStr()}
            </div>

            {/* Theme toggle */}
            <button
              onClick={toggleTheme}
              aria-label={theme === "dark" ? "Byt till ljust läge" : "Byt till mörkt läge"}
              style={{
                background: "transparent",
                border: "1px solid var(--color-border)",
                borderRadius: 999,
                width: 32,
                height: 32,
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
        </div>

        {/* ── Riksdag feature sub-tabs ────────────────────────────────── */}
        {showRiksdagTabs && (
          <div style={{ display: "flex", borderTop: "1px solid var(--color-border)" }}>
            {[
              { to: "/",          label: "Partimål & röstning", match: ["/", "/parties"] },
              { to: "/votes",     label: "Omröstningar",        match: []                },
              { to: "/budget",    label: "Statsbudget",         match: ["/budget"]       },
              { to: "/politicians", label: "Enskilda politiker",match: ["/politicians"]  },
              { to: "/manifestos",  label: "Manifest",          match: ["/manifestos"]  },
            ].map((tab) => {
              const active = tab.match.some(
                (m) => location.pathname === m || (m !== "/" && location.pathname.startsWith(m))
              );
              const isPartiesTab  = tab.to === "/" && location.pathname.startsWith("/parties");
              const isVotesTab    = tab.to === "/votes" && location.pathname === "/votes";
              const isVoteDetailOnParties = tab.to === "/" && location.pathname.startsWith("/votes/");
              const isActive = active || isPartiesTab || isVotesTab || isVoteDetailOnParties;

              return (
                <NavLink
                  key={tab.to}
                  to={tab.to}
                  style={{
                    display: "block",
                    padding: "10px 20px",
                    fontSize: 13,
                    fontWeight: isActive ? 700 : 500,
                    textDecoration: "none",
                    borderBottom: isActive
                      ? "2px solid var(--color-fg)"
                      : "2px solid transparent",
                    color: isActive ? "var(--color-fg)" : "var(--color-fg-muted)",
                    transition: "color 0.15s",
                    whiteSpace: "nowrap",
                  }}
                >
                  {tab.label}
                </NavLink>
              );
            })}
          </div>
        )}
      </nav>

      {/* ── Page content ──────────────────────────────────────────────── */}
      <main style={{ maxWidth: 1280, margin: "0 auto", padding: isMobile ? "0" : "32px" }}>
        <Routes>
          <Route path="/"                                      element={<HomePage />} />
          <Route path="/riksdag"                               element={<RiksdagPage />} />
          <Route path="/riksdag/myndigheter/:slug"             element={<AuthorityDetailPage />} />
          <Route path="/parties"                               element={<PartiesPage />} />
          <Route path="/parties/:party"                        element={<PartyDetailPage />} />
          <Route path="/parties/:party/goals"                  element={<PartyGoalsPage />} />
          <Route path="/parties/:party/goals/:goalId/votes"    element={<GoalVotesPage />} />
          <Route path="/politicians"                           element={<PoliticiansPage />} />
          <Route path="/politicians/:id"                       element={<PoliticianPage />} />
          <Route path="/votes"                                 element={<VotesPage />} />
          <Route path="/votes/:beteckning/:punkt"              element={<VoteDetailPage />} />
          <Route path="/beslut/:beteckning"                    element={<BeslutDetailPage />} />
          <Route path="/budget"                                element={<BudgetPage />} />
          <Route path="/budget/areas/:code"                    element={<AreaHistoryPage />} />
          <Route path="/manifestos"                            element={<ManifestosPage />} />
          <Route path="/region"                                element={<RegionLandingPage />} />
          <Route path="/region/:code"                          element={<RegionDetailPage />} />
          <Route path="/kommun"                                element={<MunicipalityLandingPage />} />
          <Route path="/kommun/jämför"                         element={<MunicipalityComparePage />} />
          <Route path="/kommun/:code"                          element={<MunicipalityDetailPage />} />
          <Route path="/sok"                                   element={<SearchPage />} />
          <Route path="/anforanden/:id"                        element={<SpeechDetailPage />} />
          <Route path="/data"                                  element={<DataIndexPage />} />
          <Route path="/data/:id"                              element={<DataSourcePage />} />
        </Routes>
      </main>

      <footer
        style={{
          borderTop: "1px solid var(--color-border)",
          padding: "16px 14px",
          textAlign: "center",
        }}
      >
        <Link
          to="/data"
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-fg-muted)",
            textDecoration: "none",
            letterSpacing: "0.1em",
            textTransform: "uppercase",
          }}
        >
          Datakällor
        </Link>
      </footer>
    </div>
  );
}
