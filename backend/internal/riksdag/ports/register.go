package ports

import (
	"context"
	"time"

	"riksdagskollen/internal/riksdag/domain"
)

// RegisterEntry is one row scraped from SCB Myndighetsregistret.
type RegisterEntry struct {
	OrgNumber       string
	Slug            string
	Name            string
	Type            string
	PrincipalBody   string
	UnderGovernment bool
	Website         string
	SFS             string
}

type RegisterClient interface {
	FetchRegister(ctx context.Context) ([]RegisterEntry, error)
}

type AuthorityFilter struct {
	Query           string
	UnderGovernment *bool
	Limit           int
	Offset          int
}

// YearlyHeadcountSCB is one year's headcount for an agency from
// Statskontoret's Myndighetsförteckning.
type YearlyHeadcountSCB struct {
	Year         int
	HeadcountInt int
}

// HeadcountEntry is one agency's enrichment record from Myndighetsförteckning.
type HeadcountEntry struct {
	OrgNumber  string // canonical "NNNNNN-NNNN"
	Name       string // for logging / unmatched-row diagnostics
	Department string
	Year       int                  // latest year
	Headcount  int                  // latest value
	History    []YearlyHeadcountSCB // chronological, ascending
}

// HeadcountSCBClient fetches per-agency headcount + department from
// Statskontoret's Myndighetsförteckning. Named distinctly from the legacy
// HeadcountClient in external.go (which returns []HeadcountData from SCB KLS).
type HeadcountSCBClient interface {
	FetchHeadcounts(ctx context.Context) ([]HeadcountEntry, error)
}

// Enrichment is the update payload for non-register columns. Phase 1's
// UpsertAuthorities never writes these; this method only updates rows whose
// org_number already exists (no INSERT — pre-existing register row required).
type Enrichment struct {
	OrgNumber    string
	Department   string
	HeadcountInt int
	Year         int
	History      []YearlyHeadcountSCB
}

type AuthorityRepository interface {
	UpsertAuthorities(ctx context.Context, items []domain.RegisteredAuthority) (int, error)
	List(ctx context.Context, f AuthorityFilter) ([]domain.RegisteredAuthority, error)
	Count(ctx context.Context, f AuthorityFilter) (int, error)
	GetBySlug(ctx context.Context, slug string) (*domain.RegisteredAuthority, error)
	Stats(ctx context.Context) (*AuthorityStats, error)
	UpdateEnrichment(ctx context.Context, items []Enrichment) (matched, unmatched int, err error)
	UpdateExpenditure(ctx context.Context, items []ExpenditureUpdate) (matched, unmatched int, err error)
}

// AuthorityStats is the aggregated overview shown above the searchable list
// page. Coverage = how many of the 448 registered agencies have non-NULL
// expenditure / headcount; sums are over the populated rows only.
type AuthorityStats struct {
	Total                int       `json:"total"`
	UnderGovernment      int       `json:"underGovernment"`
	WithExpenditure      int       `json:"withExpenditure"`
	WithHeadcount        int       `json:"withHeadcount"`
	TotalExpenditureMdkr float64   `json:"totalExpenditureMdkr"`
	TotalHeadcount       int       `json:"totalHeadcount"`
	LatestUpdatedAt      time.Time `json:"latestUpdatedAt"`
}

// YearlyExpenditureSCB is one year's outcome + budget for an agency, in mdkr.
type YearlyExpenditureSCB struct {
	Year            int
	ExpenditureMdkr float64
	BudgetMdkr      float64
}

// AnslagYearly is one (anslag, year) row from Statskontoret's årsutfall CSV.
// AnslagName is the official anslag title — often the owning agency's name
// (e.g. "Polismyndigheten" for 0401001), sometimes a topic ("Allmänna val
// och demokrati"). The Phase 3.1 worker only attributes anslag whose
// AnslagName matches a registered agency name.
type AnslagYearly struct {
	Anslag          string
	AnslagName      string
	Year            int
	ExpenditureMdkr float64
	BudgetMdkr      float64
}

// AllAnslagClient yields every anslag-year row from the latest available
// Statskontoret year-outcome publication.
type AllAnslagClient interface {
	FetchAllAnslagYearly(ctx context.Context) ([]AnslagYearly, error)
}

// ExpenditureUpdate is the update payload for the expenditure columns of an
// authorities row. Like Enrichment, it is UPDATE-only — rows whose org_number
// does not exist are skipped (counted as unmatched).
type ExpenditureUpdate struct {
	OrgNumber       string
	ExpenditureMdkr float64
	BudgetMdkr      float64
	Year            int
	History         []YearlyExpenditureSCB
}
