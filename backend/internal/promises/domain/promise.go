package domain

import "time"

type Specificity string

const (
	SpecificityConcrete    Specificity = "concrete"
	SpecificityDirectional Specificity = "directional"
	SpecificityRhetorical  Specificity = "rhetorical"
)

type Promise struct {
	ID           int         `json:"id"`
	PoliticianID string      `json:"politicianId"`
	SpeechID     int         `json:"speechId"`
	PromiseText  string      `json:"promiseText"`
	Topic        string      `json:"topic"`
	Specificity  Specificity `json:"specificity"`
	Keywords     []string    `json:"keywords,omitempty"`
	ExtractedAt  time.Time   `json:"extractedAt"`
}
