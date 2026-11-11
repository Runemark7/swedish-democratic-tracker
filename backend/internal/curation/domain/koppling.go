package domain

import (
	"fmt"
	"strings"
)

const (
	EntityKoppling = "koppling"

	KopplingForeslagen   = "foreslagen"
	KopplingBortsorterad = "bortsorterad"

	// DefaultThreshold is the confidence below which a suggestion is dropped.
	// Open question in the spec (0.6); the request may override it.
	DefaultThreshold = 0.6
	MaxSuggestions   = 200
)

// Suggestion is one proposed Löfte↔votering link from the bot.
type Suggestion struct {
	LoefteID   int64   `json:"loefteId"`
	VoteringID string  `json:"voteringId"` // Riksdagen's votering_id
	Direction  string  `json:"direction"`  // the ballot, Ja or Nej, that is in the Löfte's direction
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

// SuggestionBatch is the submit payload. Threshold is part of the request so
// it can be tuned; nil means DefaultThreshold.
type SuggestionBatch struct {
	Threshold   *float64     `json:"threshold"`
	Suggestions []Suggestion `json:"suggestions"`
	Provenance
}

func (b *SuggestionBatch) Normalize() error {
	if b.Threshold == nil {
		t := DefaultThreshold
		b.Threshold = &t
	}
	if *b.Threshold < 0 || *b.Threshold > 1 {
		return invalid("threshold must be between 0 and 1")
	}
	if len(b.Suggestions) == 0 || len(b.Suggestions) > MaxSuggestions {
		return invalid("suggestions must hold 1 to %d items", MaxSuggestions)
	}
	seen := map[string]bool{}
	for i := range b.Suggestions {
		s := &b.Suggestions[i]
		s.VoteringID = strings.TrimSpace(s.VoteringID)
		s.Reasoning = strings.TrimSpace(s.Reasoning)
		at := fmt.Sprintf("suggestion %d", i+1)
		if s.LoefteID <= 0 || s.VoteringID == "" {
			return invalid("%s: loefteId and voteringId are required", at)
		}
		if s.Direction != "Ja" && s.Direction != "Nej" {
			return invalid(`%s: direction must be "Ja" or "Nej"`, at)
		}
		if !(s.Confidence >= 0 && s.Confidence <= 1) {
			return invalid("%s: confidence must be between 0 and 1", at)
		}
		if s.Reasoning == "" || len(s.Reasoning) > 4000 {
			return invalid("%s: reasoning is required (max 4000 characters)", at)
		}
		k := fmt.Sprintf("%d/%s", s.LoefteID, s.VoteringID)
		if seen[k] {
			return invalid("%s: the same Löfte and votering appear twice", at)
		}
		seen[k] = true
	}
	return b.Provenance.Validate()
}

// Koppling is the audit snapshot of a suggestion. It is never served publicly.
type Koppling struct {
	ID         int64   `json:"id"`
	LoefteID   int64   `json:"loefteId"`
	VoteringID string  `json:"voteringId"`
	Direction  string  `json:"direction"`
	State      string  `json:"state"`
	Confidence float64 `json:"confidence"`
	Threshold  float64 `json:"threshold"`
	Reasoning  string  `json:"reasoning"`
	Origin     string  `json:"origin"`
	// Set once the koppling has been put to the Redaktör in a Veckobatch.
	VeckobatchID *int64 `json:"veckobatchId,omitempty"`
	Nr           *int   `json:"nr,omitempty"`
	Answer       string `json:"answer,omitempty"` // "" while unanswered
}

// SuggestionResult is one line of the answer. Status is "created" or "existing"
// (the pair was suggested before and is left as it is).
type SuggestionResult struct {
	KopplingID int64  `json:"kopplingId"`
	LoefteID   int64  `json:"loefteId"`
	VoteringID string `json:"voteringId"`
	State      string `json:"state"`
	Status     string `json:"status"`
}

type SuggestionResponse struct {
	Replayed bool               `json:"replayed"`
	Results  []SuggestionResult `json:"results"`
}

// Votering is what the mandate-period rule needs to know about a votering.
type Votering struct {
	Ref        int64
	VoteringID string
	InPeriod   bool // inside the Löfte's mandate period
}

// Candidate types: the read the bot uses to generate suggestions.
type CandidateMotion struct {
	Riksmote   string `json:"riksmote"`
	Beteckning string `json:"beteckning"`
	Title      string `json:"title"`
	Party      string `json:"party"`
	Yrkande    int    `json:"yrkande"`
	Wording    string `json:"wording"`
}

type CandidateLoefte struct {
	ID        int64    `json:"id"`
	Party     string   `json:"party"`
	Text      string   `json:"text"`
	Quote     string   `json:"quote"`
	Locator   string   `json:"locator"`
	SourceURL string   `json:"sourceUrl"`
	Topic     string   `json:"topic"`
	Utskott   []string `json:"utskott"`
}

type CandidateBetankande struct {
	Riksmote   string `json:"riksmote"`
	Beteckning string `json:"beteckning"`
	Title      string `json:"title"`
	Punkt      string `json:"punkt"`
	Rubrik     string `json:"rubrik"`
	Forslag    string `json:"forslagText"`
	Alternativ string `json:"alternativ"`
}

type Candidate struct {
	VoteringID  string              `json:"voteringId"`
	Beteckning  string              `json:"beteckning"`
	Forslagspkt string              `json:"forslagspunkt"`
	Session     string              `json:"session"`
	Datum       *string             `json:"datum"`
	Utskott     string              `json:"utskott"`
	Betankande  CandidateBetankande `json:"betankande"`
	Motions     []CandidateMotion   `json:"motions"`
	Loften      []CandidateLoefte   `json:"loften"`
}

type Candidates struct {
	Since      *string     `json:"since"` // closed_at of the last closed Veckobatch; null when there is none
	Next       *int64      `json:"next"`  // cursor for the next page; null on the last
	Voteringar []Candidate `json:"voteringar"`
}
