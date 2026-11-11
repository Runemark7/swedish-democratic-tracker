package domain

import "testing"

func TestNormalizeWording(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Riksdagen ställer sig bakom det som anförs i motionen.", "riksdagenställersigbakomdetsomanförsimotionen"},
		{"  RIKSDAGEN,\n ställer\tsig  bakom: det som anförs i motionen  ", "riksdagenställersigbakomdetsomanförsimotionen"},
		{"Åke Öst – ”citat” (1)", "åkeöstcitat1"},
		{"", ""},
		{" . , ", ""},
	}
	for _, c := range cases {
		if got := NormalizeWording(c.in); got != c.want {
			t.Errorf("NormalizeWording(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestWordingHash(t *testing.T) {
	a := WordingHash("Att tillsätta en utredning.")
	b := WordingHash("att  tillsätta en utredning")
	if a == "" || a != b {
		t.Errorf("hashes differ for the same normalised wording: %q vs %q", a, b)
	}
	if a == WordingHash("att tillsätta två utredningar") {
		t.Error("different wording gave the same hash")
	}
	if WordingHash(" ,. ") != "" {
		t.Error("wording with nothing left after normalising must have no hash")
	}
}
