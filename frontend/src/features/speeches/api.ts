import { api } from "@/shared/api-client";

export interface Speech {
  id: number;
  dokId: string;
  anforandeNummer?: string;
  politicianId: string;
  politicianName: string;
  party: string;
  date: string; // ISO date-time
  topicHeading?: string;
  snippet: string;
  speechText?: string;
}

export const speechesApi = {
  listRecent: (limit = 100) =>
    api.get<Speech[]>(`/speeches/recent?limit=${limit}`),
  getById: (id: number) =>
    api.get<Speech>(`/speeches/${id}`),
};
