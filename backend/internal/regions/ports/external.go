package ports

import "context"

type KoladaClient interface {
	FetchKPIs(ctx context.Context, munCode string, kpiCodes []string, years []int) ([]KPIValue, error)
}

type KPIValue struct {
	KPI    string  `json:"kpi"`
	Year   int     `json:"year"`
	Value  float64 `json:"value"`
	Status string  `json:"status"`
}

type SCBClient interface {
	FetchPopulationTrend(ctx context.Context, munCode string, years []int) ([]PopulationEntry, error)
}

type PopulationEntry struct {
	Year       int `json:"year"`
	Population int `json:"population"`
}
