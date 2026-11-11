package domain

import (
	"fmt"
	"strings"
)

// Curated motion links (issue #163). A motion koppling is a link the bot
// suggests and the Redaktör confirms in the Veckobatch: either a fråga or
// interpellation that cites a motion only in prose (kind "fraga"), or two
// yrkanden of different wording that are the same Återkommande krav (kind
// "krav"). Only a bekraftad one is ever public; the rule it was made under
// stays recorded.
const (
	EntityMotionKoppling = "motionkoppling"

	KindFraga = "fraga"
	KindKrav  = "krav"

	RuleProseCitation  = "prose-citation"
	RuleSimilarWording = "similar-wording"

	// Item kinds in a Veckobatch.
	ItemLoefte = "loefte"
)

// QuestionRef names a fråga or interpellation.
type QuestionRef struct {
	Kind       string `json:"kind"` // fraga | interpellation
	Riksmote   string `json:"riksmote"`
	Beteckning string `json:"beteckning"`
}

// MotionRef names a motion, and for kind "krav" one of its yrkanden.
type MotionRef struct {
	Riksmote   string `json:"riksmote"`
	Beteckning string `json:"beteckning"`
	Yrkande    int    `json:"yrkande,omitempty"`
}

// MotionSuggestion is one proposed link from the bot.
//
//	kind "fraga": Question + Motion (the cited motion)
//	kind "krav":  Motion + Other, each with its Yrkande number
type MotionSuggestion struct {
	Kind       string       `json:"kind"`
	Question   *QuestionRef `json:"question,omitempty"`
	Motion     MotionRef    `json:"motion"`
	Other      *MotionRef   `json:"other,omitempty"`
	Confidence float64      `json:"confidence"`
	Reasoning  string       `json:"reasoning"`
}

type MotionSuggestionBatch struct {
	Threshold   *float64           `json:"threshold"`
	Suggestions []MotionSuggestion `json:"suggestions"`
	Provenance
}

func (r MotionRef) key() string { return fmt.Sprintf("%s:%s/%d", r.Riksmote, r.Beteckning, r.Yrkande) }

func (b *MotionSuggestionBatch) Normalize() error {
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
		at := fmt.Sprintf("suggestion %d", i+1)
		s.Reasoning = strings.TrimSpace(s.Reasoning)
		s.Motion.Riksmote, s.Motion.Beteckning = strings.TrimSpace(s.Motion.Riksmote), strings.TrimSpace(s.Motion.Beteckning)
		if s.Motion.Riksmote == "" || s.Motion.Beteckning == "" {
			return invalid("%s: motion riksmote and beteckning are required", at)
		}
		if !(s.Confidence >= 0 && s.Confidence <= 1) {
			return invalid("%s: confidence must be between 0 and 1", at)
		}
		if s.Reasoning == "" || len(s.Reasoning) > 4000 {
			return invalid("%s: reasoning is required (max 4000 characters)", at)
		}
		var k string
		switch s.Kind {
		case KindFraga:
			q := s.Question
			if q == nil || (q.Kind != "fraga" && q.Kind != "interpellation") || strings.TrimSpace(q.Riksmote) == "" || strings.TrimSpace(q.Beteckning) == "" {
				return invalid(`%s: question needs kind "fraga" or "interpellation", riksmote and beteckning`, at)
			}
			if s.Other != nil || s.Motion.Yrkande != 0 {
				return invalid("%s: a fraga suggestion names a question and a motion only", at)
			}
			k = fmt.Sprintf("fraga|%s|%s|%s|%s", q.Kind, q.Riksmote, q.Beteckning, s.Motion.key())
		case KindKrav:
			o := s.Other
			if o == nil || strings.TrimSpace(o.Riksmote) == "" || strings.TrimSpace(o.Beteckning) == "" {
				return invalid("%s: other (riksmote, beteckning, yrkande) is required", at)
			}
			if s.Question != nil {
				return invalid("%s: a krav suggestion names two yrkanden, not a question", at)
			}
			if s.Motion.Yrkande < 1 || o.Yrkande < 1 {
				return invalid("%s: both yrkande numbers must be 1 or higher", at)
			}
			if s.Motion.key() == o.key() {
				return invalid("%s: a yrkande cannot be linked to itself", at)
			}
			// The pair is unordered: keep the earlier-named side first so the
			// same pair always reads the same.
			a, c := s.Motion, *o
			if a.key() > c.key() {
				s.Motion, s.Other = c, &a
			}
			k = fmt.Sprintf("krav|%s|%s", s.Motion.key(), s.Other.key())
		default:
			return invalid(`%s: kind must be "fraga" or "krav"`, at)
		}
		if seen[k] {
			return invalid("%s: the same link appears twice", at)
		}
		seen[k] = true
	}
	return b.Provenance.Validate()
}

