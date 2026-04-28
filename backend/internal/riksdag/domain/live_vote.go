package domain

import "time"

type LiveVote struct {
	Beteckning string    `json:"beteckning"`
	Title      string    `json:"title"`
	Status     string    `json:"status"` // "Bifall" | "Avslag"
	JaCount    int       `json:"jaCount"`
	NejCount   int       `json:"nejCount"`
	Margin     string    `json:"margin"`
	Tag        string    `json:"tag"`
	Date       time.Time `json:"date"`
}
