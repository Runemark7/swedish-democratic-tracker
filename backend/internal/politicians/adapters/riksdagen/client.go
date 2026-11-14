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

const defaultBaseURL = "https://data.riksdagen.se"

type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient() *Client { return NewClientWithBaseURL("") }

// NewClientWithBaseURL points the client at baseURL. An empty base URL uses
// the production data.riksdagen.se origin.
func NewClientWithBaseURL(baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		http:    &http.Client{Timeout: 30 * time.Second},
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

// FetchMembers fetches current members for a party.
// status: "tjanstgorande" for active members.
func (c *Client) FetchMembers(ctx context.Context, party, status string) ([]*domain.Politician, error) {
	url := fmt.Sprintf("%s/personlista/?rdlstatus=%s&parti=%s&utformat=json", c.baseURL, status, party)

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
			Person []rawPerson `json:"person"`
		} `json:"personlista"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode members: %w", err)
	}

	now := time.Now()
	politicians := make([]*domain.Politician, 0, len(payload.Personlista.Person))
	for _, p := range payload.Personlista.Person {
		politicians = append(politicians, p.toPolitician(now))
	}
	return politicians, nil
}

// FetchPerson reads one person from /personlista/?iid=. Riksdagen answers with
// a JSON object, not a list, when there is exactly one hit, and with no person
// when it does not know the id.
func (c *Client) FetchPerson(ctx context.Context, intressentID string) (*domain.Politician, error) {
	url := fmt.Sprintf("%s/personlista/?iid=%s&utformat=json", c.baseURL, intressentID)
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
		return nil, fmt.Errorf("riksdagen API returned %d for person %s", resp.StatusCode, intressentID)
	}

	var payload struct {
		Personlista struct {
			Person json.RawMessage `json:"person"`
		} `json:"personlista"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode person: %w", err)
	}
	raw := payload.Personlista.Person
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var people []rawPerson
	if raw[0] == '[' {
		if err := json.Unmarshal(raw, &people); err != nil {
			return nil, fmt.Errorf("decode person: %w", err)
		}
	} else {
		var one rawPerson
		if err := json.Unmarshal(raw, &one); err != nil {
			return nil, fmt.Errorf("decode person: %w", err)
		}
		people = []rawPerson{one}
	}
	for _, p := range people {
		if p.IntressentID != intressentID {
			continue
		}
		return p.toPolitician(time.Now()), nil
	}
	return nil, nil
}
