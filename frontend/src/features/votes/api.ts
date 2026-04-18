import { api } from "@/shared/api-client";
import type { VoteDetail, VoteSummaryListResponse } from "@/shared/types";

export const votesApi = {
  list: (params: { page?: number; pageSize?: number } = {}) =>
    api.get<VoteSummaryListResponse>(
      `/votes?page=${params.page ?? 1}&pageSize=${params.pageSize ?? 50}`,
    ),
  getDetail: (beteckning: string, punkt: string) =>
    api.get<VoteDetail>(`/votes/${beteckning}/${punkt}`),
};
