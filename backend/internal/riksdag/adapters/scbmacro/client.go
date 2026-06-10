// Package scbmacro fetches national macro indicators from SCB's PxWeb API:
// the consumer price index (for the 12-month inflation rate) and the AKU
// labour-force survey (for the seasonally-adjusted unemployment rate). It
// exists so the Riksdag-page national KPI strip is sourced live rather than
// hardcoded.
package scbmacro

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"riksdagskollen/internal/riksdag/ports"
)

const (
	defaultKPIURL    = "https://api.scb.se/OV0104/v1/doris/sv/ssd/PR/PR0101/PR0101A/KPI2020M"
	defaultAKUURL    = "https://api.scb.se/OV0104/v1/doris/sv/ssd/AM/AM0401/AM0401A/AKURLBefK"
	kpiContentsCode  = "00000804" // Årsförändring — the 12-month inflation rate, %
	akuContentsValue = "ALÖSP"    // arbetslöshetstal, procent
)

type Client struct {
	http   *http.Client
	kpiURL string
	akuURL string
}

func NewClient() *Client {
	return &Client{
		http:   &http.Client{Timeout: 30 * time.Second},
		kpiURL: defaultKPIURL,
		akuURL: defaultAKUURL,
	}
}

// scbResponse is the shared PxWeb json shape.
type scbResponse struct {
	Data []struct {
		Key    []string `json:"key"`
		Values []string `json:"values"`
	} `json:"data"`
}

func (c *Client) post(ctx context.Context, url string, body any) (*scbResponse, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SCB returned %d", resp.StatusCode)
	}
	var out scbResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchInflationRate returns SCB's published 12-month KPI inflation rate
// (Årsförändring) for the two most recent months, latest first. The caller
// uses [0] as the current value and [1] for the month-over-month delta.
func (c *Client) FetchInflationRate(ctx context.Context) ([]ports.MacroPoint, error) {
	body := map[string]any{
		"query": []map[string]any{
			{"code": "ContentsCode", "selection": map[string]any{"filter": "item", "values": []string{kpiContentsCode}}},
			{"code": "Tid", "selection": map[string]any{"filter": "top", "values": []string{"2"}}},
		},
		"response": map[string]any{"format": "json"},
	}
	resp, err := c.post(ctx, c.kpiURL, body)
	if err != nil {
		return nil, fmt.Errorf("fetch KPI årsförändring: %w", err)
	}
	pts, err := parsePoints(resp)
	if err != nil {
		return nil, fmt.Errorf("KPI årsförändring: %w", err)
	}
	return pts, nil
}

// FetchUnemploymentRate returns the seasonally-adjusted unemployment rate
// (ages 15–74, both sexes) for the two most recent quarters, latest first.
func (c *Client) FetchUnemploymentRate(ctx context.Context) ([]ports.MacroPoint, error) {
	body := map[string]any{
		"query": []map[string]any{
			{"code": "Arbetskraftstillh", "selection": map[string]any{"filter": "item", "values": []string{akuContentsValue}}},
			{"code": "Kon", "selection": map[string]any{"filter": "item", "values": []string{"1+2"}}},
			{"code": "Alder", "selection": map[string]any{"filter": "item", "values": []string{"tot15-74"}}},
			{"code": "TypData", "selection": map[string]any{"filter": "item", "values": []string{"SR_DATA"}}},
			{"code": "Tid", "selection": map[string]any{"filter": "top", "values": []string{"2"}}},
		},
		"response": map[string]any{"format": "json"},
	}
	resp, err := c.post(ctx, c.akuURL, body)
	if err != nil {
		return nil, fmt.Errorf("fetch AKU unemployment: %w", err)
	}
	pts, err := parsePoints(resp)
	if err != nil {
		return nil, fmt.Errorf("AKU unemployment: %w", err)
	}
	return pts, nil
}

// parsePoints extracts (period, value) rows, skipping unpublished cells
// (".."), and returns the latest two points, latest first. SCB period labels
// (YYYYMmm / YYYYKq) sort chronologically as strings.
func parsePoints(resp *scbResponse) ([]ports.MacroPoint, error) {
	var pts []ports.MacroPoint
	for _, row := range resp.Data {
		if len(row.Key) == 0 || len(row.Values) == 0 {
			continue
		}
		var v float64
		if _, err := fmt.Sscanf(row.Values[0], "%f", &v); err != nil {
			continue // ".." = not yet published
		}
		pts = append(pts, ports.MacroPoint{Period: row.Key[len(row.Key)-1], Value: round1(v)})
	}
	if len(pts) == 0 {
		return nil, fmt.Errorf("no published data points")
	}
	sort.Slice(pts, func(i, j int) bool { return pts[i].Period < pts[j].Period })
	out := []ports.MacroPoint{pts[len(pts)-1]}
	if len(pts) >= 2 {
		out = append(out, pts[len(pts)-2])
	}
	return out, nil
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5*sign(v))) / 10
}

func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}
