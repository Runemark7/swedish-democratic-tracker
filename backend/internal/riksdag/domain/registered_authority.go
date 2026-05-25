package domain

import "time"

// RegisteredAuthority is a state agency as stored in the authorities table.
// Expenditure/headcount are pointers: nil means "no source data yet" (rendered
// as "saknas"), distinct from a real zero.
type RegisteredAuthority struct {
	OrgNumber       string
	Slug            string
	Name            string
	Type            string
	PrincipalBody   string
	Department      string
	UnderGovernment bool
	Website         string
	SFS             string
	ExpenditureMdkr *float64
	BudgetMdkr      *float64
	HeadcountInt    *int
	Year            int
	UpdatedAt       time.Time
}
