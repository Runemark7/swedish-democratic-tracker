import type { LevelData } from "@/types/democracy";

export const mockRiksdag: LevelData = {
  title: "Riksdagen",
  subtitle: "Sveriges nationella parlament — 349 ledamöter",
  ruling: {
    type: "Regeringskoalition (stöd av SD)",
    parties: [
      { name: "Moderaterna",       short: "M",  seats: 68, color: "#1E88E5" },
      { name: "Kristdemokraterna", short: "KD", seats: 19, color: "#00558E" },
      { name: "Liberalerna",       short: "L",  seats: 16, color: "#006AB3" },
    ],
    support: [
      { name: "Sverigedemokraterna", short: "SD", seats: 73, color: "#DDD014" },
    ],
    opposition: [
      { name: "Socialdemokraterna", short: "S",  seats: 107, color: "#E8112D" },
      { name: "Vänsterpartiet",     short: "V",  seats: 24,  color: "#AF0000" },
      { name: "Centerpartiet",      short: "C",  seats: 24,  color: "#009933" },
      { name: "Miljöpartiet",       short: "MP", seats: 18,  color: "#83CF39" },
    ],
  },
  liveVotes: [
    { time: "Idag 14:20",  title: "Ändring i socialförsäkringsbalken", status: "Bifall",  margin: "196–149", tag: "Välfärd",    beteckning: "SfU2425:18" },
    { time: "Idag 11:05",  title: "Höjda pensioner 2026",              status: "Bifall",  margin: "245–98",  tag: "Pension",    beteckning: "SfU2425:21" },
    { time: "Igår 16:40",  title: "Skärpt straff för narkotikabrott",  status: "Bifall",  margin: "203–142", tag: "Rättsväsen", beteckning: "JuU2425:14" },
    { time: "Igår 10:15",  title: "Sänkt bensinskatt",                 status: "Avslag",  margin: "149–196", tag: "Skatt",      beteckning: "SkU2425:9"  },
    { time: "17 apr",      title: "Utökad försvarsbudget",             status: "Bifall",  margin: "289–54",  tag: "Försvar",    beteckning: "FöU2425:6"  },
  ],
  budget: {
    total: "1 527 mdkr",
    year: "2026",
    sourceUrl: "https://www.riksdagen.se/sv/dokument-och-lagar/dokument/proposition/budgetpropositionen-for-2026_hd031/",
    areas: [
      { name: "Socialförsäkring",    value: 286, pct: 18.5 },
      { name: "Hälsovård",           value: 128, pct: 8.3  },
      { name: "Utbildning",          value: 143, pct: 9.3  },
      { name: "Försvar",             value: 225, pct: 14.6 },
      { name: "Rättsväsen",          value: 95,  pct: 6.1  },
      { name: "Infrastruktur",       value: 107, pct: 6.9  },
      { name: "Statsskuld & räntor", value: 27,  pct: 1.7  },
      { name: "Övrigt",              value: 534, pct: 34.6 },
    ],
  },
  agenda: [
    { id: 1, title: "Reformerad arbetslöshetsförsäkring", description: "Skärpta krav på aktivitet och tydligare matchning mot lediga tjänster. Ersättningsnivåer kopplas till tidigare inkomst med ett nytt trappstegssystem.", source: "Tidöavtalet 2022", status: "in_progress" },
    { id: 2, title: "Ny migrationslagstiftning", description: "Permanenta uppehållstillstånd ersätts av tidsbegränsade med krav på självförsörjning. Asylreglerna anpassas till EU:s miniminivå.", source: "Tidöavtalet 2022", status: "active" },
    { id: 3, title: "Skattesänkningar för arbete", description: "Jobbskatteavdraget utökas i tre steg under mandatperioden med fokus på låg- och medelinkomsttagare för att öka sysselsättningsgraden.", source: "Budgetpropositionen 2024", status: "in_progress" },
    { id: 4, title: "Höjd pensionsålder utreds", description: "En parlamentarisk kommission ska senast 2025 lägga fram förslag om gradvis höjd riktålder för pension i linje med ökad medellivslängd.", source: "Pensionsgruppens direktiv 2023", status: "active" },
    { id: 5, title: "Stärkt försvar mot 2030", description: "Försvarsanslaget når 2,5 % av BNP 2030. Värnplikt utökas och tre nya regementen återaktiveras för att möta förändrat säkerhetsläge.", source: "Försvarspropositionen 2024–2030", status: "in_progress" },
  ],
  authorities: [
    { slug: "polismyndigheten",   name: "Polismyndigheten",   ministry: "Justitiedepartementet",       role: "Ordning & utredning",                  headcount: "35 500", headcountInt: 35500, expenditureMdkr: 41.4, year: 2024, history: [{year:2015,expenditureMdkr:21.3},{year:2016,expenditureMdkr:22.5},{year:2017,expenditureMdkr:23.7},{year:2018,expenditureMdkr:25.1},{year:2019,expenditureMdkr:26.8},{year:2020,expenditureMdkr:28.4},{year:2021,expenditureMdkr:30.9},{year:2022,expenditureMdkr:33.5},{year:2023,expenditureMdkr:37.2},{year:2024,expenditureMdkr:41.4}] },
    { slug: "kriminalvarden",     name: "Kriminalvården",     ministry: "Justitiedepartementet",       role: "Kriminalvård & häkte",                 headcount: "14 200", headcountInt: 14200, expenditureMdkr: 17.6, year: 2024, history: [{year:2015,expenditureMdkr:8.9},{year:2016,expenditureMdkr:9.2},{year:2017,expenditureMdkr:9.6},{year:2018,expenditureMdkr:10.1},{year:2019,expenditureMdkr:10.9},{year:2020,expenditureMdkr:11.5},{year:2021,expenditureMdkr:12.4},{year:2022,expenditureMdkr:13.8},{year:2023,expenditureMdkr:15.6},{year:2024,expenditureMdkr:17.6}] },
    { slug: "forsakringskassan",  name: "Försäkringskassan",  ministry: "Socialdepartementet",         role: "Administration socialförsäkring",      headcount: "14 200", headcountInt: 14200, expenditureMdkr: 9.6,  year: 2024, history: [{year:2015,expenditureMdkr:7.1},{year:2016,expenditureMdkr:7.3},{year:2017,expenditureMdkr:7.5},{year:2018,expenditureMdkr:7.8},{year:2019,expenditureMdkr:8.0},{year:2020,expenditureMdkr:8.3},{year:2021,expenditureMdkr:8.6},{year:2022,expenditureMdkr:8.9},{year:2023,expenditureMdkr:9.2},{year:2024,expenditureMdkr:9.6}] },
    { slug: "skatteverket",       name: "Skatteverket",       ministry: "Finansdepartementet",         role: "Skatt & folkbokföring",                headcount: "11 100", headcountInt: 11100, expenditureMdkr: 8.6,  year: 2024, history: [{year:2015,expenditureMdkr:6.8},{year:2016,expenditureMdkr:7.0},{year:2017,expenditureMdkr:7.1},{year:2018,expenditureMdkr:7.3},{year:2019,expenditureMdkr:7.4},{year:2020,expenditureMdkr:7.6},{year:2021,expenditureMdkr:7.8},{year:2022,expenditureMdkr:8.0},{year:2023,expenditureMdkr:8.3},{year:2024,expenditureMdkr:8.6}] },
    { slug: "sveriges-domstolar", name: "Sveriges Domstolar", ministry: "Justitiedepartementet",       role: "Domstolar & nämnder",                  headcount: "7 200",  headcountInt: 7200,  expenditureMdkr: 7.7,  year: 2024, history: [{year:2015,expenditureMdkr:5.4},{year:2016,expenditureMdkr:5.6},{year:2017,expenditureMdkr:5.8},{year:2018,expenditureMdkr:6.0},{year:2019,expenditureMdkr:6.2},{year:2020,expenditureMdkr:6.4},{year:2021,expenditureMdkr:6.7},{year:2022,expenditureMdkr:7.0},{year:2023,expenditureMdkr:7.3},{year:2024,expenditureMdkr:7.7}] },
    { slug: "arbetsformedlingen", name: "Arbetsförmedlingen", ministry: "Arbetsmarknadsdepartementet", role: "Matchning & arbetsmarknadspolitik",    headcount: "9 400",  headcountInt: 9400,  expenditureMdkr: 7.5,  year: 2024, history: [{year:2015,expenditureMdkr:9.1},{year:2016,expenditureMdkr:9.4},{year:2017,expenditureMdkr:9.8},{year:2018,expenditureMdkr:9.3},{year:2019,expenditureMdkr:8.5},{year:2020,expenditureMdkr:8.1},{year:2021,expenditureMdkr:7.6},{year:2022,expenditureMdkr:7.4},{year:2023,expenditureMdkr:7.4},{year:2024,expenditureMdkr:7.5}] },
    { slug: "migrationsverket",   name: "Migrationsverket",   ministry: "Justitiedepartementet",       role: "Uppehållstillstånd & asyl",            headcount: "5 800",  headcountInt: 5800,  expenditureMdkr: 4.8,  year: 2024, history: [{year:2015,expenditureMdkr:5.2},{year:2016,expenditureMdkr:7.1},{year:2017,expenditureMdkr:6.3},{year:2018,expenditureMdkr:5.5},{year:2019,expenditureMdkr:4.9},{year:2020,expenditureMdkr:4.6},{year:2021,expenditureMdkr:4.2},{year:2022,expenditureMdkr:4.5},{year:2023,expenditureMdkr:4.9},{year:2024,expenditureMdkr:4.8}] },
    { slug: "tullverket",         name: "Tullverket",         ministry: "Finansdepartementet",         role: "Tull & gränskontroll",                 headcount: "2 700",  headcountInt: 2700,  expenditureMdkr: 2.9,  year: 2024, history: [{year:2015,expenditureMdkr:2.0},{year:2016,expenditureMdkr:2.1},{year:2017,expenditureMdkr:2.2},{year:2018,expenditureMdkr:2.3},{year:2019,expenditureMdkr:2.4},{year:2020,expenditureMdkr:2.5},{year:2021,expenditureMdkr:2.6},{year:2022,expenditureMdkr:2.7},{year:2023,expenditureMdkr:2.8},{year:2024,expenditureMdkr:2.9}] },
    { slug: "aklagarmyndigheten", name: "Åklagarmyndigheten", ministry: "Justitiedepartementet",       role: "Åklagare",                             headcount: "1 900",  headcountInt: 1900,  expenditureMdkr: 2.6,  year: 2024, history: [{year:2015,expenditureMdkr:1.6},{year:2016,expenditureMdkr:1.7},{year:2017,expenditureMdkr:1.8},{year:2018,expenditureMdkr:1.9},{year:2019,expenditureMdkr:2.0},{year:2020,expenditureMdkr:2.1},{year:2021,expenditureMdkr:2.2},{year:2022,expenditureMdkr:2.3},{year:2023,expenditureMdkr:2.4},{year:2024,expenditureMdkr:2.6}] },
    { slug: "sakerhetspolisen",   name: "Säkerhetspolisen",   ministry: "Justitiedepartementet",       role: "Nationell säkerhet",                   headcount: "2 100",  headcountInt: 2100,  expenditureMdkr: 2.4,  year: 2024, history: [{year:2015,expenditureMdkr:1.2},{year:2016,expenditureMdkr:1.3},{year:2017,expenditureMdkr:1.5},{year:2018,expenditureMdkr:1.6},{year:2019,expenditureMdkr:1.8},{year:2020,expenditureMdkr:1.9},{year:2021,expenditureMdkr:2.0},{year:2022,expenditureMdkr:2.1},{year:2023,expenditureMdkr:2.2},{year:2024,expenditureMdkr:2.4}] },
  ],
  kpis: [
    {
      label: "Statsskuld/BNP",
      description: "Statens samlade skuld i förhållande till BNP. Hög skuld begränsar statens möjlighet att investera i välfärd och krisberedskap utan att höja skatterna.",
      value: "38 %", raw: 38, target: 35, worseHigher: true, unit: "%", trend: "down", delta: "−1,2 pp", note: "Källa: Riksgälden", sourceUrl: "https://www.riksgalden.se/sv/var-verksamhet/statsskulden/",
    },
    {
      label: "Arbetslöshet",
      description: "Andel av arbetskraften som är arbetslösa. Hög arbetslöshet minskar skatteintäkterna och ökar kostnaderna för socialförsäkringssystemet.",
      value: "8,5 %", raw: 8.5, target: 5, worseHigher: true, unit: "%", trend: "up", delta: "+0,3 pp", note: "Källa: SCB", sourceUrl: "https://www.scb.se/hitta-statistik/statistik-efter-amne/arbetsmarknad/",
    },
    {
      label: "Inflation (KPI)",
      description: "Konsumentprisindex visar hur snabbt priserna stiger. Riksbankens mål är 2 %. För hög inflation urholkar hushållens köpkraft.",
      value: "1,8 %", raw: 1.8, target: 2, worseHigher: false, unit: "%", trend: "down", delta: "−0,6 pp", note: "Källa: SCB", sourceUrl: "https://www.scb.se/hitta-statistik/statistik-efter-amne/priser-och-konsumtion/konsumentprisindex/konsumentprisindex-kpi/",
    },
    {
      label: "BNP-tillväxt",
      description: "Hur snabbt ekonomin växer. Positiv tillväxt innebär fler jobb och ökade skatteintäkter som kan finansiera välfärden.",
      value: "1,2 %", raw: 1.2, target: 2, worseHigher: false, unit: "%", trend: "up", delta: "+0,8 pp", note: "Källa: Konjunkturinstitutet", sourceUrl: "https://www.konj.se/statistik/prognoser.html",
    },
  ],
};

