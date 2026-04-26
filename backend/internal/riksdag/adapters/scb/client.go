package scb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"riksdagskollen/internal/riksdag/ports"
)

// headcountURL is the SCB PxWebApi endpoint for government employees by authority.
const headcountURL = "https://api.scb.se/OV0104/v1/doris/sv/ssd/OE/OE0108/OE0108A/OE0108T01A"

// scbNameMap maps SCB agency names to our canonical agency names.
// SCB uses full legal names; we match our 10 tracked agencies.
var scbNameMap = map[string]string{
	"Polismyndigheten":                             "Polismyndigheten",
	"Säkerhetspolisen":                             "Säkerhetspolisen",
	"Åklagarmyndigheten":                           "Åklagarmyndigheten",
	"Sveriges Domstolar":                           "Sveriges Domstolar",
	"Kriminalvården":                               "Kriminalvården",
	"Skatteverket":                                 "Skatteverket",
	"Tullverket":                                   "Tullverket",
	"Migrationsverket":                             "Migrationsverket",
	"Försäkringskassan":                            "Försäkringskassan",
	"Arbetsförmedlingen":                           "Arbetsförmedlingen",
}

type Client struct {
	http     *http.Client
	mu       sync.Mutex
	cached   []ports.HeadcountData
	cachedAt time.Time
	cacheTTL time.Duration
}

func NewClient() *Client {
	return &Client{
		http:     &http.Client{Timeout: 15 * time.Second},
		cacheTTL: 24 * time.Hour,
	}
}

func (c *Client) FetchHeadcounts(ctx context.Context) ([]ports.HeadcountData, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.cached) > 0 && time.Since(c.cachedAt) < c.cacheTTL {
		return c.cached, nil
	}

	data, err := c.fetch(ctx)
	if err != nil {
		return nil, err
	}
	c.cached = data
	c.cachedAt = time.Now()
	return data, nil
}

func (c *Client) fetch(ctx context.Context) ([]ports.HeadcountData, error) {
	year := time.Now().Year() - 1
	body := map[string]any{
		"query": []map[string]any{
			{
				"code":      "ContentsCode",
				"selection": map[string]any{"filter": "item", "values": []string{"OE0108A1"}},
			},
			{
				"code":      "Tid",
				"selection": map[string]any{"filter": "item", "values": []string{strconv.Itoa(year)}},
			},
		},
		"response": map[string]any{"format": "json"},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("scb headcount marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, headcountURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("scb headcount request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scb headcount: HTTP %d", resp.StatusCode)
	}

	var result struct {
		Columns []struct {
			Code  string `json:"code"`
			Text  string `json:"text"`
			Type  string `json:"type"`
		} `json:"columns"`
		Data []struct {
			Key    []string `json:"key"`
			Values []string `json:"values"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("scb headcount decode: %w", err)
	}

	var out []ports.HeadcountData
	for _, row := range result.Data {
		if len(row.Key) < 2 || len(row.Values) == 0 {
			continue
		}
		scbName := row.Key[0]
		canonical, ok := scbNameMap[scbName]
		if !ok {
			continue
		}
		yr, err := strconv.Atoi(row.Key[1])
		if err != nil {
			continue
		}
		count, err := strconv.Atoi(row.Values[0])
		if err != nil {
			continue
		}
		out = append(out, ports.HeadcountData{
			Name:         canonical,
			HeadcountInt: count,
			Year:         yr,
		})
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("scb headcount: no matching agencies in response")
	}
	return out, nil
}
