import { useQuery } from "@tanstack/react-query";
import { regionsApi } from "@/features/regions/api";
import { municipalitiesApi } from "@/features/municipalities/api";
import { riksdagApi, budgetApi, type BudgetYearDetail, type RiksdagKpi, type RiksdagGovernment, type RiksdagAgendaItem, type RiksdagLiveVote, type RegisteredAuthorityList, type MyndigheterListFilter, type AuthorityStats } from "@/features/riksdag/api";
import { speechesApi, type Speech } from "@/features/speeches/api";
import { votesApi } from "@/features/votes/api";
import { TC_PARTY_COLORS, type LevelData, type Party, type LiveVote, type Budget, type BudgetArea, type Kpi, type AgendaItem, type Ruling } from "@/types/democracy";
import type {
  RecordCoverage, ElectionResult, KPIRank, RegionSummary, MunicipalitySummary, MunicipalityKPIItem, RiksdagDocument, RiksdagDocumentFull, RegionBudgetSnapshot, RegionAreaDataPoint, RegionKPIRankEntry, RegionPlan, MunicipalityBudgetSnapshot, MunicipalityAreaDataPoint, BudgetSnapshot } from "@/shared/types";
import { STRIP_KPI_META, STRIP_ORDER, type KpiMeta } from "@/features/municipalities/kpiMeta";
import { REGION_KPI_META, REGION_STRIP_ORDER } from "@/features/regions/regionKpiMeta";

const EMPTY_BUDGET: Budget = { total: "–", year: "–", areas: [] };
const EMPTY_RULING: Ruling = { type: "–", parties: [], opposition: [] };

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
  N45014: "Vatten & avlopp",
};

const KPI_ORDER = ["N11004","N15028","N17014","N20014","N30005","N07037","N09022","N05011","N45014"];

// Converts Kolada spending KPIs (kr/invånare) × population → Budget.
// Returns null when data is absent; caller uses EMPTY_BUDGET instead.
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
// STRIP_KPI_META and STRIP_ORDER are imported from kpiMeta.ts

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
      kpiId: code,
      label: m.label,
      description: m.description,
      value: m.format(latest.value),
      raw: latest.value,
      worseHigher: m.worseHigher,
      unit: m.unit,
      trend,
      delta: prev ? `${sign}${absDelta} ${m.unit}` : "–",
      note: `Källa: Kolada ${latest.year}`,
    } satisfies Kpi];
  });
}

// Region-level strip KPIs: imported from regionKpiMeta.ts
// (kept as const alias so the useRegion hook below can reference it locally)
const REGION_STRIP_KPI_META = REGION_KPI_META;

// ── Riksdag budget helpers ────────────────────────────────────────────────────

const UO_TO_GROUP: Record<string, string> = {
  UO4: "Rättsväsen",
  UO6: "Försvar",
  UO9: "Hälsovård",
  UO10: "Socialförsäkring", UO11: "Socialförsäkring", UO12: "Socialförsäkring",
  UO15: "Utbildning", UO16: "Utbildning",
  UO22: "Infrastruktur",
  UO26: "Statsskuld & räntor",
};

const BUDGET_GROUP_ORDER = [
  "Socialförsäkring", "Hälsovård", "Utbildning", "Försvar",
  "Rättsväsen", "Infrastruktur", "Statsskuld & räntor", "Övrigt",
];

function mapBudgetDetail(detail: BudgetYearDetail): Budget {
  const grouped = new Map<string, number>();
  for (const alloc of detail.allocations) {
    const group = UO_TO_GROUP[alloc.area.code] ?? "Övrigt";
    grouped.set(group, (grouped.get(group) ?? 0) + alloc.amountKsek);
  }
  const totalKsek = detail.totalKsek;
  const areas: BudgetArea[] = BUDGET_GROUP_ORDER.map(name => {
    const ksek = grouped.get(name) ?? 0;
    return {
      name,
      value: Math.round(ksek / 1_000_000),
      pct: totalKsek > 0 ? Math.round((ksek / totalKsek) * 1000) / 10 : 0,
    };
  });
  const mdkr = Math.round(totalKsek / 1_000_000);
  const sourceUrl = detail.documents?.[0]?.url;
  return {
    total: `${mdkr.toLocaleString("sv-SE")} mdkr`,
    year: String(detail.year),
    areas,
    sourceUrl,
  };
}

