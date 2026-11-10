package domain

import "testing"

func TestPropositionsIn(t *testing.T) {
	got := PropositionsIn("<BR/>Riksdagen antar regeringens förslag i proposition 2021/22:5 och proposition 2021/22:6 och avslår motion 2021/22:7 yrkande 1.<BR/>")
	if len(got) != 2 || got[0].Beteckning != "5" || got[1].Beteckning != "6" || got[0].Riksmote != "2021/22" {
		t.Errorf("got %+v", got)
	}
	if got := PropositionsIn("Riksdagen avslår motionerna 2021/22:3312 av X (M) och 2021/22:3637."); len(got) != 0 {
		t.Errorf("motions only: got %+v", got)
	}
	if got := PropositionsIn("proposition 2021/22:5 punkt 1 och 2021/22:5 punkt 2"); len(got) != 1 {
		t.Errorf("duplicate: got %+v", got)
	}
}

func TestSoleParty(t *testing.T) {
	m := func(p ...string) ProposingMotion { return ProposingMotion{Parties: p} }
	cases := []struct {
		name string
		in   ProposedBy
		want string
	}{
		{"none", ProposedBy{}, ""},
		{"prop", ProposedBy{Propositions: []PropositionRef{{}}}, "Regeringen"},
		{"one party", ProposedBy{Motions: []ProposingMotion{m("M"), m("M")}}, "M"},
		{"several parties", ProposedBy{Motions: []ProposingMotion{m("M"), m("C")}}, ""},
		{"prop and motion", ProposedBy{Motions: []ProposingMotion{m("M")}, Propositions: []PropositionRef{{}}}, ""},
	}
	for _, c := range cases {
		if got := c.in.SoleParty(); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}
