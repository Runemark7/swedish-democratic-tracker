import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

export type Loefte = components["schemas"]["Loefte"];
export type LoefteMedStatus = components["schemas"]["LoefteMedStatus"];
export type BekraftadKoppling = components["schemas"]["BekraftadKoppling"];
export type LoftePartySummary = components["schemas"]["LoftePartySummary"];
export type LoefteRef = components["schemas"]["LoefteRef"];
export type Valmanifest = components["schemas"]["Valmanifest"];

export interface LoftenFilter {
  party?: string;
  utskott?: string;
  status?: string;
  /** riksmöte:beteckning, e.g. 2026/27:SkU2 */
  betankande?: string;
}

export const loftenApi = {
  list: (f: LoftenFilter = {}) => {
    const q = new URLSearchParams();
    if (f.party) q.set("party", f.party);
    if (f.utskott) q.set("utskott", f.utskott);
    if (f.status) q.set("status", f.status);
    if (f.betankande) q.set("betankande", f.betankande);
    const qs = q.toString();
    return api.get<LoefteMedStatus[]>(`/loften${qs ? `?${qs}` : ""}`);
  },
  summary: () => api.get<LoftePartySummary[]>("/loften/summary"),
  get: (id: number) => api.get<LoefteMedStatus>(`/loften/${id}`),
  koppling: (id: number) => api.get<BekraftadKoppling>(`/kopplingar/${id}`),
  valmanifest: (party?: string) =>
    api.get<Valmanifest[]>(party ? `/parties/${encodeURIComponent(party)}/valmanifest` : "/valmanifest"),
};
