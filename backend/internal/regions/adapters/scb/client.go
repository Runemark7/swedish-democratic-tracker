package scb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"riksdagskollen/internal/regions/ports"
)

const (
	populationURL   = "https://api.scb.se/OV0104/v1/doris/sv/ssd/BE/BE0101/BE0101A/BefolkningNy"
	regionBudgetURL = "https://api.scb.se/OV0104/v1/doris/sv/ssd/OE/OE0107/OE0107D/KostnDRLT"
)

type Client struct {
	http           *http.Client
	regionBudgetBase string // overridable for tests; empty means use regionBudgetURL
	populationBase   string // overridable for tests; empty means use populationURL
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 15 * time.Second}}
}

// NewClientWithHTTP creates a Client for testing, routing all requests to baseURL.
func NewClientWithHTTP(baseURL string, httpClient *http.Client) *Client {
	return &Client{
		http:             httpClient,
		regionBudgetBase: baseURL,
		populationBase:   baseURL,
	}
}

func (c *Client) regionBudgetURL() string {
	if c.regionBudgetBase != "" {
		return c.regionBudgetBase
	}
	return regionBudgetURL
}

func (c *Client) populationTrendURL() string {
	if c.populationBase != "" {
		return c.populationBase
	}
	return populationURL
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
	// SCB uses e.g. "06L" for Jönköping. Gotland is special: it's both
	// kommun and region, so SCB uses "0980L" (kommun-style) in the region
	// budget table.
	scbCode := scbRegionCode(regionCode)

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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.regionBudgetURL(), bytes.NewReader(bodyBytes))
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
			Pct:   math.Round((val/total)*1000) / 10,
		})
	}
	return result, nil
}

// scbRegionCode converts a 2-digit county code to the SCB "NNL" / "0980L" format used
// in the region budget table (OE0107D/KostnDRLT).
func scbRegionCode(regionCode string) string {
	if regionCode == "09" {
		return "0980L"
	}
	return regionCode + "L"
}

// scbToRegionCode is the inverse of scbRegionCode.
func scbToRegionCode(scbCode string) string {
	if scbCode == "0980L" {
		return "09"
	}
	if len(scbCode) > 1 && scbCode[len(scbCode)-1] == 'L' {
		return scbCode[:len(scbCode)-1]
	}
	return scbCode
}

// FetchRegionBudgetMultiYear fetches budget data for a single region across multiple years.
// It returns one RegionBudgetSnapshot per (area, year) combination.
func (c *Client) FetchRegionBudgetMultiYear(ctx context.Context, regionCode string, years []int) ([]ports.RegionBudgetSnapshot, error) {
	code := scbRegionCode(regionCode)

	areaCodes := make([]string, len(regionBudgetAreas)+1)
	for i, a := range regionBudgetAreas {
		areaCodes[i] = a.code
	}
	areaCodes[len(regionBudgetAreas)] = "0-9"

	yearStrs := make([]string, len(years))
	for i, y := range years {
		yearStrs[i] = strconv.Itoa(y)
	}

	body := map[string]any{
		"query": []map[string]any{
			{"code": "Region", "selection": map[string]any{"filter": "item", "values": []string{code}}},
			{"code": "Verksomrkom", "selection": map[string]any{"filter": "item", "values": areaCodes}},
			{"code": "ContentsCode", "selection": map[string]any{"filter": "item", "values": []string{"000000A7"}}},
			{"code": "Tid", "selection": map[string]any{"filter": "item", "values": yearStrs}},
		},
		"response": map[string]any{"format": "json"},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.regionBudgetURL(), bytes.NewReader(bodyBytes))
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
		return nil, fmt.Errorf("decode SCB region budget multi-year: %w", err)
	}

	// Group by year: year → areaCode → value
	type yearData struct {
		byCode map[string]float64
	}
	perYear := map[int]*yearData{}

	for _, d := range payload.Data {
		// key layout: [region, verksamhetsområde, år]
		if len(d.Key) < 3 || len(d.Values) == 0 || d.Values[0] == ".." {
			continue
		}
		year, err := strconv.Atoi(d.Key[2])
		if err != nil {
			continue
		}
		val, err := strconv.ParseFloat(d.Values[0], 64)
		if err != nil {
			continue
		}
		if perYear[year] == nil {
			perYear[year] = &yearData{byCode: map[string]float64{}}
		}
		perYear[year].byCode[d.Key[1]] = val
	}

	var result []ports.RegionBudgetSnapshot
	for _, yearStr := range yearStrs {
		year, _ := strconv.Atoi(yearStr)
		yd, ok := perYear[year]
		if !ok {
			continue
		}
		total := yd.byCode["0-9"]
		if total == 0 {
			continue
		}
		for _, a := range regionBudgetAreas {
			val, ok := yd.byCode[a.code]
			if !ok {
				continue
			}
			result = append(result, ports.RegionBudgetSnapshot{
				RegionCode: regionCode,
				AreaName:   a.label,
				Year:       year,
				ValueMnkr:  val,
				TotalMnkr:  total,
				Pct:        math.Round((val/total)*1000) / 10,
			})
		}
	}
	return result, nil
}

