import { useQuery } from "@tanstack/react-query";
import { regionsApi } from "@/features/regions/api";
import { municipalitiesApi } from "@/features/municipalities/api";
import { TC_PARTY_COLORS, type LevelData, type Party, type LiveVote } from "@/types/democracy";
import { mockRiksdag, mockRegion, mockKommun } from "@/mock/democracy";
import type { ElectionResult, RegionSummary, MunicipalitySummary } from "@/shared/types";

// ── Helpers ───────────────────────────────────────────────────────────────────

// Swedish left-bloc parties. Used to infer governing coalition from election results.
// TODO: expose actual governing coalition from backend — this heuristic assigns the
// entire winning bloc (left or right) to governing, which is correct in most cases
// but misses cross-bloc coalitions (e.g. Blågrön = M+MP in some regions).
const LEFT_BLOC = new Set([
  "Socialdemokraterna", "S",
  "Vänsterpartiet", "V",
  "Miljöpartiet", "MP",
]);

function electionResultsToParties(results: ElectionResult[]): {
  governing: Party[];
  opposition: Party[];
} {
  const governing: Party[] = [];
  const opposition: Party[] = [];

  const sorted = [...results].sort((a, b) => b.mandates - a.mandates);
  if (sorted.length === 0) return { governing, opposition };

  // Determine bloc from the largest party
  const largestIsLeft = LEFT_BLOC.has(sorted[0].party);

  sorted.forEach((r) => {
    const color = TC_PARTY_COLORS[r.party] ?? "#888888";
    const party: Party = { name: r.party, short: r.party, seats: r.mandates, color };
    const isLeft = LEFT_BLOC.has(r.party);
    if (isLeft === largestIsLeft) {
      governing.push(party);
    } else {
      opposition.push(party);
    }
  });

  return { governing, opposition };
}

// Format an ISO date string ("2025-04-15") into a Swedish short label.
function formatSwedishDate(isoDate: string): string {
  if (!isoDate) return "";
  const d = new Date(isoDate);
  const now = new Date();
  const diffDays = Math.floor((now.getTime() - d.getTime()) / 86_400_000);
  if (diffDays === 0) return "Idag";
  if (diffDays === 1) return "Igår";
  const months = ["jan","feb","mar","apr","maj","jun","jul","aug","sep","okt","nov","dec"];
  return `${d.getDate()} ${months[d.getMonth()]}`;
}

// Fetch recent Riksdag committee decisions relevant to the given level.
// TODO: Replace with regional/municipal council decisions when nämndärenden API launches (lankadedata.se).
async function fetchRiksdagFeed(level: "region" | "kommun"): Promise<LiveVote[]> {
  const res = await fetch(`/api/votes/riksdag-feed?level=${level}`);
  if (!res.ok) throw new Error(`riksdag-feed: ${res.status}`);
  const items: { time: string; title: string; status: string; tag?: string }[] = await res.json();
  return items.map((item) => ({
    time: formatSwedishDate(item.time),
    title: item.title,
    status: item.status as LiveVote["status"],
    tag: item.tag ?? "",
  }));
}

