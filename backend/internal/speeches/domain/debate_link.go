package domain

import (
	"regexp"
	"strings"
	"time"
)

// Rules by which a speech is linked to the betänkande it debates. The rule
// is stored with the link so the page can say how it was made.
const (
	// LinkRuleRelDokID: the speech's rel_dok_id names a betänkande.
	LinkRuleRelDokID = "rel_dok_id"
	// LinkRuleTitleDate: avsnittsrubrik equals the betänkande's title and the
	// protocol date equals its debate date.
	LinkRuleTitleDate = "avsnittsrubrik_datum"
)

// DebateCandidate is a betänkande a speech may belong to.
type DebateCandidate struct {
	DokID      string
	Title      string
	DebateDate time.Time // zero when unknown
}

// SpeechForLinking is what the linking decides from.
type SpeechForLinking struct {
	RelatedDokID string
	TopicHeading string
	Date         time.Time
}

// DebateLink is the outcome: a betänkande's dok_id and the rule used.
type DebateLink struct {
	BetankandeDokID string
	Rule            string
}

// betankandeDokID matches a betänkande's dok_id: H901MJU33, H501SkU11,
// GQ01NU11 (two-character prefix, two-digit year, committee letters, number).
// Interpellations, motions and propositions end in digits only.
var betankandeDokID = regexp.MustCompile(`^[A-Z0-9]{2}[0-9]{2}[A-Za-zÅÄÖåäö]+[0-9]+$`)

// IsBetankandeDokID reports whether id has the shape of a betänkande dok_id.
func IsBetankandeDokID(id string) bool { return betankandeDokID.MatchString(strings.TrimSpace(id)) }

// LinkSpeech links a speech to the betänkande it debates. It tries, in order:
//
//  1. rel_dok_id, when it names a betänkande;
//  2. avsnittsrubrik = title and protocol date = debate date, when exactly
//     one betänkande matches;
//  3. nothing: the speech stays unlinked. It never guesses.
func LinkSpeech(s SpeechForLinking, candidates []DebateCandidate) (DebateLink, bool) {
	if id := strings.ToUpper(strings.TrimSpace(s.RelatedDokID)); IsBetankandeDokID(id) {
		return DebateLink{BetankandeDokID: id, Rule: LinkRuleRelDokID}, true
	}
	heading := normalizeTitle(s.TopicHeading)
	if heading == "" || s.Date.IsZero() {
		return DebateLink{}, false
	}
	var hit string
	for _, c := range candidates {
		if c.DebateDate.IsZero() || !sameDay(c.DebateDate, s.Date) || normalizeTitle(c.Title) != heading {
			continue
		}
		id := strings.ToUpper(c.DokID)
		if hit != "" && hit != id {
			return DebateLink{}, false // two betänkanden match: ambiguous
		}
		hit = id
	}
	if hit == "" {
		return DebateLink{}, false
	}
	return DebateLink{BetankandeDokID: hit, Rule: LinkRuleTitleDate}, true
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// normalizeTitle lowercases and collapses whitespace so "Fristående skolor "
// equals "fristående  skolor".
func normalizeTitle(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}
