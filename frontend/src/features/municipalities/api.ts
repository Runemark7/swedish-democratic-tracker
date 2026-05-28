import { api } from "@/shared/api-client";
import type { MunicipalityDetail, MunicipalitySummary, MunicipalityKPIItem, PopulationTrendEntry, ProcurementCategorySummary, MunicipalityBudgetSnapshot, MunicipalityAreaDataPoint } from "@/shared/types";

export const municipalitiesApi = {
  listMunicipalities: (regionCode?: string) =>
    api.get<MunicipalitySummary[]>(
      `/municipalities${regionCode ? `?regionCode=${regionCode}` : ""}`
    ),
  getMunicipality: (code: string) =>
    api.get<MunicipalityDetail>(`/municipalities/${code}`),
  getMunicipalityKPIs: (code: string) =>
    api.get<MunicipalityKPIItem[]>(`/municipalities/${code}/kpi`),
  getPopulationTrend: (code: string) =>
    api.get<PopulationTrendEntry[]>(`/municipalities/${code}/population-trend`),
  getMunicipalitySpending: (code: string) =>
    api.get<MunicipalityKPIItem[]>(`/municipalities/${code}/spending`),
  getMunicipalityProcurement: (code: string) =>
    api.get<ProcurementCategorySummary[]>(`/municipalities/${code}/procurement`),
  getMunicipalityBudgetHistory: (code: string) =>
    api.get<MunicipalityBudgetSnapshot[]>(`/municipalities/${code}/budget/history`),
  getAreaAcrossMunicipalities: (areaName: string, year?: number) =>
    api.get<MunicipalityAreaDataPoint[]>(
      `/municipalities/budget/area/${encodeURIComponent(areaName)}${year ? `?year=${year}` : ""}`
    ),
  getKPIRanking: (kpiCode: string) =>
    api.get<{ mun_code: string; name: string; value: number; year: number; rank: number; total: number }[]>(
      `/municipalities/kpi/${kpiCode}/ranking`
    ),
  getMunicipalityKPIRanks: (code: string) =>
    api.get<{ kpi: string; rank: number; total: number; mean: number }[]>(
      `/municipalities/${code}/kpi-ranks`
    ),
};
