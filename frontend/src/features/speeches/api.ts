import { api } from "@/shared/api-client";

export interface Speech {
  id: number;
  dokId: string;
  anforandeNummer?: string;
  politicianId: string;
  politicianName: string;
  politicianImageUrl?: string;
  party: string;
  date: string; // ISO date-time
  topicHeading?: string;
  snippet: string;
  speechText?: string;
  relatedDokId?: string;
}

export const speechesApi = {
  listRecent: (limit = 100) =>
    api.get<Speech[]>(`/speeches/recent?limit=${limit}`),
  getById: (id: number) =>
    api.get<Speech>(`/speeches/${id}`),
  listByDocument: (dokId: string) =>
    api.get<Speech[]>(`/speeches/by-document/${encodeURIComponent(dokId)}`),
  listByPolitician: (intressentId: string, limit = 20) =>
    api.get<Speech[]>(`/speeches/by-politician/${encodeURIComponent(intressentId)}?limit=${limit}`),
  listByParty: (party: string, limit = 50) =>
    api.get<Speech[]>(`/speeches/by-party/${encodeURIComponent(party)}?limit=${limit}`),
};
