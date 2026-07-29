package domain

import "time"

type Specificity string

const (
	SpecificityConcrete     Specificity = "concrete"
	SpecificityDirectional  Specificity = "directional"
	SpecificityRhetorical   Specificity = "rhetorical"
)

type Goal struct {
	ID                 int         `json:"id"`
	Party              string      `json:"party"`
	GoalText           string      `json:"goalText"`
	Topic              string      `json:"topic"`
	Specificity        Specificity `json:"specificity"`
	SourceDocument     string      `json:"sourceDocument"`
	SourceURL          string      `json:"sourceUrl,omitempty"`
	SourceQuote        string      `json:"sourceQuote,omitempty"`
	Keywords           []string    `json:"keywords,omitempty"`
	RelevantCommittees []string    `json:"relevantCommittees,omitempty"`
	CreatedAt          time.Time   `json:"createdAt"`
}
