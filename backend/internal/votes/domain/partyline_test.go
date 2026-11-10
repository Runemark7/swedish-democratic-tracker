package domain

import "testing"

func TestPartilinje(t *testing.T) {
	tests := []struct {
		name string
		t    Tally
		want Partilinje
	}{
		{"Ja plurality", Tally{Ja: 5, Nej: 2, Avstar: 1}, PartilinjeJa},
		{"Nej plurality", Tally{Ja: 1, Nej: 4, Avstar: 3}, PartilinjeNej},
		{"Avstår plurality", Tally{Ja: 2, Nej: 1, Avstar: 6}, PartilinjeAvstar},
		{"plurality without majority", Tally{Ja: 3, Nej: 2, Avstar: 2}, PartilinjeJa},
		{"Ja/Nej tie is Delad", Tally{Ja: 4, Nej: 4}, PartilinjeDelad},
		{"Ja/Avstår tie is Delad", Tally{Ja: 7, Avstar: 7, Nej: 1}, PartilinjeDelad},
		{"three-way tie is Delad", Tally{Ja: 2, Nej: 2, Avstar: 2}, PartilinjeDelad},
		{"Frånvarande excluded from plurality", Tally{Ja: 2, Nej: 1, Franvarande: 20}, PartilinjeJa},
		{"Frånvarande cannot create a tie", Tally{Ja: 3, Nej: 1, Franvarande: 3}, PartilinjeJa},
		{"only Frånvarande: no cast ballots", Tally{Franvarande: 9}, PartilinjeSaknas},
		{"empty party", Tally{}, PartilinjeSaknas},
		{"single cast ballot", Tally{Nej: 1, Franvarande: 8}, PartilinjeNej},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := PartilinjeOf(tc.t); got != tc.want {
				t.Fatalf("PartilinjeOf(%+v) = %q, want %q", tc.t, got, tc.want)
			}
		})
	}
}

func ballot(id, party string, r VoteResult) *Vote {
	return &Vote{VoteringID: "V1", PoliticianID: id, Party: party, VoteResult: r}
}

func TestTallyByParty(t *testing.T) {
	got := TallyByParty([]*Vote{
		ballot("1", "S", VoteJa), ballot("2", "S", VoteNej), ballot("3", "S", VoteFranvarande),
		ballot("4", "M", VoteAvstar),
	})
	if got["S"] != (Tally{Ja: 1, Nej: 1, Franvarande: 1}) || got["M"] != (Tally{Avstar: 1}) {
		t.Fatalf("unexpected tallies: %+v", got)
	}
}

func TestAvvikare(t *testing.T) {
	tests := []struct {
		name  string
		votes []*Vote
		want  []string // politician ids
	}{
		{"member against Ja line",
			[]*Vote{ballot("1", "S", VoteJa), ballot("2", "S", VoteJa), ballot("3", "S", VoteNej)},
			[]string{"3"}},
		{"Avstår against Nej line counts",
			[]*Vote{ballot("1", "S", VoteNej), ballot("2", "S", VoteNej), ballot("3", "S", VoteAvstar)},
			[]string{"3"}},
		{"Frånvarande is never an Avvikare",
			[]*Vote{ballot("1", "S", VoteJa), ballot("2", "S", VoteJa), ballot("3", "S", VoteFranvarande)},
			nil},
		{"Avstår line has no Avvikare",
			[]*Vote{ballot("1", "S", VoteAvstar), ballot("2", "S", VoteAvstar), ballot("3", "S", VoteJa)},
			nil},
		{"Delad line has no Avvikare",
			[]*Vote{ballot("1", "S", VoteJa), ballot("2", "S", VoteNej)},
			nil},
		{"party with no cast ballots",
			[]*Vote{ballot("1", "S", VoteFranvarande)},
			nil},
		{"parties are independent",
			[]*Vote{ballot("1", "S", VoteJa), ballot("2", "S", VoteJa), ballot("3", "M", VoteNej), ballot("4", "M", VoteNej), ballot("5", "M", VoteJa)},
			[]string{"5"}},
		{"no-party placeholder is not a party",
			[]*Vote{ballot("1", "-", VoteJa), ballot("2", "-", VoteJa), ballot("3", "-", VoteNej)},
			nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Avvikare(tc.votes)
			var ids []string
			for _, a := range got {
				ids = append(ids, a.PoliticianID)
			}
			if len(ids) != len(tc.want) {
				t.Fatalf("Avvikare ids = %v, want %v", ids, tc.want)
			}
			for i := range ids {
				if ids[i] != tc.want[i] {
					t.Fatalf("Avvikare ids = %v, want %v", ids, tc.want)
				}
			}
		})
	}
}

func TestAvvikareCarriesPartilinje(t *testing.T) {
	got := Avvikare([]*Vote{ballot("1", "S", VoteJa), ballot("2", "S", VoteJa), ballot("3", "S", VoteNej)})
	if len(got) != 1 || got[0].Partilinje != PartilinjeJa || got[0].VoteResult != VoteNej || got[0].Party != "S" || got[0].VoteringID != "V1" {
		t.Fatalf("unexpected: %+v", got)
	}
}
