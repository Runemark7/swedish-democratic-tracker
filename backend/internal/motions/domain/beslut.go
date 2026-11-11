package domain

import (
	"sort"
	"strings"
	"time"

	votes "riksdagskollen/internal/votes/domain"
)

// Values of BeslutFilter.State.
const (
	BeslutDecided  = "decided"
	BeslutUpcoming = "upcoming"
)

// BeslutFilter narrows the Beslut list. State empty means both. Within one
// field the values are alternatives; the fields combine with AND.
//
// Parties matches a betänkande that decides a motion signed by the party or
// carries a reservation of the party. It does not mean "the party voted":
// every party votes in every votering.
type BeslutFilter struct {
	State   string
	Utskott []string // "SkU"
	Parties []string // upper case
}

// BeslutCursor is the keyset position after the last row returned. Upcoming
// rows come first, soonest first; decided rows follow, newest first.
type BeslutCursor struct {
	Decided    bool
	Date       time.Time
	Riksmote   string
	Beteckning string
}

// BeslutRow is one stored Riksdagen listing row.
type BeslutRow struct {
	Riksmote, Beteckning, DokID, Title, Organ, Status string
	// DecisionDate is Riksdagen's beslutsdag ("2026-11-25"), empty when none.
	DecisionDate   string
	ChamberDecided *bool
	// SortDate is beslutsdag, else the publication date.
	SortDate time.Time
	// SQLDecided is the repository's own reading of the decided rule. The
	// service uses RiksdagDocument.Decided and treats a mismatch as a bug.
	SQLDecided bool
}

// Key is "riksmöte:beteckning".
func (r BeslutRow) Key() string { return r.Riksmote + ":" + r.Beteckning }

// BallotCount is one party's ballots with one result in one votering.
type BallotCount struct {
	VoteringID, Party, Result string
	N                         int
}

// BeslutDetail is what is known about a listed betänkande beyond Riksdagen's
// listing row.
type BeslutDetail struct {
	// Beslutstyper holds the beslutstyp of each stored förslagspunkt.
	Beslutstyper []string
	Ballots      []BallotCount
}

// PartilinjeVaries is shown for a party whose Partilinje differs between the
// betänkande's voteringar.
const PartilinjeVaries = "Varierar"

// PartyLineSummary is a party's Partilinje over a betänkande's voteringar:
// the common Partilinje when every votering gives the same, "Varierar" when
// they differ, empty when the party cast no ballot.
type PartyLineSummary struct {
	Party      string
	Partilinje string
	Voteringar int
}

// SummarisePartilinje applies the one shared Partilinje rule per votering and
// then compares the voteringar. Parties are returned in the order of order.
func SummarisePartilinje(ballots []BallotCount, order []string) []PartyLineSummary {
	type key struct{ vot, party string }
	tallies := map[key]*votes.Tally{}
	for _, b := range ballots {
		k := key{b.VoteringID, b.Party}
		t := tallies[k]
		if t == nil {
			t = &votes.Tally{}
			tallies[k] = t
		}
		for i := 0; i < b.N; i++ {
			t.Add(votes.VoteResult(b.Result))
		}
	}
	lines := map[string]map[string]bool{}
	counts := map[string]int{}
	for k, t := range tallies {
		l := votes.PartilinjeFor(k.party, *t)
		if l == votes.PartilinjeSaknas {
			continue
		}
		if lines[k.party] == nil {
			lines[k.party] = map[string]bool{}
		}
		lines[k.party][string(l)] = true
		counts[k.party]++
	}
	out := make([]PartyLineSummary, 0, len(order))
	for _, p := range order {
		s := PartyLineSummary{Party: p, Voteringar: counts[p]}
		switch ls := lines[p]; len(ls) {
		case 0:
		case 1:
			for l := range ls {
				s.Partilinje = l
			}
		default:
			s.Partilinje = PartilinjeVaries
		}
		out = append(out, s)
	}
	return out
}

// RiksdagParties are the eight parties in the Riksdag, in the order the site
// shows them.
var RiksdagParties = []string{"S", "M", "SD", "C", "V", "KD", "MP", "L"}

// BeslutOutcome words the decision of a betänkande from how its punkter were
// decided. It says nothing about which side won.
func BeslutOutcome(decided bool, beslutstyper []string, hasVoteringar bool) string {
	if !decided {
		return OutcomeEjAvgjord
	}
	var votering, ackl bool
	for _, t := range beslutstyper {
		switch ModeOf(t, "") {
		case ModeVotering:
			votering = true
		case ModeAcklamation:
			ackl = true
		}
	}
	if !votering && !ackl && hasVoteringar {
		votering = true
	}
	switch {
	case votering && ackl:
		return "Avgjord med och utan omröstning"
	case votering:
		return ModeVotering.Label()
	case ackl:
		return ModeAcklamation.Label()
	}
	return "Avgjord"
}

// SortedUnique upper-cases and de-duplicates.
func SortedUnique(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		u := strings.ToUpper(strings.TrimSpace(s))
		if u != "" && !seen[u] {
			seen[u] = true
			out = append(out, u)
		}
	}
	sort.Strings(out)
	return out
}
