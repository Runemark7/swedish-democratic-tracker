import { Routes, Route, NavLink, useLocation } from "react-router-dom";
import { PartiesPage } from "./features/parties/PartiesPage";
import { PartyGoalsPage } from "./features/parties/PartyGoalsPage";
import { GoalVotesPage } from "./features/parties/GoalVotesPage";
import { PoliticiansPage } from "./features/politicians/PoliticiansPage";
import { PoliticianPage } from "./features/politicians/PoliticianPage";
import { VotesPage } from "./features/votes/VotesPage";
import { VoteDetailPage } from "./features/votes/VoteDetailPage";
import { BudgetPage } from "./features/budget/BudgetPage";
import { AreaHistoryPage } from "./features/budget/AreaHistoryPage";
import { ManifestosPage } from "./features/manifestos/ManifestosPage";
import { RegionLandingPage } from "./features/regions/RegionLandingPage";
import { RegionDetailPage } from "./features/regions/RegionDetailPage";
import { MunicipalityLandingPage } from "./features/municipalities/MunicipalityLandingPage";
import { MunicipalityDetailPage } from "./features/municipalities/MunicipalityDetailPage";

const CATEGORIES = [
  {
    key: "riksdag" as const,
    to: "/",
    label: "Riksdag",
    sub: "Nationell nivå",
    description: "Partimål, voteringar och statsbudget",
  },
  {
    key: "region" as const,
    to: "/region",
    label: "Region",
    sub: "21 regioner",
    description: "Mandatfördelning och styrande koalitioner",
  },
  {
    key: "kommun" as const,
    to: "/kommun",
    label: "Kommun",
    sub: "290 kommuner",
    description: "Valresultat och styrande partier",
  },
] as const;

function HeroCategoryCards({ section }: { section: Section }) {
  return (
    <div className="grid grid-cols-3 gap-3 mt-5">
      {CATEGORIES.map((cat) => {
        const isActive = section === cat.key;
        return (
          <NavLink
            key={cat.key}
            to={cat.to}
            className="rounded-xl px-5 py-4 transition-all"
            style={{
              background: isActive ? "#1e1e1e" : "#151515",
              border: isActive ? "1px solid #444" : "1px solid #262626",
              boxShadow: isActive ? "0 0 0 1px rgba(255,255,255,0.06)" : "none",
            }}
          >
            <div className="flex items-start justify-between mb-2">
              <span
                className="text-xl font-extrabold tracking-tight"
                style={{ color: isActive ? "#fff" : "#888" }}
              >
                {cat.label}
              </span>
              {isActive && (
                <span className="text-[10px] font-semibold uppercase tracking-widest text-neutral-500 bg-neutral-800 rounded px-1.5 py-0.5">
                  Aktiv
                </span>
              )}
            </div>
            <div
              className="text-[11px] font-semibold uppercase tracking-wider mb-1"
              style={{ color: isActive ? "#666" : "#444" }}
            >
              {cat.sub}
            </div>
            <div
              className="text-[13px] leading-snug"
              style={{ color: isActive ? "#aaa" : "#555" }}
            >
              {cat.description}
            </div>
          </NavLink>
        );
      })}
    </div>
  );
}

type Section = "riksdag" | "region" | "kommun";

function sectionFromPath(pathname: string): Section {
  if (pathname.startsWith("/region")) return "region";
  if (pathname.startsWith("/kommun")) return "kommun";
  return "riksdag";
}

