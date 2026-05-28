package workers

import (
	"context"
	"log/slog"
	"sort"
	"strings"

	"riksdagskollen/internal/riksdag/ports"
)

// AuthorityExpenditureWorker attributes Statskontoret årsutfall outcomes to
// agencies in the authorities table by exact-name matching the CSV's
// `Anslagsnamn` against the registered agency name. Anslag whose name does
// not match any registered agency (topic anslag, sub-posts, etc.) are
// ignored.
type AuthorityExpenditureWorker struct {
	client ports.AllAnslagClient
	repo   ports.AuthorityRepository
}

func NewAuthorityExpenditureWorker(client ports.AllAnslagClient, repo ports.AuthorityRepository) AuthorityExpenditureWorker {
	return AuthorityExpenditureWorker{client: client, repo: repo}
}

func (w *AuthorityExpenditureWorker) Name() string { return "authority-expenditure" }

func (w *AuthorityExpenditureWorker) Run(ctx context.Context) error {
	rows, err := w.client.FetchAllAnslagYearly(ctx)
	if err != nil {
		// Keep last good expenditure — do not touch the table on fetch failure.
		return err
	}

	authorities, err := w.repo.List(ctx, ports.AuthorityFilter{Limit: 10000})
	if err != nil {
		return err
	}
	nameToOrg := make(map[string]string, len(authorities))
	for _, a := range authorities {
		key := normName(a.Name)
		if key == "" {
			continue
		}
		// First entry wins on conflicts (rare; logged for visibility).
		if _, exists := nameToOrg[key]; !exists {
			nameToOrg[key] = a.OrgNumber
		}
	}

	// Accumulate per (org_number, year). One agency may have multiple anslag
	// matching its name in the same year — sum them.
	type yearAcc struct{ utfall, budget float64 }
	byOrg := map[string]map[int]*yearAcc{}
	matchedRows := 0
	for _, r := range rows {
		org, ok := nameToOrg[normName(r.AnslagName)]
		if !ok {
			continue
		}
		matchedRows++
		yrs := byOrg[org]
		if yrs == nil {
			yrs = map[int]*yearAcc{}
			byOrg[org] = yrs
		}
		a := yrs[r.Year]
		if a == nil {
			a = &yearAcc{}
			yrs[r.Year] = a
		}
		a.utfall += r.ExpenditureMdkr
		a.budget += r.BudgetMdkr
	}

	items := make([]ports.ExpenditureUpdate, 0, len(byOrg))
	for org, yrs := range byOrg {
		years := make([]int, 0, len(yrs))
		for y := range yrs {
			years = append(years, y)
		}
		sort.Ints(years)
		history := make([]ports.YearlyExpenditureSCB, len(years))
		for i, y := range years {
			a := yrs[y]
			history[i] = ports.YearlyExpenditureSCB{Year: y, ExpenditureMdkr: a.utfall, BudgetMdkr: a.budget}
		}
		latest := history[len(history)-1]
		items = append(items, ports.ExpenditureUpdate{
			OrgNumber:       org,
			ExpenditureMdkr: latest.ExpenditureMdkr,
			BudgetMdkr:      latest.BudgetMdkr,
			Year:            latest.Year,
			History:         history,
		})
	}

	matched, unmatched, err := w.repo.UpdateExpenditure(ctx, items)
	if err != nil {
		return err
	}
	slog.Info("authority-expenditure: applied",
		"matched", matched, "unmatched", unmatched,
		"agencies_with_anslag", len(byOrg), "anslag_rows_matched", matchedRows,
		"anslag_rows_total", len(rows))
	return nil
}

// normName lower-cases + folds Swedish vowels for tolerant exact match.
// Trims whitespace. Returns "" for empty.
func normName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	r := strings.NewReplacer(
		"å", "a", "Å", "a",
		"ä", "a", "Ä", "a",
		"ö", "o", "Ö", "o",
		"é", "e", "É", "e",
	)
	return strings.ToLower(r.Replace(s))
}
