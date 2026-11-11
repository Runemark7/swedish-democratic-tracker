package domain

import (
	"strings"
	"time"
)

// Outcomes a reader can filter the list by. Delvis bifallen and Blandat utfall
// are motion-level words; the others are yrkande-level and motion-level.
var FilterableOutcomes = []string{
	OutcomeAvslagen, OutcomeBifallen, OutcomeDelvis, OutcomeBesvarad, OutcomeEjAvgjord, OutcomeBlandat,
}

// Kinds a reader can filter the list by, in the order the UI shows them.
var FilterableKinds = []string{KindKommittemotion, KindPartimotion, KindEnskild, KindFlerparti}

// ListFilter narrows the motion list. Every field is optional; the values of
// one field are alternatives, the fields are combined with AND.
type ListFilter struct {
	Parties   []string // signer party, upper case
	Utskott   []string // utskott code ("SkU")
	Riksmoten []string // "2017/18"
	Kinds     []string // one of FilterableKinds
	Outcomes  []string // one of FilterableOutcomes; applied by the service
	Authors   []string // signer politician ids (intressent_id)
	Query     string   // free text over title and yrkande wording
}

// ListCursor is the keyset position after the last row returned. The list is
// ordered by (Decided, Key, DokID) descending.
type ListCursor struct {
	Decided bool
	Key     time.Time
	DokID   string
}

// ListRow is one motion in the list with what is needed to compute its
// outcome. Key is decided-at when Decided, else the filing date.
type ListRow struct {
	Motion   Motion
	Decided  bool
	Key      time.Time
	Cursor   ListCursor
	Decision map[int][]string // yrkande number → decision text per linked punkt
	// HandledIn lists the betänkanden that decided its yrkanden ("2018/19:SkU11").
	HandledIn []string
}

// Outcome aggregates the row's yrkanden the same way the motion page does.
func (r ListRow) Outcome() Aggregate {
	words := make([]string, 0, len(r.Motion.Yrkanden))
	for _, y := range r.Motion.Yrkanden {
		words = append(words, YrkandeWord(y.ChamberDecision, r.Decision[y.Number]))
	}
	return MotionOutcome(words)
}

// Parties lists the distinct signer parties in signing order.
func (m Motion) Parties() []string {
	seen := map[string]bool{}
	out := []string{}
	for _, a := range m.Authors {
		if a.Role != "" && !equalFold(a.Role, "undertecknare") {
			continue
		}
		p := a.Party
		if p == "" || p == "-" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

func equalFold(a, b string) bool { return strings.EqualFold(a, b) }
