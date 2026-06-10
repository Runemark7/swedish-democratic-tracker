package seeder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	scbBase       = "https://api.scb.se/OV0104/v1/doris/sv/ssd"
	kfmandatURL   = scbBase + "/ME/ME0104/ME0104A/Kfmandat"
	ltmandatURL   = scbBase + "/ME/ME0104/ME0104B/Ltmandat"
	befolkningURL = scbBase + "/BE/BE0101/BE0101A/BefolkningNy"
)

var (
	rightBloc = map[string]bool{"M": true, "SD": true, "KD": true, "L": true, "C": true}
	leftBloc  = map[string]bool{"S": true, "V": true, "MP": true}
)

var allParties = []string{"M", "C", "FP", "KD", "MP", "S", "V", "SD", "ÖVRIGA"}

type scbMeta struct {
	Variables []struct {
		Code       string   `json:"code"`
		Values     []string `json:"values"`
		ValueTexts []string `json:"valueTexts"`
	} `json:"variables"`
}

type scbData struct {
	Data []struct {
		Key    []string `json:"key"`
		Values []string `json:"values"`
	} `json:"data"`
}

// Run seeds all 291 municipalities and 21 regions from the SCB PxWeb API (2022 election data).
func Run(ctx context.Context, pool *pgxpool.Pool) error {
	httpClient := &http.Client{Timeout: 30 * time.Second}

	slog.Info("seeder: fetching municipality metadata from SCB")
	munMeta, err := getMeta(ctx, httpClient, kfmandatURL)
	if err != nil {
		return fmt.Errorf("fetch kfmandat meta: %w", err)
	}
	allMunCodes, allMunNames := codesAndNames(munMeta)
	var munCodes, munNames []string
	for i, code := range allMunCodes {
		if len(code) == 4 {
			munCodes = append(munCodes, code)
			munNames = append(munNames, allMunNames[i])
		}
	}
	slog.Info("seeder: found municipalities", "count", len(munCodes))

	slog.Info("seeder: fetching region metadata from SCB")
	regMeta, err := getMeta(ctx, httpClient, ltmandatURL)
	if err != nil {
		return fmt.Errorf("fetch ltmandat meta: %w", err)
	}
	allRegCodes, allRegNames := codesAndNames(regMeta)

	// Valid current region codes come from the regions table (seeded by
	// migration). SCB's Ltmandat list also contains "Summa regioner" (00L) and
	// historical landsting — e.g. 12LG Malmöhus, 14LG Bohuslandstinget, 11L
	// Kristianstad — that must NOT map to a current region. Conversely, some
	// current regions use an "NNLG" code rather than "NNL": notably Dalarna is
	// "20LG" with no plain "20L". So we select, per current region code (the
	// leading two digits), the plain "NNL" code when present and fall back to
	// the "NNLG" variant otherwise. The earlier HasSuffix("L") filter silently
	// dropped Dalarna, leaving it on the migration's placeholder mandate count.
	validRegion := make(map[string]bool)
	rows, qerr := pool.Query(ctx, `SELECT code FROM regions`)
	if qerr != nil {
		return fmt.Errorf("load region codes: %w", qerr)
	}
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			rows.Close()
			return fmt.Errorf("scan region code: %w", err)
		}
		validRegion[c] = true
	}
	rows.Close()

	chosenCode := make(map[string]string) // dbCode -> SCB Ltmandat api code
	chosenName := make(map[string]string)
	for i, code := range allRegCodes {
		if code == "00L" || len(code) < 2 {
			continue
		}
		db := code[:2]
		if !validRegion[db] {
			continue
		}
		if existing, ok := chosenCode[db]; ok {
			// Prefer the plain "NNL" code over an "NNLG" variant.
			if strings.HasSuffix(existing, "G") && !strings.HasSuffix(code, "G") {
				chosenCode[db], chosenName[db] = code, allRegNames[i]
			}
			continue
		}
		chosenCode[db], chosenName[db] = code, allRegNames[i]
	}

	var regCodes, regDBCodes, regNames []string
	for db, code := range chosenCode {
		regCodes = append(regCodes, code)
		regDBCodes = append(regDBCodes, db)
		regNames = append(regNames, chosenName[db])
	}
	slog.Info("seeder: found regions", "count", len(regCodes))

	slog.Info("seeder: fetching municipal mandates (2022)")
	munMandates, err := fetchMandates(ctx, httpClient, kfmandatURL, munCodes, "ME0104C1")
	if err != nil {
		return fmt.Errorf("fetch municipal mandates: %w", err)
	}

	slog.Info("seeder: fetching regional mandates (2022)")
	regMandates, err := fetchMandates(ctx, httpClient, ltmandatURL, regCodes, "ME0104C2")
	if err != nil {
		return fmt.Errorf("fetch regional mandates: %w", err)
	}

	slog.Info("seeder: fetching population metadata")
	popMeta, err := getMeta(ctx, httpClient, befolkningURL)
	if err != nil {
		return fmt.Errorf("fetch population meta: %w", err)
	}
	validPopCodes := make(map[string]bool)
	for _, v := range popMeta.Variables {
		if v.Code == "Region" {
			for _, c := range v.Values {
				validPopCodes[c] = true
			}
		}
	}
	var popQueryCodes []string
	for _, code := range munCodes {
		if validPopCodes[code] {
			popQueryCodes = append(popQueryCodes, code)
		}
	}
	slog.Info("seeder: fetching population", "querying", len(popQueryCodes))
	populations, err := fetchPopulations(ctx, httpClient, popQueryCodes)
	if err != nil {
		return fmt.Errorf("fetch populations: %w", err)
	}

	slog.Info("seeder: upserting municipalities")
	for i, code := range munCodes {
		name := munNames[i]
		regionCode := code[:2]
		pop := populations[code]
		byParty := munMandates[code]
		total := sumMandates(byParty)
		governing := deriveGoverning(byParty, total)

		if _, err := pool.Exec(ctx, `
			INSERT INTO municipalities (code, name, region_code, population, governing_parties, election_year, total_mandates)
			VALUES ($1, $2, $3, $4, $5, 2022, $6)
			ON CONFLICT (code) DO UPDATE SET
				name = EXCLUDED.name,
				population = EXCLUDED.population,
				governing_parties = EXCLUDED.governing_parties,
				total_mandates = EXCLUDED.total_mandates,
				election_year = 2022
		`, code, name, regionCode, pop, governing, total); err != nil {
			return fmt.Errorf("upsert municipality %s: %w", code, err)
		}
	}
	slog.Info("seeder: municipalities done", "count", len(munCodes))

	slog.Info("seeder: replacing municipal election results")
	for _, code := range munCodes {
		byParty := munMandates[code]
		total := sumMandates(byParty)
		if total == 0 {
			continue
		}
		if _, err := pool.Exec(ctx, `DELETE FROM municipal_election_results WHERE municipality_code = $1`, code); err != nil {
			return fmt.Errorf("delete results %s: %w", code, err)
		}
		for party, mandates := range byParty {
			if mandates == 0 {
				continue
			}
			votePct := math.Round(float64(mandates)/float64(total)*10000) / 100
			if _, err := pool.Exec(ctx, `
				INSERT INTO municipal_election_results (municipality_code, party, mandates, vote_pct, total_mandates)
				VALUES ($1, $2, $3, $4, $5)
			`, code, party, mandates, votePct, total); err != nil {
				return fmt.Errorf("insert result %s/%s: %w", code, party, err)
			}
		}
	}
	slog.Info("seeder: municipal election results done")

	slog.Info("seeder: updating regions")
	for i, apiCode := range regCodes {
		dbCode := regDBCodes[i]
		name := regNames[i]
		byParty := regMandates[apiCode]
		total := sumMandates(byParty)
		governing := deriveGoverning(byParty, total)

		if _, err := pool.Exec(ctx, `
			UPDATE regions SET governing_parties = $1, total_mandates = $2, election_year = 2022 WHERE code = $3
		`, governing, total, dbCode); err != nil {
			return fmt.Errorf("update region %s (%s): %w", dbCode, name, err)
		}
	}
	slog.Info("seeder: regions done")

	slog.Info("seeder: replacing regional election results")
	for i, apiCode := range regCodes {
		dbCode := regDBCodes[i]
		byParty := regMandates[apiCode]
		total := sumMandates(byParty)
		if total == 0 {
			continue
		}
		if _, err := pool.Exec(ctx, `DELETE FROM regional_election_results WHERE region_code = $1`, dbCode); err != nil {
			return fmt.Errorf("delete regional results %s: %w", dbCode, err)
		}
		for party, mandates := range byParty {
			if mandates == 0 {
				continue
			}
			votePct := math.Round(float64(mandates)/float64(total)*10000) / 100
			if _, err := pool.Exec(ctx, `
				INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates)
				VALUES ($1, $2, $3, $4, $5)
			`, dbCode, party, mandates, votePct, total); err != nil {
				return fmt.Errorf("insert regional result %s/%s: %w", dbCode, party, err)
			}
		}
	}
	slog.Info("seeder: updating region populations")
	regionDBCodes, err := allRegionCodes(ctx, pool)
	if err != nil {
		return fmt.Errorf("load region codes: %w", err)
	}
	// SCB BE0101 accepts 2-digit county codes, which are exactly our region
	// codes. Region population had been a frozen, hand-entered migration value;
	// fetch it live like municipalities so it stays sourced and current.
	regionPops, err := fetchPopulations(ctx, httpClient, regionDBCodes)
	if err != nil {
		return fmt.Errorf("fetch region populations: %w", err)
	}
	for code, pop := range regionPops {
		if _, err := pool.Exec(ctx, `UPDATE regions SET population = $1 WHERE code = $2`, pop, code); err != nil {
			return fmt.Errorf("update region population %s: %w", code, err)
		}
	}
	slog.Info("seeder: region populations done", "count", len(regionPops))

	slog.Info("seeder: complete")
	return nil
}

