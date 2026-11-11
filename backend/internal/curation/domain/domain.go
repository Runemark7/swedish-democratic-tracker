// Package domain holds the Ändringslogg vocabulary of CONTEXT.md: an entry
// records who changed what, when, from which value to which, and the chat
// answer it came from. Nothing here knows about storage or HTTP.
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	ActionUpsert = "upsert"
	ActionUndo   = "undo"

	EntityValmanifest = "valmanifest"
)

// Errors the service returns; the HTTP adapter maps each to a status.
var (
	ErrNotFound = errors.New("not found")
	// ErrConflict: the value changed since the entry (undo), or the change
	// collides with other data.
	ErrConflict = errors.New("conflict")
	// ErrKeyReuse: the Idempotency-Key was already used for a different request.
	ErrKeyReuse = errors.New("idempotency key reused with a different request")
)

// ValidationError carries a message safe to show the caller.
type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }

func invalid(format string, a ...any) error { return &ValidationError{fmt.Sprintf(format, a...)} }

// Provenance says who wrote and why. Required on every write.
type Provenance struct {
	OnBehalfOf string `json:"onBehalfOf"`
	ChatRef    string `json:"chatRef"`
}

func (p Provenance) Validate() error {
	if strings.TrimSpace(p.OnBehalfOf) == "" {
		return invalid("onBehalfOf is required")
	}
	if strings.TrimSpace(p.ChatRef) == "" {
		return invalid("chatRef is required")
	}
	if len(p.OnBehalfOf) > 200 || len(p.ChatRef) > 500 {
		return invalid("onBehalfOf or chatRef is too long")
	}
	return nil
}

// Valmanifest is the editable snapshot of a Valmanifest row. Optional text
// fields use "" for "none" so snapshots compare with ==.
type Valmanifest struct {
	ID           int64  `json:"id"`
	Party        string `json:"party"`
	ElectionYear int    `json:"electionYear"`
	Title        string `json:"title"`
	PublishedOn  string `json:"publishedOn"` // YYYY-MM-DD or ""
	URL          string `json:"url"`
	PDFURL       string `json:"pdfUrl"`
	State        string `json:"state"`
}

// ValmanifestInput is the upsert payload; the key is (party, electionYear, title).
type ValmanifestInput struct {
	Party        string `json:"party"`
	ElectionYear int    `json:"electionYear"`
	Title        string `json:"title"`
	PublishedOn  string `json:"publishedOn"`
	URL          string `json:"url"`
	PDFURL       string `json:"pdfUrl"`
	State        string `json:"state"`
	Provenance
}

func httpURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// Normalize trims and validates in place.
func (in *ValmanifestInput) Normalize() error {
	in.Party = strings.ToUpper(strings.TrimSpace(in.Party))
	in.Title = strings.TrimSpace(in.Title)
	in.URL = strings.TrimSpace(in.URL)
	in.PDFURL = strings.TrimSpace(in.PDFURL)
	in.PublishedOn = strings.TrimSpace(in.PublishedOn)
	if l := len(in.Party); l < 1 || l > 4 {
		return invalid("party must be a party code")
	}
	if in.ElectionYear < 1900 || in.ElectionYear > 2100 {
		return invalid("electionYear must be between 1900 and 2100")
	}
	if in.Title == "" || len(in.Title) > 300 {
		return invalid("title is required (max 300 characters)")
	}
	if !httpURL(in.URL) {
		return invalid("url must be an http(s) URL")
	}
	if in.PDFURL != "" && !httpURL(in.PDFURL) {
		return invalid("pdfUrl must be an http(s) URL")
	}
	if in.State != "draft" && in.State != "final" {
		return invalid(`state must be "draft" or "final"`)
	}
	if in.PublishedOn != "" {
		if _, err := time.Parse("2006-01-02", in.PublishedOn); err != nil {
			return invalid("publishedOn must be YYYY-MM-DD")
		}
	}
	return in.Provenance.Validate()
}

func (in ValmanifestInput) Snapshot() Valmanifest {
	return Valmanifest{Party: in.Party, ElectionYear: in.ElectionYear, Title: in.Title,
		PublishedOn: in.PublishedOn, URL: in.URL, PDFURL: in.PDFURL, State: in.State}
}

// Entry is one Ändringslogg row. Old/New are JSON snapshots; nil means the
// entity did not exist before / does not exist after.
type Entry struct {
	ID         int64           `json:"id"`
	CreatedAt  time.Time       `json:"createdAt"`
	Actor      string          `json:"actor"`
	OnBehalfOf string          `json:"onBehalfOf"`
	ChatRef    string          `json:"chatRef"`
	Action     string          `json:"action"`
	Entity     string          `json:"entity"`
	EntityID   int64           `json:"entityId"`
	OldValue   json.RawMessage `json:"oldValue"`
	NewValue   json.RawMessage `json:"newValue"`
	Undoes     *int64          `json:"undoes"`

	// Never serialised: the idempotency bookkeeping.
	RequestKey  string `json:"-"`
	RequestHash string `json:"-"`
}

// Result of a write. Changed is false when the request matched the stored
// value, in which case no entry exists.
type Result struct {
	Changed  bool   `json:"changed"`
	Replayed bool   `json:"replayed"`
	Entry    *Entry `json:"entry"`
}
