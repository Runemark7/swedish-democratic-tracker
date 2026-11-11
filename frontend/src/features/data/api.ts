import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

export type DataCoverage = components["schemas"]["DataCoverage"];

export const dataApi = {
  coverage: () => api.get<{ data: DataCoverage[] }>("/data/coverage"),
};
