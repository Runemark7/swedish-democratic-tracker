package regions

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"riksdagskollen/internal/regions/domain"
	"riksdagskollen/internal/regions/ports"
)

// rollingYears returns the last n completed calendar years (excluding the current year).
// Source: Kolada and SCB data is published with a 1-year lag.
func rollingYears(n int) []int {
	current := time.Now().Year()
	years := make([]int, n)
	for i := range years {
		years[i] = current - 1 - i
	}
	return years
}

// Municipal KPIs fetched from Kolada — source: https://api.kolada.se/v3/kpi/{code}
var defaultKPIs = []string{
	// Verksamhet
	"N00900", // Kommunalskatt, skattesats (%) — kolada.se/kpi/N00900
	"N11037", // Nettokostnad förskola, kr/barn — kolada.se/kpi/N11037
	"N15027", // Nettokostnad grundskola, kr/elev — kolada.se/kpi/N15027
	"N20043", // Nettokostnad äldreomsorg totalt, kr/inv 80+ år — kolada.se/kpi/N20043
	"N03010", // Investeringar, kr/inv — kolada.se/kpi/N03010
	// Budget & ekonomi
	"N03007", // Nettokostnader totalt, kr/inv — kolada.se/kpi/N03007
	"N03102", // Resultat i % av skatteintäkter — kolada.se/kpi/N03102
	"N03106", // Soliditet inkl. ansvarsförbindelser (%) — kolada.se/kpi/N03106
	"N03040", // Kassalikviditet (%) — kolada.se/kpi/N03040
	"N03132", // Långfristiga skulder, kr/inv — kolada.se/kpi/N03132
	// Skola & arbetsmarknad
	"N15428", // Elever behöriga till gymnasiet, andel (%) — kolada.se/kpi/N15428
	"N00708", // Öppet arbetslösa och i program 20–64 år, andel (%) — kolada.se/kpi/N00708
}

// regionStripKPIs — shown in the region header strip.
// Region KPIs use the N6xxxx/N7xxxx/N8xxxx Kolada namespaces (mun_type "L" for landsting/region).
// Source: https://api.kolada.se/v3/kpi/{code}
var regionStripKPIs = []string{
	"N60008", // Nettokostnad hälso- och sjukvård totalt, kr/inv — kolada.se/kpi/N60008
	"N63016", // Resultat i % av skatteintäkter, region — kolada.se/kpi/N63016
	"N63007", // Soliditet inkl. ansvarsförbindelser, region (%) — kolada.se/kpi/N63007
	"N79173", // Medicinsk bedömning inom 3 dagar i primärvård, andel (%) — kolada.se/kpi/N79173
	"N79179", // Telefonsamtal till primärvården besvarade samma dag, andel (%) — kolada.se/kpi/N79179
	"N60404", // Resor med kollektivtrafik, resor/inv — kolada.se/kpi/N60404
	"N85012", // Nettokostnad regional utveckling totalt, kr/inv — kolada.se/kpi/N85012
}

// koladaRegionCode converts a Swedish 2-digit county code (e.g. "09") to the
// 4-digit zero-prefixed code Kolada expects (e.g. "0009").
func koladaRegionCode(code string) string {
	if len(code) >= 4 {
		return code
	}
	return "00" + code
}

// spendingKPIs — municipal spending breakdown by service area (kr/inv from Kolada).
// Multiplied by population to derive budget-equivalent SEK totals.
// Source: https://api.kolada.se/v3/kpi/{code}
var spendingKPIs = []string{
	"N11004", // Nettokostnad förskola, kr/inv — kolada.se/kpi/N11004
	"N15028", // Nettokostnad grundskola, kr/inv — kolada.se/kpi/N15028
	"N17014", // Nettokostnad gymnasieskola, kr/inv — kolada.se/kpi/N17014
	"N20014", // Nettokostnad äldreomsorg totalt, kr/inv — kolada.se/kpi/N20014
	"N30005", // Nettokostnad individ- och familjeomsorg, kr/inv — kolada.se/kpi/N30005
	"N07037", // Nettokostnad gata/väg, park och plan, kr/inv — kolada.se/kpi/N07037
	"N09022", // Nettokostnad fritid och kultur, kr/inv — kolada.se/kpi/N09022
	"N05011", // Nettokostnad politisk verksamhet, kr/inv — kolada.se/kpi/N05011
	"N45014", // Nettokostnad VA (vatten och avlopp), kr/inv — kolada.se/kpi/N45014
}

// spendingKPINames maps Kolada KPI codes to human-readable Swedish area names.
var spendingKPINames = map[string]string{
	"N11004": "Förskola",
	"N15028": "Grundskola",
	"N17014": "Gymnasieskola",
	"N20014": "Äldreomsorg",
	"N30005": "Individ & familj",
	"N07037": "Gata, park, plan",
	"N09022": "Fritid & kultur",
	"N05011": "Politisk verksamhet",
	"N45014": "Vatten & avlopp",
}

