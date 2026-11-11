package http

import (
	"context"
	"strings"
	"unicode"

	"riksdagskollen/internal/motions/domain"
)

const excerptRunes = 240

// debateSpeechResponse is one anförande in a debate. The excerpt is the
// start of the speech, nothing selected or summarised.
type debateSpeechResponse struct {
	SpeechID     int    `json:"speechId"`
	PoliticianID string `json:"politicianId"`
	Name         string `json:"name"`
	Party        string `json:"party"`
	Date         string `json:"date"`
	Excerpt      string `json:"excerpt"`
	// ExcerptAvailable is false while the speech text has not been fetched.
	ExcerptAvailable bool `json:"excerptAvailable"`
	// RiksdagenURL is the speech on Riksdagen's site.
	RiksdagenURL string `json:"riksdagenUrl"`
	// LinkRule is how the speech was tied to the betänkande:
	// "rel_dok_id" or "avsnittsrubrik_datum".
	LinkRule string `json:"linkRule"`
}

// debateResponse is the chamber debate on a betänkande. Total can exceed the
// number of speeches listed.
type debateResponse struct {
	Date     *string                `json:"date"`
	Total    int                    `json:"total"`
	Speeches []debateSpeechResponse `json:"speeches"`
}

func toDebate(d *domain.Debate) *debateResponse {
	out := &debateResponse{Total: d.Total, Speeches: make([]debateSpeechResponse, 0, len(d.Speeches))}
	if d.Total > 0 && !d.Date.IsZero() {
		s := d.Date.Format("2006-01-02")
		out.Date = &s
	}
	for _, s := range d.Speeches {
		ex := excerptOf(s.Excerpt)
		out.Speeches = append(out.Speeches, debateSpeechResponse{
			SpeechID: s.SpeechID, PoliticianID: s.PoliticianID, Name: s.Name, Party: s.Party,
			Date: s.Date.Format("2006-01-02"), Excerpt: ex, ExcerptAvailable: ex != "",
			RiksdagenURL: "https://data.riksdagen.se/anforande/" + s.DokID + "-" + s.AnforandeNummer + "/html",
			LinkRule:     s.LinkRule,
		})
	}
	return out
}

// excerptOf turns the start of a speech (HTML) into one line of plain text of
// at most excerptRunes runes, cut at a word.
func excerptOf(raw string) string {
	var b strings.Builder
	inTag := false
	for _, r := range raw {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	text := strings.Join(strings.FieldsFunc(b.String(), unicode.IsSpace), " ")
	r := []rune(text)
	if len(r) <= excerptRunes {
		return text
	}
	cut := string(r[:excerptRunes])
	if i := strings.LastIndex(cut, " "); i > excerptRunes/2 {
		cut = cut[:i]
	}
	return cut + "…"
}

// debates reads each betänkande's debate once per request.
type debates struct {
	read  func(ctx context.Context, betDokID string) (*domain.Debate, error)
	cache map[string]*debateResponse
}

func (d *debates) get(ctx context.Context, betDokID string) (*debateResponse, error) {
	if betDokID == "" {
		return nil, nil
	}
	if r, ok := d.cache[betDokID]; ok {
		return r, nil
	}
	raw, err := d.read(ctx, betDokID)
	if err != nil {
		return nil, err
	}
	r := toDebate(raw)
	d.cache[betDokID] = r
	return r, nil
}
