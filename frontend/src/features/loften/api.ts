import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

export type Loefte = components["schemas"]["Loefte"];
export type Valmanifest = components["schemas"]["Valmanifest"];

export const loftenApi = {
  list: (party?: string) =>
    api.get<Loefte[]>(party ? `/parties/${encodeURIComponent(party)}/loften` : "/loften"),
  get: (id: number) => api.get<Loefte>(`/loften/${id}`),
  valmanifest: (party?: string) =>
    api.get<Valmanifest[]>(party ? `/parties/${encodeURIComponent(party)}/valmanifest` : "/valmanifest"),
};
