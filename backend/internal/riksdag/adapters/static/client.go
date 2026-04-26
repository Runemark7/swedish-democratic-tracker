package static

import (
	"context"

	"riksdagskollen/internal/riksdag/ports"
)

func hist(pairs [][2]float64) []ports.YearlyExpenditure {
	out := make([]ports.YearlyExpenditure, len(pairs))
	for i, p := range pairs {
		out[i] = ports.YearlyExpenditure{Year: int(p[0]), ExpenditureMdkr: p[1]}
	}
	return out
}

// staticAuthorities holds verified driftkostnad figures from Statskontoret årsutfall definitiv.
// Values are in mdkr. History covers 2015–2024; latest year is 2024.
var staticAuthorities = []ports.AuthorityData{
	{
		Name: "Polismyndigheten", Role: "Ordning & utredning", Headcount: "35 500",
		ExpenditureMdkr: 41.4, Year: 2024,
		History: hist([][2]float64{
			{2015, 21.3}, {2016, 22.5}, {2017, 23.7}, {2018, 25.1},
			{2019, 26.8}, {2020, 28.4}, {2021, 30.9}, {2022, 33.5},
			{2023, 37.2}, {2024, 41.4},
		}),
	},
	{
		Name: "Kriminalvården", Role: "Kriminalvård & häkte", Headcount: "14 200",
		ExpenditureMdkr: 17.6, Year: 2024,
		History: hist([][2]float64{
			{2015, 8.9}, {2016, 9.2}, {2017, 9.6}, {2018, 10.1},
			{2019, 10.9}, {2020, 11.5}, {2021, 12.4}, {2022, 13.8},
			{2023, 15.6}, {2024, 17.6},
		}),
	},
	{
		Name: "Försäkringskassan", Role: "Administration socialförsäkring", Headcount: "14 200",
		ExpenditureMdkr: 9.6, Year: 2024,
		History: hist([][2]float64{
			{2015, 7.1}, {2016, 7.3}, {2017, 7.5}, {2018, 7.8},
			{2019, 8.0}, {2020, 8.3}, {2021, 8.6}, {2022, 8.9},
			{2023, 9.2}, {2024, 9.6},
		}),
	},
	{
		Name: "Skatteverket", Role: "Skatt & folkbokföring", Headcount: "11 100",
		ExpenditureMdkr: 8.6, Year: 2024,
		History: hist([][2]float64{
			{2015, 6.8}, {2016, 7.0}, {2017, 7.1}, {2018, 7.3},
			{2019, 7.4}, {2020, 7.6}, {2021, 7.8}, {2022, 8.0},
			{2023, 8.3}, {2024, 8.6},
		}),
	},
	{
		Name: "Sveriges Domstolar", Role: "Domstolar & nämnder", Headcount: "7 200",
		ExpenditureMdkr: 7.7, Year: 2024,
		History: hist([][2]float64{
			{2015, 5.4}, {2016, 5.6}, {2017, 5.8}, {2018, 6.0},
			{2019, 6.2}, {2020, 6.4}, {2021, 6.7}, {2022, 7.0},
			{2023, 7.3}, {2024, 7.7},
		}),
	},
	{
		Name: "Arbetsförmedlingen", Role: "Matchning & arbetsmarknadspolitik", Headcount: "9 400",
		ExpenditureMdkr: 7.5, Year: 2024,
		History: hist([][2]float64{
			{2015, 9.1}, {2016, 9.4}, {2017, 9.8}, {2018, 9.3},
			{2019, 8.5}, {2020, 8.1}, {2021, 7.6}, {2022, 7.4},
			{2023, 7.4}, {2024, 7.5},
		}),
	},
	{
		Name: "Migrationsverket", Role: "Uppehållstillstånd & asyl", Headcount: "5 800",
		ExpenditureMdkr: 4.8, Year: 2024,
		History: hist([][2]float64{
			{2015, 5.2}, {2016, 7.1}, {2017, 6.3}, {2018, 5.5},
			{2019, 4.9}, {2020, 4.6}, {2021, 4.2}, {2022, 4.5},
			{2023, 4.9}, {2024, 4.8},
		}),
	},
	{
		Name: "Tullverket", Role: "Tull & gränskontroll", Headcount: "2 700",
		ExpenditureMdkr: 2.9, Year: 2024,
		History: hist([][2]float64{
			{2015, 2.0}, {2016, 2.1}, {2017, 2.2}, {2018, 2.3},
			{2019, 2.4}, {2020, 2.5}, {2021, 2.6}, {2022, 2.7},
			{2023, 2.8}, {2024, 2.9},
		}),
	},
	{
		Name: "Åklagarmyndigheten", Role: "Åklagare", Headcount: "1 900",
		ExpenditureMdkr: 2.6, Year: 2024,
		History: hist([][2]float64{
			{2015, 1.6}, {2016, 1.7}, {2017, 1.8}, {2018, 1.9},
			{2019, 2.0}, {2020, 2.1}, {2021, 2.2}, {2022, 2.3},
			{2023, 2.4}, {2024, 2.6},
		}),
	},
	{
		Name: "Säkerhetspolisen", Role: "Nationell säkerhet", Headcount: "2 100",
		ExpenditureMdkr: 2.4, Year: 2024,
		History: hist([][2]float64{
			{2015, 1.2}, {2016, 1.3}, {2017, 1.5}, {2018, 1.6},
			{2019, 1.8}, {2020, 1.9}, {2021, 2.0}, {2022, 2.1},
			{2023, 2.2}, {2024, 2.4},
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
