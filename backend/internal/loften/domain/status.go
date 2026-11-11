package domain

import (
	"fmt"
	"sort"
	"time"

	votes "riksdagskollen/internal/votes/domain"
)

// Status is Löftesstatus (ADR-002): what a party did about a Löfte, judged
// only from its votes. It is never "hållet", "brutet" or "uppfyllt".
type Status string

const (
	// StatusNone is no status at all: an archived Löfte has no Bekräftade
	// kopplingar, and a status would state our curation gap as a fact about
	// the party.
	StatusNone       Status = ""
	StatusFor        Status = "Röstat för"
	StatusEmot       Status = "Röstat emot"
	StatusBlandat    Status = "Blandat"
	StatusInteProvat Status = "Inte prövat än"
)

// Effect is how one Bekräftad koppling counts toward the status.
type Effect string

const (
	EffectFor     Effect = "for"
	EffectEmot    Effect = "emot"
	EffectNone    Effect = "ingen"   // Avstår, Delad or no cast ballots: listed, not counted
	EffectOutside Effect = "utanfor" // votering outside the Löfte's mandate period: listed, not counted
)

// VoteInput is one Bekräftad koppling with the party's ballots in its votering.
type VoteInput struct {
	KopplingID  int
	VoteringID  string
	Direction   string // "Ja" or "Nej": the ballot in the Löfte's direction, confirmed by a person
	Session     string // riksmöte of the votering
	Tally       votes.Tally
	ConfirmedAt time.Time
}

// EvaluatedVote is a VoteInput after Partilinje and the window are applied.
type EvaluatedVote struct {
	VoteInput
	Partilinje  votes.Partilinje
	Effect      Effect
	Explanation string
}

// Result is a Löftesstatus with the votes behind it and when and why it last changed.
type Result struct {
	Status       Status
	Votes        []EvaluatedVote
	ChangedAt    time.Time      // zero when the status never moved off Inte prövat än
	ChangedBy    *EvaluatedVote // the koppling whose confirmation moved it
	ChangeReason string
}

// evaluate applies the window and the party's Partilinje to one koppling.
// Avvikare never enter here: only the Partilinje does.
func evaluate(window []string, in VoteInput) EvaluatedVote {
	ev := EvaluatedVote{VoteInput: in}
	inWindow := false
	for _, s := range window {
		if s == in.Session {
			inWindow = true
		}
	}
	ev.Partilinje = votes.PartilinjeOf(in.Tally)
	line := string(ev.Partilinje)
	switch {
	case !inWindow:
		ev.Effect = EffectOutside
		ev.Explanation = fmt.Sprintf("Omröstningen hölls i riksmöte %s, utanför Löftets mandatperiod, och räknas därför inte.", in.Session)
	case ev.Partilinje == votes.PartilinjeJa || ev.Partilinje == votes.PartilinjeNej:
		other := "Nej"
		if in.Direction == "Nej" {
			other = "Ja"
		}
		if line == in.Direction {
			ev.Effect = EffectFor
			ev.Explanation = fmt.Sprintf("Redaktören har bekräftat att %s i denna omröstning går i Löftets riktning. Partilinjen var %s, så omröstningen räknas som röstat för.", in.Direction, line)
		} else {
			ev.Effect = EffectEmot
			ev.Explanation = fmt.Sprintf("Redaktören har bekräftat att %s i denna omröstning går i Löftets riktning, och %s går i motsatt riktning. Partilinjen var %s, så omröstningen räknas som röstat emot.", in.Direction, other, line)
		}
	case ev.Partilinje == votes.PartilinjeAvstar:
		ev.Effect = EffectNone
		ev.Explanation = "Partilinjen var Avstår. Det anger ingen riktning, så omröstningen räknas varken för eller emot."
	case ev.Partilinje == votes.PartilinjeDelad:
		ev.Effect = EffectNone
		ev.Explanation = "Partiet var delat: lika många röstade för två alternativ. Det anger ingen riktning, så omröstningen räknas varken för eller emot."
	default:
		ev.Effect = EffectNone
		ev.Explanation = "Ingen ledamot från partiet röstade, så det finns ingen Partilinje och omröstningen räknas inte."
	}
	return ev
}

func statusOfEvaluated(evs []EvaluatedVote) Status {
	var f, e bool
	for _, v := range evs {
		switch v.Effect {
		case EffectFor:
			f = true
		case EffectEmot:
			e = true
		}
	}
	switch {
	case f && e:
		return StatusBlandat
	case f:
		return StatusFor
	case e:
		return StatusEmot
	}
	return StatusInteProvat
}

// StatusOf computes Löftesstatus on read. archived Löften get StatusNone.
// window is the riksmöten of the Löfte's mandate period. Votes are listed in
// the order they were confirmed; the status history is replayed in that order
// to find the confirmation that last changed it.
func StatusOf(archived bool, window []string, in []VoteInput) Result {
	if archived {
		return Result{Status: StatusNone, Votes: []EvaluatedVote{}}
	}
	sorted := append([]VoteInput(nil), in...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if !sorted[i].ConfirmedAt.Equal(sorted[j].ConfirmedAt) {
			return sorted[i].ConfirmedAt.Before(sorted[j].ConfirmedAt)
		}
		return sorted[i].KopplingID < sorted[j].KopplingID
	})
	evs := make([]EvaluatedVote, len(sorted))
	for i, v := range sorted {
		evs[i] = evaluate(window, v)
	}
	r := Result{Votes: evs, Status: statusOfEvaluated(evs)}
	prev := StatusInteProvat
	for i := range evs {
		s := statusOfEvaluated(evs[:i+1])
		if s != prev {
			v := evs[i]
			r.ChangedAt, r.ChangedBy = v.ConfirmedAt, &v
			r.ChangeReason = fmt.Sprintf("Från %q till %q när koppling %d bekräftades (omröstning %s).", prev, s, v.KopplingID, v.VoteringID)
			prev = s
		}
	}
	return r
}
