import { NavLink } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { PartyBadge } from "@/shared/components";
import { regionsApi } from "./api";
import type { RegionSummary } from "@/shared/types";

function RegionCard({ region }: { region: RegionSummary }) {
  return (
    <NavLink
      to={`/region/${region.code}`}
      className="block rounded-xl p-4 border transition-all hover:shadow-ambient"
      style={{ background: "var(--color-surface-lowest)", borderColor: "var(--color-surface-high)" }}
    >
      <div className="flex items-start justify-between mb-2">
        <div>
          <h3 className="font-display text-sm font-extrabold text-on-surface leading-tight">
            {region.name}
          </h3>
          <p className="text-[11px] text-on-surface-variant mt-0.5">{region.capital}</p>
        </div>
        <span className="text-[10px] text-on-surface-variant font-mono bg-surface-high rounded px-1.5 py-0.5">
          {region.code}
        </span>
      </div>

      <div className="flex gap-1 flex-wrap mb-3">
        {region.governingParties.map((p) => (
          <PartyBadge key={p} party={p} />
        ))}
      </div>

      <div className="flex gap-4">
        <div>
          <div className="text-[11px] font-mono font-bold text-on-surface">
            {region.population.toLocaleString("sv-SE")}
          </div>
          <div className="text-[9px] text-on-surface-variant uppercase tracking-widest">invånare</div>
        </div>
        <div>
          <div className="text-[11px] font-mono font-bold text-on-surface">
            {region.totalMandates}
          </div>
          <div className="text-[9px] text-on-surface-variant uppercase tracking-widest">mandat</div>
        </div>
        <div>
          <div className="text-[11px] font-mono font-bold text-on-surface">{region.electionYear}</div>
          <div className="text-[9px] text-on-surface-variant uppercase tracking-widest">val</div>
        </div>
      </div>
    </NavLink>
  );
}

// Minimal interactive SVG map of Sweden showing the 21 regions.
// Each region is a simplified clickable hotspot. The proportions are
// approximate — accurate enough to orient the user without a full GIS tile.
function SwedenRegionMap({ regions, onSelect }: {
  regions: RegionSummary[];
  onSelect: (code: string) => void;
}) {
  // Approximate centroids for each region (SVG coords, 0-200 x 0-500 space)
  const centroids: Record<string, [number, number]> = {
    "25": [100, 40],  // Norrbotten
    "24": [100, 90],  // Västerbotten
    "22": [110, 135], // Västernorrland
    "23": [75, 140],  // Jämtland
    "21": [120, 180], // Gävleborg
    "20": [90, 200],  // Dalarna
    "17": [65, 220],  // Värmland
    "19": [120, 215], // Västmanland
    "18": [105, 240], // Örebro
    "03": [130, 250], // Uppsala
    "01": [145, 265], // Stockholm
    "04": [130, 280], // Södermanland
    "05": [125, 295], // Östergötland
    "14": [70, 310],  // Västra Götaland
    "06": [100, 320], // Jönköping
    "07": [115, 335], // Kronoberg
    "08": [130, 350], // Kalmar
    "13": [80, 355],  // Halland
    "12": [95, 375],  // Skåne
    "10": [130, 370], // Blekinge
    "09": [155, 330], // Gotland
  };

  return (
    <div className="relative mx-auto" style={{ width: 200, height: 500 }}>
      <svg
        viewBox="0 0 200 500"
        width="200"
        height="500"
        style={{ overflow: "visible" }}
        aria-label="Karta över Sveriges regioner"
      >
        {/* Very rough Sweden outline */}
        <path
          d="M120,10 L140,20 L155,40 L160,65 L150,90 L155,110 L160,130
             L150,155 L140,170 L145,195 L155,215 L165,240 L160,265
             L170,285 L165,305 L155,320 L160,340 L150,360 L135,380
             L115,395 L95,400 L75,390 L60,370 L55,350 L60,330
             L50,310 L45,290 L55,270 L50,250 L45,230 L55,210
             L50,195 L60,175 L55,155 L60,130 L65,110 L60,85
             L70,60 L80,40 L95,25 Z"
          fill="var(--color-surface-low)"
          stroke="var(--color-surface-high)"
          strokeWidth="1.5"
        />
        {/* Region dots */}
        {regions.map((r) => {
          const pos = centroids[r.code];
          if (!pos) return null;
          return (
            <g
              key={r.code}
              onClick={() => onSelect(r.code)}
              style={{ cursor: "pointer" }}
              role="button"
              aria-label={r.name}
            >
              <circle
                cx={pos[0]}
                cy={pos[1]}
                r={7}
                fill="var(--color-primary)"
                opacity={0.7}
                className="transition-all"
              />
              <circle
                cx={pos[0]}
                cy={pos[1]}
                r={7}
                fill="transparent"
                stroke="var(--color-on-surface)"
                strokeWidth={0}
                className="hover:stroke-2 transition-all"
              />
            </g>
          );
        })}
      </svg>
    </div>
  );
}

export function RegionLandingPage() {
  const { data: regions, isLoading, error } = useQuery({
    queryKey: ["regions"],
    queryFn: regionsApi.listRegions,
    staleTime: 60 * 60 * 1000,
  });

  function handleMapSelect(code: string) {
    const el = document.getElementById(`region-${code}`);
    el?.scrollIntoView({ behavior: "smooth", block: "center" });
  }

  return (
    <div>
      <div className="mb-6">
        <h2 className="font-display text-2xl font-extrabold tracking-tight text-on-surface mb-1">
          Regionkoll
        </h2>
        <p className="text-sm text-on-surface-variant">
          Utforska valresultat och styrande koalitioner i Sveriges 21 regioner.
          Data: Valmyndigheten, val 2022.
        </p>
      </div>

      {isLoading && (
        <p className="text-sm text-on-surface-variant py-16 text-center">Laddar regioner…</p>
      )}
      {error && (
        <p className="text-sm text-red-500 py-8 text-center">Kunde inte hämta regiondata.</p>
      )}

      {regions && (
        <div className="flex gap-8">
          {/* Map sidebar */}
          <div className="hidden md:block shrink-0">
            <p className="text-[10px] text-on-surface-variant uppercase tracking-widest font-semibold mb-3">
              Klicka på en region
            </p>
            <SwedenRegionMap regions={regions} onSelect={handleMapSelect} />
          </div>

          {/* Region cards grid */}
          <div className="flex-1 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {regions.map((region) => (
              <div key={region.code} id={`region-${region.code}`}>
                <RegionCard region={region} />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
