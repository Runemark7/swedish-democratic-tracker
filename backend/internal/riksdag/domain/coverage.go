package domain

// RecordCoverage states how much of a mandate period's voting record the site
// actually holds.
//
// Completeness is a claim like any other, so it must be checkable. The
// denominator comes from Riksdagen's own count, not from ours: a site that
// supplies both sides of its own coverage claim is asking to be trusted rather
// than checked.
type RecordCoverage struct {
	Mandate MandatePeriod `json:"mandate"`
	// Expected is Riksdagen's count of voteringar in the period.
	Expected int `json:"expected"`
	// Ingested is how many of those we hold.
	Ingested int `json:"ingested"`
	// Unreachable are voteringar Riksdagen lists but that carry no beteckning,
	// so they cannot be fetched. Reported so a shortfall is explained rather
	// than left as an unexplained gap.
	Unreachable int `json:"unreachable"`
	// LastDecisionDate is when the newest decision we hold was taken. It comes
	// from the votering's own date, never from an ingest timestamp.
	LastDecisionDate *string          `json:"lastDecisionDate"`
	ByRiksmote       []RiksmoteRecord `json:"byRiksmote"`
}

// MandatePeriod bounds a record so votes from one parliament are never
// attributed to another.
type MandatePeriod struct {
	Code  string `json:"code"`
	Label string `json:"label"`
	// Ended reports whether the period has closed. Once it has, the record is
	// final and the incoming parliament accumulates a separate one.
	Ended bool `json:"ended"`
}

type RiksmoteRecord struct {
	Riksmote         string  `json:"riksmote"`
	Expected         int     `json:"expected"`
	Ingested         int     `json:"ingested"`
	Unreachable      int     `json:"unreachable"`
	LastDecisionDate *string `json:"lastDecisionDate"`
}