export const mockRegion: LevelData = {
  title: "Region Stockholm",
  subtitle: "Regional nivå — hälso- och sjukvård, kollektivtrafik",
  population: "2,44 miljoner",
  ruling: {
    type: "Blågrön majoritet",
    parties: [
      { name: "Moderaterna",      short: "M",  seats: 42, color: "#1E88E5" },
      { name: "Miljöpartiet",     short: "MP", seats: 12, color: "#83CF39" },
      { name: "Liberalerna",      short: "L",  seats: 8,  color: "#006AB3" },
      { name: "Kristdemokraterna",short: "KD", seats: 6,  color: "#00558E" },
      { name: "Centerpartiet",    short: "C",  seats: 5,  color: "#009933" },
    ],
    opposition: [
      { name: "Socialdemokraterna",  short: "S",  seats: 35, color: "#E8112D" },
      { name: "Sverigedemokraterna", short: "SD", seats: 18, color: "#DDD014" },
      { name: "Vänsterpartiet",      short: "V",  seats: 13, color: "#AF0000" },
    ],
  },
  liveVotes: [
    { time: "Idag 09:30", title: "Ny vårdcentral i Norrtälje",  status: "Bifall", tag: "Vård",   beteckning: "SoU2425:12" },
    { time: "Igår 15:00", title: "Biljettprishöjning SL 2026",  status: "Bifall", tag: "Trafik", beteckning: "TU2425:8"   },
    { time: "16 apr",     title: "Utökade BVC-tider",           status: "Bifall", tag: "Vård",   beteckning: "SoU2425:15" },
    { time: "14 apr",     title: "Extra medel akutvård",        status: "Bifall", tag: "Vård",   beteckning: "SoU2425:11" },
  ],
  budget: {
    total: "118 mdkr",
    year: "2026",
    areas: [
      { name: "Hälso- och sjukvård",  value: 82, pct: 69.5 },
      { name: "Kollektivtrafik",      value: 22, pct: 18.6 },
      { name: "Regional utveckling",  value: 6,  pct: 5.1  },
      { name: "Administration",       value: 4,  pct: 3.4  },
      { name: "Övrigt",              value: 4,  pct: 3.4  },
    ],
  },
  agenda: [
    { id: 1, title: "Korta vårdköer till akutsjukvård", description: "Garantera att alla akutpatienter tas om hand inom 4 timmar. Extra resurser tillförs akutmottagningarna vid Karolinska, Södersjukhuset och Danderyds sjukhus.", source: "Regionalt mandatprogram M+MP+L+KD+C 2022–2026", status: "in_progress" },
    { id: 2, title: "Förbättra psykiatrisk vård", description: "Tillföra 150 nya vårdplatser inom psykiatrin och korta väntetiden till BUP från 90 till 30 dagar under mandatperioden.", source: "Regionalt mandatprogram M+MP+L+KD+C 2022–2026", status: "active" },
    { id: 3, title: "Utbyggd tunnelbana mot 2030", description: "Förlänga tunnelbanan till Nacka, Barkarby och Älvsjö enligt Stockholmsöverenskommelsen. Byggnation pågår och beräknas klar 2030.", source: "Stockholmsöverenskommelsen + regionalt mandatprogram", status: "in_progress" },
    { id: 4, title: "Digital vårdplattform", description: "Samla journalsystem, remisshantering och patientkommunikation i en gemensam regional plattform för att minska dubbelarbete och förbättra informationsflödet.", source: "Regionalt mandatprogram M+MP+L+KD+C 2022–2026", status: "active" },
  ],
  kpis: [
    { label: "Väntetid akut",       description: "Genomsnittlig väntetid på akutmottagning från ankomst till läkarbedömning.",         value: "3 h 12 min", raw: 3.2, target: 3.0, worseHigher: true,  unit: "h",  trend: "up",   delta: "+8 min",    note: "Mål: under 3 h" },
    { label: "Bedömning inom 3 mån",description: "Andel patienter med specialistvårdsremiss som fått tid inom 90 dagar.",              value: "78 %",       raw: 78,  target: 80,  worseHigher: false, unit: "%",  trend: "up",   delta: "+2 pp",     note: "Mål: 80 %"      },
    { label: "Punktlighet SL",      description: "Andel avgångar inom kollektivtrafiken som avgår inom 3 minuter från tidtabell.",     value: "94,2 %",     raw: 94.2,target: 95,  worseHigher: false, unit: "%",  trend: "down", delta: "−0,5 pp",   note: "Mål: 95 %"      },
  ],
};

