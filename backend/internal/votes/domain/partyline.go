package domain

import "sort"

// Partilinje is a party's position in one votering. It is THE definition on
// the site; every surface that shows a party position reads it from here so
// two pages can never disagree (CONTEXT.md: Partilinje, Avvikare).
//
// The rule: the strict plurality of the party's cast ballots (Ja, Nej,
// Avstår). Frånvarande is not a cast ballot and is excluded. When two or
// more positions share the top count the party is Delad. A party with no
// cast ballots has no Partilinje (PartilinjeSaknas).
type Partilinje string

const (
	PartilinjeJa     Partilinje = "Ja"
	PartilinjeNej    Partilinje = "Nej"
	PartilinjeAvstar Partilinje = "Avstår"
	PartilinjeDelad  Partilinje = "Delad"
	// PartilinjeSaknas means no member of the party cast a ballot. It is
	// serialised as an absent position, never as a made-up one.
	PartilinjeSaknas Partilinje = ""
)

// Tally counts one party's ballots in one votering (or in the set of
// voteringar a surface chooses to aggregate).
type Tally struct {
	Ja, Nej, Avstar, Franvarande int
}

// Add counts one ballot.
func (t *Tally) Add(r VoteResult) {
	switch r {
	case VoteJa:
		t.Ja++
	case VoteNej:
		t.Nej++
	case VoteAvstar:
		t.Avstar++
	case VoteFranvarande:
		t.Franvarande++
	}
}

// PartilinjeOf is the pure Partilinje function.
func PartilinjeOf(t Tally) Partilinje {
	max := t.Ja
	if t.Nej > max {
		max = t.Nej
	}
	if t.Avstar > max {
		max = t.Avstar
	}
	if max == 0 {
		return PartilinjeSaknas
	}
	var winner Partilinje
	n := 0
	if t.Ja == max {
		winner, n = PartilinjeJa, n+1
	}
	if t.Nej == max {
		winner, n = PartilinjeNej, n+1
	}
	if t.Avstar == max {
		winner, n = PartilinjeAvstar, n+1
	}
	if n > 1 {
		return PartilinjeDelad
	}
	return winner
}

// NoParty is the record's placeholder for ballots cast without a party
// affiliation. It is not a party and has no Partilinje.
const NoParty = "-"

// TallyByParty groups ballots by party. The NoParty placeholder gets a tally
// like any key (callers sum ballots), but PartilinjeFor refuses it a line.
func TallyByParty(votes []*Vote) map[string]Tally {
	out := map[string]Tally{}
	for _, v := range votes {
		t := out[v.Party]
		t.Add(v.VoteResult)
		out[v.Party] = t
	}
	return out
}

// PartilinjeFor is PartilinjeOf for a named party: the NoParty placeholder
// is not a party and never has a Partilinje.
func PartilinjeFor(party string, t Tally) Partilinje {
	if party == NoParty || party == "" {
		return PartilinjeSaknas
	}
	return PartilinjeOf(t)
}

// AvvikareBallot is a member whose cast ballot differs from their party's
// Ja or Nej Partilinje in one votering.
type AvvikareBallot struct {
	VoteringID   string     `json:"voteringId"`
	PoliticianID string     `json:"politicianId"`
	Party        string     `json:"party"`
	VoteResult   VoteResult `json:"voteResult"`
	Partilinje   Partilinje `json:"partilinje"`
}

// Avvikare is the pure Avvikare function for the ballots of ONE votering.
// A member is an Avvikare when they cast Ja, Nej or Avstår and the party's
// Partilinje is Ja or Nej and differs from it. Frånvarande is never an
// Avvikare; an Avstår, Delad or missing Partilinje has none. Result order is
// by party, then politician id.
func Avvikare(votes []*Vote) []AvvikareBallot {
	lines := map[string]Partilinje{}
	for p, t := range TallyByParty(votes) {
		lines[p] = PartilinjeFor(p, t)
	}
	var out []AvvikareBallot
	for _, v := range votes {
		line := lines[v.Party]
		if line != PartilinjeJa && line != PartilinjeNej {
			continue
		}
		switch v.VoteResult {
		case VoteJa, VoteNej, VoteAvstar:
		default:
			continue
		}
		if string(v.VoteResult) == string(line) {
			continue
		}
		out = append(out, AvvikareBallot{
			VoteringID: v.VoteringID, PoliticianID: v.PoliticianID,
			Party: v.Party, VoteResult: v.VoteResult, Partilinje: line,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Party != out[j].Party {
			return out[i].Party < out[j].Party
		}
		return out[i].PoliticianID < out[j].PoliticianID
	})
	return out
}

// AvvikarePerVotering applies Avvikare to each votering in votes separately,
// so ballots from two voteringar on the same förslagspunkt are never blended
// into one party tally. Voteringar are visited in id order.
func AvvikarePerVotering(votes []*Vote) []AvvikareBallot {
	by := map[string][]*Vote{}
	for _, v := range votes {
		by[v.VoteringID] = append(by[v.VoteringID], v)
	}
	ids := make([]string, 0, len(by))
	for id := range by {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var out []AvvikareBallot
	for _, id := range ids {
		out = append(out, Avvikare(by[id])...)
	}
	return out
}
