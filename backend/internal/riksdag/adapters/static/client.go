package static

import (
	"context"

	"riksdagskollen/internal/riksdag/ports"
)

// staticAuthorities holds verified 2024 driftkostnad figures from Statskontoret årsutfall definitiv.
// Values are in mdkr (miljarder kronor). These are appropriation outturns, not transfer payments.
var staticAuthorities = []ports.AuthorityData{
	{Name: "Polismyndigheten", Role: "Ordning & utredning", Headcount: "35 500", ExpenditureMdkr: 41.4, Year: 2024},
	{Name: "Kriminalvården", Role: "Kriminalvård & häkte", Headcount: "14 200", ExpenditureMdkr: 17.6, Year: 2024},
	{Name: "Försäkringskassan", Role: "Administration socialförsäkring", Headcount: "14 200", ExpenditureMdkr: 9.6, Year: 2024},
	{Name: "Skatteverket", Role: "Skatt & folkbokföring", Headcount: "11 100", ExpenditureMdkr: 8.6, Year: 2024},
	{Name: "Sveriges Domstolar", Role: "Domstolar & nämnder", Headcount: "7 200", ExpenditureMdkr: 7.7, Year: 2024},
	{Name: "Arbetsförmedlingen", Role: "Matchning & arbetsmarknadspolitik", Headcount: "9 400", ExpenditureMdkr: 7.5, Year: 2024},
	{Name: "Migrationsverket", Role: "Uppehållstillstånd & asyl", Headcount: "5 800", ExpenditureMdkr: 4.8, Year: 2024},
	{Name: "Tullverket", Role: "Tull & gränskontroll", Headcount: "2 700", ExpenditureMdkr: 2.9, Year: 2024},
	{Name: "Åklagarmyndigheten", Role: "Åklagare", Headcount: "1 900", ExpenditureMdkr: 2.6, Year: 2024},
	{Name: "Säkerhetspolisen", Role: "Nationell säkerhet", Headcount: "2 100", ExpenditureMdkr: 2.4, Year: 2024},
}

type Client struct{}

func NewClient() *Client { return &Client{} }

func (c *Client) FetchAuthorities(_ context.Context) ([]ports.AuthorityData, error) {
	result := make([]ports.AuthorityData, len(staticAuthorities))
	copy(result, staticAuthorities)
	return result, nil
}
