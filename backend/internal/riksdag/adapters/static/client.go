package static

import (
	"context"

	"riksdagskollen/internal/riksdag/ports"
)

// hist converts {year, utfall_mdkr, budget_mdkr} triples to []YearlyExpenditure.
func hist(triples [][3]float64) []ports.YearlyExpenditure {
	out := make([]ports.YearlyExpenditure, len(triples))
	for i, t := range triples {
		out[i] = ports.YearlyExpenditure{Year: int(t[0]), ExpenditureMdkr: t[1], BudgetMdkr: t[2]}
	}
	return out
}

// staticAuthorities holds verified utfall + anslag figures from Statskontoret årsutfall definitiv.
// Columns: year, utfall (mdkr), budget/anslag (mdkr). History covers 2015–2024.
var staticAuthorities = []ports.AuthorityData{
	{
		Name: "Polismyndigheten", Role: "Ordning & utredning",
		Ministry: "Justitiedepartementet", Headcount: "35 500", HeadcountInt: 35500,
		Description:     "Sveriges största myndighet, ansvarar för brottsbekämpning, utredning och ordningshållning.",
		Mandate:         "Inrättades 2015 genom sammanslagning av 21 polismyndigheter (polislagen 1984:387). Nationell civil myndighet med ansvar för brottsförebyggande arbete, brottsutredning och ordningshållning i hela landet.",
		MandateURL:      "https://www.riksdagen.se/sv/dokument-och-lagar/dokument/svensk-forfattningssamling/polislag-1984387_sfs-1984-387/",
		WebsiteURL:      "https://polisen.se",
		AnnualReportURL: "https://polisen.se/aktuellt/publikationer/",
		ExpenditureMdkr: 41.38, BudgetMdkr: 41.17, Year: 2024,
		History: hist([][3]float64{
			{2015, 20.76, 21.16}, {2016, 21.88, 21.84}, {2017, 22.91, 22.62}, {2018, 23.77, 24.72},
			{2019, 26.55, 26.11}, {2020, 28.80, 28.55}, {2021, 30.98, 30.99}, {2022, 33.31, 33.84},
			{2023, 37.43, 37.04}, {2024, 41.38, 41.17},
		}),
	},
	{
		Name: "Kriminalvården", Role: "Kriminalvård & häkte",
		Ministry: "Justitiedepartementet", Headcount: "14 200", HeadcountInt: 14200,
		Description:     "Ansvarar för häkten, fängelser och frivård med målet att minska återfall i brott.",
		Mandate:         "Grundad 1974 (kriminalvårdslagen). Verkställer fängelsedomar, häktning och skyddstillsyn med uppdrag att minska återfall i brott och bidra till ett tryggare samhälle.",
		MandateURL:      "https://www.riksdagen.se/sv/dokument-och-lagar/dokument/svensk-forfattningssamling/fangelselag-2010610_sfs-2010-610/",
		WebsiteURL:      "https://www.kriminalvarden.se",
		AnnualReportURL: "https://www.kriminalvarden.se/om-kriminalvarden/publikationer/",
		ExpenditureMdkr: 17.57, BudgetMdkr: 17.80, Year: 2024,
		History: hist([][3]float64{
			{2015, 7.97, 7.84}, {2016, 8.20, 8.12}, {2017, 8.62, 8.35}, {2018, 8.96, 8.65},
			{2019, 9.02, 9.55}, {2020, 9.59, 9.47}, {2021, 10.61, 10.65}, {2022, 11.94, 12.36},
			{2023, 14.55, 14.13}, {2024, 17.57, 17.80},
		}),
	},
	{
		Name: "Försäkringskassan", Role: "Administration socialförsäkring",
		Ministry: "Socialdepartementet", Headcount: "14 200", HeadcountInt: 14200,
		Description:     "Administrerar socialförsäkringssystemet inklusive sjukpenning, föräldrapenning och aktivitetsersättning.",
		Mandate:         "Inrättades 2005 (socialförsäkringsbalken). Administrerar den svenska socialförsäkringen — sjukpenning, föräldrapenning, aktivitetsersättning och handikappersättning — och säkerställer att rätt person får rätt ersättning i rätt tid.",
		MandateURL:      "https://www.riksdagen.se/sv/dokument-och-lagar/dokument/svensk-forfattningssamling/socialforsakringsbalk-2010110_sfs-2010-110/",
		WebsiteURL:      "https://www.forsakringskassan.se",
		AnnualReportURL: "https://www.forsakringskassan.se/om-forsakringskassan/publikationer",
		ExpenditureMdkr: 9.63, BudgetMdkr: 10.47, Year: 2024,
		History: hist([][3]float64{
			{2015, 7.82, 8.03}, {2016, 8.26, 8.20}, {2017, 8.44, 8.42}, {2018, 8.47, 8.73},
			{2019, 8.90, 8.64}, {2020, 9.14, 9.39}, {2021, 9.26, 9.25}, {2022, 9.47, 9.31},
			{2023, 9.83, 9.28}, {2024, 9.63, 10.47},
		}),
	},
	{
		Name: "Skatteverket", Role: "Skatt & folkbokföring",
		Ministry: "Finansdepartementet", Headcount: "11 100", HeadcountInt: 11100,
		Description:     "Ansvarar för beskattning, folkbokföring och bouppteckningar i hela Sverige.",
		Mandate:         "Inrättades 2004 genom sammanslagning av Riksskatteverket och de tio länsskattemyndigheterna (skatteförfarandelagen 2011:1244). Säkerställer korrekt och effektiv beskattning, folkbokföring och fastighetstaxering.",
		MandateURL:      "https://www.riksdagen.se/sv/dokument-och-lagar/dokument/svensk-forfattningssamling/skatteforfarandelag-20111244_sfs-2011-1244/",
		WebsiteURL:      "https://www.skatteverket.se",
		AnnualReportURL: "https://www.skatteverket.se/omoss/varverksamhet/styrningochuppfoljning/arsredovisning.4.48cfd212185efbb440b2d9c.html",
		ExpenditureMdkr: 8.64, BudgetMdkr: 8.67, Year: 2024,
		History: hist([][3]float64{
			{2015, 7.19, 7.09}, {2016, 7.22, 7.19}, {2017, 7.38, 7.38}, {2018, 7.63, 7.60},
			{2019, 7.55, 7.57}, {2020, 7.90, 8.17}, {2021, 8.17, 8.41}, {2022, 8.25, 8.30},
			{2023, 8.60, 8.17}, {2024, 8.64, 8.67},
		}),
	},
	{
		Name: "Sveriges Domstolar", Role: "Domstolar & nämnder",
		Ministry: "Justitiedepartementet", Headcount: "7 200", HeadcountInt: 7200,
		Description:     "Samlingsnamn för landets domstolar och nämnder — tingsrätter, hovrätter, förvaltningsrätter och Högsta domstolen.",
		Mandate:         "Regleras av rättegångsbalken (1942:740) och förvaltningsprocesslagen. Domstolsverket administrerar ca 80 domstolar och nämnder. Uppdraget är att avgöra tvister och mål opartiskt och rättssäkert inom rimlig tid.",
		MandateURL:      "https://www.riksdagen.se/sv/dokument-och-lagar/dokument/svensk-forfattningssamling/rattegangsbalk-1942740_sfs-1942-740/",
		WebsiteURL:      "https://www.domstol.se",
		AnnualReportURL: "https://www.domstol.se/om-sveriges-domstolar/statistik-styrning-och-utveckling/arsredovisning/",
		ExpenditureMdkr: 7.71, BudgetMdkr: 7.96, Year: 2024,
		History: hist([][3]float64{
			{2015, 5.32, 5.37}, {2016, 5.40, 5.42}, {2017, 5.51, 5.52}, {2018, 5.76, 5.61},
			{2019, 5.89, 5.99}, {2020, 6.14, 6.27}, {2021, 6.46, 6.47}, {2022, 6.77, 6.68},
			{2023, 7.36, 7.05}, {2024, 7.71, 7.96},
		}),
	},
	{
		Name: "Arbetsförmedlingen", Role: "Matchning & arbetsmarknadspolitik",
		Ministry: "Arbetsmarknadsdepartementet", Headcount: "9 400", HeadcountInt: 9400,
		Description:     "Ansvarar för arbetsförmedling, matchning mellan arbetsgivare och arbetssökande, och genomförande av arbetsmarknadspolitiken.",
		Mandate:         "Inrättades 2008 (lag 2007:1030). Genomför den statliga arbetsmarknadspolitiken — matchning, rustning och stöd till arbetssökande och arbetsgivare — för att bidra till en väl fungerande arbetsmarknad.",
		MandateURL:      "https://www.riksdagen.se/sv/dokument-och-lagar/dokument/svensk-forfattningssamling/lag-20071030_sfs-2007-1030/",
		WebsiteURL:      "https://www.arbetsformedlingen.se",
		AnnualReportURL: "https://arbetsformedlingen.se/om-oss/var-verksamhet/styrning-och-resultat",
		ExpenditureMdkr: 7.45, BudgetMdkr: 7.41, Year: 2024,
		History: hist([][3]float64{
			{2015, 7.47, 7.45}, {2016, 8.15, 7.96}, {2017, 8.50, 8.39}, {2018, 8.19, 8.41},
			{2019, 7.64, 7.62}, {2020, 7.52, 7.55}, {2021, 7.74, 7.85}, {2022, 7.87, 7.86},
			{2023, 7.68, 7.64}, {2024, 7.45, 7.41},
		}),
	},
	{
		Name: "Migrationsverket", Role: "Uppehållstillstånd & asyl",
		Ministry: "Justitiedepartementet", Headcount: "5 800", HeadcountInt: 5800,
		Description:     "Prövar ansökningar om uppehållstillstånd, asyl, medborgarskap och arbetstillstånd.",
		Mandate:         "Inrättades 1969 (utlänningslagen, nu 2005:716). Prövar ansökningar om uppehållstillstånd, asyl, medborgarskap och arbetstillstånd samt ansvarar för mottagning av asylsökande och återvändande.",
		MandateURL:      "https://www.riksdagen.se/sv/dokument-och-lagar/dokument/svensk-forfattningssamling/utlanningslag-2005716_sfs-2005-716/",
		WebsiteURL:      "https://www.migrationsverket.se",
		AnnualReportURL: "https://www.migrationsverket.se/om-migrationsverket/styrning-och-uppfoljning.html",
		ExpenditureMdkr: 4.79, BudgetMdkr: 4.73, Year: 2024,
		History: hist([][3]float64{
			{2015, 4.72, 4.54}, {2016, 5.98, 6.88}, {2017, 6.01, 5.90}, {2018, 5.19, 5.35},
			{2019, 4.55, 4.38}, {2020, 4.34, 4.44}, {2021, 4.09, 4.42}, {2022, 4.44, 4.32},
			{2023, 4.71, 4.72}, {2024, 4.79, 4.73},
		}),
	},
	{
		Name: "Tullverket", Role: "Tull & gränskontroll",
		Ministry: "Finansdepartementet", Headcount: "2 700", HeadcountInt: 2700,
		Description:     "Kontrollerar in- och utförsel av varor vid Sveriges gränser och bekämpar smuggling.",
		Mandate:         "Inrättades 1637, nuvarande form regleras av tullagen (2016:253). Kontrollerar varuflöden vid gränsen, uppbär tullavgifter och bekämpar smuggling av narkotika, vapen och andra otillåtna varor.",
		MandateURL:      "https://www.riksdagen.se/sv/dokument-och-lagar/dokument/svensk-forfattningssamling/tullag-2016253_sfs-2016-253/",
		WebsiteURL:      "https://www.tullverket.se",
		AnnualReportURL: "https://www.tullverket.se/omoss/dethargortullverket/verksamhetochorganisation/arsredovisning.html",
		ExpenditureMdkr: 2.92, BudgetMdkr: 2.88, Year: 2024,
		History: hist([][3]float64{
			{2015, 1.77, 1.68}, {2016, 1.72, 1.73}, {2017, 1.72, 1.74}, {2018, 1.90, 1.93},
			{2019, 1.99, 1.98}, {2020, 2.14, 2.13}, {2021, 2.28, 2.30}, {2022, 2.47, 2.48},
			{2023, 2.64, 2.59}, {2024, 2.92, 2.88},
		}),
	},
	{
		Name: "Åklagarmyndigheten", Role: "Åklagare",
		Ministry: "Justitiedepartementet", Headcount: "1 900", HeadcountInt: 1900,
		Description:     "Leder förundersökningar och väcker åtal i brottmål vid Sveriges allmänna domstolar.",
		Mandate:         "Inrättades 2005 (åklagarlag 1998:162, förordning 2004:1265). Leder förundersökningar i samarbete med polisen och beslutar om åtal vid allmänna domstolar. Säkerställer en rättssäker och effektiv lagföring av brott.",
		MandateURL:      "https://www.riksdagen.se/sv/dokument-och-lagar/dokument/svensk-forfattningssamling/forordning-20041265_sfs-2004-1265/",
		WebsiteURL:      "https://www.aklagare.se",
		AnnualReportURL: "https://www.aklagare.se/om-oss/dokument/planering-och-uppfoljning/",
		ExpenditureMdkr: 2.59, BudgetMdkr: 2.63, Year: 2024,
		History: hist([][3]float64{
			{2015, 1.36, 1.40}, {2016, 1.44, 1.45}, {2017, 1.50, 1.47}, {2018, 1.51, 1.52},
			{2019, 1.60, 1.61}, {2020, 1.69, 1.70}, {2021, 1.83, 1.81}, {2022, 2.01, 2.07},
			{2023, 2.26, 2.30}, {2024, 2.59, 2.63},
		}),
	},
	{
		Name: "Säkerhetspolisen", Role: "Nationell säkerhet",
		Ministry: "Justitiedepartementet", Headcount: "2 100", HeadcountInt: 2100,
		Description:     "Skyddar Sverige mot terrorism, spionage och andra hot mot den nationella säkerheten.",
		Mandate:         "Inrättades 1989 som självständig myndighet (polislagen 1984:387, säkerhetspolisens instruktion). Förebygger och avslöjar brott mot rikets säkerhet — terrorism, spionage, författningsskyddande verksamhet och kontraspionage.",
		MandateURL:      "https://www.riksdagen.se/sv/dokument-och-lagar/dokument/svensk-forfattningssamling/polislag-1984387_sfs-1984-387/",
		WebsiteURL:      "https://www.sakerhetspolisen.se",
		AnnualReportURL: "https://www.sakerhetspolisen.se/om-sakerhetspolisen/styrning-och-uppfoljning.html",
		ExpenditureMdkr: 2.40, BudgetMdkr: 2.45, Year: 2024,
		History: hist([][3]float64{
			{2015, 1.15, 1.14}, {2016, 1.20, 1.20}, {2017, 1.36, 1.31}, {2018, 1.47, 1.49},
			{2019, 1.55, 1.58}, {2020, 1.59, 1.68}, {2021, 1.73, 1.74}, {2022, 1.94, 1.90},
			{2023, 2.11, 2.11}, {2024, 2.40, 2.45},
		}),
	},
}

type Client struct{}

func NewClient() *Client { return &Client{} }

func (c *Client) FetchAuthorities(_ context.Context) ([]ports.AuthorityData, error) {
	result := make([]ports.AuthorityData, len(staticAuthorities))
	copy(result, staticAuthorities)
	return result, nil
}
