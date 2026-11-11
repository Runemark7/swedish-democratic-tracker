package domain

import (
	"strconv"
	"strings"
	"time"
)

const (
	EntityVeckobatch = "veckobatch"

	KopplingBekraftad = "bekraftad"
	KopplingAvvisad   = "avvisad"

	AnswerJa       = "ja"
	AnswerNej      = "nej"
	AnswerJaOmvant = "ja-omvaent" // ja, with the direction reversed

	// DefaultBatchSize is also the cap: "at most about 20".
	DefaultBatchSize = 20
)

// BatchRequest creates a Veckobatch from the strongest open suggestions.
type BatchRequest struct {
	Max *int `json:"max"`
	Provenance
}

func (r *BatchRequest) Normalize() (int, error) {
	n := DefaultBatchSize
	if r.Max != nil {
		n = *r.Max
	}
	if n < 1 || n > DefaultBatchSize {
		return 0, invalid("max must be between 1 and %d", DefaultBatchSize)
	}
	return n, r.Provenance.Validate()
}

// BatchItem is one numbered line of a Veckobatch, with what the Redaktör needs
// to judge it in chat. Direction is the ballot in the Löfte's direction
// (flipped once answered ja-omvaent).
type BatchItem struct {
	Nr            int     `json:"nr"`
	KopplingID    int64   `json:"kopplingId"`
	State         string  `json:"state"`
	RolledOver    bool    `json:"rolledOver"`
	LoefteID      int64   `json:"loefteId"`
	Party         string  `json:"party"`
	LoefteText    string  `json:"loefteText"`
	Quote         string  `json:"quote"`
	Locator       string  `json:"locator"`
	SourceURL     string  `json:"sourceUrl"`
	VoteringID    string  `json:"voteringId"`
	Beteckning    string  `json:"beteckning"`
	Forslagspunkt string  `json:"forslagspunkt"`
	Betankande    string  `json:"betankande"`
	Datum         *string `json:"datum"`
	Direction     string  `json:"direction"`
	Confidence    float64 `json:"confidence"`
	Reasoning     string  `json:"reasoning"`
	Answer        *string `json:"answer"`
}

type Batch struct {
	ID        int64       `json:"id"`
	CreatedAt time.Time   `json:"createdAt"`
	ClosedAt  *time.Time  `json:"closedAt"`
	Replayed  bool        `json:"replayed,omitempty"`
	Items     []BatchItem `json:"items"`
}

func (b *Batch) Item(nr int) *BatchItem {
	for i := range b.Items {
		if b.Items[i].Nr == nr {
			return &b.Items[i]
		}
	}
	return nil
}

// OpenSuggestion is a foreslagen koppling waiting for a batch.
type OpenSuggestion struct {
	KopplingID int64
	RolledOver bool
}

// Answer decides one numbered item.
type Answer struct {
	Nr     int    `json:"nr"`
	Answer string `json:"answer"`
}

// AnswerRequest is the Redaktör's short-form answer. Give either Text (the
// chat sentence, parsed strictly) or Answers (+ optional Rest). Items that are
// neither named nor covered by Rest stay unanswered and roll over.
type AnswerRequest struct {
	Text    string   `json:"text"`
	Answers []Answer `json:"answers"`
	Rest    string   `json:"rest"`
	Provenance
}

func validAnswer(a string) bool { return a == AnswerJa || a == AnswerNej || a == AnswerJaOmvant }

// Normalize resolves Text into Answers/Rest and validates the shape. It does
// not know the batch; the service checks numbers against it.
func (r *AnswerRequest) Normalize() error {
	if err := r.Provenance.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(r.Text) != "" {
		if len(r.Answers) > 0 || r.Rest != "" {
			return invalid("give either text or answers/rest, not both")
		}
		a, rest, err := ParseAnswerText(r.Text)
		if err != nil {
			return err
		}
		r.Answers, r.Rest = a, rest
	}
	if r.Rest != "" && !validAnswer(r.Rest) {
		return invalid(`rest must be "ja", "nej" or "ja-omvaent"`)
	}
	seen := map[int]bool{}
	for _, a := range r.Answers {
		if a.Nr < 1 {
			return invalid("answer nr must be 1 or higher")
		}
		if !validAnswer(a.Answer) {
			return invalid(`item %d: answer must be "ja", "nej" or "ja-omvaent"`, a.Nr)
		}
		if seen[a.Nr] {
			return invalid("item %d is answered twice", a.Nr)
		}
		seen[a.Nr] = true
	}
	if len(r.Answers) == 0 && r.Rest == "" {
		return invalid("no answers given")
	}
	return nil
}