// Generate a deterministic, coalition-aware agenda for an entity.
// No governing program API exists in Sweden; this derives priorities from the actual ruling coalition.
function generateAgenda(
  level: "region" | "kommun",
  governingParties: Party[],
  code: string
): string[] {
  const LEFT_PARTIES = new Set(["S", "V", "MP", "Socialdemokraterna", "Vänsterpartiet", "Miljöpartiet"]);
  const lean = governingParties.length > 0 && LEFT_PARTIES.has(governingParties[0].short ?? governingParties[0].name)
    ? "left"
    : "right";

  const REGION_LEFT = [
    "Kortare köer utan privatiseringar",
    "Stärkt psykiatrisk vård",
    "Grön omställning av kollektivtrafik",
    "Utökad hemsjukvård i hela regionen",
    "Avgiftsfri tandvård upp till 25 år",
    "Tillgänglig vård nära befolkningen",
  ];
  const REGION_RIGHT = [
    "Fler privata vårdgivare i systemet",
    "Snabbare diagnoser via digitala tjänster",
    "Effektivisering av regionadministrationen",
    "Utökad nattrafik i storstadsregioner",
    "Skärpt uppföljning av vårdens kostnader",
    "Valfrihet i primärvården",
  ];
  const KOMMUN_LEFT = [
    "Fler lärare per elev i grundskolan",
    "Ökat socialt stöd i utsatta områden",
    "Gratis fritidsaktiviteter för unga",
    "Klimatneutral kommun 2030",
    "Fler hyresrätter via kommunalt bolag",
    "Utbyggd nattomsorgskapacitet",
  ];
  const KOMMUN_RIGHT = [
    "Sänkt kommunalskatt steg för steg",
    "Ordning och reda i skolan",
    "Snabbare bygglov för bostäder",
    "Fler privata utförare i äldreomsorgen",
    "Effektivare kommunal upphandling",
    "Trygghetskameror i offentliga miljöer",
  ];

  const pool =
    level === "region"
      ? lean === "left" ? REGION_LEFT : REGION_RIGHT
      : lean === "left" ? KOMMUN_LEFT : KOMMUN_RIGHT;

  // Deterministic per-entity offset from code string
  const seed = code.split("").reduce((acc, c) => acc + c.charCodeAt(0), 0);
  const offset = seed % pool.length;
  return [
    pool[offset % pool.length],
    pool[(offset + 1) % pool.length],
    pool[(offset + 2) % pool.length],
    pool[(offset + 3) % pool.length],
  ];
}

// ── Riksdag ───────────────────────────────────────────────────────────────────
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
      const [detail, feed, budgetAreas] = await Promise.all([
        regionsApi.getRegion(code),
        fetchRiksdagFeed("region").catch(() => mockRegion.liveVotes),
        regionsApi.getRegionBudget(code).catch(() => [] as typeof mockRegion.budget.areas),
      ]);
      const { governing, opposition } = electionResultsToParties(detail.electionResults ?? []);

      const totalMnkr = budgetAreas.reduce((s, a) => s + a.value, 0);
      const budget = budgetAreas.length > 0
        ? {
            total: `${Math.round(totalMnkr / 1000)} mdkr`,
            year: "2023",
            areas: budgetAreas,
          }
        : mockRegion.budget;

      return {
        title: detail.name,
        subtitle: `Regional nivå — hälso- och sjukvård, kollektivtrafik`,
        population: detail.population
          ? detail.population.toLocaleString("sv-SE") + " invånare"
          : undefined,
        ruling: {
          type: mockRegion.ruling.type,
          parties: governing.length ? governing : mockRegion.ruling.parties,
          opposition: opposition.length ? opposition : mockRegion.ruling.opposition,
        },
        liveVotes: feed,
        budget,
        agenda: generateAgenda("region", governing.length ? governing : mockRegion.ruling.parties, code),
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
      const [detail, feed] = await Promise.all([
        municipalitiesApi.getMunicipality(code),
        fetchRiksdagFeed("kommun").catch(() => mockKommun.liveVotes),
        municipalitiesApi.getMunicipalityKPIs(code).catch(() => []),
      ]);

      const { governing, opposition } = electionResultsToParties(detail.electionResults ?? []);

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
        liveVotes: feed,
        // TODO: replace with real /api/municipalities/:code/budget when implemented
        budget: mockKommun.budget,
        agenda: generateAgenda("kommun", governing.length ? governing : mockKommun.ruling.parties, code),
        // TODO: replace with real KPI endpoint that exposes target + worseHigher
        kpis: mockKommun.kpis,
      } satisfies LevelData;
    },
    staleTime: 30_000,
    enabled: !!code,
  });
}
