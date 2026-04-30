package ports

import (
	"context"
	"time"

	"riksdagskollen/internal/riksdag/domain"
)

type YearlyExpenditure struct {
	Year            int
	ExpenditureMdkr float64
	BudgetMdkr      float64 // Statens budget + Ändringsbudgetar (col6+col7 in CSV)
}

type AuthorityData struct {
	Name             string
	Role             string
	Ministry         string
	Headcount        string
	HeadcountInt     int
	Description      string
	Mandate          string
	MandateURL       string
	StatsliggarenID  int // numeric ID in statskontoret.se/statsliggaren
	ActiveSince      int // first year with regleringsbrev in Statsliggaren
	WebsiteURL       string
	AnnualReportURL  string
	ExpenditureMdkr  float64
	BudgetMdkr       float64
	Year             int
	History          []YearlyExpenditure
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

type RiksdagenDoc struct {
	DokID   string
	Year    int
	Date    time.Time
	Title   string
	Summary string
	DocType string
	URL     string
}

type AgencyDocClient interface {
	FetchRegleringsbrev(ctx context.Context, agencyName string) ([]RiksdagenDoc, error)
	FetchDecisions(ctx context.Context, agencyName string) ([]RiksdagenDoc, error)
}

type AgencyIntelRepository interface {
	UpsertRegleringsbrev(ctx context.Context, slug string, docs []RiksdagenDoc) error
	UpsertDecisions(ctx context.Context, slug string, docs []RiksdagenDoc) error
	GetRegleringsbrev(ctx context.Context, slug string) ([]domain.Regleringsbrev, error)
	GetDecisions(ctx context.Context, slug string) ([]domain.AgencyDecision, error)
}

type KpiRepository interface {
	ListKpis(ctx context.Context) ([]domain.Kpi, error)
}

type GovRepository interface {
	GetCurrentGovernment(ctx context.Context) (*domain.Government, error)
}

type AgendaRepository interface {
	ListAgenda(ctx context.Context) ([]domain.AgendaItem, error)
}

type LiveVotesRepository interface {
	ListLiveVotes(ctx context.Context, limit int) ([]domain.LiveVote, error)
}
