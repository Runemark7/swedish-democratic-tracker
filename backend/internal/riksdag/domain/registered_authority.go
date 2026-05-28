package domain

import "time"

// RegisteredAuthority is a state agency as stored in the authorities table.
// Expenditure/headcount are pointers: nil means "no source data yet" (rendered
// as "saknas"), distinct from a real zero.
type RegisteredAuthority struct {
	OrgNumber       string    `json:"orgNumber"`
	Slug            string    `json:"slug"`
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	PrincipalBody   string    `json:"principalBody"`
	Department      string    `json:"department"`
	UnderGovernment bool      `json:"underGovernment"`
	Website         string    `json:"website,omitempty"`
	SFS             string    `json:"sfs,omitempty"`
	ExpenditureMdkr *float64  `json:"expenditureMdkr,omitempty"`
	BudgetMdkr      *float64  `json:"budgetMdkr,omitempty"`
	HeadcountInt    *int      `json:"headcountInt,omitempty"`
	Year            int       `json:"year"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
