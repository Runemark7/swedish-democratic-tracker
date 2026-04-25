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
    total: "1 389 mdkr",
    year: "2026",
    areas: [
      { name: "Socialförsäkring",    value: 316, pct: 22.7 },
      { name: "Hälsovård",           value: 97,  pct: 7.0  },
      { name: "Utbildning",          value: 99,  pct: 7.1  },
      { name: "Försvar",             value: 143, pct: 10.3 },
      { name: "Rättsväsen",          value: 64,  pct: 4.6  },
      { name: "Infrastruktur",       value: 78,  pct: 5.6  },
      { name: "Statsskuld & räntor", value: 34,  pct: 2.4  },
      { name: "Övrigt",              value: 558, pct: 40.3 },
    ],
  },
  agenda: [
    { title: "Reformerad arbetslöshetsförsäkring", description: "Skärpta krav på aktivitet och tydligare matchning mot lediga tjänster. Ersättningsnivåer kopplas till tidigare inkomst med ett nytt trappstegssystem.", source: "Tidöavtalet 2022", status: "in_progress" },
    { title: "Ny migrationslagstiftning", description: "Permanenta uppehållstillstånd ersätts av tidsbegränsade med krav på självförsörjning. Asylreglerna anpassas till EU:s miniminivå.", source: "Tidöavtalet 2022", status: "active" },
    { title: "Skattesänkningar för arbete", description: "Jobbskatteavdraget utökas i tre steg under mandatperioden med fokus på låg- och medelinkomsttagare för att öka sysselsättningsgraden.", source: "Budgetpropositionen 2024", status: "in_progress" },
    { title: "Höjd pensionsålder utreds", description: "En parlamentarisk kommission ska senast 2025 lägga fram förslag om gradvis höjd riktålder för pension i linje med ökad medellivslängd.", source: "Pensionsgruppens direktiv 2023", status: "active" },
    { title: "Stärkt försvar mot 2030", description: "Försvarsanslaget når 2,5 % av BNP 2030. Värnplikt utökas och tre nya regementen återaktiveras för att möta förändrat säkerhetsläge.", source: "Försvarspropositionen 2024–2030", status: "in_progress" },
  ],
  authorities: [
    { name: "Försäkringskassan",  role: "Utbetalar socialförsäkring",          headcount: "14 200", budget: "259 mdkr" },
    { name: "Arbetsförmedlingen", role: "Matchning & arbetsmarknadspolitik",    headcount: "9 400",  budget: "71 mdkr"  },
    { name: "Skatteverket",       role: "Skatt & folkbokföring",                headcount: "11 100", budget: "8,4 mdkr" },
    { name: "Migrationsverket",   role: "Uppehållstillstånd & asyl",            headcount: "5 800",  budget: "5,9 mdkr" },
    { name: "Polismyndigheten",   role: "Ordning & utredning",                  headcount: "35 500", budget: "42 mdkr"  },
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
    { title: "Korta vårdköer till akutsjukvård", description: "Garantera att alla akutpatienter tas om hand inom 4 timmar. Extra resurser tillförs akutmottagningarna vid Karolinska, Södersjukhuset och Danderyds sjukhus.", source: "Regionalt mandatprogram M+MP+L+KD+C 2022–2026", status: "in_progress" },
    { title: "Förbättra psykiatrisk vård", description: "Tillföra 150 nya vårdplatser inom psykiatrin och korta väntetiden till BUP från 90 till 30 dagar under mandatperioden.", source: "Regionalt mandatprogram M+MP+L+KD+C 2022–2026", status: "active" },
    { title: "Utbyggd tunnelbana mot 2030", description: "Förlänga tunnelbanan till Nacka, Barkarby och Älvsjö enligt Stockholmsöverenskommelsen. Byggnation pågår och beräknas klar 2030.", source: "Stockholmsöverenskommelsen + regionalt mandatprogram", status: "in_progress" },
    { title: "Digital vårdplattform", description: "Samla journalsystem, remisshantering och patientkommunikation i en gemensam regional plattform för att minska dubbelarbete och förbättra informationsflödet.", source: "Regionalt mandatprogram M+MP+L+KD+C 2022–2026", status: "active" },
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
      { name: "Kultur & fritid",  value: 1.2, pct: 4.8  },
      { name: "Övrigt",           value: 3.2, pct: 12.9 },
    ],
  },
  agenda: [
    { title: "Fler lärare per elev", description: "Rekrytera 80 nya lärartjänster och minska klasstorleken från 26 till 22 elever per klass. Fokus på skolor med lägst meritvärden i Malmö.", source: "Mandatprogrammet S+MP+V 2022–2026", status: "in_progress" },
    { title: "Trygghetsskapande belysning", description: "Uppgradera gatubelysningen i Rosengård, Herrgården och Husie under 2024–2025 med LED-armaturer och rörelsestyrning för att förbättra den upplevda tryggheten.", source: "Mandatprogrammet S+MP+V 2022–2026", status: "in_progress" },
    { title: "Snabbare bygglov", description: "Korta handläggningstiden för standardärenden från 11 till 6 veckor genom digitalt system och ökad bemanning på stadsbyggnadskontoret.", source: "Mandatprogrammet S+MP+V 2022–2026", status: "active" },
    { title: "Fossilfritt Malmö 2030", description: "Kommunens fordonsflotta ska vara 100 % fossilfri 2026. Solceller på alla kommunala tak 2028. Nettopositiv koldioxidbalans 2030 enligt antagen klimatplan.", source: "Malmö stads klimatplan 2022–2030", status: "in_progress" },
  ],
  kpis: [
    { label: "Lärare per 100 elever",  description: "Antal heltidsanställda lärare per 100 elever i kommunens grundskolor.",                      value: "8,1",  raw: 8.1, target: 8.5, worseHigher: false, unit: "",  trend: "flat", delta: "±0",    note: "Mål: 8,5"        },
    { label: "Handläggning bygglov",   description: "Genomsnittlig handläggningstid i veckor för standardärenden på plan- och byggavdelningen.",   value: "11 v", raw: 11,  target: 10,  worseHigher: true,  unit: "v", trend: "down", delta: "−2 v",  note: "Mål: under 10 v" },
    { label: "Otrygghet (SCB)",        description: "Andel invånare som uppger att de känner sig otrygga i sitt bostadsområde (SCB trygghetsmätning).", value: "28 %", raw: 28,  target: 22,  worseHigher: true,  unit: "%", trend: "down", delta: "−3 pp", note: "Mål: 22 %"       },
  ],
};
