// Package domain holds the Motion entity. It has no external dependencies.
package domain

import (
	"strings"
	"time"
)

// Motion is a proposal to the Riksdag, identified by riksmöte and beteckning
// (2017/18:3293). Fields hold what Riksdagen gave; Kind is the only derived
// value and is computed on read, never stored.
type Motion struct {
	DokID      string
	Riksmote   string // "2017/18", always the slash form
	Beteckning string // "3293"
	Title      string
	// Subtyp is Riksdagen's own value, stored as given.
	Subtyp string
	// Date is the filing date (`datum`).
	Date *time.Time
	// Organ is the utskott code Riksdagen files the motion under ("SkU").
	Organ string
	// Tilldelat is the utskott's name as given ("Skatteutskottet").
	Tilldelat string
	// Status is Riksdagen's own status word ("Klar").
	Status      string
	SystemDatum time.Time
	// BehandlasI is the betänkande the motion is handled in, as Riksdagen
	// writes it ("2017/18:SkU11"). Empty while it has no betänkande.
	BehandlasI string
	Authors    []Author
	Yrkanden   []Yrkande
}

// Author is one signer, with the party Riksdagen recorded on the motion. That
// is the party at the time of signing, not the member's party today.
type Author struct {
	Order        int
	PoliticianID string
	Name         string
	Party        string
	Role         string
}

// Yrkande is one numbered demand with its wording and the decision as given.
type Yrkande struct {
	Number int
	// Wording is the full text of the demand (`lydelse`).
	Wording string
	// UtskottProposal is the utskott's proposal as given ("Avslag").
	UtskottProposal string
	// ChamberDecision is the chamber's decision as given. Empty before it
	// decides.
	ChamberDecision string
	BehandlasI      string
	BehandlasIPunkt string
	// WordingHash is the hash of the normalised wording; empty when it has
	// not been computed yet or the wording has no letters or digits.
	WordingHash string
}

// Kind labels. Kommittémotion, Partimotion and Enskild motion are Riksdagen's
// own subtyp values; Flerpartimotion is derived from the signers.
const (
	KindKommittemotion = "Kommittémotion"
	KindPartimotion    = "Partimotion"
	KindEnskild        = "Enskild motion"
	KindFlerparti      = "Flerpartimotion"
)

// Kind returns the label shown for the motion and a one-line explanation in
// the glossary's wording. A subtyp this code does not know is returned exactly
// as Riksdagen gave it, with no explanation.
func (m Motion) Kind() (label, explanation string) {
	subtyp := strings.TrimSpace(m.Subtyp)
	switch strings.ToLower(subtyp) {
	case "kommittémotion":
		return KindKommittemotion, "Väckt i namn av ett partis ledamöter i ett utskott och uttrycker oftast partiets linje på området."
	case "partimotion":
		return KindPartimotion, "Väckt av partiledningen i partiets namn."
	case "enskild motion":
		if m.signingParties() > 1 {
			return KindFlerparti, "En enskild motion vars undertecknare kommer från mer än ett parti. Riksdagen har ingen egen typ för detta; sajten härleder den från undertecknarna."
		}
		return KindEnskild, "Väckt av en eller flera enskilda ledamöter."
	}
	return subtyp, ""
}

// signingParties counts the distinct parties among the signers. Members
// without a party ("-" or empty) do not count as a party.
func (m Motion) signingParties() int {
	seen := map[string]bool{}
	for _, a := range m.Authors {
		if a.Role != "" && !strings.EqualFold(a.Role, "undertecknare") {
			continue
		}
		p := strings.TrimSpace(a.Party)
		if p == "" || p == "-" {
			continue
		}
		seen[strings.ToUpper(p)] = true
	}
	return len(seen)
}

// Coverage is what Riksdagen's listing says exists for one riksmöte from one
// source. How many motions we hold is counted from our rows at read time.
type Coverage struct {
	Riksmote string
	Source   string
	Expected int
	// Held is counted from stored motions when coverage is read.
	Held             int
	LastEnumeratedAt time.Time
}

// SourceRiksdagenDokumentlista is the coverage source key for motions listed
// by data.riksdagen.se/dokumentlista (doktyp=mot).
const SourceRiksdagenDokumentlista = "riksdagen-dokumentlista"
