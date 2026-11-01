/**
 * Regeringen is one page. Budget, the Riksdag overview and the old Regering
 * page redirect here. Agency detail URLs and minister detail URLs stay.
 */
export const REGERINGEN_PATH = "/regeringen";

export const REGERINGEN_SECTIONS = {
  statsrad: "statsrad",
  budget: "budget",
  riksdag: "riksdag",
  myndigheter: "myndigheter",
} as const;

export type RegeringenSection = (typeof REGERINGEN_SECTIONS)[keyof typeof REGERINGEN_SECTIONS];

/** Location for a section on the Regeringen page. Hash includes "#". */
export function regeringenSection(section: RegeringenSection): { pathname: string; hash: string } {
  return { pathname: REGERINGEN_PATH, hash: `#${section}` };
}

/** Budget area detail, a sub-page of Regeringen. */
export function budgetAreaPath(code: string): string {
  return `${REGERINGEN_PATH}/budget/${encodeURIComponent(code)}`;
}

/**
 * Where an absorbed URL should land. Null when this path is not one of them,
 * including a minister (`/regering/:id`) and the agency register
 * (`/riksdag/myndigheter` and `/riksdag/myndigheter/:slug`).
 */
export function regeringenRedirect(pathname: string): string | { pathname: string; hash: string } | null {
  if (pathname === "/regering") return REGERINGEN_PATH;
  if (pathname === "/budget") return regeringenSection("budget");
  if (pathname === "/riksdag") return regeringenSection("riksdag");
  const area = /^\/budget\/areas\/([^/]+)$/.exec(pathname);
  if (area) return budgetAreaPath(decodeURIComponent(area[1]));
  return null;
}
