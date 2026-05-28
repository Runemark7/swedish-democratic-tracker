package domain

import "time"

type Minister struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Title          string    `json:"title"`
	Department     string    `json:"department"`
	DepartmentCode string    `json:"departmentCode"`
	Party          string    `json:"party"`
	PoliticianID   *string   `json:"politicianId,omitempty"`
	PhotoURL       *string   `json:"photoUrl,omitempty"`
	Bio            *string   `json:"bio,omitempty"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Proposal struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	DepartmentCode string     `json:"departmentCode"`
	RiksdagYear    string     `json:"riksdagYear"`
	PublishedAt    *time.Time `json:"publishedAt,omitempty"`
	URL            string     `json:"url"`
}
