import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

type TopicContext = components["schemas"]["TopicContext"];

export const contextApi = {
  getTopic: (topic: string) => api.get<TopicContext>(`/context/topic/${topic}`),
};