// ParseAnswerText reads the Redaktör's chat answer strictly: lists of item
// numbers (or ranges 2-4) followed by ja, nej, or "ja men omvänt", and
// optionally "resten"/"övriga" followed by one of those. "1,3,7 ja, resten
// nej" and "4 ja men omvänt" are the canonical examples. Anything it does not
// recognise is an error: the bot must never guess an answer.
func ParseAnswerText(text string) (answers []Answer, rest string, err error) {
	repl := strings.NewReplacer(",", " ", ";", " ", ".", " ", "\n", " ", "!", " ")
	toks := strings.Fields(strings.ToLower(repl.Replace(text)))
	var pending []int
	restPending := false
	seen := map[int]bool{}
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		switch {
		case t == "och":
		case t == "resten" || t == "övriga" || t == "ovriga":
			if restPending || rest != "" || len(pending) > 0 {
				return nil, "", invalid("answer text: unexpected %q", t)
			}
			restPending = true
		case t == "ja" || t == "nej":
			if len(pending) == 0 && !restPending {
				return nil, "", invalid("answer text: %q without item numbers", t)
			}
			ans := t
			if t == "ja" {
				j := i + 1
				if j < len(toks) && toks[j] == "men" {
					j++
				}
				if j < len(toks) && isReversed(toks[j]) {
					ans, i = AnswerJaOmvant, j
				}
			}
			if restPending {
				rest, restPending = ans, false
			}
			for _, n := range pending {
				answers = append(answers, Answer{Nr: n, Answer: ans})
			}
			pending = nil
		default:
			ns, ok := parseNumbers(t)
			if !ok || restPending {
				return nil, "", invalid("answer text: cannot read %q", t)
			}
			for _, n := range ns {
				if seen[n] {
					return nil, "", invalid("answer text: item %d appears twice", n)
				}
				seen[n] = true
			}
			pending = append(pending, ns...)
		}
	}
	if len(pending) > 0 || restPending {
		return nil, "", invalid("answer text: missing ja or nej at the end")
	}
	return answers, rest, nil
}

func isReversed(t string) bool {
	switch t {
	case "omvänt", "omvant", "omvänd", "omvand", "omvända", "omvanda":
		return true
	}
	return false
}

func parseNumbers(t string) ([]int, bool) {
	if a, b, ok := strings.Cut(t, "-"); ok {
		x, e1 := strconv.Atoi(a)
		y, e2 := strconv.Atoi(b)
		if e1 != nil || e2 != nil || x < 1 || y < x || y-x > 100 {
			return nil, false
		}
		var out []int
		for n := x; n <= y; n++ {
			out = append(out, n)
		}
		return out, true
	}
	n, err := strconv.Atoi(t)
	if err != nil || n < 1 {
		return nil, false
	}
	return []int{n}, true
}

// ItemOutcome is one decided item in the answer to AnswerRequest.
type ItemOutcome struct {
	Nr         int    `json:"nr"`
	KopplingID int64  `json:"kopplingId"`
	Answer     string `json:"answer"`
	State      string `json:"state"`
	Direction  string `json:"direction"`
	EntryID    int64  `json:"entryId"`
}

type AnswerResponse struct {
	Replayed   bool          `json:"replayed"`
	Decided    []ItemOutcome `json:"decided"`
	Unanswered []int         `json:"unanswered"` // item numbers still open; they roll over
	Closed     bool          `json:"closed"`
}

// Decision maps an answer to the koppling's new state and direction.
func Decision(answer, direction string) (state, dir string) {
	switch answer {
	case AnswerNej:
		return KopplingAvvisad, direction
	case AnswerJaOmvant:
		if direction == "Ja" {
			return KopplingBekraftad, "Nej"
		}
		return KopplingBekraftad, "Ja"
	}
	return KopplingBekraftad, direction
}
