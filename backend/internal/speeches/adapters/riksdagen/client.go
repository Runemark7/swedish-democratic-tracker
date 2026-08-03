package riksdagen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"riksdagskollen/internal/speeches/domain"
	"riksdagskollen/internal/speeches/ports"
)

const defaultBaseURL = "https://data.riksdagen.se"

type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient() *Client {
	return &Client{
		http:    &http.Client{Timeout: 30 * time.Second},
		baseURL: defaultBaseURL,
	}
}

func (c *Client) FetchSpeeches(ctx context.Context, f ports.FetchSpeechesFilter) ([]*domain.Speech, error) {
	size := f.Size
	if size == 0 {
		size = 200
	}
	url := fmt.Sprintf("%s/anforandelista/?rm=%s&parti=%s&iid=%s&sz=%d&anftyp=Akt&utformat=json",
		c.baseURL, f.Session, f.Party, f.PoliticianID, size)

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
		return nil, fmt.Errorf("riksdagen speeches API returned %d", resp.StatusCode)
	}

	var payload struct {
		Anforandelista struct {
			Anforande []struct {
				DokID           string `json:"dok_id"`
				AnforandeNummer string `json:"anforande_nummer"`
				IntressentID    string `json:"intressent_id"`
				Parti           string `json:"parti"`
				DokDatum        string `json:"dok_datum"`
				Avsnittsrubrik  string `json:"avsnittsrubrik"`
				Anforandetext   string `json:"anforandetext"`
				RelDokID        string `json:"rel_dok_id"`
			} `json:"anforande"`
		} `json:"anforandelista"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode speeches: %w", err)
	}

	speeches := make([]*domain.Speech, 0, len(payload.Anforandelista.Anforande))
	for _, a := range payload.Anforandelista.Anforande {
		date, _ := time.Parse("2006-01-02", strings.TrimSpace(a.DokDatum))
		if !f.Since.IsZero() && !date.After(f.Since) {
			continue
		}
		speeches = append(speeches, &domain.Speech{
			DokID:           a.DokID,
			AnforandeNummer: a.AnforandeNummer,
			PoliticianID:    a.IntressentID,
			Party:           a.Parti,
			Date:            date,
			TopicHeading:    a.Avsnittsrubrik,
			SpeechText:      a.Anforandetext,
			RelatedDokID:    a.RelDokID,
		})
	}
	return speeches, nil
}

// FetchSpeechText calls /anforande/{dokID}-{nr}.json and returns the
// `anforandetext` body. The list endpoint omits the prose, so this is
// the only way to populate `speeches.speech_text`.
func (c *Client) FetchSpeechText(ctx context.Context, dokID, anforandeNummer string) (string, error) {
	url := fmt.Sprintf("%s/anforande/%s-%s.json", c.baseURL, dokID, anforandeNummer)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", nil // nothing upstream — caller can mark text empty/sentinel
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("riksdagen speech text returned %d", resp.StatusCode)
	}
	var payload struct {
		Anforande struct {
			Anforandetext string `json:"anforandetext"`
		} `json:"anforande"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode speech text: %w", err)
	}
	return payload.Anforande.Anforandetext, nil
}
