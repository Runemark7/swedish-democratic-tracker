package main

import (
	"fmt"
	"strings"
	"time"
)

// renderReport produces the dated markdown audit document. The "Needs review"
// summary (Status==Review only) goes on top so problems are seen first;
// per-tier tables follow; volatile Info rows are grouped at the end.
func renderReport(results []Result, runAt time.Time) string {
	var b strings.Builder
	day := runAt.Format("2006-01-02")
	fmt.Fprintf(&b, "# Value Verification — %s\n\n", day)
	fmt.Fprintf(&b, "Run: %s UTC\n\n", runAt.Format("2006-01-02 15:04:05"))

	// Needs review summary.
	b.WriteString("## ⚠ Needs review")
	var review []Result
	for _, r := range results {
		if r.Status == StatusReview {
			review = append(review, r)
		}
	}
	fmt.Fprintf(&b, " (%d)\n\n", len(review))
	if len(review) == 0 {
		b.WriteString("None — every checked value matches upstream.\n\n")
	} else {
		writeTable(&b, []string{"Tier", "Entity", "Metric", "Ours", "Upstream", "Source"}, review,
			func(r Result) []string { return []string{r.Tier, r.Entity, r.Metric, r.Ours, r.Upstream, r.SourceID} })
	}

	// Per-tier tables (non-Info).
	for _, tier := range []string{"region", "kommun", "riksdag"} {
		var rows []Result
		for _, r := range results {
			if r.Tier == tier && r.Status != StatusInfo {
				rows = append(rows, r)
			}
		}
		if len(rows) == 0 {
			continue
		}
		fmt.Fprintf(&b, "## %s (%d)\n\n", strings.Title(tier), len(rows)) //nolint:staticcheck
		writeTable(&b, []string{"Entity", "Metric", "Ours", "Upstream", "Status", "Source"}, rows,
			func(r Result) []string { return []string{r.Entity, r.Metric, r.Ours, r.Upstream, string(r.Status), r.SourceID} })
	}

	// Informational (volatile).
	var info []Result
	for _, r := range results {
		if r.Status == StatusInfo {
			info = append(info, r)
		}
	}
	if len(info) > 0 {
		b.WriteString("### Informational (volatile)\n\n")
		writeTable(&b, []string{"Entity", "Metric", "Ours", "Upstream", "Source"}, info,
			func(r Result) []string { return []string{r.Entity, r.Metric, r.Ours, r.Upstream, r.SourceID} })
	}

	return b.String()
}

func writeTable(b *strings.Builder, headers []string, rows []Result, cols func(Result) []string) {
	b.WriteString("| " + strings.Join(headers, " | ") + " |\n")
	b.WriteString("|" + strings.Repeat("---|", len(headers)) + "\n")
	for _, r := range rows {
		b.WriteString("| " + strings.Join(cols(r), " | ") + " |\n")
	}
	b.WriteString("\n")
}
