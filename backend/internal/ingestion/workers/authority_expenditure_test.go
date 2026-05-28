package workers_test

import (
	"context"
	"errors"
	"testing"

	"riksdagskollen/internal/ingestion/workers"
	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
)

type mockAuthorityFetchClient struct {
	data []ports.AuthorityData
	err  error
}

func (m *mockAuthorityFetchClient) FetchAuthorities(_ context.Context) ([]ports.AuthorityData, error) {
	return m.data, m.err
}

type mockExpenditureRepo struct {
	captured  []ports.ExpenditureUpdate
	matched   int
	updateErr error
}

func (m *mockExpenditureRepo) UpsertAuthorities(_ context.Context, _ []domain.RegisteredAuthority) (int, error) {
	return 0, nil
}
func (m *mockExpenditureRepo) List(_ context.Context, _ ports.AuthorityFilter) ([]domain.RegisteredAuthority, error) {
	return nil, nil
}
func (m *mockExpenditureRepo) Count(_ context.Context, _ ports.AuthorityFilter) (int, error) {
	return 0, nil
}
func (m *mockExpenditureRepo) UpdateEnrichment(_ context.Context, _ []ports.Enrichment) (int, int, error) {
	return 0, 0, nil
}
func (m *mockExpenditureRepo) UpdateExpenditure(_ context.Context, items []ports.ExpenditureUpdate) (int, int, error) {
	if m.updateErr != nil {
		return 0, 0, m.updateErr
	}
	m.captured = append(m.captured, items...)
	return m.matched, len(items) - m.matched, nil
}

func TestAuthorityExpenditureWorker_MapsKnownAndSkipsUnknown(t *testing.T) {
	client := &mockAuthorityFetchClient{data: []ports.AuthorityData{
		{Name: "Polismyndigheten", ExpenditureMdkr: 44.8, BudgetMdkr: 45.0, Year: 2024, History: []ports.YearlyExpenditure{{Year: 2023, ExpenditureMdkr: 42.0}, {Year: 2024, ExpenditureMdkr: 44.8}}},
		{Name: "Helt okänd myndighet", ExpenditureMdkr: 1.0, Year: 2024},
	}}
	repo := &mockExpenditureRepo{matched: 1}
	w := workers.NewAuthorityExpenditureWorker(client, repo)

	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.captured) != 1 {
		t.Fatalf("expected 1 mapped, got %d", len(repo.captured))
	}
	got := repo.captured[0]
	if got.OrgNumber != "202100-0076" || got.ExpenditureMdkr != 44.8 || got.Year != 2024 {
		t.Errorf("mapped wrong: %+v", got)
	}
	if len(got.History) != 2 {
		t.Errorf("history not propagated: %+v", got.History)
	}
}

func TestAuthorityExpenditureWorker_FetchErrorNoUpdate(t *testing.T) {
	client := &mockAuthorityFetchClient{err: errors.New("statskontoret down")}
	repo := &mockExpenditureRepo{}
	w := workers.NewAuthorityExpenditureWorker(client, repo)

	if err := w.Run(context.Background()); err == nil {
		t.Error("expected error when fetch fails")
	}
	if len(repo.captured) != 0 {
		t.Error("must not update on fetch failure")
	}
}
