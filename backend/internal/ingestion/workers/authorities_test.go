package workers_test

import (
	"context"
	"errors"
	"testing"

	"riksdagskollen/internal/ingestion/workers"
	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
)

type mockRegisterClient struct {
	entries []ports.RegisterEntry
	err     error
}

func (m *mockRegisterClient) FetchRegister(_ context.Context) ([]ports.RegisterEntry, error) {
	return m.entries, m.err
}

type mockAuthorityRepo struct {
	upserted  []domain.RegisteredAuthority
	upsertErr error
}

func (m *mockAuthorityRepo) UpsertAuthorities(_ context.Context, items []domain.RegisteredAuthority) (int, error) {
	if m.upsertErr != nil {
		return 0, m.upsertErr
	}
	m.upserted = append(m.upserted, items...)
	return len(items), nil
}
func (m *mockAuthorityRepo) List(_ context.Context, _ ports.AuthorityFilter) ([]domain.RegisteredAuthority, error) {
	return nil, nil
}
func (m *mockAuthorityRepo) Count(_ context.Context, _ ports.AuthorityFilter) (int, error) {
	return 0, nil
}

func TestAuthoritiesWorker_MapsAndUpserts(t *testing.T) {
	client := &mockRegisterClient{entries: []ports.RegisterEntry{
		{OrgNumber: "202100-2114", Slug: "arbetsformedlingen", Name: "Arbetsförmedlingen", Type: "Förvaltningsmyndighet", PrincipalBody: "Regeringen", UnderGovernment: true, Website: "www.arbetsformedlingen.se", SFS: "2007:1030"},
	}}
	repo := &mockAuthorityRepo{}
	w := workers.NewAuthoritiesWorker(client, repo)

	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.upserted) != 1 {
		t.Fatalf("expected 1 upserted, got %d", len(repo.upserted))
	}
	got := repo.upserted[0]
	if got.Name != "Arbetsförmedlingen" || got.OrgNumber != "202100-2114" || !got.UnderGovernment {
		t.Errorf("mapped wrong: %+v", got)
	}
	if got.ExpenditureMdkr != nil || got.HeadcountInt != nil {
		t.Errorf("Phase 1 must leave expenditure/headcount nil, got %+v", got)
	}
}

func TestAuthoritiesWorker_FetchErrorPropagates(t *testing.T) {
	client := &mockRegisterClient{err: errors.New("scrape down")}
	repo := &mockAuthorityRepo{}
	w := workers.NewAuthoritiesWorker(client, repo)

	if err := w.Run(context.Background()); err == nil {
		t.Error("expected error when fetch fails")
	}
	if len(repo.upserted) != 0 {
		t.Error("must not upsert on fetch failure (keep last good data)")
	}
}
