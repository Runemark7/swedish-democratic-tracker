import type { Authority, AuthorityDetail } from "@/types/democracy";

export const riksdagApi = {
  getAuthorities: (): Promise<Authority[]> =>
    fetch("/api/riksdag/authorities").then((r) => {
      if (!r.ok) throw new Error(`riksdag/authorities: ${r.status}`);
      return r.json();
    }),

  getAuthority: (slug: string): Promise<AuthorityDetail> =>
    fetch(`/api/riksdag/authorities/${encodeURIComponent(slug)}`).then((r) => {
      if (!r.ok) throw new Error(`riksdag/authorities/${slug}: ${r.status}`);
      return r.json();
    }),
};
