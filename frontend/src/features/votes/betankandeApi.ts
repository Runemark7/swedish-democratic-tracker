import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

export type BetankandePage = components["schemas"]["BetankandePage"];
export type BetankandePunkt = components["schemas"]["BetankandePunkt"];

/** Canonical page URL. The riksmöte keeps its slash form as Riksdagen writes
 *  it (2025/26), never 2025-26; an optional punkt becomes the anchor. */
export function betankandePath(riksmote: string, beteckning: string, punkt?: string): string {
  const hash = punkt ? `#punkt-${encodeURIComponent(punkt)}` : "";
  return `/beslut/${riksmote}/${encodeURIComponent(beteckning)}${hash}`;
}

export const betankandeApi = {
  /** The riksmöte sits in one path segment, so its slash is encoded (2025%2F26). */
  get: (riksmote: string, beteckning: string) =>
    api.get<BetankandePage>(`/betankanden/${encodeURIComponent(riksmote)}/${encodeURIComponent(beteckning)}`),
  resolve: (beteckning: string, punkt?: string, rm?: string) => {
    const p = new URLSearchParams({ beteckning });
    if (punkt) p.set("punkt", punkt);
    if (rm) p.set("rm", rm);
    return api.get<{ riksmote: string; beteckning: string; punkt: string }>(`/betankanden/resolve?${p}`);
  },
};
