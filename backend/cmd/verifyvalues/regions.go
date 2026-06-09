package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"riksdagskollen/internal/regions"
	"riksdagskollen/internal/regions/adapters/kolada"
	"riksdagskollen/internal/regions/seeder"
)

// KPI codes mirror the strips the UI renders (kept in sync with
// backend/internal/regions/service.go).
var verifyRegionKPIs = []string{"N60008", "N63016", "N63007", "N79173", "N79179", "N60404", "N85012"}
var verifyKommunKPIs = []string{"N00900", "N03102", "N03106", "N15428", "N00708"}
var verifySpendingKPIs = []string{"N11004", "N15028", "N17014", "N20014", "N30005", "N07037", "N09022", "N05011", "N45014"}

// popYear is the year our seeded population figures represent (the startup
// seeder fetches SCB BE0101 BefolkningNy for Tid 2023). Population is verified
// against that same year so we compare like-for-like instead of flagging
// natural year-over-year drift.
const popYear = 2023

// kpiYear is the latest full year Kolada reliably exposes for the KPI strips.
const kpiYear = 2024

// regionAndKommunChecks compares DB values to live SCB/Kolada values.
func regionAndKommunChecks(ctx context.Context, svc *regions.Service, koladaCl *kolada.Client) []Result {
	var out []Result
	httpCl := &http.Client{Timeout: 60 * time.Second}

	munMandates, regMandates, mErr := seeder.FetchMandateTotals(ctx, httpCl)

	regionsList, rErr := svc.ListRegions(ctx)
	munList, mListErr := svc.ListMunicipalities(ctx, "")

	// Batch-fetch municipality populations in one shot (per-entity requests
	// would be hundreds of sequential SCB calls). Only 4-digit municipality
	// codes are valid in this SCB table; region population is migration-derived
	// (the seeder does not set it) so it is not verified here.
	regionCodes := make(map[string]bool)
	munCodes := make(map[string]bool)
	var munCodeList []string
	if rErr == nil {
		for _, r := range regionsList {
			regionCodes[r.Code] = true
		}
	}
	if mListErr == nil {
		for _, m := range munList {
			munCodes[m.Code] = true
			munCodeList = append(munCodeList, m.Code)
		}
	}
	popByCode, popErr := seeder.FetchPopulations(ctx, httpCl, munCodeList)

	// Regions: mandate only.
	if rErr == nil {
		for _, r := range regionsList {
			entity := fmt.Sprintf("%s (%s)", r.Name, r.Code)
			out = append(out, mandateCheck("region", entity, r.Name, r.TotalMandates, regMandates, mErr))
		}
	} else {
		out = append(out, Result{Tier: "region", Entity: "(all)", Metric: "list", SourceID: "scb-ltmandat", Status: StatusError, Upstream: rErr.Error()})
	}

	// Kommuner: population + mandate.
	if mListErr == nil {
		for _, m := range munList {
			entity := fmt.Sprintf("%s (%s)", m.Name, m.Code)
			out = append(out, populationCheck("kommun", entity, m.Code, m.Population, popByCode, popErr))
			out = append(out, mandateCheck("kommun", entity, m.Name, m.TotalMandates, munMandates, mErr))
		}
	} else {
		out = append(out, Result{Tier: "kommun", Entity: "(all)", Metric: "list", SourceID: "scb-kfmandat", Status: StatusError, Upstream: mListErr.Error()})
	}

	// KPI / spending. Kolada's batch endpoint returns municipality-level rows
	// AND sub-municipal operating units (G-prefixed ou ids); we compare only
	// the codes that are real entities in our DB so the report stays meaningful.
	out = append(out, kpiChecks(ctx, "kommun", append(append([]string{}, verifyKommunKPIs...), verifySpendingKPIs...), kpiYear, svc, koladaCl, false, munCodes)...)
	out = append(out, kpiChecks(ctx, "region", verifyRegionKPIs, kpiYear, svc, koladaCl, true, regionCodes)...)

	return out
}

