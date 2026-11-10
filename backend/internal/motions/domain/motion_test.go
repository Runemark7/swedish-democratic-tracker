package domain

import "testing"

func signers(parties ...string) []Author {
	var a []Author
	for i, p := range parties {
		a = append(a, Author{Order: i + 1, Party: p, Role: "undertecknare"})
	}
	return a
}

func TestKind(t *testing.T) {
	cases := []struct {
		name    string
		m       Motion
		want    string
		hasText bool
	}{
		{"kommitté", Motion{Subtyp: "Kommittémotion", Authors: signers("M", "M")}, KindKommittemotion, true},
		{"parti", Motion{Subtyp: "Partimotion", Authors: signers("S")}, KindPartimotion, true},
		{"enskild one party", Motion{Subtyp: "Enskild motion", Authors: signers("C", "C")}, KindEnskild, true},
		{"enskild many parties", Motion{Subtyp: "Enskild motion", Authors: signers("C", "KD", "M")}, KindFlerparti, true},
		{"independent is not a party", Motion{Subtyp: "Enskild motion", Authors: signers("C", "-", "")}, KindEnskild, true},
		{"kommittémotion across parties stays as given", Motion{Subtyp: "Kommittémotion", Authors: signers("C", "M")}, KindKommittemotion, true},
		{"unknown shown as given", Motion{Subtyp: "Ny motionstyp", Authors: signers("C", "M")}, "Ny motionstyp", false},
	}
	for _, c := range cases {
		got, expl := c.m.Kind()
		if got != c.want || (expl != "") != c.hasText {
			t.Errorf("%s: Kind() = %q, %q", c.name, got, expl)
		}
	}
}
