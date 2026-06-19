// Official annual-plan documents per region ("mål och budget" / "regionplan
// och budget" / "verksamhetsplan med budget"). Every region is required by
// Kommunallagen (11 kap.) to adopt an annual budget containing goals and
// guidelines for its operations. Sweden publishes no structured, machine-
// readable feed of these plans, so we link the official source document
// instead of generating or summarising it. Each URL was verified live to
// resolve to the region's own budget/governing-document page.
//
// Keys are SCB region codes (zero-padded), matching `region.code` from the API.
// `label` is the title the region itself gives the document.

export interface RegionBudgetDoc {
  /** What the region calls its annual plan document. */
  label: string;
  /** Official page or document hosting the plan, on the region's own domain. */
  url: string;
}

export const REGION_BUDGET_DOCS: Record<string, RegionBudgetDoc> = {
  "01": { label: "Budget för Region Stockholm", url: "https://www.regionstockholm.se/om-region-stockholm/ekonomi-och-budget/budget/" },
  "03": { label: "Regionplan och budget", url: "https://regionuppsala.se/politik-och-paverkan/handlingar/styrande-dokument/" },
  "04": { label: "Mål och budget", url: "https://regionsormland.se/om-region-sormland/ekonomi-och-budget/mal-budget/" },
  "05": { label: "Strategisk plan med treårsbudget", url: "https://www.regionostergotland.se/ro/om-region-ostergotland/ekonomi/planering-och-uppfoljning" },
  "06": { label: "Budget med verksamhetsplan", url: "https://www.rjl.se/om-oss/Budget-och-styrdokument/budget-med-verksamhetsplan-och-flerarsplan/" },
  "07": { label: "Budget (styrande dokument)", url: "https://www.regionkronoberg.se/om-region-kronoberg/styrdokument/" },
  "08": { label: "Regionplan och budget", url: "https://regionkalmar.se/politik-och-demokrati/styrande-dokument/mal-budget-och-uppfoljning/" },
  "09": { label: "Mål och budget", url: "https://gotland.se/region-och-politik/regionfakta-och-statistik/ekonomi-budget-och-publikationer/region-gotlands-budget" },
  "10": { label: "Budget med flerårsplan", url: "https://regionblekinge.se/om-region-blekinge/styrning-och-mal/budget-2025-med-planer-2026-2027.html" },
  "12": { label: "Verksamhetsplan och budget", url: "https://www.skane.se/om-region-skane/detta-gor-region-skane/publikationer/budget-2025/" },
  "13": { label: "Budget och ekonomi", url: "https://www.regionhalland.se/om-region-halland/budget-och-ekonomi/" },
  "14": { label: "Budget", url: "https://www.vgregion.se/politik/sa-styrs-vastra-gotalandsregionen/budget/" },
  "17": { label: "Regionplan och budget", url: "https://www.regionvarmland.se/regionvarmland/om-regionen/om-region-varmland/styrande-dokument" },
  "18": { label: "Verksamhetsplan med budget", url: "https://www.regionorebrolan.se/sv/organisation-och-politik/ekonomi-och-budget/budget/" },
  "19": { label: "Regionplan och budget", url: "https://regionvastmanland.se/contentassets/8afcc008e2e04f3a86bc106b417915e5/reviderad-regionplan-och-budget-2026-2028.pdf" },
  "20": { label: "Regionplan med budget", url: "https://www.regiondalarna.se/om-oss/regionplan/" },
  "21": { label: "Budget", url: "https://www.regiongavleborg.se/om/ekonomi/" },
  "22": { label: "Regionplan", url: "https://www.rvn.se/sv/Demokrati-och-insyn/Sa-styrs-regionen/Regionplanen/" },
  "23": { label: "Regionplan och budget", url: "https://www.regionjh.se/om-regionen/ekonomi-och-budget" },
  "24": { label: "Regionplan", url: "https://www.regionvasterbotten.se/ekonomi-och-planering" },
  "25": { label: "Strategisk plan med budget", url: "https://www.norrbotten.se/sv/demokrati-och-politik/styrande-dokument/" },
};
