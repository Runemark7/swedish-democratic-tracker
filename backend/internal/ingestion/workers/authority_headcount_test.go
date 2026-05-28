package workers_test

import (
	"context"
	"errors"
	"testing"

	"riksdagskollen/internal/ingestion/workers"
	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
)

type mockHeadcountClient struct {
	entries []ports.HeadcountEntry
	err     error
}

func (m *mockHeadcountClient) FetchHeadcounts(_ context.Context) ([]ports.HeadcountEntry, error) {
	return m.entries, m.err
}

type mockEnrichRepo struct {
	captured  []ports.Enrichment
	matched   int
	updateErr error
}

func (m *mockEnrichRepo) UpsertAuthorities(_ context.Context, _ []domain.RegisteredAuthority) (int, error) {
	return 0, nil
}
func (m *mockEnrichRepo) List(_ context.Context, _ ports.AuthorityFilter) ([]domain.RegisteredAuthority, error) {
	return nil, nil
}
func (m *mockEnrichRepo) Count(_ context.Context, _ ports.AuthorityFilter) (int, error) {
	return 0, nil
}
func (m *mockEnrichRepo) UpdateEnrichment(_ context.Context, items []ports.Enrichment) (int, int, error) {
	if m.updateErr != nil {
		return 0, 0, m.updateErr
	}
	m.captured = append(m.captured, items...)
	return m.matched, len(items) - m.matched, nil
}
func (m *mockEnrichRepo) UpdateExpenditure(_ context.Context, _ []ports.ExpenditureUpdate) (int, int, error) {
	return 0, 0, nil
}

func TestAuthorityHeadcountWorker_MapsAndUpdates(t *testing.T) {
	client := &mockHeadcountClient{entries: []ports.HeadcountEntry{
		{OrgNumber: "202100-2114", Name: "Arbetsförmedlingen", Department: "Arbetsmarknadsdepartementet", Year: 2024, Headcount: 9400, History: []ports.YearlyHeadcountSCB{{2023, 9300}, {2024, 9400}}},
	}}
	repo := &mockEnrichRepo{matched: 1}
	w := workers.NewAuthorityHeadcountWorker(client, repo)

	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.captured) != 1 {
		t.Fatalf("expected 1 enrichment, got %d", len(repo.captured))
	}
	got := repo.captured[0]
	if got.OrgNumber != "202100-2114" || got.HeadcountInt != 9400 || got.Department != "Arbetsmarknadsdepartementet" || got.Year != 2024 {
		t.Errorf("payload mapped wrong: %+v", got)
	}
	if len(got.History) != 2 {
		t.Errorf("history not propagated: %+v", got.History)
	}
}

func TestAuthorityHeadcountWorker_FetchErrorNoUpdate(t *testing.T) {
	client := &mockHeadcountClient{err: errors.New("xlsx down")}
	repo := &mockEnrichRepo{}
	w := workers.NewAuthorityHeadcountWorker(client, repo)

	if err := w.Run(context.Background()); err == nil {
		t.Error("expected error when fetch fails")
	}
	if len(repo.captured) != 0 {
		t.Error("must not update on fetch failure (keep last good data)")
	}
}
