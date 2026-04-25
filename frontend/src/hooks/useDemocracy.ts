import { useQuery } from "@tanstack/react-query";
import { regionsApi } from "@/features/regions/api";
import { municipalitiesApi } from "@/features/municipalities/api";
import { TC_PARTY_COLORS, type LevelData, type Party, type LiveVote, type Budget, type BudgetArea, type Kpi } from "@/types/democracy";
import { mockRiksdag, mockRegion, mockKommun } from "@/mock/democracy";
import type { ElectionResult, RegionSummary, MunicipalitySummary, MunicipalityKPIItem } from "@/shared/types";

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
  const items: { time: string; title: string; status: string; tag?: string; beteckning?: string }[] = await res.json();
  return items.map((item) => ({
    time: formatSwedishDate(item.time),
    title: item.title,
    status: item.status as LiveVote["status"],
    tag: item.tag ?? "",
    beteckning: item.beteckning,
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

// ── Municipality budget helpers ────────────────────────────────────────────────

const SPENDING_NAMES: Record<string, string> = {
  N11004: "Förskola",
  N15028: "Grundskola",
  N17014: "Gymnasieskola",
  N20014: "Äldreomsorg",
  N30005: "Individ & familj",
  N07037: "Gata, park, plan",
  N09022: "Fritid & kultur",
  N05011: "Politisk verksamhet",
};

const KPI_ORDER = ["N11004","N15028","N17014","N20014","N30005","N07037","N09022","N05011"];

// Converts Kolada spending KPIs (kr/invånare) × population → Budget.
// Returns null when data is absent so the caller can fall back to mock.
function spendingToBudget(items: MunicipalityKPIItem[], population: number): Budget | null {
  if (!items.length || !population) return null;

  // Pick the latest entry per KPI code
  const byKpi = new Map<string, MunicipalityKPIItem>();
  for (const item of items) {
    const existing = byKpi.get(item.kpi);
    if (!existing || item.year > existing.year) byKpi.set(item.kpi, item);
  }

  const areasKr: BudgetArea[] = [];
  let totalKr = 0;
  for (const code of KPI_ORDER) {
    const entry = byKpi.get(code);
    if (entry && entry.value > 0) {
      const kr = entry.value * population;
      areasKr.push({ name: SPENDING_NAMES[code] ?? code, value: kr, pct: 0 });
      totalKr += kr;
    }
  }
  if (!areasKr.length || totalKr === 0) return null;

  for (const a of areasKr) a.pct = Math.round((a.value / totalKr) * 1000) / 10;
  const areas: BudgetArea[] = areasKr.map(a => ({
    ...a,
    value: Math.round((a.value / 1e9) * 100) / 100,
  }));
  const totalMdkr = Math.round((totalKr / 1e9) * 10) / 10;
  const latestYear = Math.max(...[...byKpi.values()].map(i => i.year));

  return {
    total: `${totalMdkr.toLocaleString("sv-SE")} mdkr`,
    year: String(latestYear),
    areas,
  };
}

// ── Municipality KPI strip helpers ────────────────────────────────────────────

const STRIP_KPI_META: Record<string, {
  label: string;
  unit: string;
  target: number;
  worseHigher: boolean;
  format: (v: number) => string;
}> = {
  N00900: { label: "Kommunalskatt",  unit: "%", target: 31.0, worseHigher: true,  format: v => `${v.toFixed(2)} %` },
  N03102: { label: "Resultat/skatt", unit: "%", target:  2.0, worseHigher: false, format: v => `${v.toFixed(1)} %` },
  N03106: { label: "Soliditet",      unit: "%", target: 25.0, worseHigher: false, format: v => `${v.toFixed(0)} %` },
};
const STRIP_ORDER = ["N00900", "N03102", "N03106"];

type KpiMeta = {
  label: string;
  unit: string;
  target: number;
  worseHigher: boolean;
  format: (v: number) => string;
};

// Converts raw Kolada KPI items → Kpi[] for the header strip.
function kpiItemsToStrip(
  items: MunicipalityKPIItem[],
  meta: Record<string, KpiMeta>,
  order: string[],
): Kpi[] {
  const byKpi = new Map<string, MunicipalityKPIItem[]>();
  for (const item of items) {
    const list = byKpi.get(item.kpi) ?? [];
    list.push(item);
    byKpi.set(item.kpi, list);
  }

  return order.flatMap(code => {
    const m = meta[code];
    const yearItems = (byKpi.get(code) ?? []).sort((a, b) => b.year - a.year);
    if (!m || yearItems.length === 0) return [];

    const latest = yearItems[0];
    const prev   = yearItems[1];
    const delta  = prev ? latest.value - prev.value : 0;
    const trend: Kpi["trend"] = delta > 0.01 ? "up" : delta < -0.01 ? "down" : "flat";
    const sign   = delta >= 0 ? "+" : "−";
    const absDelta = Math.abs(delta).toFixed(2);

    return [{
      label: m.label,
      value: m.format(latest.value),
      raw: latest.value,
      target: m.target,
      worseHigher: m.worseHigher,
      unit: m.unit,
      trend,
      delta: prev ? `${sign}${absDelta} ${m.unit}` : "–",
      note: `Källa: Kolada ${latest.year}`,
    } satisfies Kpi];
  });
}

// Region-level strip KPIs: financial health metrics from Kolada's region (N6xxxx) namespace.
const REGION_STRIP_KPI_META: Record<string, KpiMeta> = {
  N60008: { label: "Nettokostnad/inv", unit: "kr", target: 40000, worseHigher: true,  format: v => `${Math.round(v).toLocaleString("sv-SE")} kr` },
  N63016: { label: "Resultat/skatt",   unit: "%",  target:   2.0, worseHigher: false, format: v => `${v.toFixed(1)} %` },
  N63007: { label: "Soliditet",        unit: "%",  target:  25.0, worseHigher: false, format: v => `${v.toFixed(0)} %` },
};
const REGION_STRIP_ORDER = ["N60008", "N63016", "N63007"];

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
      const [detail, feed, budgetAreas, kpiItems] = await Promise.all([
        regionsApi.getRegion(code),
        fetchRiksdagFeed("region").catch(() => mockRegion.liveVotes),
        regionsApi.getRegionBudget(code).catch(() => [] as typeof mockRegion.budget.areas),
        regionsApi.getRegionKPIs(code).catch(() => [] as MunicipalityKPIItem[]),
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
        kpis: kpiItemsToStrip(kpiItems, REGION_STRIP_KPI_META, REGION_STRIP_ORDER),
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
      const [detail, feed, kpiItems, spendingItems] = await Promise.all([
        municipalitiesApi.getMunicipality(code),
        fetchRiksdagFeed("kommun").catch(() => mockKommun.liveVotes),
        municipalitiesApi.getMunicipalityKPIs(code).catch(() => [] as MunicipalityKPIItem[]),
        municipalitiesApi.getMunicipalitySpending(code).catch(() => [] as MunicipalityKPIItem[]),
      ]);

      const { governing, opposition } = electionResultsToParties(detail.electionResults ?? []);
      const govSeats   = governing.reduce((s, p) => s + p.seats, 0);
      const totalSeats = [...governing, ...opposition].reduce((s, p) => s + p.seats, 0);
      const isLeft     = governing.length > 0 && LEFT_BLOC.has(governing[0].name);
      const majority   = govSeats > totalSeats / 2 ? "majoritet" : "minoritet";
      const coalType   = governing.length === 0
        ? mockKommun.ruling.type
        : isLeft ? `Rödgrön ${majority}` : `Borgerlig ${majority}`;

      const budget = spendingToBudget(spendingItems, detail.population) ?? mockKommun.budget;
      const kpis   = kpiItemsToStrip(kpiItems, STRIP_KPI_META, STRIP_ORDER);

      return {
        title: detail.name,
        subtitle: `Kommunal nivå — skola, omsorg, gator, plan`,
        population: detail.population
          ? detail.population.toLocaleString("sv-SE") + " invånare"
          : undefined,
        ruling: {
          type: coalType,
          parties: governing.length ? governing : mockKommun.ruling.parties,
          opposition: opposition.length ? opposition : mockKommun.ruling.opposition,
        },
        liveVotes: feed,
        budget,
        agenda: generateAgenda("kommun", governing.length ? governing : mockKommun.ruling.parties, code),
        kpis,
      } satisfies LevelData;
    },
    staleTime: 30_000,
    enabled: !!code,
  });
}
