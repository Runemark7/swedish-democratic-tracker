package ingestion

import (
	"context"
	"fmt"
	"time"
)

// How long a gap may be before the daily chain logs it at error.
//
// ContentStaleAfter is a backstop, not a recess calendar. A Swedish summer
// break can outlast it, and the log will say so; the external check compares
// our dates with Riksdagen's, which stays quiet when both are old. Sixty days
// is long enough to cover the shorter recesses and short enough that the
// 2026 speech freeze (newest speech 13 August, still stale in late October)
// would have been logged while it was happening.
//
// Job thresholds are about the chain, not the chamber. A daily worker that
// has not succeeded in 36 hours missed more than one night. A weekly worker
// gets nine days so a single Sunday failure is the error log from that run,
// and a second missed week is this alert.
const (
	ContentStaleAfter   = 60 * 24 * time.Hour
	DailyJobStaleAfter  = 36 * time.Hour
	WeeklyJobStaleAfter = 9 * 24 * time.Hour
)

// ScheduledJob is one worker in the chain, in dependency order.
type ScheduledJob struct {
	Name string
	Spec string // "@daily" or "@weekly"
}

// JobFreshness is the latest recorded outcome of one worker.
type JobFreshness struct {
	Name           string
	Spec           string
	LastSuccessAt  *time.Time
	LastFinishedAt *time.Time
	LastStatus     string
}

// Snapshot is the dates and job outcomes a freshness check reads.
type Snapshot struct {
	SpeechesLatest *time.Time
	VotesLatest    *time.Time
	Jobs           []JobFreshness
}

// FreshnessSource reads the snapshot from the database.
type FreshnessSource interface {
	Snapshot(ctx context.Context) (Snapshot, error)
}

// MergeCatalog overlays database rows onto the workers we expect to run.
// A worker that has never been recorded stays in the list with empty times,
// which is the signal that it has never succeeded. Rows for workers no longer
// registered are kept after the catalog.
func MergeCatalog(catalog []ScheduledJob, rows []JobFreshness) []JobFreshness {
	byName := make(map[string]JobFreshness, len(rows))
	for _, row := range rows {
		byName[row.Name] = row
	}
	out := make([]JobFreshness, 0, len(catalog)+len(rows))
	seen := map[string]bool{}
	for _, job := range catalog {
		row := byName[job.Name]
		row.Name = job.Name
		row.Spec = job.Spec
		out = append(out, row)
		seen[job.Name] = true
	}
	for _, row := range rows {
		if !seen[row.Name] {
			out = append(out, row)
		}
	}
	return out
}

// Alerts is the list of staleness conditions, in a stable order: speeches,
// votes, then jobs in catalog order. An empty list means nothing crossed a
// threshold. The strings are the error-log text.
func Alerts(snap Snapshot, now time.Time) []string {
	var out []string
	if msg := contentAlert("speeches", snap.SpeechesLatest, now); msg != "" {
		out = append(out, msg)
	}
	if msg := contentAlert("votes", snap.VotesLatest, now); msg != "" {
		out = append(out, msg)
	}
	for _, job := range snap.Jobs {
		if msg := jobAlert(job, now); msg != "" {
			out = append(out, msg)
		}
	}
	return out
}

func contentAlert(kind string, latest *time.Time, now time.Time) string {
	if latest == nil {
		return kind + ": no records ingested"
	}
	age := now.Sub(latest.UTC())
	if age <= ContentStaleAfter {
		return ""
	}
	return fmt.Sprintf("%s: newest record %s is %s old (threshold %s)",
		kind, latest.UTC().Format("2006-01-02"), formatAge(age), formatAge(ContentStaleAfter))
}

func jobAlert(job JobFreshness, now time.Time) string {
	limit := DailyJobStaleAfter
	schedule := "daily"
	if job.Spec == "@weekly" {
		limit = WeeklyJobStaleAfter
		schedule = "weekly"
	}
	if job.LastSuccessAt == nil {
		return fmt.Sprintf("%s: has never succeeded", job.Name)
	}
	age := now.Sub(job.LastSuccessAt.UTC())
	if age <= limit {
		return ""
	}
	return fmt.Sprintf("%s: last success %s is %s old (%s threshold %s)",
		job.Name, job.LastSuccessAt.UTC().Format(time.RFC3339), formatAge(age), schedule, formatAge(limit))
}

func formatAge(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d <= 48*time.Hour {
		return fmt.Sprintf("%.0fh", d.Hours())
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

// ScheduleLabel is the public word for a cron spec.
func ScheduleLabel(spec string) string {
	if spec == "@weekly" {
		return "weekly"
	}
	return "daily"
}