// FetchAreaAcrossRegions fetches a single budget area across multiple regions for one year.
func (c *Client) FetchAreaAcrossRegions(ctx context.Context, areaName string, regionCodes []string, year int) ([]ports.RegionAreaDataPoint, error) {
	// Look up area code by label.
	areaCode := ""
	for _, a := range regionBudgetAreas {
		if a.label == areaName {
			areaCode = a.code
			break
		}
	}
	if areaCode == "" {
		return nil, fmt.Errorf("unknown area name: %s", areaName)
	}

	// Convert region codes to SCB codes and keep a reverse map.
	scbCodes := make([]string, len(regionCodes))
	reverseMap := map[string]string{} // scbCode → regionCode
	for i, rc := range regionCodes {
		sc := scbRegionCode(rc)
		scbCodes[i] = sc
		reverseMap[sc] = rc
	}

	body := map[string]any{
		"query": []map[string]any{
			{"code": "Region", "selection": map[string]any{"filter": "item", "values": scbCodes}},
			{"code": "Verksomrkom", "selection": map[string]any{"filter": "item", "values": []string{areaCode, "0-9"}}},
			{"code": "ContentsCode", "selection": map[string]any{"filter": "item", "values": []string{"000000A7"}}},
			{"code": "Tid", "selection": map[string]any{"filter": "item", "values": []string{strconv.Itoa(year)}}},
		},
		"response": map[string]any{"format": "json"},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.regionBudgetURL(), bytes.NewReader(bodyBytes))
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
		return nil, fmt.Errorf("SCB area-across-regions API returned %d", resp.StatusCode)
	}

	var payload struct {
		Data []struct {
			Key    []string `json:"key"`
			Values []string `json:"values"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode SCB area-across-regions: %w", err)
	}

	// Build per-region data: scbCode → {areaValue, totalValue}
	type regionVals struct {
		area  float64
		total float64
	}
	byRegion := map[string]*regionVals{}

	for _, d := range payload.Data {
		// key layout: [region, verksamhetsområde, år]
		if len(d.Key) < 3 || len(d.Values) == 0 || d.Values[0] == ".." {
			continue
		}
		scbCode := d.Key[0]
		vType := d.Key[1]
		val, err := strconv.ParseFloat(d.Values[0], 64)
		if err != nil {
			continue
		}
		if byRegion[scbCode] == nil {
			byRegion[scbCode] = &regionVals{}
		}
		switch vType {
		case areaCode:
			byRegion[scbCode].area = val
		case "0-9":
			byRegion[scbCode].total = val
		}
	}

	var result []ports.RegionAreaDataPoint
	for scbCode, rv := range byRegion {
		rc, ok := reverseMap[scbCode]
		if !ok {
			rc = scbToRegionCode(scbCode)
		}
		pct := 0.0
		if rv.total > 0 {
			pct = math.Round((rv.area/rv.total)*1000) / 10
		}
		result = append(result, ports.RegionAreaDataPoint{
			RegionCode: rc,
			ValueMnkr:  rv.area,
			TotalMnkr:  rv.total,
			Pct:        pct,
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.populationTrendURL(), bytes.NewReader(bodyBytes))
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
