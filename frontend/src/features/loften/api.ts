import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";
import { loftenQuery, type LoftenFilter } from "./query";

export type Loefte = components["schemas"]["Loefte"];
export type LoefteMedStatus = components["schemas"]["LoefteMedStatus"];
export type BekraftadKoppling = components["schemas"]["BekraftadKoppling"];
export type LoftePartySummary = components["schemas"]["LoftePartySummary"];
export type LoefteRef = components["schemas"]["LoefteRef"];
export type Valmanifest = components["schemas"]["Valmanifest"];

export type { LoftenFilter } from "./query";

export const loftenApi = {
  list: (f: LoftenFilter = {}) => api.get<LoefteMedStatus[]>(`/loften${loftenQuery(f)}`),
  summary: (mandatperiod = "") =>
    api.get<LoftePartySummary[]>(`/loften/summary${mandatperiod ? `?mandatperiod=${encodeURIComponent(mandatperiod)}` : ""}`),
  get: (id: number) => api.get<LoefteMedStatus>(`/loften/${id}`),
  koppling: (id: number) => api.get<BekraftadKoppling>(`/kopplingar/${id}`),
  valmanifest: (party?: string) =>
    api.get<Valmanifest[]>(party ? `/parties/${encodeURIComponent(party)}/valmanifest` : "/valmanifest"),
};
