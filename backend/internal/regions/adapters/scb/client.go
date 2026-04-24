package scb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"riksdagskollen/internal/regions/ports"
)

const (
	populationURL  = "https://api.scb.se/OV0104/v1/doris/sv/ssd/BE/BE0101/BE0101A/BefolkningNy"
	regionBudgetURL = "https://api.scb.se/OV0104/v1/doris/sv/ssd/OE/OE0107/OE0107D/KostnDRLT"
)

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 15 * time.Second}}
}

// regionBudgetAreas maps SCB verksamhetsområde codes to Swedish labels.
// We fetch individual subcategories and compute percentages against total (0-9).
var regionBudgetAreas = []struct {
	code  string
	label string
}{
	{"0", "Primärvård"},
	{"1", "Specialiserad somatisk vård"},
	{"2", "Specialiserad psykiatrisk vård"},
	{"3", "Tandvård"},
	{"7", "Trafik & infrastruktur"},
	{"8", "Regional utveckling"},
	{"4", "Övrig hälso- och sjukvård"},
}

func (c *Client) FetchRegionBudget(ctx context.Context, regionCode string, year int) ([]ports.RegionBudgetArea, error) {
	scbCode := regionCode + "L" // SCB uses e.g. "06L" for Jönköping

	areaCodes := make([]string, len(regionBudgetAreas)+1)
	for i, a := range regionBudgetAreas {
		areaCodes[i] = a.code
	}
	areaCodes[len(regionBudgetAreas)] = "0-9" // total

	yearStr := strconv.Itoa(year)
	body := map[string]any{
		"query": []map[string]any{
			{"code": "Region", "selection": map[string]any{"filter": "item", "values": []string{scbCode}}},
			{"code": "Verksomrkom", "selection": map[string]any{"filter": "item", "values": areaCodes}},
			{"code": "ContentsCode", "selection": map[string]any{"filter": "item", "values": []string{"000000A7"}}},
			{"code": "Tid", "selection": map[string]any{"filter": "item", "values": []string{yearStr}}},
		},
		"response": map[string]any{"format": "json"},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, regionBudgetURL, bytes.NewReader(bodyBytes))
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
		return nil, fmt.Errorf("SCB region budget API returned %d", resp.StatusCode)
	}

	var payload struct {
		Data []struct {
			Key    []string `json:"key"`
			Values []string `json:"values"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode SCB region budget: %w", err)
	}

	// Build a map from area code to mnkr value.
	byCode := map[string]float64{}
	for _, d := range payload.Data {
		// key layout: [region, verksamhetsområde, år]
		if len(d.Key) < 3 || len(d.Values) == 0 || d.Values[0] == ".." {
			continue
		}
		val, err := strconv.ParseFloat(d.Values[0], 64)
		if err != nil {
			continue
		}
		byCode[d.Key[1]] = val
	}

	total := byCode["0-9"]
	if total == 0 {
		return []ports.RegionBudgetArea{}, nil
	}

	result := make([]ports.RegionBudgetArea, 0, len(regionBudgetAreas))
	for _, a := range regionBudgetAreas {
		val, ok := byCode[a.code]
		if !ok {
			continue
		}
		result = append(result, ports.RegionBudgetArea{
			Name:  a.label,
			Value: val,
			Pct:   (val / total) * 100,
		})
	}
	return result, nil
}

func (c *Client) FetchPopulationTrend(ctx context.Context, munCode string, years []int) ([]ports.PopulationEntry, error) {
	yearStrs := make([]string, len(years))
	for i, y := range years {
		yearStrs[i] = strconv.Itoa(y)
	}

	body := map[string]any{
		"query": []map[string]any{
			{
				"code":      "Region",
				"selection": map[string]any{"filter": "item", "values": []string{munCode}},
			},
			{
				"code":      "ContentsCode",
				"selection": map[string]any{"filter": "item", "values": []string{"BE0101N1"}},
			},
			{
				"code":      "Tid",
				"selection": map[string]any{"filter": "item", "values": yearStrs},
			},
		},
		"response": map[string]any{"format": "json"},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, populationURL, bytes.NewReader(bodyBytes))
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
		return nil, fmt.Errorf("SCB API returned %d", resp.StatusCode)
	}

	var payload struct {
		Data []struct {
			Key    []string `json:"key"`
			Values []string `json:"values"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode SCB response: %w", err)
	}

	result := []ports.PopulationEntry{}
	for _, d := range payload.Data {
		// SCB key layout is [region, year] — verified against live API 2026-04-19.
		// ContentsCode is declared via the query but does not appear in the key;
		// it is returned as the (only) entry in values[].
		if len(d.Key) < 2 || len(d.Values) == 0 {
			continue
		}
		year, err := strconv.Atoi(d.Key[1])
		if err != nil {
			continue
		}
		pop, err := strconv.Atoi(d.Values[0])
		if err != nil {
			continue
		}
		result = append(result, ports.PopulationEntry{Year: year, Population: pop})
	}
	return result, nil
}