export const mockKommun: LevelData = {
  title: "Malmö kommun",
  subtitle: "Kommunal nivå — skola, omsorg, gator, plan",
  population: "365 000",
  ruling: {
    type: "Rödgrön minoritet",
    parties: [
      { name: "Socialdemokraterna", short: "S",  seats: 20, color: "#E8112D" },
      { name: "Miljöpartiet",       short: "MP", seats: 5,  color: "#83CF39" },
      { name: "Vänsterpartiet",     short: "V",  seats: 5,  color: "#AF0000" },
    ],
    opposition: [
      { name: "Moderaterna",         short: "M",  seats: 14, color: "#1E88E5" },
      { name: "Sverigedemokraterna", short: "SD", seats: 12, color: "#DDD014" },
      { name: "Liberalerna",         short: "L",  seats: 3,  color: "#006AB3" },
      { name: "Centerpartiet",       short: "C",  seats: 2,  color: "#009933" },
    ],
  },
  liveVotes: [
    { time: "Idag 13:15", title: "Ny förskola i Sofielund",         status: "Bifall",     tag: "Skola",      beteckning: "UbU2425:17" },
    { time: "Idag 10:40", title: "Parkeringsavgifter innerstaden",   status: "Bifall",     tag: "Trafik",     beteckning: "TU2425:10"  },
    { time: "Igår 14:00", title: "Upprustning Folkets park",         status: "Bifall",     tag: "Stadsmiljö", beteckning: "CU2425:5"   },
    { time: "16 apr",     title: "Sommarjobb till unga, budget",     status: "Bifall",     tag: "Arbete",     beteckning: "AU2425:13"  },
    { time: "15 apr",     title: "Ny detaljplan Västra hamnen",      status: "Återremiss", tag: "Plan",       beteckning: "CU2425:8"   },
  ],
  budget: {
    total: "24,8 mdkr",
    year: "2026",
    areas: [
      { name: "Förskola & skola", value: 8.9, pct: 35.9 },
      { name: "Äldreomsorg",      value: 4.2, pct: 16.9 },
      { name: "Individ & familj", value: 3.1, pct: 12.5 },
      { name: "Funktionsstöd",    value: 2.4, pct: 9.7  },
      { name: "Gata, park, plan", value: 1.8, pct: 7.3  },
      { name: "Vatten & avlopp", value: 1.8, pct: 7.3  },
      { name: "Kultur & fritid",  value: 1.2, pct: 4.8  },
      { name: "Övrigt",           value: 1.4, pct: 5.6  },
    ],
  },
  agenda: [
    { id: 1, title: "Fler lärare per elev", description: "Rekrytera 80 nya lärartjänster och minska klasstorleken från 26 till 22 elever per klass. Fokus på skolor med lägst meritvärden i Malmö.", source: "Mandatprogrammet S+MP+V 2022–2026", status: "in_progress" },
    { id: 2, title: "Trygghetsskapande belysning", description: "Uppgradera gatubelysningen i Rosengård, Herrgården och Husie under 2024–2025 med LED-armaturer och rörelsestyrning för att förbättra den upplevda tryggheten.", source: "Mandatprogrammet S+MP+V 2022–2026", status: "in_progress" },
    { id: 3, title: "Snabbare bygglov", description: "Korta handläggningstiden för standardärenden från 11 till 6 veckor genom digitalt system och ökad bemanning på stadsbyggnadskontoret.", source: "Mandatprogrammet S+MP+V 2022–2026", status: "active" },
    { id: 4, title: "Fossilfritt Malmö 2030", description: "Kommunens fordonsflotta ska vara 100 % fossilfri 2026. Solceller på alla kommunala tak 2028. Nettopositiv koldioxidbalans 2030 enligt antagen klimatplan.", source: "Malmö stads klimatplan 2022–2030", status: "in_progress" },
  ],
  kpis: [
    { label: "Lärare per 100 elever",  description: "Antal heltidsanställda lärare per 100 elever i kommunens grundskolor.",                      value: "8,1",  raw: 8.1, target: 8.5, worseHigher: false, unit: "",  trend: "flat", delta: "±0",    note: "Mål: 8,5"        },
    { label: "Handläggning bygglov",   description: "Genomsnittlig handläggningstid i veckor för standardärenden på plan- och byggavdelningen.",   value: "11 v", raw: 11,  target: 10,  worseHigher: true,  unit: "v", trend: "down", delta: "−2 v",  note: "Mål: under 10 v" },
    { label: "Otrygghet (SCB)",        description: "Andel invånare som uppger att de känner sig otrygga i sitt bostadsområde (SCB trygghetsmätning).", value: "28 %", raw: 28,  target: 22,  worseHigher: true,  unit: "%", trend: "down", delta: "−3 pp", note: "Mål: 22 %"       },
  ],
};
