package regions

import (
	"context"
	"sort"

	"riksdagskollen/internal/regions/domain"
	"riksdagskollen/internal/regions/ports"
)

var defaultKPIs = []string{
	// Verksamhet
	"N00900", "N11037", "N15027", "N20043", "N03010",
	// Budget & ekonomi
	"N03007", "N03102", "N03106", "N03040", "N03132",
}
var defaultKPIYears = []int{2019, 2020, 2021, 2022, 2023}

// regionStripKPIs are the three KPIs shown in the region header strip.
// N63007 = Soliditet region (%), N63016 = Resultat/skatt region (%), N60008 = Nettokostnad/inv (kr).
// Region KPIs use the N6xxxx namespace (mun_type "L" in Kolada), not N0xxxx (which is municipality-only).
var regionStripKPIs = []string{"N60008", "N63016", "N63007"}
var regionKPIYears  = []int{2020, 2021, 2022, 2023}

// koladaRegionCode converts a Swedish 2-digit county code (e.g. "09") to the
// 4-digit zero-prefixed code Kolada expects (e.g. "0009").
func koladaRegionCode(code string) string {
	if len(code) >= 4 {
		return code
	}
	return "00" + code
}
var defaultPopYears = []int{2019, 2020, 2021, 2022, 2023}

var spendingKPIs  = []string{"N11004", "N15028", "N17014", "N20014", "N30005", "N07037", "N09022", "N05011"}
var spendingYears = []int{2019, 2020, 2021, 2022, 2023}

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

func (s *Service) GetMunicipalityKPIs(ctx context.Context, munCode string) ([]ports.KPIValue, error) {
	return s.kolada.FetchKPIs(ctx, munCode, defaultKPIs, defaultKPIYears)
}

func (s *Service) GetRegionKPIs(ctx context.Context, regionCode string) ([]ports.KPIValue, error) {
	return s.kolada.FetchKPIs(ctx, koladaRegionCode(regionCode), regionStripKPIs, regionKPIYears)
}

func (s *Service) GetPopulationTrend(ctx context.Context, munCode string) ([]ports.PopulationEntry, error) {
	return s.scb.FetchPopulationTrend(ctx, munCode, defaultPopYears)
}

func (s *Service) GetMunicipalitySpending(ctx context.Context, munCode string) ([]ports.KPIValue, error) {
	return s.kolada.FetchKPIs(ctx, munCode, spendingKPIs, spendingYears)
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
