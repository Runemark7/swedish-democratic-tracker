import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

export type DepartmentGroup = components["schemas"]["DepartmentGroup"];
export type Minister = components["schemas"]["Minister"];
export type MinisterDetail = components["schemas"]["MinisterDetail"];

export const regeringApi = {
  listMinisters: () => api.get<DepartmentGroup[]>("/ministers"),
  getMinister: (id: string) => api.get<MinisterDetail>(`/ministers/${id}`),
};