// MotionSuggestionResult is one line of the answer; Status is "created" or
// "existing" (the pair was suggested before and is left as it is).
type MotionSuggestionResult struct {
	MotionKopplingID int64  `json:"motionKopplingId"`
	Kind             string `json:"kind"`
	State            string `json:"state"`
	Status           string `json:"status"`
}

type MotionSuggestionResponse struct {
	Replayed bool                     `json:"replayed"`
	Results  []MotionSuggestionResult `json:"results"`
}

// MotionKoppling is the audit snapshot of a motion link. Never served publicly.
type MotionKoppling struct {
	ID            int64   `json:"id"`
	Kind          string  `json:"kind"`
	Rule          string  `json:"rule"`
	State         string  `json:"state"`
	QuestionDokID string  `json:"questionDokId,omitempty"`
	MotionDokID   string  `json:"motionDokId"`
	Yrkande       int     `json:"yrkande,omitempty"`
	OtherMotionID string  `json:"otherMotionDokId,omitempty"`
	OtherYrkande  int     `json:"otherYrkande,omitempty"`
	Confidence    float64 `json:"confidence"`
	Threshold     float64 `json:"threshold"`
	Reasoning     string  `json:"reasoning"`
	Origin        string  `json:"origin"`
	VeckobatchID  *int64  `json:"veckobatchId,omitempty"`
	Nr            *int    `json:"nr,omitempty"`
	Answer        string  `json:"answer,omitempty"`
}

// MotionBatchItem is what the Redaktör needs to judge a motion link in chat.
type MotionBatchItem struct {
	Kind     string        `json:"kind"`
	Rule     string        `json:"rule"`
	Question *QuestionView `json:"question,omitempty"`
	Motion   MotionView    `json:"motion"`
	Other    *MotionView   `json:"other,omitempty"`
}

type QuestionView struct {
	Kind       string `json:"kind"`
	Riksmote   string `json:"riksmote"`
	Beteckning string `json:"beteckning"`
	Title      string `json:"title"`
	Asker      string `json:"asker"`
	Excerpt    string `json:"excerpt"`
	SourceURL  string `json:"sourceUrl"`
}

type MotionView struct {
	Riksmote   string   `json:"riksmote"`
	Beteckning string   `json:"beteckning"`
	Title      string   `json:"title"`
	Parties    []string `json:"parties"`
	Yrkande    int      `json:"yrkande,omitempty"`
	Wording    string   `json:"wording,omitempty"`
	SourceURL  string   `json:"sourceUrl"`
}

// MotionDecision maps an answer to the new state. "ja men omvänt" only means
// something for a Löfte koppling (it has a direction); here it is refused.
func MotionDecision(answer string) (string, error) {
	switch answer {
	case AnswerJa:
		return KopplingBekraftad, nil
	case AnswerNej:
		return KopplingAvvisad, nil
	}
	return "", invalid(`"ja men omvänt" has no meaning for a motion link; answer ja or nej`)
}
