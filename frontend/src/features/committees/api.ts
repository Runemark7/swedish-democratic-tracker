import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

type Committee = components["schemas"]["Committee"];
type CommitteeVoteringPage = components["schemas"]["CommitteeVoteringPage"];

/**
 * Committee fetchers.
 *
 * `period` is required by every one of these endpoints and has no default: an
 * unscoped query would blend mandate periods together the moment the record
 * holds more than one, and an unknown period answers 404 rather than an empty
 * list, because "this committee decided nothing" is a claim about the record
 * that we are in no position to make. Callers read the code from
 * useRecordCoverage().mandate.code rather than hardcoding a vintage.
 */
export const committeesApi = {
  listCommittees: (period: string) =>
    api.get<Committee[]>(`/committees?period=${encodeURIComponent(period)}`),

  getCommittee: (code: string, period: string) =>
    api.get<Committee>(
      `/committees/${encodeURIComponent(code)}?period=${encodeURIComponent(period)}`,
    ),

  listVotes: (code: string, period: string, limit = 50, offset = 0) =>
    api.get<CommitteeVoteringPage>(
      `/committees/${encodeURIComponent(code)}/votes` +
        `?period=${encodeURIComponent(period)}&limit=${limit}&offset=${offset}`,
    ),
};
