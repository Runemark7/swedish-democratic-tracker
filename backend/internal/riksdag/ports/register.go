package ports

import (
	"context"

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
	UpdateEnrichment(ctx context.Context, items []Enrichment) (matched, unmatched int, err error)
}
