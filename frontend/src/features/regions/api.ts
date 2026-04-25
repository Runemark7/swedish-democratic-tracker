import { api } from "@/shared/api-client";
import type { MunicipalityKPIItem, RegionBudgetArea, RegionDetail, RegionSummary } from "@/shared/types";

export const regionsApi = {
  listRegions: () => api.get<RegionSummary[]>("/regions"),
  getRegion: (code: string) => api.get<RegionDetail>(`/regions/${code}`),
  getRegionBudget: (code: string) => api.get<RegionBudgetArea[]>(`/regions/${code}/budget`),
  getRegionKPIs: (code: string) => api.get<MunicipalityKPIItem[]>(`/regions/${code}/kpi`),
};
