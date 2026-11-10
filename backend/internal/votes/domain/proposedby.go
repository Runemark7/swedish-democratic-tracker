package domain

import (
	"regexp"
	"strings"
)

// ProposedBy is who proposed what a votering decided. It is derived at read
// time from the förslagspunkt's links (riksmöte + beteckning + förslagspunkt):
// the motions and yrkanden the punkt decided, and the propositions its text
// cites. Nothing here is stored on the votering.
type ProposedBy struct {
	Motions      []ProposingMotion `json:"motions"`
	Propositions []PropositionRef  `json:"propositions"`
	// Alternativ is the alternative put to the vote, as Riksdagen words it
	// ("Utskottets förslag mot reservation 12 (M)"). Empty when none was.
	Alternativ string `json:"alternativ,omitempty"`
}

// ProposingMotion is one motion the förslagspunkt decided.
type ProposingMotion struct {
	DokID      string `json:"dokId"`
	Riksmote   string `json:"riksmote"`
	Beteckning string `json:"beteckning"`
	Title      string `json:"title"`
	Subtyp     string `json:"subtyp,omitempty"`
	// Parties are the distinct parties of the motion's authors, in signing order.
	Parties []string `json:"parties"`
	// Yrkanden are the yrkande numbers of this motion the punkt decided.
	Yrkanden []int `json:"yrkanden"`
}

// PropositionRef is a proposition cited in the förslagspunkt's text.
type PropositionRef struct {
	Riksmote   string `json:"riksmote"`
	Beteckning string `json:"beteckning"`
}

var propTokens = regexp.MustCompile(`(?i)(motion\w*|proposition\w*|reservation\w*)|(\d{4}/\d{2}):(\d+)`)

// PropositionsIn returns the propositions a förslagspunkt text cites. A
// number counts as a proposition when the nearest preceding keyword is
// "proposition" (so "proposition 2021/22:5 och motion 2021/22:7" yields only
// 2021/22:5), and each is returned once, in order of appearance.
func PropositionsIn(text string) []PropositionRef {
	var out []PropositionRef
	seen := map[string]bool{}
	mode := ""
	for _, m := range propTokens.FindAllStringSubmatch(text, -1) {
		if m[1] != "" {
			mode = strings.ToLower(m[1][:3])
			continue
		}
		if mode != "pro" {
			continue
		}
		key := m[2] + ":" + m[3]
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, PropositionRef{Riksmote: m[2], Beteckning: m[3]})
	}
	return out
}

// SoleParty names the single party behind the proposal when that is
// unambiguous: "Regeringen" for propositions alone, otherwise the one party
// every motion shares. It is "" when the proposal has several proposers,
// because picking one would be an interpretation.
func (p ProposedBy) SoleParty() string {
	if len(p.Motions) == 0 {
		if len(p.Propositions) > 0 {
			return "Regeringen"
		}
		return ""
	}
	if len(p.Propositions) > 0 {
		return ""
	}
	party := ""
	for _, m := range p.Motions {
		for _, mp := range m.Parties {
			if party == "" {
				party = mp
			} else if party != mp {
				return ""
			}
		}
	}
	return party
}

// Kind is "prop" when only propositions, "mot" when only motions, else "".
func (p ProposedBy) Kind() string {
	switch {
	case len(p.Motions) > 0 && len(p.Propositions) == 0:
		return "mot"
	case len(p.Motions) == 0 && len(p.Propositions) > 0:
		return "prop"
	}
	return ""
}