type Service struct {
	repo   ports.RegionRepository
	kolada ports.KoladaClient
	scb    ports.SCBClient
	ted    ports.TEDClient
}

func NewService(repo ports.RegionRepository, kolada ports.KoladaClient, scb ports.SCBClient, ted ports.TEDClient) *Service {
	return &Service{repo: repo, kolada: kolada, scb: scb, ted: ted}
}

func (s *Service) ListRegions(ctx context.Context) ([]*domain.Region, error) {
	return s.repo.ListRegions(ctx)
}

func (s *Service) GetRegion(ctx context.Context, code string) (*domain.RegionDetail, error) {
	return s.repo.GetRegion(ctx, code)
}

func (s *Service) ListMunicipalities(ctx context.Context, regionCode string) ([]*domain.Municipality, error) {
	return s.repo.ListMunicipalities(ctx, regionCode)
}

func (s *Service) GetMunicipality(ctx context.Context, code string) (*domain.MunicipalityDetail, error) {
	return s.repo.GetMunicipality(ctx, code)
}

func (s *Service) GetRegionBudget(ctx context.Context, regionCode string, year int) ([]ports.RegionBudgetArea, error) {
	return s.scb.FetchRegionBudget(ctx, regionCode, year)
}

func (s *Service) GetRegionBudgetMultiYear(ctx context.Context, regionCode string, years []int) ([]ports.RegionBudgetSnapshot, error) {
	return s.scb.FetchRegionBudgetMultiYear(ctx, regionCode, years)
}

func (s *Service) GetRegionBudgetHistory(ctx context.Context, regionCode string, years []int) ([]ports.RegionBudgetSnapshot, error) {
	return s.repo.GetRegionBudgetHistory(ctx, regionCode, years)
}

func (s *Service) GetAreaAcrossRegions(ctx context.Context, areaName string, year int) ([]ports.RegionAreaDataPoint, error) {
	return s.repo.GetAreaAcrossRegions(ctx, areaName, year)
}

func (s *Service) GetAreaAcrossMunicipalities(ctx context.Context, areaName string, year int) ([]ports.MunicipalityAreaDataPoint, error) {
	return s.repo.GetAreaAcrossMunicipalities(ctx, areaName, year)
}

func (s *Service) UpsertRegionBudgetSnapshots(ctx context.Context, snapshots []ports.RegionBudgetSnapshot) (int, error) {
	return s.repo.UpsertRegionBudgetSnapshots(ctx, snapshots)
}

func (s *Service) GetMunicipalityKPIs(ctx context.Context, munCode string) ([]ports.KPIValue, error) {
	return s.kolada.FetchKPIs(ctx, munCode, defaultKPIs, rollingYears(5))
}

func (s *Service) GetRegionKPIs(ctx context.Context, regionCode string) ([]ports.KPIValue, error) {
	return s.kolada.FetchKPIs(ctx, koladaRegionCode(regionCode), regionStripKPIs, rollingYears(5))
}

func (s *Service) GetPopulationTrend(ctx context.Context, munCode string) ([]ports.PopulationEntry, error) {
	return s.scb.FetchPopulationTrend(ctx, munCode, rollingYears(5))
}

func (s *Service) GetMunicipalitySpending(ctx context.Context, munCode string) ([]ports.KPIValue, error) {
	return s.kolada.FetchKPIs(ctx, munCode, spendingKPIs, rollingYears(5))
}

// GetMunicipalityBudgetMultiYear fetches Kolada spending KPIs for multiple years,
// multiplies kr/inv by population to derive mnkr totals, and returns snapshots.
// NOTE: population is the municipality's current figure, applied uniformly across all
// years. Year-specific populations are not fetched. Year-over-year deltas are directionally
// correct but slightly imprecise for years where population differed significantly.
func (s *Service) GetMunicipalityBudgetMultiYear(ctx context.Context, munCode string, population int, years []int) ([]ports.MunicipalityBudgetSnapshot, error) {
	if population <= 0 {
		return nil, fmt.Errorf("GetMunicipalityBudgetMultiYear: invalid population %d for %s", population, munCode)
	}

	kpis, err := s.kolada.FetchKPIs(ctx, munCode, spendingKPIs, years)
	if err != nil {
		return nil, err
	}

	// Group by year → kpiCode → value.
	byYear := map[int]map[string]float64{}
	for _, k := range kpis {
		if byYear[k.Year] == nil {
			byYear[k.Year] = map[string]float64{}
		}
		byYear[k.Year][k.KPI] = k.Value
	}

	var snapshots []ports.MunicipalityBudgetSnapshot
	for _, year := range years {
		kpiMap, ok := byYear[year]
		if !ok {
			continue
		}
		var totalMnkr float64
		type pair struct{ code, name string }
		var areas []pair
		for _, code := range spendingKPIs {
			if v, ok := kpiMap[code]; ok && v > 0 {
				areas = append(areas, pair{code, spendingKPINames[code]})
				totalMnkr += (v * float64(population)) / 1_000_000
			}
		}
		if totalMnkr == 0 {
			continue
		}
		// Pct is computed from unrounded totalMnkr for accuracy.
		// Stored TotalMnkr is rounded for display; do not recompute Pct from stored fields.
		for _, a := range areas {
			v := kpiMap[a.code]
			valueMnkr := (v * float64(population)) / 1_000_000
			pct := math.Round((valueMnkr/totalMnkr)*1000) / 10
			snapshots = append(snapshots, ports.MunicipalityBudgetSnapshot{
				MunCode:   munCode,
				AreaName:  a.name,
				Year:      year,
				ValueMnkr: math.Round(valueMnkr*10) / 10,
				TotalMnkr: math.Round(totalMnkr*10) / 10,
				Pct:       pct,
			})
		}
	}
	return snapshots, nil
}

