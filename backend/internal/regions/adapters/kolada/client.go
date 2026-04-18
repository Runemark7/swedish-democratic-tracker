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

const baseURL = "https://api.kolada.se/v3"

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
	yearsStr := strings.Join(yearStrs, ",")

	var result []ports.KPIValue
	for _, kpi := range kpiCodes {
		url := fmt.Sprintf("%s/data/kpi/%s/municipality/%s/year/%s", baseURL, kpi, munCode, yearsStr)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("kolada API returned %d for KPI %s", resp.StatusCode, kpi)
		}

		var payload struct {
			Values []struct {
				KPI          string `json:"kpi"`
				Period       int    `json:"period"`
				Municipality string `json:"municipality"`
				Values       []struct {
					Gender    string  `json:"gender"`
					Value     float64 `json:"value"`
					Status    string  `json:"status"`
					IsDeleted bool    `json:"isdeleted"`
				} `json:"values"`
			} `json:"values"`
		}
		err = json.NewDecoder(resp.Body).Decode(&payload)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decode kolada response for %s: %w", kpi, err)
		}

		for _, v := range payload.Values {
			for _, val := range v.Values {
				if val.Gender == "T" && !val.IsDeleted {
					result = append(result, ports.KPIValue{
						KPI:    v.KPI,
						Year:   v.Period,
						Value:  val.Value,
						Status: val.Status,
					})
					break
				}
			}
		}
	}
	return result, nil
}
