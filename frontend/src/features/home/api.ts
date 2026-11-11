import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

export type FrontPage = components["schemas"]["FrontPageResponse"];

export const frontPageApi = {
  get: () => api.get<FrontPage>("/frontpage"),
};
