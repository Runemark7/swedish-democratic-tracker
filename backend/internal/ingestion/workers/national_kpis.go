package workers

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
)

// NationalKpisWorker keeps the Riksdag-page national KPI strip sourced live
// from SCB (AKU unemployment, KPI inflation) instead of hardcoded snapshots.
type NationalKpisWorker struct {
	client ports.MacroClient
	repo   ports.KpiRepository
}

func NewNationalKpisWorker(client ports.MacroClient, repo ports.KpiRepository) NationalKpisWorker {
	return NationalKpisWorker{client: client, repo: repo}
}

func (w *NationalKpisWorker) Name() string { return "national-kpis" }

func (w *NationalKpisWorker) Run(ctx context.Context) error {
	type spec struct {
		fetch       func(context.Context) ([]ports.MacroPoint, error)
		label       string
		description string
		worseHigher bool
		sourceURL   string
		sortOrder   int
	}
	specs := []spec{
		{
			fetch:       w.client.FetchUnemploymentRate,
			label:       "Arbetslöshet",
			description: "Andel av arbetskraften 15–74 år som är arbetslösa, säsongrensat (AKU).",
			worseHigher: true,
			sourceURL:   "https://www.statistikdatabasen.scb.se/pxweb/sv/ssd/START__AM__AM0401__AM0401A/AKURLBefK/",
			sortOrder:   1,
		},
		{
			fetch:       w.client.FetchInflationRate,
			label:       "Inflation (KPI)",
			description: "Konsumentprisernas förändring de senaste tolv månaderna. Riksbankens inflationsmål är 2 %.",
			worseHigher: true,
			sourceURL:   "https://www.statistikdatabasen.scb.se/pxweb/sv/ssd/START__PR__PR0101__PR0101A/KPI2020M/",
			sortOrder:   2,
		},
	}

	var firstErr error
	for _, s := range specs {
		pts, err := s.fetch(ctx)
		if err != nil {
			// Keep last good data; report the failure.
			slog.Warn("national-kpis: fetch failed", "kpi", s.label, "error", err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		latest := pts[0]
		kpi := domain.Kpi{
			Label:       s.label,
			Description: s.description,
			Raw:         latest.Value,
			WorseHigher: s.worseHigher,
			Unit:        "%",
			Trend:       "flat",
			Delta:       "–",
			Note:        "Källa: SCB · " + formatPeriod(latest.Period),
			SourceURL:   s.sourceURL,
			Year:        periodYear(latest.Period),
		}
		if len(pts) >= 2 {
			kpi.Trend, kpi.Delta = trendAndDelta(latest.Value, pts[1].Value)
		}
		if err := w.repo.UpsertKpi(ctx, kpi, s.sortOrder); err != nil {
			slog.Warn("national-kpis: upsert failed", "kpi", s.label, "error", err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		slog.Info("national-kpis: updated", "kpi", s.label, "value", latest.Value, "period", latest.Period)
	}
	return firstErr
}

// trendAndDelta derives the trend arrow and a Swedish-formatted delta
// ("−0,4 pp") from the latest and previous observation.
func trendAndDelta(latest, prev float64) (trend, delta string) {
	d := latest - prev
	switch {
	case d > 0.04:
		trend = "up"
	case d < -0.04:
		trend = "down"
	default:
		return "flat", "±0,0 pp"
	}
	sign := "+"
	if d < 0 {
		sign = "−"
		d = -d
	}
	return trend, fmt.Sprintf("%s%s pp", sign, strings.ReplaceAll(strconv.FormatFloat(d, 'f', 1, 64), ".", ","))
}

// periodYear extracts the year from an SCB period label ("2026M04", "2026K1").
func periodYear(period string) int {
	if len(period) < 4 {
		return 0
	}
	y, _ := strconv.Atoi(period[:4])
	return y
}

// formatPeriod renders an SCB period label in Swedish ("april 2026",
// "kv 1 2026"). Falls back to the raw label.
func formatPeriod(period string) string {
	months := []string{"januari", "februari", "mars", "april", "maj", "juni",
		"juli", "augusti", "september", "oktober", "november", "december"}
	if i := strings.IndexByte(period, 'M'); i == 4 && len(period) >= 6 {
		if m, err := strconv.Atoi(period[i+1:]); err == nil && m >= 1 && m <= 12 {
			return months[m-1] + " " + period[:4]
		}
	}
	if i := strings.IndexByte(period, 'K'); i == 4 && len(period) >= 6 {
		return "kv " + period[i+1:] + " " + period[:4]
	}
	return period
}
