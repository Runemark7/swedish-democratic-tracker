package ports

import "context"

type YearlyExpenditure struct {
	Year            int
	ExpenditureMdkr float64
}

type AuthorityData struct {
	Name            string
	Role            string
	Ministry        string
	Headcount       string
	HeadcountInt    int
	Description     string
	WebsiteURL      string
	AnnualReportURL string
	ExpenditureMdkr float64
	Year            int
	History         []YearlyExpenditure
}

type AuthorityClient interface {
	FetchAuthorities(ctx context.Context) ([]AuthorityData, error)
}

type HeadcountData struct {
	Name         string
	HeadcountInt int
	Year         int
}

type HeadcountClient interface {
	FetchHeadcounts(ctx context.Context) ([]HeadcountData, error)
}
