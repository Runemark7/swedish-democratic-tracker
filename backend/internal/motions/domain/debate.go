package domain

import "time"

// DebateSpeech is one anförande of the debate on a betänkande.
type DebateSpeech struct {
	SpeechID        int
	DokID           string // the protocol
	AnforandeNummer string
	PoliticianID    string
	Name, Party     string
	Date            time.Time
	// Excerpt is the start of the speech as plain text; empty while the text
	// has not been fetched.
	Excerpt string
	// LinkRule is how the speech was tied to the betänkande.
	LinkRule string
}

// Debate is the chamber debate on one betänkande: the speeches linked to it,
// in speaking order. Total can exceed len(Speeches) when the list is capped.
type Debate struct {
	Date     time.Time // date of the first speech
	Total    int
	Speeches []DebateSpeech
}
