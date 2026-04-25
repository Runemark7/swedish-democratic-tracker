package statskontoret

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"riksdagskollen/internal/riksdag/ports"
)

// anslagInfo holds the canonical name, role, and headcount for a target appropriation code.
type anslagInfo struct {
	name      string
	role      string
	headcount string
}

// targetAnslag maps Statskontoret appropriation codes (format "UUAASSS") to agency metadata.
// Values verified against the 2025 årsutfall definitiv (covers 2024 spend).
var targetAnslag = map[string]anslagInfo{
	"0301001": {"Skatteverket", "Skatt & folkbokföring", "11 100"},
	"0301002": {"Tullverket", "Tull & gränskontroll", "2 700"},
	"0401001": {"Polismyndigheten", "Ordning & utredning", "35 500"},
	"0401002": {"Säkerhetspolisen", "Nationell säkerhet", "2 100"},
	"0401003": {"Åklagarmyndigheten", "Åklagare", "1 900"},
	"0401005": {"Sveriges Domstolar", "Domstolar & nämnder", "7 200"},
	"0401006": {"Kriminalvården", "Kriminalvård & häkte", "14 200"},
	"0801001": {"Migrationsverket", "Uppehållstillstånd & asyl", "5 800"},
	"1002001": {"Försäkringskassan", "Administration socialförsäkring", "14 200"},
	"1401001": {"Arbetsförmedlingen", "Matchning & arbetsmarknadspolitik", "9 400"},
}

// downloadURL builds the Statskontoret open data download URL for the given year.
func downloadURL(year int) string {
	filename := fmt.Sprintf("%%C3%%85rsutfall%%20utgifter%%201997%%20-%%20%d,%%20definitivt.zip", year)
	return fmt.Sprintf(
		"https://www.statskontoret.se/OpenDataArsUtfallPage/GetFile?documentType=Utgift&fileType=Zip&fileName=%s&Year=%d&month=0&status=Definitiv",
		filename, year,
	)
}

type Client struct {
	http      *http.Client
	mu        sync.Mutex
	cached    []ports.AuthorityData
	cachedAt  time.Time
	cacheTTL  time.Duration
}

func NewClient() *Client {
	return &Client{
		http:     &http.Client{Timeout: 30 * time.Second},
		cacheTTL: 24 * time.Hour,
	}
}

func (c *Client) FetchAuthorities(ctx context.Context) ([]ports.AuthorityData, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.cached) > 0 && time.Since(c.cachedAt) < c.cacheTTL {
		return c.cached, nil
	}

	now := time.Now()
	// Try current year - 1 (definitiv data for last year); fall back to year - 2.
	for _, year := range []int{now.Year() - 1, now.Year() - 2} {
		data, err := c.fetchYear(ctx, year)
		if err != nil {
			slog.Warn("statskontoret fetch failed", "year", year, "error", err)
			continue
		}
		c.cached = data
		c.cachedAt = now
		return data, nil
	}
	return nil, fmt.Errorf("statskontoret: all year attempts failed")
}

func (c *Client) fetchYear(ctx context.Context, year int) ([]ports.AuthorityData, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL(year), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, fmt.Errorf("unzip: %w", err)
	}
	if len(zr.File) == 0 {
		return nil, fmt.Errorf("unzip: empty archive")
	}

	f, err := zr.File[0].Open()
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	return parseCSV(f, year)
}

// parseCSV reads the semicolon-delimited UTF-8 CSV and extracts per-agency utfall.
//
// Expected columns (0-indexed):
//
//	2: Anslag          (e.g. "0401001")
//	4: År              (e.g. "2024")
//	10: Utfall         (Mkr, Swedish decimal comma)
func parseCSV(r io.Reader, year int) ([]ports.AuthorityData, error) {
	// Strip UTF-8 BOM if present.
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})

	cr := csv.NewReader(bytes.NewReader(raw))
	cr.Comma = ';'
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	// Skip header row.
	if _, err := cr.Read(); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	yearStr := strconv.Itoa(year)
	sums := make(map[string]float64)

	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}
		if len(row) < 11 {
			continue
		}
		anslag := strings.TrimSpace(row[2])
		if _, ok := targetAnslag[anslag]; !ok {
			continue
		}
		if strings.TrimSpace(row[4]) != yearStr {
			continue
		}
		utfallStr := strings.TrimSpace(row[10])
		if utfallStr == "" {
			continue
		}
		// Swedish decimal: replace comma with period.
		utfallStr = strings.ReplaceAll(utfallStr, ",", ".")
		val, err := strconv.ParseFloat(utfallStr, 64)
		if err != nil {
			continue
		}
		sums[anslag] += val
	}

	if len(sums) == 0 {
		return nil, fmt.Errorf("no matching rows found for year %d", year)
	}

	result := make([]ports.AuthorityData, 0, len(sums))
	for code, mkr := range sums {
		info := targetAnslag[code]
		result = append(result, ports.AuthorityData{
			Name:            info.name,
			Role:            info.role,
			Headcount:       info.headcount,
			ExpenditureMdkr: mkr / 1000, // Mkr → mdkr
			Year:            year,
		})
	}
	return result, nil
}
