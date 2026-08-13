import { api } from "@/shared/api-client";
import type {
  Committee,
  CommitteeVoteringPage,
  GoalWithAlignment,
} from "@/shared/types";

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

  /** The LOVAT row. Not period-scoped: a party's stated goals are not a fact
   *  about a mandate period's voting record. */
  listGoals: (code: string) =>
    api.get<GoalWithAlignment[]>(`/committees/${encodeURIComponent(code)}/goals`),

  listVotes: (code: string, period: string, limit = 50, offset = 0) =>
    api.get<CommitteeVoteringPage>(
      `/committees/${encodeURIComponent(code)}/votes` +
        `?period=${encodeURIComponent(period)}&limit=${limit}&offset=${offset}`,
    ),
};
