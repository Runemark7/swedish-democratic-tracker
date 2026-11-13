/**
 * Data-source docs link to each other as `./<id>.md` so they read well on
 * GitHub. On the site the same page lives at `/data/<id>`.
 * Returns the in-site path for such a link, or null for anything else.
 */
export function resolveDocHref(href: string | undefined): string | null {
  if (!href) return null;
  const m = href.match(/^(?:\.\/)?([a-z0-9][a-z0-9-]*)\.md(#[^\s]*)?$/i);
  if (!m) return null;
  return `/data/${m[1]}${m[2] ?? ""}`;
}
