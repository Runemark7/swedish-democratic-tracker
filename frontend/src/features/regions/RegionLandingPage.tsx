import { NavLink } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { PartyBadge } from "@/shared/components";
import { regionsApi } from "./api";
import { SwedenRegionMap } from "./components/SwedenRegionMap";
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

export function RegionLandingPage() {
  const { data: regions, isLoading, error } = useQuery({
    queryKey: ["regions"],
    queryFn: regionsApi.listRegions,
    staleTime: 60 * 60 * 1000,
  });

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
        <div className="flex flex-col md:flex-row gap-8">
          {/* Map sidebar */}
          <div className="shrink-0 md:w-[280px]">
            <p className="text-[10px] text-on-surface-variant uppercase tracking-widest font-semibold mb-3 text-center md:text-left">
              Klicka på en region
            </p>
            <SwedenRegionMap regions={regions} />
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
