package domain

import (
	"testing"
	"time"

	votes "riksdagskollen/internal/votes/domain"
)

var window = []string{"2026/27", "2027/28"}

func vote(id int, dir, session string, t votes.Tally, at string) VoteInput {
	c, _ := time.Parse("2006-01-02", at)
	return VoteInput{KopplingID: id, VoteringID: "V", Direction: dir, Session: session, Tally: t, ConfirmedAt: c}
}

var (
	ja   = votes.Tally{Ja: 90, Nej: 3, Franvarande: 5}
	nej  = votes.Tally{Ja: 2, Nej: 80}
	avst = votes.Tally{Ja: 1, Avstar: 50}
	del  = votes.Tally{Ja: 10, Nej: 10}
)

func TestStatusOf(t *testing.T) {
	tests := []struct {
		name     string
		archived bool
		in       []VoteInput
		want     Status
	}{
		{"no kopplingar is Inte prövat än", false, nil, StatusInteProvat},
		{"Ja in direction Ja is Röstat för", false, []VoteInput{vote(1, "Ja", "2026/27", ja, "2026-10-01")}, StatusFor},
		{"Nej in direction Ja is Röstat emot", false, []VoteInput{vote(1, "Ja", "2026/27", nej, "2026-10-01")}, StatusEmot},
		{"reversed direction: Nej line with direction Nej is för", false, []VoteInput{vote(1, "Nej", "2026/27", nej, "2026-10-01")}, StatusFor},
		{"reversed direction: Ja line with direction Nej is emot", false, []VoteInput{vote(1, "Nej", "2026/27", ja, "2026-10-01")}, StatusEmot},
		{"för and emot is Blandat", false, []VoteInput{
			vote(1, "Ja", "2026/27", ja, "2026-10-01"), vote(2, "Ja", "2027/28", nej, "2026-10-02")}, StatusBlandat},
		{"Avstår only is Inte prövat än", false, []VoteInput{vote(1, "Ja", "2026/27", avst, "2026-10-01")}, StatusInteProvat},
		{"Delad only is Inte prövat än", false, []VoteInput{vote(1, "Ja", "2026/27", del, "2026-10-01")}, StatusInteProvat},
		{"Avstår does not dilute a för", false, []VoteInput{
			vote(1, "Ja", "2026/27", avst, "2026-10-01"), vote(2, "Ja", "2026/27", ja, "2026-10-02")}, StatusFor},
		{"only Frånvarande is Inte prövat än", false, []VoteInput{vote(1, "Ja", "2026/27", votes.Tally{Franvarande: 9}, "2026-10-01")}, StatusInteProvat},
		{"Frånvarande is excluded from the line", false, []VoteInput{vote(1, "Ja", "2026/27", votes.Tally{Ja: 1, Nej: 0, Franvarande: 40}, "2026-10-01")}, StatusFor},
		{"vote outside the window does not count", false, []VoteInput{vote(1, "Ja", "2022/23", ja, "2026-10-01")}, StatusInteProvat},
		{"outside window ignored next to inside", false, []VoteInput{
			vote(1, "Ja", "2022/23", nej, "2026-10-01"), vote(2, "Ja", "2026/27", ja, "2026-10-02")}, StatusFor},
		{"archive gives no status", true, []VoteInput{vote(1, "Ja", "2026/27", ja, "2026-10-01")}, StatusNone},
		{"archive without kopplingar gives no status", true, nil, StatusNone},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := StatusOf(tc.archived, window, tc.in).Status; got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestEvaluationExplainsEachVote(t *testing.T) {
	r := StatusOf(false, window, []VoteInput{
		vote(1, "Ja", "2026/27", ja, "2026-10-01"),
		vote(2, "Nej", "2026/27", ja, "2026-10-02"),
		vote(3, "Ja", "2026/27", avst, "2026-10-03"),
		vote(4, "Ja", "2022/23", ja, "2026-10-04"),
	})
	want := []Effect{EffectFor, EffectEmot, EffectNone, EffectOutside}
	for i, v := range r.Votes {
		if v.Effect != want[i] {
			t.Errorf("vote %d effect %q want %q", i+1, v.Effect, want[i])
		}
		if v.Explanation == "" {
			t.Errorf("vote %d has no explanation", i+1)
		}
	}
}

func TestStatusHistoryNamesTheChangingKoppling(t *testing.T) {
	in := []VoteInput{
		vote(1, "Ja", "2026/27", avst, "2026-10-01"), // Inte prövat än -> Inte prövat än (no change)
		vote(2, "Ja", "2026/27", ja, "2026-10-05"),   // -> för
		vote(3, "Ja", "2026/27", ja, "2026-10-09"),   // still för
	}
	r := StatusOf(false, window, in)
	if r.Status != StatusFor || r.ChangedBy == nil || r.ChangedBy.KopplingID != 2 {
		t.Fatalf("changed by %+v status %q", r.ChangedBy, r.Status)
	}
	if got := r.ChangedAt.Format("2006-01-02"); got != "2026-10-05" {
		t.Fatalf("changedAt %s", got)
	}
	if r.ChangeReason == "" {
		t.Fatal("no reason")
	}
	// Adding an emot makes it Blandat, changed by that koppling.
	in = append(in, vote(4, "Ja", "2027/28", nej, "2026-11-01"))
	r = StatusOf(false, window, in)
	if r.Status != StatusBlandat || r.ChangedBy.KopplingID != 4 {
		t.Fatalf("got %q by %+v", r.Status, r.ChangedBy)
	}
}

func TestStatusWithoutKopplingarHasNoChange(t *testing.T) {
	r := StatusOf(false, window, nil)
	if r.ChangedBy != nil || !r.ChangedAt.IsZero() {
		t.Fatalf("%+v", r)
	}
}
