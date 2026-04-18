import { api } from "@/shared/api-client";
import type { MunicipalityDetail, MunicipalitySummary } from "@/shared/types";

export const municipalitiesApi = {
  listMunicipalities: (regionCode?: string) =>
    api.get<MunicipalitySummary[]>(
      `/municipalities${regionCode ? `?regionCode=${regionCode}` : ""}`
    ),
  getMunicipality: (code: string) =>
    api.get<MunicipalityDetail>(`/municipalities/${code}`),
};
