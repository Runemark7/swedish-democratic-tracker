package domain

type YearlyExpenditure struct {
	Year            int     `json:"year"`
	ExpenditureMdkr float64 `json:"expenditureMdkr"`
}

type Authority struct {
	Name            string              `json:"name"`
	Role            string              `json:"role"`
	Headcount       string              `json:"headcount"`
	ExpenditureMdkr float64             `json:"expenditureMdkr"`
	Year            int                 `json:"year"`
	History         []YearlyExpenditure `json:"history"`
}
