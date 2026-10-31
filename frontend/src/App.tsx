import { useEffect } from "react";
import { Routes, Route, useLocation, Link, Navigate, useParams } from "react-router-dom";
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
import { LokaltPage } from "./features/lokalt/LokaltPage";
import { RiksdagPage } from "./features/riksdag/RiksdagPage";
import { AuthorityDetailPage } from "./features/riksdag/AuthorityDetailPage";
import { MyndigheterListPage } from "./features/riksdag/MyndigheterListPage";
import { RegeringPage } from "./features/regering/RegeringPage";
import { MinisterDetailPage } from "./features/regering/MinisterDetailPage";
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
import { Wordmark } from "./components/Wordmark";
import { LokaltMenu } from "./components/LokaltMenu";
import { PRIMARY_NAV, SECONDARY_NAV, documentTitle, isActive } from "./shared/nav";

// ── Nav pill ───────────────────────────────────────────────────────────────
interface PillNavProps {
  to: string;
  label: string;
  active: boolean;
}

function PillNav({ to, label, active }: PillNavProps) {
  return (
    <Link
      to={to}
      aria-current={active ? "page" : undefined}
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
    </Link>
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
  const pathname = location.pathname;

  useEffect(() => {
    document.title = documentTitle(pathname);
  }, [pathname]);

  // Share-only pages render without app chrome
  if (location.pathname.endsWith("/dela")) {
    return (
      <Routes>
        <Route path="/votes/:beteckning/:punkt/dela" element={<VoteSharePage />} />
      </Routes>
    );
  }

  return (
    <div className="sdt-page">
      {/* ── Mobile nav (≤640 px) ──────────────────────────────────────── */}
      <MobileNav
        pathname={pathname}
        theme={theme}
        toggleTheme={toggleTheme}
      />

      {/* ── Desktop top nav (>640 px) ─────────────────────────────────── */}
      <header
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
            minHeight: 60,
            padding: "8px 0",
            flexWrap: "wrap",
          }}
        >
          <Wordmark />

          <div style={{ display: "flex", alignItems: "center", gap: 6, flexWrap: "wrap", flex: 1 }}>
            <nav aria-label="Huvudmeny" style={{ display: "flex", alignItems: "center", gap: 6, flexWrap: "wrap" }}>
              {PRIMARY_NAV.map((item) => (
                <PillNav
                  key={item.to}
                  to={item.to}
                  label={item.label}
                  active={isActive(item, pathname)}
                />
              ))}
            </nav>
            <span
              aria-hidden="true"
              style={{
                width: 1,
                alignSelf: "stretch",
                margin: "6px 8px",
                background: "var(--color-border)",
              }}
            />
            <nav aria-label="Lokalt">
              <LokaltMenu pathname={pathname} />
            </nav>
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

            <nav aria-label="Mer" style={{ display: "flex", alignItems: "center", gap: 16 }}>
              {SECONDARY_NAV.map((item) => (
                <Link
                  key={item.to}
                  to={item.to}
                  aria-current={isActive(item, pathname) ? "page" : undefined}
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 11,
                    letterSpacing: "0.08em",
                    textTransform: "uppercase",
                    textDecoration: "none",
                    color: isActive(item, pathname) ? "var(--color-fg)" : "var(--color-fg-muted)",
                    whiteSpace: "nowrap",
                  }}
                >
                  {item.label}
                </Link>
              ))}
            </nav>

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
      </header>

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
          {/* The list moved to /beslut. Detail and share URLs stay. */}
          <Route path="/votes"                                 element={<Navigate to="/beslut" replace />} />
          <Route path="/votes/:beteckning/:punkt"              element={<VoteDetailPage />} />
          <Route path="/beslut"                                element={<VotesPage />} />
          <Route path="/beslut/:beteckning"                    element={<BeslutDetailPage />} />
          <Route path="/budget"                                element={<BudgetPage />} />
          <Route path="/budget/areas/:code"                    element={<AreaHistoryPage />} />
          <Route path="/manifestos"                            element={<ManifestosPage />} />
          <Route path="/lokalt"                               element={<LokaltPage />} />
          <Route path="/region"                                element={<RegionLandingPage />} />
          <Route path="/region/:code/budget/:areaName"         element={<RegionBudgetAreaPage />} />
          <Route path="/region/:code/kpi/:kpiId"               element={<RegionKpiRankingPage />} />
          <Route path="/region/:code"                          element={<RegionDetailPage />} />
          <Route path="/kommun/:code/budget/:areaName"         element={<KommunBudgetAreaPage />} />
          <Route path="/kommun/:code/kpi/:kpiId"              element={<KommunKpiRankingPage />} />
          <Route path="/kommun"                                element={<MunicipalityLandingPage />} />
          <Route path="/kommun/jämför"                         element={<MunicipalityComparePage />} />
          <Route path="/kommun/:code"                          element={<MunicipalityDetailPage />} />
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
          {SECONDARY_NAV.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                color: "var(--color-fg-muted)",
                textDecoration: "none",
                letterSpacing: "0.1em",
                textTransform: "uppercase",
              }}
            >
              {item.label}
            </Link>
          ))}
        </div>
      </footer>
    </div>
  );
}