// allRegionCodes returns every region code currently in the regions table.
func allRegionCodes(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	rows, err := pool.Query(ctx, `SELECT code FROM regions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var codes []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

func getMeta(ctx context.Context, c *http.Client, url string) (*scbMeta, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SCB meta returned %d", resp.StatusCode)
	}
	var m scbMeta
	return &m, json.NewDecoder(resp.Body).Decode(&m)
}

func codesAndNames(m *scbMeta) (codes, names []string) {
	for _, v := range m.Variables {
		if v.Code == "Region" {
			return v.Values, v.ValueTexts
		}
	}
	return nil, nil
}

func fetchMandates(ctx context.Context, c *http.Client, url string, codes []string, contentsCode string) (map[string]map[string]int, error) {
	body := map[string]any{
		"query": []map[string]any{
			{"code": "Region", "selection": map[string]any{"filter": "item", "values": codes}},
			{"code": "Parti", "selection": map[string]any{"filter": "item", "values": allParties}},
			{"code": "ContentsCode", "selection": map[string]any{"filter": "item", "values": []string{contentsCode}}},
			{"code": "Tid", "selection": map[string]any{"filter": "item", "values": []string{"2022"}}},
		},
		"response": map[string]any{"format": "json"},
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SCB mandates returned %d", resp.StatusCode)
	}
	var data scbData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	result := make(map[string]map[string]int)
	for _, row := range data.Data {
		if len(row.Key) < 2 || len(row.Values) == 0 {
			continue
		}
		regionCode := row.Key[0]
		party := normaliseParty(row.Key[1])
		mandates, err := strconv.Atoi(row.Values[0])
		if err != nil {
			continue
		}
		if result[regionCode] == nil {
			result[regionCode] = make(map[string]int)
		}
		result[regionCode][party] += mandates
	}
	return result, nil
}

func fetchPopulations(ctx context.Context, c *http.Client, munCodes []string) (map[string]int, error) {
	result := make(map[string]int, len(munCodes))
	for i := 0; i < len(munCodes); i += 50 {
		end := i + 50
		if end > len(munCodes) {
			end = len(munCodes)
		}
		partial, err := fetchPopulationBatch(ctx, c, munCodes[i:end])
		if err != nil {
			return nil, err
		}
		for k, v := range partial {
			result[k] = v
		}
	}
	return result, nil
}

func fetchPopulationBatch(ctx context.Context, c *http.Client, munCodes []string) (map[string]int, error) {
	body := map[string]any{
		"query": []map[string]any{
			{"code": "Region", "selection": map[string]any{"filter": "item", "values": munCodes}},
			{"code": "ContentsCode", "selection": map[string]any{"filter": "item", "values": []string{"BE0101N1"}}},
			{"code": "Tid", "selection": map[string]any{"filter": "item", "values": []string{"2023"}}},
		},
		"response": map[string]any{"format": "json"},
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, befolkningURL, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SCB population returned %d for batch starting %s", resp.StatusCode, munCodes[0])
	}
	var data scbData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	result := make(map[string]int, len(munCodes))
	for _, row := range data.Data {
		if len(row.Key) < 2 || len(row.Values) == 0 {
			continue
		}
		pop, err := strconv.Atoi(row.Values[0])
		if err != nil {
			continue
		}
		result[row.Key[0]] = pop
	}
	return result, nil
}

func normaliseParty(p string) string {
	if p == "FP" {
		return "L"
	}
	return p
}

func sumMandates(byParty map[string]int) int {
	total := 0
	for _, m := range byParty {
		total += m
	}
	return total
}

// FetchPopulations returns population per municipality code, fetched live from
// SCB in batches (the same call the startup seeder uses, Tid 2023). Intended
// for the value-verification audit. It first reads the table metadata and drops
// any code SCB does not list as a valid Region value, so one stale/unknown code
// cannot reject (HTTP 400) an entire batch. Codes filtered out are simply
// absent from the result; the caller reports them as "not in SCB set".
func FetchPopulations(ctx context.Context, c *http.Client, codes []string) (map[string]int, error) {
	meta, err := getMeta(ctx, c, befolkningURL)
	if err != nil {
		return nil, fmt.Errorf("population meta: %w", err)
	}
	valid := make(map[string]bool)
	for _, v := range meta.Variables {
		if v.Code == "Region" {
			for _, code := range v.Values {
				valid[code] = true
			}
		}
	}
	known := make([]string, 0, len(codes))
	for _, code := range codes {
		if valid[code] {
			known = append(known, code)
		}
	}
	return fetchPopulations(ctx, c, known)
}

// FetchMandateTotals returns total elected council seats per entity NAME,
// fetched live from SCB's Kfmandat (municipalities) and Ltmandat (regions)
// tables. It reuses the same PxWeb calls the startup seeder uses. Keyed by
// name so callers need not reproduce the region code→SCB code mapping.
func FetchMandateTotals(ctx context.Context, c *http.Client) (mun, reg map[string]int, err error) {
	munMeta, err := getMeta(ctx, c, kfmandatURL)
	if err != nil {
		return nil, nil, fmt.Errorf("kfmandat meta: %w", err)
	}
	munCodes, munNames := codesAndNames(munMeta)
	munByCode, err := fetchMandates(ctx, c, kfmandatURL, munCodes, "ME0104C1")
	if err != nil {
		return nil, nil, fmt.Errorf("kfmandat: %w", err)
	}
	regMeta, err := getMeta(ctx, c, ltmandatURL)
	if err != nil {
		return nil, nil, fmt.Errorf("ltmandat meta: %w", err)
	}
	regCodes, regNames := codesAndNames(regMeta)
	regByCode, err := fetchMandates(ctx, c, ltmandatURL, regCodes, "ME0104C2")
	if err != nil {
		return nil, nil, fmt.Errorf("ltmandat: %w", err)
	}

	mun = totalsByName(munCodes, munNames, munByCode)
	reg = totalsByName(regCodes, regNames, regByCode)
	return mun, reg, nil
}

// totalsByName maps each code's per-party mandate map to a name→total entry.
// Entries with zero mandates (e.g. historical landsting no longer in use) are omitted.
func totalsByName(codes, names []string, byCode map[string]map[string]int) map[string]int {
	out := make(map[string]int, len(codes))
	for i, code := range codes {
		total := sumMandates(byCode[code])
		if total == 0 {
			continue
		}
		name := code
		if i < len(names) {
			name = names[i]
		}
		out[name] = total
	}
	return out
}

func deriveGoverning(byParty map[string]int, total int) []string {
	if total == 0 {
		return []string{}
	}
	right, left := 0, 0
	for p, m := range byParty {
		if rightBloc[p] {
			right += m
		} else if leftBloc[p] {
			left += m
		}
	}
	majority := total/2 + 1
	var governing []string
	var bloc map[string]bool
	if right >= majority {
		bloc = rightBloc
	} else if left >= majority {
		bloc = leftBloc
	} else {
		return []string{}
	}
	for p := range bloc {
		if byParty[p] > 0 {
			governing = append(governing, p)
		}
	}
	return governing
}
