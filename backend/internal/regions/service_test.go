package regions_test

import (
	"context"
	"slices"
	"testing"

	"riksdagskollen/internal/regions"
	"riksdagskollen/internal/regions/domain"
	"riksdagskollen/internal/regions/ports"
)

// fakeRepo embeds the interface; only ListRegions is implemented — calling
// anything else panics, which is fine for these tests.
type fakeRepo struct {
	ports.RegionRepository
	regions []*domain.Region
}

func (f *fakeRepo) ListRegions(context.Context) ([]*domain.Region, error) {
	return f.regions, nil
}

// fakeKolada records the KPI codes requested via FetchKPIs and serves a fixed
// batch for FetchKPIAllMunicipalities.
type fakeKolada struct {
	requestedKPIs []string
	batch         []ports.KPIValueWithMun
}

func (f *fakeKolada) FetchKPIs(_ context.Context, _ string, kpiCodes []string, _ []int) ([]ports.KPIValue, error) {
	f.requestedKPIs = kpiCodes
	return nil, nil
}

func (f *fakeKolada) FetchKPIAllMunicipalities(_ context.Context, _ string, _ []int) ([]ports.KPIValueWithMun, error) {
	return f.batch, nil
}

func TestGetRegionKPIRanking_ExcludesGotlandForN85012(t *testing.T) {
	repo := &fakeRepo{regions: []*domain.Region{
		{Code: "09", Name: "Region Gotland"},
		{Code: "10", Name: "Region Blekinge"},
	}}
	kolada := &fakeKolada{batch: []ports.KPIValueWithMun{
		// Gotland's regional-development cost is booked municipally; Kolada
		// publishes a structural 0.0 for the region entity every year.
		{MunCode: "0009", KPI: "N85012", Year: 2024, Value: 0.0},
		{MunCode: "0010", KPI: "N85012", Year: 2024, Value: 3428.25},
	}}
	svc := regions.NewService(repo, kolada, nil, nil)

	entries, err := svc.GetRegionKPIRanking(context.Background(), "N85012")
	if err != nil {
		t.Fatalf("GetRegionKPIRanking: %v", err)
	}
	for _, e := range entries {
		if e.RegionCode == "09" {
			t.Fatalf("Gotland must be excluded from N85012 ranking, got %+v", e)
		}
	}
	if len(entries) != 1 || entries[0].RegionCode != "10" || entries[0].Total != 1 {
		t.Fatalf("want only Blekinge with Total=1, got %+v", entries)
	}
}

func TestGetRegionKPIRanking_KeepsGotlandForOtherKPIs(t *testing.T) {
	repo := &fakeRepo{regions: []*domain.Region{
		{Code: "09", Name: "Region Gotland"},
	}}
	kolada := &fakeKolada{batch: []ports.KPIValueWithMun{
		{MunCode: "0009", KPI: "N60008", Year: 2024, Value: 41577.14},
	}}
	svc := regions.NewService(repo, kolada, nil, nil)

	entries, err := svc.GetRegionKPIRanking(context.Background(), "N60008")
	if err != nil {
		t.Fatalf("GetRegionKPIRanking: %v", err)
	}
	if len(entries) != 1 || entries[0].RegionCode != "09" {
		t.Fatalf("Gotland must stay in other KPI rankings, got %+v", entries)
	}
}

func TestGetRegionKPIs_FiltersExcludedKPIForGotland(t *testing.T) {
	kolada := &fakeKolada{}
	svc := regions.NewService(&fakeRepo{}, kolada, nil, nil)

	if _, err := svc.GetRegionKPIs(context.Background(), "09"); err != nil {
		t.Fatalf("GetRegionKPIs: %v", err)
	}
	if slices.Contains(kolada.requestedKPIs, "N85012") {
		t.Fatalf("N85012 must not be requested for Gotland, requested: %v", kolada.requestedKPIs)
	}
	if len(kolada.requestedKPIs) == 0 {
		t.Fatal("other strip KPIs must still be requested for Gotland")
	}

	// Any other region requests the full strip.
	if _, err := svc.GetRegionKPIs(context.Background(), "10"); err != nil {
		t.Fatalf("GetRegionKPIs: %v", err)
	}
	if !slices.Contains(kolada.requestedKPIs, "N85012") {
		t.Fatalf("N85012 must be requested for other regions, requested: %v", kolada.requestedKPIs)
	}
}
