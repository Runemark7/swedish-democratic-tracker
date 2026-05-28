package domain

import "time"

type Party struct {
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	FoundedYear *int      `json:"foundedYear,omitempty"`
	Ideology    *string   `json:"ideology,omitempty"`
	ColorHex    string    `json:"colorHex"`
	TextHex     string    `json:"textHex"`
	Active      bool      `json:"active"`
	WebsiteURL  *string   `json:"websiteUrl,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}
