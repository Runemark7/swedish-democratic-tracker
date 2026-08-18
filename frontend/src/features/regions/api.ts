import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

type KPIRank = components["schemas"]["KPIRank"];
type MunicipalityKPIItem = components["schemas"]["MunicipalityKPIItem"];
type RegionAreaDataPoint = components["schemas"]["RegionAreaDataPoint"];
type RegionBudgetArea = components["schemas"]["RegionBudgetArea"];
type RegionBudgetSnapshot = components["schemas"]["RegionBudgetSnapshot"];
type RegionDetail = components["schemas"]["RegionDetail"];
type RegionKPIRankEntry = components["schemas"]["RegionKPIRankEntry"];
type RegionPlan = components["schemas"]["RegionPlan"];
type RegionSummary = components["schemas"]["RegionSummary"];

export const regionsApi = {
  listRegions: () => api.get<RegionSummary[]>("/regions"),
  getRegion: (code: string) => api.get<RegionDetail>(`/regions/${code}`),
  getRegionPlan: (code: string) => api.get<RegionPlan>(`/regions/${code}/plan`),
  getRegionBudget: (code: string) => api.get<RegionBudgetArea[]>(`/regions/${code}/budget`),
  getRegionBudgetHistory: (code: string) =>
    api.get<RegionBudgetSnapshot[]>(`/regions/${code}/budget/history`),
  getRegionKPIs: (code: string) => api.get<MunicipalityKPIItem[]>(`/regions/${code}/kpi`),
  getAreaAcrossRegions: (areaName: string, year?: number) =>
    api.get<RegionAreaDataPoint[]>(
      `/regions/budget/area/${encodeURIComponent(areaName)}${year != null ? `?year=${year}` : ""}`
    ),
  getRegionKPIRanking: (kpiCode: string) =>
    api.get<RegionKPIRankEntry[]>(`/regions/kpi/${encodeURIComponent(kpiCode)}/ranking`),
  getRegionKPIRanks: (code: string) =>
    api.get<KPIRank[]>(`/regions/${code}/kpi-ranks`),
};
