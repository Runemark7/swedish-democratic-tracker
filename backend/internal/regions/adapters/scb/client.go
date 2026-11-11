package scb

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"riksdagskollen/internal/regions/ports"
	"riksdagskollen/internal/scbpx"
)

// Client reads SCB's KostnDRLT (region budgets) and BefolkningNy tables via
// PxWebApi v2, falling back to v1 while it exists (see package scbpx).
type Client struct {
	px *scbpx.Client

	mu    sync.Mutex
	trend map[string]cachedTrend // munCode|years -> result
}

// BE0101 is published once a year, so a municipality's trend is cached.
const trendTTL = 12 * time.Hour

type cachedTrend struct {
	entries []ports.PopulationEntry
	at      time.Time
}

func NewClient() *Client {
	return &Client{px: scbpx.New(&http.Client{Timeout: 15 * time.Second})}
}

// NewClientWithHTTP creates a Client for testing: baseURL stands in for the
// PxWebApi v2 base and the v1 fallback is disabled.
func NewClientWithHTTP(baseURL string, httpClient *http.Client) *Client {
	px := scbpx.NewWithBaseURLs(baseURL, "", httpClient)
	px.SetRetryDelay(time.Millisecond)
	return &Client{px: px}
}

type scbRow = scbpx.Row

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
	rows, err := c.px.Data(ctx, scbpx.KostnDRLT, []scbpx.Select{
		{Code: "Region", Values: []string{scbCode}},
		{Code: "Verksomrkom", Values: areaCodes},
		{Code: "ContentsCode", Values: []string{"000000A7"}},
		{Code: "Tid", Values: []string{yearStr}},
	})
	if err != nil {
		return nil, fmt.Errorf("SCB region budget: %w", err)
	}
	payload := struct{ Data []scbRow }{Data: rows}

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

	rows, err := c.px.Data(ctx, scbpx.KostnDRLT, []scbpx.Select{
		{Code: "Region", Values: []string{code}},
		{Code: "Verksomrkom", Values: areaCodes},
		{Code: "ContentsCode", Values: []string{"000000A7"}},
		{Code: "Tid", Values: yearStrs},
	})
	if err != nil {
		return nil, fmt.Errorf("SCB region budget: %w", err)
	}
	payload := struct{ Data []scbRow }{Data: rows}

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

	rows, err := c.px.Data(ctx, scbpx.KostnDRLT, []scbpx.Select{
		{Code: "Region", Values: scbCodes},
		{Code: "Verksomrkom", Values: []string{areaCode, "0-9"}},
		{Code: "ContentsCode", Values: []string{"000000A7"}},
		{Code: "Tid", Values: []string{strconv.Itoa(year)}},
	})
	if err != nil {
		return nil, fmt.Errorf("SCB area-across-regions: %w", err)
	}
	payload := struct{ Data []scbRow }{Data: rows}

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
	key := fmt.Sprint(munCode, years)
	c.mu.Lock()
	if e, ok := c.trend[key]; ok && time.Since(e.at) < trendTTL {
		c.mu.Unlock()
		return e.entries, nil
	}
	c.mu.Unlock()
	entries, err := c.fetchPopulationTrend(ctx, munCode, years)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	if c.trend == nil {
		c.trend = make(map[string]cachedTrend)
	}
	c.trend[key] = cachedTrend{entries: entries, at: time.Now()}
	c.mu.Unlock()
	return entries, nil
}

func (c *Client) fetchPopulationTrend(ctx context.Context, munCode string, years []int) ([]ports.PopulationEntry, error) {
	yearStrs := make([]string, len(years))
	for i, y := range years {
		yearStrs[i] = strconv.Itoa(y)
	}
	// Ask only for published years: SCB rejects the whole query (400) if one
	// year is missing, and BE0101 lags the calendar by more than a year.
	yearStrs, err := c.px.Has(ctx, scbpx.BefolkningNy, "Tid", yearStrs)
	if err != nil {
		return nil, fmt.Errorf("SCB population trend: %w", err)
	}
	if len(yearStrs) == 0 {
		return []ports.PopulationEntry{}, nil
	}

	rows, err := c.px.Data(ctx, scbpx.BefolkningNy, []scbpx.Select{
		{Code: "Region", Values: []string{munCode}},
		{Code: "Alder", Values: []string{"tot"}},
		{Code: "ContentsCode", Values: []string{"BE0101N1"}},
		{Code: "Tid", Values: yearStrs},
	})
	if err != nil {
		return nil, fmt.Errorf("SCB population trend: %w", err)
	}
	payload := struct{ Data []scbRow }{Data: rows}

	result := []ports.PopulationEntry{}
	for _, d := range payload.Data {
		// Key layout is [region, alder, year] (Alder pinned to "tot").
		// ContentsCode is not in the key; it is the (only) entry in values[].
		if len(d.Key) < 2 || len(d.Values) == 0 {
			continue
		}
		year, err := strconv.Atoi(d.Key[len(d.Key)-1])
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
