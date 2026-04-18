package kolada

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"riksdagskollen/internal/regions/ports"
)

const baseURL = "https://api.kolada.se/v2"

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) FetchKPIs(ctx context.Context, munCode string, kpiCodes []string, years []int) ([]ports.KPIValue, error) {
	yearStrs := make([]string, len(years))
	for i, y := range years {
		yearStrs[i] = strconv.Itoa(y)
	}
	url := fmt.Sprintf("%s/data/kpi/%s/municipality/%s/year/%s",
		baseURL,
		strings.Join(kpiCodes, ","),
		munCode,
		strings.Join(yearStrs, ","),
	)

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
		return nil, fmt.Errorf("kolada API returned %d", resp.StatusCode)
	}

	var payload struct {
		Values []struct {
			KPI            string `json:"kpi"`
			MunicipalityID string `json:"municipality_id"`
			Period         string `json:"period"`
			Values         []struct {
				Gender string  `json:"gender"`
				Count  float64 `json:"count"`
				Status string  `json:"status"`
			} `json:"values"`
		} `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode kolada response: %w", err)
	}

	result := []ports.KPIValue{}
	for _, v := range payload.Values {
		year, err := strconv.Atoi(v.Period)
		if err != nil {
			continue
		}
		for _, val := range v.Values {
			if val.Gender == "T" {
				result = append(result, ports.KPIValue{
					KPI:    v.KPI,
					Year:   year,
					Value:  val.Count,
					Status: val.Status,
				})
				break
			}
		}
	}
	return result, nil
}
