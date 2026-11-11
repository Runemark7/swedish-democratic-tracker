// Package loften serves Valmanifest and Löfte as public, read-only data.
package loften

import (
	"context"

	"riksdagskollen/internal/loften/domain"
	"riksdagskollen/internal/loften/ports"
)

type Service struct{ repo ports.Repository }

func NewService(repo ports.Repository) *Service { return &Service{repo: repo} }

// List returns public Löften, optionally for one party ("" = all).
func (s *Service) List(ctx context.Context, party string) ([]*domain.Loefte, error) {
	return s.repo.ListPublic(ctx, party)
}

// Get returns a public Löfte, or nil when it does not exist or is not public.
func (s *Service) Get(ctx context.Context, id int) (*domain.Loefte, error) {
	return s.repo.GetPublic(ctx, id)
}

// Valmanifest lists Valmanifest, newest election first, optionally for one party.
func (s *Service) Valmanifest(ctx context.Context, party string) ([]*domain.Valmanifest, error) {
	return s.repo.ListValmanifest(ctx, party)
}
