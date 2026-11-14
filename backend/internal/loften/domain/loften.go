// Package domain holds Löfte and Valmanifest, the Phase 2 vocabulary of
// CONTEXT.md. Nothing here knows about storage or HTTP.
package domain

import "time"

// Visibility is the only gate on what is public.
type Visibility string

const (
	Proposed Visibility = "proposed"
	Approved Visibility = "approved"
	Struck   Visibility = "struck"
	Archived Visibility = "archived"
)

// Public reports whether a Löfte with this visibility may leave the API.
// Proposed and struck Löften never may.
func (v Visibility) Public() bool { return v == Approved || v == Archived }

// Loefte is one concrete, testable statement in a Valmanifest. It deliberately
// has no stored status field: Löftesstatus is computed on read from Bekräftade
// kopplingar. An archived 2022 Löfte without any gets none, since a status
// would state our curation gap as a fact about the party.
type Loefte struct {
	ID            int        `json:"id"`
	Party         string     `json:"party"`
	ValmanifestID *int       `json:"valmanifestId"`
	Text          string     `json:"text"`
	Quote         string     `json:"quote,omitempty"`
	Locator       string     `json:"locator,omitempty"`
	SourceURL     string     `json:"sourceUrl,omitempty"`
	SourceDoc     string     `json:"sourceDocument,omitempty"`
	Topic         string     `json:"topic,omitempty"`
	Utskott       []string   `json:"utskott"`
	MandatePeriod string     `json:"mandatePeriod"`
	Visibility    Visibility `json:"visibility"`
	Archived      bool       `json:"archived"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// Valmanifest is a party's published election programme, a source for Löften.
// State is the party's own marking: "draft" (utkast, preliminär) or "final".
type Valmanifest struct {
	ID           int        `json:"id"`
	Party        string     `json:"party"`
	ElectionYear int        `json:"electionYear"`
	Title        string     `json:"title"`
	PublishedOn  *time.Time `json:"publishedOn"`
	URL          string     `json:"url"`
	PDFURL       string     `json:"pdfUrl,omitempty"`
	State        string     `json:"state"`
}
