import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

type PoliticianListResponse = components["schemas"]["PoliticianListResponse"];
type PoliticianSummary = components["schemas"]["PoliticianSummary"];
type VoteListResponse = components["schemas"]["VoteListResponse"];

export const politiciansApi = {
  list: (params: { party?: string; active?: boolean; page?: number; pageSize?: number } = {}) => {
    const q = new URLSearchParams();
    if (params.party) q.set("party", params.party);
    if (params.active === false) q.set("active", "false");
    if (params.page) q.set("page", String(params.page));
    if (params.pageSize) q.set("pageSize", String(params.pageSize));
    const qs = q.toString();
    return api.get<PoliticianListResponse>(`/politicians${qs ? `?${qs}` : ""}`);
  },
  getById: (id: string) => api.get<PoliticianSummary>(`/politicians/${id}`),
  listVotes: (id: string, params: { session?: string; page?: number; pageSize?: number } = {}) => {
    const q = new URLSearchParams();
    if (params.session) q.set("session", params.session);
    if (params.page) q.set("page", String(params.page));
    if (params.pageSize) q.set("pageSize", String(params.pageSize));
    const qs = q.toString();
    return api.get<VoteListResponse>(`/politicians/${id}/votes${qs ? `?${qs}` : ""}`);
  },
};