function mapApiKpis(apiKpis: RiksdagKpi[]): Kpi[] {
  return apiKpis.map(k => {
    const rawStr = k.raw % 1 === 0
      ? `${k.raw}${k.unit ? " " + k.unit : ""}`
      : `${k.raw.toFixed(1).replace(".", ",")}${k.unit ? " " + k.unit : ""}`;
    return {
      kpiId: String(k.id),
      label: k.label,
      description: k.description,
      value: rawStr,
      raw: k.raw,
      worseHigher: k.worseHigher,
      unit: k.unit,
      trend: k.trend,
      delta: k.delta,
      note: k.note,
      sourceUrl: k.sourceUrl,
    };
  });
}

function mapGovernment(gov: RiksdagGovernment): LevelData["ruling"] {
  const toParty = (p: RiksdagGovernment["parties"][number]): Party => ({
    name: p.name, short: p.short, seats: p.seats, color: p.color,
  });
  return {
    type: gov.typeLabel,
    parties: gov.parties.map(toParty),
    support: gov.support.map(toParty),
    opposition: gov.opposition.map(toParty),
  };
}

function mapAgenda(items: RiksdagAgendaItem[]): AgendaItem[] {
  return items.map(i => ({ id: i.id, title: i.title, description: i.description, source: i.source, status: i.status as AgendaItem["status"] }));
}

function mapLiveVotes(votes: RiksdagLiveVote[]): LiveVote[] {
  return votes.map(v => ({
    time: formatSwedishDate(v.date),
    title: v.title,
    status: v.status as LiveVote["status"],
    tag: v.tag,
    beteckning: v.beteckning,
  }));
}

// ── Riksdag ───────────────────────────────────────────────────────────────────
export function useRiksdag() {
  return useQuery<LevelData>({
    queryKey: ["riksdag"],
    queryFn: async () => {
      const [authorities, years, apiKpis, apiGov, apiAgenda, apiLiveVotes] = await Promise.all([
        riksdagApi.getAuthorities().catch(() => []),
        budgetApi.listYears().catch((): import("@/features/riksdag/api").BudgetYearSummary[] => []),
        riksdagApi.getKpis().catch(() => [] as RiksdagKpi[]),
        riksdagApi.getGovernment().catch(() => null as RiksdagGovernment | null),
        riksdagApi.getAgenda().catch(() => [] as RiksdagAgendaItem[]),
        riksdagApi.getLiveVotes().catch(() => [] as RiksdagLiveVote[]),
      ]);

      const latestDecided = [...(years ?? [])]
        .filter(y => y.status === "decided")
        .sort((a, b) => b.year - a.year)[0];

      const budget = latestDecided
        ? await budgetApi.getYear(latestDecided.year)
            .then(mapBudgetDetail)
            .catch(() => EMPTY_BUDGET)
        : EMPTY_BUDGET;

      const kpis = apiKpis.length > 0 ? mapApiKpis(apiKpis) : [];
      const ruling = apiGov ? mapGovernment(apiGov) : EMPTY_RULING;
      const agenda = apiAgenda.length > 0 ? mapAgenda(apiAgenda) : [];
      const liveVotes = apiLiveVotes.length > 0 ? mapLiveVotes(apiLiveVotes) : [];

      return {
        title: "Riksdagen",
        subtitle: "Sveriges nationella parlament — 349 ledamöter",
        authorities,
        budget,
        kpis,
        ruling,
        agenda,
        liveVotes,
      };
    },
    staleTime: 60_000,
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
        fetchRiksdagFeed("region").catch(() => []),
        regionsApi.getRegionBudget(code).catch(() => []),
        regionsApi.getRegionKPIs(code).catch(() => [] as MunicipalityKPIItem[]),
      ]);
      const { governing, opposition } = electionResultsToParties(detail.electionResults ?? []);

      const totalMnkr = budgetAreas.reduce((s, a) => s + a.value, 0);
      const budgetYear = String(new Date().getFullYear() - 1);
      const budget = budgetAreas.length > 0
        ? {
            total: `${Math.round(totalMnkr / 1000)} mdkr`,
            year: budgetYear,
            areas: budgetAreas,
            sourceUrl: "https://www.scb.se/hitta-statistik/statistik-efter-amne/offentlig-ekonomi/finanser-for-den-kommunala-sektorn/rakenskapssammandrag-for-kommuner-och-regioner/",
          }
        : EMPTY_BUDGET;

      return {
        title: detail.name,
        subtitle: `Regional nivå — hälso- och sjukvård, kollektivtrafik`,
        population: detail.population
          ? detail.population.toLocaleString("sv-SE") + " invånare"
          : undefined,
        ruling: {
          type: "Regionstyre",
          parties: governing,
          opposition,
        },
        liveVotes: feed,
        budget,
        kpis: kpiItemsToStrip(kpiItems, REGION_STRIP_KPI_META, REGION_STRIP_ORDER),
      } satisfies LevelData;
    },
    staleTime: 30_000,
    enabled: !!code,
  });
}

