package domain

import (
	"strings"
	"time"
)

type Politician struct {
	IntressentID string    `json:"intressentId"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Party        string    `json:"party"`
	Constituency string    `json:"constituency,omitempty"`
	ImageURL     string    `json:"imageUrl,omitempty"`
	IsActive     bool      `json:"isActive"`
	UpdatedAt    time.Time `json:"updatedAt"`

	// Email is the official public address Riksdagen publishes for the member.
	Email string `json:"email,omitempty"`
	// MemberSince is the year the member's current unbroken run as
	// riksdagsledamot began. Zero when it cannot be established.
	MemberSince int `json:"memberSince,omitempty"`
	// Committees are the member's current committee assignments.
	Committees []Assignment `json:"committees,omitempty"`
	// RiksdagenURL is the member's official page on riksdagen.se.
	RiksdagenURL string `json:"riksdagenUrl,omitempty"`
}

// Assignment is one current uppdrag in a Riksdag organ (utskott).
type Assignment struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Role string `json:"role"`
	From string `json:"from,omitempty"`
}

// RiksdagenURL builds the member's riksdagen.se page. The site resolves the
// page from the intressent_id after the underscore; the slug is cosmetic.
func RiksdagenURL(firstName, lastName, intressentID string) string {
	return "https://www.riksdagen.se/sv/ledamoter-och-partier/ledamot/" +
		slug(firstName+" "+lastName) + "_" + intressentID + "/"
}

func slug(name string) string {
	repl := strings.NewReplacer("å", "a", "ä", "a", "ö", "o", "é", "e", "è", "e", "ü", "u", "ø", "o", "æ", "ae", "ç", "c", "á", "a", "à", "a", "ñ", "n", "ë", "e", "ó", "o", "í", "i", "ï", "i", "ß", "ss")
	var b strings.Builder
	dash := false
	for _, r := range repl.Replace(strings.ToLower(name)) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimRight(b.String(), "-")
}

func (p *Politician) FullName() string {
	return p.FirstName + " " + p.LastName
}
