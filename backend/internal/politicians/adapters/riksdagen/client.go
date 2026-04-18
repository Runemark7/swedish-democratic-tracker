package riksdagen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"riksdagskollen/internal/politicians/domain"
)

const baseURL = "https://data.riksdagen.se"

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchMembers fetches current members for a party.
// status: "tjanstgorande" for active members.
func (c *Client) FetchMembers(ctx context.Context, party, status string) ([]*domain.Politician, error) {
	url := fmt.Sprintf("%s/personlista/?rdlstatus=%s&parti=%s&utformat=json", baseURL, status, party)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("riksdagen API returned %d for members", resp.StatusCode)
	}

	var payload struct {
		Personlista struct {
			Person []struct {
				IntressentID string `json:"intressent_id"`
				Tilltalsnamn string `json:"tilltalsnamn"`
				Efternamn    string `json:"efternamn"`
				Parti        string `json:"parti"`
				Valkrets     string `json:"valkrets"`
				BildURL192   string `json:"bild_url_192"`
				Status       string `json:"status"`
			} `json:"person"`
		} `json:"personlista"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode members: %w", err)
	}

	politicians := make([]*domain.Politician, 0, len(payload.Personlista.Person))
	for _, p := range payload.Personlista.Person {
		politicians = append(politicians, &domain.Politician{
			IntressentID: p.IntressentID,
			FirstName:    p.Tilltalsnamn,
			LastName:     p.Efternamn,
			Party:        p.Parti,
			Constituency: p.Valkrets,
			ImageURL:     p.BildURL192,
			IsActive:     strings.Contains(strings.ToLower(p.Status), "tjänstgörande"),
			UpdatedAt:    time.Now(),
		})
	}
	return politicians, nil
}
