package domain

// Committee is one utskott as it appears in the record for a mandate period.
//
// Derived from the beteckningar actually present, not from a fixed list: a
// decision belongs to whoever decided it, so a committee that existed then
// appears then, and one that exists now but decided nothing does not.
type Committee struct {
	Code string `json:"code"`
	Name string `json:"name"`
	// Voteringar is how many voteringar the committee decided in the period,
	// counted as the record identifies them (one votering_id each). A
	// förslagspunkt decided by two voteringar counts twice because the record
	// holds two. A count of the record, not a measure of importance.
	Voteringar int `json:"voteringar"`
	// ExpenditureAreas are the utgiftsområden the committee bereder, per the
	// Bilaga to riksdagsordningen. Empty for a joint committee, which handles
	// no expenditure area of its own, and empty on the list endpoint, which
	// does not query them per committee.
	ExpenditureAreas []ExpenditureArea `json:"expenditureAreas"`
}

// ExpenditureArea is one utgiftsområde with the amount allocated to it.
//
// Never carries a share of the total or a rank. A figure that invites "who is
// bigger" is where bias enters even when every number is correct, and a share
// would misrepresent committees whose remit exceeds their areas — SkU handles
// all taxation while reading as 1,0 % of expenditure, because revenue is not
// an utgiftsområde.
type ExpenditureArea struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	AmountKsek int64  `json:"amountKsek"`
}
