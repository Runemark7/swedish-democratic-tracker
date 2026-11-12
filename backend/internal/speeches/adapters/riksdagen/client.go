package riksdagen

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
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
				DokRm           string `json:"dok_rm"`
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
			Riksmote:        riksmoteOf(a.DokRm, f.Session),
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return parseSpeechTextBody(body)
}

// parseSpeechTextBody reads anförandetext from a JSON body, or from the XML
// document Riksdagen sometimes serves on the .json URL (the body then starts
// with '<', and json.Decoder fails with invalid character '<'). An HTML error
// page or other markup without anförandetext is treated as unavailable
// (empty, no error) so a backfill can finish instead of retrying forever.
// Non-markup garbage still returns a decode error so a flaky response is retried.
func parseSpeechTextBody(body []byte) (string, error) {
	trim := bytes.TrimSpace(body)
	var payload struct {
		Anforande struct {
			Anforandetext string `json:"anforandetext"`
		} `json:"anforande"`
	}
	if err := json.Unmarshal(trim, &payload); err == nil {
		return payload.Anforande.Anforandetext, nil
	} else if len(trim) == 0 || trim[0] != '<' {
		return "", fmt.Errorf("decode speech text: %w", err)
	}
	var xmlPayload struct {
		Anforandetext string `xml:"anforandetext"`
	}
	if err := xml.Unmarshal(trim, &xmlPayload); err == nil {
		return xmlPayload.Anforandetext, nil
	}
	return "", nil
}

// riksmoteOf is the riksmöte the protocol names, or the one asked for when the
// row does not name it.
func riksmoteOf(dokRm, asked string) string {
	if rm := strings.TrimSpace(dokRm); rm != "" {
		return rm
	}
	return asked
}
