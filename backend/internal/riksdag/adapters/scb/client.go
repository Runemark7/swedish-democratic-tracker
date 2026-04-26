package scb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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

const historyFromYear = 2015

func (c *Client) fetch(ctx context.Context) ([]ports.HeadcountData, error) {
	latestYear := time.Now().Year() - 1
	years := make([]string, 0, latestYear-historyFromYear+1)
	for y := historyFromYear; y <= latestYear; y++ {
		years = append(years, strconv.Itoa(y))
	}

	body := map[string]any{
		"query": []map[string]any{
			{
				"code":      "ContentsCode",
				"selection": map[string]any{"filter": "item", "values": []string{"OE0108A1"}},
			},
			{
				"code":      "Tid",
				"selection": map[string]any{"filter": "item", "values": years},
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
		Data []struct {
			Key    []string `json:"key"`
			Values []string `json:"values"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("scb headcount decode: %w", err)
	}

	// Aggregate per agency: collect all years into HeadcountHistory.
	type agencyAcc struct {
		latestYear int
		latest     int
		history    []ports.YearlyHeadcount
	}
	byName := map[string]*agencyAcc{}

	for _, row := range result.Data {
		if len(row.Key) < 2 || len(row.Values) == 0 {
			continue
		}
		canonical, ok := scbNameMap[row.Key[0]]
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
		acc := byName[canonical]
		if acc == nil {
			acc = &agencyAcc{}
			byName[canonical] = acc
		}
		acc.history = append(acc.history, ports.YearlyHeadcount{Year: yr, HeadcountInt: count})
		if yr > acc.latestYear {
			acc.latestYear = yr
			acc.latest = count
		}
	}

	if len(byName) == 0 {
		return nil, fmt.Errorf("scb headcount: no matching agencies in response")
	}

	out := make([]ports.HeadcountData, 0, len(byName))
	for name, acc := range byName {
		// Sort history ascending by year.
		sort.Slice(acc.history, func(i, j int) bool { return acc.history[i].Year < acc.history[j].Year })
		out = append(out, ports.HeadcountData{
			Name:             name,
			HeadcountInt:     acc.latest,
			Year:             acc.latestYear,
			HeadcountHistory: acc.history,
		})
	}
	return out, nil
}
