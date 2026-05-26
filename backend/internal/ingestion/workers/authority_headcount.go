package workers

import (
	"context"
	"log/slog"

	"riksdagskollen/internal/riksdag/ports"
)

type AuthorityHeadcountWorker struct {
	client ports.HeadcountSCBClient
	repo   ports.AuthorityRepository
}

func NewAuthorityHeadcountWorker(client ports.HeadcountSCBClient, repo ports.AuthorityRepository) AuthorityHeadcountWorker {
	return AuthorityHeadcountWorker{client: client, repo: repo}
}

func (w *AuthorityHeadcountWorker) Name() string { return "authority-headcount" }

func (w *AuthorityHeadcountWorker) Run(ctx context.Context) error {
	entries, err := w.client.FetchHeadcounts(ctx)
	if err != nil {
		// Keep last good enrichment — do not touch the table on fetch failure.
		return err
	}

	items := make([]ports.Enrichment, 0, len(entries))
	for _, e := range entries {
		items = append(items, ports.Enrichment{
			OrgNumber:    e.OrgNumber,
			Department:   e.Department,
			HeadcountInt: e.Headcount,
			Year:         e.Year,
			History:      e.History,
		})
	}

	matched, unmatched, err := w.repo.UpdateEnrichment(ctx, items)
	if err != nil {
		return err
	}
	slog.Info("authority-headcount: enrichment applied", "matched", matched, "unmatched", unmatched)
	return nil
}
