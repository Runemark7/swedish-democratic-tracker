package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
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

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL not set")
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("connect db", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	httpClient := &http.Client{Timeout: 30 * time.Second}

	slog.Info("fetching municipality metadata from SCB")
	munMeta, err := getMeta(ctx, httpClient, kfmandatURL)
	if err != nil {
		slog.Error("fetch kfmandat meta", "err", err)
		os.Exit(1)
	}
	allMunCodes, allMunNames := codesAndNames(munMeta)
	// Keep only 4-digit numeric codes (actual municipalities, not national aggregates)
	var munCodes, munNames []string
	for i, code := range allMunCodes {
		if len(code) == 4 {
			munCodes = append(munCodes, code)
			munNames = append(munNames, allMunNames[i])
		}
	}
	slog.Info("found municipalities", "count", len(munCodes))

	slog.Info("fetching region metadata from SCB")
	regMeta, err := getMeta(ctx, httpClient, ltmandatURL)
	if err != nil {
		slog.Error("fetch ltmandat meta", "err", err)
		os.Exit(1)
	}
	allRegCodes, allRegNames := codesAndNames(regMeta)
	// Ltmandat uses codes like "01L", "03L" etc; "00L" is national aggregate
	var regCodes, regNames []string   // API codes (with L suffix, for querying)
	var regDBCodes []string           // DB codes (without L suffix)
	for i, code := range allRegCodes {
		if strings.HasSuffix(code, "L") && code != "00L" {
			regCodes = append(regCodes, code)
			regDBCodes = append(regDBCodes, strings.TrimSuffix(code, "L"))
			regNames = append(regNames, allRegNames[i])
		}
	}
	slog.Info("found regions", "count", len(regCodes))

	slog.Info("fetching municipal mandates (2022)")
	munMandates, err := fetchMandates(ctx, httpClient, kfmandatURL, munCodes, "ME0104C1")
	if err != nil {
		slog.Error("fetch municipal mandates", "err", err)
		os.Exit(1)
	}

	slog.Info("fetching regional mandates (2022)")
	regMandates, err := fetchMandates(ctx, httpClient, ltmandatURL, regCodes, "ME0104C2")
	if err != nil {
		slog.Error("fetch regional mandates", "err", err)
		os.Exit(1)
	}

	slog.Info("fetching population metadata to get valid codes")
	popMeta, err := getMeta(ctx, httpClient, befolkningURL)
	if err != nil {
		slog.Error("fetch population meta", "err", err)
		os.Exit(1)
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
	slog.Info("fetching population for all municipalities", "querying", len(popQueryCodes))
	populations, err := fetchPopulations(ctx, httpClient, popQueryCodes)
	if err != nil {
		slog.Error("fetch populations", "err", err)
		os.Exit(1)
	}

	// ── Upsert municipalities ─────────────────────────────────────────────
	slog.Info("upserting municipalities")
	for i, code := range munCodes {
		name := munNames[i]
		regionCode := code[:2]
		pop := populations[code]
		byParty := munMandates[code]
		total := sumMandates(byParty)
		governing := deriveGoverning(byParty, total)

		_, err := pool.Exec(ctx, `
			INSERT INTO municipalities (code, name, region_code, population, governing_parties, election_year, total_mandates)
			VALUES ($1, $2, $3, $4, $5, 2022, $6)
			ON CONFLICT (code) DO UPDATE SET
				name = EXCLUDED.name,
				population = EXCLUDED.population,
				governing_parties = EXCLUDED.governing_parties,
				total_mandates = EXCLUDED.total_mandates,
				election_year = 2022
		`, code, name, regionCode, pop, governing, total)
		if err != nil {
			slog.Error("upsert municipality", "code", code, "err", err)
			os.Exit(1)
		}
	}
	slog.Info("municipalities done", "count", len(munCodes))

	// ── Replace municipal election results ────────────────────────────────
	slog.Info("replacing municipal election results")
	for _, code := range munCodes {
		byParty := munMandates[code]
		total := sumMandates(byParty)
		if total == 0 {
			continue
		}

		_, err := pool.Exec(ctx, `DELETE FROM municipal_election_results WHERE municipality_code = $1`, code)
		if err != nil {
			slog.Error("delete results", "code", code, "err", err)
			os.Exit(1)
		}

		for party, mandates := range byParty {
			if mandates == 0 {
				continue
			}
			votePct := math.Round(float64(mandates)/float64(total)*10000) / 100
			_, err := pool.Exec(ctx, `
				INSERT INTO municipal_election_results (municipality_code, party, mandates, vote_pct, total_mandates)
				VALUES ($1, $2, $3, $4, $5)
			`, code, party, mandates, votePct, total)
			if err != nil {
				slog.Error("insert result", "code", code, "party", party, "err", err)
				os.Exit(1)
			}
		}
	}
	slog.Info("municipal election results done")

	// ── Update regions ────────────────────────────────────────────────────
	slog.Info("updating regions")
	for i, apiCode := range regCodes {
		dbCode := regDBCodes[i]
		name := regNames[i]
		byParty := regMandates[apiCode]
		total := sumMandates(byParty)
		governing := deriveGoverning(byParty, total)

		_, err := pool.Exec(ctx, `
			UPDATE regions SET
				governing_parties = $1,
				total_mandates = $2,
				election_year = 2022
			WHERE code = $3
		`, governing, total, dbCode)
		if err != nil {
			slog.Error("update region", "code", dbCode, "name", name, "err", err)
			os.Exit(1)
		}
	}
	slog.Info("regions done")

	// ── Replace regional election results ─────────────────────────────────
	slog.Info("replacing regional election results")
	for i, apiCode := range regCodes {
		dbCode := regDBCodes[i]
		byParty := regMandates[apiCode]
		total := sumMandates(byParty)
		if total == 0 {
			continue
		}

		_, err := pool.Exec(ctx, `DELETE FROM regional_election_results WHERE region_code = $1`, dbCode)
		if err != nil {
			slog.Error("delete regional results", "code", dbCode, "err", err)
			os.Exit(1)
		}

		for party, mandates := range byParty {
			if mandates == 0 {
				continue
			}
			votePct := math.Round(float64(mandates)/float64(total)*10000) / 100
			_, err := pool.Exec(ctx, `
				INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates)
				VALUES ($1, $2, $3, $4, $5)
			`, dbCode, party, mandates, votePct, total)
			if err != nil {
				slog.Error("insert regional result", "code", dbCode, "party", party, "err", err)
				os.Exit(1)
			}
		}
	}
	slog.Info("regional election results done")

	slog.Info("seeding complete")
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
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
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
	// Batch into chunks of 50 to avoid SCB request size limits
	for i := 0; i < len(munCodes); i += 50 {
		end := i + 50
		if end > len(munCodes) {
			end = len(munCodes)
		}
		batch := munCodes[i:end]
		partial, err := fetchPopulationBatch(ctx, c, batch)
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
