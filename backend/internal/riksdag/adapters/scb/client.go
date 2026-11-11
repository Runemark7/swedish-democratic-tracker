package scb

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"riksdagskollen/internal/riksdag/ports"
	"riksdagskollen/internal/scbpx"
)

// The SCB KLS table (scbpx.KLSHeadcount): monthly employees in statlig sektor
// by agency, read via PxWebApi v2 with a v1 fallback.

// klsCodeMap maps KLS agency codes to our canonical agency names.
// C021 = polisväsendet (includes Polismyndigheten + Säkerhetspolisen).
// Åklagarmyndigheten and Säkerhetspolisen have no separate codes in this table.
var klsCodeMap = map[string]string{
	"C021": "Polismyndigheten",
	"C022": "Sveriges Domstolar",
	"C025": "Kriminalvården",
	"C026": "Migrationsverket",
	"C051": "Försäkringskassan",
	"C071": "Skatteverket",
	"C076": "Tullverket",
	"C172": "Arbetsförmedlingen",
}

type Client struct {
	px       *scbpx.Client
	mu       sync.Mutex
	cached   []ports.HeadcountData
	cachedAt time.Time
	cacheTTL time.Duration
}

func NewClient() *Client {
	return &Client{
		px:       scbpx.New(&http.Client{Timeout: 30 * time.Second}),
		cacheTTL: 24 * time.Hour,
	}
}

func (c *Client) FetchHeadcounts(ctx context.Context) ([]ports.HeadcountData, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.cached) > 0 && time.Since(c.cachedAt) < c.cacheTTL {
		return c.cached, nil
	}

	data, err := c.fetch(ctx)
	if err != nil {
		return nil, err
	}
	c.cached = data
	c.cachedAt = time.Now()
	return data, nil
}

const historyFromYear = 2015

func (c *Client) fetch(ctx context.Context) ([]ports.HeadcountData, error) {
	latestYear := time.Now().Year() - 1

	// Request December of each year as the annual snapshot.
	var tidValues []string
	for y := historyFromYear; y <= latestYear; y++ {
		tidValues = append(tidValues, fmt.Sprintf("%dM12", y))
	}

	codes := make([]string, 0, len(klsCodeMap))
	for k := range klsCodeMap {
		codes = append(codes, k)
	}

	// Ask only for published months; one missing month rejects the query.
	tidValues, err := c.px.Has(ctx, scbpx.KLSHeadcount, "Tid", tidValues)
	if err != nil {
		return nil, fmt.Errorf("scb headcount: %w", err)
	}
	if len(tidValues) == 0 {
		return nil, fmt.Errorf("scb headcount: no published December months")
	}

	rows, err := c.px.Data(ctx, scbpx.KLSHeadcount, []scbpx.Select{
		{Code: "Myndighet", Values: codes},
		{Code: "Kon", Values: []string{"1+2"}},
		{Code: "Heltiddeltid", Values: []string{"HT+DT"}},
		{Code: "ContentsCode", Values: []string{"AM0102AB"}},
		{Code: "Tid", Values: tidValues},
	})
	if err != nil {
		return nil, fmt.Errorf("scb headcount: %w", err)
	}

	// Aggregate per agency: key layout is [Myndighet, Kon, Heltiddeltid, Tid].
	type agencyAcc struct {
		latestYear int
		latest     int
		history    []ports.YearlyHeadcount
	}
	byName := map[string]*agencyAcc{}

	for _, row := range rows {
		if len(row.Key) < 4 || len(row.Values) == 0 {
			continue
		}
		canonical, ok := klsCodeMap[row.Key[0]]
		if !ok {
			continue
		}
		// Tid is "YYYYMmm", extract year from first 4 chars.
		yr, err := strconv.Atoi(row.Key[3][:4])
		if err != nil {
			continue
		}
		count, err := strconv.Atoi(row.Values[0])
		if err != nil {
			continue
		}
		acc := byName[canonical]
		if acc == nil {
			acc = &agencyAcc{}
			byName[canonical] = acc
		}
		acc.history = append(acc.history, ports.YearlyHeadcount{Year: yr, HeadcountInt: count})
		if yr > acc.latestYear {
			acc.latestYear = yr
			acc.latest = count
		}
	}

	if len(byName) == 0 {
		return nil, fmt.Errorf("scb headcount: no matching agencies in response")
	}

	out := make([]ports.HeadcountData, 0, len(byName))
	for name, acc := range byName {
		sort.Slice(acc.history, func(i, j int) bool { return acc.history[i].Year < acc.history[j].Year })
		out = append(out, ports.HeadcountData{
			Name:             name,
			HeadcountInt:     acc.latest,
			Year:             acc.latestYear,
			HeadcountHistory: acc.history,
		})
	}
	return out, nil
}
