import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

export type Motion = components["schemas"]["Motion"];

/** Page URL for a motion. The riksmöte keeps its slash form, encoded as
 *  2017%2F18 because it sits in one path segment. Never 2017-18. */
export function motionPath(riksmote: string, beteckning: string): string {
  return `/motioner/${encodeURIComponent(riksmote)}/${encodeURIComponent(beteckning)}`;
}

export const motionsApi = {
  get: (riksmote: string, beteckning: string) =>
    api.get<Motion>(`/motions/${encodeURIComponent(riksmote)}/${encodeURIComponent(beteckning)}`),
};
