// Package scbmacro fetches national macro indicators from SCB's Statistical
// Database: the consumer price index (for the 12-month inflation rate) and the
// AKU labour-force survey (for the seasonally-adjusted unemployment rate). It
// exists so the Riksdag-page national KPI strip is sourced live rather than
// hardcoded.
//
// SCB replaced PxWebApi v1 (api.scb.se/OV0104/v1/doris/...) with PxWebApi v2
// (statistikdatabasen.scb.se/api/v2) in October 2025; v1 is switched off at the
// turn of 2026/2027. Each series is therefore read from v2 first (table ids
// TAB6516 for AKU per quarter, TAB6596 for KPI) and from v1 only as a fallback
// while it still exists.
//
// Since November 2026 SCB's AKU tables fail intermittently on both versions
// (v1 answers 400 Bad Request, v2 500, to the same valid query that succeeds
// on the next try), so every request is retried a few times with a short
// backoff before falling back. The query selects every variable explicitly,
// including ContentsCode, so it never depends on SCB's elimination defaults.
package scbmacro

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"riksdagskollen/internal/riksdag/ports"
)

const (
	defaultV2Base = "https://statistikdatabasen.scb.se/api/v2"
	defaultV1Base = "https://api.scb.se/OV0104/v1/doris/sv/ssd"
	timeVar       = "Tid"
	topN          = "2"
	attempts      = 4 // per API version
)

// selection is one non-time variable and the single value we want for it.
type selection struct{ code, value string }

// series describes one SCB time series in both API versions.
type series struct {
	name     string      // for error messages
	v2Table  string      // PxWebApi v2 table id
	v1Path   string      // PxWebApi v1 path below the v1 base
	selected []selection // every non-time variable, one value each
}

var (
	// AKU: arbetslöshetstal (ALÖSP), säsongrensad, båda könen, 15–74 år, per kvartal.
	akuSeries = series{
		name:    "AKU unemployment",
		v2Table: "TAB6516",
		v1Path:  "AM/AM0401/AM0401A/AKURLBefK",
		selected: []selection{
			{"Arbetskraftstillh", "ALÖSP"},
			{"TypData", "SR_DATA"},
			{"Kon", "1+2"},
			{"Alder", "tot15-74"},
			{"ContentsCode", "000007V3"},
		},
	}
	// KPI 2020=100: Årsförändring (00000804) = 12-month inflation rate, %.
	kpiSeries = series{
		name:     "KPI årsförändring",
		v2Table:  "TAB6596",
		v1Path:   "PR/PR0101/PR0101A/KPI2020M",
		selected: []selection{{"ContentsCode", "00000804"}},
	}
)

type Client struct {
	http       *http.Client
	v2Base     string
	v1Base     string        // empty disables the v1 fallback
	retryDelay time.Duration // first backoff; doubles per attempt
}

func NewClient() *Client {
	return NewClientWithBaseURLs(defaultV2Base, defaultV1Base)
}

// NewClientWithBaseURLs points the client at other PxWebApi hosts (tests use
// an httptest server). An empty v1Base disables the v1 fallback.
func NewClientWithBaseURLs(v2Base, v1Base string) *Client {
	return &Client{
		http:       &http.Client{Timeout: 30 * time.Second},
		v2Base:     strings.TrimRight(v2Base, "/"),
		v1Base:     strings.TrimRight(v1Base, "/"),
		retryDelay: time.Second,
	}
}

// FetchInflationRate returns SCB's published 12-month KPI inflation rate
// (Årsförändring) for the two most recent months, latest first. The caller
// uses [0] as the current value and [1] for the month-over-month delta.
func (c *Client) FetchInflationRate(ctx context.Context) ([]ports.MacroPoint, error) {
	return c.fetch(ctx, kpiSeries)
}

// FetchUnemploymentRate returns the seasonally-adjusted unemployment rate
// (ages 15–74, both sexes) for the two most recent quarters, latest first.
func (c *Client) FetchUnemploymentRate(ctx context.Context) ([]ports.MacroPoint, error) {
	return c.fetch(ctx, akuSeries)
}

func (c *Client) fetch(ctx context.Context, s series) ([]ports.MacroPoint, error) {
	pts, errV2 := c.fetchV2(ctx, s)
	if errV2 == nil {
		return pts, nil
	}
	if c.v1Base == "" {
		return nil, fmt.Errorf("fetch %s: v2: %w", s.name, errV2)
	}
	pts, errV1 := c.fetchV1(ctx, s)
	if errV1 == nil {
		return pts, nil
	}
	return nil, fmt.Errorf("fetch %s: v2: %v; v1 fallback: %w", s.name, errV2, errV1)
}

// ---- PxWebApi v2 (GET, json-stat2) ----

