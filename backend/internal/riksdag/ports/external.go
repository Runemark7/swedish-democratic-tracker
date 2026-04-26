package ports

import "context"

type YearlyExpenditure struct {
	Year            int
	ExpenditureMdkr float64
}

type AuthorityData struct {
	Name            string
	Role            string
	Headcount       string
	ExpenditureMdkr float64
	Year            int
	History         []YearlyExpenditure
}

type AuthorityClient interface {
	FetchAuthorities(ctx context.Context) ([]AuthorityData, error)
}
