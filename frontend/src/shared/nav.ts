/**
 * The site's navigation, defined once.
 *
 * Desktop (`App.tsx`) and mobile (`MobileNav.tsx`) both read this. A second
 * copy is how the same route ended up with two labels.
 *
 * Core menu: Beslut · Motioner · Ledamöter · Partier · Regeringen. Regeringen
 * opens /regeringen, which holds the government, the budget, the Riksdag
 * overview and the agencies. /regering, /budget and /riksdag redirect there.
 *
 * Lokalt (kommuner and regioner) is a side menu, not a core item. Om sajten
 * and Data are secondary. Löften are a section under Partier (/loften), not a
 * menu item; /manifestos redirects to the Valmanifest list. Budget and Riksdag
 * are not menu items; they are sections of Regeringen.
 */
export interface NavItem {
  to: string;
  label: string;
  /** Path prefixes that mark this item active, including retired URLs. */
  match: string[];
}

export const PRIMARY_NAV: NavItem[] = [
  { to: "/beslut", label: "Beslut", match: ["/beslut", "/votes"] },
  { to: "/motioner", label: "Motioner", match: ["/motioner"] },
  { to: "/politicians", label: "Ledamöter", match: ["/politicians"] },
  { to: "/parties", label: "Partier", match: ["/parties", "/loften"] },
  { to: "/regeringen", label: "Regeringen", match: ["/regeringen", "/regering", "/budget", "/riksdag"] },
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

/** Browser tab title. Every route keeps the product name. */
export function documentTitle(pathname: string): string {
  if (pathname === "/loften" || pathname.startsWith("/loften/")) return "Riksdagskollen – Löften";
  const primary = PRIMARY_NAV.find((item) => isActive(item, pathname));
  if (primary) return `Riksdagskollen – ${primary.label}`;
  if (isLokaltSection(pathname)) return "Riksdagskollen – Lokalt";
  const secondary = SECONDARY_NAV.find((item) => isActive(item, pathname));
  if (secondary) return `Riksdagskollen – ${secondary.label}`;
  return "Riksdagskollen";
}