func (c *Client) fetchV2(ctx context.Context, s series) ([]ports.MacroPoint, error) {
	q := url.Values{}
	q.Set("lang", "sv")
	q.Set("outputFormat", "json-stat2")
	for _, sel := range s.selected {
		q.Set("valueCodes["+sel.code+"]", sel.value)
	}
	q.Set("valueCodes["+timeVar+"]", "top("+topN+")")
	u := c.v2Base + "/tables/" + url.PathEscape(s.v2Table) + "/data?" + q.Encode()
	body, err := c.do(ctx, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err == nil {
			req.Header.Set("Accept", "application/json")
		}
		return req, err
	})
	if err != nil {
		return nil, err
	}
	return parseJSONStat2(body)
}

// jsonStat2 is the part of a json-stat2 dataset we read.
type jsonStat2 struct {
	ID        []string                   `json:"id"`
	Size      []int                      `json:"size"`
	Value     []*float64                 `json:"value"`
	Dimension map[string]jsonStat2Dim    `json:"dimension"`
	Role      map[string][]string        `json:"role"`
	Status    map[string]json.RawMessage `json:"status"`
}

type jsonStat2Dim struct {
	Category struct {
		Index map[string]int `json:"index"`
	} `json:"category"`
}

// parseJSONStat2 reads a dataset in which every dimension except time has
// exactly one category, so the value array is indexed by time alone.
func parseJSONStat2(body []byte) ([]ports.MacroPoint, error) {
	var ds jsonStat2
	if err := json.Unmarshal(body, &ds); err != nil {
		return nil, fmt.Errorf("decode json-stat2: %w", err)
	}
	if len(ds.ID) == 0 || len(ds.ID) != len(ds.Size) {
		return nil, errors.New("json-stat2: missing id/size")
	}
	tdim := timeVar
	if t := ds.Role["time"]; len(t) == 1 {
		tdim = t[0]
	}
	cells := 1
	found := false
	for i, id := range ds.ID {
		cells *= ds.Size[i]
		if id == tdim {
			found = true
		} else if ds.Size[i] != 1 {
			return nil, fmt.Errorf("json-stat2: dimension %s has %d categories, want 1", id, ds.Size[i])
		}
	}
	if !found {
		return nil, fmt.Errorf("json-stat2: no time dimension %s", tdim)
	}
	if len(ds.Value) != cells {
		return nil, fmt.Errorf("json-stat2: %d values for %d cells", len(ds.Value), cells)
	}
	var pts []ports.MacroPoint
	for period, i := range ds.Dimension[tdim].Category.Index {
		if i < 0 || i >= len(ds.Value) || ds.Value[i] == nil {
			continue // null = not yet published / confidential
		}
		pts = append(pts, ports.MacroPoint{Period: period, Value: round1(*ds.Value[i])})
	}
	return latestTwo(pts)
}

// ---- PxWebApi v1 (POST, json) — fallback until SCB removes it ----

// scbResponse is the PxWeb v1 json shape.
type scbResponse struct {
	Data []struct {
		Key    []string `json:"key"`
		Values []string `json:"values"`
	} `json:"data"`
}

func (c *Client) fetchV1(ctx context.Context, s series) ([]ports.MacroPoint, error) {
	query := make([]map[string]any, 0, len(s.selected)+1)
	for _, sel := range s.selected {
		query = append(query, map[string]any{"code": sel.code, "selection": map[string]any{"filter": "item", "values": []string{sel.value}}})
	}
	query = append(query, map[string]any{"code": timeVar, "selection": map[string]any{"filter": "top", "values": []string{topN}}})
	b, err := json.Marshal(map[string]any{"query": query, "response": map[string]any{"format": "json"}})
	if err != nil {
		return nil, err
	}
	body, err := c.do(ctx, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.v1Base+"/"+s.v1Path, bytes.NewReader(b))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
		return req, err
	})
	if err != nil {
		return nil, err
	}
	var out scbResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode v1 json: %w", err)
	}
	return parsePoints(&out)
}

// do sends the request up to `attempts` times. It retries network errors,
// 5xx and 400 (SCB's flaky AKU answer), and 429; other statuses are final.
func (c *Client) do(ctx context.Context, newReq func() (*http.Request, error)) ([]byte, error) {
	delay := c.retryDelay
	var lastErr error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("%w (after: %v)", ctx.Err(), lastErr)
			case <-time.After(delay):
			}
			delay *= 2
		}
		req, err := newReq()
		if err != nil {
			return nil, err
		}
		body, status, err := c.once(req)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			break
		}
		if status != 0 && status != http.StatusBadRequest && status != http.StatusTooManyRequests && status < 500 {
			break
		}
	}
	return nil, lastErr
}

func (c *Client) once(req *http.Request) ([]byte, int, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("SCB returned %d", resp.StatusCode)
	}
	return body, resp.StatusCode, nil
}

// parsePoints extracts (period, value) rows from a v1 response, skipping
// unpublished cells (".."). The time code is the last key element because
// every other key variable precedes it and ContentsCode is a value column.
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
	return latestTwo(pts)
}

// latestTwo returns the latest two points, latest first. SCB period labels
// (YYYYMmm / YYYYKq) sort chronologically as strings.
func latestTwo(pts []ports.MacroPoint) ([]ports.MacroPoint, error) {
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
