// Package scbpx reads tables from SCB's Statistical Database (Statistikdatabasen).
//
// SCB replaced PxWebApi v1 (api.scb.se/OV0104/v1/doris/...) with PxWebApi v2
// (statistikdatabasen.scb.se/api/v2) in October 2025, and v1 is switched off at
// the turn of 2026/2027. Every call here goes to v2 first and to v1 only as a
// fallback while it still exists.
//
// Since November 2026 SCB's larger tables (e.g. BE0101 BefolkningNy, KLS) fail
// intermittently on both versions: v1 answers 400 Bad Request and v2 500 to the
// same valid query that succeeds on the next try. So requests are retried a few
// times with a doubling backoff (network errors, 429 and 5xx; on v1 also 400)
// before the client falls back to v1. On v2 a 400 is a real answer (a value the
// table does not have, e.g. an unpublished year) and is returned at once,
// without the v1 fallback, which would reject the same query.
//
// Data responses use the v2 "json-px" output format, which has the same
// {"data":[{"key":[...],"values":[...]}]} shape as v1's "json" format, so
// callers parse one row type regardless of which version answered. Keys hold
// one code per selected non-content variable, in table order; ContentsCode
// values are the row's values.
//
// v2 cannot eliminate (sum over) several variables at once on some tables, so
// callers should select every variable explicitly, using the table's total
// value (e.g. Alder "tot") where one exists.
package scbpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	DefaultV2Base = "https://statistikdatabasen.scb.se/api/v2"
	DefaultV1Base = "https://api.scb.se/OV0104/v1/doris/sv/ssd"
	attempts      = 5 // per API version
	metaTTL       = 12 * time.Hour
)

// StatusError is a non-200 answer from SCB.
type StatusError struct{ Status int }

func (e *StatusError) Error() string { return fmt.Sprintf("SCB returned %d", e.Status) }

func isV2BadRequest(err error) bool {
	var se *StatusError
	return errors.As(err, &se) && se.Status == http.StatusBadRequest
}

// Table identifies one SCB table in both API versions.
type Table struct {
	V2ID   string // PxWebApi v2 table id, e.g. "TAB638"
	V1Path string // PxWebApi v1 path below the v1 base, e.g. "BE/BE0101/BE0101A/BefolkningNy"
}

// Known tables. Variable codes and value codes are identical in v1 and v2
// (verified against both metadata endpoints on 2026-11-11).
var (
	BefolkningNy = Table{"TAB638", "BE/BE0101/BE0101A/BefolkningNy"}      // Folkmängd efter region, civilstånd, ålder, kön
	Kfmandat     = Table{"TAB1971", "ME/ME0104/ME0104A/Kfmandat"}         // Kommunfullmäktigval, mandat
	Ltmandat     = Table{"TAB2590", "ME/ME0104/ME0104B/Ltmandat"}         // Regionfullmäktigval, mandat
	KostnDRLT    = Table{"TAB4242", "OE/OE0107/OE0107D/KostnDRLT"}        // Regioners kostnader per verksamhetsområde
	KLSHeadcount = Table{"TAB1978", "AM/AM0102/AM0102A/KLStabell14LpMan"} // KLS statlig sektor per myndighet
)

func (t Table) String() string { return t.V2ID + " (" + t.V1Path + ")" }

// Select picks values for one variable.
type Select struct {
	Code   string
	Values []string
}

// Row is one data row: key codes in table order, then the content values.
type Row struct {
	Key    []string `json:"key"`
	Values []string `json:"values"`
}

// Variable is one table variable with its value codes and labels.
type Variable struct {
	Code       string
	Values     []string
	ValueTexts []string
}

// Meta is a table's variables in table order.
type Meta struct {
	Variables []Variable
}

// Variable returns the variable with the given code, or nil.
func (m *Meta) Variable(code string) *Variable {
	for i := range m.Variables {
		if m.Variables[i].Code == code {
			return &m.Variables[i]
		}
	}
	return nil
}

type Client struct {
	http       *http.Client
	v2Base     string
	v1Base     string        // empty disables the v1 fallback
	retryDelay time.Duration // first backoff; doubles per attempt

	mu        sync.Mutex
	metaCache map[string]cachedMeta
	v2Down    map[string]time.Time // table -> until when to try v1 first
}

// v2Cooldown is how long a table whose v2 data call exhausted its retries is
// read from v1 first (v2 still serves as the fallback), so that a v2 outage
// costs the retry wait once rather than on every request.
const v2Cooldown = 10 * time.Minute

