package workers

import (
	"context"
	"log/slog"
	"time"

	riksdagPorts "riksdagskollen/internal/riksdag/ports"
)

type AgencyInfo struct {
	Name string
	Slug string
}

type AgencyIntelWorker struct {
	client   riksdagPorts.AgencyDocClient
	repo     riksdagPorts.AgencyIntelRepository
	agencies []AgencyInfo
}

func NewAgencyIntelWorker(
	client riksdagPorts.AgencyDocClient,
	repo riksdagPorts.AgencyIntelRepository,
	agencies []AgencyInfo,
) AgencyIntelWorker {
	return AgencyIntelWorker{client: client, repo: repo, agencies: agencies}
}

func (w *AgencyIntelWorker) Name() string { return "agency-intel" }

func (w *AgencyIntelWorker) Run(ctx context.Context) error {
	for i, a := range w.agencies {
		if err := w.syncAgency(ctx, a); err != nil {
			slog.Warn("agency-intel: sync failed, continuing", "agency", a.Name, "error", err)
		}
		if i < len(w.agencies)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	return nil
}

func (w *AgencyIntelWorker) syncAgency(ctx context.Context, a AgencyInfo) error {
	dec, err := w.client.FetchDecisions(ctx, a.Name)
	if err != nil {
		slog.Warn("agency-intel: fetch decisions failed", "agency", a.Name, "error", err)
	} else if len(dec) > 0 {
		if err := w.repo.UpsertDecisions(ctx, a.Slug, dec); err != nil {
			return err
		}
		slog.Info("agency-intel: upserted decisions", "agency", a.Name, "count", len(dec))
	}

	return nil
}
