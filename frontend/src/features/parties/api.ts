import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

export type PartyMeta = components["schemas"]["Party"];
type GoalWithAlignment = components["schemas"]["GoalWithAlignment"];
type PartySummary = components["schemas"]["PartySummary"];

export const partiesApi = {
  listParties: () => api.get<PartySummary[]>("/parties"),
  listGoals: (party: string, topic?: string) =>
    api.get<GoalWithAlignment[]>(`/parties/${party}/goals${topic ? `?topic=${topic}` : ""}`),
  listPartyMeta: () => api.get<PartyMeta[]>("/party-meta"),
};

export type PartyRecord = components["schemas"]["PartyRecord"];
export type PartyDecision = components["schemas"]["PartyDecision"];

export const partyRecordApi = {
  get: (party: string) => api.get<PartyRecord>(`/parties/${encodeURIComponent(party)}/record`),
};