func (c *Client) preferV1(t Table) bool {
	if c.v1Base == "" || t.V1Path == "" {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return time.Now().Before(c.v2Down[t.V2ID])
}

func (c *Client) markV2(t Table, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ok {
		delete(c.v2Down, t.V2ID)
		return
	}
	if c.v2Down == nil {
		c.v2Down = make(map[string]time.Time)
	}
	c.v2Down[t.V2ID] = time.Now().Add(v2Cooldown)
}

type cachedMeta struct {
	meta *Meta
	at   time.Time
}

// New returns a client for the live SCB APIs.
func New(hc *http.Client) *Client {
	return NewWithBaseURLs(DefaultV2Base, DefaultV1Base, hc)
}

// NewWithBaseURLs points the client at other hosts (tests use an httptest
// server). An empty v1Base disables the v1 fallback. A nil hc gets a 30 s
// timeout client.
func NewWithBaseURLs(v2Base, v1Base string, hc *http.Client) *Client {
	if hc == nil {
		hc = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		http:       hc,
		v2Base:     strings.TrimRight(v2Base, "/"),
		v1Base:     strings.TrimRight(v1Base, "/"),
		retryDelay: time.Second,
	}
}

// SetRetryDelay sets the first retry backoff (tests use a tiny one).
func (c *Client) SetRetryDelay(d time.Duration) { c.retryDelay = d }

// Data returns the rows for the selection, from v2 or else v1. After v2 has
// failed for a table, v1 is tried first for v2Cooldown.
func (c *Client) Data(ctx context.Context, t Table, sel []Select) ([]Row, error) {
	if c.preferV1(t) {
		rows, errV1 := c.dataV1(ctx, t, sel)
		if errV1 == nil {
			return rows, nil
		}
		rows, errV2 := c.dataV2(ctx, t, sel)
		if errV2 == nil {
			c.markV2(t, true)
			return rows, nil
		}
		return nil, fmt.Errorf("SCB %s: v1 (v2 cooling down): %v; v2: %w", t.V2ID, errV1, errV2)
	}
	rows, errV2 := c.dataV2(ctx, t, sel)
	if errV2 == nil {
		c.markV2(t, true)
		return rows, nil
	}
	if !isV2BadRequest(errV2) && ctx.Err() == nil {
		c.markV2(t, false)
	}
	if c.v1Base == "" || t.V1Path == "" || isV2BadRequest(errV2) {
		return nil, fmt.Errorf("SCB %s: v2: %w", t.V2ID, errV2)
	}
	rows, errV1 := c.dataV1(ctx, t, sel)
	if errV1 == nil {
		return rows, nil
	}
	return nil, fmt.Errorf("SCB %s: v2: %v; v1 fallback: %w", t.V2ID, errV2, errV1)
}

// Meta returns the table's variables, from v2 or else v1. Results are cached
// for metaTTL per client.
func (c *Client) Meta(ctx context.Context, t Table) (*Meta, error) {
	c.mu.Lock()
	if e, ok := c.metaCache[t.V2ID]; ok && time.Since(e.at) < metaTTL {
		c.mu.Unlock()
		return e.meta, nil
	}
	c.mu.Unlock()
	m, err := c.fetchMeta(ctx, t)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	if c.metaCache == nil {
		c.metaCache = make(map[string]cachedMeta)
	}
	c.metaCache[t.V2ID] = cachedMeta{meta: m, at: time.Now()}
	c.mu.Unlock()
	return m, nil
}

// Has reports which of values the table lists for variable code (e.g. which
// years are published), keeping their order.
func (c *Client) Has(ctx context.Context, t Table, code string, values []string) ([]string, error) {
	m, err := c.Meta(ctx, t)
	if err != nil {
		return nil, err
	}
	v := m.Variable(code)
	if v == nil {
		return nil, fmt.Errorf("SCB %s: no variable %s", t.V2ID, code)
	}
	known := make(map[string]bool, len(v.Values))
	for _, x := range v.Values {
		known[x] = true
	}
	out := make([]string, 0, len(values))
	for _, x := range values {
		if known[x] {
			out = append(out, x)
		}
	}
	return out, nil
}

func (c *Client) fetchMeta(ctx context.Context, t Table) (*Meta, error) {
	m, errV2 := c.metaV2(ctx, t)
	if errV2 == nil {
		return m, nil
	}
	if c.v1Base == "" || t.V1Path == "" {
		return nil, fmt.Errorf("SCB %s metadata: v2: %w", t.V2ID, errV2)
	}
	m, errV1 := c.metaV1(ctx, t)
	if errV1 == nil {
		return m, nil
	}
	return nil, fmt.Errorf("SCB %s metadata: v2: %v; v1 fallback: %w", t.V2ID, errV2, errV1)
}

type dataResponse struct {
	Data []Row `json:"data"`
}

func (c *Client) dataV2(ctx context.Context, t Table, sel []Select) ([]Row, error) {
	type v2sel struct {
		VariableCode string   `json:"variableCode"`
		ValueCodes   []string `json:"valueCodes"`
	}
	body := struct {
		Selection []v2sel `json:"selection"`
	}{}
	for _, s := range sel {
		body.Selection = append(body.Selection, v2sel{s.Code, s.Values})
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	u := c.v2Base + "/tables/" + url.PathEscape(t.V2ID) + "/data?lang=sv&outputFormat=json-px"
	raw, err := c.do(ctx, http.MethodPost, u, b, false)
	if err != nil {
		return nil, err
	}
	var out dataResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode v2 json-px: %w", err)
	}
	return out.Data, nil
}

func (c *Client) dataV1(ctx context.Context, t Table, sel []Select) ([]Row, error) {
	query := make([]map[string]any, 0, len(sel))
	for _, s := range sel {
		query = append(query, map[string]any{"code": s.Code, "selection": map[string]any{"filter": "item", "values": s.Values}})
	}
	b, err := json.Marshal(map[string]any{"query": query, "response": map[string]any{"format": "json"}})
	if err != nil {
		return nil, err
	}
	raw, err := c.do(ctx, http.MethodPost, c.v1Base+"/"+t.V1Path, b, true)
	if err != nil {
		return nil, err
	}
	var out dataResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode v1 json: %w", err)
	}
	return out.Data, nil
}

