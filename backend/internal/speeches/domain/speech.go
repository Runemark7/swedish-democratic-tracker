package domain

import "time"

type Speech struct {
	ID              int       `json:"id"`
	DokID           string    `json:"dokId"`
	AnforandeNummer string    `json:"anforandeNummer"`
	PoliticianID    string    `json:"politicianId"`
	Party           string    `json:"party"`
	Date            time.Time `json:"date"`
	TopicHeading    string    `json:"topicHeading"`
	SpeechText      string    `json:"speechText"`
	RelatedDokID    string    `json:"relatedDokId,omitempty"`
	AIProcessed     bool      `json:"aiProcessed"`
	CreatedAt       time.Time `json:"createdAt"`
}
