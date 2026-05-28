import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";
import type { GoalVoteBreakdown, GoalWithAlignment, PartySummary } from "@/shared/types";

export type PartyMeta = components["schemas"]["Party"];

export const partiesApi = {
  listParties: () => api.get<PartySummary[]>("/parties"),
  listGoals: (party: string, topic?: string) =>
    api.get<GoalWithAlignment[]>(`/parties/${party}/goals${topic ? `?topic=${topic}` : ""}`),
  getGoalVotes: (party: string, goalId: number) =>
    api.get<GoalVoteBreakdown>(`/parties/${party}/goals/${goalId}/votes`),
  listPartyMeta: () => api.get<PartyMeta[]>("/party-meta"),
};
