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
//
// `goals` are the region's OWN overarching goals, copied verbatim from the
// plan — the complete top-level set the regionfullmäktige decided, never a
// selection or a paraphrase. If we cannot reproduce them faithfully we leave
// `goals` empty and show only the link. The fact layer is the region's words;
// interpretation stays with the reader.

export interface RegionPlanGoal {
  /** Goal/heading exactly as the plan prints it (verbatim, region's wording). */
  title: string;
  /** Optional verbatim clarifying line from the document. */
  description?: string;
}

export interface RegionBudgetDoc {
  /** What the region calls its annual plan document. */
  label: string;
  /** Official page or document hosting the plan, on the region's own domain. */
  url: string;
  /** What the region itself calls these goals (e.g. "Effektmål", "Fokusområden"). */
  goalsLabel?: string;
  /** Plan period the goals are taken from, e.g. "2025–2027". */
  period?: string;
  /** The region's own overarching goals, verbatim and complete. */
  goals?: RegionPlanGoal[];
}

export const REGION_BUDGET_DOCS: Record<string, RegionBudgetDoc> = {
  "01": {
    label: "Budget för Region Stockholm",
    url: "https://www.regionstockholm.se/om-region-stockholm/ekonomi-och-budget/budget/",
    goalsLabel: "Mål i budget",
    period: "2026",
    goals: [
      { title: "Hälso- och sjukvården är behovsstyrd, jämlik och förebygger ohälsa" },
      { title: "Kollektivtrafiken är tillgänglig och attraktiv" },
      { title: "Den hållbara regionala utvecklingen ligger i framkant" },
      { title: "Verksamheten är långsiktigt hållbar och kostnadseffektiv" },
      { title: "Kompetensförsörjningen är långsiktig" },
    ],
  },
  "03": {
    label: "Regionplan och budget",
    url: "https://regionuppsala.se/politik-och-paverkan/handlingar/styrande-dokument/",
    goalsLabel: "Strategiska mål",
    goals: [
      { title: "Attraktiv arbetsplats" },
      { title: "Hög tillgänglighet" },
      { title: "God kvalitet med ekonomi i balans" },
      { title: "Hållbar och konkurrenskraftig utveckling" },
    ],
  },
  "04": {
    label: "Mål och budget",
    url: "https://regionsormland.se/om-region-sormland/ekonomi-och-budget/mal-budget/",
    goalsLabel: "Politiska målområden",
    period: "2025–2027",
    goals: [
      { title: "Region Sörmland har en jämlik, tillgänglig, kvalitativ och effektiv verksamhet" },
      { title: "Region Sörmland arbetar hälsofrämjande och förebyggande" },
      { title: "Region Sörmland använder sina resurser hållbart och effektivt" },
    ],
  },
  "05": { label: "Strategisk plan med treårsbudget", url: "https://www.regionostergotland.se/ro/om-region-ostergotland/ekonomi/planering-och-uppfoljning" },
  "06": {
    label: "Budget med verksamhetsplan",
    url: "https://www.rjl.se/om-oss/Budget-och-styrdokument/budget-med-verksamhetsplan-och-flerarsplan/",
    goalsLabel: "Strategiska mål",
    goals: [
      { title: "Bästa platsen att växa upp, leva, verka och åldras på" },
      { title: "Enklare vardag för invånare och medarbetare" },
      { title: "Bästa möjliga kvalitet" },
      { title: "Bäst på att förbättra och förnya" },
      { title: "Långsiktigt hållbar och innovativ tillväxtregion" },
      { title: "Sveriges bästa offentliga arbetsgivare med Sveriges bästa arbetsplatser" },
      { title: "God och hållbar hushållning" },
    ],
  },
  "07": { label: "Budget (styrande dokument)", url: "https://www.regionkronoberg.se/om-region-kronoberg/styrdokument/" },
  "08": { label: "Regionplan och budget", url: "https://regionkalmar.se/politik-och-demokrati/styrande-dokument/mal-budget-och-uppfoljning/" },
  "09": { label: "Mål och budget", url: "https://gotland.se/region-och-politik/regionfakta-och-statistik/ekonomi-budget-och-publikationer/region-gotlands-budget" },
  "10": {
    label: "Budget med flerårsplan",
    url: "https://regionblekinge.se/om-region-blekinge/styrning-och-mal/budget-2025-med-planer-2026-2027.html",
    goalsLabel: "Övergripande mål",
    period: "2025",
    goals: [
      { title: "Ett Blekinge som växer med fler Blekingebor och god kompetensförsörjning" },
      { title: "En tillgänglig verksamhet med hög kvalitet" },
      { title: "En attraktiv arbetsgivare med medarbetare som trivs och utvecklas" },
      { title: "En ansvarsfull och resurseffektiv ekonomi med budget i balans" },
    ],
  },
  "12": { label: "Verksamhetsplan och budget", url: "https://www.skane.se/om-region-skane/detta-gor-region-skane/publikationer/budget-2025/" },
  "13": { label: "Budget och ekonomi", url: "https://www.regionhalland.se/om-region-halland/budget-och-ekonomi/" },
  "14": {
    label: "Budget",
    url: "https://www.vgregion.se/politik/sa-styrs-vastra-gotalandsregionen/budget/",
    goalsLabel: "Mål i budget",
    period: "2026",
    goals: [
      { title: "Västra Götalandsregionen ska stärka sin beredskap och motståndskraft för att bättre kunna hantera samhällsstörningar, kriser och ett förändrat säkerhetsläge" },
      { title: "Vård ska ges jämlikt i Västra Götalandsregionen" },
      { title: "Sjukvårdens köer ska kortas och tillgängligheten ska öka" },
      { title: "Invånarnas psykiska hälsa ska förbättras och den psykiatriska vården ska stärkas" },
      { title: "Invånarna ska få en god vård på rätt vårdnivå" },
      { title: "Folkhälsan ska förbättras för invånarna genom mer hälsofrämjande och sjukdomsförebyggande arbete i bred samverkan" },
      { title: "Västra Götalandsregionen ska verka för en minskad miljö- och klimatpåverkan i Västra Götaland" },
      { title: "Stärka innovationskraften – för ett hållbart och konkurrenskraftigt näringsliv i framkant" },
      { title: "Bygga kompetens för bättre kompetensförsörjning och livslångt lärande" },
      { title: "Knyta samman Västra Götaland med hållbara resor och förbättrad tillgänglighet" },
      { title: "Ett fritt, rikt och starkt kulturliv i hela Västra Götaland" },
      { title: "Västra Götalandsregionen ska vara en ledande och attraktiv arbetsgivare med konkurrenskraftiga löner och ett hälsofrämjande ledarskap som grund" },
      { title: "Medarbetarnas arbetsmiljö ska förbättras" },
      { title: "Ökad måluppfyllelse för Västra Götalandsregionens hållbarhetsmål" },
    ],
  },
  "17": {
    label: "Regionplan och budget",
    url: "https://www.regionvarmland.se/regionvarmland/om-regionen/om-region-varmland/styrande-dokument",
    goalsLabel: "Effektmål",
    period: "2025–2027",
    goals: [
      { title: "Trygga och nöjda invånare" },
      { title: "Attraktivt och hållbart Värmland" },
      { title: "God, jämlik och jämställd hälsa" },
      { title: "Hållbar organisation" },
    ],
  },
  "18": {
    label: "Verksamhetsplan med budget",
    url: "https://www.regionorebrolan.se/sv/organisation-och-politik/ekonomi-och-budget/budget/",
    goalsLabel: "Övergripande målsättningar",
    period: "2025",
    goals: [
      { title: "Hälso- och sjukvården är god, säker och jämlik" },
      { title: "Tandvården är god, säker och jämlik" },
      { title: "En hållbar och attraktiv kollektivtrafik" },
      { title: "Den hållbara regionala utvecklingen ligger i framkant" },
      { title: "Forskning och utbildning av hög kvalitet med internationell spets" },
      { title: "Digitala lösningar för en enklare vardag för invånare och medarbetare" },
      { title: "Ansvarsfulla verksamheter med höga ambitioner för hållbar utveckling" },
      { title: "Kompetensförsörjningen är långsiktig" },
      { title: "En långsiktig stark och hållbar ekonomi med finansiell god hushållning" },
    ],
  },
  "19": {
    label: "Regionplan och budget",
    url: "https://regionvastmanland.se/contentassets/8afcc008e2e04f3a86bc106b417915e5/reviderad-regionplan-och-budget-2026-2028.pdf",
    goalsLabel: "Regionfullmäktiges övergripande mål",
    period: "2026–2028",
    goals: [
      { title: "Nöjda och trygga invånare", description: "Perspektiv: Invånare" },
      { title: "Effektiva och ändamålsenliga tjänster av god kvalitet", description: "Perspektiv: Tjänst/process" },
      { title: "Engagerade medarbetare med rätt kompetens", description: "Perspektiv: Medarbetare" },
      { title: "Ekonomi i balans", description: "Perspektiv: Ekonomi" },
    ],
  },
  "20": {
    label: "Regionplan med budget",
    url: "https://www.regiondalarna.se/om-oss/regionplan/",
    goalsLabel: "Målområden",
    period: "2025–2027",
    goals: [
      { title: "Till nytta för Dalarnas invånare" },
      { title: "Ett gott medarbetarskap och ledarskap" },
      { title: "Hållbar utveckling" },
    ],
  },
  "21": { label: "Budget", url: "https://www.regiongavleborg.se/om/ekonomi/" },
  "22": {
    label: "Regionplan",
    url: "https://www.rvn.se/sv/Demokrati-och-insyn/Sa-styrs-regionen/Regionplanen/",
    goalsLabel: "Målområden",
    period: "2026–2028",
    goals: [
      { title: "Ett attraktivt Västernorrland med hållbar utveckling" },
      { title: "Region Västernorrland utvecklar arbetet för en god, jämlik och nära vård" },
      { title: "Vi skapar mervärde för befolkningen genom vår verksamhet" },
    ],
  },
  "23": {
    label: "Regionplan och budget",
    url: "https://www.regionjh.se/om-regionen/ekonomi-och-budget",
    goalsLabel: "Övergripande mål",
    period: "2025–2027",
    goals: [
      { title: "God och nära vård i hela Jämtland Härjedalen", description: "Målområde för hälso- och sjukvård" },
      { title: "God och tillgänglig tandvård i hela Jämtland Härjedalen", description: "Målområde för tandvård" },
      { title: "Utvecklingskraft i hela Jämtland Härjedalen", description: "Målområde för regionalt utvecklingsansvar" },
      { title: "Tillgängligt kulturutbud i hela Jämtland Härjedalen", description: "Målområde för kultur" },
      { title: "Goda kommunikationer inom samt till och från Jämtland Härjedalen", description: "Målområde för kollektivtrafik" },
      { title: "Folkhögskolor med hög kvalitet", description: "Målområde för folkhögskolor" },
      { title: "Ändamålsenlig kompetens för effektivt stöd till verksamhet och utveckling", description: "Målområde för stödverksamhet" },
    ],
  },
  "24": {
    label: "Regionplan",
    url: "https://www.regionvasterbotten.se/ekonomi-och-planering",
    goalsLabel: "Regionmål",
    period: "2024",
    goals: [
      { title: "Trygga uppväxtvillkor" },
      { title: "Goda levnadsvillkor" },
      { title: "Ett attraktivt och hållbart Västerbotten" },
      { title: "Tillgänglig vård med god kvalitet och kontinuitet" },
      { title: "Goda arbetsplatser och attraktiva arbetsgivare" },
      { title: "En hållbar miljö- och klimatregion" },
      { title: "Stärkt forskning, utbildning och innovation" },
    ],
  },
  "25": { label: "Strategisk plan med budget", url: "https://www.norrbotten.se/sv/demokrati-och-politik/styrande-dokument/" },
};
