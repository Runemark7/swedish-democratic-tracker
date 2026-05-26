package myndighetsforteckning

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xuri/excelize/v2"

	"riksdagskollen/internal/riksdag/ports"
)

const datasetURL = "https://www.statskontoret.se/contentassets/bbd19bc969054c86bbc1aa757e7d5c85/statskontorets-myndighetsforteckning-2025.xlsx"

// Sheet that contains the longitudinal "one row per agency-year" data.
const sheetName = "Förteckning 2007-2025"

var (
	digitsRe    = regexp.MustCompile(`^\d{10}$`)
	canonicalRe = regexp.MustCompile(`^\d{6}-\d{4}$`)
)

func normalizeOrgNumber(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if canonicalRe.MatchString(s) {
		return s
	}
	if digitsRe.MatchString(s) {
		return s[:6] + "-" + s[6:]
	}
	return ""
}

type Client struct {
	http     *http.Client
	mu       sync.Mutex
	cached   []ports.HeadcountEntry
	cachedAt time.Time
	cacheTTL time.Duration
}

func NewClient() *Client {
	return &Client{
		http:     &http.Client{Timeout: 60 * time.Second},
		cacheTTL: 24 * time.Hour,
	}
}

func (c *Client) FetchHeadcounts(ctx context.Context) ([]ports.HeadcountEntry, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.cached) > 0 && time.Since(c.cachedAt) < c.cacheTTL {
		return c.cached, nil
	}
	raw, err := c.download(ctx)
	if err != nil {
		return nil, err
	}
	entries, err := parseWorkbook(raw)
	if err != nil {
		return nil, err
	}
	c.cached = entries
	c.cachedAt = time.Now()
	return entries, nil
}

func (c *Client) download(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, datasetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "riksdagskollen/1.0")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("myndighetsforteckning download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("myndighetsforteckning download: HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func parseWorkbook(raw []byte) ([]ports.HeadcountEntry, error) {
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("read sheet %q: %w", sheetName, err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("sheet %q has %d rows, expected >1", sheetName, len(rows))
	}

	header := rows[0]
	idx := func(name string) int {
		for i, h := range header {
			if strings.EqualFold(strings.TrimSpace(h), name) {
				return i
			}
		}
		return -1
	}
	iOrg := idx("orgnr")
	iName := idx("myndighet")
	iDept := idx("departement")
	iYear := idx("år")
	iFTE := idx("årsarbetskrafter")
	if iOrg < 0 || iName < 0 || iYear < 0 || iFTE < 0 {
		return nil, fmt.Errorf("missing required columns; header=%v", header)
	}

	type acc struct {
		name    string
		dept    string
		history []ports.YearlyHeadcountSCB
	}
	byOrg := map[string]*acc{}

	cell := func(row []string, i int) string {
		if i < 0 || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	for _, row := range rows[1:] {
		org := normalizeOrgNumber(cell(row, iOrg))
		if org == "" {
			continue
		}
		year, err := strconv.Atoi(cell(row, iYear))
		if err != nil {
			continue
		}
		fte := 0
		if v := cell(row, iFTE); v != "" {
			normalized := strings.ReplaceAll(v, " ", "")
			if n, err := strconv.Atoi(normalized); err == nil {
				fte = n
			} else if fval, err := strconv.ParseFloat(strings.ReplaceAll(normalized, ",", "."), 64); err == nil {
				fte = int(fval + 0.5)
			}
		}

		a := byOrg[org]
		if a == nil {
			a = &acc{}
			byOrg[org] = a
		}
		if n := cell(row, iName); n != "" {
			a.name = n
		}
		if iDept >= 0 {
			if d := cell(row, iDept); d != "" {
				a.dept = d
			}
		}
		a.history = append(a.history, ports.YearlyHeadcountSCB{Year: year, HeadcountInt: fte})
	}

	out := make([]ports.HeadcountEntry, 0, len(byOrg))
	for org, a := range byOrg {
		sortAscByYear(a.history)
		snapYear, snapFTE := latestNonZero(a.history)
		out = append(out, ports.HeadcountEntry{
			OrgNumber:  org,
			Name:       a.name,
			Department: a.dept,
			Year:       snapYear,
			Headcount:  snapFTE,
			History:    a.history,
		})
	}
	return out, nil
}

func sortAscByYear(h []ports.YearlyHeadcountSCB) {
	for i := 1; i < len(h); i++ {
		for j := i; j > 0 && h[j-1].Year > h[j].Year; j-- {
			h[j-1], h[j] = h[j], h[j-1]
		}
	}
}

func latestNonZero(h []ports.YearlyHeadcountSCB) (int, int) {
	for i := len(h) - 1; i >= 0; i-- {
		if h[i].HeadcountInt > 0 {
			return h[i].Year, h[i].HeadcountInt
		}
	}
	if len(h) > 0 {
		return h[len(h)-1].Year, 0
	}
	return 0, 0
}
