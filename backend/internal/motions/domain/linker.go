package domain

import (
	"html"
	"regexp"
	"strconv"
	"strings"
)

// How a yrkande was tied to a förslagspunkt.
const (
	LinkedByBehandlasIPunkt = "behandlas_i_punkt"
	LinkedByForslagText     = "forslag_text"
)

// What the förslagspunkt's text says was done with one motion.
const (
	DecisionAvslag       = "Avslag"
	DecisionBifall       = "Bifall"
	DecisionDelvisBifall = "Delvis bifall"
	DecisionBesvarad     = "Besvarad"
)

// Punkt is the part of a förslagspunkt the Linker reads: its number and the
// utskottet's förslag text as Riksdagen serves it (HTML breaks and entities
// are handled here).
type Punkt struct {
	Number string
	Text   string
}

// YrkandeRef identifies the yrkande to link.
type YrkandeRef struct {
	Riksmote   string // "2017/18"
	Beteckning string // "3293"
	Number     int
	// BehandlasIPunkt is Riksdagen's own pointer, often empty.
	BehandlasIPunkt string
}

// PunktLink is one förslagspunkt that decided the yrkande.
type PunktLink struct {
	Punkt string
	How   string
	// Decision is what the text says was done with the motion (a Decision*
	// constant), empty when the text does not say or the link is explicit.
	Decision string
}

var (
	brRe       = regexp.MustCompile(`(?i)<br\s*/?>`)
	tagRe      = regexp.MustCompile(`<[^>]*>`)
	spaceRe    = regexp.MustCompile(`\s+`)
	citationRe = regexp.MustCompile(`(\d{4}/\d{2}):(\d+)`)
	verbRe     = regexp.MustCompile(`(?i)(bifaller delvis|bifaller|avslår|är besvarad|besvarad)`)
	yrkandeRe  = regexp.MustCompile(`(?i)yrkand(?:ena|e)\s+(\d+(?:\s*[-–]\s*\d+)?(?:\s*(?:,|och|samt)\s*\d+(?:\s*[-–]\s*\d+)?)*)`)
	itemRe     = regexp.MustCompile(`(\d+)(?:\s*[-–]\s*(\d+))?`)
)

func normalizeText(s string) string {
	s = brRe.ReplaceAllString(s, " ")
	s = tagRe.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	return strings.TrimSpace(spaceRe.ReplaceAllString(s, " "))
}

func decisionFromVerb(v string) string {
	switch strings.ToLower(v) {
	case "bifaller delvis":
		return DecisionDelvisBifall
	case "bifaller":
		return DecisionBifall
	case "avslår":
		return DecisionAvslag
	case "är besvarad", "besvarad":
		return DecisionBesvarad
	}
	return ""
}

// PunktDecisionOf reads what a förslagspunkt decided from its first verb. It
// returns "" when the utskottet's proposal did not win (vinnare is something
// other than "utskottet") or the text has no verb it knows: the caller then
// shows no outcome rather than a guess.
func PunktDecisionOf(text, vinnare string) string {
	if v := strings.TrimSpace(vinnare); v != "" && !strings.EqualFold(v, "utskottet") {
		return ""
	}
	m := verbRe.FindString(normalizeText(text))
	return decisionFromVerb(m)
}

// LinkYrkande is the Linker: a pure function from a yrkande and the
// förslagspunkter of the betänkande it is handled in to the punkt(er) that
// decided it. behandlas_i_punkt is used when it names a punkt that exists;
// otherwise the förslag text is parsed ("yrkandena 1–3 och 5", a bare motion
// meaning all its yrkanden, "bifaller delvis"). A yrkande no punkt cites gets
// no link.
func LinkYrkande(y YrkandeRef, punkter []Punkt) []PunktLink {
	if p := strings.TrimSpace(y.BehandlasIPunkt); p != "" {
		for _, pk := range punkter {
			if pk.Number == p {
				return []PunktLink{{Punkt: p, How: LinkedByBehandlasIPunkt}}
			}
		}
	}
	var out []PunktLink
	for _, pk := range punkter {
		if d, ok := citedIn(normalizeText(pk.Text), y); ok {
			out = append(out, PunktLink{Punkt: pk.Number, How: LinkedByForslagText, Decision: d})
		}
	}
	return out
}

func citedIn(text string, y YrkandeRef) (string, bool) {
	locs := citationRe.FindAllStringSubmatchIndex(text, -1)
	for i, loc := range locs {
		if loc[0] > 0 && text[loc[0]-1] >= '0' && text[loc[0]-1] <= '9' {
			continue
		}
		if text[loc[2]:loc[3]] != y.Riksmote || text[loc[4]:loc[5]] != y.Beteckning {
			continue
		}
		segEnd := len(text)
		if i+1 < len(locs) {
			segEnd = locs[i+1][0]
		}
		head := text[loc[1]:segEnd]
		if i+1 < len(locs) {
			// The last verb of the segment opens the next citation's clause.
			if all := verbRe.FindAllStringIndex(head, -1); len(all) > 0 {
				head = head[:all[len(all)-1][0]]
			}
		}
		decision := ""
		if all := verbRe.FindAllString(text[:loc[0]], -1); len(all) > 0 {
			decision = decisionFromVerb(all[len(all)-1])
		}
		if ms := yrkandeRe.FindAllStringSubmatch(head, -1); len(ms) > 0 {
			found := false
			for _, m := range ms {
				if listHas(m[1], y.Number) {
					found = true
				}
			}
			if !found {
				continue
			}
		}
		if decision == DecisionBifall && strings.Contains(strings.ToLower(head), "delvis") {
			decision = DecisionDelvisBifall
		}
		return decision, true
	}
	return "", false
}

func listHas(list string, n int) bool {
	for _, m := range itemRe.FindAllStringSubmatch(list, -1) {
		lo, _ := strconv.Atoi(m[1])
		hi := lo
		if m[2] != "" {
			hi, _ = strconv.Atoi(m[2])
		}
		if n >= lo && n <= hi {
			return true
		}
	}
	return false
}
