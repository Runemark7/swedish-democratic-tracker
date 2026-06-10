package domain

type Kpi struct {
	ID          int     `json:"id"`
	Label       string  `json:"label"`
	Description string  `json:"description"`
	Raw         float64 `json:"raw"`
	WorseHigher bool    `json:"worseHigher"`
	Unit        string  `json:"unit"`
	Trend       string  `json:"trend"`
	Delta       string  `json:"delta"`
	Note        string  `json:"note"`
	SourceURL   string  `json:"sourceUrl,omitempty"`
	Year        int     `json:"year"`
}
