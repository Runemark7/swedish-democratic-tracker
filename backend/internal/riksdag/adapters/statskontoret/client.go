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
	name            string
	role            string
	ministry        string
	headcount       string
	headcountInt    int
	description     string
	websiteURL      string
	annualReportURL string
}

// targetAnslag maps Statskontoret appropriation codes (UUAASSS) to agency metadata.
var targetAnslag = map[string]anslagInfo{
	"0301001": {
		name: "Skatteverket", role: "Skatt & folkbokföring",
		ministry: "Finansdepartementet", headcount: "11 100", headcountInt: 11100,
		description:     "Ansvarar för beskattning, folkbokföring och bouppteckningar i hela Sverige.",
		websiteURL:      "https://www.skatteverket.se",
		annualReportURL: "https://www.skatteverket.se/omoss/omskatteverket/publikationer/arsredovisning.4.html",
	},
	"0301002": {
		name: "Tullverket", role: "Tull & gränskontroll",
		ministry: "Finansdepartementet", headcount: "2 700", headcountInt: 2700,
		description:     "Kontrollerar in- och utförsel av varor vid Sveriges gränser och bekämpar smuggling.",
		websiteURL:      "https://www.tullverket.se",
		annualReportURL: "https://www.tullverket.se/omtullverket/publikationer/arsredovisningar.4.html",
	},
	"0401001": {
		name: "Polismyndigheten", role: "Ordning & utredning",
		ministry: "Justitiedepartementet", headcount: "35 500", headcountInt: 35500,
		description:     "Sveriges största myndighet, ansvarar för brottsbekämpning, utredning och ordningshållning.",
		websiteURL:      "https://polisen.se",
		annualReportURL: "https://polisen.se/om-polisen/organisation/arsredovisning/",
	},
	"0401002": {
		name: "Säkerhetspolisen", role: "Nationell säkerhet",
		ministry: "Justitiedepartementet", headcount: "2 100", headcountInt: 2100,
		description:     "Skyddar Sverige mot terrorism, spionage och andra hot mot den nationella säkerheten.",
		websiteURL:      "https://www.sakerhetspolisen.se",
		annualReportURL: "https://www.sakerhetspolisen.se/om-sapo/publikationer/arsredovisningar.html",
	},
	"0401003": {
		name: "Åklagarmyndigheten", role: "Åklagare",
		ministry: "Justitiedepartementet", headcount: "1 900", headcountInt: 1900,
		description:     "Leder förundersökningar och väcker åtal i brottmål vid Sveriges allmänna domstolar.",
		websiteURL:      "https://www.aklagare.se",
		annualReportURL: "https://www.aklagare.se/om-aklagarmyndigheten/publikationer/arsredovisningar/",
	},
	"0401005": {
		name: "Sveriges Domstolar", role: "Domstolar & nämnder",
		ministry: "Justitiedepartementet", headcount: "7 200", headcountInt: 7200,
		description:     "Samlingsnamn för landets domstolar och nämnder — tingsrätter, hovrätter, förvaltningsrätter och Högsta domstolen.",
		websiteURL:      "https://www.domstol.se",
		annualReportURL: "https://www.domstol.se/om-sveriges-domstolar/publikationer/arsredovisningar/",
	},
	"0401006": {
		name: "Kriminalvården", role: "Kriminalvård & häkte",
		ministry: "Justitiedepartementet", headcount: "14 200", headcountInt: 14200,
		description:     "Ansvarar för häkten, fängelser och frivård med målet att minska återfall i brott.",
		websiteURL:      "https://www.kriminalvarden.se",
		annualReportURL: "https://www.kriminalvarden.se/om-kriminalvarden/publikationer/arsredovisningar/",
	},
	"0801001": {
		name: "Migrationsverket", role: "Uppehållstillstånd & asyl",
		ministry: "Justitiedepartementet", headcount: "5 800", headcountInt: 5800,
		description:     "Prövar ansökningar om uppehållstillstånd, asyl, medborgarskap och arbetstillstånd.",
		websiteURL:      "https://www.migrationsverket.se",
		annualReportURL: "https://www.migrationsverket.se/Om-Migrationsverket/Fakta-och-statistik/Publikationer/Arsredovisningar.html",
	},
	"1002001": {
		name: "Försäkringskassan", role: "Administration socialförsäkring",
		ministry: "Socialdepartementet", headcount: "14 200", headcountInt: 14200,
		description:     "Administrerar socialförsäkringssystemet inklusive sjukpenning, föräldrapenning och aktivitetsersättning.",
		websiteURL:      "https://www.forsakringskassan.se",
		annualReportURL: "https://www.forsakringskassan.se/om-forsakringskassan/publikationer/arsredovisning",
	},
	"1401001": {
		name: "Arbetsförmedlingen", role: "Matchning & arbetsmarknadspolitik",
		ministry: "Arbetsmarknadsdepartementet", headcount: "9 400", headcountInt: 9400,
		description:     "Ansvarar för arbetsförmedling, matchning mellan arbetsgivare och arbetssökande, och genomförande av arbetsmarknadspolitiken.",
		websiteURL:      "https://www.arbetsformedlingen.se",
		annualReportURL: "https://www.arbetsformedlingen.se/om-oss/fakta-om-af/publikationer/arsredovisning.html",
	},
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
			Ministry:        info.ministry,
			Headcount:       info.headcount,
			HeadcountInt:    info.headcountInt,
			Description:     info.description,
			WebsiteURL:      info.websiteURL,
			AnnualReportURL: info.annualReportURL,
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
