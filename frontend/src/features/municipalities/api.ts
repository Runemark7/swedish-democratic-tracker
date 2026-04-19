import { api } from "@/shared/api-client";
import type { MunicipalityDetail, MunicipalitySummary, MunicipalityKPIItem, PopulationTrendEntry, ProcurementCategorySummary } from "@/shared/types";

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
};
