import { api } from "@/shared/api-client";
import type { KPIRank, MunicipalityKPIItem, RegionAreaDataPoint, RegionBudgetArea, RegionBudgetSnapshot, RegionDetail, RegionKPIRankEntry, RegionPlan, RegionSummary } from "@/shared/types";

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
