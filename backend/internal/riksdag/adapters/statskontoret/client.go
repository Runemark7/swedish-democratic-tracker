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
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"riksdagskollen/internal/riksdag/ports"
)

type anslagInfo struct {
	name      string
	role      string
	headcount string
}

// targetAnslag maps Statskontoret appropriation codes (UUAASSS) to agency metadata.
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

// historyFromYear is the earliest year included in per-agency history.
const historyFromYear = 2010

func downloadURL(year int) string {
	filename := fmt.Sprintf("%%C3%%85rsutfall%%20utgifter%%201997%%20-%%20%d,%%20definitivt.zip", year)
	return fmt.Sprintf(
		"https://www.statskontoret.se/OpenDataArsUtfallPage/GetFile?documentType=Utgift&fileType=Zip&fileName=%s&Year=%d&month=0&status=Definitiv",
		filename, year,
	)
}

type Client struct {
	http     *http.Client
	mu       sync.Mutex
	cached   []ports.AuthorityData
	cachedAt time.Time
	cacheTTL time.Duration
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

func (c *Client) fetchYear(ctx context.Context, latestYear int) ([]ports.AuthorityData, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL(latestYear), nil)
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

	return parseCSV(f, latestYear)
}

// parseCSV reads the semicolon-delimited CSV and extracts per-agency utfall for all years.
//
// Relevant columns (0-indexed):
//
//	2:  Anslag   (e.g. "0401001")
//	4:  År       (e.g. "2024")
//	10: Utfall   (Mkr, Swedish decimal comma)
func parseCSV(r io.Reader, latestYear int) ([]ports.AuthorityData, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})

	cr := csv.NewReader(bytes.NewReader(raw))
	cr.Comma = ';'
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	if _, err := cr.Read(); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	// yearSums[anslagCode][year] → total Mkr
	yearSums := make(map[string]map[int]float64)
	for code := range targetAnslag {
		yearSums[code] = make(map[int]float64)
	}

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
		year, err := strconv.Atoi(strings.TrimSpace(row[4]))
		if err != nil || year < historyFromYear {
			continue
		}
		utfallStr := strings.TrimSpace(row[10])
		if utfallStr == "" {
			continue
		}
		val, err := strconv.ParseFloat(strings.ReplaceAll(utfallStr, ",", "."), 64)
		if err != nil {
			continue
		}
		yearSums[anslag][year] += val
	}

	var result []ports.AuthorityData
	for code, byYear := range yearSums {
		if len(byYear) == 0 {
			continue
		}
		info := targetAnslag[code]

		history := make([]ports.YearlyExpenditure, 0, len(byYear))
		for yr, mkr := range byYear {
			history = append(history, ports.YearlyExpenditure{
				Year:            yr,
				ExpenditureMdkr: mkr / 1000,
			})
		}
		sort.Slice(history, func(i, j int) bool { return history[i].Year < history[j].Year })

		latest := history[len(history)-1]
		result = append(result, ports.AuthorityData{
			Name:            info.name,
			Role:            info.role,
			Headcount:       info.headcount,
			ExpenditureMdkr: latest.ExpenditureMdkr,
			Year:            latest.Year,
			History:         history,
		})
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no matching rows found")
	}
	return result, nil
}
