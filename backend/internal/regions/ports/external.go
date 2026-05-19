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
	FetchRegionBudget(ctx context.Context, regionCode string, year int) ([]RegionBudgetArea, error)
	FetchRegionBudgetMultiYear(ctx context.Context, regionCode string, years []int) ([]RegionBudgetSnapshot, error)
	FetchAreaAcrossRegions(ctx context.Context, areaName string, regionCodes []string, year int) ([]RegionAreaDataPoint, error)
}

type RegionBudgetSnapshot struct {
	RegionCode string  `json:"region_code"`
	AreaName   string  `json:"area_name"`
	Year       int     `json:"year"`
	ValueMnkr  float64 `json:"value_mnkr"`
	TotalMnkr  float64 `json:"total_mnkr"`
	Pct        float64 `json:"pct"`
}

type RegionAreaDataPoint struct {
	RegionCode string  `json:"region_code"`
	ValueMnkr  float64 `json:"value_mnkr"`
	TotalMnkr  float64 `json:"total_mnkr"`
	Pct        float64 `json:"pct"`
}

type PopulationEntry struct {
	Year       int `json:"year"`
	Population int `json:"population"`
}

type RegionBudgetArea struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Pct   float64 `json:"pct"`
}

type TEDClient interface {
	FetchProcurement(ctx context.Context, buyerName string) ([]ProcurementNotice, error)
}

type ProcurementNotice struct {
	CPVDivision string
	Value       float64
	Currency    string
	Year        int
}

type ProcurementCategorySummary struct {
	CPVDivision string  `json:"cpv_division"`
	Label       string  `json:"label"`
	TotalValueSEK float64 `json:"total_value_sek"`
	Pct         float64 `json:"pct"`
	Count       int     `json:"count"`
}
