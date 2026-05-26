package workers

import (
	"context"
	"log/slog"

	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
)

type AuthoritiesWorker struct {
	client ports.RegisterClient
	repo   ports.AuthorityRepository
}

func NewAuthoritiesWorker(client ports.RegisterClient, repo ports.AuthorityRepository) AuthoritiesWorker {
	return AuthoritiesWorker{client: client, repo: repo}
}

func (w *AuthoritiesWorker) Name() string { return "authorities" }

func (w *AuthoritiesWorker) Run(ctx context.Context) error {
	entries, err := w.client.FetchRegister(ctx)
	if err != nil {
		// Keep last good data — do not touch the table on a fetch failure.
		return err
	}

	items := make([]domain.RegisteredAuthority, 0, len(entries))
	for _, e := range entries {
		items = append(items, domain.RegisteredAuthority{
			OrgNumber:       e.OrgNumber,
			Slug:            e.Slug,
			Name:            e.Name,
			Type:            e.Type,
			PrincipalBody:   e.PrincipalBody,
			UnderGovernment: e.UnderGovernment,
			Website:         e.Website,
			SFS:             e.SFS,
			// ExpenditureMdkr / HeadcountInt intentionally nil in Phase 1.
		})
	}

	n, err := w.repo.UpsertAuthorities(ctx, items)
	if err != nil {
		return err
	}
	slog.Info("authorities: upserted register", "count", n)
	return nil
}
