package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"riksdagskollen/internal/ingestion"
)

// Handler serves the public freshness report. It does not include worker
// error text: that can carry upstream response bodies.
type Handler struct {
	src  ingestion.FreshnessSource
	jobs []ingestion.ScheduledJob
	now  func() time.Time
}

func NewHandler(src ingestion.FreshnessSource, jobs []ingestion.ScheduledJob) *Handler {
	return &Handler{src: src, jobs: jobs, now: func() time.Time { return time.Now().UTC() }}
}

func (h *Handler) ServeFreshness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	snap, err := h.src.Snapshot(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "freshness unavailable"})
		return
	}
	now := h.now()
	snap.Jobs = ingestion.MergeCatalog(h.jobs, snap.Jobs)
	alerts := ingestion.Alerts(snap, now)
	if alerts == nil {
		alerts = []string{}
	}

	jobs := make([]jobJSON, 0, len(snap.Jobs))
	for _, job := range snap.Jobs {
		jobs = append(jobs, jobJSON{
			Name:           job.Name,
			Schedule:       ingestion.ScheduleLabel(job.Spec),
			LastSuccessAt:  formatTime(job.LastSuccessAt),
			LastFinishedAt: formatTime(job.LastFinishedAt),
			LastStatus:     emptyAsNull(job.LastStatus),
		})
	}
	resp := freshnessJSON{
		GeneratedAt: now.Format(time.RFC3339),
		Speeches:    seriesJSON{LatestDate: formatDate(snap.SpeechesLatest)},
		Votes: seriesJSON{
			LatestDate: formatDate(snap.VotesLatest),
			LatestAt:   formatTime(snap.VotesLatest),
		},
		Jobs:   jobs,
		Alerts: alerts,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

type freshnessJSON struct {
	GeneratedAt string     `json:"generatedAt"`
	Speeches    seriesJSON `json:"speeches"`
	Votes       seriesJSON `json:"votes"`
	Jobs        []jobJSON  `json:"jobs"`
	Alerts      []string   `json:"alerts"`
}

type seriesJSON struct {
	LatestDate *string `json:"latestDate"`
	LatestAt   *string `json:"latestAt,omitempty"`
}

type jobJSON struct {
	Name           string  `json:"name"`
	Schedule       string  `json:"schedule"`
	LastSuccessAt  *string `json:"lastSuccessAt"`
	LastFinishedAt *string `json:"lastFinishedAt"`
	LastStatus     *string `json:"lastStatus"`
}

func formatTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

func formatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format("2006-01-02")
	return &s
}

func emptyAsNull(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
