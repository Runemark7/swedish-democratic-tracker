package workers

import (
	"context"
	"log/slog"

	"riksdagskollen/internal/riksdag/ports"
)

// nameToOrgNumber maps the curated set of agency names returned by the
// statskontoret/static AuthorityClient to canonical org-numbers in the
// authorities table. Phase 3 MVP covers only this curated set (~10 agencies);
// extending to all ~449 needs a Statsliggaren scrape (anslag→myndighet
// mapping), tracked as a follow-up.
var nameToOrgNumber = map[string]string{
	"Skatteverket":       "202100-5448",
	"Tullverket":         "202100-0969",
	"Polismyndigheten":   "202100-0076",
	"Säkerhetspolisen":   "202100-6594",
	"Åklagarmyndigheten": "202100-0084",
	"Sveriges Domstolar": "cfar:20738233", // Domstolsverket (umbrella org)
	"Kriminalvården":     "202100-0225",
	"Migrationsverket":   "202100-2163",
	"Försäkringskassan":  "202100-5521",
	"Arbetsförmedlingen": "202100-2114",
}

type AuthorityExpenditureWorker struct {
	client ports.AuthorityClient
	repo   ports.AuthorityRepository
}

func NewAuthorityExpenditureWorker(client ports.AuthorityClient, repo ports.AuthorityRepository) AuthorityExpenditureWorker {
	return AuthorityExpenditureWorker{client: client, repo: repo}
}

func (w *AuthorityExpenditureWorker) Name() string { return "authority-expenditure" }

func (w *AuthorityExpenditureWorker) Run(ctx context.Context) error {
	entries, err := w.client.FetchAuthorities(ctx)
	if err != nil {
		// Keep last good expenditure — do not touch the table on fetch failure.
		return err
	}

	items := make([]ports.ExpenditureUpdate, 0, len(entries))
	skipped := 0
	for _, e := range entries {
		org, ok := nameToOrgNumber[e.Name]
		if !ok {
			skipped++
			continue
		}
		history := make([]ports.YearlyExpenditureSCB, len(e.History))
		for i, h := range e.History {
			history[i] = ports.YearlyExpenditureSCB{
				Year:            h.Year,
				ExpenditureMdkr: h.ExpenditureMdkr,
				BudgetMdkr:      h.BudgetMdkr,
			}
		}
		items = append(items, ports.ExpenditureUpdate{
			OrgNumber:       org,
			ExpenditureMdkr: e.ExpenditureMdkr,
			BudgetMdkr:      e.BudgetMdkr,
			Year:            e.Year,
			History:         history,
		})
	}

	matched, unmatched, err := w.repo.UpdateExpenditure(ctx, items)
	if err != nil {
		return err
	}
	slog.Info("authority-expenditure: applied", "matched", matched, "unmatched", unmatched, "skipped_unmapped", skipped)
	return nil
}
