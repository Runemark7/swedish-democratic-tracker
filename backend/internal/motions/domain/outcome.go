package domain

import "strings"

// Outcome words shown to the reader.
const (
	OutcomeAvslagen  = "Avslagen"
	OutcomeBifallen  = "Bifallen"
	OutcomeDelvis    = "Delvis bifallen"
	OutcomeBesvarad  = "Besvarad"
	OutcomeEjAvgjord = "Ej avgjord än"
	// OutcomeBlandat is used when a motion's yrkanden ended differently and
	// none of them was bifallen (for example Avslagen and Besvarad).
	OutcomeBlandat = "Blandat utfall"
)

func mapOutcomeWord(w string) string {
	l := strings.ToLower(strings.TrimSpace(w))
	switch {
	case l == "":
		return ""
	case strings.Contains(l, "delvis"):
		return OutcomeDelvis
	case strings.Contains(l, "avslag"):
		return OutcomeAvslagen
	case strings.Contains(l, "bifall"):
		return OutcomeBifallen
	case strings.Contains(l, "besvar"):
		return OutcomeBesvarad
	}
	return strings.TrimSpace(w)
}

// YrkandeOutcome is the outcome of one yrkande: the chamber's decision as
// Riksdagen gave it, or, when kammaren is empty, the decision of the
// förslagspunkt. A word we do not know is returned as Riksdagen wrote it.
// With neither, the yrkande is Ej avgjord än.
func YrkandeOutcome(kammaren, punktDecision string) string {
	if w := mapOutcomeWord(kammaren); w != "" {
		return w
	}
	if w := mapOutcomeWord(punktDecision); w != "" {
		return w
	}
	return OutcomeEjAvgjord
}

// Aggregate is a motion's outcome with the counts behind it.
type Aggregate struct {
	Word   string
	Counts map[string]int
	Total  int
}

// MotionOutcome aggregates the outcomes of a motion's yrkanden. All alike:
// that word. Any still undecided: Ej avgjord än. Mixed with a bifall: Delvis
// bifallen. Mixed without one: Blandat utfall. Counts are always returned.
func MotionOutcome(words []string) Aggregate {
	a := Aggregate{Counts: map[string]int{}, Total: len(words)}
	for _, w := range words {
		a.Counts[w]++
	}
	switch {
	case len(words) == 0 || a.Counts[OutcomeEjAvgjord] > 0:
		a.Word = OutcomeEjAvgjord
	case len(a.Counts) == 1:
		a.Word = words[0]
	case a.Counts[OutcomeBifallen] > 0 || a.Counts[OutcomeDelvis] > 0:
		a.Word = OutcomeDelvis
	default:
		a.Word = OutcomeBlandat
	}
	return a
}

// DecisionMode is how the chamber decided a förslagspunkt.
type DecisionMode string

const (
	ModeAcklamation DecisionMode = "acklamation"
	ModeVotering    DecisionMode = "votering"
	ModeEjBeslutad  DecisionMode = "ej_beslutad"
	// ModeAnnat is any beslutstyp this code does not know; the caller shows
	// Riksdagen's own word.
	ModeAnnat DecisionMode = "annat"
)

// ModeOf reads the mode from Riksdagen's beslutstyp.
func ModeOf(beslutstyp, voteringID string) DecisionMode {
	switch strings.ToLower(strings.TrimSpace(beslutstyp)) {
	case "acklamation":
		return ModeAcklamation
	case "röstning":
		return ModeVotering
	case "":
		return ModeEjBeslutad
	}
	return ModeAnnat
}

// Label is the plain Swedish line for the mode.
func (m DecisionMode) Label() string {
	switch m {
	case ModeAcklamation:
		return "Avgjord utan omröstning"
	case ModeVotering:
		return "Avgjord genom omröstning"
	case ModeEjBeslutad:
		return OutcomeEjAvgjord
	}
	return ""
}

// YrkandeWord is the outcome of one yrkande given the chamber's decision as
// Riksdagen gave it and the decision text of each förslagspunkt the yrkande is
// linked to. Without a linked förslagspunkt there is no decision to show: the
// yrkande is Ej avgjord än, whatever the listing's own columns say.
func YrkandeWord(kammaren string, punktDecisions []string) string {
	switch len(punktDecisions) {
	case 0:
		return OutcomeEjAvgjord
	case 1:
		return YrkandeOutcome(kammaren, punktDecisions[0])
	}
	words := make([]string, 0, len(punktDecisions))
	for _, d := range punktDecisions {
		words = append(words, YrkandeOutcome(kammaren, d))
	}
	return MotionOutcome(words).Word
}
