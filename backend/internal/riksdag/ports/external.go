package ports

import (
	"context"

	"riksdagskollen/internal/riksdag/domain"
)

type YearlyExpenditure struct {
	Year            int
	ExpenditureMdkr float64
	BudgetMdkr      float64 // Statens budget + Ändringsbudgetar (col6+col7 in CSV)
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
	BudgetMdkr      float64
	Year            int
	History         []YearlyExpenditure
}

type AuthorityClient interface {
	FetchAuthorities(ctx context.Context) ([]AuthorityData, error)
}

type YearlyHeadcount struct {
	Year         int
	HeadcountInt int
}

type HeadcountData struct {
	Name             string
	HeadcountInt     int
	Year             int
	HeadcountHistory []YearlyHeadcount
}

type HeadcountClient interface {
	FetchHeadcounts(ctx context.Context) ([]HeadcountData, error)
}

type KpiRepository interface {
	ListKpis(ctx context.Context) ([]domain.Kpi, error)
}
