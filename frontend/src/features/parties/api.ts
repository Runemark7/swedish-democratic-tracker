import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

export type PartyMeta = components["schemas"]["Party"];
type GoalVoteBreakdown = components["schemas"]["GoalVoteBreakdown"];
type GoalWithAlignment = components["schemas"]["GoalWithAlignment"];
type PartySummary = components["schemas"]["PartySummary"];

export const partiesApi = {
  listParties: () => api.get<PartySummary[]>("/parties"),
  listGoals: (party: string, topic?: string) =>
    api.get<GoalWithAlignment[]>(`/parties/${party}/goals${topic ? `?topic=${topic}` : ""}`),
  getGoalVotes: (party: string, goalId: number) =>
    api.get<GoalVoteBreakdown>(`/parties/${party}/goals/${goalId}/votes`),
  listPartyMeta: () => api.get<PartyMeta[]>("/party-meta"),
};