func (c *Client) metaV2(ctx context.Context, t Table) (*Meta, error) {
	raw, err := c.do(ctx, http.MethodGet, c.v2Base+"/tables/"+url.PathEscape(t.V2ID)+"/metadata?lang=sv", nil, false)
	if err != nil {
		return nil, err
	}
	var ds struct {
		ID        []string `json:"id"`
		Dimension map[string]struct {
			Category struct {
				Index map[string]int    `json:"index"`
				Label map[string]string `json:"label"`
			} `json:"category"`
		} `json:"dimension"`
	}
	if err := json.Unmarshal(raw, &ds); err != nil {
		return nil, fmt.Errorf("decode v2 metadata: %w", err)
	}
	if len(ds.ID) == 0 {
		return nil, fmt.Errorf("v2 metadata: no dimensions")
	}
	m := &Meta{}
	for _, id := range ds.ID {
		cat := ds.Dimension[id].Category
		v := Variable{Code: id, Values: make([]string, len(cat.Index)), ValueTexts: make([]string, len(cat.Index))}
		for code, i := range cat.Index {
			if i < 0 || i >= len(v.Values) {
				return nil, fmt.Errorf("v2 metadata: %s index %d out of range", id, i)
			}
			v.Values[i] = code
			v.ValueTexts[i] = cat.Label[code]
		}
		m.Variables = append(m.Variables, v)
	}
	return m, nil
}

func (c *Client) metaV1(ctx context.Context, t Table) (*Meta, error) {
	raw, err := c.do(ctx, http.MethodGet, c.v1Base+"/"+t.V1Path, nil, true)
	if err != nil {
		return nil, err
	}
	var out struct {
		Variables []struct {
			Code       string   `json:"code"`
			Values     []string `json:"values"`
			ValueTexts []string `json:"valueTexts"`
		} `json:"variables"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode v1 metadata: %w", err)
	}
	m := &Meta{}
	for _, v := range out.Variables {
		m.Variables = append(m.Variables, Variable{Code: v.Code, Values: v.Values, ValueTexts: v.ValueTexts})
	}
	return m, nil
}

// do sends the request up to `attempts` times. It retries network errors, 429
// and 5xx, and 400 when retry400 is set (SCB's flaky v1 answer); other
// statuses are final.
func (c *Client) do(ctx context.Context, method, u string, body []byte, retry400 bool) ([]byte, error) {
	delay := c.retryDelay
	var lastErr error
	tries := 0
	for i := 0; i < attempts; i++ {
		tries++
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("%w (after: %v)", ctx.Err(), lastErr)
			case <-time.After(delay):
			}
			delay *= 2
		}
		var rd io.Reader
		if body != nil {
			rd = bytes.NewReader(body)
		}
		req, err := http.NewRequestWithContext(ctx, method, u, rd)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		out, status, err := c.once(req)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			break
		}
		if status == http.StatusBadRequest && retry400 {
			continue
		}
		if status != 0 && status != http.StatusTooManyRequests && status < 500 {
			break
		}
	}
	return nil, fmt.Errorf("%w (%d attempts)", lastErr, tries)
}

func (c *Client) once(req *http.Request) ([]byte, int, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, &StatusError{Status: resp.StatusCode}
	}
	return b, resp.StatusCode, nil
}
