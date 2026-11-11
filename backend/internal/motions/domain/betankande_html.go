package domain

import (
	"regexp"
	"strconv"
	"strings"
)

// The parts of a betänkande that exist only in its HTML.

const (
	KindReservation      = "reservation"
	KindSarskiltYttrande = "sarskilt_yttrande"
)

// Statement is a reservation or a särskilt yttrande, as the betänkande
// prints it. Fields are what the document says; nothing is summarised.
type Statement struct {
	Kind   string
	Number int
	// Rubrik and Punkt: "Övriga frågor, punkt 6 (M)" gives rubrik
	// "Övriga frågor" and punkt "6".
	Rubrik  string
	Punkt   string
	Parties []string // from the heading: (M, C, L, KD) -> M C L KD
	Signers []string // "Cecilia Widegren (M)"
	// Proposal is the reservation's "Förslag till riksdagsbeslut".
	Proposal string
	Text     string
}

// RubrikReasoning is "Utskottets ställningstagande" under one rubrik of
// Utskottets överväganden.
type RubrikReasoning struct {
	Rubrik string
	Text   string
}

// BetankandeHTML is what could be read out of a betänkande's HTML by section
// heading. An empty result means the sections were absent or unrecognised.
type BetankandeHTML struct {
	Statements []Statement
	Reasoning  []RubrikReasoning
	// HasReasoning and HasReservations say the document had a section by
	// that heading. Without either, a punkt with no reservation is unknown,
	// not reservation-free.
	HasReasoning, HasReservations bool
}

// NormalizeRubrik makes two spellings of the same rubrik compare equal.
func NormalizeRubrik(s string) string {
	s = strings.ReplaceAll(s, "­", "")
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

// ReasoningFor returns the utskott's reasoning under the rubrik, or "" when
// the document has none or the rubrik is ambiguous (two sections share it).
func (h *BetankandeHTML) ReasoningFor(rubrik string) string {
	want := NormalizeRubrik(rubrik)
	if want == "" {
		return ""
	}
	var out string
	n := 0
	for _, r := range h.Reasoning {
		if NormalizeRubrik(r.Rubrik) == want {
			n++
			out = r.Text
		}
	}
	if n != 1 {
		return ""
	}
	return out
}

var votedAltRe = regexp.MustCompile(`(?i)mot reservation (\d+)(?:\s*\(([^)]*)\))?`)

// VotedReservation reads the reservation named in the alternative put to the
// vote ("Utskottets förslag mot reservation 12 (M)"). ok is false when the
// alternative names none (for instance "Utskottets förslag mot
// kammarens...").
func VotedReservation(alternativ string) (number int, party string, ok bool) {
	m := votedAltRe.FindStringSubmatch(alternativ)
	if m == nil {
		return 0, "", false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, "", false
	}
	return n, strings.TrimSpace(m[2]), true
}