// ── Region budget history ─────────────────────────────────────────────────────
export function useRegionBudgetHistory(code: string) {
  return useQuery<RegionBudgetSnapshot[]>({
    queryKey: ["region-budget-history", code],
    queryFn: () => regionsApi.getRegionBudgetHistory(code),
    staleTime: 300_000,
    enabled: !!code,
  });
}

// ── Region plan (official annual-plan doc + verbatim overarching goals) ───────
export function useRegionPlan(code: string) {
  return useQuery<RegionPlan>({
    queryKey: ["region-plan", code],
    queryFn: () => regionsApi.getRegionPlan(code),
    staleTime: 60 * 60 * 1000,
    enabled: !!code,
  });
}

// ── Municipality budget history ───────────────────────────────────────────────
export function useKommunBudgetHistory(code: string) {
  return useQuery<MunicipalityBudgetSnapshot[]>({
    queryKey: ["kommun-budget-history", code],
    queryFn: () => municipalitiesApi.getMunicipalityBudgetHistory(code),
    staleTime: 300_000,
    enabled: !!code,
  });
}

// ── Myndigheter stats (aggregated coverage for list page header) ─────────────
export function useMyndigheterStats() {
  return useQuery<AuthorityStats>({
    queryKey: ["myndigheter-stats"],
    queryFn: () => riksdagApi.getAuthorityStats(),
    staleTime: 300_000,
  });
}

// ── Myndigheter list (DB-backed agency register, Phase 4) ────────────────────
export function useMyndigheterList(filter: MyndigheterListFilter) {
  return useQuery<RegisteredAuthorityList>({
    queryKey: ["myndigheter-list", filter],
    queryFn: () => riksdagApi.listMyndigheter(filter),
    staleTime: 60_000,
  });
}

// ── Riksdag budget history ────────────────────────────────────────────────────
export function useRiksdagBudgetHistory() {
  return useQuery<BudgetSnapshot[]>({
    queryKey: ["riksdag-budget-history"],
    queryFn: async () => {
      const years = await budgetApi.listYears();
      const decided = years.filter(y => y.status === "decided");
      const details = await Promise.all(decided.map(y => budgetApi.getYear(y.year)));
      const snapshots: BudgetSnapshot[] = [];
      for (const detail of details) {
        const total_mnkr = detail.totalKsek / 1000;
        for (const alloc of detail.allocations) {
          const value_mnkr = alloc.amountKsek / 1000;
          snapshots.push({
            area_name: alloc.area.name,
            year: detail.year,
            value_mnkr,
            total_mnkr,
            pct: total_mnkr > 0 ? (value_mnkr / total_mnkr) * 100 : 0,
          });
        }
      }
      return snapshots;
    },
    staleTime: 300_000,
  });
}

// ── Area across regions ───────────────────────────────────────────────────────
export function useAreaAcrossRegions(areaName: string, year?: number) {
  return useQuery<RegionAreaDataPoint[]>({
    queryKey: ["area-across-regions", areaName, year],
    queryFn: () => regionsApi.getAreaAcrossRegions(areaName, year),
    staleTime: 300_000,
    enabled: !!areaName,
  });
}

// ── Area across municipalities ────────────────────────────────────────────────
export function useAreaAcrossMunicipalities(areaName: string, year?: number) {
  return useQuery<MunicipalityAreaDataPoint[]>({
    queryKey: ["area-across-municipalities", areaName, year],
    queryFn: () => municipalitiesApi.getAreaAcrossMunicipalities(areaName, year),
    staleTime: 300_000,
    enabled: !!areaName,
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
        fetchRiksdagFeed("kommun").catch(() => []),
        municipalitiesApi.getMunicipalityKPIs(code).catch(() => [] as MunicipalityKPIItem[]),
        municipalitiesApi.getMunicipalitySpending(code).catch(() => [] as MunicipalityKPIItem[]),
      ]);

      const { governing, opposition } = electionResultsToParties(detail.electionResults ?? []);
      const govSeats   = governing.reduce((s, p) => s + p.seats, 0);
      const totalSeats = [...governing, ...opposition].reduce((s, p) => s + p.seats, 0);
      const isLeft     = governing.length > 0 && LEFT_BLOC.has(governing[0].name);
      const majority   = govSeats > totalSeats / 2 ? "majoritet" : "minoritet";
      const coalType   = governing.length === 0
        ? "–"
        : isLeft ? `Rödgrön ${majority}` : `Borgerlig ${majority}`;

      const budget = spendingToBudget(spendingItems, detail.population) ?? EMPTY_BUDGET;
      const kpis   = kpiItemsToStrip(kpiItems, STRIP_KPI_META, STRIP_ORDER);

      return {
        title: detail.name,
        subtitle: `Kommunal nivå — skola, omsorg, gator, plan`,
        population: detail.population
          ? detail.population.toLocaleString("sv-SE") + " invånare"
          : undefined,
        ruling: {
          type: coalType,
          parties: governing,
          opposition,
        },
        liveVotes: feed,
        budget,
        kpis,
      } satisfies LevelData;
    },
    staleTime: 30_000,
    enabled: !!code,
  });
}

