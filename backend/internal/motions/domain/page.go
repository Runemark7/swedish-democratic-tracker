package domain

// PunktMotion is a motion a förslagspunkt decided, with the yrkanden decided.
type PunktMotion struct {
	DokID, Riksmote, Beteckning, Title, Subtyp string
	// Parties are the distinct parties of the signers, in signing order.
	Parties  []string
	Yrkanden []int
}

// Ballot is one member's vote in a votering, with the member's name.
type Ballot struct {
	PoliticianID, Name, Party, Result string
}

// VoteringPunkt is a votering we hold, as a förslagspunkt of a betänkande we
// may not hold (fallback when the betänkande was never ingested).
type VoteringPunkt struct {
	VoteringID, Punkt, Title string
}

// BetankandeStored is a stored betänkande with what is known about its HTML.
type BetankandeStored struct {
	Betankande
	HTMLRead, HasReasoningSection, HasReservationsSection bool
}
