import { useState } from "react";
import { NavLink } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { PartyBadge } from "@/shared/components";
import { municipalitiesApi } from "./api";
import { SwedenKommunMap } from "./components/SwedenKommunMap";
import type { MunicipalitySummary } from "@/shared/types";

function MunicipalityCard({ mun }: { mun: MunicipalitySummary }) {
  return (
    <NavLink
      to={`/kommun/${mun.code}`}
      className="block rounded-xl p-4 border transition-all hover:shadow-ambient"
      style={{ background: "var(--color-surface-lowest)", borderColor: "var(--color-surface-high)" }}
    >
      <div className="flex items-start justify-between mb-2">
        <div>
          <h3 className="font-display text-sm font-extrabold text-on-surface leading-tight">
            {mun.name}
          </h3>
          <p className="text-[11px] text-on-surface-variant mt-0.5">{mun.regionName}</p>
        </div>
        <span className="text-[10px] text-on-surface-variant font-mono bg-surface-high rounded px-1.5 py-0.5">
          {mun.code}
        </span>
      </div>

      <div className="flex gap-1 flex-wrap mb-3">
        {mun.governingParties.map((p) => (
          <PartyBadge key={p} party={p} />
        ))}
      </div>

      <div className="flex gap-4">
        <div>
          <div className="text-[11px] font-mono font-bold text-on-surface">
            {mun.population.toLocaleString("sv-SE")}
          </div>
          <div className="text-[9px] text-on-surface-variant uppercase tracking-widest">inv.</div>
        </div>
        <div>
          <div className="text-[11px] font-mono font-bold text-on-surface">
            {mun.totalMandates}
          </div>
          <div className="text-[9px] text-on-surface-variant uppercase tracking-widest">mandat</div>
        </div>
      </div>
    </NavLink>
  );
}

export function MunicipalityLandingPage() {
  const [search, setSearch] = useState("");

  const { data: municipalities, isLoading, error } = useQuery({
    queryKey: ["municipalities"],
    queryFn: () => municipalitiesApi.listMunicipalities(),
    staleTime: 60 * 60 * 1000,
  });

  const filtered = municipalities?.filter(
    (m) =>
      search === "" ||
      m.name.toLowerCase().includes(search.toLowerCase()) ||
      m.regionName.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div>
      <div className="mb-6">
        <h2 className="font-display text-2xl font-extrabold tracking-tight text-on-surface mb-1">
          Kommunalkollen
        </h2>
        <p className="text-sm text-on-surface-variant">
          Utforska valresultat och styrande partier i Sveriges kommuner.
          Data: Valmyndigheten, val 2022.
        </p>
      </div>

      {/* Search */}
      <div className="mb-5">
        <input
          type="search"
          placeholder="Sök kommun eller region…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full max-w-sm px-3 py-2 text-sm rounded-lg border bg-surface-lowest text-on-surface placeholder:text-on-surface-variant focus:outline-none focus:ring-1"
          style={{ borderColor: "var(--color-surface-high)" }}
        />
        {municipalities && (
          <span className="ml-3 text-xs text-on-surface-variant">
            {filtered?.length ?? 0} av {municipalities.length} kommuner
          </span>
        )}
      </div>

      {isLoading && (
        <p className="text-sm text-on-surface-variant py-16 text-center">Laddar kommuner…</p>
      )}
      {error && (
        <p className="text-sm text-red-500 py-8 text-center">Kunde inte hämta kommundata.</p>
      )}

      {filtered && municipalities && (
        <div className="flex flex-col lg:flex-row gap-8">
          <div className="shrink-0 lg:w-[260px]">
            <p className="text-[10px] text-on-surface-variant uppercase tracking-widest font-semibold mb-3 text-center lg:text-left">
              Klicka på en kommun
            </p>
            <SwedenKommunMap municipalities={municipalities} />
            <p className="mt-3 text-[10px] text-on-surface-variant text-center lg:text-left leading-snug">
              Karta: Lokal_Profil / Wikimedia Commons (CC BY-SA 2.5),
              grunddata från SCB.
            </p>
          </div>

          <div className="flex-1 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3 gap-3">
            {filtered.map((mun) => (
              <MunicipalityCard key={mun.code} mun={mun} />
            ))}
            {filtered.length === 0 && (
              <p className="col-span-3 text-sm text-on-surface-variant text-center py-8">
                Inga kommuner matchade sökningen.
              </p>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
