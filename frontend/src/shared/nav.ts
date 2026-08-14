/**
 * The site's navigation, defined once.
 *
 * Four copies of this existed across App.tsx and MobileNav.tsx and had already
 * drifted — the same route was labelled "Partimål & röstning" in one and
 * "Partier" in the other. One definition makes that impossible rather than
 * policed.
 */
export interface NavItem {
  to: string;
  label: string;
  /** Path prefixes that mark this item active. */
  match: string[];
}

/** The primary tier: the questions the site answers. */
export const PRIMARY_NAV: NavItem[] = [
  { to: "/", label: "FRÅGOR", match: ["/", "/committees"] },
  { to: "/om-sajten", label: "OM SAJTEN", match: ["/om-sajten"] },
  { to: "/data", label: "DATA", match: ["/data"] },
];

/**
 * The Riksdag section's own sub-tabs, shown beneath the Fördjupning row
 * whenever the current page sits inside Riksdag territory.
 *
 * Labels are the descriptive set (not the terse one MobileNav previously
 * used): "Partimål & röstning" over "Partier", "Statsbudget" over "Budget",
 * "Enskilda politiker" over "Politiker".
 *
 * Declared before FORDJUPNING_NAV so its Riksdag entry can derive its match
 * list from these `to` fields instead of hand-listing them a second time —
 * see the comment there.
 */
export const RIKSDAG_SUB_NAV: NavItem[] = [
  { to: "/parties",     label: "Partimål & röstning", match: ["/parties"] },
  { to: "/votes",       label: "Omröstningar",        match: ["/votes"] },
  { to: "/budget",      label: "Statsbudget",         match: ["/budget"] },
  { to: "/politicians", label: "Enskilda politiker",  match: ["/politicians"] },
  { to: "/manifestos",  label: "Manifest",            match: ["/manifestos"] },
];

/**
 * The second tier: the four levels of government, as reference.
 *
 * Alphabetical. The previous nav numbered each chamber 1st through 4th, in
 * the order Riksdag, Region, Kommun, Regering — and numerals read as
 * precedence, an order neither the Riksdag nor the constitution publishes.
 * That order also placed Regering last, even though Riksdag and Regering are
 * both national.
 */
export const FORDJUPNING_NAV: NavItem[] = [
  { to: "/kommun",   label: "Kommun",   match: ["/kommun"] },
  { to: "/region",   label: "Region",   match: ["/region"] },
  { to: "/regering", label: "Regering", match: ["/regering"] },
  {
    to: "/riksdag",
    label: "Riksdag",
    // Derived, not listed: RIKSDAG_SUB_NAV's five `to` fields already say
    // what counts as Riksdag. A hand-written copy here agreed with that list
    // only by coincidence — the exact drift this task removed, reappeared
    // inside one file instead of across two. Add a sub-tab to
    // RIKSDAG_SUB_NAV and this list of matches grows with it; nothing to
    // keep in sync by hand. Do not "simplify" this back into a literal array.
    match: ["/riksdag", ...RIKSDAG_SUB_NAV.map((item) => item.to)],
  },
];

/** True when the pathname sits under the item. */
export function isActive(item: NavItem, pathname: string): boolean {
  return item.match.some((m) =>
    m === "/" ? pathname === "/" : pathname === m || pathname.startsWith(m + "/"),
  );
}

/**
 * True while pathname sits inside the Riksdag section — the landing page or
 * any of its five sub-pages. Desktop and mobile both use this to decide
 * whether to show the Riksdag sub-tab row; deriving it here instead of
 * duplicating the lookup in both files is what keeps it in sync.
 */
export function isRiksdagSection(pathname: string): boolean {
  const riksdag = FORDJUPNING_NAV.find((item) => item.to === "/riksdag");
  return riksdag !== undefined && isActive(riksdag, pathname);
}
