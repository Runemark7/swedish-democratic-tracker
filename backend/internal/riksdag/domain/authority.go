package domain

type YearlyExpenditure struct {
	Year            int     `json:"year"`
	ExpenditureMdkr float64 `json:"expenditureMdkr"`
	BudgetMdkr      float64 `json:"budgetMdkr"`
}

type YearlyHeadcount struct {
	Year         int `json:"year"`
	HeadcountInt int `json:"headcountInt"`
}

type Authority struct {
	Slug             string               `json:"slug"`
	Name             string               `json:"name"`
	Role             string               `json:"role"`
	Ministry         string               `json:"ministry"`
	Headcount        string               `json:"headcount"`
	HeadcountInt     int                  `json:"headcountInt"`
	Description      string               `json:"description"`
	WebsiteURL       string               `json:"websiteUrl"`
	AnnualReportURL  string               `json:"annualReportUrl"`
	ExpenditureMdkr  float64              `json:"expenditureMdkr"`
	BudgetMdkr       float64              `json:"budgetMdkr"`
	Year             int                  `json:"year"`
	History          []YearlyExpenditure  `json:"history"`
	HeadcountHistory []YearlyHeadcount    `json:"headcountHistory"`
}

type Regleringsbrev struct {
	Year    int    `json:"year"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	URL     string `json:"url"`
}

type AgencyDecision struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	DocType string `json:"docType"`
	Summary string `json:"summary"`
	URL     string `json:"url"`
}

type AuthorityDetail struct {
	Authority
	Mandate         string           `json:"mandate"`
	Regleringsbrev  []Regleringsbrev `json:"regleringsbrev"`
	RecentDecisions []AgencyDecision `json:"recentDecisions"`
}
