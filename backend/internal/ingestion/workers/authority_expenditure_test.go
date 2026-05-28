package workers_test

import (
	"context"
	"errors"
	"testing"

	"riksdagskollen/internal/ingestion/workers"
	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
)

type mockAllAnslagClient struct {
	rows []ports.AnslagYearly
	err  error
}

func (m *mockAllAnslagClient) FetchAllAnslagYearly(_ context.Context) ([]ports.AnslagYearly, error) {
	return m.rows, m.err
}

type mockExpenditureRepo struct {
	authorities []domain.RegisteredAuthority
	listErr     error
	captured    []ports.ExpenditureUpdate
	matched     int
	updateErr   error
}

func (m *mockExpenditureRepo) UpsertAuthorities(_ context.Context, _ []domain.RegisteredAuthority) (int, error) {
	return 0, nil
}
func (m *mockExpenditureRepo) List(_ context.Context, _ ports.AuthorityFilter) ([]domain.RegisteredAuthority, error) {
	return m.authorities, m.listErr
}
func (m *mockExpenditureRepo) Count(_ context.Context, _ ports.AuthorityFilter) (int, error) {
	return len(m.authorities), nil
}
func (m *mockExpenditureRepo) UpdateEnrichment(_ context.Context, _ []ports.Enrichment) (int, int, error) {
	return 0, 0, nil
}
func (m *mockExpenditureRepo) GetBySlug(_ context.Context, _ string) (*domain.RegisteredAuthority, error) {
	return nil, nil
}
func (m *mockExpenditureRepo) UpdateExpenditure(_ context.Context, items []ports.ExpenditureUpdate) (int, int, error) {
	if m.updateErr != nil {
		return 0, 0, m.updateErr
	}
	m.captured = append(m.captured, items...)
	return m.matched, len(items) - m.matched, nil
}

func TestAuthorityExpenditureWorker_NameMatchAndAggregate(t *testing.T) {
	client := &mockAllAnslagClient{rows: []ports.AnslagYearly{
		{Anslag: "0401001", AnslagName: "Polismyndigheten", Year: 2023, ExpenditureMdkr: 40.0, BudgetMdkr: 41.0},
		{Anslag: "0401001", AnslagName: "Polismyndigheten", Year: 2024, ExpenditureMdkr: 44.8, BudgetMdkr: 45.2},
		// Second matching anslag for same agency same year → sum.
		{Anslag: "0401099", AnslagName: "Polismyndigheten", Year: 2024, ExpenditureMdkr: 0.5, BudgetMdkr: 0.5},
		// Topic anslag — no matching agency, must be skipped.
		{Anslag: "0106001", AnslagName: "Allmänna val och demokrati", Year: 2024, ExpenditureMdkr: 700, BudgetMdkr: 700},
	}}
	repo := &mockExpenditureRepo{
		authorities: []domain.RegisteredAuthority{
			{OrgNumber: "202100-0076", Name: "POLISMYNDIGHETEN"}, // upper-case in register; matcher must fold.
		},
		matched: 1,
	}
	w := workers.NewAuthorityExpenditureWorker(client, repo)

	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.captured) != 1 {
		t.Fatalf("expected 1 ExpenditureUpdate, got %d", len(repo.captured))
	}
	got := repo.captured[0]
	if got.OrgNumber != "202100-0076" {
		t.Errorf("org: got %q", got.OrgNumber)
	}
	if got.Year != 2024 {
		t.Errorf("year: got %d", got.Year)
	}
	if got.ExpenditureMdkr != 45.3 { // 44.8 + 0.5
		t.Errorf("expenditure sum: got %v want 45.3", got.ExpenditureMdkr)
	}
	if len(got.History) != 2 { // 2023 + 2024
		t.Errorf("history len: got %d want 2", len(got.History))
	}
	if got.History[0].Year != 2023 || got.History[1].Year != 2024 {
		t.Errorf("history order: %+v", got.History)
	}
}

func TestAuthorityExpenditureWorker_FetchErrorNoUpdate(t *testing.T) {
	client := &mockAllAnslagClient{err: errors.New("statskontoret down")}
	repo := &mockExpenditureRepo{}
	w := workers.NewAuthorityExpenditureWorker(client, repo)

	if err := w.Run(context.Background()); err == nil {
		t.Error("expected error when fetch fails")
	}
	if len(repo.captured) != 0 {
		t.Error("must not update on fetch failure")
	}
}