export default function App() {
  const location = useLocation();
  const section = sectionFromPath(location.pathname);

  return (
    <div className="min-h-screen bg-surface font-body">
      {/* ── Hero (dark) ─────────────────────────────────────────────── */}
      <div className="bg-[#0f0f0f] text-white pb-8 pt-9 px-8">
        <div className="mx-auto max-w-[960px]">
          <div className="flex items-baseline gap-3 mb-1">
            <NavLink to="/" className="hover:opacity-80 transition-opacity">
              <h1 className="font-display text-[28px] font-extrabold tracking-tight m-0">
                RIKSDAGSKOLLEN
              </h1>
            </NavLink>
          </div>
          <p className="text-sm text-neutral-500 mb-0">
            Följ demokratin på alla nivåer — från riksdag till din hemkommun.
          </p>
          <HeroCategoryCards section={section} />
        </div>
      </div>

      {/* ── Sticky Nav ──────────────────────────────────────────────── */}
      <div
        className="glass sticky top-0 z-10 px-8"
        style={{ borderBottom: "1px solid color-mix(in srgb, var(--color-outline-variant) 30%, transparent)" }}
      >
        <div className="mx-auto max-w-[960px]">
          {/* Feature tabs (Riksdag only) */}
          {section === "riksdag" && (
            <div className="flex">
              {[
                { to: "/", label: "Partimål & röstning", match: ["/", "/parties"] },
                { to: "/votes", label: "Omröstningar", match: [] },
                { to: "/budget", label: "Statsbudget", match: ["/budget"] },
                { to: "/politicians", label: "Enskilda politiker", match: ["/politicians"] },
                { to: "/manifestos", label: "Manifest", match: ["/manifestos"] },
              ].map((tab) => {
                const active = tab.match.some(
                  (m) => location.pathname === m || (m !== "/" && location.pathname.startsWith(m))
                );
                const isPartiesTab = tab.to === "/" && location.pathname.startsWith("/parties");
                const isVotesTab = tab.to === "/votes" && location.pathname === "/votes";
                const isVoteDetailOnParties = tab.to === "/" && location.pathname.startsWith("/votes/");
                const isActive = active || isPartiesTab || isVotesTab || isVoteDetailOnParties;
                return (
                  <NavLink
                    key={tab.to}
                    to={tab.to}
                    className="block px-5 py-3 text-[13px] font-medium transition-colors"
                    style={{
                      borderBottom: isActive ? "2px solid var(--color-on-surface)" : "2px solid transparent",
                      color: isActive ? "var(--color-on-surface)" : "var(--color-on-surface-variant)",
                      fontWeight: isActive ? 700 : 500,
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

      {/* ── Content ─────────────────────────────────────────────────── */}
      <main className="mx-auto max-w-[960px] px-8 py-6">
        <Routes>
          <Route path="/" element={<PartiesPage />} />
          <Route path="/parties" element={<PartiesPage />} />
          <Route path="/parties/:party/goals" element={<PartyGoalsPage />} />
          <Route path="/parties/:party/goals/:goalId/votes" element={<GoalVotesPage />} />
          <Route path="/politicians" element={<PoliticiansPage />} />
          <Route path="/politicians/:id" element={<PoliticianPage />} />
          <Route path="/votes" element={<VotesPage />} />
          <Route path="/votes/:beteckning/:punkt" element={<VoteDetailPage />} />
          <Route path="/budget" element={<BudgetPage />} />
          <Route path="/budget/areas/:code" element={<AreaHistoryPage />} />
          <Route path="/manifestos" element={<ManifestosPage />} />
          <Route path="/region" element={<RegionLandingPage />} />
          <Route path="/region/:code" element={<RegionDetailPage />} />
          <Route path="/kommun" element={<MunicipalityLandingPage />} />
          <Route path="/kommun/:code" element={<MunicipalityDetailPage />} />
        </Routes>

        {/* Source disclaimer */}
        <div className="mt-8 p-4 bg-surface-low rounded-lg">
          <p className="text-xs text-on-surface-variant leading-relaxed m-0">
            <strong>Datakälla:</strong>{" "}
            {section === "riksdag"
              ? "Sveriges riksdag (data.riksdagen.se). Partimål extraheras från valmanifest och partiprogram. Matchning mot voteringar sker automatiskt med manuell verifiering."
              : "Valmyndigheten (val.se) — val 2022."}
          </p>
        </div>
      </main>
    </div>
  );
}
