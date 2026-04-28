package domain

import "time"

type GovernmentParty struct {
	Name      string `json:"name"`
	Short     string `json:"short"`
	Seats     int    `json:"seats"`
	Color     string `json:"color"`
	Role      string `json:"role"`
	SortOrder int    `json:"-"`
}

type Government struct {
	TypeLabel  string            `json:"type"`
	ValidFrom  time.Time         `json:"validFrom"`
	Parties    []GovernmentParty `json:"parties"`
	Support    []GovernmentParty `json:"support,omitempty"`
	Opposition []GovernmentParty `json:"opposition"`
}
