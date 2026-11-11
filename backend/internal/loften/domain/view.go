package domain

import (
	"strconv"
	"time"
)

// ConfirmedKoppling is a Bekräftad koppling as read for the public: the
// facts about the votering plus the party's ballots in it.
type ConfirmedKoppling struct {
	VoteInput
	LoefteID      int
	Beteckning    string
	Forslagspunkt string
	Betankande    string
	VoteDate      *time.Time
}

// KopplingView is the public shape of one Bekräftad koppling.
type KopplingView struct {
	ID            int    `json:"id"`
	LoefteID      int    `json:"loefteId"`
	URL           string `json:"url"`
	VoteringID    string `json:"voteringId"`
	Beteckning    string `json:"beteckning"`
	Forslagspunkt string `json:"forslagspunkt"`
	Session       string `json:"session"`
	Betankande    string `json:"betankande,omitempty"`
	VoteDate      string `json:"voteDate,omitempty"`
	// Direction is the ballot a person confirmed to be in the Löfte's direction.
	Direction  string `json:"direction"`
	Partilinje string `json:"partilinje,omitempty"`
	// Tally is the party's ballots; Frånvarande is shown but never counted.
	Ja          int    `json:"ja"`
	Nej         int    `json:"nej"`
	Avstar      int    `json:"avstar"`
	Franvarande int    `json:"franvarande"`
	Effect      Effect `json:"effect"`
	Explanation string `json:"explanation"`
	ConfirmedAt string `json:"confirmedAt"`
}

// LoefteView is a Löfte with its Löftesstatus. Status is absent for archived
// Löften. Kopplingar is only filled on the single-Löfte read.
type LoefteView struct {
	Loefte
	Status          *Status        `json:"status"`
	StatusChangedAt *string        `json:"statusChangedAt"`
	StatusReason    string         `json:"statusReason,omitempty"`
	Kopplingar      []KopplingView `json:"kopplingar,omitempty"`
}

// PartySummary counts a party's public Löften per Löftesstatus.
type PartySummary struct {
	Party      string `json:"party"`
	Total      int    `json:"total"`
	For        int    `json:"for"`
	Emot       int    `json:"emot"`
	Blandat    int    `json:"blandat"`
	InteProvat int    `json:"inteProvat"`
	Arkiv      int    `json:"arkiv"`
}

func KopplingURL(id int) string { return "/kopplingar/" + strconv.Itoa(id) }

// ToView converts an evaluated vote for the API.
func ToView(ev EvaluatedVote, c ConfirmedKoppling) KopplingView {
	v := KopplingView{
		ID: ev.KopplingID, LoefteID: c.LoefteID, URL: KopplingURL(ev.KopplingID), VoteringID: ev.VoteringID,
		Beteckning: c.Beteckning, Forslagspunkt: c.Forslagspunkt, Session: ev.Session, Betankande: c.Betankande,
		Direction: ev.Direction, Partilinje: string(ev.Partilinje),
		Ja: ev.Tally.Ja, Nej: ev.Tally.Nej, Avstar: ev.Tally.Avstar, Franvarande: ev.Tally.Franvarande,
		Effect: ev.Effect, Explanation: ev.Explanation, ConfirmedAt: ev.ConfirmedAt.UTC().Format(time.RFC3339),
	}
	if c.VoteDate != nil {
		v.VoteDate = c.VoteDate.UTC().Format("2006-01-02")
	}
	return v
}
