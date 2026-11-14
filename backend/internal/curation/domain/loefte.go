package domain

import (
	"strings"
)

const (
	EntityLoefte = "loefte"

	VisProposed = "proposed"
	VisApproved = "approved"
	VisStruck   = "struck"
	VisArchived = "archived"
)

// PublicVisibility reports whether a Löfte with this visibility is served by
// the public API. Mirrors loften/domain.Visibility.Public.
func PublicVisibility(v string) bool { return v == VisApproved || v == VisArchived }

// Loefte is the snapshot of a loften row that the write API edits and logs.
// Optional text uses "" for "none" so snapshots compare field by field.
type Loefte struct {
	ID            int64    `json:"id"`
	Party         string   `json:"party"`
	ValmanifestID int64    `json:"valmanifestId"`
	Text          string   `json:"text"`
	Quote         string   `json:"quote"`
	Locator       string   `json:"locator"`
	SourceURL     string   `json:"sourceUrl"`
	Topic         string   `json:"topic"`
	Specificity   string   `json:"specificity"`
	Utskott       []string `json:"utskott"`
	MandatePeriod string   `json:"mandatePeriod"`
	Visibility    string   `json:"visibility"`
}

// Equal compares two snapshots.
func (l Loefte) Equal(o Loefte) bool {
	if len(l.Utskott) != len(o.Utskott) {
		return false
	}
	for i := range l.Utskott {
		if l.Utskott[i] != o.Utskott[i] {
			return false
		}
	}
	a, b := l, o
	a.Utskott, b.Utskott = nil, nil
	return a.ID == b.ID && a.Party == b.Party && a.ValmanifestID == b.ValmanifestID && a.Text == b.Text &&
		a.Quote == b.Quote && a.Locator == b.Locator && a.SourceURL == b.SourceURL && a.Topic == b.Topic &&
		a.Specificity == b.Specificity && a.MandatePeriod == b.MandatePeriod && a.Visibility == b.Visibility
}

// LoefteFields are the editable fields shared by propose and edit.
type LoefteFields struct {
	Text        string   `json:"text"`
	Quote       string   `json:"quote"`
	Locator     string   `json:"locator"`
	SourceURL   string   `json:"sourceUrl"`
	Topic       string   `json:"topic"`
	Specificity string   `json:"specificity"`
	Utskott     []string `json:"utskott"`
}

func (f *LoefteFields) normalize() error {
	f.Text = strings.TrimSpace(f.Text)
	f.Quote = strings.TrimSpace(f.Quote)
	f.Locator = strings.TrimSpace(f.Locator)
	f.SourceURL = strings.TrimSpace(f.SourceURL)
	f.Topic = strings.TrimSpace(f.Topic)
	f.Specificity = strings.TrimSpace(f.Specificity)
	if f.Text == "" || len(f.Text) > 1000 {
		return invalid("text is required (max 1000 characters)")
	}
	if f.Quote == "" {
		return invalid("quote is required: the exact words of the Valmanifest")
	}
	if len(f.Quote) > 4000 {
		return invalid("quote is too long (max 4000 characters)")
	}
	if !httpURL(f.SourceURL) {
		return invalid("sourceUrl is required and must be an http(s) URL (page or anchor of the original)")
	}
	if len(f.Locator) > 200 || len(f.Topic) > 100 {
		return invalid("locator or topic is too long")
	}
	switch f.Specificity {
	case "", "concrete", "directional", "rhetorical":
	default:
		return invalid(`specificity must be "concrete", "directional" or "rhetorical"`)
	}
	us := make([]string, 0, len(f.Utskott))
	for _, u := range f.Utskott {
		if u = strings.TrimSpace(u); u != "" {
			us = append(us, u)
		}
	}
	f.Utskott = us
	return nil
}

// LoefteProposal creates a Löfte as "proposed" (never public).
type LoefteProposal struct {
	Party         string `json:"party"`
	ValmanifestID int64  `json:"valmanifestId"`
	MandatePeriod string `json:"mandatePeriod"`
	LoefteFields
	Provenance
}

func (in *LoefteProposal) Normalize() error {
	in.Party = strings.ToUpper(strings.TrimSpace(in.Party))
	in.MandatePeriod = strings.TrimSpace(in.MandatePeriod)
	if in.MandatePeriod == "" {
		in.MandatePeriod = "2026-2030"
	}
	if l := len(in.Party); l < 1 || l > 4 {
		return invalid("party must be a party code")
	}
	if in.ValmanifestID <= 0 {
		return invalid("valmanifestId is required")
	}
	if err := in.LoefteFields.normalize(); err != nil {
		return err
	}
	return in.Provenance.Validate()
}

func (in LoefteProposal) Snapshot() Loefte {
	return Loefte{Party: in.Party, ValmanifestID: in.ValmanifestID, Text: in.Text, Quote: in.Quote,
		Locator: in.Locator, SourceURL: in.SourceURL, Topic: in.Topic, Specificity: in.Specificity,
		Utskott: in.Utskott, MandatePeriod: in.MandatePeriod, Visibility: VisProposed}
}

// LoefteEdit replaces the editable fields of a proposed or approved Löfte.
type LoefteEdit struct {
	LoefteFields
	Provenance
}

func (in *LoefteEdit) Normalize() error {
	if err := in.LoefteFields.normalize(); err != nil {
		return err
	}
	return in.Provenance.Validate()
}

// TakesKopplingar reports whether the bot may suggest kopplingar for the Löfte:
// every approved Löfte, and an archived one only when it has both a quote and a
// source URL, so that a Löftesstatus on it can always be checked against the source.
func (l Loefte) TakesKopplingar() bool {
	switch l.Visibility {
	case VisApproved:
		return true
	case VisArchived:
		return strings.TrimSpace(l.Quote) != "" && strings.TrimSpace(l.SourceURL) != ""
	}
	return false
}
