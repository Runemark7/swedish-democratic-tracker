// Package domain holds the pure entities of skriftliga frågor and
// interpellationer.
package domain

import (
	"regexp"
	"strings"
	"time"
)

const (
	KindFraga          = "fraga"
	KindInterpellation = "interpellation"
)

// Link rules. Only a link that passes one of them is stored or shown.
const (
	RuleReference = "riksdagen-reference"
	RuleCited     = "cited-beteckning"
)

// Person is one of the people on the document.
type Person struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Party string `json:"party"`
}

// Question is a skriftlig fråga (with its svar) or an interpellation.
type Question struct {
	DokID       string
	Kind        string
	Riksmote    string
	Beteckning  string
	Title       string
	AskedDate   *time.Time
	Status      string
	SystemDatum time.Time

	Asker     Person
	Addressed Person
	Answerer  Person

	// AnswerDokID is the svar document of a fråga.
	AnswerDokID string
	AnswerDate  *time.Time

	// BodyText is the question's plain text.
	BodyText string
	// MotionRefs are the motions Riksdagen's references point at, "rm:bet".
	MotionRefs []string
}

// Ref is "rm:beteckning" of a motion.
func Ref(rm, bet string) string { return rm + ":" + bet }

var beteckningRe = regexp.MustCompile(`(\d{4}/\d{2}):(\d{1,5})`)

// Keywords that say what kind of document a beteckning names. A bare
// "2022/23:100" can be a motion, a proposition or a skrivelse, so a citation
// counts only when the nearest keyword before it says motion.
var (
	motionWords = []string{"motion", "mot."}
	otherWords  = []string{"proposition", "prop.", "skrivelse", "skr.", "betänkande", "bet.", "rskr", "fråga", "interpellation", "svar på"}
)

// CitedMotions returns the "rm:beteckning" of every motion the text cites by
// beteckning, in order of appearance, without duplicates. A beteckning that is
// not introduced by the word motion (or mot.) is not a citation, and neither is
// the document's own beteckning. A motion named only in prose ("motionen från
// 2017") has no beteckning and yields nothing.
func CitedMotions(text, ownRiksmote, ownBeteckning string) []string {
	lower := strings.ToLower(text)
	var out []string
	seen := map[string]bool{}
	for _, loc := range beteckningRe.FindAllStringSubmatchIndex(text, -1) {
		rm, bet := text[loc[2]:loc[3]], text[loc[4]:loc[5]]
		key := Ref(rm, bet)
		if seen[key] || (rm == ownRiksmote && bet == ownBeteckning) {
			continue
		}
		start := loc[0] - 60
		if start < 0 {
			start = 0
		}
		window := lower[start:loc[0]]
		if lastIndexAny(window, motionWords) > lastIndexAny(window, otherWords) {
			seen[key] = true
			out = append(out, key)
		}
	}
	return out
}

func lastIndexAny(s string, words []string) int {
	best := -1
	for _, w := range words {
		if i := strings.LastIndex(s, w); i > best {
			best = i
		}
	}
	return best
}

// Link is one stored question→motion link.
type Link struct {
	QuestionDokID string
	MotionDokID   string
	Rule          string
}

// Resolve resolves candidate links against the motions held. It returns
// the links that pass a rule: a Riksdagen reference first, a cited beteckning
// otherwise. A candidate for a motion that is not held yields no link.
func Resolve(q Question, held func(ref string) (dokID string, ok bool)) []Link {
	var links []Link
	seen := map[string]bool{}
	add := func(ref, rule string) {
		id, ok := held(ref)
		if !ok || seen[id] {
			return
		}
		seen[id] = true
		links = append(links, Link{QuestionDokID: q.DokID, MotionDokID: id, Rule: rule})
	}
	for _, r := range q.MotionRefs {
		add(r, RuleReference)
	}
	for _, r := range CitedMotions(q.BodyText, q.Riksmote, q.Beteckning) {
		add(r, RuleCited)
	}
	return links
}

// Item is a question as listed, with the member's role and the motion link.
type Item struct {
	DokID      string     `json:"dokId"`
	Kind       string     `json:"kind"`
	Riksmote   string     `json:"riksmote"`
	Beteckning string     `json:"beteckning"`
	Title      string     `json:"title"`
	AskedDate  *time.Time `json:"askedDate"`
	Status     string     `json:"status"`
	Asker      Person     `json:"asker"`
	Addressed  Person     `json:"addressed"`
	Answerer   Person     `json:"answerer"`
	AnswerDate *time.Time `json:"answerDate"`
	// AnswerDokID is the svar document of a fråga.
	AnswerDokID string `json:"answerDokId"`
	// Role is the member's role in a by-member listing: asker or answerer.
	Role string `json:"role,omitempty"`
	// Rule is the rule behind the link in a by-motion listing.
	Rule      string `json:"rule,omitempty"`
	SourceURL string `json:"sourceUrl"`
}

// Coverage is what Riksdagen lists against what is held.
type Coverage struct {
	Riksmote         string    `json:"riksmote"`
	Kind             string    `json:"kind"`
	Expected         int       `json:"expected"`
	Held             int       `json:"held"`
	LastEnumeratedAt time.Time `json:"lastEnumeratedAt"`
}

// DebateSpeech is an anförande in an interpellation's debate.
type DebateSpeech struct {
	SpeechID     int       `json:"speechId"`
	DokID        string    `json:"dokId"`
	PoliticianID string    `json:"politicianId"`
	Name         string    `json:"name"`
	Party        string    `json:"party"`
	Date         time.Time `json:"date"`
	Excerpt      string    `json:"excerpt"`
}

// Detail is one question with, for an interpellation, its debate.
type Detail struct {
	Item
	Text    string         `json:"text"`
	Debate  []DebateSpeech `json:"debate"`
	Motions []LinkedMotion `json:"motions"`
}

// LinkedMotion is a motion a question is linked to.
type LinkedMotion struct {
	DokID      string `json:"dokId"`
	Riksmote   string `json:"riksmote"`
	Beteckning string `json:"beteckning"`
	Title      string `json:"title"`
	Rule       string `json:"rule"`
}

// SourceURL is the document at Riksdagen.
func SourceURL(dokID string) string {
	return "https://data.riksdagen.se/dokument/" + dokID
}
