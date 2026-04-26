package domain

type YearlyExpenditure struct {
	Year            int     `json:"year"`
	ExpenditureMdkr float64 `json:"expenditureMdkr"`
}

type Authority struct {
	Slug            string              `json:"slug"`
	Name            string              `json:"name"`
	Role            string              `json:"role"`
	Ministry        string              `json:"ministry"`
	Headcount       string              `json:"headcount"`
	HeadcountInt    int                 `json:"headcountInt"`
	Description     string              `json:"description"`
	WebsiteURL      string              `json:"websiteUrl"`
	AnnualReportURL string              `json:"annualReportUrl"`
	ExpenditureMdkr float64             `json:"expenditureMdkr"`
	Year            int                 `json:"year"`
	History         []YearlyExpenditure `json:"history"`
}

type AuthorityDetail struct {
	Authority
	RegleringsbrevURL string `json:"regleringsbrevUrl"`
}
