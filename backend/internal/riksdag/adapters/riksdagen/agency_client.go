package riksdagen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"riksdagskollen/internal/riksdag/ports"
)

const (
	baseURL    = "https://data.riksdagen.se/dokumentlista/"
	docBaseURL = "https://data.riksdagen.se/dokument/"
	minYear    = 2018
)

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

func stripHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, " ")
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
}

type dokument struct {
	DokID          string `json:"dok_id"`
	Datum          string `json:"datum"`
	Titel          string `json:"titel"`
	Sammanfattning string `json:"sammanfattning"`
	Typ            string `json:"typ"`
}

type dokumentlista struct {
	Dokument []dokument `json:"dokument"`
}

type apiResponse struct {
	Dokumentlista dokumentlista `json:"dokumentlista"`
}

type AgencyClient struct {
	http *http.Client
}

func NewAgencyClient() *AgencyClient {
	return &AgencyClient{
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *AgencyClient) fetch(ctx context.Context, params url.Values) ([]dokument, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("riksdagen API status %d", resp.StatusCode)
	}

	var out apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Dokumentlista.Dokument, nil
}

func toDoc(d dokument, docType string) (ports.RiksdagenDoc, error) {
	t, err := time.Parse("2006-01-02", d.Datum)
	if err != nil {
		return ports.RiksdagenDoc{}, fmt.Errorf("parse datum %q: %w", d.Datum, err)
	}
	return ports.RiksdagenDoc{
		DokID:   d.DokID,
		Year:    t.Year(),
		Date:    t,
		Title:   strings.TrimSpace(d.Titel),
		Summary: stripHTML(d.Sammanfattning),
		DocType: docType,
		URL:     docBaseURL + d.DokID + ".html",
	}, nil
}

// FetchRegleringsbrev fetches yearly regulatory instructions for an agency.
// Returns one document per year (newest per year), only years >= minYear.
func (c *AgencyClient) FetchRegleringsbrev(ctx context.Context, agencyName string) ([]ports.RiksdagenDoc, error) {
	params := url.Values{
		"doktyp":   {"Rb"},
		"titel":    {agencyName},
		"utformat": {"json"},
		"sz":       {"30"},
	}
	raw, err := c.fetch(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetch regleringsbrev %q: %w", agencyName, err)
	}

	byYear := map[int]ports.RiksdagenDoc{}
	for _, d := range raw {
		doc, err := toDoc(d, "Rb")
		if err != nil || doc.Year < minYear {
			continue
		}
		if existing, ok := byYear[doc.Year]; !ok || doc.Date.After(existing.Date) {
			byYear[doc.Year] = doc
		}
	}

	result := make([]ports.RiksdagenDoc, 0, len(byYear))
	for _, doc := range byYear {
		result = append(result, doc)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Year > result[j].Year })
	return result, nil
}

// FetchDecisions fetches recent committee reports (betänkanden) mentioning the agency.
func (c *AgencyClient) FetchDecisions(ctx context.Context, agencyName string) ([]ports.RiksdagenDoc, error) {
	params := url.Values{
		"sok":      {agencyName},
		"doktyp":   {"bet"},
		"utformat": {"json"},
		"sz":       {"20"},
	}
	raw, err := c.fetch(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetch decisions %q: %w", agencyName, err)
	}

	result := make([]ports.RiksdagenDoc, 0, len(raw))
	for _, d := range raw {
		doc, err := toDoc(d, "bet")
		if err != nil {
			continue
		}
		result = append(result, doc)
	}
	return result, nil
}
