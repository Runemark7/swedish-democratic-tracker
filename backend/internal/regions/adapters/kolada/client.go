package kolada

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"riksdagskollen/internal/regions/ports"
)

const defaultBaseURL = "https://api.kolada.se/v3"

type Client struct {
	http    *http.Client
	baseURL string // overridable for tests
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 10 * time.Second}, baseURL: defaultBaseURL}
}

func (c *Client) FetchKPIs(ctx context.Context, munCode string, kpiCodes []string, years []int) ([]ports.KPIValue, error) {
	yearStrs := make([]string, len(years))
	for i, y := range years {
		yearStrs[i] = strconv.Itoa(y)
	}
	yearsStr := strings.Join(yearStrs, ",")

	result := make([]ports.KPIValue, 0, len(kpiCodes)*len(years))
	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error

	for _, kpi := range kpiCodes {
		wg.Add(1)
		go func(kpi string) {
			defer wg.Done()
			vals, err := c.fetchOneKPI(ctx, kpi, munCode, yearsStr)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			result = append(result, vals...)
		}(kpi)
	}
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	return result, nil
}

func (c *Client) fetchOneKPI(ctx context.Context, kpi, munCode, yearsStr string) ([]ports.KPIValue, error) {
	url := fmt.Sprintf("%s/data/kpi/%s/municipality/%s/year/%s", c.baseURL, kpi, munCode, yearsStr)

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
			KPI    string `json:"kpi"`
			Period int    `json:"period"`
			Values []struct {
				Gender    string   `json:"gender"`
				Value     *float64 `json:"value"`
				Status    string   `json:"status"`
				IsDeleted bool     `json:"isdeleted"`
			} `json:"values"`
		} `json:"values"`
	}
	err = json.NewDecoder(resp.Body).Decode(&payload)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("decode kolada response for %s: %w", kpi, err)
	}

	var result []ports.KPIValue
	for _, v := range payload.Values {
		for _, val := range v.Values {
			// A null value means the entity has not reported for that period.
			// Treat it as no data (skip) rather than letting JSON null decode
			// to 0.0 and surface as a real zero.
			if val.Gender == "T" && !val.IsDeleted && val.Value != nil {
				result = append(result, ports.KPIValue{
					KPI:    v.KPI,
					Year:   v.Period,
					Value:  *val.Value,
					Status: val.Status,
				})
				break
			}
		}
	}
	return result, nil
}

func (c *Client) FetchKPIAllMunicipalities(ctx context.Context, kpiCode string, years []int) ([]ports.KPIValueWithMun, error) {
	yearStrs := make([]string, len(years))
	for i, y := range years {
		yearStrs[i] = strconv.Itoa(y)
	}
	yearsStr := strings.Join(yearStrs, ",")

	url := fmt.Sprintf("%s/data/kpi/%s/year/%s", c.baseURL, kpiCode, yearsStr)

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
		return nil, fmt.Errorf("kolada API returned %d for KPI %s/all", resp.StatusCode, kpiCode)
	}

	var payload struct {
		Values []struct {
			KPI          string `json:"kpi"`
			Municipality string `json:"municipality"`
			Period       int    `json:"period"`
			Values       []struct {
				Gender    string   `json:"gender"`
				Value     *float64 `json:"value"`
				Status    string   `json:"status"`
				IsDeleted bool     `json:"isdeleted"`
			} `json:"values"`
		} `json:"values"`
	}
	err = json.NewDecoder(resp.Body).Decode(&payload)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("decode kolada all-municipalities response for %s: %w", kpiCode, err)
	}

	var result []ports.KPIValueWithMun
	for _, v := range payload.Values {
		if v.Municipality == "0000" {
			continue // skip national aggregate
		}
		for _, val := range v.Values {
			// Skip null values: a missing reading is "no data", not 0.0, and
			// must not surface as a real zero (e.g. a region falsely ranked
			// last for the latest year before it has reported).
			if val.Gender == "T" && !val.IsDeleted && val.Value != nil {
				result = append(result, ports.KPIValueWithMun{
					MunCode: v.Municipality,
					KPI:     v.KPI,
					Year:    v.Period,
					Value:   *val.Value,
					Status:  val.Status,
				})
				break
			}
		}
	}
	return result, nil
}
