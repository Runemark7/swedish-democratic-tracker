import { api } from "@/shared/api-client";
import type { MunicipalityKPIItem, RegionAreaDataPoint, RegionBudgetArea, RegionBudgetSnapshot, RegionDetail, RegionSummary } from "@/shared/types";

export const regionsApi = {
  listRegions: () => api.get<RegionSummary[]>("/regions"),
  getRegion: (code: string) => api.get<RegionDetail>(`/regions/${code}`),
  getRegionBudget: (code: string) => api.get<RegionBudgetArea[]>(`/regions/${code}/budget`),
  getRegionBudgetHistory: (code: string) =>
    api.get<RegionBudgetSnapshot[]>(`/regions/${code}/budget/history`),
  getRegionKPIs: (code: string) => api.get<MunicipalityKPIItem[]>(`/regions/${code}/kpi`),
  getAreaAcrossRegions: (areaName: string, year?: number) =>
    api.get<RegionAreaDataPoint[]>(
      `/regions/budget/area/${encodeURIComponent(areaName)}${year != null ? `?year=${year}` : ""}`
    ),
};
