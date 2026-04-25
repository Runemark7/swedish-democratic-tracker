package ports

import "context"

type AuthorityData struct {
	Name            string
	Role            string
	Headcount       string
	ExpenditureMdkr float64
	Year            int
}

type AuthorityClient interface {
	FetchAuthorities(ctx context.Context) ([]AuthorityData, error)
}