func (s *Service) UpsertMunicipalityBudgetSnapshots(ctx context.Context, snapshots []ports.MunicipalityBudgetSnapshot) (int, error) {
	return s.repo.UpsertMunicipalityBudgetSnapshots(ctx, snapshots)
}

func (s *Service) GetMunicipalityBudgetHistory(ctx context.Context, munCode string, years []int) ([]ports.MunicipalityBudgetSnapshot, error) {
	return s.repo.GetMunicipalityBudgetHistory(ctx, munCode, years)
}

func (s *Service) GetMunicipalityProcurement(ctx context.Context, munCode string) ([]ports.ProcurementCategorySummary, error) {
	mun, err := s.repo.GetMunicipality(ctx, munCode)
	if err != nil {
		return nil, err
	}

	notices, err := s.ted.FetchProcurement(ctx, mun.Name)
	if err != nil {
		return nil, err
	}

	// Aggregate by CPV division.
	type agg struct {
		value float64
		count int
	}
	byDiv := map[string]*agg{}
	var totalSEK float64
	for _, n := range notices {
		// Only use SEK notices for now; skip EUR (rare for Swedish municipalities).
		if n.Currency != "SEK" {
			continue
		}
		if byDiv[n.CPVDivision] == nil {
			byDiv[n.CPVDivision] = &agg{}
		}
		byDiv[n.CPVDivision].value += n.Value
		byDiv[n.CPVDivision].count++
		totalSEK += n.Value
	}

	if totalSEK == 0 {
		return []ports.ProcurementCategorySummary{}, nil
	}

	// Build result slice, sorted by value descending.
	result := make([]ports.ProcurementCategorySummary, 0, len(byDiv))
	for div, a := range byDiv {
		label := cpvDivisionLabel(div)
		result = append(result, ports.ProcurementCategorySummary{
			CPVDivision:   div,
			Label:         label,
			TotalValueSEK: a.value,
			Pct:           (a.value / totalSEK) * 100,
			Count:         a.count,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalValueSEK > result[j].TotalValueSEK
	})
	// Cap at top 10 categories.
	if len(result) > 10 {
		result = result[:10]
	}
	return result, nil
}

// cpvDivisionLabel returns a Swedish label for a 2-digit EU CPV division code.
// Source: Official EU Common Procurement Vocabulary (CPV) 2008, OJ 2007 L 74
// Reference: https://simap.ted.europa.eu/cpv
func cpvDivisionLabel(div string) string {
	labels := map[string]string{
		"03": "Jordbruk & naturresurser",
		"09": "Bränsle & energi",
		"14": "Gruvdrift",
		"15": "Livsmedel",
		"18": "Kläder & textil",
		"22": "Trycksaker",
		"24": "Kemikalier",
		"30": "Kontorsutrustning & IT-hårdvara",
		"31": "Elektrisk utrustning",
		"32": "Radio- & kommunikationsutrustning",
		"33": "Medicinsk utrustning",
		"34": "Transportutrustning",
		"35": "Säkerhet & försvar",
		"37": "Sport & fritid",
		"38": "Laboratorieutrustning",
		"39": "Möbler & inredning",
		"41": "Vatten",
		"42": "Industriella maskiner",
		"43": "Gruvmaskiner",
		"44": "Byggnadsstruktur",
		"45": "Bygg & anläggning",
		"48": "IT-programvara",
		"50": "Reparation & underhåll",
		"51": "Installationstjänster",
		"55": "Hotell & restaurang",
		"60": "Transporttjänster",
		"63": "Stödtjänster transport",
		"64": "Post & telekommunikation",
		"65": "Vatten & energiförsörjning",
		"66": "Finansiella tjänster",
		"70": "Fastigheter",
		"71": "Arkitektur & ingenjörstjänster",
		"72": "IT-tjänster",
		"73": "Forskning & utveckling",
		"75": "Offentlig förvaltning",
		"77": "Skogs- & trädgårdstjänster",
		"79": "Affärstjänster",
		"80": "Utbildning",
		"85": "Hälsa & omsorg",
		"90": "Miljötjänster",
		"92": "Kultur & fritidstjänster",
		"98": "Övriga samhällstjänster",
	}
	if l, ok := labels[div]; ok {
		return l
	}
	return "CPV " + div + "xxx"
}
