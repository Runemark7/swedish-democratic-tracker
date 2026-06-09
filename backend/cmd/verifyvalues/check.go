package main

// Status is the verdict for one compared value.
type Status string

const (
	StatusMatch  Status = "match"  // ours == upstream
	StatusReview Status = "review" // ours != upstream, needs human eyes
	StatusInfo   Status = "info"   // volatile metric, recorded but never flagged
	StatusError  Status = "error"  // upstream fetch failed
)

// Result is one (entity, metric) comparison row in the report.
type Result struct {
	Tier     string // "region" | "kommun" | "riksdag"
	Entity   string // "Dalarna (20)" | "Stockholm (0180)" | "Riksdag"
	Metric   string // "population 2024" | "mandate total" | "KPI N60008 2023"
	SourceID string // registry id, e.g. "scb-befolkning"
	Ours     string
	Upstream string
	Status   Status
}

// classify compares two canonical string values. Volatile metrics (e.g.
// speech counts) are always Info: their delta is recorded but never flagged.
func classify(ours, upstream string, volatile bool) Status {
	if volatile {
		return StatusInfo
	}
	if ours == upstream {
		return StatusMatch
	}
	return StatusReview
}
