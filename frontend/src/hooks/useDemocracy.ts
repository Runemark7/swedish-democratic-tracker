import { useQuery } from "@tanstack/react-query";
import { regionsApi } from "@/features/regions/api";
import { municipalitiesApi } from "@/features/municipalities/api";
import { TC_PARTY_COLORS, type LevelData, type Party } from "@/types/democracy";
import { mockRiksdag, mockRegion, mockKommun } from "@/mock/democracy";
import type { ElectionResult, RegionSummary, MunicipalitySummary } from "@/shared/types";

// Map SCB election results to the Party shape used in LevelData
function electionResultsToParties(results: ElectionResult[]): {
  governing: Party[];
  opposition: Party[];
} {
  const governing: Party[] = [];
  const opposition: Party[] = [];

  // Sort by mandates descending so the largest party leads
  const sorted = [...results].sort((a, b) => b.mandates - a.mandates);

  // We don't have governing/opposition info from the API — fall back to grouping
  // TODO: expose governing coalition from backend so we can distinguish properly
  sorted.forEach((r) => {
    const color = TC_PARTY_COLORS[r.party] ?? "#888888";
    const party: Party = { name: r.party, short: r.party, seats: r.mandates, color };
    // Treat top-mandate party + its allies as governing (rough heuristic — real data via TODO above)
    if (governing.length === 0 || governing.reduce((s, p) => s + p.seats, 0) < r.totalMandates / 2) {
      governing.push(party);
    } else {
      opposition.push(party);
    }
  });

  return { governing, opposition };
}

// ── Riksdag ──────────────────────────────────────────────────────────────────
// TODO: replace with real /api/riksdag endpoint when implemented
export function useRiksdag() {
  return useQuery<LevelData>({
    queryKey: ["riksdag"],
    queryFn: async () => mockRiksdag,
    staleTime: 30_000,
  });
}

// ── Region list ───────────────────────────────────────────────────────────────
export function useRegionList() {
  return useQuery<RegionSummary[]>({
    queryKey: ["regions"],
    queryFn: () => regionsApi.listRegions(),
    staleTime: 30_000,
  });
}

// ── Single region ─────────────────────────────────────────────────────────────
export function useRegion(code: string) {
  return useQuery<LevelData>({
    queryKey: ["region", code],
    queryFn: async () => {
      const detail = await regionsApi.getRegion(code);
      const { governing, opposition } = electionResultsToParties(detail.electionResults ?? []);

      return {
        title: detail.name,
        subtitle: `Regional nivå — hälso- och sjukvård, kollektivtrafik`,
        population: detail.population
          ? detail.population.toLocaleString("sv-SE") + " invånare"
          : undefined,
        ruling: {
          // TODO: expose ruling coalition type from backend
          type: mockRegion.ruling.type,
          parties: governing.length ? governing : mockRegion.ruling.parties,
          opposition: opposition.length ? opposition : mockRegion.ruling.opposition,
        },
        // TODO: replace with real /api/regions/:code/votes when implemented
        liveVotes: mockRegion.liveVotes,
        // TODO: replace with real /api/regions/:code/budget when implemented
        budget: mockRegion.budget,
        // TODO: replace with real /api/regions/:code/agenda when implemented
        agenda: mockRegion.agenda,
        // TODO: replace with real region KPI endpoint when implemented
        kpis: mockRegion.kpis,
      } satisfies LevelData;
    },
    staleTime: 30_000,
    enabled: !!code,
  });
}

// ── Kommun list ───────────────────────────────────────────────────────────────
export function useKommunList(regionCode?: string) {
  return useQuery<MunicipalitySummary[]>({
    queryKey: ["kommuner", regionCode],
    queryFn: () => municipalitiesApi.listMunicipalities(regionCode),
    staleTime: 30_000,
  });
}

// ── Single kommun ─────────────────────────────────────────────────────────────
export function useKommun(code: string) {
  return useQuery<LevelData>({
    queryKey: ["kommun", code],
    queryFn: async () => {
      const [detail, kpiItems] = await Promise.all([
        municipalitiesApi.getMunicipality(code),
        municipalitiesApi.getMunicipalityKPIs(code).catch(() => []),
      ]);

      const { governing, opposition } = electionResultsToParties(detail.electionResults ?? []);

      // Map real KPI items if available — target/worseHigher not exposed yet, fall back to mock shape
      // TODO: expose target + worseHigher from KPI endpoint so TargetBar renders real goals
      const kpis = kpiItems.length > 0
        ? mockKommun.kpis  // keep mock until backend exposes targets
        : mockKommun.kpis;

      return {
        title: detail.name,
        subtitle: `Kommunal nivå — skola, omsorg, gator, plan`,
        population: detail.population
          ? detail.population.toLocaleString("sv-SE") + " invånare"
          : undefined,
        ruling: {
          // TODO: expose ruling coalition type from backend
          type: mockKommun.ruling.type,
          parties: governing.length ? governing : mockKommun.ruling.parties,
          opposition: opposition.length ? opposition : mockKommun.ruling.opposition,
        },
        // TODO: replace with real /api/municipalities/:code/votes when implemented
        liveVotes: mockKommun.liveVotes,
        // TODO: replace with real /api/municipalities/:code/budget when implemented
        budget: mockKommun.budget,
        // TODO: replace with real /api/municipalities/:code/agenda when implemented
        agenda: mockKommun.agenda,
        kpis,
      } satisfies LevelData;
    },
    staleTime: 30_000,
    enabled: !!code,
  });
}
