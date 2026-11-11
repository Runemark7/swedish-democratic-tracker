package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
	"unicode"
)

// RuleIdenticalWording is the rule that groups yrkanden into an Återkommande
// krav in Phase 1: the normalised wording is identical.
const RuleIdenticalWording = "identical_wording"

// NormalizeWording lowercases the wording and removes everything that is not a
// letter or a digit (whitespace, punctuation, quotes), so two filings of the
// same demand compare equal despite typographic differences.
func NormalizeWording(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// WordingHash is the hex SHA-256 of the normalised wording, or "" when nothing
// is left after normalising (such a yrkande is never grouped).
func WordingHash(wording string) string {
	n := NormalizeWording(wording)
	if n == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(n))
	return hex.EncodeToString(sum[:])
}

// RecurringMember is one filing of a recurring demand.
type RecurringMember struct {
	DokID      string
	Riksmote   string
	Beteckning string
	Number     int
	Date       *time.Time
	// Parties are the distinct parties among the signers, in signing order.
	Parties []string
	// ChamberDecision is the listing's decision, used with the linked
	// decisions to resolve Outcome.
	ChamberDecision string
	// Rule is why this filing is in the group.
	Rule    string
	Outcome string
}

// Recurring is an Återkommande krav as seen from one yrkande: every filing
// sharing its wording, earliest first.
type Recurring struct {
	Rule    string
	Members []RecurringMember
}
