import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

export type Motion = components["schemas"]["Motion"];
export type MotionListResponse = components["schemas"]["MotionListResponse"];
export type MotionListRow = components["schemas"]["MotionListRow"];

/** The filters of the Motioner list; they live in the URL so a view can be shared. */
export interface MotionFilters {
  q?: string;
  party?: string;
  utskott?: string;
  outcome?: string;
  riksmote?: string;
  kind?: string;
}

/** Page URL for a motion. The riksmöte keeps its slash form, encoded as
 *  2017%2F18 because it sits in one path segment. Never 2017-18. */
export function motionPath(riksmote: string, beteckning: string): string {
  return `/motioner/${encodeURIComponent(riksmote)}/${encodeURIComponent(beteckning)}`;
}

export const motionsApi = {
  list: (f: MotionFilters, cursor?: string) => {
    const p = new URLSearchParams();
    for (const [k, v] of Object.entries(f)) if (v) p.set(k, v);
    if (cursor) p.set("cursor", cursor);
    p.set("limit", "25");
    return api.get<MotionListResponse>(`/motions?${p.toString()}`);
  },
  get: (riksmote: string, beteckning: string) =>
    api.get<Motion>(`/motions/${encodeURIComponent(riksmote)}/${encodeURIComponent(beteckning)}`),
};
