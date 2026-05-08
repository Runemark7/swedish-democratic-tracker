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
	ID             int            `json:"id"`
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
}

// DocumentStatus represents the parsed response from /dokumentstatus/{dok_id}.json.
type DocumentStatus struct {
	DokID      string              `json:"dokId"`
	Title      string              `json:"title"`
	Type       string              `json:"type"`
	References []DocumentReference `json:"references"`
	Date       string
	Subtitle   string
	Summary    string
	// BodyHTML is the full document content from /dokumentstatus.dokument.html.
	// Inline-rendered on the BeslutDetailPage so users see the proposal,
	// motivation and debate transcript without leaving the site.
	BodyHTML   string `json:"bodyHtml,omitempty"`
}

type DocumentReference struct {
	RefDokTyp string `json:"refDokTyp"` // "prop" or "mot"
	RefDokID  string `json:"refDokId"`
	PartyBet  string `json:"partyBet"` // proposing party (for mot)
}
