// Package datacoverage serves the Data section's per-source, per-riksmöte
// coverage, read from the ingestion coverage records.
package datacoverage

import (
	"context"

	"riksdagskollen/internal/datacoverage/domain"
	"riksdagskollen/internal/datacoverage/ports"
)

type Service struct{ repo ports.Repository }

func NewService(repo ports.Repository) *Service { return &Service{repo: repo} }

func (s *Service) Coverage(ctx context.Context) ([]domain.Row, error) { return s.repo.List(ctx) }
