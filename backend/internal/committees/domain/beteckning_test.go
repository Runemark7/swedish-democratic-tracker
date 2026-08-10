package domain_test

import (
	"testing"

	"riksdagskollen/internal/committees/domain"
)

func TestCommitteeCode(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		// Ordinary codes.
		{"SoU12", "SoU"},
		{"AU9", "AU"},
		{"FiU1", "FiU"},
		// Interior uppercase and Swedish vowels.
		{"MJU5", "MJU"},
		{"FöU3", "FöU"},
		// Joint committee: real, 11 voteringar in 2022-2026.
		{"UFöU2", "UFöU"},
		// Abolished 2006, present in 2002-2006.
		{"LU21", "LU"},
		{"BoU8", "BoU"},
		// ALL-CAPS vintage from 2002-2006 normalises to the canonical form,
		// otherwise the same committee appears twice.
		{"UBU14", "UbU"},
		{"JUU9", "JuU"},
		{"SOU1", "SoU"},
		{"MJU1", "MJU"},
		// Suffixed beteckning (yttrande) still resolves.
		{"AU1y", "AU"},
		// Unknown committee codes must survive rather than be dropped; making the lookup
		// a gate is what previously hid 11 real UFöU voteringar. This is a regression
		// test: a future edit that returns "" for unmapped codes must fail here.
		{"ZZU12", "ZZU"},
		// No committee in the beteckning.
		{"", ""},
		{"1234", ""},
		{"prop.2025/26:1", ""},
	}
	for _, c := range cases {
		if got := domain.CommitteeCode(c.in); got != c.want {
			t.Errorf("CommitteeCode(%q) = %q, want %q", c.in, got, c.want)
		}
	}

	if got := domain.Canonical("UBU"); got != "UbU" {
		t.Errorf("Canonical(UBU) = %q, want UbU", got)
	}
	if got := domain.Canonical("prop."); got != "" {
		t.Errorf("Canonical(prop.) = %q, want empty", got)
	}
	// Regression test: unknown code ending in U must survive. This defends the
	// critical behavior that losing an unknown code loses voteringar; earlier this
	// hid 11 UFöU voteringar from the frontend.
	if got := domain.Canonical("ZZU"); got != "ZZU" {
		t.Errorf("Canonical(ZZU) = %q, want ZZU", got)
	}
}
