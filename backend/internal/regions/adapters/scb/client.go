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

const populationURL = "https://api.scb.se/OV0104/v1/doris/sv/ssd/BE/BE0101/BE0101A/BefolkningNy"

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 15 * time.Second}}
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
		// key layout: [region, contentsCode, year]
		if len(d.Key) < 3 || len(d.Values) == 0 {
			continue
		}
		year, err := strconv.Atoi(d.Key[2])
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
