package domain

import "time"

type Politician struct {
	IntressentID string    `json:"intressentId"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Party        string    `json:"party"`
	Constituency string    `json:"constituency,omitempty"`
	ImageURL     string    `json:"imageUrl,omitempty"`
	IsActive     bool      `json:"isActive"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (p *Politician) FullName() string {
	return p.FirstName + " " + p.LastName
}
