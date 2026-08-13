import { Routes, Route, NavLink, useLocation, Link, Navigate, useParams } from "react-router-dom";
import { PartiesPage } from "./features/parties/PartiesPage";
import { PartyDetailPage } from "./features/parties/PartyDetailPage";
import { GoalVotesPage } from "./features/parties/GoalVotesPage";
import { CommitteePage } from "./features/committees/CommitteePage";
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
import { RegionBudgetAreaPage } from "./features/regions/RegionBudgetAreaPage";
import { RegionKpiRankingPage } from "./features/regions/RegionKpiRankingPage";
import { KommunBudgetAreaPage } from "./features/municipalities/KommunBudgetAreaPage";
import { KommunKpiRankingPage } from "./features/municipalities/KommunKpiRankingPage";
import { MunicipalityLandingPage } from "./features/municipalities/MunicipalityLandingPage";
import { MunicipalityDetailPage } from "./features/municipalities/MunicipalityDetailPage";
import { MunicipalityComparePage } from "./features/municipalities/MunicipalityComparePage";
import { HomePage } from "./features/home/HomePage";
import { RiksdagPage } from "./features/riksdag/RiksdagPage";
import { AuthorityDetailPage } from "./features/riksdag/AuthorityDetailPage";
import { MyndigheterListPage } from "./features/riksdag/MyndigheterListPage";
import { RegeringPage } from "./features/regering/RegeringPage";
import { MinisterDetailPage } from "./features/regering/MinisterDetailPage";
import { SearchPage } from "./features/search/SearchPage";
import { SpeechDetailPage } from "./features/speeches/SpeechDetailPage";
import { DebateDetailPage } from "./features/speeches/DebateDetailPage";
import { DataIndexPage } from "./features/data/DataIndexPage";
import { OmSajtenPage } from "./features/about/OmSajtenPage";
import { DataSourcePage } from "./features/data/DataSourcePage";
import { useTheme } from "./contexts/theme";
import { useMediaQuery } from "./hooks/useMediaQuery";
import { useRecordCoverage } from "./hooks/useDemocracy";
import { swedishDate } from "./shared/dates";
import { MobileNav } from "./components/MobileNav";

// ── Section detection ─────────────────────────────────────────────────────
type NavSection = "start" | "riksdag" | "region" | "kommun" | "regering" | "sok";

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
  if (pathname.startsWith("/regering")) return "regering";
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
/**
 * /parties/:party/goals used to be a separate full record view. That view now
 * lives in the party page's Mål tab, so the old URL redirects rather than 404s —
 * it has been linked and shared.
 *
 * The redirect is absolute on purpose. A relative <Navigate to=".." /> resolves
 * against the route hierarchy, and because this route is declared flat that
 * lands on /parties, the list, rather than the party the reader asked for.
 */
function GoalsRedirect() {
  const { party = "" } = useParams<{ party: string }>();
  return <Navigate to={`/parties/${party}`} replace />;
}

export default function App() {
  const location = useLocation();
  const { theme, toggleTheme } = useTheme();
  // These must sit above the /dela early return: a hook called after a
  // conditional return changes the hook count between renders, so any
  // client-side navigation to or from a share page would throw. Today /dela is
  // only ever opened as a fresh page load (VoteDetailPage builds it as an
  // absolute URL), which is the only reason this has not bitten yet.
  const isMobile = useMediaQuery("(max-width: 640px)");
  const { data: coverage } = useRecordCoverage();

  // Share-only pages render without app chrome
  if (location.pathname.endsWith("/dela")) {
    return (
      <Routes>
        <Route path="/votes/:beteckning/:punkt/dela" element={<VoteSharePage />} />
      </Routes>
    );
  }

  const section = sectionFromPath(location.pathname);
  const showRiksdagTabs = isRiksdagSection(location.pathname);

  const navPills = [
    { key: "start",    to: "/",         label: "◆ START"     },
    { key: "riksdag",  to: "/riksdag",  label: "I RIKSDAG"   },
    { key: "region",   to: "/region",  label: "II REGION"   },
    { key: "kommun",   to: "/kommun",  label: "III KOMMUN"  },
    { key: "regering", to: "/regering", label: "IV REGERING" },
    { key: "sok",      to: "/sok",     label: "⌕ SÖK"       },
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
              {coverage?.lastDecisionDate
                ? `Senaste beslut · ${swedishDate(coverage.lastDecisionDate)}`
                : "Ingen omröstningsdata"}
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
              { to: "/parties",     label: "Partimål & röstning", match: ["/parties"]     },
              { to: "/votes",       label: "Omröstningar",        match: ["/votes"]       },
              { to: "/budget",      label: "Statsbudget",         match: ["/budget"]      },
              { to: "/politicians", label: "Enskilda politiker",  match: ["/politicians"] },
              { to: "/manifestos",  label: "Manifest",            match: ["/manifestos"]  },
            ].map((tab) => {
              // Exact match, or a nested route beneath it (/parties/S/goals).
              // The trailing slash keeps /votes from matching a sibling like
              // /votes-archive.
              const isActive = tab.match.some(
                (m) => location.pathname === m || location.pathname.startsWith(m + "/")
              );

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
          <Route path="/regering"     element={<RegeringPage />} />
          <Route path="/regering/:id" element={<MinisterDetailPage />} />
          <Route path="/riksdag"                               element={<RiksdagPage />} />
          <Route path="/riksdag/myndigheter"                   element={<MyndigheterListPage />} />
          <Route path="/riksdag/myndigheter/:slug"             element={<AuthorityDetailPage />} />
          <Route path="/parties"                               element={<PartiesPage />} />
          <Route path="/parties/:party"                        element={<PartyDetailPage />} />
          <Route path="/committees/:code"                      element={<CommitteePage />} />
          <Route path="/parties/:party/goals"                  element={<GoalsRedirect />} />
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
          <Route path="/region/:code/budget/:areaName"         element={<RegionBudgetAreaPage />} />
          <Route path="/region/:code/kpi/:kpiId"               element={<RegionKpiRankingPage />} />
          <Route path="/region/:code"                          element={<RegionDetailPage />} />
          <Route path="/kommun/:code/budget/:areaName"         element={<KommunBudgetAreaPage />} />
          <Route path="/kommun/:code/kpi/:kpiId"              element={<KommunKpiRankingPage />} />
          <Route path="/kommun"                                element={<MunicipalityLandingPage />} />
          <Route path="/kommun/jämför"                         element={<MunicipalityComparePage />} />
          <Route path="/kommun/:code"                          element={<MunicipalityDetailPage />} />
          <Route path="/sok"                                   element={<SearchPage />} />
          <Route path="/anforanden/:id"                        element={<SpeechDetailPage />} />
          <Route path="/debatt/:dokId"                         element={<DebateDetailPage />} />
          <Route path="/om-sajten"                             element={<OmSajtenPage />} />
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
        <div style={{ display: "flex", gap: 20, justifyContent: "center", flexWrap: "wrap" }}>
          {[
            { to: "/om-sajten", label: "Om sajten" },
            { to: "/data", label: "Datakällor" },
          ].map((l) => (
            <Link
              key={l.to}
              to={l.to}
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                color: "var(--color-fg-muted)",
                textDecoration: "none",
                letterSpacing: "0.1em",
                textTransform: "uppercase",
              }}
            >
              {l.label}
            </Link>
          ))}
        </div>
      </footer>
    </div>
  );
}
