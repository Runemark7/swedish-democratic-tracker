package ted

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"riksdagskollen/internal/regions/ports"
)

const tedSearchURL = "https://api.ted.europa.eu/v3/notices/search"

// maxPages caps the number of paginated TED API calls per request.
const maxPages = 10

// tedDateRange returns a TED publication-date filter covering the last 4 full calendar years.
func tedDateRange() (from, to string) {
	now := time.Now()
	end := now.Year() - 1
	start := end - 3
	return fmt.Sprintf("%d0101", start), fmt.Sprintf("%d1231", end)
}

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 20 * time.Second}}
}

func (c *Client) FetchProcurement(ctx context.Context, buyerName string) ([]ports.ProcurementNotice, error) {
	from, to := tedDateRange()
	query := fmt.Sprintf(
		"buyer-name ~ %s AND BT-02-notice = can-standard AND publication-date >= %s AND publication-date <= %s",
		quoteIfNeeded(buyerName), from, to,
	)

	var all []ports.ProcurementNotice
	for page := 1; page <= maxPages; page++ {
		body := map[string]any{
			"query":  query,
			"fields": []string{"publication-date", "classification-cpv", "total-value", "total-value-cur"},
			"limit":  100,
			"page":   page,
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, tedSearchURL, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("TED API request failed: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("TED API returned %d", resp.StatusCode)
		}

		var payload struct {
			Notices []struct {
				PublicationDate string   `json:"publication-date"`
				CPV             []string `json:"classification-cpv"`
				TotalValue      *float64 `json:"total-value"`
				TotalValueCur   []string `json:"total-value-cur"`
			} `json:"notices"`
			TotalNoticeCount int `json:"totalNoticeCount"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode TED response: %w", err)
		}
		resp.Body.Close()

		for _, n := range payload.Notices {
			if len(n.CPV) == 0 || n.TotalValue == nil || *n.TotalValue <= 0 {
				continue
			}
			div := ""
			if len(n.CPV[0]) >= 2 {
				div = n.CPV[0][:2]
			}
			cur := "SEK"
			if len(n.TotalValueCur) > 0 {
				cur = n.TotalValueCur[0]
			}
			year := extractYear(n.PublicationDate)
			all = append(all, ports.ProcurementNotice{
				CPVDivision: div,
				Value:       *n.TotalValue,
				Currency:    cur,
				Year:        year,
			})
		}

		fetched := (page-1)*100 + len(payload.Notices)
		if fetched >= payload.TotalNoticeCount || len(payload.Notices) < 100 {
			break
		}
	}
	return all, nil
}

func quoteIfNeeded(s string) string {
	if strings.ContainsAny(s, " \t") {
		return `"` + s + `"`
	}
	return s
}

func extractYear(dateStr string) int {
	// format: "2023-01-02+01:00" or "2023-01-02"
	if len(dateStr) >= 4 {
		y, err := strconv.Atoi(dateStr[:4])
		if err == nil {
			return y
		}
	}
	return 0
}
