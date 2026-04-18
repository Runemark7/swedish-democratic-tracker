import { api } from "@/shared/api-client";
import type { TopicContext } from "@/shared/types";

export const contextApi = {
  getTopic: (topic: string) => api.get<TopicContext>(`/context/topic/${topic}`),
};
