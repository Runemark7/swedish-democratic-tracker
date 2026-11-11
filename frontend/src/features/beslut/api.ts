import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

export type BeslutListResponse = components["schemas"]["BeslutListResponse"];
export type BeslutListRow = components["schemas"]["BeslutListRow"];

/** The filters of the Beslut list; they live in the URL so a view can be shared. */
export interface BeslutFilters {
  state?: string;
  utskott?: string;
  party?: string;
}

export const beslutApi = {
  list: (f: BeslutFilters, cursor?: string) => {
    const p = new URLSearchParams();
    for (const [k, v] of Object.entries(f)) if (v) p.set(k, v);
    if (cursor) p.set("cursor", cursor);
    p.set("limit", "25");
    return api.get<BeslutListResponse>(`/beslut?${p.toString()}`);
  },
};
