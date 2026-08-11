package domain_test

import (
	"testing"

	"riksdagskollen/internal/committees/domain"
)

func TestDisplayName(t *testing.T) {
	if got := domain.DisplayName("SoU"); got != "Socialutskottet" {
		t.Errorf("DisplayName(SoU) = %q", got)
	}
	if got := domain.DisplayName("UFöU"); got != "Sammansatta utrikes- och försvarsutskottet" {
		t.Errorf("DisplayName(UFöU) = %q", got)
	}
	if got := domain.DisplayName("LU"); got != "Lagutskottet" {
		t.Errorf("DisplayName(LU) = %q", got)
	}
	// An unknown code must fall back to itself, never to "" and never be
	// dropped: the lookup is a display aid, not a gate on what exists.
	if got := domain.DisplayName("ZZU"); got != "ZZU" {
		t.Errorf("DisplayName(ZZU) = %q, want the code itself", got)
	}
}
