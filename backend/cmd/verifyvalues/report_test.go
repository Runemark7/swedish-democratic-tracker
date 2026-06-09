package main

import (
	"strings"
	"testing"
	"time"
)

func TestRenderReport(t *testing.T) {
	results := []Result{
		{Tier: "region", Entity: "Dalarna (20)", Metric: "population 2024", SourceID: "scb-befolkning", Ours: "287966", Upstream: "287966", Status: StatusMatch},
		{Tier: "kommun", Entity: "Malmö (1280)", Metric: "population 2024", SourceID: "scb-befolkning", Ours: "357377", Upstream: "357891", Status: StatusReview},
		{Tier: "riksdag", Entity: "Riksdag", Metric: "anföranden total", SourceID: "riksdagen", Ours: "41233", Upstream: "41980", Status: StatusInfo},
	}
	out := renderReport(results, time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC))

	if !strings.Contains(out, "# Value Verification — 2026-06-07") {
		t.Errorf("missing dated title:\n%s", out)
	}
	reviewSection := section(out, "## ⚠ Needs review")
	if !strings.Contains(reviewSection, "Malmö (1280)") {
		t.Errorf("review summary missing Malmö row:\n%s", reviewSection)
	}
	if strings.Contains(reviewSection, "Dalarna") {
		t.Errorf("review summary wrongly includes a Match row:\n%s", reviewSection)
	}
	if strings.Contains(reviewSection, "anföranden") {
		t.Errorf("review summary wrongly includes an Info row:\n%s", reviewSection)
	}
	if !strings.Contains(out, "### Informational") || !strings.Contains(out, "anföranden total") {
		t.Errorf("missing Informational section:\n%s", out)
	}
}

// section returns the substring from the given heading up to the next "## "
// heading (test helper).
func section(doc, heading string) string {
	i := strings.Index(doc, heading)
	if i < 0 {
		return ""
	}
	rest := doc[i+len(heading):]
	if j := strings.Index(rest, "\n## "); j >= 0 {
		return rest[:j]
	}
	return rest
}
