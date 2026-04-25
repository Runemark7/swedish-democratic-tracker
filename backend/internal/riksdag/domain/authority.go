package domain

type Authority struct {
	Name            string  `json:"name"`
	Role            string  `json:"role"`
	Headcount       string  `json:"headcount"`
	ExpenditureMdkr float64 `json:"expenditureMdkr"`
	Year            int     `json:"year"`
}