func populationCheck(tier, entity, code string, ours int, popByCode map[string]int, fetchErr error) Result {
	r := Result{Tier: tier, Entity: entity, Metric: fmt.Sprintf("population %d", popYear), SourceID: "scb-befolkning", Ours: strconv.Itoa(ours)}
	if fetchErr != nil {
		r.Status = StatusError
		r.Upstream = fetchErr.Error()
		return r
	}
	up, ok := popByCode[code]
	if !ok {
		r.Status = StatusError
		r.Upstream = "code not in SCB population set"
		return r
	}
	r.Upstream = strconv.Itoa(up)
	r.Status = classify(r.Ours, r.Upstream, false)
	return r
}

func mandateCheck(tier, entity, name string, ours int, byName map[string]int, fetchErr error) Result {
	src := "scb-ltmandat"
	if tier == "kommun" {
		src = "scb-kfmandat"
	}
	r := Result{Tier: tier, Entity: entity, Metric: "mandate total", SourceID: src, Ours: strconv.Itoa(ours)}
	if fetchErr != nil {
		r.Status = StatusError
		r.Upstream = fetchErr.Error()
		return r
	}
	up, ok := byName[name]
	if !ok {
		r.Status = StatusError
		r.Upstream = "name not found in SCB mandate set"
		return r
	}
	r.Upstream = strconv.Itoa(up)
	r.Status = classify(r.Ours, r.Upstream, false)
	return r
}

// kpiChecks compares each entity's KPI value (as our service serves it) against
// Kolada's batch result. isRegion selects GetRegionKPIs vs GetMunicipalityKPIs.
// validCodes restricts the comparison to entities that exist in our DB, so the
// G-prefixed sub-municipal units Kolada also returns are ignored.
//
// Upstream is fetched once per KPI (batch, all entities); ours is fetched once
// per entity (all KPIs) — avoiding a fetch per (entity, KPI) pair.
func kpiChecks(ctx context.Context, tier string, kpiCodes []string, year int, svc *regions.Service, koladaCl *kolada.Client, isRegion bool, validCodes map[string]bool) []Result {
	var out []Result

	// Upstream: kpi -> entity code -> value (restricted to our entities).
	upstream := make(map[string]map[string]float64, len(kpiCodes))
	for _, kpi := range kpiCodes {
		rows, err := koladaCl.FetchKPIAllMunicipalities(ctx, kpi, []int{year})
		if err != nil {
			out = append(out, Result{Tier: tier, Entity: "(all)", Metric: "KPI " + kpi, SourceID: "kolada", Status: StatusError, Upstream: err.Error()})
			continue
		}
		byCode := make(map[string]float64)
		for _, u := range rows {
			if validCodes[u.MunCode] {
				byCode[u.MunCode] = u.Value
			}
		}
		upstream[kpi] = byCode
	}

	// Ours: one fetch per entity, then compare each KPI we both have.
	for code := range validCodes {
		oursList, oerr := svc.GetMunicipalityKPIs(ctx, code)
		if isRegion {
			oursList, oerr = svc.GetRegionKPIs(ctx, code)
		}
		if oerr != nil {
			continue
		}
		oursByKPI := make(map[string]float64, len(oursList))
		for _, kv := range oursList {
			oursByKPI[kv.KPI] = kv.Value
		}
		for _, kpi := range kpiCodes {
			upVal, ok := upstream[kpi][code]
			if !ok {
				continue
			}
			oursVal, ok := oursByKPI[kpi]
			if !ok {
				continue
			}
			oursStr := strconv.FormatFloat(oursVal, 'f', -1, 64)
			upStr := strconv.FormatFloat(upVal, 'f', -1, 64)
			// KPIs are not stored facts: our service proxies them live from
			// Kolada at the latest available year per KPI, while this batch
			// pins a fixed year. So a difference reflects year selection, not a
			// data error — recorded as informational, never flagged for review.
			out = append(out, Result{
				Tier:     tier,
				Entity:   code,
				Metric:   fmt.Sprintf("KPI %s %d", kpi, year),
				SourceID: "kolada",
				Ours:     oursStr,
				Upstream: upStr,
				Status:   classify(oursStr, upStr, true),
			})
		}
	}
	return out
}
