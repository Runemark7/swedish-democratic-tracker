import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";
import type { VoteDetail, VoteSummaryListResponse, RiksdagDocument, RiksdagDocumentFull } from "@/shared/types";

type RecentBetankande = components["schemas"]["RecentBetankande"];

export const votesApi = {
  list: (params: { page?: number; pageSize?: number } = {}) =>
    api.get<VoteSummaryListResponse>(
      `/votes?page=${params.page ?? 1}&pageSize=${params.pageSize ?? 50}`,
    ),
  getDetail: (beteckning: string, punkt: string) =>
    api.get<VoteDetail>(`/votes/${beteckning}/${punkt}`),
  getDocument: (dokId: string) =>
    api.get<RiksdagDocument>(`/documents/${encodeURIComponent(dokId)}`),
  getDocumentFull: (dokId: string) =>
    api.get<RiksdagDocumentFull>(`/documents/${encodeURIComponent(dokId)}/full`),
  /** No committee filter — Riksdagen ignores `organ` on the underlying feed,
   *  so nothing here selects which committees appear. The list is not stable
   *  between calls: the upstream intermittently answers an identical query with
   *  zero documents, and it mixes decided betänkanden with planned ones. */
  listRecent: (count = 40) =>
    api.get<RecentBetankande[]>(`/votes/recent?count=${count}`),
};
