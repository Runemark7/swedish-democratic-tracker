/**
 * The site's navigation, defined once.
 *
 * Desktop (`App.tsx`) and mobile (`MobileNav.tsx`) both read this. A second
 * copy is how the same route ended up with two labels.
 *
 * The core product — what the Riksdag decides — sits in PRIMARY_NAV. Kommun
 * and region are a side product under Lokalt. Om sajten and Data are
 * secondary: they stay linked, they are not the front door.
 */
export interface NavItem {
  to: string;
  label: string;
  /** Path prefixes that mark this item active. */
  match: string[];
}

/**
 * Top-level nav. The old Riksdag sub-tabs (partier, omröstningar, politiker,
 * manifest, plus statsbudget so that page stays one click away) live here.
 * Riksdag and Regering stay too: dropping the Fördjupning row must not hide
 * those routes.
 *
 * Home is the wordmark, not a pill. There is no "Frågor" item.
 */
export const PRIMARY_NAV: NavItem[] = [
  { to: "/parties", label: "Partier", match: ["/parties"] },
  { to: "/votes", label: "Omröstningar", match: ["/votes", "/beslut"] },
  { to: "/politicians", label: "Politiker", match: ["/politicians"] },
  { to: "/manifestos", label: "Manifest", match: ["/manifestos"] },
  { to: "/budget", label: "Budget", match: ["/budget"] },
  { to: "/riksdag", label: "Riksdag", match: ["/riksdag"] },
  { to: "/regering", label: "Regering", match: ["/regering"] },
];

/** Kommun and region, collapsed. The routes themselves are unchanged. */
export const LOKALT_NAV: NavItem[] = [
  { to: "/kommun", label: "Kommuner", match: ["/kommun"] },
  { to: "/region", label: "Regioner", match: ["/region"] },
];

/** Footer and the quiet header links. Not primary pills. */
export const SECONDARY_NAV: NavItem[] = [
  { to: "/om-sajten", label: "Om sajten", match: ["/om-sajten"] },
  { to: "/data", label: "Data", match: ["/data"] },
];

/** True when the pathname sits under the item. */
export function isActive(item: NavItem, pathname: string): boolean {
  return item.match.some((m) =>
    m === "/" ? pathname === "/" : pathname === m || pathname.startsWith(m + "/"),
  );
}

/** True on the Lokalt landing and on every kommun or region page. */
export function isLokaltSection(pathname: string): boolean {
  return (
    pathname === "/lokalt" ||
    pathname.startsWith("/lokalt/") ||
    LOKALT_NAV.some((item) => isActive(item, pathname))
  );
}
