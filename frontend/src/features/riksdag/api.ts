import type { Authority } from "@/types/democracy";

export const riksdagApi = {
  getAuthorities: (): Promise<Authority[]> =>
    fetch("/api/riksdag/authorities").then((r) => {
      if (!r.ok) throw new Error(`riksdag/authorities: ${r.status}`);
      return r.json();
    }),
};