// ── Record coverage ───────────────────────────────────────────────────────────
/** How much of the mandate record the site holds, and when its newest decision
 *  was taken. Both are stated in the UI rather than implied. */
export function useRecordCoverage() {
  return useQuery<RecordCoverage>({
    queryKey: ["record-coverage"],
    queryFn: () => fetch("/api/riksdag/coverage").then((r) => r.json()),
    staleTime: 5 * 60_000,
  });
}

// ── Recent speeches ───────────────────────────────────────────────────────────
export function useRecentSpeeches(limit = 100) {
  return useQuery<Speech[]>({
    queryKey: ["speeches-recent", limit],
    queryFn: () => speechesApi.listRecent(limit),
    staleTime: 60_000,
  });
}

export function useSpeechesByDocument(dokId: string | undefined) {
  return useQuery<Speech[]>({
    queryKey: ["speeches-by-document", dokId],
    queryFn: () => speechesApi.listByDocument(dokId ?? ""),
    enabled: !!dokId,
    staleTime: 60_000,
  });
}

export function useSpeechesByPolitician(
  intressentId: string | undefined,
  limit = 20,
) {
  return useQuery<Speech[]>({
    queryKey: ["speeches-by-politician", intressentId, limit],
    queryFn: () => speechesApi.listByPolitician(intressentId ?? "", limit),
    enabled: !!intressentId,
    staleTime: 60_000,
  });
}

export function useSpeechesByParty(
  party: string | undefined,
  limit = 50,
) {
  return useQuery<Speech[]>({
    queryKey: ["speeches-by-party", party, limit],
    queryFn: () => speechesApi.listByParty(party ?? "", limit),
    enabled: !!party,
    staleTime: 60_000,
  });
}

export function useDocument(dokId: string | undefined) {
  return useQuery<RiksdagDocument>({
    queryKey: ["document", dokId],
    queryFn: () => votesApi.getDocument(dokId ?? ""),
    enabled: !!dokId,
    staleTime: 5 * 60 * 1000,
  });
}

export function useDocumentFull(dokId: string | undefined) {
  return useQuery<RiksdagDocumentFull>({
    queryKey: ["document-full", dokId],
    queryFn: () => votesApi.getDocumentFull(dokId ?? ""),
    enabled: !!dokId,
    staleTime: 5 * 60 * 1000,
  });
}

// ── KPI ranking — all municipalities for one KPI ──────────────────────────────
export interface KPIRankEntry {
  mun_code: string;
  name: string;
  value: number;
  year: number;
  rank: number;
  total: number;
}

export function useKpiRanking(kpiCode: string) {
  return useQuery<KPIRankEntry[]>({
    queryKey: ["kpi-ranking", kpiCode],
    queryFn: () => municipalitiesApi.getKPIRanking(kpiCode),
    staleTime: 10 * 60 * 1000,
    enabled: !!kpiCode,
  });
}

// ── KPI ranks for one municipality ────────────────────────────────────────────
export type { KPIRank }; // re-export from @/shared/types for backward compat

export function useKommunKpiRanks(munCode: string) {
  return useQuery<KPIRank[]>({
    queryKey: ["kommun-kpi-ranks", munCode],
    queryFn: () => municipalitiesApi.getMunicipalityKPIRanks(munCode),
    staleTime: 10 * 60 * 1000,
    enabled: !!munCode,
  });
}

// ── Region KPI ranking ────────────────────────────────────────────────────────
export type { RegionKPIRankEntry }; // re-export from @/shared/types

export function useRegionKpiRanking(kpiCode: string) {
  return useQuery<RegionKPIRankEntry[]>({
    queryKey: ["region-kpi-ranking", kpiCode],
    queryFn: () => regionsApi.getRegionKPIRanking(kpiCode),
    staleTime: 10 * 60 * 1000,
    enabled: !!kpiCode,
  });
}

// ── Region KPI ranks for one region ──────────────────────────────────────────
export function useRegionKpiRanks(regionCode: string) {
  return useQuery<KPIRank[]>({
    queryKey: ["region-kpi-ranks", regionCode],
    queryFn: () => regionsApi.getRegionKPIRanks(regionCode),
    staleTime: 10 * 60 * 1000,
    enabled: !!regionCode,
  });
}
