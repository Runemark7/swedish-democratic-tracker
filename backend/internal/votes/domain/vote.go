package domain

import "time"

type VoteResult string

const (
	VoteJa          VoteResult = "Ja"
	VoteNej         VoteResult = "Nej"
	VoteAvstar      VoteResult = "Avstår"
	VoteFranvarande VoteResult = "Frånvarande"
)

type ProposalType string

const (
	ProposalProp ProposalType = "prop"
	ProposalMot  ProposalType = "mot"
	ProposalBet  ProposalType = "bet"
)

// ProposalOrigin describes who initiated the proposal being voted on.
type ProposalOrigin struct {
	ProposedByParty string       `json:"proposedByParty,omitempty"`
	ProposalType    ProposalType `json:"proposalType,omitempty"`
	ProposalDokID   string       `json:"proposalDokId,omitempty"`
	DocumentTitle   string       `json:"documentTitle,omitempty"`
}

type Vote struct {
	// No ID: a ballot is identified by the votering it belongs to and the
	// member who cast it, which is exactly the ballots primary key. The old
	// per-ballot SERIAL was a storage surrogate that reached the API without
	// ever being read -- no query filtered on it and no consumer used it.
	VoteringID     string         `json:"voteringId"`
	PoliticianID   string         `json:"politicianId"`
	Party          string         `json:"party"`
	VoteResult     VoteResult     `json:"voteResult"`
	Beteckning     string         `json:"beteckning"`
	Forslagspunkt  string         `json:"forslagspunkt"`
	Session        string         `json:"session"`
	DokID          string         `json:"dokId,omitempty"`
	ProposalOrigin ProposalOrigin `json:"proposalOrigin"`
	OriginEnriched bool           `json:"originEnriched"`
	CreatedAt      time.Time      `json:"createdAt"`
	// SystemDatum is Riksdagen's own timestamp for the record. It is the only
	// safe basis for the ingestion cursor; CreatedAt is our insert time and
	// says nothing about how current the upstream data is.
	SystemDatum time.Time `json:"systemDatum"`
}

// Intressent is a politician formally attached to a Riksdagen document.
type Intressent struct {
	IntressentID string `json:"intressentId"`
	Name         string `json:"name"`
	Party        string `json:"party"`
	Role         string `json:"role"`
}

// DocumentStatus represents the parsed response from /dokumentstatus/{dok_id}.json.
type DocumentStatus struct {
	DokID        string              `json:"dokId"`
	Title        string              `json:"title"`
	Type         string              `json:"type"`
	References   []DocumentReference `json:"references"`
	Intressenter []Intressent        `json:"intressenter,omitempty"`
	Date         string
	Subtitle     string
	Summary      string
	Beteckning   string
	// BodyHTML is the full document content from /dokumentstatus.dokument.html.
	// Inline-rendered on the BeslutDetailPage so users see the proposal,
	// motivation and debate transcript without leaving the site.
	BodyHTML string `json:"bodyHtml,omitempty"`
	// Chamber schedule + plain-language summary parsed from dokumentstatus.dokuppgift.
	// All raw Riksdagen values, shown verbatim (fact layer).
	DebattDate string `json:"debattDate,omitempty"` // dokuppgift kod=debattdatumtid, "YYYY-MM-DD HH:MM:SS"
	BeslutDate string `json:"beslutDate,omitempty"` // dokuppgift kod=beslutdatumtid, "YYYY-MM-DD HH:MM:SS"
	StatusText string `json:"statusText,omitempty"` // dokuppgift kod=statustext
	Notis      string `json:"notis,omitempty"`      // dokuppgift kod=notis, "Beslut i korthet" — raw HTML, rendered (not stripped like Summary)
}

type DocumentReference struct {
	RefDokTyp string `json:"refDokTyp"` // "prop" or "mot"
	RefDokID  string `json:"refDokId"`
	PartyBet  string `json:"partyBet"` // proposing party (for mot)
}
