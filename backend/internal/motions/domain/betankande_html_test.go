package domain

import "testing"

func TestPlaceUnpointedPutsStatementsOnTheOnlyForslagspunkt(t *testing.T) {
	h := &BetankandeHTML{Statements: []Statement{
		{Kind: KindReservation, Number: 1, Rubrik: "Vilande förslag"},
		{Kind: KindSarskiltYttrande, Number: 1, Rubrik: "Vilande förslag", Punkt: "7"},
	}}
	h.PlaceUnpointed([]string{"1"})
	if h.Statements[0].Punkt != "1" {
		t.Errorf("unpointed statement on punkt %q, want 1", h.Statements[0].Punkt)
	}
	if h.Statements[1].Punkt != "7" {
		t.Errorf("a statement that names its punkt was moved to %q", h.Statements[1].Punkt)
	}
}

func TestPlaceUnpointedGuessesNothingWhenThereAreSeveralForslagspunkter(t *testing.T) {
	for _, punkter := range [][]string{nil, {"1", "2"}} {
		h := &BetankandeHTML{Statements: []Statement{{Kind: KindReservation, Number: 1}}}
		h.PlaceUnpointed(punkter)
		if h.Statements[0].Punkt != "" {
			t.Errorf("punkter %v: placed on %q", punkter, h.Statements[0].Punkt)
		}
	}
}
