package domain

import "time"

// Betankande is a betänkande the motions are handled in. Held fields are what
// Riksdagen gave.
type Betankande struct {
	DokID       string
	Riksmote    string // "2017/18"
	Beteckning  string // "SkU11"
	Title       string
	Organ       string
	SystemDatum time.Time
	Punkter     []Forslagspunkt
}

// Forslagspunkt is one numbered decision point of a betänkande.
type Forslagspunkt struct {
	Number string
	Rubrik string
	// Text is the utskottets förslag as served (HTML breaks and entities kept).
	Text       string
	Beslutstyp string // "acklamation", "röstning", as given
	VoteringID string // empty when decided without a votering
	// Alternativ is the alternative put to the vote, as Riksdagen words it
	// ("Utskottets förslag mot reservation 12 (M)").
	Alternativ string
	Vinnare    string
}

// PartyTally is one party's cast ballots in a votering.
type PartyTally struct {
	Party                        string
	Ja, Nej, Avstar, Franvarande int
}

// YrkandeDecision is the stored link from a yrkande to a förslagspunkt, with
// what is needed to show the decision.
type YrkandeDecision struct {
	YrkandeNumber int
	LinkedBy      string
	// Decision is what the punkt text says was done with this motion.
	Decision   string
	Riksmote   string
	Betankande string // beteckning, "SkU11"
	BetTitle   string
	BetDokID   string
	Punkt      Forslagspunkt
	// VoteringStored is true when the votering is one of ours (ballots held).
	VoteringStored bool
	Parties        []PartyTally
}
