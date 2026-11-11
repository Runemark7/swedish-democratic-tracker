export const MANDATE_LABEL: Record<string, string> = {
  "2022-2026": "2022–2026",
  "2026-2030": "2026–2030",
};

export const STATUS_FILTERS: { slug: string; label: string }[] = [
  { slug: "for", label: "Röstat för" },
  { slug: "emot", label: "Röstat emot" },
  { slug: "blandat", label: "Blandat" },
  { slug: "inte-provat", label: "Inte prövat än" },
  { slug: "arkiv", label: "Arkiv (ingen status)" },
];
