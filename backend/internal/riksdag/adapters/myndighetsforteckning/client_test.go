package myndighetsforteckning

import "testing"

func TestNormalizeOrgNumber(t *testing.T) {
	cases := map[string]string{
		"2021004284":   "202100-4284", // 10-digit flat → canonical
		"202100-4284":  "202100-4284", // already canonical
		"857209-0606":  "857209-0606", // already canonical (AP-fond style)
		"8572090606":   "857209-0606", // 10-digit flat AP-fond
		" 2021004284 ": "202100-4284", // whitespace tolerated
		"":             "",            // empty stays empty
		"abc":          "",            // garbage rejected
	}
	for in, want := range cases {
		if got := normalizeOrgNumber(in); got != want {
			t.Errorf("normalizeOrgNumber(%q) = %q, want %q", in, got, want)
		}
	}
}
